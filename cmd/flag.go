package cmd

import (
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

type flag struct {
	ConfigFile string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.ConfigFile, "file", "f", "images.yaml", "retagger config file to use")
}

func (f *flag) Validate() error {
	if f.ConfigFile == "" {
		return microerror.Maskf(invalidFlagsError, "file flag must not be empty")
	}

	return nil
}
