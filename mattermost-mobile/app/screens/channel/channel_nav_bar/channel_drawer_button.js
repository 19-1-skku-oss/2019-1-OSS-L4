// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {PureComponent} from 'react';
import PropTypes from 'prop-types';
import {connect} from 'react-redux';
import {
    TouchableOpacity,
    View,
} from 'react-native';

import Icon from 'react-native-vector-icons/Ionicons';

import Badge from 'app/components/badge';
import {getTheme} from 'mattermost-redux/selectors/entities/preferences';
import {preventDoubleTap} from 'app/utils/tap';
import {makeStyleSheetFromTheme} from 'app/utils/theme';

import telemetry from 'app/telemetry';

import {getUnreadsInCurrentTeam} from 'mattermost-redux/selectors/entities/channels';
import {getCurrentTeamId, getTeamMemberships} from 'mattermost-redux/selectors/entities/teams';

class ChannelDrawerButton extends PureComponent {
    static propTypes = {
        currentTeamId: PropTypes.string.isRequired,
        openDrawer: PropTypes.func.isRequired,
        messageCount: PropTypes.number,
        mentionCount: PropTypes.number,
        myTeamMembers: PropTypes.object,
        theme: PropTypes.object,
        visible: PropTypes.bool,
    };

    static defaultProps = {
        currentChannel: {},
        theme: {},
        messageCount: 0,
        mentionCount: 0,
        visible: true,
    };

    handlePress = preventDoubleTap(() => {
        telemetry.start(['channel:open_drawer']);
        this.props.openDrawer();
    });

    render() {
        const {
            currentTeamId,
            mentionCount,
            messageCount,
            myTeamMembers,
            theme,
            visible,
        } = this.props;

        const style = getStyleFromTheme(theme);

        let mentions = mentionCount;
        let messages = messageCount;

        const members = Object.values(myTeamMembers).filter((m) => m.team_id !== currentTeamId);
        members.forEach((m) => {
            mentions += (m.mention_count || 0);
            messages += (m.msg_count || 0);
        });

        let badgeCount = 0;
        if (mentions) {
            badgeCount = mentions;
        } else if (messages) {
            badgeCount = -1;
        }

        let badge;
        if (badgeCount && visible) {
            badge = (
                <Badge
                    style={style.badge}
                    countStyle={style.mention}
                    count={badgeCount}
                    onPress={this.handlePress}
                />
            );
        }

        let icon;
        let containerStyle;
        if (visible) {
            icon = (
                <Icon
                    name='md-menu'
                    size={25}
                    color={theme.sidebarHeaderTextColor}
                />
            );
            containerStyle = style.container;
        } else {
            icon = (<View style={style.tabletIcon}/>);
            containerStyle = style.tabletContainer;
        }

        return (
            <TouchableOpacity
                onPress={this.handlePress}
                style={containerStyle}
            >
                <View style={[style.wrapper]}>
                    <View>
                        {icon}
                        {badge}
                    </View>
                </View>
            </TouchableOpacity>
        );
    }
}

const getStyleFromTheme = makeStyleSheetFromTheme((theme) => {
    return {
        container: {
            width: 55,
        },
        tabletContainer: {
            width: 10,
        },
        tabletIcon: {
            height: 25,
        },
        wrapper: {
            alignItems: 'center',
            flex: 1,
            flexDirection: 'column',
            justifyContent: 'center',
            paddingHorizontal: 10,
        },
        badge: {
            backgroundColor: theme.mentionBg,
            borderColor: theme.sidebarHeaderBg,
            borderRadius: 10,
            borderWidth: 1,
            left: -13,
            padding: 3,
            position: 'absolute',
            right: 0,
            top: -4,
        },
        mention: {
            color: theme.mentionColor,
            fontSize: 10,
        },
    };
});

function mapStateToProps(state) {
    return {
        currentTeamId: getCurrentTeamId(state),
        myTeamMembers: getTeamMemberships(state),
        theme: getTheme(state),
        ...getUnreadsInCurrentTeam(state),
    };
}

export default connect(mapStateToProps)(ChannelDrawerButton);
