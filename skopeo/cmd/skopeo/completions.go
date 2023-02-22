package main

import (
	"github.com/containers/image/v5/transports"
	"github.com/spf13/cobra"
)

// autocompleteSupportedTransports list all supported transports with the colon suffix.
func autocompleteSupportedTransports(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	tps := transports.ListNames()
	suggestions := make([]string, 0, len(tps))
	for _, tp := range tps {
		suggestions = append(suggestions, tp+":")
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}
