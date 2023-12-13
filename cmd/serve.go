/*
Copyright © 2023 Jon <No mails please>
*/
package cmd

import (
	"log/slog"
	"strings"

	"github.com/jonmol/http-skeleton/cmd/serve"
	"github.com/jonmol/http-skeleton/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A skeleton for a web service",
	Long: `Even with the net/http library being really good, the boilerplate code
needed to have a reasonable service with defaults and such is a lot. This is me
being lazy and doing it once and hoping to not have to do these parts again.

One thing to note is that the Duration flags/config is in nano seconds if no unit is
given. That means "--http-read-timeout 2" gives a 2ns timeout. It's more likely you
want to use 1500ms or 2s.

Another thing to note is that array flags use , as the delimiter so
"--mid-cors-methods HeaderA,HeaderB" is the way to use them.
`,
	Run: func(cmd *cobra.Command, args []string) {
		handleGlobalFlags()
		serveStruct := serve.Serve{}
		serveStruct.Run()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	addFlags()
	if err := viper.BindPFlags(serveCmd.Flags()); err != nil {
		slog.Error("Failed to bind flags!", logging.Err(err))
	}
	setDefaults()

	// handle hyphens from flags in env variables, and while at it make maps/arrays accessible by using __
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))
	viper.AutomaticEnv()
}

func addFlags() {
	for _, flag := range serve.ConfigStructure.Durations {
		serveCmd.Flags().Duration(flag.Name, flag.Def, flag.Desc)
	}

	for _, flag := range serve.ConfigStructure.Strings {
		serveCmd.Flags().String(flag.Name, flag.Def, flag.Desc)
	}

	for _, flag := range serve.ConfigStructure.Ints {
		serveCmd.Flags().Int(flag.Name, flag.Def, flag.Desc)
	}

	for _, flag := range serve.ConfigStructure.Bools {
		serveCmd.Flags().Bool(flag.Name, flag.Def, flag.Desc)
	}

	for _, flag := range serve.ConfigStructure.StringArrays {
		serveCmd.Flags().StringArray(flag.Name, flag.Def, flag.Desc)
	}
}

func setDefaults() {
	for _, flag := range serve.ConfigStructure.Durations {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range serve.ConfigStructure.Strings {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range serve.ConfigStructure.Ints {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range serve.ConfigStructure.Bools {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range serve.ConfigStructure.StringArrays {
		viper.SetDefault(flag.Name, flag.Def)
	}
}
