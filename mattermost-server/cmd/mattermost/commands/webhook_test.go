// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/api4"
	"github.com/mattermost/mattermost-server/model"
)

func TestListWebhooks(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	adminClient := th.SystemAdminClient

	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	*config.ServiceSettings.EnableIncomingWebhooks = true
	*config.ServiceSettings.EnableOutgoingWebhooks = true
	*config.ServiceSettings.EnablePostUsernameOverride = true
	*config.ServiceSettings.EnablePostIconOverride = true
	th.SetConfig(config)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableIncomingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOutgoingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostUsernameOverride = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostIconOverride = true })

	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()
	th.AddPermissionToRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)

	dispName := "myhookinc"
	hook := &model.IncomingWebhook{DisplayName: dispName, ChannelId: th.BasicChannel.Id, TeamId: th.BasicChannel.TeamId}
	_, resp := adminClient.CreateIncomingWebhook(hook)
	api4.CheckNoError(t, resp)

	dispName2 := "myhookout"
	outHook := &model.OutgoingWebhook{DisplayName: dispName2, ChannelId: th.BasicChannel.Id, TeamId: th.BasicChannel.TeamId, CallbackURLs: []string{"http://nowhere.com"}, Username: "some-user-name", IconURL: "http://some-icon-url/"}
	_, resp = adminClient.CreateOutgoingWebhook(outHook)
	api4.CheckNoError(t, resp)

	output := th.CheckCommand(t, "webhook", "list", th.BasicTeam.Name)

	if !strings.Contains(string(output), dispName) {
		t.Fatal("should have incoming webhooks")
	}

	if !strings.Contains(string(output), dispName2) {
		t.Fatal("should have outgoing webhooks")
	}

}

func TestShowWebhook(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	adminClient := th.SystemAdminClient

	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	*config.ServiceSettings.EnableIncomingWebhooks = true
	*config.ServiceSettings.EnableOutgoingWebhooks = true
	*config.ServiceSettings.EnablePostUsernameOverride = true
	*config.ServiceSettings.EnablePostIconOverride = true
	th.SetConfig(config)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableIncomingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOutgoingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostUsernameOverride = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostIconOverride = true })

	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()
	th.AddPermissionToRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)

	dispName := "incominghook"
	hook := &model.IncomingWebhook{
		DisplayName: dispName,
		ChannelId:   th.BasicChannel.Id,
		TeamId:      th.BasicChannel.TeamId,
	}
	incomingWebhook, resp := adminClient.CreateIncomingWebhook(hook)
	api4.CheckNoError(t, resp)

	// should return an error when no webhookid is provided
	require.Error(t, th.RunCommand(t, "webhook", "show"))

	// invalid webhook should return error
	require.Error(t, th.RunCommand(t, "webhook", "show", "invalid-webhook"))

	// valid incoming webhook should return webhook data
	output := th.CheckCommand(t, "webhook", "show", incomingWebhook.Id)
	if !strings.Contains(string(output), "DisplayName: \""+dispName+"\"") {
		t.Fatal("incoming: should have incominghook as displayname")
	}
	if !strings.Contains(string(output), "ChannelId: \""+hook.ChannelId+"\"") {
		t.Fatal("incoming: should have a valid channelId")
	}

	dispName = "outgoinghook"
	outgoingHook := &model.OutgoingWebhook{
		DisplayName:  dispName,
		ChannelId:    th.BasicChannel.Id,
		TeamId:       th.BasicChannel.TeamId,
		CallbackURLs: []string{"http://nowhere.com"},
		Username:     "some-user-name",
		IconURL:      "http://some-icon-url/",
	}
	outgoingWebhook, resp := adminClient.CreateOutgoingWebhook(outgoingHook)
	api4.CheckNoError(t, resp)

	// valid outgoing webhook should return webhook data
	output = th.CheckCommand(t, "webhook", "show", outgoingWebhook.Id)
	if !strings.Contains(string(output), "DisplayName: \""+dispName+"\"") {
		t.Fatal("outgoing: should have outgoinghook as displayname")
	}
	if !strings.Contains(string(output), "ChannelId: \""+hook.ChannelId+"\"") {
		t.Fatal("outgoing: should have a valid channelId")
	}

}

