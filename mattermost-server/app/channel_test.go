// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
)

func TestPermanentDeleteChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.ServiceSettings.EnableIncomingWebhooks = true
		*cfg.ServiceSettings.EnableOutgoingWebhooks = true
	})

	channel, err := th.App.CreateChannel(&model.Channel{DisplayName: "deletion-test", Name: "deletion-test", Type: model.CHANNEL_OPEN, TeamId: th.BasicTeam.Id}, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		th.App.PermanentDeleteChannel(channel)
	}()

	incoming, err := th.App.CreateIncomingWebhookForChannel(th.BasicUser.Id, channel, &model.IncomingWebhook{ChannelId: channel.Id})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer th.App.DeleteIncomingWebhook(incoming.Id)

	if incoming, err = th.App.GetIncomingWebhook(incoming.Id); incoming == nil || err != nil {
		t.Fatal("unable to get new incoming webhook")
	}

	outgoing, err := th.App.CreateOutgoingWebhook(&model.OutgoingWebhook{
		ChannelId:    channel.Id,
		TeamId:       channel.TeamId,
		CreatorId:    th.BasicUser.Id,
		CallbackURLs: []string{"http://foo"},
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer th.App.DeleteOutgoingWebhook(outgoing.Id)

	if outgoing, err = th.App.GetOutgoingWebhook(outgoing.Id); outgoing == nil || err != nil {
		t.Fatal("unable to get new outgoing webhook")
	}

	err = th.App.PermanentDeleteChannel(channel)
	require.Nil(t, err)

	if incoming, err = th.App.GetIncomingWebhook(incoming.Id); incoming != nil || err == nil {
		t.Error("incoming webhook wasn't deleted")
	}

	if outgoing, err = th.App.GetOutgoingWebhook(outgoing.Id); outgoing != nil || err == nil {
		t.Error("outgoing webhook wasn't deleted")
	}
}

func TestMoveChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	sourceTeam := th.CreateTeam()
	targetTeam := th.CreateTeam()
	channel1 := th.CreateChannel(sourceTeam)
	defer func() {
		th.App.PermanentDeleteChannel(channel1)
		th.App.PermanentDeleteTeam(sourceTeam)
		th.App.PermanentDeleteTeam(targetTeam)
	}()

	if _, err := th.App.AddUserToTeam(sourceTeam.Id, th.BasicUser.Id, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := th.App.AddUserToTeam(sourceTeam.Id, th.BasicUser2.Id, ""); err != nil {
		t.Fatal(err)
	}

	if _, err := th.App.AddUserToTeam(targetTeam.Id, th.BasicUser.Id, ""); err != nil {
		t.Fatal(err)
	}

	if _, err := th.App.AddUserToChannel(th.BasicUser, channel1); err != nil {
		t.Fatal(err)
	}
	if _, err := th.App.AddUserToChannel(th.BasicUser2, channel1); err != nil {
		t.Fatal(err)
	}

	if err := th.App.MoveChannel(targetTeam, channel1, th.BasicUser, false); err == nil {
		t.Fatal("Should have failed due to mismatched members.")
	}

	if _, err := th.App.AddUserToTeam(targetTeam.Id, th.BasicUser2.Id, ""); err != nil {
		t.Fatal(err)
	}

	if err := th.App.MoveChannel(targetTeam, channel1, th.BasicUser, false); err != nil {
		t.Fatal(err)
	}

	// Test moving a channel with a deactivated user who isn't in the destination team.
	// It should fail, unless removeDeactivatedMembers is true.
	deacivatedUser := th.CreateUser()
	channel2 := th.CreateChannel(sourceTeam)

	if _, err := th.App.AddUserToTeam(sourceTeam.Id, deacivatedUser.Id, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := th.App.AddUserToChannel(th.BasicUser, channel2); err != nil {
		t.Fatal(err)
	}
	if _, err := th.App.AddUserToChannel(deacivatedUser, channel2); err != nil {
		t.Fatal(err)
	}

	if _, err := th.App.UpdateActive(deacivatedUser, false); err != nil {
		t.Fatal(err)
	}

	if err := th.App.MoveChannel(targetTeam, channel2, th.BasicUser, false); err == nil {
		t.Fatal("Should have failed due to mismatched deacivated member.")
	}

	if err := th.App.MoveChannel(targetTeam, channel2, th.BasicUser, true); err != nil {
		t.Fatal(err)
	}

	// Test moving a channel with no members.
	channel3 := &model.Channel{
		DisplayName: "dn_" + model.NewId(),
		Name:        "name_" + model.NewId(),
		Type:        model.CHANNEL_OPEN,
		TeamId:      sourceTeam.Id,
		CreatorId:   th.BasicUser.Id,
	}

	var err *model.AppError
	channel3, err = th.App.CreateChannel(channel3, false)
	require.Nil(t, err)

	err = th.App.MoveChannel(targetTeam, channel3, th.BasicUser, false)
	assert.Nil(t, err)
}

func TestJoinDefaultChannelsCreatesChannelMemberHistoryRecordTownSquare(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// figure out the initial number of users in town square
	townSquareChannelId := store.Must(th.App.Srv.Store.Channel().GetByName(th.BasicTeam.Id, "town-square", true)).(*model.Channel).Id
	initialNumTownSquareUsers := len(store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, townSquareChannelId)).([]*model.ChannelMemberHistoryResult))

	// create a new user that joins the default channels
	user := th.CreateUser()
	th.App.JoinDefaultChannels(th.BasicTeam.Id, user, false, "")

	// there should be a ChannelMemberHistory record for the user
	histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, townSquareChannelId)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, initialNumTownSquareUsers+1)

	found := false
	for _, history := range histories {
		if user.Id == history.UserId && townSquareChannelId == history.ChannelId {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestJoinDefaultChannelsCreatesChannelMemberHistoryRecordOffTopic(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// figure out the initial number of users in off-topic
	offTopicChannelId := store.Must(th.App.Srv.Store.Channel().GetByName(th.BasicTeam.Id, "off-topic", true)).(*model.Channel).Id
	initialNumTownSquareUsers := len(store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, offTopicChannelId)).([]*model.ChannelMemberHistoryResult))

	// create a new user that joins the default channels
	user := th.CreateUser()
	th.App.JoinDefaultChannels(th.BasicTeam.Id, user, false, "")

	// there should be a ChannelMemberHistory record for the user
	histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, offTopicChannelId)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, initialNumTownSquareUsers+1)

	found := false
	for _, history := range histories {
		if user.Id == history.UserId && offTopicChannelId == history.ChannelId {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestJoinDefaultChannelsExperimentalDefaultChannels(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	basicChannel2 := th.CreateChannel(th.BasicTeam)
	defaultChannelList := []string{th.BasicChannel.Name, basicChannel2.Name, basicChannel2.Name}
	th.App.Config().TeamSettings.ExperimentalDefaultChannels = defaultChannelList

	user := th.CreateUser()
	th.App.JoinDefaultChannels(th.BasicTeam.Id, user, false, "")

	for _, channelName := range defaultChannelList {
		channel, err := th.App.GetChannelByName(channelName, th.BasicTeam.Id, false)

		if err != nil {
			t.Errorf("Expected nil, got %s", err)
		}

		member, err := th.App.GetChannelMember(channel.Id, user.Id)

		if member == nil {
			t.Errorf("Expected member object, got nil")
		}

		if err != nil {
			t.Errorf("Expected nil object, got %s", err)
		}
	}
}

func TestCreateChannelPublicCreatesChannelMemberHistoryRecord(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// creates a public channel and adds basic user to it
	publicChannel := th.createChannel(th.BasicTeam, model.CHANNEL_OPEN)

	// there should be a ChannelMemberHistory record for the user
	histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, publicChannel.Id)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, 1)
	assert.Equal(t, th.BasicUser.Id, histories[0].UserId)
	assert.Equal(t, publicChannel.Id, histories[0].ChannelId)
}

