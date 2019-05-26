// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import PropTypes from 'prop-types';
import {
    Dimensions,
    Image,
    TouchableOpacity,
    View,
} from 'react-native';

import {makeStyleSheetFromTheme} from 'app/utils/theme';

const {width: deviceWidth, height: deviceHeight} = Dimensions.get('window');

const checkmark = require('assets/images/themes/check.png');

const tilePadding = 8;

const ThemeTile = (props) => {
    const {
        action,
        actionValue,
        imageSrc,
        isLandscape,
        isTablet,
        label,
        selected,
        theme,
    } = props;

    const style = getStyleSheet(theme);

    const labelComponent = React.cloneElement(
        label,
        {style: style.label}
    );

    const tilesPerLine = isLandscape || isTablet ? 4 : 2;
    const fullWidth = isLandscape ? deviceHeight - 40 : deviceWidth;
    const layoutStyle = {
        container: {
            width: (fullWidth / tilesPerLine) - tilePadding,
        },
        thumbnail: {
            width: (fullWidth / tilesPerLine) - (tilePadding + 20),
        },
    };

    return (
        <TouchableOpacity
            style={[style.container, layoutStyle.container]}
            onPress={() => action(actionValue)}
        >
            <View style={[style.imageWrapper, layoutStyle.thumbnail]}>
                {imageSrc && (
                    <Image
                        source={imageSrc}
                        style={[
                            style.thumbnail,
                            layoutStyle.thumbnail,
                            selected ? style.selectedThumbnail : null,
                        ]}
                    />
                )}
                {selected && (
                    <Image
                        source={checkmark}
                        style={style.check}
                    />
                )}
            </View>
            {labelComponent}
        </TouchableOpacity>
    );
};

ThemeTile.propTypes = {
    action: PropTypes.func,
    actionValue: PropTypes.string,
    imageSrc: PropTypes.oneOfType([
        PropTypes.number,
        PropTypes.object,
    ]).isRequired,
    isLandscape: PropTypes.bool.isRequired,
    isTablet: PropTypes.bool.isRequired,
    label: PropTypes.node.isRequired,
    selected: PropTypes.bool,
    theme: PropTypes.object.isRequired,
};

ThemeTile.defaultProps = {
    action: () => true,
};

const getStyleSheet = makeStyleSheetFromTheme((theme) => {
    return {
        container: {
            flexDirection: 'column',
            padding: tilePadding,
            marginTop: 8,
        },
        imageWrapper: {
            position: 'relative',
            alignItems: 'flex-start',
            marginBottom: 12,
        },
        thumbnail: {
            resizeMode: 'stretch',
        },
        selectedThumbnail: {
            opacity: 0.5,
        },
        check: {
            position: 'absolute',
            right: 10,
            bottom: 10,
        },
        label: {
            color: theme.centerChannelColor,
            fontSize: 15,
        },
    };
});

export default ThemeTile;
