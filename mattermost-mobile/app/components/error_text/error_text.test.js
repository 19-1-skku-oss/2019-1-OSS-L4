// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {shallow} from 'enzyme';
import Preferences from 'mattermost-redux/constants/preferences';

import ErrorText from './error_text.js';

describe('ErrorText', () => {
    const baseProps = {
        textStyle: {
            fontSize: 14,
            marginHorizontal: 15,
        },
        theme: Preferences.THEMES.default,
        error: {
            message: 'Username must begin with a letter and contain between 3 and 22 characters including numbers, lowercase letters, and the symbols',
        },
    };

    test('should match snapshot', () => {
        const wrapper = shallow(
            <ErrorText {...baseProps}/>,
        );

        expect(wrapper.getElement()).toMatchSnapshot();
    });
});
