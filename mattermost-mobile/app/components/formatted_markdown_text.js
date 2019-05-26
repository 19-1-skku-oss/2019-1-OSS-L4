// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {Text} from 'react-native';
import PropTypes from 'prop-types';
import Renderer from 'commonmark-react-renderer';
import {Parser} from 'commonmark';
import {injectIntl, intlShape} from 'react-intl';

import {concatStyles, changeOpacity, makeStyleSheetFromTheme} from 'app/utils/theme';
import {getMarkdownTextStyles} from 'app/utils/markdown';
import CustomPropTypes from 'app/constants/custom_prop_types';

import AtMention from 'app/components/at_mention';
import MarkdownLink from 'app/components/markdown/markdown_link';

const TARGET_BLANK_URL_PREFIX = '!';

/*
* Translations component with a similar API to <FormattedText> component except the message string
* accepts markdown. It supports the following non-block-level markdown:
* - *italic*
* - **bold**
* - `inline code`
* - ~~strikethrough~~
* - [link](http://example.com/)
* - line\nbreaks
*
* Note: Line breaks (\n) in a defaultMessage parameter string must be surrounded by curly brackets {} in JSX. Example:
* <FormattedMarkdownText id='my.example' defaultMessage={'first line\nsecond line'} theme={theme} />
*/
class FormattedMarkdownText extends React.PureComponent {
    static propTypes = {
        baseTextStyle: CustomPropTypes.Style,
        defaultMessage: PropTypes.string.isRequired,
        id: PropTypes.string.isRequired,
        navigator: PropTypes.object.isRequired,
        onPostPress: PropTypes.func,
        style: CustomPropTypes.Style,
        textStyles: PropTypes.object,
        theme: PropTypes.object.isRequired,
        values: PropTypes.object,
    };

    static contextTypes = {
        intl: intlShape,
    };

    constructor(props) {
        super(props);
        this.parser = this.createParser();
        this.renderer = this.createRenderer();
        this.textStyles = null;
        this.baseTextStyle = null;
    }

    createParser = () => {
        return new Parser();
    }

    createRenderer = () => {
        return new Renderer({
            renderers: {
                text: this.renderText,
                emph: Renderer.forwardChildren,
                strong: Renderer.forwardChildren,
                code: this.renderCodeSpan,
                link: this.renderLink,
                hardBreak: this.renderBreak,
                softBreak: this.renderBreak,
                paragraph: this.renderParagraph,
                del: Renderer.forwardChildren,
                html_inline: this.renderHTML,
                html_block: this.renderHTML,
                atMention: this.renderAtMention,
            },
        });
    }

    computeTextStyle = (baseStyle, context) => {
        return concatStyles(baseStyle, context.map((type) => this.textStyles[type]));
    }

    renderText = ({context, literal}) => {
        const style = this.computeTextStyle(this.props.style || this.baseTextStyle, context);
        return <Text style={style}>{literal}</Text>;
    }

    renderCodeSpan = ({context, literal}) => {
        const style = this.computeTextStyle([this.baseTextStyle, this.textStyles.code], context);
        return <Text style={style}>{literal}</Text>;
    }

    renderLink = ({children, href}) => {
        var url = href[0] === TARGET_BLANK_URL_PREFIX ? href.substring(1, href.length) : href;
        return <MarkdownLink href={url}>{children}</MarkdownLink>;
    }

    renderBreak = () => {
        return <Text>{'\n'}</Text>;
    }

    renderParagraph = ({children}) => {
        return <Text>{children}</Text>;
    }

    renderHTML = (props) => {
        console.warn(`HTML used in FormattedMarkdownText component with id ${this.props.id}`); // eslint-disable-line no-console
        return this.renderText(props);
    }

    renderAtMention = ({context, mentionName}) => {
        return (
            <AtMention
                mentionStyle={this.props.textStyles.mention}
                mentionName={mentionName}
                navigator={this.props.navigator}
                onPostPress={this.props.onPostPress}
                textStyle={this.computeTextStyle(this.props.baseTextStyle, context)}
            />
        );
    }

    render() {
        const {id, defaultMessage, values, theme} = this.props;
        const messageDescriptor = {id, defaultMessage};
        const {formatMessage} = this.context.intl;
        const message = formatMessage(messageDescriptor, values);
        const ast = this.parser.parse(message);
        this.textStyles = getMarkdownTextStyles(theme);
        this.baseTextStyle = getStyleSheet(theme).message;
        return this.renderer.render(ast);
    }
}

const getStyleSheet = makeStyleSheetFromTheme((theme) => {
    return {
        message: {
            color: changeOpacity(theme.centerChannelColor, 0.8),
            fontSize: 15,
            lineHeight: 22,
        },
    };
});

export default injectIntl(FormattedMarkdownText);
