// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {bindActionCreators} from 'redux';
import {connect} from 'react-redux';
import {createSelector} from 'reselect';

import {searchChannels} from 'mattermost-redux/actions/channels';
import {getProfilesInTeam, searchProfiles} from 'mattermost-redux/actions/users';
import {makeGroupMessageVisibleIfNecessary} from 'mattermost-redux/actions/preferences';
import {General} from 'mattermost-redux/constants';
import {
    getChannelsWithUnreadSection,
    getCurrentChannel,
    getGroupChannels,
    getOtherChannels,
} from 'mattermost-redux/selectors/entities/channels';
import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {getCurrentTeam} from 'mattermost-redux/selectors/entities/teams';
import {getCurrentUserId, getProfilesInCurrentTeam, getUsers, getUserIdsInChannels, getUserStatuses} from 'mattermost-redux/selectors/entities/users';
import {getDirectShowPreferences, getTeammateNameDisplaySetting, getTheme} from 'mattermost-redux/selectors/entities/preferences';

import Config from 'assets/config';

import FilteredList from './filtered_list';

const DEFAULT_SEARCH_ORDER = ['unreads', 'dms', 'channels', 'members', 'nonmembers'];

const pastDirectMessages = createSelector(
    getDirectShowPreferences,
    (directChannelsFromPreferences) => directChannelsFromPreferences.filter((d) => d.value === 'false').map((d) => d.name)
);

const getTeamProfiles = createSelector(
    getProfilesInCurrentTeam,
    (members) => {
        return members.reduce((memberProfiles, member) => {
            memberProfiles[member.id] = member;

            return memberProfiles;
        }, {});
    }
);

// Fill an object for each group channel with concatenated strings for username, email, fullname, and nickname
function getGroupDetails(currentUserId, userIdsInChannels, profiles, groupChannels) {
    return groupChannels.reduce((groupMemberDetails, channel) => {
        if (!userIdsInChannels.hasOwnProperty(channel.id)) {
            return groupMemberDetails;
        }

        const members = Array.from(userIdsInChannels[channel.id]).reduce((memberDetails, member) => {
            if (member === currentUserId) {
                return memberDetails;
            }

            const details = {...memberDetails};

            const profile = profiles[member];
            details.username.push(profile.username);
            if (profile.email) {
                details.email.push(profile.email);
            }
            if (profile.nickname) {
                details.nickname.push(profile.nickname);
            }
            if (profile.fullname) {
                details.fullname.push(`${profile.first_name} ${profile.last_name}`);
            }

            return details;
        }, {
            email: [],
            fullname: [],
            nickname: [],
            username: [],
        });

        groupMemberDetails[channel.id] = {
            email: members.email.join(','),
            fullname: members.fullname.join(','),
            nickname: members.nickname.join(','),
            username: members.username.join(','),
        };

        return groupMemberDetails;
    }, {});
}

const getGroupChannelMemberDetails = createSelector(
    getCurrentUserId,
    getUserIdsInChannels,
    getUsers,
    getGroupChannels,
    getGroupDetails
);

function mapStateToProps(state) {
    const {currentUserId} = state.entities.users;

    const profiles = getUsers(state);
    let teamProfiles = {};
    const restrictDms = getConfig(state).RestrictDirectMessage !== General.RESTRICT_DIRECT_MESSAGE_ANY;
    if (restrictDms) {
        teamProfiles = getTeamProfiles(state);
    }

    const searchOrder = Config.DrawerSearchOrder ? Config.DrawerSearchOrder : DEFAULT_SEARCH_ORDER;

    return {
        channels: getChannelsWithUnreadSection(state),
        currentChannel: getCurrentChannel(state),
        currentTeam: getCurrentTeam(state),
        currentUserId,
        otherChannels: getOtherChannels(state),
        groupChannelMemberDetails: getGroupChannelMemberDetails(state),
        profiles,
        teamProfiles,
        teammateNameDisplay: getTeammateNameDisplaySetting(state),
        statuses: getUserStatuses(state),
        searchOrder,
        pastDirectMessages: pastDirectMessages(state),
        restrictDms,
        theme: getTheme(state),
    };
}

function mapDispatchToProps(dispatch) {
    return {
        actions: bindActionCreators({
            getProfilesInTeam,
            makeGroupMessageVisibleIfNecessary,
            searchChannels,
            searchProfiles,
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(FilteredList);