func TestCreateChannelPrivateCreatesChannelMemberHistoryRecord(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// creates a private channel and adds basic user to it
	privateChannel := th.createChannel(th.BasicTeam, model.CHANNEL_PRIVATE)

	// there should be a ChannelMemberHistory record for the user
	histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, privateChannel.Id)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, 1)
	assert.Equal(t, th.BasicUser.Id, histories[0].UserId)
	assert.Equal(t, privateChannel.Id, histories[0].ChannelId)
}

func TestUpdateChannelPrivacy(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	privateChannel := th.createChannel(th.BasicTeam, model.CHANNEL_PRIVATE)
	privateChannel.Type = model.CHANNEL_OPEN

	if publicChannel, err := th.App.UpdateChannelPrivacy(privateChannel, th.BasicUser); err != nil {
		t.Fatal("Failed to update channel privacy. Error: " + err.Error())
	} else {
		assert.Equal(t, publicChannel.Id, privateChannel.Id)
		assert.Equal(t, publicChannel.Type, model.CHANNEL_OPEN)
	}
}

func TestCreateGroupChannelCreatesChannelMemberHistoryRecord(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user1 := th.CreateUser()
	user2 := th.CreateUser()

	groupUserIds := make([]string, 0)
	groupUserIds = append(groupUserIds, user1.Id)
	groupUserIds = append(groupUserIds, user2.Id)
	groupUserIds = append(groupUserIds, th.BasicUser.Id)

	if channel, err := th.App.CreateGroupChannel(groupUserIds, th.BasicUser.Id); err != nil {
		t.Fatal("Failed to create group channel. Error: " + err.Message)
	} else {
		// there should be a ChannelMemberHistory record for each user
		histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, channel.Id)).([]*model.ChannelMemberHistoryResult)
		assert.Len(t, histories, 3)

		channelMemberHistoryUserIds := make([]string, 0)
		for _, history := range histories {
			assert.Equal(t, channel.Id, history.ChannelId)
			channelMemberHistoryUserIds = append(channelMemberHistoryUserIds, history.UserId)
		}

		sort.Strings(groupUserIds)
		sort.Strings(channelMemberHistoryUserIds)
		assert.Equal(t, groupUserIds, channelMemberHistoryUserIds)
	}
}