func TestCreateIncomingWebhook(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	*config.ServiceSettings.EnableIncomingWebhooks = true
	*config.ServiceSettings.EnableOutgoingWebhooks = true
	*config.ServiceSettings.EnablePostUsernameOverride = true
	*config.ServiceSettings.EnablePostIconOverride = true
	th.SetConfig(config)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableIncomingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOutgoingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostUsernameOverride = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostIconOverride = true })

	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()
	th.AddPermissionToRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)

	// should fail because you need to specify valid channel
	require.Error(t, th.RunCommand(t, "webhook", "create-incoming"))
	require.Error(t, th.RunCommand(t, "webhook", "create-incoming", "--channel", th.BasicTeam.Name+":doesnotexist"))

	// should fail because you need to specify valid user
	require.Error(t, th.RunCommand(t, "webhook", "create-incoming", "--channel", th.BasicChannel.Id))
	require.Error(t, th.RunCommand(t, "webhook", "create-incoming", "--channel", th.BasicChannel.Id, "--user", "doesnotexist"))

	description := "myhookinc"
	displayName := "myhookinc"
	th.CheckCommand(t, "webhook", "create-incoming", "--channel", th.BasicChannel.Id, "--user", th.BasicUser.Email, "--description", description, "--display-name", displayName)

	webhooks, err := th.App.GetIncomingWebhooksPage(0, 1000)
	if err != nil {
		t.Fatal("unable to retrieve incoming webhooks")
	}

	found := false
	for _, webhook := range webhooks {
		if webhook.Description == description && webhook.UserId == th.BasicUser.Id {
			found = true
		}
	}
	if !found {
		t.Fatal("Failed to create incoming webhook")
	}
}

func TestModifyIncomingWebhook(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	*config.ServiceSettings.EnableIncomingWebhooks = true
	*config.ServiceSettings.EnableOutgoingWebhooks = true
	*config.ServiceSettings.EnablePostUsernameOverride = true
	*config.ServiceSettings.EnablePostIconOverride = true
	th.SetConfig(config)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableIncomingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOutgoingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostUsernameOverride = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostIconOverride = true })

	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()
	th.AddPermissionToRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)

	description := "myhookincdesc"
	displayName := "myhookincname"

	incomingWebhook := &model.IncomingWebhook{
		ChannelId:   th.BasicChannel.Id,
		DisplayName: displayName,
		Description: description,
	}

	oldHook, err := th.App.CreateIncomingWebhookForChannel(th.BasicUser.Id, th.BasicChannel, incomingWebhook)
	if err != nil {
		t.Fatal("unable to create incoming webhooks")
	}
	defer func() {
		th.App.DeleteIncomingWebhook(oldHook.Id)
	}()

	// should fail because you need to specify valid incoming webhook
	require.Error(t, th.RunCommand(t, "webhook", "modify-incoming", "doesnotexist"))
	// should fail because you need to specify valid channel
	require.Error(t, th.RunCommand(t, "webhook", "modify-incoming", oldHook.Id, "--channel", th.BasicTeam.Name+":doesnotexist"))

	modifiedDescription := "myhookincdesc2"
	modifiedDisplayName := "myhookincname2"
	modifiedIconUrl := "myhookincicon2"
	modifiedChannelLocked := true
	modifiedChannelId := th.BasicChannel2.Id

	th.CheckCommand(t, "webhook", "modify-incoming", oldHook.Id, "--channel", modifiedChannelId, "--description", modifiedDescription, "--display-name", modifiedDisplayName, "--icon", modifiedIconUrl, "--lock-to-channel", strconv.FormatBool(modifiedChannelLocked))

	modifiedHook, err := th.App.GetIncomingWebhook(oldHook.Id)
	if err != nil {
		t.Fatal("unable to retrieve modified incoming webhook")
	}
	if modifiedHook.DisplayName != modifiedDisplayName || modifiedHook.Description != modifiedDescription || modifiedHook.IconURL != modifiedIconUrl || modifiedHook.ChannelLocked != modifiedChannelLocked || modifiedHook.ChannelId != modifiedChannelId {
		t.Fatal("Failed to update incoming webhook")
	}
}

