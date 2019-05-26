// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Platform} from 'react-native';
import NetInfo from '@react-native-community/netinfo';
import RNFetchBlob from 'rn-fetch-blob';

import {Client4} from 'mattermost-redux/client';

import mattermostBucket from 'app/mattermost_bucket';
import mattermostManaged from 'app/mattermost_managed';

let certificate = '';
let previousState;
export async function checkConnection(isConnected) {
    if (!Client4.getBaseRoute().startsWith('http')) {
        // If we don't have a connection or have a server yet, return the default implementation
        return {hasInternet: isConnected, serverReachable: false};
    }

    // Ping the Mattermost server to detect if the we have network connection even if the websocket cannot connect
    const server = `${Client4.getBaseRoute()}/system/ping?time=${Date.now()}`;

    let managedConfig;
    let waitsForConnectivity = false;
    let timeoutIntervalForResource = 30;

    try {
        managedConfig = await mattermostManaged.getConfig();
    } catch {
        // no managed config
    }

    if (managedConfig?.useVPN === 'true') {
        waitsForConnectivity = true;
    }

    if (managedConfig?.timeoutVPN) {
        timeoutIntervalForResource = parseInt(managedConfig.timeoutVPN, 10);
    }

    const config = {
        auto: true,
        waitsForConnectivity,
        timeoutIntervalForResource,
    };

    if (Platform.OS === 'ios' && certificate === '') {
        certificate = await mattermostBucket.getPreference('cert');
        config.certificate = certificate;
    }

    try {
        await RNFetchBlob.config(config).fetch('GET', server);
        return {hasInternet: isConnected, serverReachable: true};
    } catch (error) {
        return {hasInternet: isConnected, serverReachable: false};
    }
}

function handleConnectionChange(onChange) {
    return async (isConnected) => {
        if (isConnected !== previousState) {
            previousState = isConnected;

            // Check if connected to server
            const result = await checkConnection(isConnected);
            onChange(result);
        }
    };
}

export default function networkConnectionListener(onChange) {
    const connectionChanged = handleConnectionChange(onChange);

    NetInfo.isConnected.fetch().then((isConnected) => {
        NetInfo.isConnected.addEventListener('connectionChange', connectionChanged);
        connectionChanged(isConnected);
    });

    const removeEventListener = () => NetInfo.isConnected.removeEventListener('connectionChange', connectionChanged);

    return {
        removeEventListener,
    };
}
