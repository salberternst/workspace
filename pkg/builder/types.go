package builder

import (
	"github.com/spf13/cobra"
)

type ArgsBuilder interface {
	BuildValues(cmd *cobra.Command) map[string]interface{}
}

type Args struct {
	values Values
	Prefix string
}

func (o *Args) buildValueIfChanged(cmd *cobra.Command, value interface{}, name string, path string) {
	if cmd.Flags().Changed(name) {
		o.values.Set(value, path)
	}
}

func (o *Args) addPrefix(name string) string {
	if o.Prefix != "" {
		return o.Prefix + "-" + name
	}
	return name
}
