// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOAuthAccessTokenForImplicitFlow(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOAuthServiceProvider = true })

	oapp := &model.OAuthApp{
		Name:         "fakeoauthapp" + model.NewRandomString(10),
		CreatorId:    th.BasicUser2.Id,
		Homepage:     "https://nowhere.com",
		Description:  "test",
		CallbackUrls: []string{"https://nowhere.com"},
	}

	oapp, err := th.App.CreateOAuthApp(oapp)
	require.Nil(t, err)

	authRequest := &model.AuthorizeRequest{
		ResponseType: model.IMPLICIT_RESPONSE_TYPE,
		ClientId:     oapp.Id,
		RedirectUri:  oapp.CallbackUrls[0],
		Scope:        "",
		State:        "123",
	}

	session, err := th.App.GetOAuthAccessTokenForImplicitFlow(th.BasicUser.Id, authRequest)
	assert.Nil(t, err)
	assert.NotNil(t, session)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOAuthServiceProvider = false })

	session, err = th.App.GetOAuthAccessTokenForImplicitFlow(th.BasicUser.Id, authRequest)
	assert.NotNil(t, err, "should fail - oauth2 disabled")
	assert.Nil(t, session)

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableOAuthServiceProvider = true })
	authRequest.ClientId = "junk"

	session, err = th.App.GetOAuthAccessTokenForImplicitFlow(th.BasicUser.Id, authRequest)
	assert.NotNil(t, err, "should fail - bad client id")
	assert.Nil(t, session)

	authRequest.ClientId = oapp.Id

	session, err = th.App.GetOAuthAccessTokenForImplicitFlow("junk", authRequest)
	assert.NotNil(t, err, "should fail - bad user id")
	assert.Nil(t, session)
}

func TestOAuthRevokeAccessToken(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	if err := th.App.RevokeAccessToken(model.NewRandomString(16)); err == nil {
		t.Fatal("Should have failed bad token")
	}

	session := &model.Session{}
	session.CreateAt = model.GetMillis()
	session.UserId = model.NewId()
	session.Token = model.NewId()
	session.Roles = model.SYSTEM_USER_ROLE_ID
	session.SetExpireInDays(1)

	session, _ = th.App.CreateSession(session)
	if err := th.App.RevokeAccessToken(session.Token); err == nil {
		t.Fatal("Should have failed does not have an access token")
	}

	accessData := &model.AccessData{}
	accessData.Token = session.Token
	accessData.UserId = session.UserId
	accessData.RedirectUri = "http://example.com"
	accessData.ClientId = model.NewId()
	accessData.ExpiresAt = session.ExpiresAt

	if result := <-th.App.Srv.Store.OAuth().SaveAccessData(accessData); result.Err != nil {
		t.Fatal(result.Err)
	}

	if err := th.App.RevokeAccessToken(accessData.Token); err != nil {
		t.Fatal(err)
	}
}

func TestOAuthDeleteApp(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	*th.App.Config().ServiceSettings.EnableOAuthServiceProvider = true

	a1 := &model.OAuthApp{}
	a1.CreatorId = model.NewId()
	a1.Name = "TestApp" + model.NewId()
	a1.CallbackUrls = []string{"https://nowhere.com"}
	a1.Homepage = "https://nowhere.com"

	var err *model.AppError
	a1, err = th.App.CreateOAuthApp(a1)
	if err != nil {
		t.Fatal(err)
	}

	session := &model.Session{}
	session.CreateAt = model.GetMillis()
	session.UserId = model.NewId()
	session.Token = model.NewId()
	session.Roles = model.SYSTEM_USER_ROLE_ID
	session.IsOAuth = true
	session.SetExpireInDays(1)

	session, _ = th.App.CreateSession(session)

	accessData := &model.AccessData{}
	accessData.Token = session.Token
	accessData.UserId = session.UserId
	accessData.RedirectUri = "http://example.com"
	accessData.ClientId = a1.Id
	accessData.ExpiresAt = session.ExpiresAt

	if result := <-th.App.Srv.Store.OAuth().SaveAccessData(accessData); result.Err != nil {
		t.Fatal(result.Err)
	}

	if err := th.App.DeleteOAuthApp(a1.Id); err != nil {
		t.Fatal(err)
	}

	if _, err := th.App.GetSession(session.Token); err == nil {
		t.Fatal("should not get session from cache or db")
	}
}

