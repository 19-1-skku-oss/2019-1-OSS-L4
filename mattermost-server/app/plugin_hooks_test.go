// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
	"github.com/mattermost/mattermost-server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetAppEnvironmentWithPlugins(t *testing.T, pluginCode []string, app *App, apiFunc func(*model.Manifest) plugin.API) (func(), []string, []error) {
	pluginDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	webappPluginDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	env, err := plugin.NewEnvironment(apiFunc, pluginDir, webappPluginDir, app.Log)
	require.NoError(t, err)

	app.SetPluginsEnvironment(env)
	pluginIds := []string{}
	activationErrors := []error{}
	for _, code := range pluginCode {
		pluginId := model.NewId()
		backend := filepath.Join(pluginDir, pluginId, "backend.exe")
		utils.CompileGo(t, code, backend)

		ioutil.WriteFile(filepath.Join(pluginDir, pluginId, "plugin.json"), []byte(`{"id": "`+pluginId+`", "backend": {"executable": "backend.exe"}}`), 0600)
		_, _, activationErr := env.Activate(pluginId)
		pluginIds = append(pluginIds, pluginId)
		activationErrors = append(activationErrors, activationErr)
	}

	return func() {
		os.RemoveAll(pluginDir)
		os.RemoveAll(webappPluginDir)
	}, pluginIds, activationErrors
}

