// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// Based on the work done by https://github.com/wcandillon/react-native-expo-image-cache/

import {Platform} from 'react-native';
import RNFetchBlob from 'rn-fetch-blob';

import {Client4} from 'mattermost-redux/client';

import {DeviceTypes} from 'app/constants';
import {
    getExtensionFromMime,
    getExtensionFromContentDisposition,
} from 'app/utils/file';
import mattermostBucket from 'app/mattermost_bucket';

const {IMAGES_PATH} = DeviceTypes;
const DEFAULT_MIME_TYPE = 'image/png';
let siteUrl;

export default class ImageCacheManager {
    static listeners = {};

    static cache = async (filename, uri, listener) => {
        if (!listener) {
            console.warn('Unable to cache image when no listener is provided'); // eslint-disable-line no-console
        }

        const {path, exists} = await getCacheFile(filename, uri);
        const prefix = Platform.OS === 'android' ? 'file://' : '';
        let pathWithPrefix = `${prefix}${path}`;

        if (exports.isDownloading(uri)) {
            addListener(uri, listener);
        } else if (exists) {
            listener(pathWithPrefix);
        } else {
            addListener(uri, listener);
            if (uri.startsWith('http')) {
                try {
                    const certificate = await mattermostBucket.getPreference('cert');
                    const options = {
                        session: uri,
                        timeout: 10000,
                        indicator: true,
                        overwrite: true,
                        path,
                        certificate,
                    };

                    const headers = {};
                    if (uri.includes(Client4.getUrl()) || uri.includes(siteUrl)) {
                        headers.Authorization = `Bearer ${Client4.getToken()}`;
                        headers['X-Requested-With'] = 'XMLHttpRequest';
                    }

                    this.downloadTask = await RNFetchBlob.config(options).fetch('GET', uri, headers);
                    if (this.downloadTask.respInfo.respType === 'text') {
                        throw new Error();
                    }

                    const contentDisposition = this.downloadTask.respInfo.headers['Content-Disposition'];
                    const mimeType = this.downloadTask.respInfo.headers['Content-Type'];
                    const ext = `.${
                        getExtensionFromContentDisposition(contentDisposition) ||
                        getExtensionFromMime(mimeType) ||
                        getExtensionFromMime(DEFAULT_MIME_TYPE)
                    }`;

                    if (!path.endsWith(ext)) {
                        const oldExt = path.substring(path.lastIndexOf('.'));
                        const newPath = path.replace(oldExt, ext);
                        await RNFetchBlob.fs.mv(path, newPath);

                        pathWithPrefix = `${prefix}${newPath}`;
                    }

                    notifyAll(uri, pathWithPrefix);
                } catch (e) {
                    RNFetchBlob.fs.unlink(pathWithPrefix);
                    notifyAll(uri, uri);
                    return null;
                }
            } else {
                // In case the uri we are trying to cache is already a local file just notify and return
                notifyAll(uri, uri);
            }

            unsubscribe(uri);
        }

        return pathWithPrefix;
    };
}

export const getCacheFile = async (name, uri) => {
    const filename = name || uri.substring(uri.lastIndexOf('/'), uri.indexOf('?') === -1 ? uri.length : uri.indexOf('?'));
    const defaultExt = `.${getExtensionFromMime(DEFAULT_MIME_TYPE)}`;
    const ext = filename.indexOf('.') === -1 ? defaultExt : filename.substring(filename.lastIndexOf('.'));

    let path = `${IMAGES_PATH}/${Math.abs(hashCode(uri))}${ext}`;

    try {
        const isDir = await RNFetchBlob.fs.isDir(IMAGES_PATH);
        if (!isDir) {
            await RNFetchBlob.fs.mkdir(IMAGES_PATH);
        }
    } catch (error) {
        // do nothing
    }

    let exists = await RNFetchBlob.fs.exists(path);
    if (!exists) {
        const pathWithDiffExt = await RNFetchBlob.fs.existsWithDiffExt(path);
        if (pathWithDiffExt) {
            exists = true;
            path = pathWithDiffExt;
        }
    }

    return {exists, path};
};

export const getSiteUrl = () => {
    return siteUrl;
};

export const setSiteUrl = (url) => {
    siteUrl = url;
};

export const isDownloading = (uri) => Boolean(ImageCacheManager.listeners[uri]);

const addListener = (uri, listener) => {
    if (!ImageCacheManager.listeners[uri]) {
        ImageCacheManager.listeners[uri] = [];
    }
    ImageCacheManager.listeners[uri].push(listener);
};

const unsubscribe = (uri) => Reflect.deleteProperty(ImageCacheManager.listeners, uri);

const notifyAll = (uri, path) => {
    ImageCacheManager.listeners[uri].forEach((listener) => {
        if (typeof listener === 'function') {
            listener(path);
        }
    });
};

// taken from https://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript-jquery
const hashCode = (str) => {
    let hash = 0;
    let i;
    let chr;
    if (!str || str.length === 0) {
        return hash;
    }

    for (i = 0; i < str.length; i++) {
        chr = str.charCodeAt(i);
        hash = ((hash << 5) - hash) + chr;
        hash |= 0; // Convert to 32bit integer
    }
    return hash;
};
