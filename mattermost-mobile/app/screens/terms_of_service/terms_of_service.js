// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {PureComponent} from 'react';
import PropTypes from 'prop-types';
import {
    Alert,
    ScrollView,
    View,
} from 'react-native';
import {intlShape} from 'react-intl';

import FailedNetworkAction from 'app/components/failed_network_action';
import Loading from 'app/components/loading';
import Markdown from 'app/components/markdown';
import StatusBar from 'app/components/status_bar';

import {t} from 'app/utils/i18n';
import {getMarkdownTextStyles, getMarkdownBlockStyles} from 'app/utils/markdown';
import {makeStyleSheetFromTheme, setNavigatorStyles} from 'app/utils/theme';

const errorTitle = {
    id: t('mobile.terms_of_service.get_terms_error_title'),
    defaultMessage: 'Unable to load terms of service.',
};

const errorDescription = {
    id: t('mobile.terms_of_service.get_terms_error_description'),
    defaultMessage: 'Make sure you have an active internet connection and try again. If this issue persists, contact your System Administrator.',
};

export default class TermsOfService extends PureComponent {
    static propTypes = {
        actions: PropTypes.shape({
            logout: PropTypes.func.isRequired,
            getTermsOfService: PropTypes.func.isRequired,
            updateMyTermsOfServiceStatus: PropTypes.func.isRequired,
        }).isRequired,
        closeButton: PropTypes.object,
        navigator: PropTypes.object,
        siteName: PropTypes.string,
        theme: PropTypes.object,
    };

    static contextTypes = {
        intl: intlShape,
    };

    static defaultProps = {
        siteName: 'Mattermost',
    };

    leftButton = {
        id: 'reject-terms-of-service',
    };

    rightButton = {
        id: 'accept-terms-of-service',
        showAsAction: 'always',
    };

    constructor(props, context) {
        super(props);

        this.state = {
            getTermsError: false,
            loading: true,
            termsId: '',
            termsText: '',
        };

        this.rightButton.title = context.intl.formatMessage({id: 'terms_of_service.agreeButton', defaultMessage: 'I Agree'});
        this.leftButton.icon = props.closeButton;

        props.navigator.setOnNavigatorEvent(this.onNavigatorEvent);
        this.setNavigatorButtons(false);
    }

    componentDidMount() {
        this.getTerms();
    }

    componentDidUpdate(prevProps) {
        if (this.props.theme !== prevProps.theme) {
            setNavigatorStyles(this.props.navigator, this.props.theme);
        }
    }

    onNavigatorEvent = (event) => {
        if (event.type === 'NavBarButtonPress') {
            switch (event.id) {
            case 'close-terms-of-service':
                this.closeTermsAndLogout();
                break;

            case 'reject-terms-of-service':
                this.handleRejectTerms();
                break;

            case 'accept-terms-of-service':
                this.handleAcceptTerms();
                break;
            }
        }
    };

    setNavigatorButtons = (enabled = true) => {
        const buttons = {
            leftButtons: [{...this.leftButton, disabled: !enabled}],
            rightButtons: [{...this.rightButton, disabled: !enabled}],
        };

        this.props.navigator.setButtons(buttons);
    };

    enableNavigatorLogout = () => {
        const buttons = {
            leftButtons: [{...this.leftButton, id: 'close-terms-of-service', disabled: false}],
            rightButtons: [{...this.rightButton, disabled: true}],
        };

        this.props.navigator.setButtons(buttons);
    };

    closeTermsAndLogout = () => {
        const {actions} = this.props;

        this.props.navigator.dismissAllModals();
        actions.logout();
    };

    getTerms = async () => {
        const {actions} = this.props;

        this.setState({
            termsId: '',
            termsText: '',
            loading: true,
            getTermsError: false,
        });
        this.setNavigatorButtons(false);

        const {data} = await actions.getTermsOfService();
        if (data) {
            this.setState({
                termsId: data.id,
                termsText: data.text,
                loading: false,
            }, () => {
                this.setNavigatorButtons(true);
            });
        } else {
            this.setState({
                loading: false,
                getTermsError: true,
            });
            this.enableNavigatorLogout();
        }
    };

