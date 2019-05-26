// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {PureComponent} from 'react';
import PropTypes from 'prop-types';
import {
    Keyboard,
    Platform,
    StyleSheet,
    View,
} from 'react-native';

import * as CustomPropTypes from 'app/constants/custom_prop_types';

export default class KeyboardLayout extends PureComponent {
    static propTypes = {
        children: PropTypes.node,
        style: CustomPropTypes.Style,
    };

    constructor(props) {
        super(props);
        this.subscriptions = [];
        this.state = {
            keyboardHeight: 0,
        };
    }

    componentDidMount() {
        if (Platform.OS === 'ios') {
            this.subscriptions = [
                Keyboard.addListener('keyboardWillShow', this.onKeyboardWillShow),
                Keyboard.addListener('keyboardWillHide', this.onKeyboardWillHide),
            ];
        }
    }

    componentWillUnmount() {
        this.subscriptions.forEach((sub) => sub.remove());
    }

    onKeyboardWillHide = () => {
        this.setState({
            keyboardHeight: 0,
        });
    };

    onKeyboardWillShow = (e) => {
        this.setState({
            keyboardHeight: e?.endCoordinates?.height || 0,
        });
    };

    render() {
        const layoutStyle = [this.props.style, style.keyboardLayout];

        if (Platform.OS === 'ios') {
            // iOS doesn't resize the app automatically
            layoutStyle.push({paddingBottom: this.state.keyboardHeight});
        }

        return (
            <View style={layoutStyle}>
                {this.props.children}
            </View>
        );
    }
}

const style = StyleSheet.create({
    keyboardLayout: {
        position: 'relative',
        flex: 1,
    },
});