func TestCreateDirectChannelCreatesChannelMemberHistoryRecord(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user1 := th.CreateUser()
	user2 := th.CreateUser()

	if channel, err := th.App.GetOrCreateDirectChannel(user1.Id, user2.Id); err != nil {
		t.Fatal("Failed to create direct channel. Error: " + err.Message)
	} else {
		// there should be a ChannelMemberHistory record for both users
		histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, channel.Id)).([]*model.ChannelMemberHistoryResult)
		assert.Len(t, histories, 2)

		historyId0 := histories[0].UserId
		historyId1 := histories[1].UserId
		switch historyId0 {
		case user1.Id:
			assert.Equal(t, user2.Id, historyId1)
		case user2.Id:
			assert.Equal(t, user1.Id, historyId1)
		default:
			t.Fatal("Unexpected user id " + historyId0 + " in ChannelMemberHistory table")
		}
	}
}

func TestGetDirectChannelCreatesChannelMemberHistoryRecord(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user1 := th.CreateUser()
	user2 := th.CreateUser()

	// this function call implicitly creates a direct channel between the two users if one doesn't already exist
	if channel, err := th.App.GetOrCreateDirectChannel(user1.Id, user2.Id); err != nil {
		t.Fatal("Failed to create direct channel. Error: " + err.Message)
	} else {
		// there should be a ChannelMemberHistory record for both users
		histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, channel.Id)).([]*model.ChannelMemberHistoryResult)
		assert.Len(t, histories, 2)

		historyId0 := histories[0].UserId
		historyId1 := histories[1].UserId
		switch historyId0 {
		case user1.Id:
			assert.Equal(t, user2.Id, historyId1)
		case user2.Id:
			assert.Equal(t, user1.Id, historyId1)
		default:
			t.Fatal("Unexpected user id " + historyId0 + " in ChannelMemberHistory table")
		}
	}
}

func TestAddUserToChannelCreatesChannelMemberHistoryRecord(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// create a user and add it to a channel
	user := th.CreateUser()
	if _, err := th.App.AddTeamMember(th.BasicTeam.Id, user.Id); err != nil {
		t.Fatal("Failed to add user to team. Error: " + err.Message)
	}

	groupUserIds := make([]string, 0)
	groupUserIds = append(groupUserIds, th.BasicUser.Id)
	groupUserIds = append(groupUserIds, user.Id)

	channel := th.createChannel(th.BasicTeam, model.CHANNEL_OPEN)
	if _, err := th.App.AddUserToChannel(user, channel); err != nil {
		t.Fatal("Failed to add user to channel. Error: " + err.Message)
	}

	// there should be a ChannelMemberHistory record for the user
	histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, channel.Id)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, 2)
	channelMemberHistoryUserIds := make([]string, 0)
	for _, history := range histories {
		assert.Equal(t, channel.Id, history.ChannelId)
		channelMemberHistoryUserIds = append(channelMemberHistoryUserIds, history.UserId)
	}
	assert.Equal(t, groupUserIds, channelMemberHistoryUserIds)
}

