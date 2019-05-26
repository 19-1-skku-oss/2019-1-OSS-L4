// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {bindActionCreators} from 'redux';
import {connect} from 'react-redux';

import {selectFocusedPostId, selectPost} from 'mattermost-redux/actions/posts';
import {clearSearch, getFlaggedPosts} from 'mattermost-redux/actions/search';
import {RequestStatus} from 'mattermost-redux/constants';
import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import {loadChannelsByTeamName, loadThreadIfNecessary} from 'app/actions/views/channel';
import {showSearchModal} from 'app/actions/views/search';
import {makePreparePostIdsForSearchPosts} from 'app/selectors/post_list';

import FlaggedPosts from './flagged_posts';

function makeMapStateToProps() {
    const preparePostIds = makePreparePostIdsForSearchPosts();
    return (state) => {
        const postIds = preparePostIds(state, state.entities.search.flagged);
        const {flaggedPosts: flaggedPostsRequest} = state.requests.search;
        const isLoading = flaggedPostsRequest.status === RequestStatus.STARTED ||
            flaggedPostsRequest.status === RequestStatus.NOT_STARTED;

        return {
            postIds,
            isLoading,
            didFail: flaggedPostsRequest.status === RequestStatus.FAILURE,
            theme: getTheme(state),
        };
    };
}

function mapDispatchToProps(dispatch) {
    return {
        actions: bindActionCreators({
            clearSearch,
            loadChannelsByTeamName,
            loadThreadIfNecessary,
            getFlaggedPosts,
            selectFocusedPostId,
            selectPost,
            showSearchModal,
        }, dispatch),
    };
}

export default connect(makeMapStateToProps, mapDispatchToProps)(FlaggedPosts);
