package cmd

import (
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

type flag struct {
	ConfigFile   string
	Host         string
	Organization string
	Username     string
	Password     string
	WhatIf       bool
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.ConfigFile, "file", "f", "images.yaml", "retagger config file to use")
	cmd.Flags().StringVarP(&f.Host, "host", "r", "", "Registry hostname (e.g. quay.io)")
	cmd.Flags().StringVarP(&f.Organization, "organization", "o", "giantswarm", "organization to tag images for")
	cmd.Flags().StringVarP(&f.Username, "username", "u", "", "username to authenticate against registry")
	cmd.Flags().StringVarP(&f.Password, "password", "p", "", "password to authenticate against registry")
	cmd.Flags().BoolVar(&f.WhatIf, "whatif", false, "if set, will list jobs but not run them")
}

func (f *flag) Validate() error {
	if f.ConfigFile == "" {
		return microerror.Maskf(invalidFlagsError, "file flag must not be empty")
	}
	if f.Host == "" {
		return microerror.Maskf(invalidFlagsError, "host flag must not be empty")
	}
	if f.Organization == "" {
		return microerror.Maskf(invalidFlagsError, "organization flag must not be empty")
	}
	if f.Username == "" {
		return microerror.Maskf(invalidFlagsError, "username flag must not be empty")
	}
	if f.Password == "" {
		return microerror.Maskf(invalidFlagsError, "password flag must not be empty")
	}

	return nil
}
