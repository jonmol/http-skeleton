package config

import "time"

type IntConf struct {
	Name string
	Desc string
	Def  int
}
type DurationConf struct {
	Name string
	Desc string
	Def  time.Duration
}

type StringConf struct {
	Name string
	Desc string
	Def  string
}

type BoolConf struct {
	Name string
	Desc string
	Def  bool
}

type StringArrayConf struct {
	Name string
	Desc string
	Def  []string
}

type Env struct {
	FlagName string
	EnvName  string
}

type Configs struct {
	Ints         []IntConf
	Strings      []StringConf
	Durations    []DurationConf
	Bools        []BoolConf
	StringArrays []StringArrayConf
}
