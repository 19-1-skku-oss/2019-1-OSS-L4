// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {UserTypes} from 'mattermost-redux/action_types';

import {ViewTypes} from 'app/constants';

export default function banner(state = '', action) {
    switch (action.type) {
    case ViewTypes.ANNOUNCEMENT_BANNER:
        return action.data;
    case UserTypes.LOGOUT_SUCCESS:
        return '';
    default:
        return state;
    }
}