/*func TestRemoveUserFromChannelUpdatesChannelMemberHistoryRecord(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// a user creates a channel
	publicChannel := th.createChannel(th.BasicTeam, model.CHANNEL_OPEN)
	histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, publicChannel.Id)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, 1)
	assert.Equal(t, th.BasicUser.Id, histories[0].UserId)
	assert.Equal(t, publicChannel.Id, histories[0].ChannelId)
	assert.Nil(t, histories[0].LeaveTime)

	// the user leaves that channel
	if err := th.App.LeaveChannel(publicChannel.Id, th.BasicUser.Id); err != nil {
		t.Fatal("Failed to remove user from channel. Error: " + err.Message)
	}
	histories = store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, publicChannel.Id)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, 1)
	assert.Equal(t, th.BasicUser.Id, histories[0].UserId)
	assert.Equal(t, publicChannel.Id, histories[0].ChannelId)
	assert.NotNil(t, histories[0].LeaveTime)
}*/

func TestAddChannelMemberNoUserRequestor(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// create a user and add it to a channel
	user := th.CreateUser()
	if _, err := th.App.AddTeamMember(th.BasicTeam.Id, user.Id); err != nil {
		t.Fatal("Failed to add user to team. Error: " + err.Message)
	}

	groupUserIds := make([]string, 0)
	groupUserIds = append(groupUserIds, th.BasicUser.Id)
	groupUserIds = append(groupUserIds, user.Id)

	channel := th.createChannel(th.BasicTeam, model.CHANNEL_OPEN)
	userRequestorId := ""
	postRootId := ""
	if _, err := th.App.AddChannelMember(user.Id, channel, userRequestorId, postRootId); err != nil {
		t.Fatal("Failed to add user to channel. Error: " + err.Message)
	}

	// there should be a ChannelMemberHistory record for the user
	histories := store.Must(th.App.Srv.Store.ChannelMemberHistory().GetUsersInChannelDuring(model.GetMillis()-100, model.GetMillis()+100, channel.Id)).([]*model.ChannelMemberHistoryResult)
	assert.Len(t, histories, 2)
	channelMemberHistoryUserIds := make([]string, 0)
	for _, history := range histories {
		assert.Equal(t, channel.Id, history.ChannelId)
		channelMemberHistoryUserIds = append(channelMemberHistoryUserIds, history.UserId)
	}
	assert.Equal(t, groupUserIds, channelMemberHistoryUserIds)

	postList := store.Must(th.App.Srv.Store.Post().GetPosts(channel.Id, 0, 1, false)).(*model.PostList)
	if assert.Len(t, postList.Order, 1) {
		post := postList.Posts[postList.Order[0]]

		assert.Equal(t, model.POST_JOIN_CHANNEL, post.Type)
		assert.Equal(t, user.Id, post.UserId)
		assert.Equal(t, user.Username, post.Props["username"])
	}
}

func TestAppUpdateChannelScheme(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	channel := th.BasicChannel
	mockID := model.NewString("x")
	channel.SchemeId = mockID

	updatedChannel, err := th.App.UpdateChannelScheme(channel)
	if err != nil {
		t.Fatal(err)
	}

	if updatedChannel.SchemeId != mockID {
		t.Fatal("Wrong Channel SchemeId")
	}
}

