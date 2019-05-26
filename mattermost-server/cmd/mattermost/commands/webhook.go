// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var WebhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Management of webhooks",
}

var WebhookListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List webhooks",
	Long:    "list all webhooks",
	Example: "  webhook list myteam",
	RunE:    listWebhookCmdF,
}

var WebhookShowCmd = &cobra.Command{
	Use:     "show [webhookId]",
	Short:   "Show a webhook",
	Long:    "Show the webhook specified by [webhookId]",
	Args:    cobra.ExactArgs(1),
	Example: "  webhook show w16zb5tu3n1zkqo18goqry1je",
	RunE:    showWebhookCmdF,
}

var WebhookCreateIncomingCmd = &cobra.Command{
	Use:     "create-incoming",
	Short:   "Create incoming webhook",
	Long:    "create incoming webhook which allows external posting of messages to specific channel",
	Example: "  webhook create-incoming --channel [channelID] --user [userID] --display-name [displayName] --description [webhookDescription] --lock-to-channel --icon [iconURL]",
	RunE:    createIncomingWebhookCmdF,
}

var WebhookModifyIncomingCmd = &cobra.Command{
	Use:     "modify-incoming",
	Short:   "Modify incoming webhook",
	Long:    "Modify existing incoming webhook by changing its title, description, channel or icon url",
	Example: "  webhook modify-incoming [webhookID] --channel [channelID] --display-name [displayName] --description [webhookDescription] --lock-to-channel --icon [iconURL]",
	RunE:    modifyIncomingWebhookCmdF,
}

var WebhookCreateOutgoingCmd = &cobra.Command{
	Use:   "create-outgoing",
	Short: "Create outgoing webhook",
	Long:  "create outgoing webhook which allows external posting of messages from a specific channel",
	Example: `  webhook create-outgoing --team myteam --user myusername --display-name mywebhook --trigger-word "build" --trigger-word "test" --url http://localhost:8000/my-webhook-handler
	webhook create-outgoing --team myteam --channel mychannel --user myusername --display-name mywebhook --description "My cool webhook" --trigger-when start --trigger-word build --trigger-word test --icon http://localhost:8000/my-slash-handler-bot-icon.png --url http://localhost:8000/my-webhook-handler --content-type "application/json"`,
	RunE: createOutgoingWebhookCmdF,
}

var WebhookModifyOutgoingCmd = &cobra.Command{
	Use:     "modify-outgoing",
	Short:   "Modify outgoing webhook",
	Long:    "Modify existing outgoing webhook by changing its title, description, channel, icon, url, content-type, and triggers",
	Example: `  webhook modify-outgoing [webhookId] --channel [channelId] --display-name [displayName] --description "New webhook description" --icon http://localhost:8000/my-slash-handler-bot-icon.png --url http://localhost:8000/my-webhook-handler --content-type "application/json" --trigger-word test --trigger-when start`,
	RunE:    modifyOutgoingWebhookCmdF,
}

var WebhookDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete webhooks",
	Long:    "Delete webhook with given id",
	Example: "  webhook delete [webhookID]",
	RunE:    deleteWebhookCmdF,
}

func listWebhookCmdF(command *cobra.Command, args []string) error {
	app, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	var teams []*model.Team
	if len(args) < 1 {
		var getErr *model.AppError
		// If no team is specified, list all teams
		teams, getErr = app.GetAllTeams()
		if getErr != nil {
			return getErr
		}
	} else {
		teams = getTeamsFromTeamArgs(app, args)
	}

	for i, team := range teams {
		if team == nil {
			CommandPrintErrorln("Unable to find team '" + args[i] + "'")
			continue
		}

		// Fetch all hooks with a very large limit so we get them all.
		incomingResult := make(chan store.StoreResult, 1)
		go func() {
			incomingHooks, err := app.Srv.Store.Webhook().GetIncomingByTeam(team.Id, 0, 100000000)
			incomingResult <- store.StoreResult{Data: incomingHooks, Err: err}
			close(incomingResult)
		}()
		outgoingResult := make(chan store.StoreResult, 1)
		go func() {
			outgoingHooks, err := app.Srv.Store.Webhook().GetOutgoingByTeam(team.Id, 0, 100000000)
			outgoingResult <- store.StoreResult{Data: outgoingHooks, Err: err}
			close(outgoingResult)
		}()

		if result := <-incomingResult; result.Err == nil {
			CommandPrettyPrintln(fmt.Sprintf("Incoming webhooks for %s (%s):", team.DisplayName, team.Name))
			hooks := result.Data.([]*model.IncomingWebhook)
			for _, hook := range hooks {
				CommandPrettyPrintln("\t" + hook.DisplayName + " (" + hook.Id + ")")
			}
		} else {
			CommandPrintErrorln("Unable to list incoming webhooks for '" + args[i] + "'")
		}

		if result := <-outgoingResult; result.Err == nil {
			hooks := result.Data.([]*model.OutgoingWebhook)
			CommandPrettyPrintln(fmt.Sprintf("Outgoing webhooks for %s (%s):", team.DisplayName, team.Name))
			for _, hook := range hooks {
				CommandPrettyPrintln("\t" + hook.DisplayName + " (" + hook.Id + ")")
			}
		} else {
			CommandPrintErrorln("Unable to list outgoing webhooks for '" + args[i] + "'")
		}
	}
	return nil
}

