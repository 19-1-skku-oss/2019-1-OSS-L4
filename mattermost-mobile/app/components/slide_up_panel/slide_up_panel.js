// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {PureComponent} from 'react';
import PropTypes from 'prop-types';
import {Animated, Platform, StyleSheet, View} from 'react-native';
import {
    PanGestureHandler,
    NativeViewGestureHandler,
    State as GestureState,
    TapGestureHandler,
} from 'react-native-gesture-handler';

import {DeviceTypes} from 'app/constants';
import mattermostManaged from 'app/mattermost_managed';

import SlideUpPanelIndicator from './slide_up_panel_indicator';

export const BOTTOM_MARGIN = mattermostManaged.hasSafeAreaInsets ? 24 : 0;
const TOP_IOS_MARGIN = DeviceTypes.IS_IPHONE_X ? 84 : 64;
const TOP_ANDROID_MARGIN = 44;
const TOP_MARGIN = Platform.OS === 'ios' ? TOP_IOS_MARGIN : TOP_ANDROID_MARGIN;

export default class SlideUpPanel extends PureComponent {
    static propTypes = {

        // Whether or not to allow the panel to snap to the initial position after it has been opened
        allowStayMiddle: PropTypes.bool,

        containerHeight: PropTypes.number,
        children: PropTypes.oneOfType([
            PropTypes.arrayOf(PropTypes.node),
            PropTypes.node,
        ]).isRequired,
        header: PropTypes.func,
        headerHeight: PropTypes.number,

        // The initial position of the SlideUpPanel when it's first opened. If this value is between 0 and 1,
        // it is treated as a percentage of the containerHeight.
        initialPosition: PropTypes.number,

        // The space between the top of the panel and the top of the container when the SlideUpPanel is fully open.
        marginFromTop: PropTypes.number,
        onRequestClose: PropTypes.func,
    };

    static defaultProps = {
        allowStayMiddle: true,
        header: () => null,
        headerHeight: 0,
        initialPosition: 0.5,
        marginFromTop: TOP_MARGIN,
        onRequestClose: () => true,
    };

    constructor(props) {
        super(props);

        const {containerHeight, headerHeight, marginFromTop} = props;

        this.masterRef = React.createRef();
        this.panRef = React.createRef();
        this.scrollRef = React.createRef();
        this.scrollViewRef = React.createRef();
        this.headerRef = React.createRef();
        this.backdropRef = React.createRef();

        const initialUsedSpace = Math.abs(props.initialPosition);
        let initialPosition;
        if (initialUsedSpace <= 1) {
            initialPosition = ((containerHeight - (headerHeight + BOTTOM_MARGIN)) * (1 - initialUsedSpace));
        } else {
            initialPosition = ((containerHeight - (headerHeight + BOTTOM_MARGIN)) - initialUsedSpace);
        }

        // These values  correspond to when the panel is fully open, when it is initially opened, and when it is closed
        this.snapPoints = [marginFromTop, initialPosition, containerHeight];

        this.state = {
            lastSnap: initialPosition,
        };

        this.lastScrollYValue = 0;
        this.lastScrollY = new Animated.Value(0);
        this.onRegisterLastScroll = Animated.event(
            [{
                nativeEvent: {
                    contentOffset: {
                        y: this.lastScrollY,
                    },
                },
            }],
            {useNativeDriver: true},
        );
        this.lastScrollY.addListener(({value}) => {
            this.lastScrollYValue = value;
        });

        this.dragY = new Animated.Value(0);
        this.onGestureEvent = Animated.event(
            [{
                nativeEvent: {
                    translationY: this.dragY,
                },
            }],
            {useNativeDriver: true},
        );

        this.reverseLastScrollY = Animated.multiply(
            new Animated.Value(-1),
            this.lastScrollY
        );

        this.translateYOffset = new Animated.Value(containerHeight);
        this.translateY = Animated.add(
            this.translateYOffset,
            Animated.add(this.dragY, this.reverseLastScrollY)
        ).interpolate({
            inputRange: [marginFromTop, containerHeight],
            outputRange: [marginFromTop, containerHeight],
            extrapolate: 'clamp',
        });
    }

    componentDidMount() {
        Animated.timing(this.translateYOffset, {
            duration: 200,
            toValue: this.snapPoints[1],
            useNativeDriver: true,
        }).start();
    }

    closeWithAnimation = () => {
        Animated.timing(this.translateYOffset, {
            duration: 200,
            toValue: this.snapPoints[2],
            useNativeDriver: true,
        }).start(() => this.props.onRequestClose());
    };

    onHeaderHandlerStateChange = ({nativeEvent}) => {
        if (nativeEvent.oldState === GestureState.BEGAN) {
            this.lastScrollY.setValue(0);
            this.lastScrollYValue = 0;
        }
        this.onHandlerStateChange({nativeEvent});
    };

