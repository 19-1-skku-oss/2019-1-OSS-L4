// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/model"
)

func TestGetMigrationState(t *testing.T) {
	th := Setup()
	defer th.TearDown()

	migrationKey := model.NewId()

	th.DeleteAllJobsByTypeAndMigrationKey(model.JOB_TYPE_MIGRATIONS, migrationKey)

	// Test with no job yet.
	state, job, err := GetMigrationState(migrationKey, th.App.Srv.Store)
	assert.Nil(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "unscheduled", state)

	// Test with the system table showing the migration as done.
	system := model.System{
		Name:  migrationKey,
		Value: "true",
	}
	err = th.App.Srv.Store.System().Save(&system)
	assert.Nil(t, err)

	state, job, err = GetMigrationState(migrationKey, th.App.Srv.Store)
	assert.Nil(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "completed", state)

	_, err = th.App.Srv.Store.System().PermanentDeleteByName(migrationKey)
	assert.Nil(t, err)

	// Test with a job scheduled in "pending" state.
	j1 := &model.Job{
		Id:       model.NewId(),
		CreateAt: model.GetMillis(),
		Data: map[string]string{
			JOB_DATA_KEY_MIGRATION: migrationKey,
		},
		Status: model.JOB_STATUS_PENDING,
		Type:   model.JOB_TYPE_MIGRATIONS,
	}

	j1 = (<-th.App.Srv.Store.Job().Save(j1)).Data.(*model.Job)

	state, job, err = GetMigrationState(migrationKey, th.App.Srv.Store)
	assert.Nil(t, err)
	assert.Equal(t, j1.Id, job.Id)
	assert.Equal(t, "in_progress", state)

	// Test with a job scheduled in "in progress" state.
	j2 := &model.Job{
		Id:       model.NewId(),
		CreateAt: j1.CreateAt + 1,
		Data: map[string]string{
			JOB_DATA_KEY_MIGRATION: migrationKey,
		},
		Status: model.JOB_STATUS_IN_PROGRESS,
		Type:   model.JOB_TYPE_MIGRATIONS,
	}

	j2 = (<-th.App.Srv.Store.Job().Save(j2)).Data.(*model.Job)

	state, job, err = GetMigrationState(migrationKey, th.App.Srv.Store)
	assert.Nil(t, err)
	assert.Equal(t, j2.Id, job.Id)
	assert.Equal(t, "in_progress", state)

	// Test with a job scheduled in "error" state.
	j3 := &model.Job{
		Id:       model.NewId(),
		CreateAt: j2.CreateAt + 1,
		Data: map[string]string{
			JOB_DATA_KEY_MIGRATION: migrationKey,
		},
		Status: model.JOB_STATUS_ERROR,
		Type:   model.JOB_TYPE_MIGRATIONS,
	}

	j3 = (<-th.App.Srv.Store.Job().Save(j3)).Data.(*model.Job)

	state, job, err = GetMigrationState(migrationKey, th.App.Srv.Store)
	assert.Nil(t, err)
	assert.Equal(t, j3.Id, job.Id)
	assert.Equal(t, "unscheduled", state)
}
