// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';

import {getUser} from 'mattermost-redux/selectors/entities/users';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import AtMentionItem from './at_mention_item';

function mapStateToProps(state, ownProps) {
    const user = getUser(state, ownProps.userId);

    return {
        firstName: user.first_name,
        lastName: user.last_name,
        username: user.username,
        isBot: Boolean(user.is_bot),
        theme: getTheme(state),
    };
}

export default connect(mapStateToProps)(AtMentionItem);
