// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api4

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"encoding/base64"

	"github.com/mattermost/mattermost-server/app"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/services/mailservice"
	"github.com/mattermost/mattermost-server/utils"
	"github.com/mattermost/mattermost-server/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team := &model.Team{Name: GenerateTestUsername(), DisplayName: "Some Team", Type: model.TEAM_OPEN}
	rteam, resp := Client.CreateTeam(team)
	CheckNoError(t, resp)
	CheckCreatedStatus(t, resp)

	if rteam.Name != team.Name {
		t.Fatal("names did not match")
	}

	if rteam.DisplayName != team.DisplayName {
		t.Fatal("display names did not match")
	}

	if rteam.Type != team.Type {
		t.Fatal("types did not match")
	}

	_, resp = Client.CreateTeam(rteam)
	CheckBadRequestStatus(t, resp)

	rteam.Id = ""
	_, resp = Client.CreateTeam(rteam)
	CheckErrorMessage(t, resp, "store.sql_team.save.domain_exists.app_error")
	CheckBadRequestStatus(t, resp)

	rteam.Name = ""
	_, resp = Client.CreateTeam(rteam)
	CheckErrorMessage(t, resp, "model.team.is_valid.characters.app_error")
	CheckBadRequestStatus(t, resp)

	if r, err := Client.DoApiPost("/teams", "garbage"); err == nil {
		t.Fatal("should have errored")
	} else {
		if r.StatusCode != http.StatusBadRequest {
			t.Log("actual: " + strconv.Itoa(r.StatusCode))
			t.Log("expected: " + strconv.Itoa(http.StatusBadRequest))
			t.Fatal("wrong status code")
		}
	}

	Client.Logout()

	_, resp = Client.CreateTeam(rteam)
	CheckUnauthorizedStatus(t, resp)

	th.LoginBasic()

	// Test GroupConstrained flag
	groupConstrainedTeam := &model.Team{Name: GenerateTestUsername(), DisplayName: "Some Team", Type: model.TEAM_OPEN, GroupConstrained: model.NewBool(true)}
	rteam, resp = Client.CreateTeam(groupConstrainedTeam)
	CheckNoError(t, resp)
	CheckCreatedStatus(t, resp)

	if *rteam.GroupConstrained != *groupConstrainedTeam.GroupConstrained {
		t.Fatal("GroupConstrained flags do not match")
	}

	// Check the appropriate permissions are enforced.
	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()

	th.RemovePermissionFromRole(model.PERMISSION_CREATE_TEAM.Id, model.SYSTEM_USER_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_CREATE_TEAM.Id, model.SYSTEM_ADMIN_ROLE_ID)

	_, resp = Client.CreateTeam(team)
	CheckForbiddenStatus(t, resp)
}

func TestCreateTeamSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	// Non-admin users can create a team, but they become a team admin by doing so

	t.Run("team admin", func(t *testing.T) {
		team := &model.Team{
			DisplayName:    t.Name() + "_1",
			Name:           GenerateTestTeamName(),
			Email:          th.GenerateTestEmail(),
			Type:           model.TEAM_OPEN,
			AllowedDomains: "simulator.amazonses.com,dockerhost",
		}

		rteam, resp := th.Client.CreateTeam(team)
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email")
		}
	})

	t.Run("system admin", func(t *testing.T) {
		team := &model.Team{
			DisplayName:    t.Name() + "_2",
			Name:           GenerateTestTeamName(),
			Email:          th.GenerateTestEmail(),
			Type:           model.TEAM_OPEN,
			AllowedDomains: "simulator.amazonses.com,dockerhost",
		}

		rteam, resp := th.SystemAdminClient.CreateTeam(team)
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email")
		}
	})
}

func TestGetTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	rteam, resp := Client.GetTeam(team.Id, "")
	CheckNoError(t, resp)

	if rteam.Id != team.Id {
		t.Fatal("wrong team")
	}

	_, resp = Client.GetTeam("junk", "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeam("", "")
	CheckNotFoundStatus(t, resp)

	_, resp = Client.GetTeam(model.NewId(), "")
	CheckNotFoundStatus(t, resp)

	th.LoginTeamAdmin()

	team2 := &model.Team{DisplayName: "Name", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN, AllowOpenInvite: false}
	rteam2, _ := Client.CreateTeam(team2)

	team3 := &model.Team{DisplayName: "Name", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_INVITE, AllowOpenInvite: true}
	rteam3, _ := Client.CreateTeam(team3)

	th.LoginBasic()
	// AllowInviteOpen is false and team is open, and user is not on team
	_, resp = Client.GetTeam(rteam2.Id, "")
	CheckForbiddenStatus(t, resp)

	// AllowInviteOpen is true and team is invite, and user is not on team
	_, resp = Client.GetTeam(rteam3.Id, "")
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetTeam(team.Id, "")
	CheckUnauthorizedStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeam(rteam2.Id, "")
	CheckNoError(t, resp)
}

func TestGetTeamSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_1",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)

	t.Run("team user", func(t *testing.T) {
		th.LinkUserToTeam(th.BasicUser2, team)

		client := th.CreateClient()
		th.LoginBasic2WithClient(client)

		rteam, resp := client.GetTeam(team.Id, "")
		CheckNoError(t, resp)
		if rteam.Email != "" {
			t.Fatal("should've sanitized email")
		}
	})

	t.Run("team admin", func(t *testing.T) {
		rteam, resp := th.Client.GetTeam(team.Id, "")
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email")
		}
	})

	t.Run("system admin", func(t *testing.T) {
		rteam, resp := th.SystemAdminClient.GetTeam(team.Id, "")
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email")
		}
	})
}

func TestGetTeamUnread(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	teamUnread, resp := Client.GetTeamUnread(th.BasicTeam.Id, th.BasicUser.Id)
	CheckNoError(t, resp)
	if teamUnread.TeamId != th.BasicTeam.Id {
		t.Fatal("wrong team id returned for regular user call")
	}

	_, resp = Client.GetTeamUnread("junk", th.BasicUser.Id)
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamUnread(th.BasicTeam.Id, "junk")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamUnread(model.NewId(), th.BasicUser.Id)
	CheckForbiddenStatus(t, resp)

	_, resp = Client.GetTeamUnread(th.BasicTeam.Id, model.NewId())
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetTeamUnread(th.BasicTeam.Id, th.BasicUser.Id)
	CheckUnauthorizedStatus(t, resp)

	teamUnread, resp = th.SystemAdminClient.GetTeamUnread(th.BasicTeam.Id, th.BasicUser.Id)
	CheckNoError(t, resp)
	if teamUnread.TeamId != th.BasicTeam.Id {
		t.Fatal("wrong team id returned")
	}
}

func TestUpdateTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team := &model.Team{DisplayName: "Name", Description: "Some description", AllowOpenInvite: false, InviteId: "inviteid0", Name: "z-z-" + model.NewId() + "a", Email: "success+" + model.NewId() + "@simulator.amazonses.com", Type: model.TEAM_OPEN}
	team, _ = Client.CreateTeam(team)

	team.Description = "updated description"
	uteam, resp := Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if uteam.Description != "updated description" {
		t.Fatal("Update failed")
	}

	team.DisplayName = "Updated Name"
	uteam, resp = Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if uteam.DisplayName != "Updated Name" {
		t.Fatal("Update failed")
	}

	// Test GroupConstrained flag
	team.GroupConstrained = model.NewBool(true)
	rteam, resp := Client.UpdateTeam(team)
	CheckNoError(t, resp)
	CheckOKStatus(t, resp)

	if *rteam.GroupConstrained != *team.GroupConstrained {
		t.Fatal("GroupConstrained flags do not match")
	}
	team.GroupConstrained = nil

	team.AllowOpenInvite = true
	uteam, resp = Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if !uteam.AllowOpenInvite {
		t.Fatal("Update failed")
	}

	team.InviteId = "inviteid1"
	uteam, resp = Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if uteam.InviteId == "inviteid1" {
		t.Fatal("InviteID should not be updated")
	}

	team.AllowedDomains = "domain"
	uteam, resp = Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if uteam.AllowedDomains != "domain" {
		t.Fatal("Update failed")
	}

	team.Name = "Updated name"
	uteam, resp = Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if uteam.Name == "Updated name" {
		t.Fatal("Should not update name")
	}

	team.Email = "test@domain.com"
	uteam, resp = Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if uteam.Email == "test@domain.com" {
		t.Fatal("Should not update email")
	}

	team.Type = model.TEAM_INVITE
	uteam, resp = Client.UpdateTeam(team)
	CheckNoError(t, resp)

	if uteam.Type == model.TEAM_INVITE {
		t.Fatal("Should not update type")
	}

	originalTeamId := team.Id
	team.Id = model.NewId()

	r, _ := Client.DoApiPut(Client.GetTeamRoute(originalTeamId), team.ToJson())
	assert.Equal(t, http.StatusBadRequest, r.StatusCode)

	if uteam.Id != originalTeamId {
		t.Fatal("wrong team id")
	}

	team.Id = "fake"
	_, resp = Client.UpdateTeam(team)
	CheckBadRequestStatus(t, resp)

	Client.Logout()
	_, resp = Client.UpdateTeam(team)
	CheckUnauthorizedStatus(t, resp)

	team.Id = originalTeamId
	_, resp = th.SystemAdminClient.UpdateTeam(team)
	CheckNoError(t, resp)
}

func TestUpdateTeamSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_1",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)

	// Non-admin users cannot update the team

	t.Run("team admin", func(t *testing.T) {
		rteam, resp := th.Client.UpdateTeam(team)
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email for admin")
		}
	})

	t.Run("system admin", func(t *testing.T) {
		rteam, resp := th.SystemAdminClient.UpdateTeam(team)
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email for admin")
		}
	})
}

func TestPatchTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team := &model.Team{DisplayName: "Name", Description: "Some description", CompanyName: "Some company name", AllowOpenInvite: false, InviteId: "inviteid0", Name: "z-z-" + model.NewId() + "a", Email: "success+" + model.NewId() + "@simulator.amazonses.com", Type: model.TEAM_OPEN}
	team, _ = Client.CreateTeam(team)

	patch := &model.TeamPatch{}

	patch.DisplayName = model.NewString("Other name")
	patch.Description = model.NewString("Other description")
	patch.CompanyName = model.NewString("Other company name")
	patch.AllowOpenInvite = model.NewBool(true)

	rteam, resp := Client.PatchTeam(team.Id, patch)
	CheckNoError(t, resp)

	if rteam.DisplayName != "Other name" {
		t.Fatal("DisplayName did not update properly")
	}
	if rteam.Description != "Other description" {
		t.Fatal("Description did not update properly")
	}
	if rteam.CompanyName != "Other company name" {
		t.Fatal("CompanyName did not update properly")
	}
	if rteam.InviteId == "inviteid1" {
		t.Fatal("InviteId should not update")
	}
	if !rteam.AllowOpenInvite {
		t.Fatal("AllowOpenInvite did not update properly")
	}

	// Test GroupConstrained flag
	patch.GroupConstrained = model.NewBool(true)
	rteam, resp = Client.PatchTeam(team.Id, patch)
	CheckNoError(t, resp)
	CheckOKStatus(t, resp)

	if *rteam.GroupConstrained != *patch.GroupConstrained {
		t.Fatal("GroupConstrained flags do not match")
	}
	patch.GroupConstrained = nil

	_, resp = Client.PatchTeam("junk", patch)
	CheckBadRequestStatus(t, resp)

	_, resp = Client.PatchTeam(GenerateTestId(), patch)
	CheckForbiddenStatus(t, resp)

	if r, err := Client.DoApiPut("/teams/"+team.Id+"/patch", "garbage"); err == nil {
		t.Fatal("should have errored")
	} else {
		if r.StatusCode != http.StatusBadRequest {
			t.Log("actual: " + strconv.Itoa(r.StatusCode))
			t.Log("expected: " + strconv.Itoa(http.StatusBadRequest))
			t.Fatal("wrong status code")
		}
	}

	Client.Logout()
	_, resp = Client.PatchTeam(team.Id, patch)
	CheckUnauthorizedStatus(t, resp)

	th.LoginBasic2()
	_, resp = Client.PatchTeam(team.Id, patch)
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.PatchTeam(team.Id, patch)
	CheckNoError(t, resp)
}

func TestPatchTeamSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_1",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)

	// Non-admin users cannot update the team

	t.Run("team admin", func(t *testing.T) {
		rteam, resp := th.Client.PatchTeam(team.Id, &model.TeamPatch{})
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email for admin")
		}
	})

	t.Run("system admin", func(t *testing.T) {
		rteam, resp := th.SystemAdminClient.PatchTeam(team.Id, &model.TeamPatch{})
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email for admin")
		}
	})
}

func TestRegenerateTeamInviteId(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team := &model.Team{DisplayName: "Name", Description: "Some description", CompanyName: "Some company name", AllowOpenInvite: false, InviteId: "inviteid0", Name: "z-z-" + model.NewId() + "a", Email: "success+" + model.NewId() + "@simulator.amazonses.com", Type: model.TEAM_OPEN}
	team, _ = Client.CreateTeam(team)

	assert.NotEqual(t, team.InviteId, "")
	assert.NotEqual(t, team.InviteId, "inviteid0")

	rteam, resp := Client.RegenerateTeamInviteId(team.Id)
	CheckNoError(t, resp)

	assert.NotEqual(t, team.InviteId, rteam.InviteId)
	assert.NotEqual(t, team.InviteId, "")
}

func TestSoftDeleteTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team := &model.Team{DisplayName: "DisplayName", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN}
	team, _ = Client.CreateTeam(team)

	ok, resp := Client.SoftDeleteTeam(team.Id)
	CheckNoError(t, resp)

	if !ok {
		t.Fatal("should have returned true")
	}

	rteam, err := th.App.GetTeam(team.Id)
	if err != nil {
		t.Fatal("should have returned archived team")
	}
	if rteam.DeleteAt == 0 {
		t.Fatal("should have not set to zero")
	}

	ok, resp = Client.SoftDeleteTeam("junk")
	CheckBadRequestStatus(t, resp)

	if ok {
		t.Fatal("should have returned false")
	}

	otherTeam := th.BasicTeam
	_, resp = Client.SoftDeleteTeam(otherTeam.Id)
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.SoftDeleteTeam(otherTeam.Id)
	CheckUnauthorizedStatus(t, resp)

	_, resp = th.SystemAdminClient.SoftDeleteTeam(otherTeam.Id)
	CheckNoError(t, resp)
}

func TestPermanentDeleteTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team := &model.Team{DisplayName: "DisplayName", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN}
	team, _ = Client.CreateTeam(team)

	enableAPITeamDeletion := *th.App.Config().ServiceSettings.EnableAPITeamDeletion
	defer func() {
		th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnableAPITeamDeletion = &enableAPITeamDeletion })
	}()

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPITeamDeletion = false })

	// Does not error when deletion is disabled, just soft deletes
	ok, resp := Client.PermanentDeleteTeam(team.Id)
	CheckNoError(t, resp)
	assert.True(t, ok)

	rteam, err := th.App.GetTeam(team.Id)
	assert.Nil(t, err)
	assert.True(t, rteam.DeleteAt > 0)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPITeamDeletion = true })

	ok, resp = Client.PermanentDeleteTeam(team.Id)
	CheckNoError(t, resp)
	assert.True(t, ok)

	_, err = th.App.GetTeam(team.Id)
	assert.NotNil(t, err)

	ok, resp = Client.PermanentDeleteTeam("junk")
	CheckBadRequestStatus(t, resp)

	if ok {
		t.Fatal("should have returned false")
	}
}

func TestGetAllTeams(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team1 := &model.Team{DisplayName: "Name", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN, AllowOpenInvite: true}
	team1, resp := Client.CreateTeam(team1)
	CheckNoError(t, resp)

	team2 := &model.Team{DisplayName: "Name2", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN, AllowOpenInvite: true}
	team2, resp = Client.CreateTeam(team2)
	CheckNoError(t, resp)

	team3 := &model.Team{DisplayName: "Name3", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN, AllowOpenInvite: false}
	team3, resp = Client.CreateTeam(team3)
	CheckNoError(t, resp)

	testCases := []struct {
		Name          string
		Page          int
		PerPage       int
		Permissions   []string
		ExpectedTeams []string
	}{
		{
			Name:          "Get 1 team per page",
			Page:          0,
			PerPage:       1,
			Permissions:   []string{model.PERMISSION_LIST_PUBLIC_TEAMS.Id},
			ExpectedTeams: []string{team1.Id},
		},
		{
			Name:          "Get second page with 1 team per page",
			Page:          1,
			PerPage:       1,
			Permissions:   []string{model.PERMISSION_LIST_PUBLIC_TEAMS.Id},
			ExpectedTeams: []string{team2.Id},
		},
		{
			Name:          "Get no items per page",
			Page:          1,
			PerPage:       0,
			Permissions:   []string{model.PERMISSION_LIST_PUBLIC_TEAMS.Id},
			ExpectedTeams: []string{},
		},
		{
			Name:          "Get all open teams",
			Page:          0,
			PerPage:       10,
			Permissions:   []string{model.PERMISSION_LIST_PUBLIC_TEAMS.Id},
			ExpectedTeams: []string{team1.Id, team2.Id},
		},
		{
			Name:          "Get all private teams",
			Page:          0,
			PerPage:       10,
			Permissions:   []string{model.PERMISSION_LIST_PRIVATE_TEAMS.Id},
			ExpectedTeams: []string{th.BasicTeam.Id, team3.Id},
		},
		{
			Name:          "Get all teams",
			Page:          0,
			PerPage:       10,
			Permissions:   []string{model.PERMISSION_LIST_PUBLIC_TEAMS.Id, model.PERMISSION_LIST_PRIVATE_TEAMS.Id},
			ExpectedTeams: []string{th.BasicTeam.Id, team1.Id, team2.Id, team3.Id},
		},
		{
			Name:          "Get no teams because permissions",
			Page:          0,
			PerPage:       10,
			Permissions:   []string{},
			ExpectedTeams: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			defaultRolePermissions := th.SaveDefaultRolePermissions()
			defer func() {
				th.RestoreDefaultRolePermissions(defaultRolePermissions)
			}()
			th.RemovePermissionFromRole(model.PERMISSION_LIST_PUBLIC_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			th.RemovePermissionFromRole(model.PERMISSION_JOIN_PUBLIC_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			th.RemovePermissionFromRole(model.PERMISSION_LIST_PRIVATE_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			th.RemovePermissionFromRole(model.PERMISSION_JOIN_PRIVATE_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			for _, permission := range tc.Permissions {
				th.AddPermissionToRole(permission, model.SYSTEM_USER_ROLE_ID)
			}

			var teams []*model.Team
			teams, resp = Client.GetAllTeams("", tc.Page, tc.PerPage)
			CheckNoError(t, resp)
			require.Equal(t, len(tc.ExpectedTeams), len(teams))
			for idx, team := range teams {
				assert.Equal(t, tc.ExpectedTeams[idx], team.Id)
			}
		})
	}

	t.Run("Unauthorized", func(t *testing.T) {
		Client.Logout()
		_, resp = Client.GetAllTeams("", 1, 10)
		CheckUnauthorizedStatus(t, resp)
	})
}

func TestGetAllTeamsSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:     t.Name() + "_1",
		Name:            GenerateTestTeamName(),
		Email:           th.GenerateTestEmail(),
		Type:            model.TEAM_OPEN,
		AllowedDomains:  "simulator.amazonses.com,dockerhost",
		AllowOpenInvite: true,
	})
	CheckNoError(t, resp)
	team2, resp := th.SystemAdminClient.CreateTeam(&model.Team{
		DisplayName:     t.Name() + "_2",
		Name:            GenerateTestTeamName(),
		Email:           th.GenerateTestEmail(),
		Type:            model.TEAM_OPEN,
		AllowedDomains:  "simulator.amazonses.com,dockerhost",
		AllowOpenInvite: true,
	})
	CheckNoError(t, resp)

	// This may not work if the server has over 1000 open teams on it

	t.Run("team admin/non-admin", func(t *testing.T) {
		teamFound := false
		team2Found := false

		rteams, resp := th.Client.GetAllTeams("", 0, 1000)
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Id == team.Id {
				teamFound = true
				if rteam.Email == "" {
					t.Fatal("should not have sanitized email for team admin")
				}
			} else if rteam.Id == team2.Id {
				team2Found = true
				if rteam.Email != "" {
					t.Fatal("should've sanitized email for non-admin")
				}
			}
		}

		if !teamFound || !team2Found {
			t.Fatal("wasn't returned the expected teams so the test wasn't run correctly")
		}
	})

	t.Run("system admin", func(t *testing.T) {
		rteams, resp := th.SystemAdminClient.GetAllTeams("", 0, 1000)
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Id != team.Id && rteam.Id != team2.Id {
				continue
			}

			if rteam.Email == "" {
				t.Fatal("should not have sanitized email")
			}
		}
	})
}

