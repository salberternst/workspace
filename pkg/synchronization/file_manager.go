package synchronization

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/fatih/color"

	"github.com/dustin/go-humanize"

	"github.com/mutagen-io/mutagen/cmd/mutagen/common/templating"
	"github.com/mutagen-io/mutagen/pkg/logging"
	"github.com/mutagen-io/mutagen/pkg/selection"
	"github.com/mutagen-io/mutagen/pkg/synchronization"
	_ "github.com/mutagen-io/mutagen/pkg/synchronization/protocols/local"
	_ "github.com/mutagen-io/mutagen/pkg/synchronization/protocols/ssh"
	"github.com/mutagen-io/mutagen/pkg/synchronization/rsync"
	"github.com/mutagen-io/mutagen/pkg/url"
)

var monitorConfiguration struct {
	templating.TemplateFlags
}

type FileManager struct {
	sessionName            string
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

func (o *FileManager) createSession(name string, source string, target Target, ignores []string, labels map[string]string, watch bool, syncMode string) error {
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
		name,
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
				fmt.Println(computeMonitorStatusLine(sessionStates[0]))
			} else {
				fmt.Println(err.Error())
				return
			}
		}
	}()
}

func (o *FileManager) Run(name string, source string, target Target, ignores []string, labels map[string]string, watch bool, syncMode string) error {
	if err := o.createSession(name, source, target, ignores, labels, watch, syncMode); err != nil {
		return err
	}

	// wait for session to become active
	if err := o.waitForSession(); err != nil {
		return err
	}

	o.logSession()

	// do not flush if watch mode is disabled
	if watch {
		return nil
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

func computeMonitorStatusLine(state *synchronization.State) string {
	// Build the status line.
	var status string
	if state.Session.Paused {
		status += color.YellowString("[Paused]")
	} else {
		// Add a conflict flag if there are conflicts.
		if len(state.Conflicts) > 0 {
			status += color.YellowString("[C] ")
		}

		// Add a problems flag if there are problems.
		haveProblems := len(state.AlphaState.ScanProblems) > 0 ||
			len(state.BetaState.ScanProblems) > 0 ||
			len(state.AlphaState.TransitionProblems) > 0 ||
			len(state.BetaState.TransitionProblems) > 0
		if haveProblems {
			status += color.YellowString("[!] ")
		}

		// Add an error flag if there is one present.
		if state.LastError != "" {
			status += color.RedString("[X] ")
		}

		// Handle the formatting based on status. If we're in a staging mode,
		// then extract the relevant progress information. Despite not having a
		// built-in mechanism for knowing the total expected size of a staging
		// operation, we do know the number of files that the staging operation
		// is performing, so if that's equal to the number of files on the
		// source endpoint, then we know that we can use the total file size on
		// the source endpoint as an estimate for the total staging size.
		var stagingProgress *rsync.ReceiverState
		var totalExpectedSize uint64
		if state.Status == synchronization.Status_StagingAlpha {
			status += "[←] "
			stagingProgress = state.AlphaState.StagingProgress
			if stagingProgress == nil {
				status += "Preparing to stage files on alpha"
			} else if stagingProgress.ExpectedFiles == state.BetaState.Files {
				totalExpectedSize = state.BetaState.TotalFileSize
			}
		} else if state.Status == synchronization.Status_StagingBeta {
			status += "[→] "
			stagingProgress = state.BetaState.StagingProgress
			if stagingProgress == nil {
				status += "Preparing to stage files on beta"
			} else if stagingProgress.ExpectedFiles == state.AlphaState.Files {
				totalExpectedSize = state.AlphaState.TotalFileSize
			}
		} else {
			status += state.Status.Description()
		}

		// Print staging progress, if available.
		if stagingProgress != nil {
			var fractionComplete float32
			var totalSizeDenominator string
			if totalExpectedSize != 0 {
				fractionComplete = float32(stagingProgress.TotalReceivedSize) / float32(totalExpectedSize)
				totalSizeDenominator = "/" + humanize.Bytes(totalExpectedSize)
			} else {
				fractionComplete = float32(stagingProgress.ReceivedFiles) / float32(stagingProgress.ExpectedFiles)
			}
			status += fmt.Sprintf("[%d/%d - %s%s - %.0f%%] %s (%s/%s)",
				stagingProgress.ReceivedFiles, stagingProgress.ExpectedFiles,
				humanize.Bytes(stagingProgress.TotalReceivedSize), totalSizeDenominator,
				100.0*fractionComplete,
				path.Base(stagingProgress.Path),
				humanize.Bytes(stagingProgress.ReceivedSize), humanize.Bytes(stagingProgress.ExpectedSize),
			)
		}
	}

	// Done.
	return status
}