func TestFillInChannelProps(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	channelPublic1, err := th.App.CreateChannel(&model.Channel{DisplayName: "Public 1", Name: "public1", Type: model.CHANNEL_OPEN, TeamId: th.BasicTeam.Id}, false)
	require.Nil(t, err)
	defer th.App.PermanentDeleteChannel(channelPublic1)

	channelPublic2, err := th.App.CreateChannel(&model.Channel{DisplayName: "Public 2", Name: "public2", Type: model.CHANNEL_OPEN, TeamId: th.BasicTeam.Id}, false)
	require.Nil(t, err)
	defer th.App.PermanentDeleteChannel(channelPublic2)

	channelPrivate, err := th.App.CreateChannel(&model.Channel{DisplayName: "Private", Name: "private", Type: model.CHANNEL_PRIVATE, TeamId: th.BasicTeam.Id}, false)
	require.Nil(t, err)
	defer th.App.PermanentDeleteChannel(channelPrivate)

	otherTeamId := model.NewId()
	otherTeam := &model.Team{
		DisplayName: "dn_" + otherTeamId,
		Name:        "name" + otherTeamId,
		Email:       "success+" + otherTeamId + "@simulator.amazonses.com",
		Type:        model.TEAM_OPEN,
	}
	otherTeam, err = th.App.CreateTeam(otherTeam)
	require.Nil(t, err)
	defer th.App.PermanentDeleteTeam(otherTeam)

	channelOtherTeam, err := th.App.CreateChannel(&model.Channel{DisplayName: "Other Team Channel", Name: "other-team", Type: model.CHANNEL_OPEN, TeamId: otherTeam.Id}, false)
	require.Nil(t, err)
	defer th.App.PermanentDeleteChannel(channelOtherTeam)

	// Note that purpose is intentionally plaintext below.

	t.Run("single channels", func(t *testing.T) {
		testCases := []struct {
			Description          string
			Channel              *model.Channel
			ExpectedChannelProps map[string]interface{}
		}{
			{
				"channel on basic team without references",
				&model.Channel{
					TeamId:  th.BasicTeam.Id,
					Header:  "No references",
					Purpose: "No references",
				},
				nil,
			},
			{
				"channel on basic team",
				&model.Channel{
					TeamId:  th.BasicTeam.Id,
					Header:  "~public1, ~private, ~other-team",
					Purpose: "~public2, ~private, ~other-team",
				},
				map[string]interface{}{
					"channel_mentions": map[string]interface{}{
						"public1": map[string]interface{}{
							"display_name": "Public 1",
						},
					},
				},
			},
			{
				"channel on other team",
				&model.Channel{
					TeamId:  otherTeam.Id,
					Header:  "~public1, ~private, ~other-team",
					Purpose: "~public2, ~private, ~other-team",
				},
				map[string]interface{}{
					"channel_mentions": map[string]interface{}{
						"other-team": map[string]interface{}{
							"display_name": "Other Team Channel",
						},
					},
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Description, func(t *testing.T) {
				err = th.App.FillInChannelProps(testCase.Channel)
				require.Nil(t, err)

				assert.Equal(t, testCase.ExpectedChannelProps, testCase.Channel.Props)
			})
		}
	})

	t.Run("multiple channels", func(t *testing.T) {
		testCases := []struct {
			Description          string
			Channels             *model.ChannelList
			ExpectedChannelProps map[string]interface{}
		}{
			{
				"single channel on basic team",
				&model.ChannelList{
					{
						Name:    "test",
						TeamId:  th.BasicTeam.Id,
						Header:  "~public1, ~private, ~other-team",
						Purpose: "~public2, ~private, ~other-team",
					},
				},
				map[string]interface{}{
					"test": map[string]interface{}{
						"channel_mentions": map[string]interface{}{
							"public1": map[string]interface{}{
								"display_name": "Public 1",
							},
						},
					},
				},
			},
			{
				"multiple channels on basic team",
				&model.ChannelList{
					{
						Name:    "test",
						TeamId:  th.BasicTeam.Id,
						Header:  "~public1, ~private, ~other-team",
						Purpose: "~public2, ~private, ~other-team",
					},
					{
						Name:    "test2",
						TeamId:  th.BasicTeam.Id,
						Header:  "~private, ~other-team",
						Purpose: "~public2, ~private, ~other-team",
					},
					{
						Name:    "test3",
						TeamId:  th.BasicTeam.Id,
						Header:  "No references",
						Purpose: "No references",
					},
				},
				map[string]interface{}{
					"test": map[string]interface{}{
						"channel_mentions": map[string]interface{}{
							"public1": map[string]interface{}{
								"display_name": "Public 1",
							},
						},
					},
					"test2": map[string]interface{}(nil),
					"test3": map[string]interface{}(nil),
				},
			},
			{
				"multiple channels across teams",
				&model.ChannelList{
					{
						Name:    "test",
						TeamId:  th.BasicTeam.Id,
						Header:  "~public1, ~private, ~other-team",
						Purpose: "~public2, ~private, ~other-team",
					},
					{
						Name:    "test2",
						TeamId:  otherTeam.Id,
						Header:  "~private, ~other-team",
						Purpose: "~public2, ~private, ~other-team",
					},
					{
						Name:    "test3",
						TeamId:  th.BasicTeam.Id,
						Header:  "No references",
						Purpose: "No references",
					},
				},
				map[string]interface{}{
					"test": map[string]interface{}{
						"channel_mentions": map[string]interface{}{
							"public1": map[string]interface{}{
								"display_name": "Public 1",
							},
						},
					},
					"test2": map[string]interface{}{
						"channel_mentions": map[string]interface{}{
							"other-team": map[string]interface{}{
								"display_name": "Other Team Channel",
							},
						},
					},
					"test3": map[string]interface{}(nil),
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Description, func(t *testing.T) {
				err = th.App.FillInChannelsProps(testCase.Channels)
				require.Nil(t, err)

				for _, channel := range *testCase.Channels {
					assert.Equal(t, testCase.ExpectedChannelProps[channel.Name], channel.Props)
				}
			})
		}
	})
}

func TestRenameChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	testCases := []struct {
		Name                string
		Channel             *model.Channel
		ExpectError         bool
		ExpectedName        string
		ExpectedDisplayName string
	}{
		{
			"Rename open channel",
			th.createChannel(th.BasicTeam, model.CHANNEL_OPEN),
			false,
			"newchannelname",
			"New Display Name",
		},
		{
			"Fail on rename direct message channel",
			th.CreateDmChannel(th.BasicUser2),
			true,
			"",
			"",
		},
		{
			"Fail on rename direct message channel",
			th.CreateGroupChannel(th.BasicUser2, th.CreateUser()),
			true,
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			channel, err := th.App.RenameChannel(tc.Channel, "newchannelname", "New Display Name")
			if tc.ExpectError {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tc.ExpectedName, channel.Name)
				assert.Equal(t, tc.ExpectedDisplayName, channel.DisplayName)
			}
		})
	}
}

