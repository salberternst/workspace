package synchronization

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/schollz/progressbar/v3"

	"github.com/mutagen-io/mutagen/cmd/mutagen/common/templating"
	"github.com/mutagen-io/mutagen/pkg/logging"
	"github.com/mutagen-io/mutagen/pkg/selection"
	"github.com/mutagen-io/mutagen/pkg/synchronization"
	_ "github.com/mutagen-io/mutagen/pkg/synchronization/protocols/local"
	_ "github.com/mutagen-io/mutagen/pkg/synchronization/protocols/ssh"
	"github.com/mutagen-io/mutagen/pkg/url"
)

var monitorConfiguration struct {
	templating.TemplateFlags
}

type FileManager struct {
	sessionName            string
	progressBar            *progressbar.ProgressBar
	synchronizationManager *synchronization.Manager
}

func NewFileManager() (*FileManager, error) {
	logging := logging.NewLogger(logging.LevelDisabled, os.Stderr)

	manager, err := synchronization.NewManager(logging)
	if err != nil {
		return nil, err
	}

	return &FileManager{
		synchronizationManager: manager,
		progressBar:            progressbar.Default(-1),
	}, nil
}

func (o *FileManager) waitForSession() error {
	ready := make(chan bool, 1)
	failed := make(chan bool, 1)

	go func() {
		var previousStateIndex uint64
		previousStateIndex = 0

		for {
			stateIndex, sessionStates, err := o.synchronizationManager.List(context.TODO(), &selection.Selection{
				Specifications: []string{
					o.sessionName,
				},
			}, uint64(previousStateIndex))

			if err != nil {
				failed <- true
			} else {
				previousStateIndex = stateIndex
				if sessionStates[0].Status == synchronization.Status_Watching {
					ready <- true
				}
			}
		}
	}()

	select {
	case <-ready:
		return nil
	case <-failed:
		return fmt.Errorf("Failed to get status of session %s\n", o.sessionName)
	case <-time.After(30 * time.Second):
		return errors.New("timeout occured")
	}
}

func (o *FileManager) createSession(source string, target Target, ignores []string, labels map[string]string, watch bool, syncMode string) error {
	alpha, err := url.Parse(source, url.Kind_Synchronization, true)
	if err != nil {
		return err
	}

	beta, err := url.Parse(target.buildUrl(), url.Kind_Synchronization, false)
	if err != nil {
		return err
	}

	configuration := &synchronization.Configuration{
		Ignores:             ignores,
		SynchronizationMode: getSyncMode(syncMode),
	}

	if !watch {
		configuration.WatchMode = synchronization.WatchMode_WatchModeNoWatch
	}

	configurationAlpha := &synchronization.Configuration{}
	configurationBeta := &synchronization.Configuration{}

	o.sessionName, err = o.synchronizationManager.Create(context.TODO(),
		alpha,
		beta,
		configuration,
		configurationAlpha,
		configurationBeta,
		uuid.NewString(),
		labels,
		false,
		"")

	return err
}

func (o *FileManager) logSession() {
	go func() {
		var previousStateIndex uint64
		previousStateIndex = 0

		for {
			stateIndex, sessionStates, err := o.synchronizationManager.List(context.TODO(), &selection.Selection{
				Specifications: []string{
					o.sessionName,
				},
			}, uint64(previousStateIndex))

			if err == nil {
				previousStateIndex = stateIndex
				o.progressBar.Describe(sessionStates[0].Status.Description())
				o.progressBar.Add(1)
			} else {
				fmt.Println(err.Error())
				return
			}
		}
	}()
}

func (o *FileManager) Run(source string, target Target, ignores []string, labels map[string]string, watch bool, syncMode string) error {
	if err := o.createSession(source, target, ignores, labels, watch, syncMode); err != nil {
		return err
	}

	o.logSession()

	// do not flush if watch mode is disabled
	if watch {
		return nil
	}

	// wait for session to become active
	if err := o.waitForSession(); err != nil {
		return err
	}

	return o.synchronizationManager.Flush(context.TODO(), &selection.Selection{
		Specifications: []string{
			o.sessionName,
		},
	}, "", false)
}

func (o *FileManager) Stop() {
	if o.synchronizationManager != nil {
		o.synchronizationManager.Terminate(context.TODO(), &selection.Selection{
			Specifications: []string{
				o.sessionName,
			},
		}, "")

		o.synchronizationManager.Shutdown()
	}
}
