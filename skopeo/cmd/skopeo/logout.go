package main

import (
	"io"

	"github.com/containers/common/pkg/auth"
	commonFlag "github.com/containers/common/pkg/flag"
	"github.com/containers/image/v5/types"
	"github.com/spf13/cobra"
)

type logoutOptions struct {
	global     *globalOptions
	logoutOpts auth.LogoutOptions
	tlsVerify  commonFlag.OptionalBool
}

func logoutCmd(global *globalOptions) *cobra.Command {
	opts := logoutOptions{
		global: global,
	}
	cmd := &cobra.Command{
		Use:     "logout [command options] REGISTRY",
		Short:   "Logout of a container registry",
		Long:    "Logout of a container registry on a specified server.",
		RunE:    commandAction(opts.run),
		Example: `skopeo logout quay.io`,
	}
	adjustUsage(cmd)
	flags := cmd.Flags()
	commonFlag.OptionalBoolFlag(flags, &opts.tlsVerify, "tls-verify", "require HTTPS and verify certificates when accessing the registry")
	flags.AddFlagSet(auth.GetLogoutFlags(&opts.logoutOpts))
	return cmd
}

func (opts *logoutOptions) run(args []string, stdout io.Writer) error {
	opts.logoutOpts.Stdout = stdout
	opts.logoutOpts.AcceptRepositories = true
	sys := opts.global.newSystemContext()
	if opts.tlsVerify.Present() {
		sys.DockerInsecureSkipTLSVerify = types.NewOptionalBool(!opts.tlsVerify.Value())
	}
	return auth.Logout(sys, &opts.logoutOpts, args)
}