func TestGetChannelMembersTimezones(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	userRequestorId := ""
	postRootId := ""
	if _, err := th.App.AddChannelMember(th.BasicUser2.Id, th.BasicChannel, userRequestorId, postRootId); err != nil {
		t.Fatal("Failed to add user to channel. Error: " + err.Message)
	}

	user := th.BasicUser
	user.Timezone["useAutomaticTimezone"] = "false"
	user.Timezone["manualTimezone"] = "XOXO/BLABLA"
	th.App.UpdateUser(user, false)

	user2 := th.BasicUser2
	user2.Timezone["automaticTimezone"] = "NoWhere/Island"
	th.App.UpdateUser(user2, false)

	user3 := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
	ruser, _ := th.App.CreateUser(&user3)
	th.App.AddUserToChannel(ruser, th.BasicChannel)

	ruser.Timezone["automaticTimezone"] = "NoWhere/Island"
	th.App.UpdateUser(ruser, false)

	user4 := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
	ruser, _ = th.App.CreateUser(&user4)
	th.App.AddUserToChannel(ruser, th.BasicChannel)

	timezones, err := th.App.GetChannelMembersTimezones(th.BasicChannel.Id)
	if err != nil {
		t.Fatal("Failed to get the timezones for a channel. Error: " + err.Error())
	}
	assert.Equal(t, 2, len(timezones))
}

