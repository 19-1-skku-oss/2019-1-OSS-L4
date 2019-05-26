// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserTermsOfService(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	userTermsOfService, err := th.App.GetUserTermsOfService(th.BasicUser.Id)
	checkError(t, err)
	assert.Nil(t, userTermsOfService)
	assert.Equal(t, "store.sql_user_terms_of_service.get_by_user.no_rows.app_error", err.Id)

	termsOfService, err := th.App.CreateTermsOfService("terms of service", th.BasicUser.Id)
	checkNoError(t, err)

	err = th.App.SaveUserTermsOfService(th.BasicUser.Id, termsOfService.Id, true)
	checkNoError(t, err)

	userTermsOfService, err = th.App.GetUserTermsOfService(th.BasicUser.Id)
	checkNoError(t, err)
	assert.NotNil(t, userTermsOfService)
	assert.NotEmpty(t, userTermsOfService)

	assert.Equal(t, th.BasicUser.Id, userTermsOfService.UserId)
	assert.Equal(t, termsOfService.Id, userTermsOfService.TermsOfServiceId)
	assert.NotEmpty(t, userTermsOfService.CreateAt)
}