    handleAcceptTerms = () => {
        this.registerUserAction(
            true,
            () => {
                this.props.navigator.dismissModal({
                    animationType: 'slide-down',
                });
            },
            this.handleAcceptTerms
        );
    };

    handleRejectTerms = () => {
        const {siteName} = this.props;
        const {intl} = this.context;

        this.registerUserAction(
            false,
            () => {
                Alert.alert(
                    this.props.siteName,
                    intl.formatMessage({
                        id: 'mobile.terms_of_service.terms_rejected',
                        defaultMessage: 'You must agree to the terms of service before accessing {siteName}. Please contact your System Administrator for more details.',
                    }, {
                        siteName,
                    }),
                    [{
                        text: intl.formatMessage({id: 'mobile.terms_of_service.alert_ok', defaultMessage: 'OK'}),
                        onPress: this.closeTermsAndLogout,
                    }],
                );
            },
            this.handleRejectTerms
        );
    };

    registerUserAction = async (accepted, success, retry) => {
        const {actions} = this.props;
        const {intl} = this.context;

        this.setNavigatorButtons(false);
        this.setState({
            loading: true,
        });

        const {data} = await actions.updateMyTermsOfServiceStatus(this.state.termsId, accepted);

        this.setState({
            loading: false,
        });

        if (data) {
            success(data);
            this.setNavigatorButtons(true);
        } else {
            Alert.alert(
                this.props.siteName,
                intl.formatMessage({
                    id: 'terms_of_service.api_error',
                    defaultMessage: 'Unable to complete the request. If this issue persists, contact your System Administrator.',
                }),
                [{
                    text: intl.formatMessage({id: 'mobile.terms_of_service.alert_cancel', defaultMessage: 'Cancel'}),
                    style: 'cancel',
                    onPress: this.closeTermsAndLogout,
                }, {
                    text: intl.formatMessage({id: 'mobile.terms_of_service.alert_retry', defaultMessage: 'Try Again'}),
                    onPress: retry,
                }],
            );
        }
    };

    render() {
        const {navigator, theme} = this.props;
        const styles = getStyleSheet(theme);

        const blockStyles = getMarkdownBlockStyles(theme);
        const textStyles = getMarkdownTextStyles(theme);

        if (this.state.loading) {
            return <Loading/>;
        }

        if (this.state.getTermsError) {
            return (
                <View style={styles.container}>
                    <StatusBar/>
                    <FailedNetworkAction
                        onRetry={this.getTerms}
                        theme={theme}
                        errorTitle={errorTitle}
                        errorDescription={errorDescription}
                    />
                </View>
            );
        }

        return (
            <React.Fragment>
                <StatusBar/>
                <ScrollView
                    style={styles.scrollView}
                    contentContainerStyle={styles.scrollViewContent}
                >
                    <Markdown
                        baseTextStyle={styles.baseText}
                        navigator={navigator}
                        textStyles={textStyles}
                        blockStyles={blockStyles}
                        value={this.state.termsText}
                        disableHashtags={true}
                        disableAtMentions={true}
                        disableChannelLink={true}
                    />
                </ScrollView>
            </React.Fragment>
        );
    }
}

const getStyleSheet = makeStyleSheetFromTheme((theme) => {
    return {
        baseText: {
            color: theme.centerChannelColor,
            fontSize: 15,
            lineHeight: 20,
        },
        container: {
            backgroundColor: theme.centerChannelBg,
            flex: 1,
        },
        linkText: {
            color: theme.linkColor,
        },
        scrollView: {
            flex: 1,
            backgroundColor: theme.centerChannelBg,
            padding: 30,
        },
        scrollViewContent: {
            paddingBottom: 50,
        },
    };
});
