package cmd

import (
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/pymba86/bingo/pkg/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var userConfig *engine.Config

var RootCmd = &cobra.Command{
	Use:          "bingo",
	Short:        "bingo is a crypto trading bot",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		disableDotEnv, err := cmd.Flags().GetBool("no-dotenv")
		if err != nil {
			return err
		}

		if !disableDotEnv {
			dotenvFile, err := cmd.Flags().GetString("dotenv")
			if err != nil {
				return err
			}

			if _, err := os.Stat(dotenvFile); err == nil {
				if err := godotenv.Load(dotenvFile); err != nil {
					return errors.Wrap(err, "error loading dotenv file")
				}
			}
		}

		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return errors.Wrapf(err, "failed to get the config flag")
		}

		if len(configFile) > 0 {
			// if config file exists, use the config loaded from the config file.
			// otherwise, use a empty config object
			if _, err := os.Stat(configFile); err == nil {
				// load successfully
				userConfig, err = engine.Load(configFile, false)
				if err != nil {
					return errors.Wrapf(err, "can not load config file: %s", configFile)
				}

			} else if os.IsNotExist(err) {
				// config file doesn't exist, we should use the empty config
				userConfig = &engine.Config{}
			} else {
				// other error
				return errors.Wrapf(err, "config file load error: %s", configFile)
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Execute() error {

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		log.WithError(err).Errorf("failed to bind persistent flags. please check the flag settings.")
		return err
	}

	if err := viper.BindPFlags(RootCmd.Flags()); err != nil {
		log.WithError(err).Errorf("failed to bind local flags. please check the flag settings.")
		return err
	}

	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().Bool("debug", false, "debug flag")
	RootCmd.PersistentFlags().String("config", "bingo.yaml", "config file")

	RootCmd.PersistentFlags().Bool("no-dotenv", false, "disable built-in dotenv")
	RootCmd.PersistentFlags().String("dotenv", ".env.local", "the dotenv file you want to load")

	RootCmd.PersistentFlags().String("telegram-bot-token", "", "telegram bot token from bot father")
	RootCmd.PersistentFlags().String("telegram-bot-auth-token", "", "telegram auth token")

	RootCmd.PersistentFlags().String("binance-api-key", "", "binance api key")
	RootCmd.PersistentFlags().String("binance-api-secret", "", "binance api secret")
}