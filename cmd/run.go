package cmd

import (
	"context"
	"github.com/pkg/errors"
	"github.com/pymba86/bingo/pkg/engine"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	RunCmd.Flags().Bool("setup", false, "use setup mode")
	RootCmd.AddCommand(RunCmd)
}

var RunCmd = &cobra.Command{
	Use:          "run",
	Short:        "run strategies from config file",
	SilenceUsage: true,
	RunE:         run,
}

func runConfig(basectx context.Context, userConfig *engine.Config,
	enableWebServer bool, webServerBind string) error {

	return nil
}

func runSetup(baseCtx context.Context, userConfig *engine.Config, enableApiServer bool) error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	setup, err := cmd.Flags().GetBool("setup")
	if err != nil {
		return err
	}

	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	enableWebServer, err := cmd.Flags().GetBool("enable-webserver")
	if err != nil {
		return err
	}

	webServerBind, err := cmd.Flags().GetString("webserver-bind")
	if err != nil {
		return err
	}

	var userConfig = &engine.Config{}

	if !setup {
		// if it's not setup, then the config file option is required.
		if len(configFile) == 0 {
			return errors.New("--config option is required")
		}

		if _, err := os.Stat(configFile); err != nil {
			return err
		}

		userConfig, err = engine.Load(configFile, false)
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if setup {
		return runSetup(ctx, userConfig, true)
	}

	userConfig, err = engine.Load(configFile, true)
	if err != nil {
		return err
	}

	return runConfig(ctx, userConfig, enableWebServer, webServerBind)
}
