package viperext

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func MustBindCommandFlags(v *viper.Viper, cmd *cobra.Command) {
	cmd.PersistentPreRun = func(*cobra.Command, []string) {
		RunBindCommandFlags(v, cmd)
	}
}

func RunBindCommandFlags(v *viper.Viper, cmd *cobra.Command) {
	zap.L().Debug("binding command flags", zap.String("command", cmd.Name()))

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		normName := strings.ReplaceAll(f.Name, "-", "_")
		if err := v.BindPFlag(normName, f); err != nil {
			zap.L().Error("binding flag", zap.String("flag", f.Name), zap.Error(err))
		}
	})

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		normName := strings.ReplaceAll(f.Name, "-", "_")
		if err := v.BindPFlag(normName, f); err != nil {
			zap.L().Error("binding persistent flag", zap.String("flag", f.Name), zap.Error(err))
		}
	})
}
