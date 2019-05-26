// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';

import {getConfig, getLicense} from 'mattermost-redux/selectors/entities/general';
import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import AnnouncementBanner from './announcement_banner';

function mapStateToProps(state) {
    const config = getConfig(state);
    const license = getLicense(state);
    const {announcement} = state.views;

    return {
        bannerColor: config.BannerColor,
        bannerDismissed: config.BannerText === announcement,
        bannerEnabled: config.EnableBanner === 'true' && license.IsLicensed === 'true',
        bannerText: config.BannerText,
        bannerTextColor: config.BannerTextColor || '#000',
        theme: getTheme(state),
    };
}

export default connect(mapStateToProps)(AnnouncementBanner);
