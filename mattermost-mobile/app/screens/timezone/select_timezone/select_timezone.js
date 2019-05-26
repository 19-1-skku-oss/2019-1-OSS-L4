// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {PureComponent} from 'react';
import PropTypes from 'prop-types';
import {
    View,
    FlatList,
    Platform,
} from 'react-native';
import {getTimezoneRegion} from 'mattermost-redux/utils/timezone_utils';
import {intlShape} from 'react-intl';

import SearchBar from 'app/components/search_bar';
import StatusBar from 'app/components/status_bar';
import SelectTimezoneRow from './select_timezone_row';

import {ListTypes} from 'app/constants';
import {changeOpacity, makeStyleSheetFromTheme} from 'app/utils/theme';

const ITEM_HEIGHT = 45;
const VIEWABILITY_CONFIG = ListTypes.VISIBILITY_CONFIG_DEFAULTS;

export default class Timezone extends PureComponent {
    static propTypes = {
        selectedTimezone: PropTypes.string.isRequired,
        initialScrollIndex: PropTypes.number.isRequired,
        timezones: PropTypes.array.isRequired,
        navigator: PropTypes.object,
        onBack: PropTypes.func.isRequired,
        theme: PropTypes.object.isRequired,
    };

    static contextTypes = {
        intl: intlShape,
    };

    constructor(props) {
        super(props);

        this.state = {
            value: '',
            timezones: props.timezones,
        };
    }

    filteredTimezones = (timezonePrefix) => {
        if (timezonePrefix.length === 0) {
            return this.state.timezones;
        }

        const lowerCasePrefix = timezonePrefix.toLowerCase();

        return this.state.timezones.filter((t) => (
            getTimezoneRegion(t).toLowerCase().indexOf(lowerCasePrefix) >= 0 ||
            t.toLowerCase().indexOf(lowerCasePrefix) >= 0
        ));
    };

    timezoneSelected = (timezone) => {
        this.props.onBack(timezone);
        this.props.navigator.pop();
    };

    handleTextChanged = (value) => {
        this.setState({value});
    };

    keyExtractor = (item) => item;

    getItemLayout = (data, index) => ({
        length: ITEM_HEIGHT,
        offset: ITEM_HEIGHT * index,
        index,
    });

    renderItem = ({item: timezone}) => {
        return (
            <SelectTimezoneRow
                theme={this.props.theme}
                timezone={timezone}
                selectedTimezone={this.props.selectedTimezone}
                onPress={this.timezoneSelected}
            />
        );
    };

    render() {
        const {theme, initialScrollIndex} = this.props;
        const {value} = this.state;
        const {intl} = this.context;
        const style = getStyleSheet(theme);

        const searchBarInput = {
            backgroundColor: changeOpacity(theme.sidebarHeaderTextColor, 0.2),
            color: theme.sidebarHeaderTextColor,
            fontSize: 15,
        };

        return (
            <View style={style.container}>
                <StatusBar/>
                <View style={style.header}>
                    <SearchBar
                        ref='searchBar'
                        placeholder={intl.formatMessage({id: 'search_bar.search', defaultMessage: 'Search'})}
                        cancelTitle={intl.formatMessage({id: 'mobile.post.cancel', defaultMessage: 'Cancel'})}
                        backgroundColor='transparent'
                        inputHeight={Platform.OS === 'ios' ? 33 : 46}
                        inputStyle={searchBarInput}
                        placeholderTextColor={changeOpacity(theme.sidebarHeaderTextColor, 0.5)}
                        selectionColor={changeOpacity(theme.sidebarHeaderTextColor, 0.5)}
                        tintColorSearch={changeOpacity(theme.sidebarHeaderTextColor, 0.5)}
                        tintColorDelete={changeOpacity(theme.sidebarHeaderTextColor, 0.5)}
                        titleCancelColor={theme.sidebarHeaderTextColor}
                        onChangeText={this.handleTextChanged}
                        autoCapitalize='none'
                        value={value}
                        containerStyle={style.searchBarContainer}
                        showArrow={false}
                    />
                </View>
                <FlatList
                    data={this.filteredTimezones(value)}
                    renderItem={this.renderItem}
                    keyExtractor={this.keyExtractor}
                    getItemLayout={this.getItemLayout}
                    keyboardShouldPersistTaps='always'
                    keyboardDismissMode='on-drag'
                    maxToRenderPerBatch={15}
                    initialScrollIndex={initialScrollIndex}
                    viewabilityConfig={VIEWABILITY_CONFIG}
                />
            </View>
        );
    }
}

const getStyleSheet = makeStyleSheetFromTheme((theme) => {
    return {
        container: {
            flex: 1,
            backgroundColor: theme.centerChannelBg,
        },
        header: {
            backgroundColor: theme.sidebarHeaderBg,
            width: '100%',
            ...Platform.select({
                android: {
                    height: 46,
                    justifyContent: 'center',
                },
                ios: {
                    height: 44,
                },
            }),
        },
    };
});