    onHandlerStateChange = ({nativeEvent}) => {
        if (nativeEvent.oldState === GestureState.ACTIVE) {
            const {translationY, velocityY} = nativeEvent;
            const {allowStayMiddle} = this.props;
            const {lastSnap} = this.state;
            const isGoingDown = translationY > 0;
            const translation = translationY - this.lastScrollYValue;

            const endOffsetY = lastSnap + translation;
            let destSnapPoint = this.snapPoints[0];

            if (Math.abs(translationY) < 50 && allowStayMiddle) {
                // Only drag the panel after moving 50 or more points
                destSnapPoint = lastSnap;
            } else if (isGoingDown && !allowStayMiddle) {
                // Just close the panel if the user pans down and we can't snap to the middle
                destSnapPoint = this.snapPoints[2];
            } else if (isGoingDown) {
                destSnapPoint = this.snapPoints.find((s) => s >= endOffsetY);
            } else {
                destSnapPoint = this.snapPoints.find((s) => s <= endOffsetY);
            }

            if (destSnapPoint) {
                this.translateYOffset.extractOffset();
                this.translateYOffset.setValue(translationY);
                this.translateYOffset.flattenOffset();
                this.dragY.setValue(0);

                if (destSnapPoint === this.snapPoints[2]) {
                    this.closeWithAnimation();
                } else {
                    Animated.spring(this.translateYOffset, {
                        velocity: velocityY,
                        tension: 68,
                        friction: 12,
                        toValue: destSnapPoint,
                        useNativeDriver: true,
                    }).start(() => {
                        this.setState({lastSnap: destSnapPoint});

                        // When dragging down the panel when is fully open reset the scrollView to the top
                        if (isGoingDown && destSnapPoint !== this.snapPoints[0]) {
                            this.scrollToTop();
                        }
                    });
                }
            } else {
                Animated.spring(this.translateYOffset, {
                    velocity: velocityY,
                    tension: 68,
                    friction: 12,
                    toValue: lastSnap,
                    useNativeDriver: true,
                }).start();
            }
        }
    };

    onSingleTap = ({nativeEvent}) => {
        if (nativeEvent.state === GestureState.ACTIVE) {
            this.closeWithAnimation();
        }
    };

    scrollToTop = () => {
        if (this.scrollViewRef?.current) {
            this.scrollViewRef.current._component.scrollTo({ //eslint-disable-line no-underscore-dangle
                x: 0,
                y: 0,
                animated: false,
            });
        }
    };

    render() {
        const {children, header} = this.props;
        const {lastSnap} = this.state;
        const translateStyle = {
            transform: [{translateY: this.translateY}],
        };

        const headerComponent = header(this.headerRef);

        return (
            <TapGestureHandler
                maxDurationMs={100000}
                ref={this.masterRef}
                maxDeltaY={lastSnap - this.snapPoints[0]}
            >
                <View
                    style={StyleSheet.absoluteFill}
                    pointerEvents='box-none'
                >
                    <TapGestureHandler
                        waitFor={this.backdropRef}
                        onHandlerStateChange={this.onSingleTap}
                    >
                        <Animated.View style={styles.viewport}>
                            <PanGestureHandler
                                simultaneousHandlers={[this.scrollRef, this.masterRef]}
                                waitFor={this.headerRef}
                                shouldCancelWhenOutside={false}
                                onGestureEvent={this.onGestureEvent}
                                onHandlerStateChange={this.onHandlerStateChange}
                                ref={this.backdropRef}
                            >

                                <Animated.View
                                    style={styles.backdrop}
                                    pointerEvents='box-only'
                                />
                            </PanGestureHandler>
                        </Animated.View>
                    </TapGestureHandler>
                    <Animated.View style={[StyleSheet.absoluteFill, translateStyle]}>
                        <PanGestureHandler
                            simultaneousHandlers={[this.scrollRef, this.masterRef]}
                            waitFor={this.headerRef}
                            shouldCancelWhenOutside={false}
                            onGestureEvent={this.onGestureEvent}
                            onHandlerStateChange={this.onHeaderHandlerStateChange}
                        >
                            <Animated.View>
                                <SlideUpPanelIndicator/>
                                {headerComponent}
                            </Animated.View>
                        </PanGestureHandler>
                        <PanGestureHandler
                            ref={this.panRef}
                            simultaneousHandlers={[this.scrollRef, this.masterRef]}
                            waitFor={this.headerRef}
                            shouldCancelWhenOutside={false}
                            onGestureEvent={this.onGestureEvent}
                            onHandlerStateChange={this.onHandlerStateChange}
                        >
                            <Animated.View style={[styles.container, !headerComponent && styles.border]}>
                                <NativeViewGestureHandler
                                    ref={this.scrollRef}
                                    waitFor={this.masterRef}
                                    simultaneousHandlers={this.panRef}
                                >
                                    <Animated.ScrollView
                                        ref={this.scrollViewRef}
                                        bounces={false}
                                        onScrollBeginDrag={this.onRegisterLastScroll}
                                        scrollEventThrottle={1}
                                        style={{marginBottom: (this.props.marginFromTop + BOTTOM_MARGIN)}}
                                    >
                                        {children}
                                    </Animated.ScrollView>
                                </NativeViewGestureHandler>
                            </Animated.View>
                        </PanGestureHandler>
                    </Animated.View>
                </View>
            </TapGestureHandler>
        );
    }
}

const styles = StyleSheet.create({
    viewport: {
        flex: 1,
    },
    container: {
        flex: 1,
        backgroundColor: 'white',
    },
    border: {
        ...Platform.select({
            android: {
                borderTopRightRadius: 2,
                borderTopLeftRadius: 2,
            },
            ios: {
                borderTopRightRadius: 10,
                borderTopLeftRadius: 10,
            },
        }),
    },
    backdrop: {
        backgroundColor: 'rgba(0, 0, 0, 0.3)',
        position: 'absolute',
        top: 0,
        bottom: 0,
        left: 0,
        right: 0,
    },
});
