// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {latinise} from './latinise.js';
import {escapeRegex} from './markdown';

import {Files} from 'mattermost-redux/constants';

import {DeepLinkTypes} from 'app/constants';

const ytRegex = /(?:http|https):\/\/(?:www\.|m\.)?(?:(?:youtube\.com\/(?:(?:v\/)|(?:(?:watch|embed\/watch)(?:\/|.*v=))|(?:embed\/)|(?:user\/[^/]+\/u\/[0-9]\/)))|(?:youtu\.be\/))([^#&?]*)/;

export function isValidUrl(url = '') {
    const regex = /^https?:\/\//i;
    return regex.test(url);
}

export function stripTrailingSlashes(url = '') {
    return url.replace(/ /g, '').replace(/^\/+/, '').replace(/\/+$/, '');
}

export function removeProtocol(url = '') {
    return url.replace(/(^\w+:|^)\/\//, '');
}

export function extractFirstLink(text) {
    const pattern = /(^|[\s\n]|<br\/?>)((?:https?|ftp):\/\/[-A-Z0-9+\u0026\u2019@#/%?=()~_|!:,.;]*[-A-Z0-9+\u0026@#/%=~()_|])/i;
    let inText = text;

    // strip out code blocks
    inText = inText.replace(/`[^`]*`/g, '');

    // strip out inline markdown images
    inText = inText.replace(/!\[[^\]]*]\([^)]*\)/g, '');

    const match = pattern.exec(inText);
    if (match) {
        return match[0].trim();
    }

    return '';
}

export function isYoutubeLink(link) {
    return link.trim().match(ytRegex);
}

export function isImageLink(link) {
    let linkWithoutQuery = link;
    if (link.indexOf('?') !== -1) {
        linkWithoutQuery = linkWithoutQuery.split('?')[0];
    }

    for (let i = 0; i < Files.IMAGE_TYPES.length; i++) {
        const imageType = Files.IMAGE_TYPES[i];

        if (linkWithoutQuery.endsWith('.' + imageType) || linkWithoutQuery.endsWith('=' + imageType)) {
            return true;
        }
    }

    return false;
}

// Converts the protocol of a link (eg. http, ftp) to be lower case since
// Android doesn't handle uppercase links.
export function normalizeProtocol(url) {
    const index = url.indexOf(':');
    if (index === -1) {
        // There's no protocol on the link to be normalized
        return url;
    }

    const protocol = url.substring(0, index);
    return protocol.toLowerCase() + url.substring(index);
}

export function getShortenedURL(url = '', getLength = 27) {
    if (url.length > 35) {
        const subLength = getLength - 14;
        return url.substring(0, 10) + '...' + url.substring(url.length - subLength, url.length) + '/';
    }
    return url + '/';
}

export function cleanUpUrlable(input) {
    var cleaned = latinise(input);
    cleaned = cleaned.trim().replace(/-/g, ' ').replace(/[^\w\s]/gi, '').toLowerCase().replace(/\s/g, '-');
    cleaned = cleaned.replace(/-{2,}/, '-');
    cleaned = cleaned.replace(/^-+/, '');
    cleaned = cleaned.replace(/-+$/, '');
    return cleaned;
}

export function getScheme(url) {
    const match = (/([a-z0-9+.-]+):/i).exec(url);

    return match && match[1];
}

export function matchDeepLink(url, serverURL, siteURL) {
    if (!url || !serverURL || !siteURL) {
        return null;
    }

    const linkRoot = `(?:${escapeRegex(serverURL)}|${escapeRegex(siteURL)})?`;

    let match = new RegExp('^' + linkRoot + '\\/([^\\/]+)\\/channels\\/(\\S+)').exec(url);
    if (match) {
        return {type: DeepLinkTypes.CHANNEL, teamName: match[1], channelName: match[2]};
    }

    match = new RegExp('^' + linkRoot + '\\/([^\\/]+)\\/pl\\/(\\w+)').exec(url);
    if (match) {
        return {type: DeepLinkTypes.PERMALINK, teamName: match[1], postId: match[2]};
    }

    return null;
}

export function getYouTubeVideoId(link) {
    // https://youtube.com/watch?v=<id>
    let match = (/youtube\.com\/watch\?\S*\bv=([a-zA-Z0-9_-]{6,11})/g).exec(link);
    if (match) {
        return match[1];
    }

    // https://youtube.com/embed/<id>
    match = (/youtube\.com\/embed\/([a-zA-Z0-9_-]{6,11})/g).exec(link);
    if (match) {
        return match[1];
    }

    // https://youtu.be/<id>
    match = (/youtu.be\/([a-zA-Z0-9_-]{6,11})/g).exec(link);
    if (match) {
        return match[1];
    }

    return '';
}