func TestGetTeamByName(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	rteam, resp := Client.GetTeamByName(team.Name, "")
	CheckNoError(t, resp)

	if rteam.Name != team.Name {
		t.Fatal("wrong team")
	}

	_, resp = Client.GetTeamByName("junk", "")
	CheckNotFoundStatus(t, resp)

	_, resp = Client.GetTeamByName("", "")
	CheckNotFoundStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeamByName(strings.ToUpper(team.Name), "")
	CheckNoError(t, resp)

	Client.Logout()
	_, resp = Client.GetTeamByName(team.Name, "")
	CheckUnauthorizedStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeamByName(team.Name, "")
	CheckNoError(t, resp)

	th.LoginTeamAdmin()

	team2 := &model.Team{DisplayName: "Name", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_OPEN, AllowOpenInvite: false}
	rteam2, _ := Client.CreateTeam(team2)

	team3 := &model.Team{DisplayName: "Name", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_INVITE, AllowOpenInvite: true}
	rteam3, _ := Client.CreateTeam(team3)

	th.LoginBasic()
	// AllowInviteOpen is false and team is open, and user is not on team
	_, resp = Client.GetTeamByName(rteam2.Name, "")
	CheckForbiddenStatus(t, resp)

	// AllowInviteOpen is true and team is invite only, and user is not on team
	_, resp = Client.GetTeamByName(rteam3.Name, "")
	CheckForbiddenStatus(t, resp)
}

func TestGetTeamByNameSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_1",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)

	t.Run("team user", func(t *testing.T) {
		th.LinkUserToTeam(th.BasicUser2, team)

		client := th.CreateClient()
		th.LoginBasic2WithClient(client)

		rteam, resp := client.GetTeamByName(team.Name, "")
		CheckNoError(t, resp)
		if rteam.Email != "" {
			t.Fatal("should've sanitized email")
		}
	})

	t.Run("team admin/non-admin", func(t *testing.T) {
		rteam, resp := th.Client.GetTeamByName(team.Name, "")
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email")
		}
	})

	t.Run("system admin", func(t *testing.T) {
		rteam, resp := th.SystemAdminClient.GetTeamByName(team.Name, "")
		CheckNoError(t, resp)
		if rteam.Email == "" {
			t.Fatal("should not have sanitized email")
		}
	})
}

func TestSearchAllTeams(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	oTeam := th.BasicTeam
	oTeam.AllowOpenInvite = true

	if updatedTeam, err := th.App.UpdateTeam(oTeam); err != nil {
		t.Fatal(err)
	} else {
		oTeam.UpdateAt = updatedTeam.UpdateAt
	}

	pTeam := &model.Team{DisplayName: "PName", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_INVITE}
	Client.CreateTeam(pTeam)

	rteams, resp := Client.SearchTeams(&model.TeamSearch{Term: oTeam.Name})
	CheckNoError(t, resp)

	if len(rteams) != 1 {
		t.Fatal("should have returned 1 team")
	}

	if oTeam.Id != rteams[0].Id {
		t.Fatal("invalid team")
	}

	rteams, resp = Client.SearchTeams(&model.TeamSearch{Term: oTeam.DisplayName})
	CheckNoError(t, resp)

	if len(rteams) != 1 {
		t.Fatal("should have returned 1 team")
	}

	if rteams[0].Id != oTeam.Id {
		t.Fatal("invalid team")
	}

	rteams, resp = Client.SearchTeams(&model.TeamSearch{Term: pTeam.Name})
	CheckNoError(t, resp)

	if len(rteams) != 0 {
		t.Fatal("should have not returned team")
	}

	rteams, resp = Client.SearchTeams(&model.TeamSearch{Term: pTeam.DisplayName})
	CheckNoError(t, resp)

	if len(rteams) != 0 {
		t.Fatal("should have not returned team")
	}

	rteams, resp = th.SystemAdminClient.SearchTeams(&model.TeamSearch{Term: oTeam.Name})
	CheckNoError(t, resp)

	if len(rteams) != 1 {
		t.Fatal("should have returned 1 team")
	}

	rteams, resp = th.SystemAdminClient.SearchTeams(&model.TeamSearch{Term: pTeam.DisplayName})
	CheckNoError(t, resp)

	if len(rteams) != 1 {
		t.Fatal("should have returned 1 team")
	}

	rteams, resp = Client.SearchTeams(&model.TeamSearch{Term: "junk"})
	CheckNoError(t, resp)

	if len(rteams) != 0 {
		t.Fatal("should have not returned team")
	}

	Client.Logout()

	_, resp = Client.SearchTeams(&model.TeamSearch{Term: pTeam.Name})
	CheckUnauthorizedStatus(t, resp)

	_, resp = Client.SearchTeams(&model.TeamSearch{Term: pTeam.DisplayName})
	CheckUnauthorizedStatus(t, resp)
}

func TestSearchAllTeamsSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_1",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)
	team2, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_2",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)

	t.Run("non-team user", func(t *testing.T) {
		client := th.CreateClient()
		th.LoginBasic2WithClient(client)

		rteams, resp := client.SearchTeams(&model.TeamSearch{Term: t.Name()})
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Email != "" {
				t.Fatal("should've sanitized email")
			} else if rteam.AllowedDomains != "" {
				t.Fatal("should've sanitized allowed domains")
			}
		}
	})

	t.Run("team user", func(t *testing.T) {
		th.LinkUserToTeam(th.BasicUser2, team)

		client := th.CreateClient()
		th.LoginBasic2WithClient(client)

		rteams, resp := client.SearchTeams(&model.TeamSearch{Term: t.Name()})
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Email != "" {
				t.Fatal("should've sanitized email")
			} else if rteam.AllowedDomains != "" {
				t.Fatal("should've sanitized allowed domains")
			}
		}
	})

	t.Run("team admin", func(t *testing.T) {
		rteams, resp := th.Client.SearchTeams(&model.TeamSearch{Term: t.Name()})
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Id == team.Id || rteam.Id == team2.Id || rteam.Id == th.BasicTeam.Id {
				if rteam.Email == "" {
					t.Fatal("should not have sanitized email")
				}
			}
		}
	})

	t.Run("system admin", func(t *testing.T) {
		rteams, resp := th.SystemAdminClient.SearchTeams(&model.TeamSearch{Term: t.Name()})
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Email == "" {
				t.Fatal("should not have sanitized email")
			}
		}
	})
}