func TestAuthorizeOAuthUser(t *testing.T) {
	setup := func(enable, tokenEndpoint, userEndpoint bool, serverURL string) *TestHelper {
		th := Setup(t)

		th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.GitLabSettings.Enable = enable

			if tokenEndpoint {
				*cfg.GitLabSettings.TokenEndpoint = serverURL + "/token"
			} else {
				*cfg.GitLabSettings.TokenEndpoint = ""
			}

			if userEndpoint {
				*cfg.GitLabSettings.UserApiEndpoint = serverURL + "/user"
			} else {
				*cfg.GitLabSettings.UserApiEndpoint = ""
			}
		})

		return th
	}

	makeState := func(token *model.Token) string {
		return base64.StdEncoding.EncodeToString([]byte(model.MapToJson(map[string]string{
			"token": token.Token,
		})))
	}

	makeToken := func(th *TestHelper, cookie string) *model.Token {
		token, _ := th.App.CreateOAuthStateToken(generateOAuthStateTokenExtra("", "", cookie))
		return token
	}

	makeRequest := func(t *testing.T, cookie string) *http.Request {
		request, _ := http.NewRequest(http.MethodGet, "https://mattermost.example.com", nil)

		if cookie != "" {
			request.AddCookie(&http.Cookie{
				Name:  COOKIE_OAUTH,
				Value: cookie,
			})
		}

		return request
	}

	t.Run("not enabled", func(t *testing.T) {
		th := setup(false, true, true, "")
		defer th.TearDown()

		_, _, _, err := th.App.AuthorizeOAuthUser(nil, nil, model.SERVICE_GITLAB, "", "", "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.unsupported.app_error", err.Id)
	})

	t.Run("with an improperly encoded state", func(t *testing.T) {
		th := setup(true, true, true, "")
		defer th.TearDown()

		state := "!"

		_, _, _, err := th.App.AuthorizeOAuthUser(nil, nil, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.invalid_state.app_error", err.Id)
	})

	t.Run("without a stored token", func(t *testing.T) {
		th := setup(true, true, true, "")
		defer th.TearDown()

		state := base64.StdEncoding.EncodeToString([]byte(model.MapToJson(map[string]string{
			"token": model.NewId(),
		})))

		_, _, _, err := th.App.AuthorizeOAuthUser(nil, nil, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.oauth.invalid_state_token.app_error", err.Id)
		assert.NotEqual(t, "", err.DetailedError)
	})

	t.Run("with a stored token of the wrong type", func(t *testing.T) {
		th := setup(true, true, true, "")
		defer th.TearDown()

		token := model.NewToken("invalid", "")
		result := <-th.App.Srv.Store.Token().Save(token)
		require.Nil(t, result.Err)

		state := makeState(token)

		_, _, _, err := th.App.AuthorizeOAuthUser(nil, nil, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.oauth.invalid_state_token.app_error", err.Id)
		assert.Equal(t, "", err.DetailedError)
	})

	t.Run("with email missing when changing login types", func(t *testing.T) {
		th := setup(true, true, true, "")
		defer th.TearDown()

		email := ""
		action := model.OAUTH_ACTION_EMAIL_TO_SSO
		cookie := model.NewId()

		token, err := th.App.CreateOAuthStateToken(generateOAuthStateTokenExtra(email, action, cookie))
		require.Nil(t, err)

		state := base64.StdEncoding.EncodeToString([]byte(model.MapToJson(map[string]string{
			"action": action,
			"email":  email,
			"token":  token.Token,
		})))

		_, _, _, err = th.App.AuthorizeOAuthUser(nil, nil, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.invalid_state.app_error", err.Id)
	})

	t.Run("without an OAuth cookie", func(t *testing.T) {
		th := setup(true, true, true, "")
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, "")
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(nil, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.invalid_state.app_error", err.Id)
	})

	t.Run("with an invalid token", func(t *testing.T) {
		th := setup(true, true, true, "")
		defer th.TearDown()

		cookie := model.NewId()

		token, err := th.App.CreateOAuthStateToken(model.NewId())
		require.Nil(t, err)

		request := makeRequest(t, cookie)
		state := makeState(token)

		_, _, _, err = th.App.AuthorizeOAuthUser(nil, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.invalid_state.app_error", err.Id)
	})

	t.Run("with an incorrect token endpoint", func(t *testing.T) {
		th := setup(true, false, true, "")
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.token_failed.app_error", err.Id)
	})

	t.Run("with an error token response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))
		defer server.Close()

		th := setup(true, true, true, server.URL)
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.bad_response.app_error", err.Id)
		assert.Contains(t, err.DetailedError, "status_code=418")
	})

	t.Run("with an invalid token response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid"))
		}))
		defer server.Close()

		th := setup(true, true, true, server.URL)
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.bad_response.app_error", err.Id)
		assert.Contains(t, err.DetailedError, "response_body=invalid")
	})

	t.Run("with an invalid token type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(&model.AccessResponse{
				AccessToken: model.NewId(),
				TokenType:   "",
			})
		}))
		defer server.Close()

		th := setup(true, true, true, server.URL)
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.bad_token.app_error", err.Id)
	})

	t.Run("with an empty token response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(&model.AccessResponse{
				AccessToken: "",
				TokenType:   model.ACCESS_TOKEN_TYPE,
			})
		}))
		defer server.Close()

		th := setup(true, true, true, server.URL)
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.missing.app_error", err.Id)
	})

	t.Run("with an incorrect user endpoint", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(&model.AccessResponse{
				AccessToken: model.NewId(),
				TokenType:   model.ACCESS_TOKEN_TYPE,
			})
		}))
		defer server.Close()

		th := setup(true, true, false, server.URL)
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.service.app_error", err.Id)
	})

	t.Run("with an error user response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/token":
				t.Log("hit token")
				json.NewEncoder(w).Encode(&model.AccessResponse{
					AccessToken: model.NewId(),
					TokenType:   model.ACCESS_TOKEN_TYPE,
				})
			case "/user":
				t.Log("hit user")
				w.WriteHeader(http.StatusTeapot)
			}
		}))
		defer server.Close()

		th := setup(true, true, true, server.URL)
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.authorize_oauth_user.response.app_error", err.Id)
	})

	t.Run("with an error user response due to GitLab TOS", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/token":
				t.Log("hit token")
				json.NewEncoder(w).Encode(&model.AccessResponse{
					AccessToken: model.NewId(),
					TokenType:   model.ACCESS_TOKEN_TYPE,
				})
			case "/user":
				t.Log("hit user")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Terms of Service"))
			}
		}))
		defer server.Close()

		th := setup(true, true, true, server.URL)
		defer th.TearDown()

		cookie := model.NewId()
		request := makeRequest(t, cookie)
		state := makeState(makeToken(th, cookie))

		_, _, _, err := th.App.AuthorizeOAuthUser(&httptest.ResponseRecorder{}, request, model.SERVICE_GITLAB, "", state, "")
		require.NotNil(t, err)
		assert.Equal(t, "oauth.gitlab.tos.error", err.Id)
	})

	t.Run("enabled and properly configured", func(t *testing.T) {
		testCases := []struct {
			Description                   string
			SiteURL                       string
			ExpectedSetCookieHeaderRegexp string
		}{
			{"no subpath", "http://localhost:8065", "^MMOAUTH=; Path=/"},
			{"subpath", "http://localhost:8065/subpath", "^MMOAUTH=; Path=/subpath"},
		}

		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				userData := "Hello, World!"

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/token":
						json.NewEncoder(w).Encode(&model.AccessResponse{
							AccessToken: model.NewId(),
							TokenType:   model.ACCESS_TOKEN_TYPE,
						})
					case "/user":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(userData))
					}
				}))
				defer server.Close()

				th := setup(true, true, true, server.URL)
				defer th.TearDown()

				th.App.UpdateConfig(func(cfg *model.Config) {
					*cfg.ServiceSettings.SiteURL = tc.SiteURL
				})

				cookie := model.NewId()
				request := makeRequest(t, cookie)

				stateProps := map[string]string{
					"team_id": model.NewId(),
					"token":   makeToken(th, cookie).Token,
				}
				state := base64.StdEncoding.EncodeToString([]byte(model.MapToJson(stateProps)))

				recorder := httptest.ResponseRecorder{}
				body, receivedTeamId, receivedStateProps, err := th.App.AuthorizeOAuthUser(&recorder, request, model.SERVICE_GITLAB, "", state, "")

				require.NotNil(t, body)
				bodyBytes, bodyErr := ioutil.ReadAll(body)
				require.Nil(t, bodyErr)
				assert.Equal(t, userData, string(bodyBytes))

				assert.Equal(t, stateProps["team_id"], receivedTeamId)
				assert.Equal(t, stateProps, receivedStateProps)
				assert.Nil(t, err)

				cookies := recorder.Header().Get("Set-Cookie")
				assert.Regexp(t, tc.ExpectedSetCookieHeaderRegexp, cookies)
			})
		}
	})
}

