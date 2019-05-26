// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

/* eslint-disable global-require*/
import {Linking, NativeModules, Platform, Text} from 'react-native';
import AsyncStorage from '@react-native-community/async-storage';
import {setGenericPassword, getGenericPassword, resetGenericPassword} from 'react-native-keychain';

import {loadMe} from 'mattermost-redux/actions/users';
import {Client4} from 'mattermost-redux/client';
import EventEmitter from 'mattermost-redux/utils/event_emitter';

import {setDeepLinkURL} from 'app/actions/views/root';
import {ViewTypes} from 'app/constants';
import tracker from 'app/utils/time_tracker';
import {getCurrentLocale} from 'app/selectors/i18n';

import {getTranslations as getLocalTranslations} from 'app/i18n';
import {store, handleManagedConfig} from 'app/mattermost';
import avoidNativeBridge from 'app/utils/avoid_native_bridge';
import {setCSRFFromCookie} from 'app/utils/security';

const {Initialization} = NativeModules;

const TOOLBAR_BACKGROUND = 'TOOLBAR_BACKGROUND';
const TOOLBAR_TEXT_COLOR = 'TOOLBAR_TEXT_COLOR';
const APP_BACKGROUND = 'APP_BACKGROUND';

export default class App {
    constructor() {
        // Usage: app.js
        this.shouldRelaunchWhenActive = false;
        this.inBackgroundSince = null;

        // Usage: screen/entry.js
        this.startAppFromPushNotification = false;
        this.isNotificationsConfigured = false;
        this.allowOtherServers = true;
        this.appStarted = false;
        this.emmEnabled = false;
        this.performingEMMAuthentication = false;
        this.translations = null;
        this.toolbarBackground = null;
        this.toolbarTextColor = null;
        this.appBackground = null;

        // Usage utils/push_notifications.js
        this.replyNotificationData = null;
        this.deviceToken = null;

        // Usage credentials
        this.currentUserId = null;
        this.token = null;
        this.url = null;

        // Usage deeplinking
        Linking.addEventListener('url', this.handleDeepLink);

        this.setFontFamily();
        this.getStartupThemes();
        this.getAppCredentials();
    }

    setFontFamily = () => {
        // Set a global font for Android
        if (Platform.OS === 'android') {
            const defaultFontFamily = {
                style: {
                    fontFamily: 'Roboto',
                },
            };
            const TextRender = Text.render;
            const initialDefaultProps = Text.defaultProps;
            Text.defaultProps = {
                ...initialDefaultProps,
                ...defaultFontFamily,
            };
            Text.render = function render(props, ...args) {
                const oldProps = props;
                let newProps = {...props, style: [defaultFontFamily.style, props.style]};
                try {
                    return Reflect.apply(TextRender, this, [newProps, ...args]);
                } finally {
                    newProps = oldProps;
                }
            };
        }
    };

    getTranslations = () => {
        if (this.translations) {
            return this.translations;
        }

        const state = store.getState();
        const locale = getCurrentLocale(state);

        this.translations = getLocalTranslations(locale);
        return this.translations;
    };

    getAppCredentials = async () => {
        try {
            const credentials = await avoidNativeBridge(
                () => {
                    return Initialization.credentialsExist;
                },
                () => {
                    return Initialization.credentials;
                },
                () => {
                    this.waitForRehydration = true;
                    return getGenericPassword();
                }
            );

            if (credentials) {
                const usernameParsed = credentials.username.split(',');
                const passwordParsed = credentials.password.split(',');

                // username == deviceToken, currentUserId
                // password == token, url
                if (usernameParsed.length === 2 && passwordParsed.length === 2) {
                    const [deviceToken, currentUserId] = usernameParsed;
                    const [token, url] = passwordParsed;

                    // if for any case the url and the token aren't valid proceed with re-hydration
                    if (url && url !== 'undefined' && token && token !== 'undefined') {
                        this.deviceToken = deviceToken;
                        this.currentUserId = currentUserId;
                        this.token = token;
                        this.url = url;
                        Client4.setUrl(url);
                        Client4.setToken(token);
                        await setCSRFFromCookie(url);
                    } else {
                        this.waitForRehydration = true;
                    }
                }
            } else {
                this.waitForRehydration = false;
            }
        } catch (error) {
            return null;
        }

        return null;
    };

