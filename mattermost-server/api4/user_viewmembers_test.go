package api4

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/require"
)

func TestApiResctrictedViewMembers(t *testing.T) {
	th := Setup()
	defer th.TearDown()

	// Create first account for system admin
	_, err := th.App.CreateUser(&model.User{Email: th.GenerateTestEmail(), Nickname: "test user0", Password: "test-password-0", Username: "test-user-0", Roles: model.SYSTEM_USER_ROLE_ID})
	require.Nil(t, err)

	user1, err := th.App.CreateUser(&model.User{Email: th.GenerateTestEmail(), Nickname: "test user1", Password: "test-password-1", Username: "test-user-1", Roles: model.SYSTEM_USER_ROLE_ID})
	require.Nil(t, err)
	user2, err := th.App.CreateUser(&model.User{Email: th.GenerateTestEmail(), Nickname: "test user2", Password: "test-password-2", Username: "test-user-2", Roles: model.SYSTEM_USER_ROLE_ID})
	require.Nil(t, err)
	user3, err := th.App.CreateUser(&model.User{Email: th.GenerateTestEmail(), Nickname: "test user3", Password: "test-password-3", Username: "test-user-3", Roles: model.SYSTEM_USER_ROLE_ID})
	require.Nil(t, err)
	user4, err := th.App.CreateUser(&model.User{Email: th.GenerateTestEmail(), Nickname: "test user4", Password: "test-password-4", Username: "test-user-4", Roles: model.SYSTEM_USER_ROLE_ID})
	require.Nil(t, err)
	user5, err := th.App.CreateUser(&model.User{Email: th.GenerateTestEmail(), Nickname: "test user5", Password: "test-password-5", Username: "test-user-5", Roles: model.SYSTEM_USER_ROLE_ID})
	require.Nil(t, err)

	team1, err := th.App.CreateTeam(&model.Team{DisplayName: "dn_" + model.NewId(), Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN})
	require.Nil(t, err)
	team2, err := th.App.CreateTeam(&model.Team{DisplayName: "dn_" + model.NewId(), Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN})
	require.Nil(t, err)

	channel1, err := th.App.CreateChannel(&model.Channel{DisplayName: "dn_" + model.NewId(), Name: "name_" + model.NewId(), Type: model.CHANNEL_OPEN, TeamId: team1.Id, CreatorId: model.NewId()}, false)
	require.Nil(t, err)
	channel2, err := th.App.CreateChannel(&model.Channel{DisplayName: "dn_" + model.NewId(), Name: "name_" + model.NewId(), Type: model.CHANNEL_OPEN, TeamId: team1.Id, CreatorId: model.NewId()}, false)
	require.Nil(t, err)
	channel3, err := th.App.CreateChannel(&model.Channel{DisplayName: "dn_" + model.NewId(), Name: "name_" + model.NewId(), Type: model.CHANNEL_OPEN, TeamId: team2.Id, CreatorId: model.NewId()}, false)
	require.Nil(t, err)

	th.LinkUserToTeam(user1, team1)
	th.LinkUserToTeam(user2, team1)
	th.LinkUserToTeam(user3, team2)
	th.LinkUserToTeam(user4, team1)
	th.LinkUserToTeam(user4, team2)

	th.AddUserToChannel(user1, channel1)
	th.AddUserToChannel(user2, channel2)
	th.AddUserToChannel(user3, channel3)
	th.AddUserToChannel(user4, channel1)
	th.AddUserToChannel(user4, channel3)

	th.App.SetStatusOnline(user1.Id, true)
	th.App.SetStatusOnline(user2.Id, true)
	th.App.SetStatusOnline(user3.Id, true)
	th.App.SetStatusOnline(user4.Id, true)
	th.App.SetStatusOnline(user5.Id, true)

	_, resp := th.Client.Login(user1.Username, "test-password-1")
	CheckNoError(t, resp)

	t.Run("getUser", func(t *testing.T) {
		testCases := []struct {
			Name          string
			RestrictedTo  string
			UserId        string
			ExpectedError string
		}{
			{
				"Get visible user without restrictions",
				"",
				user5.Id,
				"",
			},
			{
				"Get not existing user without restrictions",
				"",
				model.NewId(),
				"store.sql_user.missing_account.const",
			},
			{
				"Get not existing user with restrictions to teams",
				"teams",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to teams",
				"teams",
				user2.Id,
				"",
			},
			{
				"Get not visible user with restrictions to teams",
				"teams",
				user5.Id,
				"api.context.permissions.app_error",
			},
			{
				"Get not existing user with restrictions to channels",
				"channels",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to channels",
				"channels",
				user4.Id,
				"",
			},
			{
				"Get not visible user with restrictions to channels",
				"channels",
				user3.Id,
				"api.context.permissions.app_error",
			},
		}
		defer th.RestoreDefaultRolePermissions(th.SaveDefaultRolePermissions())

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				if tc.RestrictedTo == "channels" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else if tc.RestrictedTo == "teams" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
				}

				_, resp := th.Client.GetUser(tc.UserId, "")
				require.Nil(t, err)
				if tc.ExpectedError != "" {
					CheckErrorMessage(t, resp, tc.ExpectedError)
				} else {
					CheckNoError(t, resp)
				}
			})
		}
	})

	t.Run("getUserByUsername", func(t *testing.T) {
		testCases := []struct {
			Name          string
			RestrictedTo  string
			Username      string
			ExpectedError string
		}{
			{
				"Get visible user without restrictions",
				"",
				user5.Username,
				"",
			},
			{
				"Get not existing user without restrictions",
				"",
				model.NewId(),
				"store.sql_user.get_by_username.app_error",
			},
			{
				"Get not existing user with restrictions to teams",
				"teams",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to teams",
				"teams",
				user2.Username,
				"",
			},
			{
				"Get not visible user with restrictions to teams",
				"teams",
				user5.Username,
				"api.context.permissions.app_error",
			},
			{
				"Get not existing user with restrictions to channels",
				"channels",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to channels",
				"channels",
				user4.Username,
				"",
			},
			{
				"Get not visible user with restrictions to channels",
				"channels",
				user3.Username,
				"api.context.permissions.app_error",
			},
		}
		defer th.RestoreDefaultRolePermissions(th.SaveDefaultRolePermissions())

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				if tc.RestrictedTo == "channels" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else if tc.RestrictedTo == "teams" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
				}

				_, resp := th.Client.GetUserByUsername(tc.Username, "")
				require.Nil(t, err)
				if tc.ExpectedError != "" {
					CheckErrorMessage(t, resp, tc.ExpectedError)
				} else {
					CheckNoError(t, resp)
				}
			})
		}
	})

	t.Run("getUserByEmail", func(t *testing.T) {
		testCases := []struct {
			Name          string
			RestrictedTo  string
			Email         string
			ExpectedError string
		}{
			{
				"Get visible user without restrictions",
				"",
				user5.Email,
				"",
			},
			{
				"Get not existing user without restrictions",
				"",
				th.GenerateTestEmail(),
				"store.sql_user.missing_account.const",
			},
			{
				"Get not existing user with restrictions to teams",
				"teams",
				th.GenerateTestEmail(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to teams",
				"teams",
				user2.Email,
				"",
			},
			{
				"Get not visible user with restrictions to teams",
				"teams",
				user5.Email,
				"api.context.permissions.app_error",
			},
			{
				"Get not existing user with restrictions to channels",
				"channels",
				th.GenerateTestEmail(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to channels",
				"channels",
				user4.Email,
				"",
			},
			{
				"Get not visible user with restrictions to channels",
				"channels",
				user3.Email,
				"api.context.permissions.app_error",
			},
		}
		defer th.RestoreDefaultRolePermissions(th.SaveDefaultRolePermissions())

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				if tc.RestrictedTo == "channels" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else if tc.RestrictedTo == "teams" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
				}

				_, resp := th.Client.GetUserByEmail(tc.Email, "")
				require.Nil(t, err)
				if tc.ExpectedError != "" {
					CheckErrorMessage(t, resp, tc.ExpectedError)
				} else {
					CheckNoError(t, resp)
				}
			})
		}
	})

	t.Run("getDefaultProfileImage", func(t *testing.T) {
		testCases := []struct {
			Name          string
			RestrictedTo  string
			UserId        string
			ExpectedError string
		}{
			{
				"Get visible user without restrictions",
				"",
				user5.Id,
				"",
			},
			{
				"Get not existing user without restrictions",
				"",
				model.NewId(),
				"store.sql_user.missing_account.const",
			},
			{
				"Get not existing user with restrictions to teams",
				"teams",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to teams",
				"teams",
				user2.Id,
				"",
			},
			{
				"Get not visible user with restrictions to teams",
				"teams",
				user5.Id,
				"api.context.permissions.app_error",
			},
			{
				"Get not existing user with restrictions to channels",
				"channels",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to channels",
				"channels",
				user4.Id,
				"",
			},
			{
				"Get not visible user with restrictions to channels",
				"channels",
				user3.Id,
				"api.context.permissions.app_error",
			},
		}
		defer th.RestoreDefaultRolePermissions(th.SaveDefaultRolePermissions())

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				if tc.RestrictedTo == "channels" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else if tc.RestrictedTo == "teams" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
				}

				_, resp := th.Client.GetDefaultProfileImage(tc.UserId)
				require.Nil(t, err)
				if tc.ExpectedError != "" {
					CheckErrorMessage(t, resp, tc.ExpectedError)
				} else {
					CheckNoError(t, resp)
				}
			})
		}
	})

	t.Run("getProfileImage", func(t *testing.T) {
		testCases := []struct {
			Name          string
			RestrictedTo  string
			UserId        string
			ExpectedError string
		}{
			{
				"Get visible user without restrictions",
				"",
				user5.Id,
				"",
			},
			{
				"Get not existing user without restrictions",
				"",
				model.NewId(),
				"store.sql_user.missing_account.const",
			},
			{
				"Get not existing user with restrictions to teams",
				"teams",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to teams",
				"teams",
				user2.Id,
				"",
			},
			{
				"Get not visible user with restrictions to teams",
				"teams",
				user5.Id,
				"api.context.permissions.app_error",
			},
			{
				"Get not existing user with restrictions to channels",
				"channels",
				model.NewId(),
				"api.context.permissions.app_error",
			},
			{
				"Get visible user with restrictions to channels",
				"channels",
				user4.Id,
				"",
			},
			{
				"Get not visible user with restrictions to channels",
				"channels",
				user3.Id,
				"api.context.permissions.app_error",
			},
		}
		defer th.RestoreDefaultRolePermissions(th.SaveDefaultRolePermissions())

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				if tc.RestrictedTo == "channels" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else if tc.RestrictedTo == "teams" {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
				} else {
					th.RemovePermissionFromRole(model.PERMISSION_VIEW_MEMBERS.Id, model.TEAM_USER_ROLE_ID)
					th.AddPermissionToRole(model.PERMISSION_VIEW_MEMBERS.Id, model.SYSTEM_USER_ROLE_ID)
				}

				_, resp := th.Client.GetProfileImage(tc.UserId, "")
				require.Nil(t, err)
				if tc.ExpectedError != "" {
					CheckErrorMessage(t, resp, tc.ExpectedError)
				} else {
					CheckNoError(t, resp)
				}
			})
		}
	})
}
