package ioc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/rumorshub/ioc/configwise"
	"github.com/rumorshub/ioc/logwise"
)

const envDotenv = "DOTENV_PATH"

func NewCommand(args []string, short, envPrefix, version string) *cobra.Command {
	var (
		cfgFile  string
		dotenv   string
		override []string
	)

	cmd := &cobra.Command{
		Use:           filepath.Base(args[0]),
		Short:         short,
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Version = version

			if absPath, err := filepath.Abs(cfgFile); err == nil {
				cfgFile = absPath
			}

			if v, ok := os.LookupEnv(envDotenv); ok {
				dotenv = v
			}

			if _, err := os.Stat(dotenv); err == nil {
				if err = godotenv.Load(dotenv); err != nil {
					return err
				}
			}

			cfg, err := configwise.NewConfigurer(
				version,
				configwise.WithPath(cfgFile),
				configwise.WithPrefix(envPrefix),
				configwise.WithFlags(override),
			)
			if err != nil {
				return err
			}

			log, err := logwise.Load(cfg)
			if err != nil {
				return err
			}

			cont := NewContainer(cfg, log)

			cmd.SetContext(WithContainer(cmd.Context(), cont))

			return nil
		},
	}

	f := cmd.PersistentFlags()
	f.StringVarP(&cfgFile, "config", "c", "config.yaml", "config file")
	f.StringVar(&dotenv, "dotenv", ".env", fmt.Sprintf("dotenv file [$%s]", envDotenv))
	f.StringArrayVarP(&override, "override", "o", nil, "override config value (dot.notation=value)")

	_ = f.Parse(args[1:])

	return cmd
}
