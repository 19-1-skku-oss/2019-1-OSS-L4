// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-server/app"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils"
)

func (w *Web) NewHandler(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	return &Handler{
		GetGlobalAppOptions: w.GetGlobalAppOptions,
		HandleFunc:          h,
		RequireSession:      false,
		TrustRequester:      false,
		RequireMfa:          false,
		IsStatic:            false,
	}
}

func (w *Web) NewStaticHandler(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	// Determine the CSP SHA directive needed for subpath support, if any. This value is fixed
	// on server start and intentionally requires a restart to take effect.
	subpath, _ := utils.GetSubpathFromConfig(w.ConfigService.Config())

	return &Handler{
		GetGlobalAppOptions: w.GetGlobalAppOptions,
		HandleFunc:          h,
		RequireSession:      false,
		TrustRequester:      false,
		RequireMfa:          false,
		IsStatic:            true,

		cspShaDirective: utils.GetSubpathScriptHash(subpath),
	}
}

type Handler struct {
	GetGlobalAppOptions app.AppOptionCreator
	HandleFunc          func(*Context, http.ResponseWriter, *http.Request)
	RequireSession      bool
	TrustRequester      bool
	RequireMfa          bool
	IsStatic            bool

	cspShaDirective string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	mlog.Debug(fmt.Sprintf("%v - %v", r.Method, r.URL.Path))

	c := &Context{}
	c.App = app.New(
		h.GetGlobalAppOptions()...,
	)
	c.App.T, _ = utils.GetTranslationsAndLocale(w, r)
	c.App.RequestId = model.NewId()
	c.App.IpAddress = utils.GetIpAddress(r)
	c.App.UserAgent = r.UserAgent()
	c.App.AcceptLanguage = r.Header.Get("Accept-Language")
	c.Params = ParamsFromRequest(r)
	c.App.Path = r.URL.Path
	c.Log = c.App.Log

	subpath, _ := utils.GetSubpathFromConfig(c.App.Config())
	siteURLHeader := app.GetProtocol(r) + "://" + r.Host + subpath
	c.SetSiteURLHeader(siteURLHeader)

	w.Header().Set(model.HEADER_REQUEST_ID, c.App.RequestId)
	w.Header().Set(model.HEADER_VERSION_ID, fmt.Sprintf("%v.%v.%v.%v", model.CurrentVersion, model.BuildNumber, c.App.ClientConfigHash(), c.App.License() != nil))

	if *c.App.Config().ServiceSettings.TLSStrictTransport {
		w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d", *c.App.Config().ServiceSettings.TLSStrictTransportMaxAge))
	}

	if h.IsStatic {
		// Instruct the browser not to display us in an iframe unless is the same origin for anti-clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		// Set content security policy. This is also specified in the root.html of the webapp in a meta tag.
		w.Header().Set("Content-Security-Policy", fmt.Sprintf(
			"frame-ancestors 'self'; script-src 'self' cdn.segment.com/analytics.js/%s",
			h.cspShaDirective,
		))
	} else {
		// All api response bodies will be JSON formatted by default
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			w.Header().Set("Expires", "0")
		}
	}

	token, tokenLocation := app.ParseAuthTokenFromRequest(r)

	if len(token) != 0 {
		session, err := c.App.GetSession(token)
		if err != nil {
			c.Log.Info("Invalid session", mlog.Err(err))
			if err.StatusCode == http.StatusInternalServerError {
				c.Err = err
			} else if h.RequireSession {
				c.RemoveSessionCookie(w, r)
				c.Err = model.NewAppError("ServeHTTP", "api.context.session_expired.app_error", nil, "token="+token, http.StatusUnauthorized)
			}
		} else if !session.IsOAuth && tokenLocation == app.TokenLocationQueryString {
			c.Err = model.NewAppError("ServeHTTP", "api.context.token_provided.app_error", nil, "token="+token, http.StatusUnauthorized)
		} else {
			c.App.Session = *session
		}

		// Rate limit by UserID
		if c.App.Srv.RateLimiter != nil && c.App.Srv.RateLimiter.UserIdRateLimit(c.App.Session.UserId, w) {
			return
		}

		csrfCheckPassed := false

		// CSRF Check
		if c.Err == nil && tokenLocation == app.TokenLocationCookie && h.RequireSession && !h.TrustRequester && r.Method != "GET" {
			csrfHeader := r.Header.Get(model.HEADER_CSRF_TOKEN)
			if csrfHeader == session.GetCSRF() {
				csrfCheckPassed = true
			} else if r.Header.Get(model.HEADER_REQUESTED_WITH) == model.HEADER_REQUESTED_WITH_XML {
				// ToDo(DSchalla) 2019/01/04: Remove after deprecation period and only allow CSRF Header (MM-13657)
				csrfErrorMessage := "CSRF Header check failed for request - Please upgrade your web application or custom app to set a CSRF Header"
				if *c.App.Config().ServiceSettings.ExperimentalStrictCSRFEnforcement {
					c.Log.Warn(csrfErrorMessage)
				} else {
					c.Log.Debug(csrfErrorMessage)
					csrfCheckPassed = true
				}
			}

			if !csrfCheckPassed {
				token = ""
				c.App.Session = model.Session{}
				c.Err = model.NewAppError("ServeHTTP", "api.context.session_expired.app_error", nil, "token="+token+" Appears to be a CSRF attempt", http.StatusUnauthorized)
			}
		}
	}

	c.Log = c.App.Log.With(
		mlog.String("path", c.App.Path),
		mlog.String("request_id", c.App.RequestId),
		mlog.String("ip_addr", c.App.IpAddress),
		mlog.String("user_id", c.App.Session.UserId),
		mlog.String("method", r.Method),
	)

	if c.Err == nil && h.RequireSession {
		c.SessionRequired()
	}

	if c.Err == nil && h.RequireMfa {
		c.MfaRequired()
	}

	if c.Err == nil {
		h.HandleFunc(c, w, r)
	}

	// Handle errors that have occurred
	if c.Err != nil {
		c.Err.Translate(c.App.T)
		c.Err.RequestId = c.App.RequestId

		if c.Err.Id == "api.context.session_expired.app_error" {
			c.LogInfo(c.Err)
		} else {
			c.LogError(c.Err)
		}

		c.Err.Where = r.URL.Path

		// Block out detailed error when not in developer mode
		if !*c.App.Config().ServiceSettings.EnableDeveloper {
			c.Err.DetailedError = ""
		}

		// Sanitize all 5xx error messages in hardened mode
		if *c.App.Config().ServiceSettings.ExperimentalEnableHardenedMode && c.Err.StatusCode >= 500 {
			c.Err.Id = ""
			c.Err.Message = "Internal Server Error"
			c.Err.DetailedError = ""
			c.Err.StatusCode = 500
			c.Err.Where = ""
			c.Err.IsOAuth = false
		}

		if IsApiCall(c.App, r) || IsWebhookCall(c.App, r) || len(r.Header.Get("X-Mobile-App")) > 0 {
			w.WriteHeader(c.Err.StatusCode)
			w.Write([]byte(c.Err.ToJson()))
		} else {
			utils.RenderWebAppError(c.App.Config(), w, r, c.Err, c.App.AsymmetricSigningKey())
		}

		if c.App.Metrics != nil {
			c.App.Metrics.IncrementHttpError()
		}
	}

	if c.App.Metrics != nil {
		c.App.Metrics.IncrementHttpRequest()

		if r.URL.Path != model.API_URL_SUFFIX+"/websocket" {
			elapsed := float64(time.Since(now)) / float64(time.Second)
			c.App.Metrics.ObserveHttpRequestDuration(elapsed)
		}
	}
}
