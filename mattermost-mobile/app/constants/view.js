// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import keyMirror from 'mattermost-redux/utils/key_mirror';

export const UpgradeTypes = {
    CAN_UPGRADE: 'can_upgrade',
    MUST_UPGRADE: 'must_upgrade',
    NO_UPGRADE: 'no_upgrade',
    IS_BETA: 'is_beta',
};

export const SidebarSectionTypes = {
    UNREADS: 'unreads',
    FAVORITE: 'favorite',
    PUBLIC: 'public',
    PRIVATE: 'private',
    DIRECT: 'direct',
    RECENT_ACTIVITY: 'recent',
    ALPHA: 'alpha',
};

const ViewTypes = keyMirror({
    DATA_CLEANUP: null,
    SERVER_URL_CHANGED: null,

    LOGIN_ID_CHANGED: null,
    PASSWORD_CHANGED: null,

    POST_DRAFT_CHANGED: null,
    COMMENT_DRAFT_CHANGED: null,
    SEARCH_DRAFT_CHANGED: null,

    POST_DRAFT_SELECTION_CHANGED: null,
    COMMENT_DRAFT_SELECTION_CHANGED: null,

    NOTIFICATION_IN_APP: null,
    NOTIFICATION_TAPPED: null,

    SET_POST_DRAFT: null,
    SET_COMMENT_DRAFT: null,

    SET_TEMP_UPLOAD_FILES_FOR_POST_DRAFT: null,
    RETRY_UPLOAD_FILE_FOR_POST: null,

    CLEAR_FILES_FOR_POST_DRAFT: null,
    CLEAR_FAILED_FILES_FOR_POST_DRAFT: null,

    REMOVE_FILE_FROM_POST_DRAFT: null,
    REMOVE_LAST_FILE_FROM_POST_DRAFT: null,

    SET_CHANNEL_LOADER: null,
    SET_CHANNEL_REFRESHING: null,
    SET_CHANNEL_RETRY_FAILED: null,
    SET_CHANNEL_DISPLAY_NAME: null,

    SET_LAST_CHANNEL_FOR_TEAM: null,
    REMOVE_LAST_CHANNEL_FOR_TEAM: null,

    GITLAB: null,
    OFFICE365: null,
    SAML: null,

    SET_INITIAL_POST_VISIBILITY: null,
    INCREASE_POST_VISIBILITY: null,
    RECEIVED_FOCUSED_POST: null,
    LOADING_POSTS: null,
    SET_LOAD_MORE_POSTS_VISIBLE: null,

    SET_INITIAL_POST_COUNT: null,
    INCREASE_POST_COUNT: null,

    RECEIVED_POSTS_FOR_CHANNEL_AT_TIME: null,

    SET_LAST_UPGRADE_CHECK: null,

    ADD_RECENT_EMOJI: null,
    EXTENSION_SELECTED_TEAM_ID: null,
    ANNOUNCEMENT_BANNER: null,

    INCREMENT_EMOJI_PICKER_PAGE: null,

    LAUNCH_LOGIN: null,
    LAUNCH_CHANNEL: null,

    SET_DEEP_LINK_URL: null,

    SET_PROFILE_IMAGE_URI: null,

    SELECTED_ACTION_MENU: null,
    SUBMIT_ATTACHMENT_MENU_ACTION: null,
    SELECT_CHANNEL_WITH_MEMBER: null,
});

export default {
    ...ViewTypes,
    POST_VISIBILITY_CHUNK_SIZE: 15,
    FEATURE_TOGGLE_PREFIX: 'feature_enabled_',
    EMBED_PREVIEW: 'embed_preview',
    LINK_PREVIEW_DISPLAY: 'link_previews',
    MIN_CHANNELNAME_LENGTH: 2,
    MAX_CHANNELNAME_LENGTH: 22,
    ANDROID_TOP_LANDSCAPE: 46,
    ANDROID_TOP_PORTRAIT: 56,
    IOS_TOP_LANDSCAPE: 32,
    IOS_TOP_PORTRAIT: 64,
    IOSX_TOP_PORTRAIT: 88,
    STATUS_BAR_HEIGHT: 20,
    PROFILE_PICTURE_SIZE: 32,
    DATA_SOURCE_USERS: 'users',
    DATA_SOURCE_CHANNELS: 'channels',
};
