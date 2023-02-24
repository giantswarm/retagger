package main

import (
	"io"
	"os"

	"github.com/containers/common/pkg/auth"
	commonFlag "github.com/containers/common/pkg/flag"
	"github.com/containers/image/v5/types"
	"github.com/spf13/cobra"
)

type loginOptions struct {
	global    *globalOptions
	loginOpts auth.LoginOptions
	tlsVerify commonFlag.OptionalBool
}

func loginCmd(global *globalOptions) *cobra.Command {
	opts := loginOptions{
		global: global,
	}
	cmd := &cobra.Command{
		Use:     "login [command options] REGISTRY",
		Short:   "Login to a container registry",
		Long:    "Login to a container registry on a specified server.",
		RunE:    commandAction(opts.run),
		Example: `skopeo login quay.io`,
	}
	adjustUsage(cmd)
	flags := cmd.Flags()
	commonFlag.OptionalBoolFlag(flags, &opts.tlsVerify, "tls-verify", "require HTTPS and verify certificates when accessing the registry")
	flags.AddFlagSet(auth.GetLoginFlags(&opts.loginOpts))
	return cmd
}

func (opts *loginOptions) run(args []string, stdout io.Writer) error {
	ctx, cancel := opts.global.commandTimeoutContext()
	defer cancel()
	opts.loginOpts.Stdout = stdout
	opts.loginOpts.Stdin = os.Stdin
	opts.loginOpts.AcceptRepositories = true
	sys := opts.global.newSystemContext()
	if opts.tlsVerify.Present() {
		sys.DockerInsecureSkipTLSVerify = types.NewOptionalBool(!opts.tlsVerify.Value())
	}
	return auth.Login(ctx, sys, &opts.loginOpts, args)
}
