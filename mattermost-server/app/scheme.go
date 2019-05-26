// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost-server/model"
)

func (a *App) GetScheme(id string) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	result := <-a.Srv.Store.Scheme().Get(id)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Scheme), nil
}

func (a *App) GetSchemeByName(name string) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	result := <-a.Srv.Store.Scheme().GetByName(name)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Scheme), nil
}

func (a *App) GetSchemesPage(scope string, page int, perPage int) ([]*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	return a.GetSchemes(scope, page*perPage, perPage)
}

func (a *App) GetSchemes(scope string, offset int, limit int) ([]*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	result := <-a.Srv.Store.Scheme().GetAllPage(scope, offset, limit)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Scheme), nil
}

func (a *App) CreateScheme(scheme *model.Scheme) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	// Clear any user-provided values for trusted properties.
	scheme.DefaultTeamAdminRole = ""
	scheme.DefaultTeamUserRole = ""
	scheme.DefaultTeamGuestRole = ""
	scheme.DefaultChannelAdminRole = ""
	scheme.DefaultChannelUserRole = ""
	scheme.DefaultChannelGuestRole = ""
	scheme.CreateAt = 0
	scheme.UpdateAt = 0
	scheme.DeleteAt = 0

	result := <-a.Srv.Store.Scheme().Save(scheme)
	if result.Err != nil {
		return nil, result.Err
	}
	return scheme, nil
}

func (a *App) PatchScheme(scheme *model.Scheme, patch *model.SchemePatch) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	scheme.Patch(patch)
	scheme, err := a.UpdateScheme(scheme)
	if err != nil {
		return nil, err
	}

	return scheme, err
}

func (a *App) UpdateScheme(scheme *model.Scheme) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	result := <-a.Srv.Store.Scheme().Save(scheme)
	if result.Err != nil {
		return nil, result.Err
	}
	return scheme, nil
}

func (a *App) DeleteScheme(schemeId string) (*model.Scheme, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	result := <-a.Srv.Store.Scheme().Delete(schemeId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Scheme), nil
}

func (a *App) GetTeamsForSchemePage(scheme *model.Scheme, page int, perPage int) ([]*model.Team, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	return a.GetTeamsForScheme(scheme, page*perPage, perPage)
}

func (a *App) GetTeamsForScheme(scheme *model.Scheme, offset int, limit int) ([]*model.Team, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	result := <-a.Srv.Store.Team().GetTeamsByScheme(scheme.Id, offset, limit)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Team), nil
}

func (a *App) GetChannelsForSchemePage(scheme *model.Scheme, page int, perPage int) (model.ChannelList, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}

	return a.GetChannelsForScheme(scheme, page*perPage, perPage)
}

func (a *App) GetChannelsForScheme(scheme *model.Scheme, offset int, limit int) (model.ChannelList, *model.AppError) {
	if err := a.IsPhase2MigrationCompleted(); err != nil {
		return nil, err
	}
	result := <-a.Srv.Store.Channel().GetChannelsByScheme(scheme.Id, offset, limit)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(model.ChannelList), nil
}

func (a *App) IsPhase2MigrationCompleted() *model.AppError {
	if a.Srv.phase2PermissionsMigrationComplete {
		return nil
	}

	if _, err := a.Srv.Store.System().GetByName(model.MIGRATION_KEY_ADVANCED_PERMISSIONS_PHASE_2); err != nil {
		return model.NewAppError("App.IsPhase2MigrationCompleted", "app.schemes.is_phase_2_migration_completed.not_completed.app_error", nil, err.Error(), http.StatusNotImplemented)
	}

	a.Srv.phase2PermissionsMigrationComplete = true

	return nil
}

func (a *App) SchemesIterator(batchSize int) func() []*model.Scheme {
	offset := 0
	return func() []*model.Scheme {
		result := <-a.Srv.Store.Scheme().GetAllPage("", offset, batchSize)
		if result.Err != nil {
			return []*model.Scheme{}
		}
		offset += batchSize
		return result.Data.([]*model.Scheme)
	}
}