func TestCreateOutgoingWebhook(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	*config.ServiceSettings.EnableIncomingWebhooks = true
	*config.ServiceSettings.EnableOutgoingWebhooks = true
	*config.ServiceSettings.EnablePostUsernameOverride = true
	*config.ServiceSettings.EnablePostIconOverride = true
	th.SetConfig(config)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableIncomingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOutgoingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostUsernameOverride = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostIconOverride = true })

	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()
	th.AddPermissionToRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)

	// team, user, display name, trigger words, callback urls are required
	team := th.BasicTeam.Id
	user := th.BasicUser.Id
	displayName := "totally radical webhook"
	triggerWord1 := "build"
	triggerWord2 := "defenestrate"
	callbackURL1 := "http://localhost:8000/my-webhook-handler"
	callbackURL2 := "http://localhost:8000/my-webhook-handler2"

	// should fail because team is not specified
	require.Error(t, th.RunCommand(t, "webhook", "create-outgoing", "--display-name", displayName, "--trigger-word", triggerWord1, "--trigger-word", triggerWord2, "--url", callbackURL1, "--url", callbackURL2, "--user", user))

	// should fail because user is not specified
	require.Error(t, th.RunCommand(t, "webhook", "create-outgoing", "--team", team, "--display-name", displayName, "--trigger-word", triggerWord1, "--trigger-word", triggerWord2, "--url", callbackURL1, "--url", callbackURL2))

	// should fail because display name is not specified
	require.Error(t, th.RunCommand(t, "webhook", "create-outgoing", "--team", team, "--trigger-word", triggerWord1, "--trigger-word", triggerWord2, "--url", callbackURL1, "--url", callbackURL2, "--user", user))

	// should fail because trigger words are not specified
	require.Error(t, th.RunCommand(t, "webhook", "create-outgoing", "--team", team, "--display-name", displayName, "--url", callbackURL1, "--url", callbackURL2, "--user", user))

	// should fail because callback URLs are not specified
	require.Error(t, th.RunCommand(t, "webhook", "create-outgoing", "--team", team, "--display-name", displayName, "--trigger-word", triggerWord1, "--trigger-word", triggerWord2, "--user", user))

	// should fail because outgoing webhooks cannot be made for private channels
	require.Error(t, th.RunCommand(t, "webhook", "create-outgoing", "--team", team, "--channel", th.BasicPrivateChannel.Id, "--display-name", displayName, "--trigger-word", triggerWord1, "--trigger-word", triggerWord2, "--url", callbackURL1, "--url", callbackURL2, "--user", user))

	th.CheckCommand(t, "webhook", "create-outgoing", "--team", team, "--channel", th.BasicChannel.Id, "--display-name", displayName, "--trigger-word", triggerWord1, "--trigger-word", triggerWord2, "--url", callbackURL1, "--url", callbackURL2, "--user", user)

	webhooks, err := th.App.GetOutgoingWebhooksPage(0, 1000)
	if err != nil {
		t.Fatal("Unable to retreive outgoing webhooks")
	}

	found := false
	for _, webhook := range webhooks {
		if webhook.DisplayName == displayName && webhook.CreatorId == th.BasicUser.Id {
			found = true
		}
	}
	if !found {
		t.Fatal("Failed to create incoming webhook")
	}
}