func TestGetTeamsForUser(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	team2 := &model.Team{DisplayName: "Name", Name: GenerateTestTeamName(), Email: th.GenerateTestEmail(), Type: model.TEAM_INVITE}
	rteam2, _ := Client.CreateTeam(team2)

	teams, resp := Client.GetTeamsForUser(th.BasicUser.Id, "")
	CheckNoError(t, resp)

	if len(teams) != 2 {
		t.Fatal("wrong number of teams")
	}

	found1 := false
	found2 := false
	for _, t := range teams {
		if t.Id == th.BasicTeam.Id {
			found1 = true
		} else if t.Id == rteam2.Id {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Fatal("missing team")
	}

	_, resp = Client.GetTeamsForUser("junk", "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamsForUser(model.NewId(), "")
	CheckForbiddenStatus(t, resp)

	_, resp = Client.GetTeamsForUser(th.BasicUser2.Id, "")
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeamsForUser(th.BasicUser2.Id, "")
	CheckNoError(t, resp)
}

func TestGetTeamsForUserSanitization(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	team, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_1",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)
	team2, resp := th.Client.CreateTeam(&model.Team{
		DisplayName:    t.Name() + "_2",
		Name:           GenerateTestTeamName(),
		Email:          th.GenerateTestEmail(),
		Type:           model.TEAM_OPEN,
		AllowedDomains: "simulator.amazonses.com,dockerhost",
	})
	CheckNoError(t, resp)

	t.Run("team user", func(t *testing.T) {
		th.LinkUserToTeam(th.BasicUser2, team)
		th.LinkUserToTeam(th.BasicUser2, team2)

		client := th.CreateClient()
		th.LoginBasic2WithClient(client)

		rteams, resp := client.GetTeamsForUser(th.BasicUser2.Id, "")
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Id != team.Id && rteam.Id != team2.Id {
				continue
			}

			if rteam.Email != "" {
				t.Fatal("should've sanitized email")
			}
		}
	})

	t.Run("team admin", func(t *testing.T) {
		rteams, resp := th.Client.GetTeamsForUser(th.BasicUser.Id, "")
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Id != team.Id && rteam.Id != team2.Id {
				continue
			}

			if rteam.Email == "" {
				t.Fatal("should not have sanitized email")
			}
		}
	})

	t.Run("system admin", func(t *testing.T) {
		rteams, resp := th.SystemAdminClient.GetTeamsForUser(th.BasicUser.Id, "")
		CheckNoError(t, resp)
		for _, rteam := range rteams {
			if rteam.Id != team.Id && rteam.Id != team2.Id {
				continue
			}

			if rteam.Email == "" {
				t.Fatal("should not have sanitized email")
			}
		}
	})
}