func TestGetAuthorizationCode(t *testing.T) {
	t.Run("not enabled", func(t *testing.T) {
		th := Setup(t)
		defer th.TearDown()

		th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.GitLabSettings.Enable = false
		})

		_, err := th.App.GetAuthorizationCode(nil, nil, model.SERVICE_GITLAB, map[string]string{}, "")
		require.NotNil(t, err)
		assert.Equal(t, "api.user.get_authorization_code.unsupported.app_error", err.Id)
	})

	t.Run("enabled and properly configured", func(t *testing.T) {
		th := Setup(t)
		defer th.TearDown()

		th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.GitLabSettings.Enable = true
		})

		testCases := []struct {
			Description                   string
			SiteURL                       string
			ExpectedSetCookieHeaderRegexp string
		}{
			{"no subpath", "http://localhost:8065", "^MMOAUTH=[a-z0-9]+; Path=/"},
			{"subpath", "http://localhost:8065/subpath", "^MMOAUTH=[a-z0-9]+; Path=/subpath"},
		}

		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				th.App.UpdateConfig(func(cfg *model.Config) {
					*cfg.ServiceSettings.SiteURL = tc.SiteURL
				})

				request, _ := http.NewRequest(http.MethodGet, "https://mattermost.example.com", nil)

				stateProps := map[string]string{
					"email":  "email@example.com",
					"action": "action",
				}

				recorder := httptest.ResponseRecorder{}
				url, err := th.App.GetAuthorizationCode(&recorder, request, model.SERVICE_GITLAB, stateProps, "")
				require.Nil(t, err)
				assert.NotEmpty(t, url)

				cookies := recorder.Header().Get("Set-Cookie")
				assert.Regexp(t, tc.ExpectedSetCookieHeaderRegexp, cookies)
			})
		}
	})
}