func TestModifyOutgoingWebhook(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	config := th.Config()
	*config.ServiceSettings.EnableOutgoingWebhooks = true
	th.SetConfig(config)

	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()
	th.AddPermissionToRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)

	description := "myhookoutdesc"
	displayName := "myhookoutname"
	triggerWords := model.StringArray{"myhookoutword1"}
	triggerWhen := 0
	callbackURLs := model.StringArray{"http://myhookouturl1"}
	iconURL := "myhookicon1"
	contentType := "myhookcontent1"

	outgoingWebhook := &model.OutgoingWebhook{
		CreatorId:    th.BasicUser.Id,
		Username:     th.BasicUser.Username,
		TeamId:       th.BasicTeam.Id,
		ChannelId:    th.BasicChannel.Id,
		DisplayName:  displayName,
		Description:  description,
		TriggerWords: triggerWords,
		TriggerWhen:  triggerWhen,
		CallbackURLs: callbackURLs,
		IconURL:      iconURL,
		ContentType:  contentType,
	}

	oldHook, err := th.App.CreateOutgoingWebhook(outgoingWebhook)
	if err != nil {
		t.Fatal("unable to create outgoing webhooks: " + err.Error())
	}
	defer func() {
		th.App.DeleteOutgoingWebhook(oldHook.Id)
	}()

	// should fail because you need to specify valid outgoing webhook
	require.Error(t, th.RunCommand(t, "webhook", "modify-outgoing", "doesnotexist"))
	// should fail because you need to specify valid channel
	require.Error(t, th.RunCommand(t, "webhook", "modify-outgoing", oldHook.Id, "--channel", th.BasicTeam.Name+":doesnotexist"))
	// should fail because you need to specify valid trigger when
	require.Error(t, th.RunCommand(t, "webhook", "modify-outgoing", oldHook.Id, "--channel", th.BasicTeam.Name+th.BasicChannel.Id, "--trigger-when", "invalid"))
	// should fail because you need to specify a valid callback URL
	require.Error(t, th.RunCommand(t, "webhook", "modify-outgoing", oldHook.Id, "--channel", th.BasicTeam.Name+th.BasicChannel.Id, "--callback-url", "invalid"))

	modifiedChannelID := th.BasicChannel2.Id
	modifiedDisplayName := "myhookoutname2"
	modifiedDescription := "myhookoutdesc2"
	modifiedTriggerWords := model.StringArray{"myhookoutword2A", "myhookoutword2B"}
	modifiedTriggerWhen := "start"
	modifiedIconURL := "myhookouticon2"
	modifiedContentType := "myhookcontent2"
	modifiedCallbackURLs := model.StringArray{"http://myhookouturl2A", "http://myhookouturl2B"}

	th.CheckCommand(t, "webhook", "modify-outgoing", oldHook.Id,
		"--channel", modifiedChannelID,
		"--display-name", modifiedDisplayName,
		"--description", modifiedDescription,
		"--trigger-word", modifiedTriggerWords[0],
		"--trigger-word", modifiedTriggerWords[1],
		"--trigger-when", modifiedTriggerWhen,
		"--icon", modifiedIconURL,
		"--content-type", modifiedContentType,
		"--url", modifiedCallbackURLs[0],
		"--url", modifiedCallbackURLs[1],
	)

	modifiedHook, err := th.App.GetOutgoingWebhook(oldHook.Id)
	if err != nil {
		t.Fatal("unable to retrieve modified outgoing webhook")
	}

	updateFailed := modifiedHook.ChannelId != modifiedChannelID ||
		modifiedHook.DisplayName != modifiedDisplayName ||
		modifiedHook.Description != modifiedDescription ||
		len(modifiedHook.TriggerWords) != len(modifiedTriggerWords) ||
		modifiedHook.TriggerWords[0] != modifiedTriggerWords[0] ||
		modifiedHook.TriggerWords[1] != modifiedTriggerWords[1] ||
		modifiedHook.TriggerWhen != 1 ||
		modifiedHook.IconURL != modifiedIconURL ||
		modifiedHook.ContentType != modifiedContentType ||
		len(modifiedHook.CallbackURLs) != len(modifiedCallbackURLs) ||
		modifiedHook.CallbackURLs[0] != modifiedCallbackURLs[0] ||
		modifiedHook.CallbackURLs[1] != modifiedCallbackURLs[1]

	if updateFailed {
		t.Fatal("Failed to update outgoing webhook")
	}
}

func TestDeleteWebhooks(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	adminClient := th.SystemAdminClient

	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	*config.ServiceSettings.EnableIncomingWebhooks = true
	*config.ServiceSettings.EnableOutgoingWebhooks = true
	*config.ServiceSettings.EnablePostUsernameOverride = true
	*config.ServiceSettings.EnablePostIconOverride = true
	th.SetConfig(config)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableIncomingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOutgoingWebhooks = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostUsernameOverride = true })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnablePostIconOverride = true })

	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()
	th.AddPermissionToRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_INCOMING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_MANAGE_OUTGOING_WEBHOOKS.Id, model.TEAM_USER_ROLE_ID)

	dispName := "myhookinc"
	inHookStruct := &model.IncomingWebhook{DisplayName: dispName, ChannelId: th.BasicChannel.Id, TeamId: th.BasicChannel.TeamId}
	incomingHook, resp := adminClient.CreateIncomingWebhook(inHookStruct)
	api4.CheckNoError(t, resp)

	dispName2 := "myhookout"
	outHookStruct := &model.OutgoingWebhook{DisplayName: dispName2, ChannelId: th.BasicChannel.Id, TeamId: th.BasicChannel.TeamId, CallbackURLs: []string{"http://nowhere.com"}, Username: "some-user-name", IconURL: "http://some-icon-url/"}
	outgoingHook, resp := adminClient.CreateOutgoingWebhook(outHookStruct)
	api4.CheckNoError(t, resp)

	hooksBeforeDeletion := th.CheckCommand(t, "webhook", "list", th.BasicTeam.Name)

	if !strings.Contains(string(hooksBeforeDeletion), dispName) {
		t.Fatal("Should have incoming webhooks")
	}

	if !strings.Contains(string(hooksBeforeDeletion), dispName2) {
		t.Fatal("Should have outgoing webhooks")
	}

	th.CheckCommand(t, "webhook", "delete", incomingHook.Id)
	th.CheckCommand(t, "webhook", "delete", outgoingHook.Id)

	hooksAfterDeletion := th.CheckCommand(t, "webhook", "list", th.BasicTeam.Name)

	if strings.Contains(string(hooksAfterDeletion), dispName) {
		t.Fatal("Should not have incoming webhooks")
	}

	if strings.Contains(string(hooksAfterDeletion), dispName2) {
		t.Fatal("Should not have outgoing webhooks")
	}
}