func TestHookMessageWillBePosted(t *testing.T) {
	t.Run("rejected", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
				return nil, "rejected"
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, th.App.NewPluginAPI)
		defer tearDown()

		post := &model.Post{
			UserId:    th.BasicUser.Id,
			ChannelId: th.BasicChannel.Id,
			Message:   "message_",
			CreateAt:  model.GetMillis() - 10000,
		}
		_, err := th.App.CreatePost(post, th.BasicChannel, false)
		if assert.NotNil(t, err) {
			assert.Equal(t, "Post rejected by plugin. rejected", err.Message)
		}
	})

	t.Run("rejected, returned post ignored", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
				post.Message = "ignored"
				return post, "rejected"
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, th.App.NewPluginAPI)
		defer tearDown()

		post := &model.Post{
			UserId:    th.BasicUser.Id,
			ChannelId: th.BasicChannel.Id,
			Message:   "message_",
			CreateAt:  model.GetMillis() - 10000,
		}
		_, err := th.App.CreatePost(post, th.BasicChannel, false)
		if assert.NotNil(t, err) {
			assert.Equal(t, "Post rejected by plugin. rejected", err.Message)
		}
	})

	t.Run("allowed", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
				return nil, ""
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, th.App.NewPluginAPI)
		defer tearDown()

		post := &model.Post{
			UserId:    th.BasicUser.Id,
			ChannelId: th.BasicChannel.Id,
			Message:   "message",
			CreateAt:  model.GetMillis() - 10000,
		}
		post, err := th.App.CreatePost(post, th.BasicChannel, false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "message", post.Message)
		if result := <-th.App.Srv.Store.Post().GetSingle(post.Id); result.Err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "message", result.Data.(*model.Post).Message)
		}
	})

	t.Run("updated", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
				post.Message = post.Message + "_fromplugin"
				return post, ""
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, th.App.NewPluginAPI)
		defer tearDown()

		post := &model.Post{
			UserId:    th.BasicUser.Id,
			ChannelId: th.BasicChannel.Id,
			Message:   "message",
			CreateAt:  model.GetMillis() - 10000,
		}
		post, err := th.App.CreatePost(post, th.BasicChannel, false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "message_fromplugin", post.Message)
		if result := <-th.App.Srv.Store.Post().GetSingle(post.Id); result.Err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "message_fromplugin", result.Data.(*model.Post).Message)
		}
	})

	t.Run("multiple updated", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
				
				post.Message = "prefix_" + post.Message
				return post, ""
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
			`
			package main

			import (
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
				post.Message = post.Message + "_suffix"
				return post, ""
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, th.App.NewPluginAPI)
		defer tearDown()

		post := &model.Post{
			UserId:    th.BasicUser.Id,
			ChannelId: th.BasicChannel.Id,
			Message:   "message",
			CreateAt:  model.GetMillis() - 10000,
		}
		post, err := th.App.CreatePost(post, th.BasicChannel, false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "prefix_message_suffix", post.Message)
	})
}

func TestHookMessageHasBeenPosted(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	var mockAPI plugintest.API
	mockAPI.On("LoadPluginConfiguration", mock.Anything).Return(nil)
	mockAPI.On("LogDebug", "message").Return(nil)

	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
			p.API.LogDebug(post.Message)
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, func(*model.Manifest) plugin.API { return &mockAPI })
	defer tearDown()

	post := &model.Post{
		UserId:    th.BasicUser.Id,
		ChannelId: th.BasicChannel.Id,
		Message:   "message",
		CreateAt:  model.GetMillis() - 10000,
	}
	_, err := th.App.CreatePost(post, th.BasicChannel, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHookMessageWillBeUpdated(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) MessageWillBeUpdated(c *plugin.Context, newPost, oldPost *model.Post) (*model.Post, string) {
			newPost.Message = newPost.Message + "fromplugin"
			return newPost, ""
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, th.App.NewPluginAPI)
	defer tearDown()

	post := &model.Post{
		UserId:    th.BasicUser.Id,
		ChannelId: th.BasicChannel.Id,
		Message:   "message_",
		CreateAt:  model.GetMillis() - 10000,
	}
	post, err := th.App.CreatePost(post, th.BasicChannel, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "message_", post.Message)
	post.Message = post.Message + "edited_"
	post, err = th.App.UpdatePost(post, true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "message_edited_fromplugin", post.Message)
}

func TestHookMessageHasBeenUpdated(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	var mockAPI plugintest.API
	mockAPI.On("LoadPluginConfiguration", mock.Anything).Return(nil)
	mockAPI.On("LogDebug", "message_edited").Return(nil)
	mockAPI.On("LogDebug", "message_").Return(nil)
	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) MessageHasBeenUpdated(c *plugin.Context, newPost, oldPost *model.Post) {
			p.API.LogDebug(newPost.Message)
			p.API.LogDebug(oldPost.Message)
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, func(*model.Manifest) plugin.API { return &mockAPI })
	defer tearDown()

	post := &model.Post{
		UserId:    th.BasicUser.Id,
		ChannelId: th.BasicChannel.Id,
		Message:   "message_",
		CreateAt:  model.GetMillis() - 10000,
	}
	post, err := th.App.CreatePost(post, th.BasicChannel, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "message_", post.Message)
	post.Message = post.Message + "edited"
	_, err = th.App.UpdatePost(post, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHookFileWillBeUploaded(t *testing.T) {
	t.Run("rejected", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		var mockAPI plugintest.API
		mockAPI.On("LoadPluginConfiguration", mock.Anything).Return(nil)
		mockAPI.On("LogDebug", "testhook.txt").Return(nil)
		mockAPI.On("LogDebug", "inputfile").Return(nil)
		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"io"
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) FileWillBeUploaded(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string) {
				return nil, "rejected"
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, func(*model.Manifest) plugin.API { return &mockAPI })
		defer tearDown()

		_, err := th.App.UploadFiles(
			"noteam",
			th.BasicChannel.Id,
			th.BasicUser.Id,
			[]io.ReadCloser{ioutil.NopCloser(bytes.NewBufferString("inputfile"))},
			[]string{"testhook.txt"},
			[]string{},
			time.Now(),
		)
		if assert.NotNil(t, err) {
			assert.Equal(t, "File rejected by plugin. rejected", err.Message)
		}
	})

	t.Run("rejected, returned file ignored", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		var mockAPI plugintest.API
		mockAPI.On("LoadPluginConfiguration", mock.Anything).Return(nil)
		mockAPI.On("LogDebug", "testhook.txt").Return(nil)
		mockAPI.On("LogDebug", "inputfile").Return(nil)
		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"io"
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) FileWillBeUploaded(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string) {
				output.Write([]byte("ignored"))
				info.Name = "ignored"
				return info, "rejected"
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, func(*model.Manifest) plugin.API { return &mockAPI })
		defer tearDown()

		_, err := th.App.UploadFiles(
			"noteam",
			th.BasicChannel.Id,
			th.BasicUser.Id,
			[]io.ReadCloser{ioutil.NopCloser(bytes.NewBufferString("inputfile"))},
			[]string{"testhook.txt"},
			[]string{},
			time.Now(),
		)
		if assert.NotNil(t, err) {
			assert.Equal(t, "File rejected by plugin. rejected", err.Message)
		}
	})

	t.Run("allowed", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		var mockAPI plugintest.API
		mockAPI.On("LoadPluginConfiguration", mock.Anything).Return(nil)
		mockAPI.On("LogDebug", "testhook.txt").Return(nil)
		mockAPI.On("LogDebug", "inputfile").Return(nil)
		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"io"
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) FileWillBeUploaded(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string) {
				return nil, ""
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, func(*model.Manifest) plugin.API { return &mockAPI })
		defer tearDown()

		response, err := th.App.UploadFiles(
			"noteam",
			th.BasicChannel.Id,
			th.BasicUser.Id,
			[]io.ReadCloser{ioutil.NopCloser(bytes.NewBufferString("inputfile"))},
			[]string{"testhook.txt"},
			[]string{},
			time.Now(),
		)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, len(response.FileInfos))
		fileId := response.FileInfos[0].Id

		fileInfo, err := th.App.GetFileInfo(fileId)
		assert.Nil(t, err)
		assert.NotNil(t, fileInfo)
		assert.Equal(t, "testhook.txt", fileInfo.Name)

		fileReader, err := th.App.FileReader(fileInfo.Path)
		assert.Nil(t, err)
		var resultBuf bytes.Buffer
		io.Copy(&resultBuf, fileReader)
		assert.Equal(t, "inputfile", resultBuf.String())
	})

	t.Run("updated", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		var mockAPI plugintest.API
		mockAPI.On("LoadPluginConfiguration", mock.Anything).Return(nil)
		mockAPI.On("LogDebug", "testhook.txt").Return(nil)
		mockAPI.On("LogDebug", "inputfile").Return(nil)
		tearDown, _, _ := SetAppEnvironmentWithPlugins(t, []string{
			`
			package main

			import (
				"io"
				"bytes"
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) FileWillBeUploaded(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string) {
				p.API.LogDebug(info.Name)
				var buf bytes.Buffer
				buf.ReadFrom(file)
				p.API.LogDebug(buf.String())

				outbuf := bytes.NewBufferString("changedtext")
				io.Copy(output, outbuf)
				info.Name = "modifiedinfo"
				return info, ""
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
			`,
		}, th.App, func(*model.Manifest) plugin.API { return &mockAPI })
		defer tearDown()

		response, err := th.App.UploadFiles(
			"noteam",
			th.BasicChannel.Id,
			th.BasicUser.Id,
			[]io.ReadCloser{ioutil.NopCloser(bytes.NewBufferString("inputfile"))},
			[]string{"testhook.txt"},
			[]string{},
			time.Now(),
		)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, len(response.FileInfos))
		fileId := response.FileInfos[0].Id

		fileInfo, err := th.App.GetFileInfo(fileId)
		assert.Nil(t, err)
		assert.NotNil(t, fileInfo)
		assert.Equal(t, "modifiedinfo", fileInfo.Name)

		fileReader, err := th.App.FileReader(fileInfo.Path)
		assert.Nil(t, err)
		var resultBuf bytes.Buffer
		io.Copy(&resultBuf, fileReader)
		assert.Equal(t, "changedtext", resultBuf.String())
	})
}

func TestUserWillLogIn_Blocked(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	err := th.App.UpdatePassword(th.BasicUser, "hunter2")

	if err != nil {
		t.Errorf("Error updating user password: %s", err)
	}

	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) UserWillLogIn(c *plugin.Context, user *model.User) string {
			return "Blocked By Plugin"
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, th.App.NewPluginAPI)
	defer tearDown()

	r := &http.Request{}
	w := httptest.NewRecorder()
	_, err = th.App.DoLogin(w, r, th.BasicUser, "")

	if !strings.HasPrefix(err.Id, "Login rejected by plugin") {
		t.Errorf("Expected Login rejected by plugin, got %s", err.Id)
	}
}

func TestUserWillLogInIn_Passed(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	err := th.App.UpdatePassword(th.BasicUser, "hunter2")

	if err != nil {
		t.Errorf("Error updating user password: %s", err)
	}

	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) UserWillLogIn(c *plugin.Context, user *model.User) string {
			return ""
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, th.App.NewPluginAPI)
	defer tearDown()

	r := &http.Request{}
	w := httptest.NewRecorder()
	session, err := th.App.DoLogin(w, r, th.BasicUser, "")

	if err != nil {
		t.Errorf("Expected nil, got %s", err)
	}

	if session.UserId != th.BasicUser.Id {
		t.Errorf("Expected %s, got %s", th.BasicUser.Id, session.UserId)
	}
}

func TestUserHasLoggedIn(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	err := th.App.UpdatePassword(th.BasicUser, "hunter2")

	if err != nil {
		t.Errorf("Error updating user password: %s", err)
	}

	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) UserHasLoggedIn(c *plugin.Context, user *model.User) {
			user.FirstName = "plugin-callback-success"
			p.API.UpdateUser(user)
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, th.App.NewPluginAPI)
	defer tearDown()

	r := &http.Request{}
	w := httptest.NewRecorder()
	_, err = th.App.DoLogin(w, r, th.BasicUser, "")

	if err != nil {
		t.Errorf("Expected nil, got %s", err)
	}

	time.Sleep(2 * time.Second)

	user, _ := th.App.GetUser(th.BasicUser.Id)

	if user.FirstName != "plugin-callback-success" {
		t.Errorf("Expected firstname overwrite, got default")
	}
}

func TestUserHasBeenCreated(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) UserHasBeenCreated(c *plugin.Context, user *model.User) {
			user.Nickname = "plugin-callback-success"
			p.API.UpdateUser(user)
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, th.App.NewPluginAPI)
	defer tearDown()

	user := &model.User{
		Email:       model.NewId() + "success+test@example.com",
		Nickname:    "Darth Vader",
		Username:    "vader" + model.NewId(),
		Password:    "passwd1",
		AuthService: "",
	}
	_, err := th.App.CreateUser(user)
	require.Nil(t, err)

	time.Sleep(1 * time.Second)

	user, err = th.App.GetUser(user.Id)
	require.Nil(t, err)

	require.Equal(t, "plugin-callback-success", user.Nickname)
}

func TestErrorString(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	t.Run("errors.New", func(t *testing.T) {
		tearDown, _, activationErrors := SetAppEnvironmentWithPlugins(t,
			[]string{
				`
			package main

			import (
				"errors"

				"github.com/mattermost/mattermost-server/plugin"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) OnActivate() error {
				return errors.New("simulate failure")
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
		`}, th.App, th.App.NewPluginAPI)
		defer tearDown()

		require.Len(t, activationErrors, 1)
		require.NotNil(t, activationErrors[0])
		require.Contains(t, activationErrors[0].Error(), "simulate failure")
	})

	t.Run("AppError", func(t *testing.T) {
		tearDown, _, activationErrors := SetAppEnvironmentWithPlugins(t,
			[]string{
				`
			package main

			import (
				"github.com/mattermost/mattermost-server/plugin"
				"github.com/mattermost/mattermost-server/model"
			)

			type MyPlugin struct {
				plugin.MattermostPlugin
			}

			func (p *MyPlugin) OnActivate() error {
				return model.NewAppError("where", "id", map[string]interface{}{"param": 1}, "details", 42)
			}

			func main() {
				plugin.ClientMain(&MyPlugin{})
			}
		`}, th.App, th.App.NewPluginAPI)
		defer tearDown()

		require.Len(t, activationErrors, 1)
		require.NotNil(t, activationErrors[0])

		cause := errors.Cause(activationErrors[0])
		require.IsType(t, &model.AppError{}, cause)

		// params not expected, since not exported
		expectedErr := model.NewAppError("where", "id", nil, "details", 42)
		require.Equal(t, expectedErr, cause)
	})
}

func TestHookContext(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// We don't actually have a session, we are faking it so just set something arbitrarily
	th.App.Session.Id = model.NewId()
	th.App.RequestId = model.NewId()
	th.App.IpAddress = model.NewId()
	th.App.AcceptLanguage = model.NewId()
	th.App.UserAgent = model.NewId()

	var mockAPI plugintest.API
	mockAPI.On("LoadPluginConfiguration", mock.Anything).Return(nil)
	mockAPI.On("LogDebug", th.App.Session.Id).Return(nil)
	mockAPI.On("LogInfo", th.App.RequestId).Return(nil)
	mockAPI.On("LogError", th.App.IpAddress).Return(nil)
	mockAPI.On("LogWarn", th.App.AcceptLanguage).Return(nil)
	mockAPI.On("DeleteTeam", th.App.UserAgent).Return(nil)

	tearDown, _, _ := SetAppEnvironmentWithPlugins(t,
		[]string{
			`
		package main

		import (
			"github.com/mattermost/mattermost-server/plugin"
			"github.com/mattermost/mattermost-server/model"
		)

		type MyPlugin struct {
			plugin.MattermostPlugin
		}

		func (p *MyPlugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
			p.API.LogDebug(c.SessionId)
			p.API.LogInfo(c.RequestId)
			p.API.LogError(c.IpAddress)
			p.API.LogWarn(c.AcceptLanguage)
			p.API.DeleteTeam(c.UserAgent)
		}

		func main() {
			plugin.ClientMain(&MyPlugin{})
		}
	`}, th.App, func(*model.Manifest) plugin.API { return &mockAPI })
	defer tearDown()

	post := &model.Post{
		UserId:    th.BasicUser.Id,
		ChannelId: th.BasicChannel.Id,
		Message:   "not this",
		CreateAt:  model.GetMillis() - 10000,
	}
	_, err := th.App.CreatePost(post, th.BasicChannel, false)
	if err != nil {
		t.Fatal(err)
	}
}
