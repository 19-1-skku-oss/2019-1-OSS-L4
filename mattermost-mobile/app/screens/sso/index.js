// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {bindActionCreators} from 'redux';
import {connect} from 'react-redux';

import {handleSuccessfulLogin, scheduleExpiredNotification} from 'app/actions/views/login';
import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import {setStoreFromLocalData} from 'mattermost-redux/actions/general';

import SSO from './sso';

function mapStateToProps(state) {
    return {
        ...state.views.selectServer,
        theme: getTheme(state),
    };
}

function mapDispatchToProps(dispatch) {
    return {
        actions: bindActionCreators({
            scheduleExpiredNotification,
            handleSuccessfulLogin,
            setStoreFromLocalData,
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(SSO);