func createIncomingWebhookCmdF(command *cobra.Command, args []string) error {
	app, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	channelArg, errChannel := command.Flags().GetString("channel")
	if errChannel != nil || channelArg == "" {
		return errors.New("Channel is required")
	}
	channel := getChannelFromChannelArg(app, channelArg)
	if channel == nil {
		return errors.New("Unable to find channel '" + channelArg + "'")
	}

	userArg, errUser := command.Flags().GetString("user")
	if errUser != nil || userArg == "" {
		return errors.New("User is required")
	}
	user := getUserFromUserArg(app, userArg)
	if user == nil {
		return errors.New("Unable to find user '" + userArg + "'")
	}

	displayName, _ := command.Flags().GetString("display-name")
	description, _ := command.Flags().GetString("description")
	iconURL, _ := command.Flags().GetString("icon")
	channelLocked, _ := command.Flags().GetBool("lock-to-channel")

	incomingWebhook := &model.IncomingWebhook{
		ChannelId:     channel.Id,
		DisplayName:   displayName,
		Description:   description,
		IconURL:       iconURL,
		ChannelLocked: channelLocked,
	}

	createdIncoming, errIncomingWebhook := app.CreateIncomingWebhookForChannel(user.Id, channel, incomingWebhook)
	if errIncomingWebhook != nil {
		return errIncomingWebhook
	}

	CommandPrettyPrintln("Id: " + createdIncoming.Id)
	CommandPrettyPrintln("Display Name: " + createdIncoming.DisplayName)

	return nil
}

func modifyIncomingWebhookCmdF(command *cobra.Command, args []string) error {
	app, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	if len(args) < 1 {
		return errors.New("WebhookID is not specified")
	}

	webhookArg := args[0]
	oldHook, getErr := app.GetIncomingWebhook(webhookArg)
	if getErr != nil {
		return errors.New("Unable to find webhook '" + webhookArg + "'")
	}

	updatedHook := oldHook

	channelArg, _ := command.Flags().GetString("channel")
	if channelArg != "" {
		channel := getChannelFromChannelArg(app, channelArg)
		if channel == nil {
			return errors.New("Unable to find channel '" + channelArg + "'")
		}
		updatedHook.ChannelId = channel.Id
	}

	displayName, _ := command.Flags().GetString("display-name")
	if displayName != "" {
		updatedHook.DisplayName = displayName
	}
	description, _ := command.Flags().GetString("description")
	if description != "" {
		updatedHook.Description = description
	}
	iconUrl, _ := command.Flags().GetString("icon")
	if iconUrl != "" {
		updatedHook.IconURL = iconUrl
	}
	channelLocked, _ := command.Flags().GetBool("lock-to-channel")
	updatedHook.ChannelLocked = channelLocked

	if _, err := app.UpdateIncomingWebhook(oldHook, updatedHook); err != nil {
		return err
	}

	return nil
}

