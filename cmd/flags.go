package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func bindPFlag(flags *pflag.FlagSet, name string) {
	err := viper.BindPFlag(name, flags.Lookup(name))
	if err != nil {
		log.WithField("error", err).Fatal("Error configuring config options")
	}
}

func addStringFlag(flags *pflag.FlagSet, name, def, desc string) {
	flags.String(name, def, desc)
	bindPFlag(flags, name)
	viper.SetDefault(name, def)
}

func addBoolFlag(flags *pflag.FlagSet, name string, def bool, desc string) {
	flags.Bool(name, def, desc)
	bindPFlag(flags, name)
	viper.SetDefault(name, def)
}

func addBoolPFlag(flags *pflag.FlagSet, name, shorthand string, def bool, desc string) {
	flags.BoolP(name, shorthand, def, desc)
	bindPFlag(flags, name)
	viper.SetDefault(name, def)
}
