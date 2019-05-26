// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api4

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/require"
)

type testHandler struct {
	t *testing.T
}

func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bb, err := ioutil.ReadAll(r.Body)
	assert.Nil(th.t, err)
	assert.NotEmpty(th.t, string(bb))
	poir := model.PostActionIntegrationRequestFromJson(bytes.NewReader(bb))
	assert.NotEmpty(th.t, poir.UserId)
	assert.NotEmpty(th.t, poir.ChannelId)
	assert.Empty(th.t, poir.TeamId)
	assert.NotEmpty(th.t, poir.PostId)
	assert.NotEmpty(th.t, poir.TriggerId)
	assert.Equal(th.t, "button", poir.Type)
	assert.Equal(th.t, "test-value", poir.Context["test-key"])
	w.Write([]byte("{}"))
	w.WriteHeader(200)
}

func TestPostActionCookies(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.ServiceSettings.AllowedUntrustedInternalConnections = "localhost 127.0.0.1"
	})

	handler := &testHandler{t}
	server := httptest.NewServer(handler)
	action := model.PostAction{
		Id:   model.NewId(),
		Name: "Test-action",
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL: server.URL,
			Context: map[string]interface{}{
				"test-key": "test-value",
			},
		},
	}

	post := &model.Post{
		Id:        model.NewId(),
		Type:      model.POST_EPHEMERAL,
		UserId:    th.BasicUser.Id,
		ChannelId: th.BasicChannel.Id,
		CreateAt:  model.GetMillis(),
		UpdateAt:  model.GetMillis(),
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{
				{
					Title:     "some-title",
					TitleLink: "https://some-url.com",
					Text:      "some-text",
					ImageURL:  "https://some-other-url.com",
					Actions:   []*model.PostAction{&action},
				},
			},
		},
	}

	post.GenerateActionIds()
	assert.Equal(t, 32, len(th.App.PostActionCookieSecret()))
	post = model.AddPostActionCookies(post, th.App.PostActionCookieSecret())

	ok, resp := Client.DoPostActionWithCookie(post.Id, action.Id, "", action.Cookie)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.RequestId)
	assert.NotNil(t, resp.ServerVersion)
}

func TestOpenDialog(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.ServiceSettings.AllowedUntrustedInternalConnections = "localhost 127.0.0.1"
	})

	_, triggerId, err := model.GenerateTriggerId(th.BasicUser.Id, th.App.AsymmetricSigningKey())
	require.Nil(t, err)

	request := model.OpenDialogRequest{
		TriggerId: triggerId,
		URL:       "http://localhost:8065",
		Dialog: model.Dialog{
			CallbackId: "callbackid",
			Title:      "Some Title",
			Elements: []model.DialogElement{
				model.DialogElement{
					DisplayName: "Element Name",
					Name:        "element_name",
					Type:        "text",
					Placeholder: "Enter a value",
				},
			},
			SubmitLabel:    "Submit",
			NotifyOnCancel: false,
			State:          "somestate",
		},
	}

	pass, resp := Client.OpenInteractiveDialog(request)
	CheckNoError(t, resp)
	assert.True(t, pass)

	// Should fail on bad trigger ID
	request.TriggerId = "junk"
	pass, resp = Client.OpenInteractiveDialog(request)
	CheckBadRequestStatus(t, resp)
	assert.False(t, pass)

	// URL is required
	request.TriggerId = triggerId
	request.URL = ""
	pass, resp = Client.OpenInteractiveDialog(request)
	CheckBadRequestStatus(t, resp)
	assert.False(t, pass)

	// Should pass with no elements
	request.URL = "http://localhost:8065"
	request.Dialog.Elements = nil
	pass, resp = Client.OpenInteractiveDialog(request)
	CheckNoError(t, resp)
	assert.True(t, pass)
	request.Dialog.Elements = []model.DialogElement{}
	pass, resp = Client.OpenInteractiveDialog(request)
	CheckNoError(t, resp)
	assert.True(t, pass)
}

func TestSubmitDialog(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.ServiceSettings.AllowedUntrustedInternalConnections = "localhost 127.0.0.1"
	})

	submit := model.SubmitDialogRequest{
		CallbackId: "callbackid",
		State:      "somestate",
		UserId:     th.BasicUser.Id,
		ChannelId:  th.BasicChannel.Id,
		TeamId:     th.BasicTeam.Id,
		Submission: map[string]interface{}{"somename": "somevalue"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request model.SubmitDialogRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.Nil(t, err)
		assert.NotNil(t, request)

		assert.Equal(t, request.URL, "")
		assert.Equal(t, request.UserId, submit.UserId)
		assert.Equal(t, request.ChannelId, submit.ChannelId)
		assert.Equal(t, request.TeamId, submit.TeamId)
		assert.Equal(t, request.CallbackId, submit.CallbackId)
		assert.Equal(t, request.State, submit.State)
		val, ok := request.Submission["somename"].(string)
		require.True(t, ok)
		assert.Equal(t, "somevalue", val)
	}))
	defer ts.Close()

	submit.URL = ts.URL

	submitResp, resp := Client.SubmitInteractiveDialog(submit)
	CheckNoError(t, resp)
	assert.NotNil(t, submitResp)

	submit.URL = ""
	submitResp, resp = Client.SubmitInteractiveDialog(submit)
	CheckBadRequestStatus(t, resp)
	assert.Nil(t, submitResp)

	submit.URL = ts.URL
	submit.ChannelId = model.NewId()
	submitResp, resp = Client.SubmitInteractiveDialog(submit)
	CheckForbiddenStatus(t, resp)
	assert.Nil(t, submitResp)

	submit.URL = ts.URL
	submit.ChannelId = th.BasicChannel.Id
	submit.TeamId = model.NewId()
	submitResp, resp = Client.SubmitInteractiveDialog(submit)
	CheckForbiddenStatus(t, resp)
	assert.Nil(t, submitResp)
}
