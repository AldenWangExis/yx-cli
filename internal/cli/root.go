package cli

import "github.com/spf13/cobra"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yx",
		Short: "Yunxiao command line client",
		Long:  "yx is a Yunxiao command line client.",
	}
	cmd.SetHelpTemplate(`{{with .Long}}{{.}}

{{end}}Usage:
  {{.UseLine}}
`)

	return cmd
}
