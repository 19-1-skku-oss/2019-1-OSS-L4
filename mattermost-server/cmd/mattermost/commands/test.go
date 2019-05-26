// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"os/signal"
	"syscall"

	"github.com/mattermost/mattermost-server/api4"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils"
	"github.com/mattermost/mattermost-server/wsapi"
	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{
	Use:    "test",
	Short:  "Testing Commands",
	Hidden: true,
}

var RunWebClientTestsCmd = &cobra.Command{
	Use:   "web_client_tests",
	Short: "Run the web client tests",
	RunE:  webClientTestsCmdF,
}

var RunServerForWebClientTestsCmd = &cobra.Command{
	Use:   "web_client_tests_server",
	Short: "Run the server configured for running the web client tests against it",
	RunE:  serverForWebClientTestsCmdF,
}

func init() {
	TestCmd.AddCommand(
		RunWebClientTestsCmd,
		RunServerForWebClientTestsCmd,
	)
	RootCmd.AddCommand(TestCmd)
}

func webClientTestsCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	utils.InitTranslations(a.Config().LocalizationSettings)
	serverErr := a.Srv.Start()
	if serverErr != nil {
		return serverErr
	}

	api4.Init(a, a.Srv.AppOptions, a.Srv.Router)
	wsapi.Init(a, a.Srv.WebSocketRouter)
	a.UpdateConfig(setupClientTests)
	runWebClientTests()

	return nil
}

func serverForWebClientTestsCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	utils.InitTranslations(a.Config().LocalizationSettings)
	serverErr := a.Srv.Start()
	if serverErr != nil {
		return serverErr
	}

	api4.Init(a, a.Srv.AppOptions, a.Srv.Router)
	wsapi.Init(a, a.Srv.WebSocketRouter)
	a.UpdateConfig(setupClientTests)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c

	return nil
}

func setupClientTests(cfg *model.Config) {
	*cfg.TeamSettings.EnableOpenServer = true
	*cfg.ServiceSettings.EnableCommands = false
	*cfg.ServiceSettings.EnableCustomEmoji = true
	*cfg.ServiceSettings.EnableIncomingWebhooks = false
	*cfg.ServiceSettings.EnableOutgoingWebhooks = false
}

func executeTestCommand(command *exec.Cmd) {
	cmdOutPipe, err := command.StdoutPipe()
	if err != nil {
		CommandPrintErrorln("Failed to run tests")
		os.Exit(1)
		return
	}

	cmdErrOutPipe, err := command.StderrPipe()
	if err != nil {
		CommandPrintErrorln("Failed to run tests")
		os.Exit(1)
		return
	}

	cmdOutReader := bufio.NewScanner(cmdOutPipe)
	cmdErrOutReader := bufio.NewScanner(cmdErrOutPipe)
	go func() {
		for cmdOutReader.Scan() {
			fmt.Println(cmdOutReader.Text())
		}
	}()

	go func() {
		for cmdErrOutReader.Scan() {
			fmt.Println(cmdErrOutReader.Text())
		}
	}()

	if err := command.Run(); err != nil {
		CommandPrintErrorln("Client Tests failed")
		os.Exit(1)
		return
	}
}

func runWebClientTests() {
	if webappDir := os.Getenv("WEBAPP_DIR"); webappDir != "" {
		os.Chdir(webappDir)
	} else {
		os.Chdir("../mattermost-webapp")
	}

	cmd := exec.Command("npm", "test")
	executeTestCommand(cmd)
}
