package builder

import "github.com/spf13/cobra"

type WorkspaceArgs struct {
	Description          string
	RequestGpu           int
	RequestGpuType       string
	RequestCpu           string
	RequestMemory        string
	LimitCpu             string
	LimitMemory          string
	Image                string
	ImageGpu             string
	ImagePullPolicy      string
	AdditionalVolumes    []string
	InstallCondaPackages []string
	InstallPipPackages   []string
	Args
}

func (o *WorkspaceArgs) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Description, o.addPrefix("description"), "", "Description of the workplace")
	cmd.Flags().IntVar(&o.RequestGpu, o.addPrefix("request-gpu"), 0, "The gpu resource to use")
	cmd.Flags().StringVar(&o.RequestGpuType, o.addPrefix("request-gpu-type"), "", "The requested gpu type")
	cmd.Flags().StringVar(&o.RequestCpu, o.addPrefix("request-cpu"), "", "The cpu resource to use")
	cmd.Flags().StringVar(&o.RequestMemory, o.addPrefix("request-memory"), "", "The memory resource to use")
	cmd.Flags().StringVar(&o.LimitCpu, o.addPrefix("limit-cpu"), "", "The cpu resource limit")
	cmd.Flags().StringVar(&o.LimitMemory, o.addPrefix("limit-memory"), "", "The memory resource limit")
	cmd.Flags().StringArrayVar(&o.AdditionalVolumes, o.addPrefix("volume"), []string{}, "List of additional volumes to mount in the form of volume:mount-path (e.g. volume-name:/data)")
	cmd.Flags().StringArrayVar(&o.InstallCondaPackages, o.addPrefix("install-conda-package"), []string{}, "List of conda-forge packages to install in the workspace")
	cmd.Flags().StringArrayVar(&o.InstallPipPackages, o.addPrefix("install-pip-package"), []string{}, "List of pip packages to install in the workspace")
	cmd.Flags().StringVar(&o.Image, o.addPrefix("override-image"), "", "Override the workspace cpu image")
	cmd.Flags().StringVar(&o.ImageGpu, o.addPrefix("override-image-gpu"), "", "Override the workspace gpu image")
	cmd.Flags().StringVar(&o.ImagePullPolicy, o.addPrefix("image-pull-policy"), "", "Set the image pull policy")
}

func (o *WorkspaceArgs) BuildValues(cmd *cobra.Command) map[string]interface{} {
	o.buildValueIfChanged(cmd, o.Description, o.addPrefix("description"), "description")
	o.buildValueIfChanged(cmd, o.RequestGpu, o.addPrefix("request-gpu"), "requests.gpu")
	o.buildValueIfChanged(cmd, o.RequestGpuType, o.addPrefix("request-gpu-type"), "requests.gpuType")
	o.buildValueIfChanged(cmd, o.RequestCpu, o.addPrefix("request-cpu"), "requests.cpu")
	o.buildValueIfChanged(cmd, o.RequestMemory, o.addPrefix("request-memory"), "requests.memory")
	o.buildValueIfChanged(cmd, o.LimitCpu, o.addPrefix("limit-cpu"), "limits.cpu")
	o.buildValueIfChanged(cmd, o.LimitMemory, o.addPrefix("limit-memory"), "limits.memory")
	o.buildValueIfChanged(cmd, o.AdditionalVolumes, o.addPrefix("volume"), "additionalVolumes")
	o.buildValueIfChanged(cmd, o.InstallCondaPackages, o.addPrefix("install-conda-package"), "installCondaPackages")
	o.buildValueIfChanged(cmd, o.InstallPipPackages, o.addPrefix("install-pip-package"), "installPipPackages")
	o.buildValueIfChanged(cmd, o.Image, o.addPrefix("override-image"), "image")
	o.buildValueIfChanged(cmd, o.ImageGpu, o.addPrefix("override-image-gpu"), "imageGpu")
	o.buildValueIfChanged(cmd, o.ImagePullPolicy, o.addPrefix("image-pull-policy"), "imagePullPolicy")
	return o.values.GetMap()
}

func NewWorkspaceArgs(prefix string) WorkspaceArgs {
	return WorkspaceArgs{
		AdditionalVolumes: []string{},
		Args: Args{
			Prefix: prefix,
			values: NewValues(),
		},
	}
}
