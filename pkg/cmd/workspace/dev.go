package workspace

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/salberternst/workspace/pkg/k8s"
	"github.com/salberternst/workspace/pkg/synchronization"
	"github.com/salberternst/workspace/pkg/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

const PrivateKeySecretKey = "ssh_host_ecdsa_key"

type DevOptions struct {
	Name            string
	Namespace       string
	SshPort         uint16
	DisableTerminal bool
	Source          string
	Target          string
	TargetVolume    string
	TargetFolder    string
	SyncFolder      string
	SyncIgnores     []string
	Labels          map[string]string
	SyncWatch       bool
	SyncMode        string
	workspacePod    *v1.Pod
	fileManager     *synchronization.FileManager
	portForward     k8s.PortForward
}

func (o *DevOptions) buildPorts() []string {
	return []string{
		strconv.Itoa(int(o.SshPort)) + ":2222",
	}
}

func (o *DevOptions) buildTarget() synchronization.Target {
	return synchronization.Target{
		Port:     2222,
		Hostname: fmt.Sprintf("%s.%s.workspace", o.Name, o.Namespace),
		Folder:   o.Target,
		Username: "workspace",
	}
}

func (o *DevOptions) createPortForward() error {
	var err error

	if o.portForward, err = k8s.GetClient().ForwardPorts(o.workspacePod.Name, o.workspacePod.Namespace, o.buildPorts()); err != nil {
		return err
	}

	return nil
}

func (o *DevOptions) createSynchronizationManager() error {
	var err error
	o.fileManager, err = synchronization.NewFileManager()
	return err
}

func (o *DevOptions) setupSshConfig() error {
	secret, err := k8s.ReadSecret(o.Name, o.Namespace)
	if err != nil {
		return err
	}

	privateKey, ok := secret.Data[PrivateKeySecretKey]
	if !ok {
		return fmt.Errorf("ssh_host_ecdsa_key does not exists in secret %s in namespace %s", o.Name, o.Namespace)
	}

	privateKeyPath, err := utils.WritePrivateKey(o.Name, o.Namespace, privateKey)
	if err != nil {
		return err
	}

	err = utils.DeleteSshConfEntry(o.Name, o.Namespace)
	if err != nil {
		return err
	}

	err = utils.AppendSshConfEntry(o.Name, o.Namespace, privateKeyPath)
	if err != nil {
		return err
	}

	return nil
}

func (o *DevOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument: name")
	}

	var err error

	o.Name = args[0]

	if o.Namespace, err = cmd.Flags().GetString("project"); err != nil {
		return err
	}

	o.workspacePod, err = k8s.GetWorkspacePod(o.Namespace, o.Name)
	if err != nil {
		return err
	}

	if o.SyncFolder != "" {
		target := strings.Split(o.SyncFolder, ":")
		if len(target) != 2 {
			return fmt.Errorf("Invalid sync folder %s", o.SyncFolder)
		}

		o.Source = target[0]
		o.Target = target[1]

		if err := o.createSynchronizationManager(); err != nil {
			return err
		}
	}

	return o.setupSshConfig()
}

func (o *DevOptions) Run() error {
	if err := o.createPortForward(); err != nil {
		return err
	}

	if o.SyncFolder != "" {
		if err := o.fileManager.Run(o.Source, o.buildTarget(), o.SyncIgnores, o.Labels, o.SyncWatch, o.SyncMode); err != nil {
			return err
		}
	}

	defer o.fileManager.Stop()

	if o.DisableTerminal {
		signalTermination := make(chan os.Signal, 1)
		signal.Notify(signalTermination, syscall.SIGINT, syscall.SIGTERM)

		fmt.Println("Press CTRL+C to stop")

		select {
		case <-o.portForward.StopChannel:
			return nil
		case <-signalTermination:
			return nil
		}
	}

	return k8s.ExecuteInPod(o.workspacePod.Namespace, o.workspacePod.Name, "workspace", []string{"bash", "--login"}, true)
}

func NewCmdDev() *cobra.Command {
	options := DevOptions{}

	var command = &cobra.Command{
		Use: "dev [name]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(cmd, args); err != nil {
				return err
			}

			fmt.Println(utils.Logo)
			fmt.Printf("Connect via: ssh %s.%s.workspace\n", options.Name, options.Namespace)

			return options.Run()
		},
	}

	command.Flags().Uint16Var(&options.SshPort, "ssh-port", 2222, "The local ssh port")
	command.Flags().BoolVar(&options.DisableTerminal, "disable-terminal", false, "Disable the terminal")
	command.Flags().StringArrayVar(&options.SyncIgnores, "sync-ignore", []string{".mutagen", ".git"}, "List of folders and files to ignore")
	command.Flags().StringToStringVar(&options.Labels, "sync-label", map[string]string{}, "List of custom labels to add")
	command.Flags().BoolVar(&options.SyncWatch, "sync-watch", false, "Continuously synchronize file changes to the workspace")
	command.Flags().StringVar(&options.SyncMode, "sync-mode", "twowaysafe", "Set the synchonization mode see https://mutagen.io/documentation/synchronization")
	command.Flags().StringVar(&options.SyncFolder, "sync-folder", "", "Synchronize a folder to the workspace")

	return command
}
