package cli

import "github.com/spf13/cobra"

type Context struct {
	Profile      string
	Organization string
	Domain       string
	JSON         bool
	Verbose      bool
}

func ContextFromCommand(cmd *cobra.Command) Context {
	flags := cmd.Flags()

	profile, _ := flags.GetString("profile")
	organization, _ := flags.GetString("org")
	domain, _ := flags.GetString("domain")
	jsonOutput, _ := flags.GetBool("json")
	verbose, _ := flags.GetBool("verbose")

	return Context{
		Profile:      profile,
		Organization: organization,
		Domain:       domain,
		JSON:         jsonOutput,
		Verbose:      verbose,
	}
}