func TestGetTeamMember(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam
	user := th.BasicUser

	rmember, resp := Client.GetTeamMember(team.Id, user.Id, "")
	CheckNoError(t, resp)

	if rmember.TeamId != team.Id {
		t.Fatal("wrong team id")
	}

	if rmember.UserId != user.Id {
		t.Fatal("wrong team id")
	}

	_, resp = Client.GetTeamMember("junk", user.Id, "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamMember(team.Id, "junk", "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamMember("junk", "junk", "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamMember(team.Id, model.NewId(), "")
	CheckNotFoundStatus(t, resp)

	_, resp = Client.GetTeamMember(model.NewId(), user.Id, "")
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeamMember(team.Id, user.Id, "")
	CheckNoError(t, resp)
}

func TestGetTeamMembers(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam
	userNotMember := th.CreateUser()

	rmembers, resp := Client.GetTeamMembers(team.Id, 0, 100, "")
	CheckNoError(t, resp)

	t.Logf("rmembers count %v\n", len(rmembers))

	if len(rmembers) == 0 {
		t.Fatal("should have results")
	}

	for _, rmember := range rmembers {
		if rmember.TeamId != team.Id || rmember.UserId == userNotMember.Id {
			t.Fatal("user should be a member of team")
		}
	}

	rmembers, resp = Client.GetTeamMembers(team.Id, 0, 1, "")
	CheckNoError(t, resp)
	if len(rmembers) != 1 {
		t.Fatal("should be 1 per page")
	}

	rmembers, resp = Client.GetTeamMembers(team.Id, 1, 1, "")
	CheckNoError(t, resp)
	if len(rmembers) != 1 {
		t.Fatal("should be 1 per page")
	}

	rmembers, resp = Client.GetTeamMembers(team.Id, 10000, 100, "")
	CheckNoError(t, resp)
	if len(rmembers) != 0 {
		t.Fatal("should be no member")
	}

	rmembers, resp = Client.GetTeamMembers(team.Id, 0, 2, "")
	CheckNoError(t, resp)
	rmembers2, resp := Client.GetTeamMembers(team.Id, 1, 2, "")
	CheckNoError(t, resp)

	for _, tm1 := range rmembers {
		for _, tm2 := range rmembers2 {
			assert.NotEqual(t, tm1.UserId+tm1.TeamId, tm2.UserId+tm2.TeamId, "different pages should not have the same members")
		}
	}

	_, resp = Client.GetTeamMembers("junk", 0, 100, "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamMembers(model.NewId(), 0, 100, "")
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetTeamMembers(team.Id, 0, 1, "")
	CheckUnauthorizedStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeamMembers(team.Id, 0, 100, "")
	CheckNoError(t, resp)
}

func TestGetTeamMembersForUser(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	members, resp := Client.GetTeamMembersForUser(th.BasicUser.Id, "")
	CheckNoError(t, resp)

	found := false
	for _, m := range members {
		if m.TeamId == th.BasicTeam.Id {
			found = true
		}
	}

	if !found {
		t.Fatal("missing team member")
	}

	_, resp = Client.GetTeamMembersForUser("junk", "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamMembersForUser(model.NewId(), "")
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetTeamMembersForUser(th.BasicUser.Id, "")
	CheckUnauthorizedStatus(t, resp)

	user := th.CreateUser()
	Client.Login(user.Email, user.Password)
	_, resp = Client.GetTeamMembersForUser(th.BasicUser.Id, "")
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeamMembersForUser(th.BasicUser.Id, "")
	CheckNoError(t, resp)
}

func TestGetTeamMembersByIds(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	tm, resp := Client.GetTeamMembersByIds(th.BasicTeam.Id, []string{th.BasicUser.Id})
	CheckNoError(t, resp)

	if tm[0].UserId != th.BasicUser.Id {
		t.Fatal("returned wrong user")
	}

	_, resp = Client.GetTeamMembersByIds(th.BasicTeam.Id, []string{})
	CheckBadRequestStatus(t, resp)

	tm1, resp := Client.GetTeamMembersByIds(th.BasicTeam.Id, []string{"junk"})
	CheckNoError(t, resp)
	if len(tm1) > 0 {
		t.Fatal("no users should be returned")
	}

	tm1, resp = Client.GetTeamMembersByIds(th.BasicTeam.Id, []string{"junk", th.BasicUser.Id})
	CheckNoError(t, resp)
	if len(tm1) != 1 {
		t.Fatal("1 user should be returned")
	}

	_, resp = Client.GetTeamMembersByIds("junk", []string{th.BasicUser.Id})
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamMembersByIds(model.NewId(), []string{th.BasicUser.Id})
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetTeamMembersByIds(th.BasicTeam.Id, []string{th.BasicUser.Id})
	CheckUnauthorizedStatus(t, resp)
}

func TestAddTeamMember(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam
	otherUser := th.CreateUser()

	if err := th.App.RemoveUserFromTeam(th.BasicTeam.Id, th.BasicUser2.Id, ""); err != nil {
		t.Fatalf(err.Error())
	}

	// Regular user can't add a member to a team they don't belong to.
	th.LoginBasic2()
	_, resp := Client.AddTeamMember(team.Id, otherUser.Id)
	CheckForbiddenStatus(t, resp)
	if resp.Error == nil {
		t.Fatalf("Error is nil")
	}
	Client.Logout()

	// Regular user can add a member to a team they belong to.
	th.LoginBasic()
	tm, resp := Client.AddTeamMember(team.Id, otherUser.Id)
	CheckNoError(t, resp)
	CheckCreatedStatus(t, resp)

	// Check all the returned data.
	if tm == nil {
		t.Fatal("should have returned team member")
	}

	if tm.UserId != otherUser.Id {
		t.Fatal("user ids should have matched")
	}

	if tm.TeamId != team.Id {
		t.Fatal("team ids should have matched")
	}

	// Check with various invalid requests.
	tm, resp = Client.AddTeamMember(team.Id, "junk")
	CheckBadRequestStatus(t, resp)

	if tm != nil {
		t.Fatal("should have not returned team member")
	}

	_, resp = Client.AddTeamMember("junk", otherUser.Id)
	CheckBadRequestStatus(t, resp)

	_, resp = Client.AddTeamMember(GenerateTestId(), otherUser.Id)
	CheckForbiddenStatus(t, resp)

	_, resp = Client.AddTeamMember(team.Id, GenerateTestId())
	CheckNotFoundStatus(t, resp)

	Client.Logout()

	// Check the appropriate permissions are enforced.
	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()

	// Set the config so that only team admins can add a user to a team.
	th.AddPermissionToRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_ADMIN_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_USER_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_USER_ROLE_ID)

	th.LoginBasic()

	// Check that a regular user can't add someone to the team.
	_, resp = Client.AddTeamMember(team.Id, otherUser.Id)
	CheckForbiddenStatus(t, resp)

	// Update user to team admin
	th.UpdateUserToTeamAdmin(th.BasicUser, th.BasicTeam)
	th.App.InvalidateAllCaches()
	th.LoginBasic()

	// Should work as a team admin.
	_, resp = Client.AddTeamMember(team.Id, otherUser.Id)
	CheckNoError(t, resp)

	// Change permission level to team user
	th.AddPermissionToRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_USER_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_USER_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_ADMIN_ROLE_ID)

	th.UpdateUserToNonTeamAdmin(th.BasicUser, th.BasicTeam)
	th.App.InvalidateAllCaches()
	th.LoginBasic()

	// Should work as a regular user.
	_, resp = Client.AddTeamMember(team.Id, otherUser.Id)
	CheckNoError(t, resp)

	// by token
	Client.Login(otherUser.Email, otherUser.Password)

	token := model.NewToken(
		app.TOKEN_TYPE_TEAM_INVITATION,
		model.MapToJson(map[string]string{"teamId": team.Id}),
	)
	<-th.App.Srv.Store.Token().Save(token)

	tm, resp = Client.AddTeamMemberFromInvite(token.Token, "")
	CheckNoError(t, resp)

	if tm == nil {
		t.Fatal("should have returned team member")
	}

	if tm.UserId != otherUser.Id {
		t.Fatal("user ids should have matched")
	}

	if tm.TeamId != team.Id {
		t.Fatal("team ids should have matched")
	}

	if result := <-th.App.Srv.Store.Token().GetByToken(token.Token); result.Err == nil {
		t.Fatal("The token must be deleted after be used")
	}

	tm, resp = Client.AddTeamMemberFromInvite("junk", "")
	CheckBadRequestStatus(t, resp)

	if tm != nil {
		t.Fatal("should have not returned team member")
	}

	// expired token of more than 50 hours
	token = model.NewToken(app.TOKEN_TYPE_TEAM_INVITATION, "")
	token.CreateAt = model.GetMillis() - 1000*60*60*50
	<-th.App.Srv.Store.Token().Save(token)

	_, resp = Client.AddTeamMemberFromInvite(token.Token, "")
	CheckBadRequestStatus(t, resp)
	th.App.DeleteToken(token)

	// invalid team id
	testId := GenerateTestId()
	token = model.NewToken(
		app.TOKEN_TYPE_TEAM_INVITATION,
		model.MapToJson(map[string]string{"teamId": testId}),
	)
	<-th.App.Srv.Store.Token().Save(token)

	_, resp = Client.AddTeamMemberFromInvite(token.Token, "")
	CheckNotFoundStatus(t, resp)
	th.App.DeleteToken(token)

	// by invite_id
	Client.Login(otherUser.Email, otherUser.Password)

	tm, resp = Client.AddTeamMemberFromInvite("", team.InviteId)
	CheckNoError(t, resp)

	if tm == nil {
		t.Fatal("should have returned team member")
	}

	if tm.UserId != otherUser.Id {
		t.Fatal("user ids should have matched")
	}

	if tm.TeamId != team.Id {
		t.Fatal("team ids should have matched")
	}

	tm, resp = Client.AddTeamMemberFromInvite("", "junk")
	CheckNotFoundStatus(t, resp)

	if tm != nil {
		t.Fatal("should have not returned team member")
	}

	// Set a team to group-constrained
	team.GroupConstrained = model.NewBool(true)
	_, err := th.App.UpdateTeam(team)
	require.Nil(t, err)

	// Attempt to use a token on a group-constrained team
	token = model.NewToken(
		app.TOKEN_TYPE_TEAM_INVITATION,
		model.MapToJson(map[string]string{"teamId": team.Id}),
	)
	<-th.App.Srv.Store.Token().Save(token)
	tm, resp = Client.AddTeamMemberFromInvite(token.Token, "")
	require.Equal(t, "app.team.invite_token.group_constrained.error", resp.Error.Id)

	// Attempt to use an invite id
	tm, resp = Client.AddTeamMemberFromInvite("", team.InviteId)
	require.Equal(t, "app.team.invite_id.group_constrained.error", resp.Error.Id)

	// User is not in associated groups so shouldn't be allowed
	_, resp = th.SystemAdminClient.AddTeamMember(team.Id, otherUser.Id)
	CheckErrorMessage(t, resp, "api.team.add_members.user_denied")

	// Associate group to team
	_, err = th.App.CreateGroupSyncable(&model.GroupSyncable{
		GroupId:    th.Group.Id,
		SyncableId: team.Id,
		Type:       model.GroupSyncableTypeTeam,
	})
	require.Nil(t, err)

	// Add user to group
	_, err = th.App.CreateOrRestoreGroupMember(th.Group.Id, otherUser.Id)
	require.Nil(t, err)

	_, resp = th.SystemAdminClient.AddTeamMember(team.Id, otherUser.Id)
	CheckNoError(t, resp)
}

func TestAddTeamMemberMyself(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	// Check the appropriate permissions are enforced.
	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()

	th.LoginBasic()

	testCases := []struct {
		Name              string
		Public            bool
		PublicPermission  bool
		PrivatePermission bool
		ExpectedSuccess   bool
	}{
		{
			Name:              "Try to join an open team without the permissions",
			Public:            true,
			PublicPermission:  false,
			PrivatePermission: false,
			ExpectedSuccess:   false,
		},
		{
			Name:              "Try to join a private team without the permissions",
			Public:            false,
			PublicPermission:  false,
			PrivatePermission: false,
			ExpectedSuccess:   false,
		},
		{
			Name:              "Try to join an open team without public permission but with private permissions",
			Public:            true,
			PublicPermission:  false,
			PrivatePermission: true,
			ExpectedSuccess:   false,
		},
		{
			Name:              "Try to join a private team without private permission but with public permission",
			Public:            false,
			PublicPermission:  true,
			PrivatePermission: false,
			ExpectedSuccess:   false,
		},
		{
			Name:              "Join an open team with the permissions",
			Public:            true,
			PublicPermission:  true,
			PrivatePermission: false,
			ExpectedSuccess:   true,
		},
		{
			Name:              "Join a private team with the permissions",
			Public:            false,
			PublicPermission:  false,
			PrivatePermission: true,
			ExpectedSuccess:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			team := th.CreateTeam()
			team.AllowOpenInvite = tc.Public
			th.App.UpdateTeam(team)
			if tc.PublicPermission {
				th.AddPermissionToRole(model.PERMISSION_JOIN_PUBLIC_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			} else {
				th.RemovePermissionFromRole(model.PERMISSION_JOIN_PUBLIC_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			}
			if tc.PrivatePermission {
				th.AddPermissionToRole(model.PERMISSION_JOIN_PRIVATE_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			} else {
				th.RemovePermissionFromRole(model.PERMISSION_JOIN_PRIVATE_TEAMS.Id, model.SYSTEM_USER_ROLE_ID)
			}
			_, resp := Client.AddTeamMember(team.Id, th.BasicUser.Id)
			if tc.ExpectedSuccess {
				CheckNoError(t, resp)
			} else {
				CheckForbiddenStatus(t, resp)
			}
		})
	}

}

func TestAddTeamMembers(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam
	otherUser := th.CreateUser()
	userList := []string{
		otherUser.Id,
	}

	if err := th.App.RemoveUserFromTeam(th.BasicTeam.Id, th.BasicUser2.Id, ""); err != nil {
		t.Fatalf(err.Error())
	}

	// Regular user can't add a member to a team they don't belong to.
	th.LoginBasic2()
	_, resp := Client.AddTeamMembers(team.Id, userList)
	CheckForbiddenStatus(t, resp)
	Client.Logout()

	// Regular user can add a member to a team they belong to.
	th.LoginBasic()
	tm, resp := Client.AddTeamMembers(team.Id, userList)
	CheckNoError(t, resp)
	CheckCreatedStatus(t, resp)

	// Check all the returned data.
	if tm[0] == nil {
		t.Fatal("should have returned team member")
	}

	if tm[0].UserId != otherUser.Id {
		t.Fatal("user ids should have matched")
	}

	if tm[0].TeamId != team.Id {
		t.Fatal("team ids should have matched")
	}

	// Check with various invalid requests.
	_, resp = Client.AddTeamMembers("junk", userList)
	CheckBadRequestStatus(t, resp)

	_, resp = Client.AddTeamMembers(GenerateTestId(), userList)
	CheckNotFoundStatus(t, resp)

	testUserList := append(userList, GenerateTestId())
	_, resp = Client.AddTeamMembers(team.Id, testUserList)
	CheckNotFoundStatus(t, resp)

	// Test with many users.
	for i := 0; i < 25; i++ {
		testUserList = append(testUserList, GenerateTestId())
	}
	_, resp = Client.AddTeamMembers(team.Id, testUserList)
	CheckBadRequestStatus(t, resp)

	Client.Logout()

	// Check the appropriate permissions are enforced.
	defaultRolePermissions := th.SaveDefaultRolePermissions()
	defer func() {
		th.RestoreDefaultRolePermissions(defaultRolePermissions)
	}()

	// Set the config so that only team admins can add a user to a team.
	th.AddPermissionToRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_ADMIN_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_USER_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_USER_ROLE_ID)

	th.LoginBasic()

	// Check that a regular user can't add someone to the team.
	_, resp = Client.AddTeamMembers(team.Id, userList)
	CheckForbiddenStatus(t, resp)

	// Update user to team admin
	th.UpdateUserToTeamAdmin(th.BasicUser, th.BasicTeam)
	th.App.InvalidateAllCaches()
	th.LoginBasic()

	// Should work as a team admin.
	_, resp = Client.AddTeamMembers(team.Id, userList)
	CheckNoError(t, resp)

	// Change permission level to team user
	th.AddPermissionToRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_USER_ROLE_ID)
	th.AddPermissionToRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_USER_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_INVITE_USER.Id, model.TEAM_ADMIN_ROLE_ID)
	th.RemovePermissionFromRole(model.PERMISSION_ADD_USER_TO_TEAM.Id, model.TEAM_ADMIN_ROLE_ID)

	th.UpdateUserToNonTeamAdmin(th.BasicUser, th.BasicTeam)
	th.App.InvalidateAllCaches()
	th.LoginBasic()

	// Should work as a regular user.
	_, resp = Client.AddTeamMembers(team.Id, userList)
	CheckNoError(t, resp)

	// Set a team to group-constrained
	team.GroupConstrained = model.NewBool(true)
	_, err := th.App.UpdateTeam(team)
	require.Nil(t, err)

	// User is not in associated groups so shouldn't be allowed
	_, resp = Client.AddTeamMembers(team.Id, userList)
	CheckErrorMessage(t, resp, "api.team.add_members.user_denied")

	// Associate group to team
	_, err = th.App.CreateGroupSyncable(&model.GroupSyncable{
		GroupId:    th.Group.Id,
		SyncableId: team.Id,
		Type:       model.GroupSyncableTypeTeam,
	})
	require.Nil(t, err)

	// Add user to group
	_, err = th.App.CreateOrRestoreGroupMember(th.Group.Id, userList[0])
	require.Nil(t, err)

	_, resp = Client.AddTeamMembers(team.Id, userList)
	CheckNoError(t, resp)
}

func TestRemoveTeamMember(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	pass, resp := Client.RemoveTeamMember(th.BasicTeam.Id, th.BasicUser.Id)
	CheckNoError(t, resp)

	if !pass {
		t.Fatal("should have passed")
	}

	_, resp = th.SystemAdminClient.AddTeamMember(th.BasicTeam.Id, th.BasicUser.Id)
	CheckNoError(t, resp)

	_, resp = Client.RemoveTeamMember(th.BasicTeam.Id, "junk")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.RemoveTeamMember("junk", th.BasicUser2.Id)
	CheckBadRequestStatus(t, resp)

	_, resp = Client.RemoveTeamMember(th.BasicTeam.Id, th.BasicUser2.Id)
	CheckForbiddenStatus(t, resp)

	_, resp = Client.RemoveTeamMember(model.NewId(), th.BasicUser.Id)
	CheckNotFoundStatus(t, resp)

	_, resp = th.SystemAdminClient.RemoveTeamMember(th.BasicTeam.Id, th.BasicUser.Id)
	CheckNoError(t, resp)

	_, resp = th.SystemAdminClient.AddTeamMember(th.BasicTeam.Id, th.SystemAdminUser.Id)
	CheckNoError(t, resp)

	// If the team is group-constrained the user cannot be removed
	th.BasicTeam.GroupConstrained = model.NewBool(true)
	_, err := th.App.UpdateTeam(th.BasicTeam)
	require.Nil(t, err)
	_, resp = th.SystemAdminClient.RemoveTeamMember(th.BasicTeam.Id, th.BasicUser.Id)
	require.Equal(t, "api.team.remove_member.group_constrained.app_error", resp.Error.Id)

	// Can remove self even if team is group-constrained
	_, resp = th.SystemAdminClient.RemoveTeamMember(th.BasicTeam.Id, th.SystemAdminUser.Id)
	CheckNoError(t, resp)
}

func TestGetTeamStats(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	rstats, resp := Client.GetTeamStats(team.Id, "")
	CheckNoError(t, resp)

	if rstats.TeamId != team.Id {
		t.Fatal("wrong team id")
	}

	if rstats.TotalMemberCount != 3 {
		t.Fatal("wrong count")
	}

	if rstats.ActiveMemberCount != 3 {
		t.Fatal("wrong count")
	}

	_, resp = Client.GetTeamStats("junk", "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamStats(model.NewId(), "")
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.GetTeamStats(team.Id, "")
	CheckNoError(t, resp)

	// deactivate BasicUser2
	th.UpdateActiveUser(th.BasicUser2, false)

	rstats, resp = th.SystemAdminClient.GetTeamStats(team.Id, "")
	CheckNoError(t, resp)

	if rstats.TotalMemberCount != 3 {
		t.Fatal("wrong count")
	}

	if rstats.ActiveMemberCount != 2 {
		t.Fatal("wrong count")
	}

	// login with different user and test if forbidden
	user := th.CreateUser()
	Client.Login(user.Email, user.Password)
	_, resp = Client.GetTeamStats(th.BasicTeam.Id, "")
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetTeamStats(th.BasicTeam.Id, "")
	CheckUnauthorizedStatus(t, resp)
}

func TestUpdateTeamMemberRoles(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	SystemAdminClient := th.SystemAdminClient

	const TEAM_MEMBER = "team_user"
	const TEAM_ADMIN = "team_user team_admin"

	// user 1 tries to promote user 2
	ok, resp := Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser2.Id, TEAM_ADMIN)
	CheckForbiddenStatus(t, resp)
	if ok {
		t.Fatal("should have returned false")
	}

	// user 1 tries to promote himself
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser.Id, TEAM_ADMIN)
	CheckForbiddenStatus(t, resp)

	// user 1 tries to demote someone
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.SystemAdminUser.Id, TEAM_MEMBER)
	CheckForbiddenStatus(t, resp)

	// system admin promotes user 1
	ok, resp = SystemAdminClient.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser.Id, TEAM_ADMIN)
	CheckNoError(t, resp)
	if !ok {
		t.Fatal("should have returned true")
	}

	// user 1 (team admin) promotes user 2
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser2.Id, TEAM_ADMIN)
	CheckNoError(t, resp)

	// user 1 (team admin) demotes user 2 (team admin)
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser2.Id, TEAM_MEMBER)
	CheckNoError(t, resp)

	// user 1 (team admin) tries to demote system admin (not member of a team)
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.SystemAdminUser.Id, TEAM_MEMBER)
	CheckNotFoundStatus(t, resp)

	// user 1 (team admin) demotes system admin (member of a team)
	th.LinkUserToTeam(th.SystemAdminUser, th.BasicTeam)
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.SystemAdminUser.Id, TEAM_MEMBER)
	CheckNoError(t, resp)
	// Note from API v3
	// Note to anyone who thinks this (above) test is wrong:
	// This operation will not affect the system admin's permissions because they have global access to all teams.
	// Their team level permissions are irrelevant. A team admin should be able to manage team level permissions.

	// System admins should be able to manipulate permission no matter what their team level permissions are.
	// system admin promotes user 2
	_, resp = SystemAdminClient.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser2.Id, TEAM_ADMIN)
	CheckNoError(t, resp)

	// system admin demotes user 2 (team admin)
	_, resp = SystemAdminClient.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser2.Id, TEAM_MEMBER)
	CheckNoError(t, resp)

	// user 1 (team admin) tries to promote himself to a random team
	_, resp = Client.UpdateTeamMemberRoles(model.NewId(), th.BasicUser.Id, TEAM_ADMIN)
	CheckForbiddenStatus(t, resp)

	// user 1 (team admin) tries to promote a random user
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, model.NewId(), TEAM_ADMIN)
	CheckNotFoundStatus(t, resp)

	// user 1 (team admin) tries to promote invalid team permission
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser.Id, "junk")
	CheckBadRequestStatus(t, resp)

	// user 1 (team admin) demotes himself
	_, resp = Client.UpdateTeamMemberRoles(th.BasicTeam.Id, th.BasicUser.Id, TEAM_MEMBER)
	CheckNoError(t, resp)
}

