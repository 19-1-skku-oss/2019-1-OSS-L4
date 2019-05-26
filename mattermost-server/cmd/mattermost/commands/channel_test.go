// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinChannel(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	channel := th.CreatePublicChannel()

	th.CheckCommand(t, "channel", "add", th.BasicTeam.Name+":"+channel.Name, th.BasicUser2.Email)

	// Joining twice should succeed
	th.CheckCommand(t, "channel", "add", th.BasicTeam.Name+":"+channel.Name, th.BasicUser2.Email)

	// should fail because channel does not exist
	require.Error(t, th.RunCommand(t, "channel", "add", th.BasicTeam.Name+":"+channel.Name+"asdf", th.BasicUser2.Email))
}

func TestRemoveChannel(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	channel := th.CreatePublicChannel()

	th.CheckCommand(t, "channel", "add", th.BasicTeam.Name+":"+channel.Name, th.BasicUser2.Email)

	// should fail because channel does not exist
	require.Error(t, th.RunCommand(t, "channel", "remove", th.BasicTeam.Name+":doesnotexist", th.BasicUser2.Email))

	time.Sleep(time.Second)

	th.CheckCommand(t, "channel", "remove", th.BasicTeam.Name+":"+channel.Name, th.BasicUser2.Email)

	time.Sleep(time.Second)

	// Leaving twice should succeed
	th.CheckCommand(t, "channel", "remove", th.BasicTeam.Name+":"+channel.Name, th.BasicUser2.Email)
}

func TestMoveChannel(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team1 := th.BasicTeam
	team2 := th.CreateTeam()
	user1 := th.BasicUser
	th.LinkUserToTeam(user1, team2)
	channel := th.BasicChannel

	th.LinkUserToTeam(user1, team1)
	th.LinkUserToTeam(user1, team2)

	adminEmail := user1.Email
	adminUsername := user1.Username
	origin := team1.Name + ":" + channel.Name
	dest := team2.Name

	th.CheckCommand(t, "channel", "add", origin, adminEmail)

	// should fail with nil because errors are logged instead of returned when a channel does not exist
	th.CheckCommand(t, "channel", "move", dest, team1.Name+":doesnotexist", "--username", adminUsername)

	th.CheckCommand(t, "channel", "move", dest, origin, "--username", adminUsername)
}

func TestListChannels(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	channel := th.CreatePublicChannel()
	th.Client.Must(th.Client.DeleteChannel(channel.Id))

	output := th.CheckCommand(t, "channel", "list", th.BasicTeam.Name)

	if !strings.Contains(string(output), "town-square") {
		t.Fatal("should have channels")
	}

	if !strings.Contains(string(output), channel.Name+" (archived)") {
		t.Fatal("should have archived channel")
	}
}

func TestRestoreChannel(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	channel := th.CreatePublicChannel()
	th.Client.Must(th.Client.DeleteChannel(channel.Id))

	th.CheckCommand(t, "channel", "restore", th.BasicTeam.Name+":"+channel.Name)

	// restoring twice should succeed
	th.CheckCommand(t, "channel", "restore", th.BasicTeam.Name+":"+channel.Name)
}

func TestCreateChannel(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	id := model.NewId()
	name := "name" + id

	th.CheckCommand(t, "channel", "create", "--display_name", name, "--team", th.BasicTeam.Name, "--name", name)

	name = name + "-private"
	th.CheckCommand(t, "channel", "create", "--display_name", name, "--team", th.BasicTeam.Name, "--private", "--name", name)
}

func TestRenameChannel(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	channel := th.CreatePublicChannel()
	th.CheckCommand(t, "channel", "rename", th.BasicTeam.Name+":"+channel.Name, "newchannelname10", "--display_name", "New Display Name")

	// Get the channel from the DB
	updatedChannel, _ := th.App.GetChannel(channel.Id)
	assert.Equal(t, "newchannelname10", updatedChannel.Name)
	assert.Equal(t, "New Display Name", updatedChannel.DisplayName)
}

func Test_searchChannelCmdF(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	channel := th.CreatePublicChannel()
	channel2 := th.CreatePublicChannel()
	th.Client.DeleteChannel(channel2.Id)

	tests := []struct {
		Name     string
		Args     []string
		Expected string
	}{
		{
			"Success find Channel in any team",
			[]string{"channel", "search", channel.Name},
			fmt.Sprintf("Channel Name :%s, Display Name :%s, Channel ID :%s", channel.Name, channel.DisplayName, channel.Id),
		},
		{
			"Failed find Channel in any team",
			[]string{"channel", "search", channel.Name + "404"},
			fmt.Sprintf("Channel %s is not found in any team", channel.Name+"404"),
		},
		{
			"Success find Channel with param team ID",
			[]string{"channel", "search", "--team", channel.TeamId, channel.Name},
			fmt.Sprintf("Channel Name :%s, Display Name :%s, Channel ID :%s", channel.Name, channel.DisplayName, channel.Id),
		},
		{
			"Failed find Channel with param team ID",
			[]string{"channel", "search", "--team", channel.TeamId, channel.Name + "404"},
			fmt.Sprintf("Channel %s is not found in team %s", channel.Name+"404", channel.TeamId),
		},
		{
			"Success find archived Channel in any team",
			[]string{"channel", "search", channel2.Name},
			fmt.Sprintf("Channel Name :%s, Display Name :%s, Channel ID :%s (archived)", channel2.Name, channel2.DisplayName, channel2.Id),
		},
		{
			"Success find archived Channel with param team ID",
			[]string{"channel", "search", "--team", channel2.TeamId, channel2.Name},
			fmt.Sprintf("Channel Name :%s, Display Name :%s, Channel ID :%s (archived)", channel2.Name, channel2.DisplayName, channel2.Id),
		},
		{
			"Failed find team",
			[]string{"channel", "search", "--team", channel.TeamId + "404", channel.Name},
			fmt.Sprintf("Team %s is not found", channel.TeamId+"404"),
		},
		{
			"Success find Channel with param team ID",
			[]string{"channel", "search", channel.Name, "--team", channel.TeamId},
			fmt.Sprintf("Channel Name :%s, Display Name :%s, Channel ID :%s", channel.Name, channel.DisplayName, channel.Id),
		},
		{
			"Success find Channel with param team ID",
			[]string{"channel", "search", channel.Name, "--team=" + channel.TeamId},
			fmt.Sprintf("Channel Name :%s, Display Name :%s, Channel ID :%s", channel.Name, channel.DisplayName, channel.Id),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			assert.Contains(t, th.CheckCommand(t, test.Args...), test.Expected)
		})
	}
}
