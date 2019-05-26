// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattermost/mattermost-server/api4"
	"github.com/mattermost/mattermost-server/app"
	"github.com/mattermost/mattermost-server/config"
	"github.com/mattermost/mattermost-server/manualtesting"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/utils"
	"github.com/mattermost/mattermost-server/web"
	"github.com/mattermost/mattermost-server/wsapi"
	"github.com/mattermost/viper"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:          "server",
	Short:        "Run the Mattermost server",
	RunE:         serverCmdF,
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(serverCmd)
	RootCmd.RunE = serverCmdF
}

func serverCmdF(command *cobra.Command, args []string) error {
	configDSN := viper.GetString("config")

	disableConfigWatch, _ := command.Flags().GetBool("disableconfigwatch")
	usedPlatform, _ := command.Flags().GetBool("platform")

	interruptChan := make(chan os.Signal, 1)

	if err := utils.TranslationsPreInit(); err != nil {
		return errors.Wrapf(err, "unable to load Mattermost translation files")
	}
	configStore, err := config.NewStore(configDSN, !disableConfigWatch)
	if err != nil {
		return err
	}

	return runServer(configStore, disableConfigWatch, usedPlatform, interruptChan)
}

func runServer(configStore config.Store, disableConfigWatch bool, usedPlatform bool, interruptChan chan os.Signal) error {
	options := []app.Option{
		app.ConfigStore(configStore),
		app.RunJobs,
		app.JoinCluster,
		app.StartElasticsearch,
		app.StartMetrics,
	}
	server, err := app.NewServer(options...)
	if err != nil {
		mlog.Critical(err.Error())
		return err
	}
	defer server.Shutdown()

	if usedPlatform {
		mlog.Error("The platform binary has been deprecated, please switch to using the mattermost binary.")
	}

	serverErr := server.Start()
	if serverErr != nil {
		mlog.Critical(serverErr.Error())
		return serverErr
	}

	api := api4.Init(server, server.AppOptions, server.Router)
	wsapi.Init(server.FakeApp(), server.WebSocketRouter)
	web.New(server, server.AppOptions, server.Router)

	// If we allow testing then listen for manual testing URL hits
	if *server.Config().ServiceSettings.EnableTesting {
		manualtesting.Init(api)
	}

	notifyReady()

	// wait for kill signal before attempting to gracefully shutdown
	// the running service
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

	return nil
}

func notifyReady() {
	// If the environment vars provide a systemd notification socket,
	// notify systemd that the server is ready.
	systemdSocket := os.Getenv("NOTIFY_SOCKET")
	if systemdSocket != "" {
		mlog.Info("Sending systemd READY notification.")

		err := sendSystemdReadyNotification(systemdSocket)
		if err != nil {
			mlog.Error(err.Error())
		}
	}
}

func sendSystemdReadyNotification(socketPath string) error {
	msg := "READY=1"
	addr := &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg))
	return err
}
