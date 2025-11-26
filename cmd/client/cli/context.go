package cli

import "github.com/spf13/viper"

type Context struct {
	Viper *viper.Viper
}

func NewContext(v *viper.Viper) *Context {
	return &Context{
		Viper: v,
	}
}
