// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';
import {getConfig, getLicense} from 'mattermost-redux/selectors/entities/general';

import LoginOptions from './login_options';

function mapStateToProps(state) {
    const config = getConfig(state);
    const license = getLicense(state);
    return {
        config,
        license,
        theme: getTheme(state),
    };
}

export default connect(mapStateToProps)(LoginOptions);