func TestGetPublicChannelsForTeam(t *testing.T) {
	th := Setup(t)
	team := th.CreateTeam()
	defer th.TearDown()

	var expectedChannels []*model.Channel

	townSquare, err := th.App.GetChannelByName("town-square", team.Id, false)
	require.Nil(t, err)
	require.NotNil(t, townSquare)
	expectedChannels = append(expectedChannels, townSquare)

	offTopic, err := th.App.GetChannelByName("off-topic", team.Id, false)
	require.Nil(t, err)
	require.NotNil(t, offTopic)
	expectedChannels = append(expectedChannels, offTopic)

	for i := 0; i < 8; i++ {
		channel := model.Channel{
			DisplayName: fmt.Sprintf("Public %v", i),
			Name:        fmt.Sprintf("public_%v", i),
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		var rchannel *model.Channel
		rchannel, err = th.App.CreateChannel(&channel, false)
		require.Nil(t, err)
		require.NotNil(t, rchannel)
		defer th.App.PermanentDeleteChannel(rchannel)

		// Store the user ids for comparison later
		expectedChannels = append(expectedChannels, rchannel)
	}

	// Fetch public channels multipile times
	channelList, err := th.App.GetPublicChannelsForTeam(team.Id, 0, 5)
	require.Nil(t, err)
	channelList2, err := th.App.GetPublicChannelsForTeam(team.Id, 5, 5)
	require.Nil(t, err)

	channels := append(*channelList, *channelList2...)
	assert.ElementsMatch(t, expectedChannels, channels)
}

func TestUpdateChannelMemberRolesChangingGuest(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	t.Run("from guest to user", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser, _ := th.App.CreateGuest(&user)

		_, err := th.App.AddUserToTeam(th.BasicTeam.Id, ruser.Id, "")
		require.Nil(t, err)

		_, err = th.App.AddUserToChannel(ruser, th.BasicChannel)
		require.Nil(t, err)

		if _, err := th.App.UpdateChannelMemberRoles(th.BasicChannel.Id, ruser.Id, "channel_user"); err == nil {
			t.Fatal("Should fail when try to modify the guest role")
		}
	})

	t.Run("from user to guest", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser, _ := th.App.CreateUser(&user)

		_, err := th.App.AddUserToTeam(th.BasicTeam.Id, ruser.Id, "")
		require.Nil(t, err)

		_, err = th.App.AddUserToChannel(ruser, th.BasicChannel)
		require.Nil(t, err)

		if _, err := th.App.UpdateChannelMemberRoles(th.BasicChannel.Id, ruser.Id, "channel_guest"); err == nil {
			t.Fatal("Should fail when try to modify the guest role")
		}
	})

	t.Run("from user to admin", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser, _ := th.App.CreateUser(&user)

		_, err := th.App.AddUserToTeam(th.BasicTeam.Id, ruser.Id, "")
		require.Nil(t, err)

		_, err = th.App.AddUserToChannel(ruser, th.BasicChannel)
		require.Nil(t, err)

		if _, err := th.App.UpdateChannelMemberRoles(th.BasicChannel.Id, ruser.Id, "channel_user channel_admin"); err != nil {
			t.Fatal("Should work when you not modify guest role")
		}
	})

	t.Run("from guest to guest plus custom", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser, _ := th.App.CreateGuest(&user)

		_, err := th.App.AddUserToTeam(th.BasicTeam.Id, ruser.Id, "")
		require.Nil(t, err)

		_, err = th.App.AddUserToChannel(ruser, th.BasicChannel)
		require.Nil(t, err)

		_, err = th.App.CreateRole(&model.Role{Name: "custom", DisplayName: "custom", Description: "custom"})
		require.Nil(t, err)

		if _, err := th.App.UpdateChannelMemberRoles(th.BasicChannel.Id, ruser.Id, "channel_guest custom"); err != nil {
			t.Fatal("Should work when you not modify guest role")
		}
	})

	t.Run("a guest cant have user role", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser, _ := th.App.CreateGuest(&user)

		_, err := th.App.AddUserToTeam(th.BasicTeam.Id, ruser.Id, "")
		require.Nil(t, err)

		_, err = th.App.AddUserToChannel(ruser, th.BasicChannel)
		require.Nil(t, err)

		if _, err := th.App.UpdateChannelMemberRoles(th.BasicChannel.Id, ruser.Id, "channel_guest channel_user"); err == nil {
			t.Fatal("Should work when you not modify guest role")
		}
	})
}

func TestDefaultChannelNames(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	actual := th.App.DefaultChannelNames()
	expect := []string{"town-square", "off-topic"}
	require.ElementsMatch(t, expect, actual)

	th.App.UpdateConfig(func(cfg *model.Config) {
		cfg.TeamSettings.ExperimentalDefaultChannels = []string{"foo", "bar"}
	})

	actual = th.App.DefaultChannelNames()
	expect = []string{"town-square", "foo", "bar"}
	require.ElementsMatch(t, expect, actual)
}
