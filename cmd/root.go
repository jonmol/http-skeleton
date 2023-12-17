/*
Copyright © 2023 Jon <No mails please>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"errors"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jonmol/http-skeleton/util/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  = "config.yml"
	validate *validator.Validate
	base     BaseConfig
)

type BaseConfig struct {
	LogOutputFormat string `mapstructure:"log-format" validate:"omitempty,oneof=text json"`
	LogMinLevel     string `mapstructure:"log-lvl" validate:"omitempty,oneof=debug info warn error"`
	LogTarget       string `mapstructure:"log-handle" validate:"omitempty,oneof=stdio stderr"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "http-skeleton",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	validate = validator.New()
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config.yml", cfgFile, "config file")
	rootCmd.PersistentFlags().StringVar(&base.LogMinLevel, "log-lvl", "info", "Minimum log level to display: debug|info|warn|error")
	rootCmd.PersistentFlags().StringVar(&base.LogOutputFormat, "log-format", "text", "Output logs in text or json")
	rootCmd.PersistentFlags().StringVar(&base.LogTarget, "log-target", "stdout", "Output logs to stdout or stderr")
	rootCmd.PersistentFlags().Bool(FlagCfgDump, false, "Prints current config and exits")
	rootCmd.PersistentFlags().Bool(FlagCfgWrite, false, "Saves current config to disk (target --config) and exits")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		slog.Error("Failed to bind flags!", logging.Err(err))
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("Failed to load config file!", logging.Err(err), slog.String("fileName", viper.ConfigFileUsed()))
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Info("No config file found", slog.String("fileName", viper.ConfigFileUsed()))
	} else {
		slog.Debug("Used config file", slog.String("fileName", viper.ConfigFileUsed()))
	}

	var base BaseConfig
	if err := viper.Unmarshal(&base); err != nil {
		slog.Error("Failed to unmarshal config file!", logging.Err(err))
	}
	if err := validate.Struct(base); err != nil {
		slog.Error("Base configuration failed to validate", logging.Err(err))
	}
	setupLogger(base)
}
