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
		Example:      "  yx auth login\n  yx repo list\n  yx repo view <repo>\n  yx mr list --repo <repo>\n  yx pipeline list",
		SilenceUsage: true,
	}
	cmd.PersistentFlags().String("profile", "", "profile to use for this invocation")
	cmd.PersistentFlags().String("org", "", "organization override for this invocation")
	cmd.PersistentFlags().String("domain", "", "domain override for this invocation")
	cmd.PersistentFlags().Bool("json", false, "write JSON output")
	cmd.PersistentFlags().Bool("verbose", false, "write verbose diagnostics")
	cmd.AddCommand(newConfigCommand(opts))
	cmd.AddCommand(newAuthCommand(opts))
	cmd.AddCommand(newRepoCommand(opts))
	cmd.AddCommand(newMergeRequestCommand(opts, "mr"))
	cmd.AddCommand(newMergeRequestCommand(opts, "pr"))
	cmd.AddCommand(newProjectCommand(opts))
	cmd.AddCommand(newWorkitemCommand(opts, "workitem"))
	cmd.AddCommand(newWorkitemCommand(opts, "issue"))
	cmd.AddCommand(newPipelineCommand(opts))
	applyHelpTemplate(cmd)

	return cmd
}

const helpTemplate = `{{with (or .Long .Short)}}{{.}}

{{end}}{{if .Example}}Examples:
{{.Example}}

{{end}}Usage:
  {{.UseLine}}
{{if .HasAvailableSubCommands}}

Available Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`

func applyHelpTemplate(cmd *cobra.Command) {
	cmd.SetHelpTemplate(helpTemplate)
	for _, child := range cmd.Commands() {
		applyHelpTemplate(child)
	}
}
