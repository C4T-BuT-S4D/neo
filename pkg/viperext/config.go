package viperext

import (
	"fmt"
	"strings"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func NewViper(envPrefix string) (*viper.Viper, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(
		strings.NewReplacer(
			"-", "_",
			".", "_",
		),
	)

	return v, nil
}

// BindPFlags binds all flags in the given FlagSet to viper with normalized names
// (kebab-case to snake_case).
func BindPFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.VisitAll(func(f *pflag.Flag) {
		normName := strings.ReplaceAll(f.Name, "-", "_")
		if err := v.BindPFlag(normName, f); err != nil {
			panic(fmt.Sprintf("binding flag %q: %v", f.Name, err))
		}
	})
}

func GetConfig[T any](v *viper.Viper, cfg *T) (*T, error) {
	if err := defaults.Set(cfg); err != nil {
		return cfg, fmt.Errorf("setting defaults: %w", err)
	}

	if err := v.Unmarshal(
		cfg,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.TextUnmarshallerHookFunc(),
				mapstructure.StringToTimeDurationHookFunc(),
			),
		),
	); err != nil {
		return cfg, fmt.Errorf("unmarshaling config: %w", err)
	}

	return cfg, nil
}