func createOutgoingWebhookCmdF(command *cobra.Command, args []string) error {
	app, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	teamArg, errTeam := command.Flags().GetString("team")
	if errTeam != nil || teamArg == "" {
		return errors.New("Team is required")
	}
	team := getTeamFromTeamArg(app, teamArg)
	if team == nil {
		return errors.New("Unable to find team: " + teamArg)
	}

	userArg, errUser := command.Flags().GetString("user")
	if errUser != nil || userArg == "" {
		return errors.New("User is required")
	}
	user := getUserFromUserArg(app, userArg)
	if user == nil {
		return errors.New("Unable to find user: " + userArg)
	}

	displayName, errName := command.Flags().GetString("display-name")
	if errName != nil || displayName == "" {
		return errors.New("Display name is required")
	}

	triggerWords, errWords := command.Flags().GetStringArray("trigger-word")
	if errWords != nil || len(triggerWords) == 0 {
		return errors.New("Trigger word or words required")
	}

	callbackURLs, errURL := command.Flags().GetStringArray("url")
	if errURL != nil || len(callbackURLs) == 0 {
		return errors.New("Callback URL or URLs required")
	}

	triggerWhenString, _ := command.Flags().GetString("trigger-when")
	var triggerWhen int
	if triggerWhenString == "exact" {
		triggerWhen = 0
	} else if triggerWhenString == "start" {
		triggerWhen = 1
	} else {
		return errors.New("Invalid trigger when parameter")
	}
	description, _ := command.Flags().GetString("description")
	contentType, _ := command.Flags().GetString("content-type")
	iconURL, _ := command.Flags().GetString("icon")

	outgoingWebhook := &model.OutgoingWebhook{
		CreatorId:    user.Id,
		Username:     user.Username,
		TeamId:       team.Id,
		TriggerWords: triggerWords,
		TriggerWhen:  triggerWhen,
		CallbackURLs: callbackURLs,
		DisplayName:  displayName,
		Description:  description,
		ContentType:  contentType,
		IconURL:      iconURL,
	}

	channelArg, _ := command.Flags().GetString("channel")
	if channelArg != "" {
		channel := getChannelFromChannelArg(app, channelArg)
		if channel != nil {
			outgoingWebhook.ChannelId = channel.Id
		}
	}

	createdOutgoing, errOutgoing := app.CreateOutgoingWebhook(outgoingWebhook)
	if errOutgoing != nil {
		return errOutgoing
	}

	CommandPrettyPrintln("Id: " + createdOutgoing.Id)
	CommandPrettyPrintln("Display Name: " + createdOutgoing.DisplayName)

	return nil
}

func modifyOutgoingWebhookCmdF(command *cobra.Command, args []string) error {
	app, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	if len(args) < 1 {
		return errors.New("WebhookID is not specified")
	}

	webhookArg := args[0]
	oldHook, appErr := app.GetOutgoingWebhook(webhookArg)
	if appErr != nil {
		return fmt.Errorf("unable to find webhook '%s'", webhookArg)
	}

	updatedHook := model.OutgoingWebhookFromJson(strings.NewReader(oldHook.ToJson()))

	channelArg, _ := command.Flags().GetString("channel")
	if channelArg != "" {
		channel := getChannelFromChannelArg(app, channelArg)
		if channel == nil {
			return fmt.Errorf("unable to find channel '%s'", channelArg)
		}
		updatedHook.ChannelId = channel.Id
	}

	displayName, _ := command.Flags().GetString("display-name")
	if displayName != "" {
		updatedHook.DisplayName = displayName
	}

	description, _ := command.Flags().GetString("description")
	if description != "" {
		updatedHook.Description = description
	}

	triggerWords, err := command.Flags().GetStringArray("trigger-word")
	if err != nil {
		return errors.Wrap(err, "invalid trigger-word parameter")
	}
	if len(triggerWords) > 0 {
		updatedHook.TriggerWords = triggerWords
	}

	triggerWhenString, _ := command.Flags().GetString("trigger-when")
	if triggerWhenString != "" {
		var triggerWhen int
		if triggerWhenString == "exact" {
			triggerWhen = 0
		} else if triggerWhenString == "start" {
			triggerWhen = 1
		} else {
			return errors.New("invalid trigger-when parameter")
		}
		updatedHook.TriggerWhen = triggerWhen
	}

	iconURL, _ := command.Flags().GetString("icon")
	if iconURL != "" {
		updatedHook.IconURL = iconURL
	}

	contentType, _ := command.Flags().GetString("content-type")
	if contentType != "" {
		updatedHook.ContentType = contentType
	}

	callbackURLs, err := command.Flags().GetStringArray("url")
	if err != nil {
		return errors.Wrap(err, "invalid URL parameter")
	}
	if len(callbackURLs) > 0 {
		updatedHook.CallbackURLs = callbackURLs
	}

	if _, appErr := app.UpdateOutgoingWebhook(oldHook, updatedHook); appErr != nil {
		return appErr
	}

	return nil
}

