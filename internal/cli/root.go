package cli

import "github.com/spf13/cobra"

func NewRootCommand() *cobra.Command {
	return NewRootCommandWithOptions(defaultOptions())
}

func NewRootCommandWithOptions(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "yx",
		Short:        "Yunxiao command line client",
		Long:         "yx is a Yunxiao command line client.",
		SilenceUsage: true,
	}
	cmd.PersistentFlags().String("profile", "", "profile to use for this invocation")
	cmd.PersistentFlags().String("org", "", "organization override for this invocation")
	cmd.PersistentFlags().String("domain", "", "domain override for this invocation")
	cmd.PersistentFlags().Bool("json", false, "write JSON output")
	cmd.PersistentFlags().Bool("verbose", false, "write verbose diagnostics")
	cmd.SetHelpTemplate(`{{with .Long}}{{.}}

{{end}}Usage:
  {{.UseLine}}
`)
	cmd.AddCommand(newConfigCommand(opts))

	return cmd
}
