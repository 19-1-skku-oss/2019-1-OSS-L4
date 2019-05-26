// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api4

import (
	"net/http"

	"github.com/mattermost/mattermost-server/model"
)

func (api *API) InitBot() {
	api.BaseRoutes.Bots.Handle("", api.ApiSessionRequired(createBot)).Methods("POST")
	api.BaseRoutes.Bot.Handle("", api.ApiSessionRequired(patchBot)).Methods("PUT")
	api.BaseRoutes.Bot.Handle("", api.ApiSessionRequired(getBot)).Methods("GET")
	api.BaseRoutes.Bots.Handle("", api.ApiSessionRequired(getBots)).Methods("GET")
	api.BaseRoutes.Bot.Handle("/disable", api.ApiSessionRequired(disableBot)).Methods("POST")
	api.BaseRoutes.Bot.Handle("/enable", api.ApiSessionRequired(enableBot)).Methods("POST")
	api.BaseRoutes.Bot.Handle("/assign/{user_id:[A-Za-z0-9]+}", api.ApiSessionRequired(assignBot)).Methods("POST")
}

func createBot(c *Context, w http.ResponseWriter, r *http.Request) {
	botPatch := model.BotPatchFromJson(r.Body)
	if botPatch == nil {
		c.SetInvalidParam("bot")
		return
	}

	bot := &model.Bot{
		OwnerId: c.App.Session.UserId,
	}
	bot.Patch(botPatch)

	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_CREATE_BOT) {
		c.SetPermissionError(model.PERMISSION_CREATE_BOT)
		return
	}

	if user, err := c.App.GetUser(c.App.Session.UserId); err == nil {
		if user.IsBot {
			c.SetPermissionError(model.PERMISSION_CREATE_BOT)
			return
		}
	}

	if !*c.App.Config().ServiceSettings.EnableBotAccountCreation {
		c.Err = model.NewAppError("createBot", "api.bot.create_disabled", nil, "", http.StatusForbidden)
		return
	}

	createdBot, err := c.App.CreateBot(bot)
	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(createdBot.ToJson())
}

func patchBot(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireBotUserId()
	if c.Err != nil {
		return
	}
	botUserId := c.Params.BotUserId

	botPatch := model.BotPatchFromJson(r.Body)
	if botPatch == nil {
		c.SetInvalidParam("bot")
		return
	}

	if err := c.App.SessionHasPermissionToManageBot(c.App.Session, botUserId); err != nil {
		c.Err = err
		return
	}

	updatedBot, err := c.App.PatchBot(botUserId, botPatch)
	if err != nil {
		c.Err = err
		return
	}

	w.Write(updatedBot.ToJson())
}

func getBot(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireBotUserId()
	if c.Err != nil {
		return
	}
	botUserId := c.Params.BotUserId

	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	bot, err := c.App.GetBot(botUserId, includeDeleted)
	if err != nil {
		c.Err = err
		return
	}

	if c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_READ_OTHERS_BOTS) {
		// Allow access to any bot.
	} else if bot.OwnerId == c.App.Session.UserId {
		if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_READ_BOTS) {
			// Pretend like the bot doesn't exist at all to avoid revealing that the
			// user is a bot. It's kind of silly in this case, sine we created the bot,
			// but we don't have read bot permissions.
			c.Err = model.MakeBotNotFoundError(botUserId)
			return
		}
	} else {
		// Pretend like the bot doesn't exist at all, to avoid revealing that the
		// user is a bot.
		c.Err = model.MakeBotNotFoundError(botUserId)
		return
	}

	if c.HandleEtag(bot.Etag(), "Get Bot", w, r) {
		return
	}

	w.Write(bot.ToJson())
}

func getBots(c *Context, w http.ResponseWriter, r *http.Request) {
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"
	onlyOrphaned := r.URL.Query().Get("only_orphaned") == "true"

	var OwnerId string
	if c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_READ_OTHERS_BOTS) {
		// Get bots created by any user.
		OwnerId = ""
	} else if c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_READ_BOTS) {
		// Only get bots created by this user.
		OwnerId = c.App.Session.UserId
	} else {
		c.SetPermissionError(model.PERMISSION_READ_BOTS)
		return
	}

	bots, err := c.App.GetBots(&model.BotGetOptions{
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
		OwnerId:        OwnerId,
		IncludeDeleted: includeDeleted,
		OnlyOrphaned:   onlyOrphaned,
	})
	if err != nil {
		c.Err = err
		return
	}

	if c.HandleEtag(bots.Etag(), "Get Bots", w, r) {
		return
	}

	w.Write(bots.ToJson())
}

func disableBot(c *Context, w http.ResponseWriter, r *http.Request) {
	updateBotActive(c, w, r, false)
}

func enableBot(c *Context, w http.ResponseWriter, r *http.Request) {
	updateBotActive(c, w, r, true)
}

func updateBotActive(c *Context, w http.ResponseWriter, r *http.Request, active bool) {
	c.RequireBotUserId()
	if c.Err != nil {
		return
	}
	botUserId := c.Params.BotUserId

	if err := c.App.SessionHasPermissionToManageBot(c.App.Session, botUserId); err != nil {
		c.Err = err
		return
	}

	bot, err := c.App.UpdateBotActive(botUserId, active)
	if err != nil {
		c.Err = err
		return
	}

	w.Write(bot.ToJson())
}

func assignBot(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	c.RequireBotUserId()
	if c.Err != nil {
		return
	}
	botUserId := c.Params.BotUserId
	userId := c.Params.UserId

	if err := c.App.SessionHasPermissionToManageBot(c.App.Session, botUserId); err != nil {
		c.Err = err
		return
	}

	if user, err := c.App.GetUser(userId); err == nil {
		if user.IsBot {
			c.SetPermissionError(model.PERMISSION_ASSIGN_BOT)
			return
		}
	}

	bot, err := c.App.UpdateBotOwner(botUserId, userId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write(bot.ToJson())
}