func TestUpdateTeamMemberSchemeRoles(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	SystemAdminClient := th.SystemAdminClient
	th.LoginBasic()

	s1 := &model.SchemeRoles{
		SchemeAdmin: false,
		SchemeUser:  false,
		SchemeGuest: false,
	}
	_, r1 := SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.BasicUser.Id, s1)
	CheckNoError(t, r1)

	tm1, rtm1 := SystemAdminClient.GetTeamMember(th.BasicTeam.Id, th.BasicUser.Id, "")
	CheckNoError(t, rtm1)
	assert.Equal(t, false, tm1.SchemeGuest)
	assert.Equal(t, false, tm1.SchemeUser)
	assert.Equal(t, false, tm1.SchemeAdmin)

	s2 := &model.SchemeRoles{
		SchemeAdmin: false,
		SchemeUser:  true,
		SchemeGuest: false,
	}
	_, r2 := SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.BasicUser.Id, s2)
	CheckNoError(t, r2)

	tm2, rtm2 := SystemAdminClient.GetTeamMember(th.BasicTeam.Id, th.BasicUser.Id, "")
	CheckNoError(t, rtm2)
	assert.Equal(t, false, tm2.SchemeGuest)
	assert.Equal(t, true, tm2.SchemeUser)
	assert.Equal(t, false, tm2.SchemeAdmin)

	s3 := &model.SchemeRoles{
		SchemeAdmin: true,
		SchemeUser:  false,
		SchemeGuest: false,
	}
	_, r3 := SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.BasicUser.Id, s3)
	CheckNoError(t, r3)

	tm3, rtm3 := SystemAdminClient.GetTeamMember(th.BasicTeam.Id, th.BasicUser.Id, "")
	CheckNoError(t, rtm3)
	assert.Equal(t, false, tm3.SchemeGuest)
	assert.Equal(t, false, tm3.SchemeUser)
	assert.Equal(t, true, tm3.SchemeAdmin)

	s4 := &model.SchemeRoles{
		SchemeAdmin: true,
		SchemeUser:  true,
		SchemeGuest: false,
	}
	_, r4 := SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.BasicUser.Id, s4)
	CheckNoError(t, r4)

	tm4, rtm4 := SystemAdminClient.GetTeamMember(th.BasicTeam.Id, th.BasicUser.Id, "")
	CheckNoError(t, rtm4)
	assert.Equal(t, false, tm4.SchemeGuest)
	assert.Equal(t, true, tm4.SchemeUser)
	assert.Equal(t, true, tm4.SchemeAdmin)

	s5 := &model.SchemeRoles{
		SchemeAdmin: false,
		SchemeUser:  false,
		SchemeGuest: true,
	}
	_, r5 := SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.BasicUser.Id, s5)
	CheckNoError(t, r5)

	tm5, rtm5 := SystemAdminClient.GetTeamMember(th.BasicTeam.Id, th.BasicUser.Id, "")
	CheckNoError(t, rtm5)
	assert.Equal(t, true, tm5.SchemeGuest)
	assert.Equal(t, false, tm5.SchemeUser)
	assert.Equal(t, false, tm5.SchemeAdmin)

	s6 := &model.SchemeRoles{
		SchemeAdmin: false,
		SchemeUser:  true,
		SchemeGuest: true,
	}
	_, resp := SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.BasicUser.Id, s6)
	CheckBadRequestStatus(t, resp)

	_, resp = SystemAdminClient.UpdateTeamMemberSchemeRoles(model.NewId(), th.BasicUser.Id, s4)
	CheckNotFoundStatus(t, resp)

	_, resp = SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, model.NewId(), s4)
	CheckNotFoundStatus(t, resp)

	_, resp = SystemAdminClient.UpdateTeamMemberSchemeRoles("ASDF", th.BasicUser.Id, s4)
	CheckBadRequestStatus(t, resp)

	_, resp = SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, "ASDF", s4)
	CheckBadRequestStatus(t, resp)

	th.LoginBasic2()
	_, resp = th.Client.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.BasicUser.Id, s4)
	CheckForbiddenStatus(t, resp)

	SystemAdminClient.Logout()
	_, resp = SystemAdminClient.UpdateTeamMemberSchemeRoles(th.BasicTeam.Id, th.SystemAdminUser.Id, s4)
	CheckUnauthorizedStatus(t, resp)
}

