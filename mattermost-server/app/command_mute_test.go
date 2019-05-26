// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"testing"
	"time"

	"github.com/mattermost/go-i18n/i18n"
	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
)

func TestMuteCommandNoChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	if testing.Short() {
		t.SkipNow()
	}

	channel1 := th.BasicChannel
	channel1M, channel1MError := th.App.GetChannelMember(channel1.Id, th.BasicUser.Id)

	if channel1MError != nil {
		t.Fatal("User is not a member of channel 1")
	}

	if channel1M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP] == model.CHANNEL_NOTIFY_MENTION {
		t.Fatal("channel shouldn't be muted on initial setup")
	}

	cmd := &MuteProvider{}
	resp := cmd.DoCommand(th.App, &model.CommandArgs{
		T:      i18n.IdentityTfunc(),
		UserId: th.BasicUser.Id,
	}, "")
	assert.Equal(t, "api.command_mute.no_channel.error", resp.Text)
}

func TestMuteCommandNoArgs(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	channel1 := th.BasicChannel
	channel1M, _ := th.App.GetChannelMember(channel1.Id, th.BasicUser.Id)

	assert.Equal(t, model.CHANNEL_NOTIFY_ALL, channel1M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP])

	cmd := &MuteProvider{}

	// First mute the channel
	resp := cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel1.Id,
		UserId:    th.BasicUser.Id,
	}, "")
	assert.Equal(t, "api.command_mute.success_mute", resp.Text)

	// Now unmute the channel
	time.Sleep(time.Millisecond)
	resp = cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel1.Id,
		UserId:    th.BasicUser.Id,
	}, "")

	assert.Equal(t, "api.command_mute.success_unmute", resp.Text)
}

func TestMuteCommandSpecificChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	if testing.Short() {
		t.SkipNow()
	}

	channel1 := th.BasicChannel
	channel2, _ := th.App.CreateChannel(&model.Channel{
		DisplayName: "AA",
		Name:        "aa" + model.NewId() + "a",
		Type:        model.CHANNEL_OPEN,
		TeamId:      th.BasicTeam.Id,
		CreatorId:   th.BasicUser.Id,
	}, true)

	channel2M, _ := th.App.GetChannelMember(channel2.Id, th.BasicUser.Id)

	assert.Equal(t, model.CHANNEL_NOTIFY_ALL, channel2M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP])

	cmd := &MuteProvider{}

	// First mute the channel
	resp := cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel1.Id,
		UserId:    th.BasicUser.Id,
	}, channel2.Name)
	assert.Equal(t, "api.command_mute.success_mute", resp.Text)
	channel2M, _ = th.App.GetChannelMember(channel2.Id, th.BasicUser.Id)
	assert.Equal(t, model.CHANNEL_NOTIFY_MENTION, channel2M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP])

	// Now unmute the channel
	resp = cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel1.Id,
		UserId:    th.BasicUser.Id,
	}, "~"+channel2.Name)

	assert.Equal(t, "api.command_mute.success_unmute", resp.Text)
	channel2M, _ = th.App.GetChannelMember(channel2.Id, th.BasicUser.Id)
	assert.Equal(t, model.CHANNEL_NOTIFY_ALL, channel2M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP])
}

func TestMuteCommandNotMember(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	if testing.Short() {
		t.SkipNow()
	}

	channel1 := th.BasicChannel
	channel2, _ := th.App.CreateChannel(&model.Channel{
		DisplayName: "AA",
		Name:        "aa" + model.NewId() + "a",
		Type:        model.CHANNEL_OPEN,
		TeamId:      th.BasicTeam.Id,
		CreatorId:   th.BasicUser.Id,
	}, false)

	cmd := &MuteProvider{}

	// First mute the channel
	resp := cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel1.Id,
		UserId:    th.BasicUser.Id,
	}, channel2.Name)
	assert.Equal(t, "api.command_mute.not_member.error", resp.Text)
}

func TestMuteCommandNotChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	if testing.Short() {
		t.SkipNow()
	}

	channel1 := th.BasicChannel

	cmd := &MuteProvider{}

	// First mute the channel
	resp := cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel1.Id,
		UserId:    th.BasicUser.Id,
	}, "~noexists")
	assert.Equal(t, "api.command_mute.error", resp.Text)
}

func TestMuteCommandDMChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	if testing.Short() {
		t.SkipNow()
	}

	channel2, _ := th.App.GetOrCreateDirectChannel(th.BasicUser.Id, th.BasicUser2.Id)
	channel2M, _ := th.App.GetChannelMember(channel2.Id, th.BasicUser.Id)

	assert.Equal(t, model.CHANNEL_NOTIFY_ALL, channel2M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP])

	cmd := &MuteProvider{}

	// First mute the channel
	resp := cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel2.Id,
		UserId:    th.BasicUser.Id,
	}, "")
	assert.Equal(t, "api.command_mute.success_mute_direct_msg", resp.Text)
	time.Sleep(time.Millisecond)
	channel2M, _ = th.App.GetChannelMember(channel2.Id, th.BasicUser.Id)
	assert.Equal(t, model.CHANNEL_NOTIFY_MENTION, channel2M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP])

	// Now unmute the channel
	resp = cmd.DoCommand(th.App, &model.CommandArgs{
		T:         i18n.IdentityTfunc(),
		ChannelId: channel2.Id,
		UserId:    th.BasicUser.Id,
	}, "")

	assert.Equal(t, "api.command_mute.success_unmute_direct_msg", resp.Text)
	time.Sleep(time.Millisecond)
	channel2M, _ = th.App.GetChannelMember(channel2.Id, th.BasicUser.Id)
	assert.Equal(t, model.CHANNEL_NOTIFY_ALL, channel2M.NotifyProps[model.MARK_UNREAD_NOTIFY_PROP])
}
