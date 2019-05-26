// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type MyPlugin struct {
	plugin.MattermostPlugin
}

func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	createdBot, err := p.API.CreateBot(&model.Bot{
		Username:    "bot",
		Description: "a plugin bot",
	})

	if err != nil {
		return nil, err.Error() + "failed to create bot"
	}

	fetchedBot, err := p.API.GetBot(createdBot.UserId, false)
	if err != nil {
		return nil, err.Error() + "failed to get bot"
	}
	if fetchedBot.Description != "a plugin bot" {
		return nil, "GetBot did not return the expected bot Description"
	}
	if fetchedBot.OwnerId != "test_bots_plugin" {
		return nil, "GetBot did not return the expected bot OwnerId"
	}

	updatedDescription := createdBot.Description + ", updated"
	patchedBot, err := p.API.PatchBot(createdBot.UserId, &model.BotPatch{
		Description: &updatedDescription,
	})
	if err != nil {
		return nil, err.Error() + "failed to patch bot"
	}

	fetchedBot, err = p.API.GetBot(patchedBot.UserId, false)
	if err != nil {
		return nil, err.Error() + "failed to get bot"
	}

	if fetchedBot.UserId != patchedBot.UserId {
		return nil, "GetBot did not return the expected bot"
	}
	if fetchedBot.Description != "a plugin bot, updated" {
		return nil, "GetBot did not return the updated bot Description"
	}

	fetchedBots, err := p.API.GetBots(&model.BotGetOptions{
		Page:           0,
		PerPage:        1,
		OwnerId:        "",
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, err.Error() + "failed to get bots"
	}

	if len(fetchedBots) != 1 {
		return nil, "GetBots did not return a single bot"
	}

	if fetchedBot.UserId != fetchedBots[0].UserId {
		return nil, "GetBots did not return the expected bot"
	}
	if _, err = p.API.UpdateBotActive(fetchedBot.UserId, false); err != nil {
		return nil, err.Error() + "failed to disable bot"
	}
	if fetchedBot, err = p.API.GetBot(patchedBot.UserId, false); err == nil {
		return nil, "expected not to find disabled bot"
	}
	if _, err = p.API.UpdateBotActive(fetchedBot.UserId, true); err != nil {
		return nil, err.Error() + "failed to disable bot"
	}
	if fetchedBot, err = p.API.GetBot(patchedBot.UserId, false); err != nil {
		return nil, err.Error() + "failed to get bot after enabling"
	}
	if fetchedBot.UserId != patchedBot.UserId {
		return nil, "GetBot did not return the expected bot after enabling"
	}
	if err = p.API.PermanentDeleteBot(patchedBot.UserId); err != nil {
		return nil, err.Error() + "failed to delete bot"
	}

	if _, err = p.API.GetBot(patchedBot.UserId, false); err == nil {
		return nil, err.Error() + "found bot after permanently deleting"
	}
	createdBotWithOverriddenCreator, err := p.API.CreateBot(&model.Bot{
		Username:    "bot",
		Description: "a plugin bot",
		OwnerId:     "abc123",
	})
	if err != nil {
		return nil, err.Error() + "failed to create bot with overridden creator"
	}
	if fetchedBot, err = p.API.GetBot(createdBotWithOverriddenCreator.UserId, false); err != nil {
		return nil, err.Error() + "failed to get bot"
	}
	if fetchedBot.Description != "a plugin bot" {
		return nil, "GetBot did not return the expected bot Description"
	}
	if fetchedBot.OwnerId != "abc123" {
		return nil, "GetBot did not return the expected bot OwnerId"
	}
	return nil, ""
}

func main() {
	plugin.ClientMain(&MyPlugin{})
}
