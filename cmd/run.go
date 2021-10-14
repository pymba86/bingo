package cmd

import (
	"context"
	"github.com/pkg/errors"
	"github.com/pymba86/bingo/pkg/cmdutil"
	"github.com/pymba86/bingo/pkg/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"syscall"
	"time"
)

func init() {
	RunCmd.Flags().Bool("setup", false, "use setup mode")
	RunCmd.Flags().Bool("enable-webserver", false, "enable webserver")
	RunCmd.Flags().String("webserver-bind", ":8080", "webserver binding")
	RootCmd.AddCommand(RunCmd)
}

var RunCmd = &cobra.Command{
	Use:          "run",
	Short:        "run strategies from config file",
	SilenceUsage: true,
	RunE:         run,
}

func BootstrapEnvironment(ctx context.Context, environ *engine.Environment, userConfig *engine.Config) error {

	if err := environ.ConfigureExchangeSessions(userConfig); err != nil {
		return errors.Wrap(err, "exchange session configure error")
	}
	return nil
}

func runConfig(basectx context.Context, userConfig *engine.Config,
	enableWebServer bool, webServerBind string) error {

	ctx, cancelTrading := context.WithCancel(basectx)
	defer cancelTrading()

	environ := engine.NewEnvironment()

	if err := BootstrapEnvironment(ctx, environ, userConfig); err != nil {
		return err
	}

	if err := environ.Init(ctx); err != nil {
		return err
	}

	trader := engine.NewTrader(environ)
	if err := trader.Configure(userConfig); err != nil {
		return err
	}

	if err := trader.Run(ctx); err != nil {
		return err
	}

	cmdutil.WaitForSignal(ctx, syscall.SIGINT, syscall.SIGTERM)

	log.Infof("shutting down stratgies...")
	shutdownCtx, cancelShutdown := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	trader.Graceful.Shutdown(shutdownCtx)
	cancelShutdown()
	cancelTrading()

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