    getStartupThemes = async () => {
        try {
            const [
                toolbarBackground,
                toolbarTextColor,
                appBackground,
            ] = await avoidNativeBridge(
                () => {
                    return Initialization.themesExist;
                },
                () => {
                    return [
                        Initialization.toolbarBackground,
                        Initialization.toolbarTextColor,
                        Initialization.appBackground,
                    ];
                },
                () => {
                    return Promise.all([
                        AsyncStorage.getItem(TOOLBAR_BACKGROUND),
                        AsyncStorage.getItem(TOOLBAR_TEXT_COLOR),
                        AsyncStorage.getItem(APP_BACKGROUND),
                    ]);
                }
            );

            if (toolbarBackground) {
                this.toolbarBackground = toolbarBackground;
                this.toolbarTextColor = toolbarTextColor;
                this.appBackground = appBackground;
            }
        } catch (error) {
            return null;
        }

        return null;
    };

    setPerformingEMMAuthentication = (authenticating) => {
        this.performingEMMAuthentication = authenticating;
    };

    setAppCredentials = (deviceToken, currentUserId, token, url) => {
        if (!currentUserId) {
            return;
        }

        const username = `${deviceToken}, ${currentUserId}`;
        const password = `${token},${url}`;

        if (this.waitForRehydration) {
            this.waitForRehydration = false;
            this.token = token;
            this.url = url;
        }

        // Only save to keychain if the url and token are set
        if (url && token) {
            try {
                setGenericPassword(username, password);
            } catch (e) {
                console.warn('could not set credentials', e); //eslint-disable-line no-console
            }
        }
    };

    setStartupThemes = (toolbarBackground, toolbarTextColor, appBackground) => {
        AsyncStorage.setItem(TOOLBAR_BACKGROUND, toolbarBackground);
        AsyncStorage.setItem(TOOLBAR_TEXT_COLOR, toolbarTextColor);
        AsyncStorage.setItem(APP_BACKGROUND, appBackground);
    };

    setStartAppFromPushNotification = (startAppFromPushNotification) => {
        this.startAppFromPushNotification = startAppFromPushNotification;
    };

    setIsNotificationsConfigured = (isNotificationsConfigured) => {
        this.isNotificationsConfigured = isNotificationsConfigured;
    };

    setAllowOtherServers = (allowOtherServers) => {
        this.allowOtherServers = allowOtherServers;
    };

    setAppStarted = (appStarted) => {
        this.appStarted = appStarted;
    };

    setEMMEnabled = (emmEnabled) => {
        this.emmEnabled = emmEnabled;
    };

    setDeviceToken = (deviceToken) => {
        this.deviceToken = deviceToken;
    };

    setReplyNotificationData = (replyNotificationData) => {
        this.replyNotificationData = replyNotificationData;
    };

    setInBackgroundSince = (inBackgroundSince) => {
        this.inBackgroundSince = inBackgroundSince;
    };

    setShouldRelaunchWhenActive = (shouldRelaunchWhenActive) => {
        this.shouldRelaunchWhenActive = shouldRelaunchWhenActive;
    };

    clearNativeCache = () => {
        resetGenericPassword();
        AsyncStorage.multiRemove([
            TOOLBAR_BACKGROUND,
            TOOLBAR_TEXT_COLOR,
            APP_BACKGROUND,
        ]);
    };

    handleDeepLink = (event) => {
        const {url} = event;
        store.dispatch(setDeepLinkURL(url));
    }

    launchApp = async () => {
        const shouldStart = await handleManagedConfig();
        if (shouldStart) {
            this.startApp();
        }
    };

    startApp = () => {
        if (this.appStarted || this.waitForRehydration) {
            return;
        }

        const {dispatch} = store;

        Linking.getInitialURL().then((url) => {
            dispatch(setDeepLinkURL(url));
        });

        let screen = 'SelectServer';
        if (this.token && this.url) {
            screen = 'Channel';
            tracker.initialLoad = Date.now();

            try {
                dispatch(loadMe());
            } catch (e) {
                // Fall through since we should have a previous version of the current user because we have a token
                console.warn('Failed to load current user when starting on Channel screen', e); // eslint-disable-line no-console
            }
        }

        switch (screen) {
        case 'SelectServer':
            EventEmitter.emit(ViewTypes.LAUNCH_LOGIN, true);
            break;
        case 'Channel':
            EventEmitter.emit(ViewTypes.LAUNCH_CHANNEL, true);
            break;
        }

        this.setAppStarted(true);
    }
}