func TestGetMyTeamsUnread(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	user := th.BasicUser
	Client.Login(user.Email, user.Password)

	teams, resp := Client.GetTeamsUnreadForUser(user.Id, "")
	CheckNoError(t, resp)
	if len(teams) == 0 {
		t.Fatal("should have results")
	}

	teams, resp = Client.GetTeamsUnreadForUser(user.Id, th.BasicTeam.Id)
	CheckNoError(t, resp)
	if len(teams) != 0 {
		t.Fatal("should not have results")
	}

	_, resp = Client.GetTeamsUnreadForUser("fail", "")
	CheckBadRequestStatus(t, resp)

	_, resp = Client.GetTeamsUnreadForUser(model.NewId(), "")
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetTeamsUnreadForUser(user.Id, "")
	CheckUnauthorizedStatus(t, resp)
}

func TestTeamExists(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	th.LoginBasic()

	exists, resp := Client.TeamExists(team.Name, "")
	CheckNoError(t, resp)
	if !exists {
		t.Fatal("team should exist")
	}

	exists, resp = Client.TeamExists("testingteam", "")
	CheckNoError(t, resp)
	if exists {
		t.Fatal("team should not exist")
	}

	Client.Logout()
	_, resp = Client.TeamExists(team.Name, "")
	CheckUnauthorizedStatus(t, resp)
}

func TestImportTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	t.Run("ImportTeam", func(t *testing.T) {
		var data []byte
		var err error
		data, err = testutils.ReadTestFile("Fake_Team_Import.zip")
		if err != nil && len(data) == 0 {
			t.Fatal("Error while reading the test file.")
		}

		// Import the channels/users/posts
		fileResp, resp := th.SystemAdminClient.ImportTeam(data, binary.Size(data), "slack", "Fake_Team_Import.zip", th.BasicTeam.Id)
		CheckNoError(t, resp)

		fileData, err := base64.StdEncoding.DecodeString(fileResp["results"])
		if err != nil {
			t.Fatal("failed to decode base64 results data")
		}

		fileReturned := fmt.Sprintf("%s", fileData)
		if !strings.Contains(fileReturned, "darth.vader@stardeath.com") {
			t.Log(fileReturned)
			t.Fatal("failed to report the user was imported")
		}

		// Checking the imported users
		importedUser, resp := th.SystemAdminClient.GetUserByUsername("bot_test", "")
		CheckNoError(t, resp)
		if importedUser.Username != "bot_test" {
			t.Fatal("username should match with the imported user")
		}

		importedUser, resp = th.SystemAdminClient.GetUserByUsername("lordvader", "")
		CheckNoError(t, resp)
		if importedUser.Username != "lordvader" {
			t.Fatal("username should match with the imported user")
		}

		// Checking the imported Channels
		importedChannel, resp := th.SystemAdminClient.GetChannelByName("testchannel", th.BasicTeam.Id, "")
		CheckNoError(t, resp)
		if importedChannel.Name != "testchannel" {
			t.Fatal("names did not match expected: testchannel")
		}

		importedChannel, resp = th.SystemAdminClient.GetChannelByName("general", th.BasicTeam.Id, "")
		CheckNoError(t, resp)
		if importedChannel.Name != "general" {
			t.Fatal("names did not match expected: general")
		}

		posts, resp := th.SystemAdminClient.GetPostsForChannel(importedChannel.Id, 0, 60, "")
		CheckNoError(t, resp)
		if posts.Posts[posts.Order[3]].Message != "This is a test post to test the import process" {
			t.Fatal("missing posts in the import process")
		}
	})

	t.Run("MissingFile", func(t *testing.T) {
		_, resp := th.SystemAdminClient.ImportTeam(nil, 4343, "slack", "Fake_Team_Import.zip", th.BasicTeam.Id)
		CheckBadRequestStatus(t, resp)
	})

	t.Run("WrongPermission", func(t *testing.T) {
		var data []byte
		var err error
		data, err = testutils.ReadTestFile("Fake_Team_Import.zip")
		if err != nil && len(data) == 0 {
			t.Fatal("Error while reading the test file.")
		}

		// Import the channels/users/posts
		_, resp := th.Client.ImportTeam(data, binary.Size(data), "slack", "Fake_Team_Import.zip", th.BasicTeam.Id)
		CheckForbiddenStatus(t, resp)
	})
}

