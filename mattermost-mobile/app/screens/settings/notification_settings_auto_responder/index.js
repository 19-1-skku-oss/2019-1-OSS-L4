// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentUserId, getStatusForUserId} from 'mattermost-redux/selectors/entities/users';

import NotificationSettingsAutoResponder from './notification_settings_auto_responder';

function mapStateToProps(state) {
    const currentUserId = getCurrentUserId(state);
    const currentUserStatus = getStatusForUserId(state, currentUserId);

    return {
        theme: getTheme(state),
        currentUserStatus,
    };
}

export default connect(mapStateToProps)(NotificationSettingsAutoResponder);
