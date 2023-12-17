# CMD

This is using Cobra for the setup and Viper for the configuration. Both packages can be pretty frustrating setting up, and if you only have one command it can feel like an overkill. However, often when you're done with your service you realise you also need a cronjob that does some cleanup. In this example application, it could be that the counters should be reset nightly. Then rather than copying the model and creating something new from scratch you can add a reset command, add the functions needed and call them from that command. You avoid duplicating code and you have your own boilerplate in place, and if you need to call existing service functions directly you can just do so.

As for Viper, there are a ton of different configuration packages out there. I selected Viper because it's very flexible: It can refresh the configuration, it supports remote key/value stores and env variables, flags and files.

When dumping and listing configuration with `--cfg-dump` everything is written in alphabethical order. For that reason are all flags prefixed with what they are about. That's why there's `mid-cors` and `mid-prom-timer`, all middleware flags will be listed together. It's also possible to use hierarchies within the flags, so that all middleware related flags are under the key (viper uses hashes) `middleware` and then a key `cors` and under there having all the cors related flags. However, configuration should be simple. Chances are that you'll not be the one scratching your head at 3 in the morning trying to fix a broken deployment.

## root.go

This is the top level command when just running it, like `go run main.go`. This is the place for the global flags, which are setup in the init() function, it also reads the configuration file if provided in initConfig() and lastsly executes any sub command provided.

The global flags this example supports are the following:
```go 
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config.yml", cfgFile, "config file")
	rootCmd.PersistentFlags().StringVar(&base.LogMinLevel, "log-lvl", "info", "Minimum log level to display: debug|info|warn|error")
	rootCmd.PersistentFlags().StringVar(&base.LogOutputFormat, "log-format", "text", "Output logs in text or json")
	rootCmd.PersistentFlags().StringVar(&base.LogTarget, "log-target", "stdout", "Output logs to stdout or stderr")
	rootCmd.PersistentFlags().Bool(FlagCfgDump, false, "Prints current config and exits")
	rootCmd.PersistentFlags().Bool(FlagCfgWrite, false, "Saves current config to disk (target --config) and exits")

``` 
### Flags

#### config

Path to the configuration file to use, it defaults to config.yml in the directory where the command is executed. The file is only read once, if you want to restart on configuration changes you can call viper.OnConfigChange and then call viper.WatchConfig. Viper supports a lot of different formats, including YAML and JSON. So if you want to use a confiuration file, write one and use this flag.

### log-lvl

Hopefully self explantory.

### log-format

Plain text or json output from the logger.

### log-target

Printing the logs to STDERR or STDOUT

### cfg-dump

Prints all current configuration, with the default values if nothing is set, to STDOUT and exits

### cfg-save

Prints all current confuration, except cfg-save and cfg-dump, to the file the config flag is set to

## serve.go

Serve.go starts the service after configuring all the flags. As there are a lot of them are they separated so it's possible to loop over them. Typing is used so that if called with the wrong type it fails to start:

```bash
you@puter:~/projects/http-skeleton$ go run main.go serve --http-port foo
Error: invalid argument "foo" for "--http-port" flag: strconv.ParseInt: parsing "foo": invalid syntax
...
      --http-port int                       Public facing http port to listen to (default 3000)
...
you@puter:~/projects/http-skeleton$ go run main.go serve --http-port 0x539 --cfg-dump
...
  "http-port": 1337,
...
you@puter:~/projects/http-skeleton$ go run main.go serve --http-port 1338 --cfg-dump
...
  "http-port": 1338,
...
```
It sets the default, and the environment format:
```go
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))
	viper.AutomaticEnv()
```
What this does is that all flags are also available as environment variables, and adjusted to work with the naming rules for them. The flag `--cfg-dump` can be instead be set by having CFG_DUMP set:
```bash
jmo@jmo:~/projects/http-skeleton$ CFG_DUMP=1 go run main.go serve
``` 
Viper also supports having a prefix if you have name colissions, say you have `--path` added, it's likely set in your environment by the OS, you can then call `viper.SetEnvPrefix("MY_APP")`, which then makes Viper to only look at env variables starting with `MY_APP_`.

When all is setup and checked, [serve](serve/README.md) is called