func deleteWebhookCmdF(command *cobra.Command, args []string) error {
	app, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	if len(args) < 1 {
		return errors.New("WebhookID is not specified")
	}

	webhookId := args[0]
	errIncomingWebhook := app.DeleteIncomingWebhook(webhookId)
	errOutgoingWebhook := app.DeleteOutgoingWebhook(webhookId)

	if errIncomingWebhook != nil && errOutgoingWebhook != nil {
		return errors.New("Unable to delete webhook '" + webhookId + "'")
	}

	return nil
}

func showWebhookCmdF(command *cobra.Command, args []string) error {
	app, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	webhookId := args[0]
	if incomingWebhook, err := app.GetIncomingWebhook(webhookId); err == nil {
		fmt.Printf("%s", prettyPrintStruct(*incomingWebhook))
		return nil
	}
	if outgoingWebhook, err := app.GetOutgoingWebhook(webhookId); err == nil {
		fmt.Printf("%s", prettyPrintStruct(*outgoingWebhook))
		return nil
	}

	return errors.New("Webhook with id " + webhookId + " not found")
}

func init() {
	WebhookCreateIncomingCmd.Flags().String("channel", "", "Channel ID (required)")
	WebhookCreateIncomingCmd.Flags().String("user", "", "User ID (required)")
	WebhookCreateIncomingCmd.Flags().String("display-name", "", "Incoming webhook display name")
	WebhookCreateIncomingCmd.Flags().String("description", "", "Incoming webhook description")
	WebhookCreateIncomingCmd.Flags().String("icon", "", "Icon URL")
	WebhookCreateIncomingCmd.Flags().Bool("lock-to-channel", false, "Lock to channel")

	WebhookModifyIncomingCmd.Flags().String("channel", "", "Channel ID")
	WebhookModifyIncomingCmd.Flags().String("display-name", "", "Incoming webhook display name")
	WebhookModifyIncomingCmd.Flags().String("description", "", "Incoming webhook description")
	WebhookModifyIncomingCmd.Flags().String("icon", "", "Icon URL")
	WebhookModifyIncomingCmd.Flags().Bool("lock-to-channel", false, "Lock to channel")

	WebhookCreateOutgoingCmd.Flags().String("team", "", "Team name or ID (required)")
	WebhookCreateOutgoingCmd.Flags().String("channel", "", "Channel name or ID")
	WebhookCreateOutgoingCmd.Flags().String("user", "", "User username, email, or ID (required)")
	WebhookCreateOutgoingCmd.Flags().String("display-name", "", "Outgoing webhook display name (required)")
	WebhookCreateOutgoingCmd.Flags().String("description", "", "Outgoing webhook description")
	WebhookCreateOutgoingCmd.Flags().StringArray("trigger-word", []string{}, "Word to trigger webhook (required)")
	WebhookCreateOutgoingCmd.Flags().String("trigger-when", "exact", "When to trigger webhook (exact: for first word matches a trigger word exactly, start: for first word starts with a trigger word)")
	WebhookCreateOutgoingCmd.Flags().String("icon", "", "Icon URL")
	WebhookCreateOutgoingCmd.Flags().StringArray("url", []string{}, "Callback URL (required)")
	WebhookCreateOutgoingCmd.Flags().String("content-type", "", "Content-type")

	WebhookModifyOutgoingCmd.Flags().String("channel", "", "Channel name or ID")
	WebhookModifyOutgoingCmd.Flags().String("display-name", "", "Outgoing webhook display name")
	WebhookModifyOutgoingCmd.Flags().String("description", "", "Outgoing webhook description")
	WebhookModifyOutgoingCmd.Flags().StringArray("trigger-word", []string{}, "Word to trigger webhook")
	WebhookModifyOutgoingCmd.Flags().String("trigger-when", "", "When to trigger webhook (exact: for first word matches a trigger word exactly, start: for first word starts with a trigger word)")
	WebhookModifyOutgoingCmd.Flags().String("icon", "", "Icon URL")
	WebhookModifyOutgoingCmd.Flags().StringArray("url", []string{}, "Callback URL")
	WebhookModifyOutgoingCmd.Flags().String("content-type", "", "Content-type")

	WebhookCmd.AddCommand(
		WebhookListCmd,
		WebhookCreateIncomingCmd,
		WebhookModifyIncomingCmd,
		WebhookCreateOutgoingCmd,
		WebhookModifyOutgoingCmd,
		WebhookDeleteCmd,
		WebhookShowCmd,
	)

	RootCmd.AddCommand(WebhookCmd)
}
