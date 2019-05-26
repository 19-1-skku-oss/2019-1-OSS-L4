// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api4

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/model"
)

func TestGetImage(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	// Prevent the test client from following a redirect
	th.Client.HttpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	t.Run("proxy disabled", func(t *testing.T) {
		imageURL := "http://foo.bar/baz.gif"

		th.App.UpdateConfig(func(cfg *model.Config) {
			cfg.ImageProxySettings.Enable = model.NewBool(false)
		})

		r, err := http.NewRequest("GET", th.Client.ApiUrl+"/image?url="+url.QueryEscape(imageURL), nil)
		require.NoError(t, err)
		r.Header.Set(model.HEADER_AUTH, th.Client.AuthType+" "+th.Client.AuthToken)

		resp, err := th.Client.HttpClient.Do(r)
		require.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, imageURL, resp.Header.Get("Location"))
	})

	t.Run("atmos/camo", func(t *testing.T) {
		imageURL := "http://foo.bar/baz.gif"
		proxiedURL := "https://proxy.foo.bar/004afe2ef382eb5f30c4490f793f8a8c5b33d8a2/687474703a2f2f666f6f2e6261722f62617a2e676966"

		th.App.UpdateConfig(func(cfg *model.Config) {
			cfg.ImageProxySettings.Enable = model.NewBool(true)
			cfg.ImageProxySettings.ImageProxyType = model.NewString("atmos/camo")
			cfg.ImageProxySettings.RemoteImageProxyOptions = model.NewString("foo")
			cfg.ImageProxySettings.RemoteImageProxyURL = model.NewString("https://proxy.foo.bar")
		})

		r, err := http.NewRequest("GET", th.Client.ApiUrl+"/image?url="+url.QueryEscape(imageURL), nil)
		require.NoError(t, err)
		r.Header.Set(model.HEADER_AUTH, th.Client.AuthType+" "+th.Client.AuthToken)

		resp, err := th.Client.HttpClient.Do(r)
		require.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, proxiedURL, resp.Header.Get("Location"))
	})

	t.Run("local", func(t *testing.T) {
		th.App.UpdateConfig(func(cfg *model.Config) {
			cfg.ImageProxySettings.Enable = model.NewBool(true)
			cfg.ImageProxySettings.ImageProxyType = model.NewString("local")

			// Allow requests to the "remote" image
			cfg.ServiceSettings.AllowedUntrustedInternalConnections = model.NewString("127.0.0.1")
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("success"))
		})

		imageServer := httptest.NewServer(handler)
		defer imageServer.Close()

		r, err := http.NewRequest("GET", th.Client.ApiUrl+"/image?url="+url.QueryEscape(imageServer.URL+"/image.png"), nil)
		require.NoError(t, err)
		r.Header.Set(model.HEADER_AUTH, th.Client.AuthType+" "+th.Client.AuthToken)

		resp, err := th.Client.HttpClient.Do(r)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "success", string(respBody))
	})
}
