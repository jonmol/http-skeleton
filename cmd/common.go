package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jonmol/http-skeleton/util/logging"
	"github.com/spf13/viper"
)

const (
	FlagCfgDump  = "cfg-dump"
	FlagCfgWrite = "cfg-save"
)

var serviceID = uuid.Must(uuid.NewV4())

func handleGlobalFlags() {
	exit := false

	if viper.GetBool(FlagCfgDump) {
		j, err := json.MarshalIndent(viper.AllSettings(), " ", " ")
		if err != nil {
			slog.Error("Failed to marshal into json:", slog.Any("settings", viper.AllSettings()))
		}
		fmt.Println(string(j))
		exit = true
	}

	if viper.GetBool(FlagCfgWrite) {
		viper.Set(FlagCfgWrite, false)
		viper.Set(FlagCfgDump, false)
		err := viper.WriteConfig()
		if err != nil {
			slog.Error("Failed to write config", logging.Err(err))
			os.Exit(1)
		}
		slog.Info("Wrote config to", slog.String("cfgFile", viper.ConfigFileUsed()))
		exit = true
	}

	if exit {
		os.Exit(0)
	}
}

// setupLogger creates a logger based on the configuration
func setupLogger(cfg BaseConfig) {
	if err := validate.Struct(cfg); err != nil {
		panic(err)
	}

	// set log level, INFO is default
	logLvl := &slog.LevelVar{}

	switch strings.ToLower(cfg.LogMinLevel) {
	case "debug":
		logLvl.Set(slog.LevelDebug)
	case "warn":
		logLvl.Set(slog.LevelWarn)
	case "error":
		logLvl.Set(slog.LevelError)
	}

	out := os.Stdout
	if cfg.LogTarget == "stderr" {
		out = os.Stderr
	}

	// pick the output format, store in the context and make it the default logger
	var logger *slog.Logger
	if cfg.LogOutputFormat == "text" {
		logger = slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: logLvl, AddSource: true}))
	} else {
		logger = slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: logLvl, AddSource: true}))
	}
	logger = logger.With(slog.String("serviceUID", serviceID.String()), slog.Int("pid", os.Getpid()))
	slog.SetDefault(logger)
	logger.Debug("Logger setup")
}
