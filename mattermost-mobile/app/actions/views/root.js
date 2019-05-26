// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {GeneralTypes} from 'mattermost-redux/action_types';
import {Client4} from 'mattermost-redux/client';
import {General} from 'mattermost-redux/constants';
import {fetchMyChannelsAndMembers} from 'mattermost-redux/actions/channels';
import {getClientConfig, getDataRetentionPolicy, getLicenseConfig} from 'mattermost-redux/actions/general';
import {receivedNewPost} from 'mattermost-redux/actions/posts';
import {getMyTeams, getMyTeamMembers, selectTeam} from 'mattermost-redux/actions/teams';

import {ViewTypes} from 'app/constants';
import {recordTime} from 'app/utils/segment';

import {handleSelectChannel} from 'app/actions/views/channel';

export function startDataCleanup() {
    return async (dispatch, getState) => {
        dispatch({
            type: ViewTypes.DATA_CLEANUP,
            payload: getState(),
        });
    };
}

export function loadConfigAndLicense() {
    return async (dispatch, getState) => {
        const {currentUserId} = getState().entities.users;
        const [configData, licenseData] = await Promise.all([
            getClientConfig()(dispatch, getState),
            getLicenseConfig()(dispatch, getState),
        ]);

        const config = configData.data || {};
        const license = licenseData.data || {};

        if (currentUserId) {
            if (config.DataRetentionEnableMessageDeletion && config.DataRetentionEnableMessageDeletion === 'true' &&
                license.IsLicensed === 'true' && license.DataRetention === 'true') {
                getDataRetentionPolicy()(dispatch, getState);
            } else {
                dispatch({type: GeneralTypes.RECEIVED_DATA_RETENTION_POLICY, data: {}});
            }
        }

        return {config, license};
    };
}

export function loadFromPushNotification(notification, startAppFromPushNotification) {
    return async (dispatch, getState) => {
        const state = getState();
        const {data} = notification;
        const {currentTeamId, teams, myMembers: myTeamMembers} = state.entities.teams;
        const {channels} = state.entities.channels;

        let channelId = '';
        let teamId = currentTeamId;
        if (data) {
            channelId = data.channel_id;

            // when the notification does not have a team id is because its from a DM or GM
            teamId = data.team_id || currentTeamId;
        }

        // load any missing data
        const loading = [];

        if (teamId && (!teams[teamId] || !myTeamMembers[teamId])) {
            loading.push(dispatch(getMyTeams()));
            loading.push(dispatch(getMyTeamMembers()));
        }

        if (channelId && !channels[channelId]) {
            loading.push(dispatch(fetchMyChannelsAndMembers(teamId)));
        }

        if (loading.length > 0) {
            await Promise.all(loading);
        }

        // when the notification is from a team other than the current team
        if (teamId !== currentTeamId) {
            dispatch(selectTeam({id: teamId}));
        }

        dispatch(handleSelectChannel(channelId, startAppFromPushNotification));
    };
}

export function purgeOfflineStore() {
    return {type: General.OFFLINE_STORE_PURGE};
}

// A non-optimistic version of the createPost action in mattermost-redux with the file handling
// removed since it's not needed.
export function createPostForNotificationReply(post) {
    return async (dispatch, getState) => {
        const state = getState();
        const currentUserId = state.entities.users.currentUserId;

        const timestamp = Date.now();
        const pendingPostId = post.pending_post_id || `${currentUserId}:${timestamp}`;

        const newPost = {
            ...post,
            pending_post_id: pendingPostId,
            create_at: timestamp,
            update_at: timestamp,
        };

        try {
            const data = await Client4.createPost({...newPost, create_at: 0});
            dispatch(receivedNewPost(data));

            return {data};
        } catch (error) {
            return {error};
        }
    };
}

export function recordLoadTime(screenName, category) {
    return async (dispatch, getState) => {
        const {currentUserId} = getState().entities.users;

        recordTime(screenName, category, currentUserId);
    };
}

export function setDeepLinkURL(url) {
    return {
        type: ViewTypes.SET_DEEP_LINK_URL,
        url,
    };
}

export default {
    loadConfigAndLicense,
    loadFromPushNotification,
    purgeOfflineStore,
};