func TestInviteUsersToTeam(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	user1 := th.GenerateTestEmail()
	user2 := th.GenerateTestEmail()

	emailList := []string{user1, user2}

	//Delete all the messages before check the sample email
	mailservice.DeleteMailBox(user1)
	mailservice.DeleteMailBox(user2)

	enableEmailInvitations := *th.App.Config().ServiceSettings.EnableEmailInvitations
	restrictCreationToDomains := th.App.Config().TeamSettings.RestrictCreationToDomains
	defer func() {
		th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnableEmailInvitations = &enableEmailInvitations })
		th.App.UpdateConfig(func(cfg *model.Config) { cfg.TeamSettings.RestrictCreationToDomains = restrictCreationToDomains })
	}()

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = false })
	_, resp := th.SystemAdminClient.InviteUsersToTeam(th.BasicTeam.Id, emailList)
	if resp.Error == nil {
		t.Fatal("Should be disabled")
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableEmailInvitations = true })
	okMsg, resp := th.SystemAdminClient.InviteUsersToTeam(th.BasicTeam.Id, emailList)
	CheckNoError(t, resp)
	if !okMsg {
		t.Fatal("should return true")
	}

	nameFormat := *th.App.Config().TeamSettings.TeammateNameDisplay
	expectedSubject := utils.T("api.templates.invite_subject",
		map[string]interface{}{"SenderName": th.SystemAdminUser.GetDisplayName(nameFormat),
			"TeamDisplayName": th.BasicTeam.DisplayName,
			"SiteName":        th.App.ClientConfig()["SiteName"]})

	//Check if the email was send to the rigth email address
	for _, email := range emailList {
		var resultsMailbox mailservice.JSONMessageHeaderInbucket
		err := mailservice.RetryInbucket(5, func() error {
			var err error
			resultsMailbox, err = mailservice.GetMailBox(email)
			return err
		})
		if err != nil {
			t.Log(err)
			t.Log("No email was received, maybe due load on the server. Disabling this verification")
		}
		if err == nil && len(resultsMailbox) > 0 {
			if !strings.ContainsAny(resultsMailbox[len(resultsMailbox)-1].To[0], email) {
				t.Fatal("Wrong To recipient")
			} else {
				if resultsEmail, err := mailservice.GetMessageFromMailbox(email, resultsMailbox[len(resultsMailbox)-1].ID); err == nil {
					if resultsEmail.Subject != expectedSubject {
						t.Log(resultsEmail.Subject)
						t.Log(expectedSubject)
						t.Fatal("Wrong Subject")
					}
				}
			}
		}
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.RestrictCreationToDomains = "@global.com,@common.com" })

	t.Run("restricted domains", func(t *testing.T) {
		err := th.App.InviteNewUsersToTeam(emailList, th.BasicTeam.Id, th.BasicUser.Id)

		if err == nil {
			t.Fatal("Adding users with non-restricted domains was allowed")
		}
		if err.Where != "InviteNewUsersToTeam" || err.Id != "api.team.invite_members.invalid_email.app_error" {
			t.Log(err)
			t.Fatal("Got wrong error message!")
		}
	})

	t.Run("override restricted domains", func(t *testing.T) {
		th.BasicTeam.AllowedDomains = "invalid.com,common.com"
		if _, err := th.App.UpdateTeam(th.BasicTeam); err == nil {
			t.Fatal("Should not update the team")
		}

		th.BasicTeam.AllowedDomains = "common.com"
		if _, err := th.App.UpdateTeam(th.BasicTeam); err != nil {
			t.Log(err)
			t.Fatal("Should update the team")
		}

		if err := th.App.InviteNewUsersToTeam([]string{"test@global.com"}, th.BasicTeam.Id, th.BasicUser.Id); err == nil || err.Where != "InviteNewUsersToTeam" {
			t.Log(err)
			t.Fatal("Per team restriction should take precedence over the global restriction")
		}

		if err := th.App.InviteNewUsersToTeam([]string{"test@common.com"}, th.BasicTeam.Id, th.BasicUser.Id); err != nil {
			t.Log(err)
			t.Fatal("Failed to invite user which was common between team and global domain restriction")
		}

		if err := th.App.InviteNewUsersToTeam([]string{"test@invalid.com"}, th.BasicTeam.Id, th.BasicUser.Id); err == nil {
			t.Log(err)
			t.Fatal("Should not invite user")
		}

	})
}

func TestGetTeamInviteInfo(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	team, resp := Client.GetTeamInviteInfo(team.InviteId)
	CheckNoError(t, resp)

	if team.DisplayName == "" {
		t.Fatal("should not be empty")
	}

	if team.Email != "" {
		t.Fatal("should be empty")
	}

	team.InviteId = "12345678901234567890123456789012"
	team, resp = th.SystemAdminClient.UpdateTeam(team)
	CheckNoError(t, resp)

	_, resp = Client.GetTeamInviteInfo(team.InviteId)
	CheckNoError(t, resp)

	_, resp = Client.GetTeamInviteInfo("junk")
	CheckNotFoundStatus(t, resp)
}

func TestSetTeamIcon(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	data, err := testutils.ReadTestFile("test.png")
	if err != nil {
		t.Fatal(err)
	}

	th.LoginTeamAdmin()

	ok, resp := Client.SetTeamIcon(team.Id, data)
	if !ok {
		t.Fatal(resp.Error)
	}
	CheckNoError(t, resp)

	ok, resp = Client.SetTeamIcon(model.NewId(), data)
	if ok {
		t.Fatal("Should return false, set team icon not allowed")
	}
	CheckForbiddenStatus(t, resp)

	th.LoginBasic()

	_, resp = Client.SetTeamIcon(team.Id, data)
	if resp.StatusCode == http.StatusForbidden {
		CheckForbiddenStatus(t, resp)
	} else if resp.StatusCode == http.StatusUnauthorized {
		CheckUnauthorizedStatus(t, resp)
	} else {
		t.Fatal("Should have failed either forbidden or unauthorized")
	}

	Client.Logout()

	_, resp = Client.SetTeamIcon(team.Id, data)
	if resp.StatusCode == http.StatusForbidden {
		CheckForbiddenStatus(t, resp)
	} else if resp.StatusCode == http.StatusUnauthorized {
		CheckUnauthorizedStatus(t, resp)
	} else {
		t.Fatal("Should have failed either forbidden or unauthorized")
	}

	teamBefore, err := th.App.GetTeam(team.Id)
	require.Nil(t, err)

	_, resp = th.SystemAdminClient.SetTeamIcon(team.Id, data)
	CheckNoError(t, resp)

	teamAfter, err := th.App.GetTeam(team.Id)
	require.Nil(t, err)
	assert.True(t, teamBefore.LastTeamIconUpdate < teamAfter.LastTeamIconUpdate, "LastTeamIconUpdate should have been updated for team")

	info := &model.FileInfo{Path: "teams/" + team.Id + "/teamIcon.png"}
	if err := th.cleanupTestFile(info); err != nil {
		t.Fatal(err)
	}
}

func TestGetTeamIcon(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	// should always fail because no initial image and no auto creation
	_, resp := Client.GetTeamIcon(team.Id, "")
	CheckNotFoundStatus(t, resp)

	Client.Logout()

	_, resp = Client.GetTeamIcon(team.Id, "")
	CheckUnauthorizedStatus(t, resp)
}

func TestRemoveTeamIcon(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	team := th.BasicTeam

	th.LoginTeamAdmin()
	data, _ := testutils.ReadTestFile("test.png")
	Client.SetTeamIcon(team.Id, data)

	_, resp := Client.RemoveTeamIcon(team.Id)
	CheckNoError(t, resp)
	teamAfter, _ := th.App.GetTeam(team.Id)
	if teamAfter.LastTeamIconUpdate != 0 {
		t.Fatal("should update LastTeamIconUpdate to 0")
	}

	Client.SetTeamIcon(team.Id, data)

	_, resp = th.SystemAdminClient.RemoveTeamIcon(team.Id)
	CheckNoError(t, resp)
	teamAfter, _ = th.App.GetTeam(team.Id)
	if teamAfter.LastTeamIconUpdate != 0 {
		t.Fatal("should update LastTeamIconUpdate to 0")
	}

	Client.SetTeamIcon(team.Id, data)
	Client.Logout()

	_, resp = Client.RemoveTeamIcon(team.Id)
	CheckUnauthorizedStatus(t, resp)

	th.LoginBasic()
	_, resp = Client.RemoveTeamIcon(team.Id)
	CheckForbiddenStatus(t, resp)
}

func TestUpdateTeamScheme(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	th.App.SetLicense(model.NewTestLicense(""))

	th.App.SetPhase2PermissionsMigrationStatus(true)

	team := &model.Team{
		DisplayName:     "Name",
		Description:     "Some description",
		CompanyName:     "Some company name",
		AllowOpenInvite: false,
		InviteId:        "inviteid0",
		Name:            "z-z-" + model.NewId() + "a",
		Email:           "success+" + model.NewId() + "@simulator.amazonses.com",
		Type:            model.TEAM_OPEN,
	}
	team, _ = th.SystemAdminClient.CreateTeam(team)

	teamScheme := &model.Scheme{
		DisplayName: "DisplayName",
		Name:        model.NewId(),
		Description: "Some description",
		Scope:       model.SCHEME_SCOPE_TEAM,
	}
	teamScheme, _ = th.SystemAdminClient.CreateScheme(teamScheme)
	channelScheme := &model.Scheme{
		DisplayName: "DisplayName",
		Name:        model.NewId(),
		Description: "Some description",
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	}
	channelScheme, _ = th.SystemAdminClient.CreateScheme(channelScheme)

	// Test the setup/base case.
	_, resp := th.SystemAdminClient.UpdateTeamScheme(team.Id, teamScheme.Id)
	CheckNoError(t, resp)

	// Test the return to default scheme
	_, resp = th.SystemAdminClient.UpdateTeamScheme(team.Id, "")
	CheckNoError(t, resp)

	// Test various invalid team and scheme id combinations.
	_, resp = th.SystemAdminClient.UpdateTeamScheme(team.Id, "x")
	CheckBadRequestStatus(t, resp)
	_, resp = th.SystemAdminClient.UpdateTeamScheme("x", teamScheme.Id)
	CheckBadRequestStatus(t, resp)
	_, resp = th.SystemAdminClient.UpdateTeamScheme("x", "x")
	CheckBadRequestStatus(t, resp)

	// Test that permissions are required.
	_, resp = th.Client.UpdateTeamScheme(team.Id, teamScheme.Id)
	CheckForbiddenStatus(t, resp)

	// Test that a license is required.
	th.App.SetLicense(nil)
	_, resp = th.SystemAdminClient.UpdateTeamScheme(team.Id, teamScheme.Id)
	CheckNotImplementedStatus(t, resp)
	th.App.SetLicense(model.NewTestLicense(""))

	// Test an invalid scheme scope.
	_, resp = th.SystemAdminClient.UpdateTeamScheme(team.Id, channelScheme.Id)
	CheckBadRequestStatus(t, resp)

	// Test that an unauthenticated user gets rejected.
	th.SystemAdminClient.Logout()
	_, resp = th.SystemAdminClient.UpdateTeamScheme(team.Id, teamScheme.Id)
	CheckUnauthorizedStatus(t, resp)
}
