// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	th.SetConfig(config)

	team := th.BasicTeam
	adminUser := th.TeamAdminUser
	user := th.BasicUser

	testCases := []struct {
		Description string
		Args        []string
		ExpectedErr string
	}{
		{
			"nil error",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"",
		},
		{
			"Team not specified",
			[]string{"command", "create", "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Error: requires at least 1 arg(s), only received 0",
		},
		{
			"Team not found",
			[]string{"command", "create", "fakeTeam", "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Error: unable to find team",
		},
		{
			"Creator not specified",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler"},
			`Error: required flag(s) "creator" not set`,
		},
		{
			"Creator not found",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", "fakeuser"},
			"unable to find user",
		},
		{
			"Creator not team admin",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", user.Username},
			"the creator must be a user who has permissions to manage slash commands",
		},
		{
			"Command not specified",
			[]string{"command", "", team.Name, "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Error: unknown flag: --trigger-word",
		},
		{
			"Trigger not specified",
			[]string{"command", "create", team.Name, "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			`Error: required flag(s) "trigger-word" not set`,
		},
		{
			"Blank trigger",
			[]string{"command", "create", team.Name, "--trigger-word", "", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Invalid trigger",
		},
		{
			"Trigger with space",
			[]string{"command", "create", team.Name, "--trigger-word", "test cmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Error: a trigger word must not contain spaces",
		},
		{
			"Trigger starting with /",
			[]string{"command", "create", team.Name, "--trigger-word", "/testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Error: a trigger word cannot begin with a /",
		},
		{
			"URL not specified",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd", "--creator", adminUser.Username},
			`Error: required flag(s) "url" not set`,
		},
		{
			"Blank URL",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd2", "--url", "", "--creator", adminUser.Username},
			"Invalid URL",
		},
		{
			"Invalid URL",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd2", "--url", "localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Invalid URL",
		},
		{
			"Duplicate Command",
			[]string{"command", "create", team.Name, "--trigger-word", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"This trigger word is already in use",
		},
		{
			"Misspelled flag",
			[]string{"command", "create", team.Name, "--trigger-wor", "testcmd", "--url", "http://localhost:8000/my-slash-handler", "--creator", adminUser.Username},
			"Error: unknown flag:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			actual, _ := th.RunCommandWithOutput(t, testCase.Args...)

			cmds, _ := th.SystemAdminClient.ListCommands(team.Id, true)

			if testCase.ExpectedErr == "" {
				if len(cmds) == 0 || cmds[0].Trigger != "testcmd" {
					t.Fatal("Failed to create command")
				}
				assert.Contains(t, string(actual), "PASS")
			} else {
				if len(cmds) > 1 {
					t.Fatal("Created command that shouldn't have been created")
				}
				assert.Contains(t, string(actual), testCase.ExpectedErr)
			}
		})
	}
}

func TestShowCommand(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	url := "http://localhost:8000/test-command"
	team := th.BasicTeam
	user := th.BasicUser
	th.LinkUserToTeam(user, team)
	trigger := "trigger_" + model.NewId()
	displayName := "dn_" + model.NewId()

	c := &model.Command{
		DisplayName: displayName,
		Method:      "G",
		TeamId:      team.Id,
		Username:    user.Username,
		CreatorId:   user.Id,
		URL:         url,
		Trigger:     trigger,
	}

	t.Run("existing command", func(t *testing.T) {
		command, err := th.App.CreateCommand(c)
		require.Nil(t, err)
		commands, err := th.App.ListTeamCommands(team.Id)
		require.Nil(t, err)
		assert.Equal(t, len(commands), 1)

		output := th.CheckCommand(t, "command", "show", command.Id)
		assert.Contains(t, string(output), command.Id)
		assert.Contains(t, string(output), command.TeamId)
		assert.Contains(t, string(output), trigger)
		assert.Contains(t, string(output), displayName)
		assert.Contains(t, string(output), user.Username)
	})

	t.Run("not existing command", func(t *testing.T) {
		assert.Error(t, th.RunCommand(t, "command", "show", "invalid"))
	})

	t.Run("no commandID", func(t *testing.T) {
		assert.Error(t, th.RunCommand(t, "command", "show"))
	})
}

func TestDeleteCommand(t *testing.T) {
	// Skipped due to v5.6 RC build issues.
	t.Skip()

	th := Setup().InitBasic()
	defer th.TearDown()
	url := "http://localhost:8000/test-command"
	team := th.BasicTeam
	user := th.BasicUser
	th.LinkUserToTeam(user, team)

	c := &model.Command{
		DisplayName: "dn_" + model.NewId(),
		Method:      "G",
		TeamId:      team.Id,
		Username:    user.Username,
		CreatorId:   user.Id,
		URL:         url,
		Trigger:     "trigger_" + model.NewId(),
	}

	t.Run("existing command", func(t *testing.T) {
		command, err := th.App.CreateCommand(c)
		require.Nil(t, err)
		commands, err := th.App.ListTeamCommands(team.Id)
		require.Nil(t, err)
		assert.Equal(t, len(commands), 1)

		th.CheckCommand(t, "command", "delete", command.Id)
		commands, err = th.App.ListTeamCommands(team.Id)
		require.Nil(t, err)
		assert.Equal(t, len(commands), 0)
	})

	t.Run("not existing command", func(t *testing.T) {
		assert.Error(t, th.RunCommand(t, "command", "delete", "invalid"))
	})
}

func TestModifyCommand(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	// set config
	config := th.Config()
	*config.ServiceSettings.EnableCommands = true
	th.SetConfig(config)

	// set team and users
	team := th.BasicTeam
	adminUser := th.TeamAdminUser
	user := th.BasicUser

	// create test command to modify
	url := "http://localhost:8000/test-command"
	th.LinkUserToTeam(user, team)
	trigger := "trigger_" + model.NewId()
	displayName := "dn_" + model.NewId()

	c := &model.Command{
		DisplayName: displayName,
		Method:      "G",
		TeamId:      team.Id,
		Username:    user.Username,
		CreatorId:   user.Id,
		URL:         url,
		Trigger:     trigger,
	}

	command, err := th.App.CreateCommand(c)
	require.Nil(t, err)
	commands, err := th.App.ListTeamCommands(team.Id)
	require.Nil(t, err)
	assert.Equal(t, len(commands), 1)

	t.Run("command not specified", func(t *testing.T) {
		args := []string{"command", "", command.Id, "--trigger-word", "sometrigger"}
		output, _ := th.RunCommandWithOutput(t, args...)
		assert.Contains(t, string(output), "Error: unknown flag: --trigger-word")
	})

	t.Run("modify command unchanged", func(t *testing.T) {
		args := []string{"command", "modify", command.Id}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.DisplayName, command.DisplayName)
		assert.Equal(t, cmd.Method, command.Method)
		assert.Equal(t, cmd.TeamId, command.TeamId)
		assert.Equal(t, cmd.Username, command.Username)
		assert.Equal(t, cmd.CreatorId, command.CreatorId)
		assert.Equal(t, cmd.URL, command.URL)
		assert.Equal(t, cmd.Trigger, command.Trigger)
		assert.Equal(t, cmd.AutoComplete, command.AutoComplete)
		assert.Equal(t, cmd.AutoCompleteDesc, command.AutoCompleteDesc)
		assert.Equal(t, cmd.AutoCompleteHint, command.AutoCompleteHint)
		assert.Equal(t, cmd.IconURL, command.IconURL)
	})

	t.Run("mispelled flag", func(t *testing.T) {
		args := []string{"command", "", command.Id, "--trigger-wor", "sometrigger"}
		output, _ := th.RunCommandWithOutput(t, args...)
		assert.Contains(t, string(output), "Error: unknown flag:")
	})

	t.Run("multiple flags nil error", func(t *testing.T) {
		testName := "multitrigger"
		testURL := "http://localhost:8000/test-modify"
		testDescription := "multiple field test"
		args := []string{"command", "modify", command.Id, "--trigger-word", testName, "--url", testURL, "--description", testDescription}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.Trigger, testName)
		assert.Equal(t, cmd.URL, testURL)
		assert.Equal(t, cmd.Description, testDescription)
	})

	t.Run("displayname nil error", func(t *testing.T) {
		testVal := "newName"
		args := []string{"command", "modify", command.Id, "--title", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.DisplayName, testVal)
	})

	t.Run("description nil error", func(t *testing.T) {
		testVal := "test description"
		args := []string{"command", "modify", command.Id, "--description", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.Description, testVal)
	})

	t.Run("trigger nil error", func(t *testing.T) {
		testVal := "testtrigger"
		args := []string{"command", "modify", command.Id, "--trigger-word", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.Trigger, testVal)
	})

	t.Run("trigger with space", func(t *testing.T) {
		testVal := "bad trigger"
		args := []string{"command", "modify", command.Id, "--trigger-word", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		assert.Contains(t, string(output), "Error: a trigger word must not contain spaces")
	})

	t.Run("trigger with leading /", func(t *testing.T) {
		testVal := "/bad-trigger"
		args := []string{"command", "modify", command.Id, "--trigger-word", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		assert.Contains(t, string(output), "Error: a trigger word cannot begin with a /")
	})

	t.Run("blank trigger", func(t *testing.T) {
		cmd_unmodified, _ := th.App.GetCommand(command.Id)
		args := []string{"command", "modify", command.Id, "--trigger-word", ""}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd_modified, _ := th.App.GetCommand(command.Id)

		// assert trigger remains unchanged
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd_unmodified.Trigger, cmd_modified.Trigger)
	})

	//url case
	t.Run("url nil error", func(t *testing.T) {
		testVal := "http://localhost:8000/modify-command"
		args := []string{"command", "modify", command.Id, "--url", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.URL, testVal)
	})

	t.Run("blank url", func(t *testing.T) {
		cmd_unmodified, _ := th.App.GetCommand(command.Id)
		args := []string{"command", "modify", command.Id, "--url", ""}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd_modified, _ := th.App.GetCommand(command.Id)

		//assert URL remains unchanged
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd_unmodified.URL, cmd_modified.URL)
	})

	t.Run("icon url nil error", func(t *testing.T) {
		testVal := "http://localhost:8000/testicon.png"
		args := []string{"command", "modify", command.Id, "--icon", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.IconURL, testVal)
	})

	t.Run("creator nil error", func(t *testing.T) {
		testVal := adminUser
		args := []string{"command", "modify", command.Id, "--creator", testVal.Username}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.CreatorId, testVal.Id)
	})

	t.Run("creator not found", func(t *testing.T) {
		testVal := "fakeuser"
		args := []string{"command", "modify", command.Id, "--creator", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		assert.Contains(t, string(output), "unable to find user")
	})

	t.Run("creator not admin user", func(t *testing.T) {
		testVal := user.Username
		args := []string{"command", "modify", command.Id, "--creator", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		assert.Contains(t, string(output), "the creator must be a user who has permissions to manage slash commands")
	})

	t.Run("response username nil error", func(t *testing.T) {
		testVal := "response-test"
		args := []string{"command", "modify", command.Id, "--response-username", testVal}
		output, _ := th.RunCommandWithOutput(t, args...)
		cmd, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output), "PASS")
		assert.Equal(t, cmd.Username, testVal)
	})

	t.Run("post set and unset", func(t *testing.T) {
		args_set := []string{"command", "modify", command.Id, "--post", ""}
		args_unset := []string{"command", "modify", command.Id, "", ""}

		// set post and check
		output_set, _ := th.RunCommandWithOutput(t, args_set...)
		cmd_set, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output_set), "PASS")
		assert.Equal(t, cmd_set.Method, "P")

		// unset post and check
		output_unset, _ := th.RunCommandWithOutput(t, args_unset...)
		cmd_unset, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output_unset), "PASS")
		assert.Equal(t, cmd_unset.Method, "G")
	})

	t.Run("autocomplete set and unset", func(t *testing.T) {
		args_set := []string{"command", "modify", command.Id, "--autocomplete", ""}
		args_unset := []string{"command", "modify", command.Id, "", ""}

		// set autocomplete and check
		output_set, _ := th.RunCommandWithOutput(t, args_set...)
		cmd_set, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output_set), "PASS")
		assert.Equal(t, cmd_set.AutoComplete, true)

		// unset autocomplete and check
		output_unset, _ := th.RunCommandWithOutput(t, args_unset...)
		cmd_unset, _ := th.App.GetCommand(command.Id)
		assert.Contains(t, string(output_unset), "PASS")
		assert.Equal(t, cmd_unset.AutoComplete, false)
	})
}
