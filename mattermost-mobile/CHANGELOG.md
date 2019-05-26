# Mattermost Mobile Apps Changelog

## 1.19.0 Release
- Release Date: May 16, 2019
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility
 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).
 - iPhone 5s devices and later with iOS 11+ is required.
 
### Bug Fixes
 - Fixed an issue where Android managed config was lost on the thread view.
 - Fixed an issue where contents of ephemeral posts did not display on the mobile app.
 - Fixed a few mobile app crash / fatal error issues.
 - Fixed an issue with an expanding animation when tapping on Jump to Channel in the channel list.
 - Fixed an issue on iOS where animated custom emoji weren't animated.
 - Fixed an issue on iOS where users were unable to create channel name of 2 characters.
 - Fixed an issue on iOS where emoji appeared too close, with uneven spacing, and too small in the info modal.
 - Added an error handler when sharing text that was over server's maximum post size with the iOS Share Extension.
 - Fixed an issue where users could upload a GIF as a profile image.
 
### Known Issues
 - Buttons inside ephemeral posts are not clickable / functional on the mobile app.

## 1.18.1 Release
- Release Date: April 18, 2019
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility
 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).
 - iPhone 5s devices and later with iOS 11+ is required.

### Bug Fixes 
 - Fixed a crash issue caused by a malformed post textbox localize string.
 - Fixed an issue where iOS crashed when trying to log in using SSO and the SSO provider set a cookie without an expiration date.

## 1.18.0 Release
- Release Date: April 16, 2019
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility
 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).
 - iPhone 5s devices and later with iOS 11+ is required.
 - ``Bot`` tags were added for bot accounts feature in server v5.10 and mobile v1.18, meaning that mobile v1.17 and earlier don't support the tags.
 
### Highlights
 - Added support for Office365 single sign-on (SSO).
 - Added support for Integrated Windows Authentication (IWA).

### Improvements
 - Added the ability for channel links to open inside the app.
 - Added ability for emojis and hyperlinks to render in the message attachment title.
 - Added Chinese support for words that trigger mentions.
 - Added a setting to the system console to change the minimum length of hashtags.
 - Added a reply option to long press context menu.

### Bug Fixes
 - Fixed an issue where blank spaces broke markdown tables.
 - Fixed an issue where deactivated users appeared on "Add Members" modal but not on the search results.
 - Fixed an issue on Android where extra text in the search box appeared after using the autocomplete drop-down.
 - Fixed an issue with multiple text entries when typing with Shift+Letter on Android.
 - Fixed an issue where push notifications badges did not always clear when read on another device.
 - Fixed an issue where opening a single or group notification did not take the user into the channel where the notification came from.
 - Fixed an issue where timezone did not automatically update on Android when travelling to another timezone.
 - Fixed an issue where the user mention autocomplete drop-down was case sensitive.
 - Fixed an issue where system admininistrators were able to see the full long press menu when long pressing a system message.
 - Fixed an issue where users were not able to unflag posts from "Flagged Posts" when opened from a read-only channel.
 - Fixed an issue where users were unable to create channel names of 2 byte characters.
 
### Known Issues
 - Content for ephemeral messages is not displayed on Mattermost Mobile Apps.

## 1.17.0 Release
- Release Date: March 20, 2019
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility
 - If **DisableLegacyMfa** setting in ``config.json`` is set to ``true`` and [multi-factor authentication](https://docs.mattermost.com/deployment/auth.html) is enabled, ensure your users have upgraded to mobile app version 1.17 or later. See [Important Upgrade Notes](https://docs.mattermost.com/administration/important-upgrade-notes.html) for more details.
 - If you are using an EMM provider via AppConfig, make sure to add two new settings, `useVPN` and `timeoutVPN`, to your AppConfig file. The settings were added for EMM connections using VPN on-demand - one to indicate if every request should wait for the VPN connection to be established, and another to set the timeout in seconds. See docs for more details on [setting AppConfig values](https://docs.mattermost.com/mobile/mobile-appconfig.html#mattermost-appconfig-values) for VPN support.
 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).
 - iPhone 5s devices and later with iOS 11+ is required.
 
### Highlights
 - iOS Share Extension now supports large file sizes and improved performance

### Bug Fixes
 - Fixed support for EMM connections using VPN on-demand. See docs for more details on [setting AppConfig values](https://docs.mattermost.com/mobile/mobile-appconfig.html#mattermost-appconfig-values) for VPN support.
 - Fixed several Android app crash / fatal error issues.
 - Fixed an issue on Android where the app crashed intermittently when selecting a link.
 - Fixed an issue where email notifications setting was out of sync with the webapp until the setting was edited.
 - Fixed an issue where notification badges were not cleared from other clients when clicking on a push notification after opening the mobile app.
 - Fixed an issue where the app did not show local notification when session expired.
 - Fixed an issue where the profile picture for webhooks was showing the hook owner picture.
 - Fixed an issue where some emoji were not rendered as jumbo.
 - Fixed an issue where jumbo emoji posted as a reply sometimes appeared with large space beneath.
 - Fixed an issue where the "No Internet Connection" banner did not always display when internet connectivity was lost.
 - Fixed an issue where the "No Internet Connection" banner did not always disappear when connection was re-estabilished.
 - Fixed an issue where opening channels with unreads had loading indicator placed above unread messages line.

## 1.16.1 Release
- Release Date: February 21, 2019
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility
 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).

### Bug Fixes
 - Fixed an issue where link previews and reactions weren't displayed when post metadata was disabled.
 - Fixed an issue on Android where the app crashed when sharing multiple files.
 
## 1.16.0 Release
- Release Date: February 16, 2019
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility

 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).

### Improvements
 - Added the ability to remove own profile picture.
 - Changed "X" to "Cancel" on Edit Profile page.
 - Added support for relative permalinks.

### Bug Fixes
 - Fixed an issue where the iOS app did not wait until the on-demand VPN connection was established. (EMM Providers)
 - Fixed an issue with a white screen caused by missing Russian translations.
 - Fixed an issue where the iOS badge notification did not always clear.
 - Fixed an issue where the thread view displayed a new message indicator.
 - Fixed an issue where quick multiple taps on the file icon opened multiple file previews.
 - Fixed an issue where the settings page did not show an option to join other teams.
 - Fixed an issue where image previews didn't work after using Delete File Cache.
 - Fixed an issue on Android where the notification trigger word modal title was "Send email notifications" instead of "Keywords".
 - Fixed an issue where the Webhook icon was misaligned and bottom edges were cut off.
 - Fixed an issue on Android where the user was not asked to authenticate to the app first when trying to share a photo, resulting in a white "Share modal" screen with a never-ending loading indicator.
 - Fixed an issue on iOS where push notifications were not preserved when opening the app via the Mattermost icon.

## 1.15.2 Release
- Release Date: January 16, 2019
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility

 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).

### Bug Fixes

 - Fixed an issue where the status changes for other users did not always stay current in the mobile app.
 - Fixed an issue where a post did not fail properly when the user attempted to send the post while there was no network access.
 - Fixed an issue where date separators did not update when changing timezones.
 - Fixed an issue where the Favorites section did not clear from a users's channel drawer.
 - Removed an extra divider below "Edit Channel" of Direct Message Channel Info.
 - Fixed an issue where a user was not returned to previously viewed channel after viewing and then closing an archived channel.
 - Fixed an issue where a quick double tap on switch of Channel Info created and extra on/off state.
 - Fixed an issue where iOS long press menu didn't have rounded corners.

## 1.15.1 Release
- Release Date: December 28, 2018
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility

 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).
 
### Bug Fixes
 - Fixed an issue preventing some users from logging in using OKTA.

## 1.15.0 Release
- Release Date: December 16, 2018
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

### Combatibility

 - Mobile App v1.13+ is required for Mattermost Server v5.4+.
 - Android operating system 7+ [is required by Google](https://android-developers.googleblog.com/2017/12/improving-app-security-and-performance.html).

### Highlights
 - Added mention and reply mention highlighting.
 - Added a sliding animation for the reaction list.
 - Added support for pinned posts.
 - Added support for jumbo emojis.
 - Added support for interactive dialogs.
 - Improved UI for the long press menu and emoji reaction viewer.

### Improvements
 - Added the ability to include custom headers with requests for custom builds.
 - Push Notifications that are grouped by channels are cleared once the channel is read.
- Improved auto-reconnect when unable to reach the server.
 - Added support for changing the mobile client status to offline when the app loses connection.
 - Added 'View Members' button to archived channels.
 - Added support on iOS for keeping the postlist in place without scrolling when new content is available.

### Bug Fixes
 - Fixed an issue where clicking on a file did not show downloading progress.
 - Fixed an issue on Android where on fresh install the share extension would not properly show available channels.
 - Fixed an issue where recently archived channels remained in in: autocomplete when they had been archived.
 - Fixed an issue where text should render when no actual custom emoji matched the named emoji pattern.
 - Fixed an issue on iOS where text got cut-off after replying to a message.
 - Fixed an issue where search modifier for channels was showing Direct Messages without usernames.
 - Fixed an issue where "Close Channel" did not work properly when viewing two archived channels in a row.
 - Fixed an issue with "Critical Error" screen when trying to upload certain file types from "+" to the left of message input box.

## 1.14.0 Release
- Release Date: November 16, 2018
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported unless the user installs the CA certificate on their device

**Combatibility Note: Mobile App v1.13+ is required for Mattermost Server v5.4+**

### Bug Fixes
- Fixed an issue where the Android app did not allow establishing a network connection with any server that used a self-signed certificate that had the CA certificate user installed on the device.
- Removed "Copy Post" option on long-press message menu for posts without text.
- Fixed an issue where the "Search Results" header was not fully scrolled to top on search "from:username".
- Fixed an issue where channel names truncated at fewer characters than necessary.
- Fixed an issue where the same uploaded photo generated a different file size.
- Fixed an issue where the "(you)" was not displayed to the right of a user's name in the channel drawer when a user opened a Direct Message channel with themself.
- Fixed an issue where a dark theme set from webapp broke mobile display.
- Fixed an issue where channel drawer transition sometimes lagged.
- Fixed an issue where sending photos to Mattermost created large files.
- Fixed an issue where the apps showed "Select a Team" screen when opened.
- Fixed an issue where at-mention, emoji, and slash command autocompletes had a double top border.
- Fixed an issue where the drawer was unable to close when showing the team list.
- Fixed an issue where team sidebar showed + sign even without more teams to join.


## 1.13.1 Release
- Release Date: October 18, 2018
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported

**Combatibility Note: Mobile App v1.13+ is required for Mattermost Server v5.4+**

### Bug Fixes
- Fixed an issue preventing some users from authenticating using OKTA

## v1.13.0 Release
- Release Date: October 16, 2018
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported

**Combatibility Note: Mobile App v1.13+ is required for Mattermost Server v5.4+**

### Highlights

#### View Emoji Reactions
- Hold down on any emoji reaction to see who reacted to the post.

#### Hashtags
- Added support for searching for hashtags in posts.

#### Dropdown menus
- Added support for dropdown menus in message attachments.

### Improvements
- Added support for iPhone XR, XS and XS Max.
- Added support for nicknames on user profile.
- On servers 5.4+, added support for searching in direct and group message channels using the "in:" modifier.
- Channel autocomplete now gets closed if multiple tildes are typed.
- Added a draft icon in sidebar and channel switcher for channels with unsent messages.
- Users are now redirected to the archived channel view (rather than to Town Square) when a channel is archived.
- When closing an archived channel, users are now returned to the previously viewed channel.

### Bug Fixes
- Refactored postlist to include Android Pie fixes and smoother scrolling.
- Fixed an issue where deactivated users were not marked as such in "Jump To" search.
- Fixed an issue where users got a permission error when trying to open a file from within the image preview screen.
- Fixed an issue where session expiry notifications were not being sent on Android.
- Fixed an issue where post attachments failed to upload.
- Fixed an issue where the "DM More..." list cut off user info.
- Fixed an issue where the user would briefly see a system message when loading a reply thread.
- Fixed an issue where the error message was incorrectly formatted if the login method was set to email/password and the user tried to log in with SAML.
- Fixed an issue on Android where the keyboard sometimes overlapped the bottom of the post textbox.
- Fixed an issue where there was no option to take video via "+" > "Take Photo or Video" on iOS.

## v1.12.0 Release
- Release Date: September 16, 2018
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported

### Highlights

#### Search Date Filters
- Search for messages before, on, or after a specified date.

### Improvements
- Added notification support for Android O and P.

### Bug Fixes
- Fixed an issue where Okta was not able to login in some deployments.
- Fixed an issue where messages in Direct Message channels did not show when clicking "Jump To".
- Fixed an issue where `Show More` on a post with a message attachment displayed a blank where content should have been.
- Prevent downloading of files when disallowed in the System Console.
- Fixed an issue where users could not click on attachment filenames to open them.
- Fixed an issue where email notification settings did not save from mobile.
- Fixed an issue where the share extension allowed users to select and attempt to share content to channels that had been archived.
- Fixed an issue where reacting to an existing emoji in an archived channel was allowed.
- Fixed an issue where archived channels sometimes remained in the drawer.
- Fixed an issue where deactivated users were not marked as such in Direct Message search.


## v1.11.0 Release
- Release Date: August 16, 2018
- Server Versions Supported: Server v4.10+ is required, Self-Signed SSL Certificates are not supported

### Highlights

#### Searching Archived Channels
- Added ability to search for archived channels. Requires Mattermost server v5.2 or later.

#### Deep Linking
- Added the ability for custom builds to open Mattermost links directly in the app rather than the default mobile browser. Learn more in our [documentation](https://docs.mattermost.com/mobile/mobile-faq.html#how-do-i-configure-deep-linking)

### Improvements
- Added profile pop-up to combined system messages.
- Force re-entering SSO auth credentials after logout.
- Added consecutive posts by the same user.
- Added a loading indicator when user info is still loading in the left-hand side.

### Bug Fixes
- Fixed an issue where Android devices showed an incorrect timestamp.
- Fixed an issue on Android where the app did not get sent to the background when pressing the hardware back button in the channel screen.
- Fixed an issue with video playback when the filename had spaces.
- Fixed an issue where the app crashed when playing YouTube videos.
- Fixed an issue with session expiration notification.
- Fixed an issue with sharing files from Google Drive in Android Share Extension.
- Fixed an issue on Android where replying to a push notification sometimes went to the wrong channel.
- Fixed an issue where the previous server URL was present on the input textbox before changing the screen to Login.
- Fixed an issue where user menu was not translated correctly.
- Fixed an issue where some field lengths in Account Settings didn't match the desktop app.
- Fixed an issue where long URLs for embedded images in message attachments got cut off and didn't render.
- Fixed an issue where link preview images were not cropped properly.
- Fixed an issue where long usernames didn't wrap properly in the Account Settings menu.
- Fixed an issue where DMs would not open if users were using "Jump To".
- Fixed an issue where no message was displayed after removing a user from a channel with join/leave messages disabled.

## v1.10.0 Release
- Release Date: July 16, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights

#### Channel drawer performance
- Android devices will notice significant performance improvements when opening and closing the channel drawer.

#### Channel loading performance
- Improved channel loading performance as post are retrieved with every push notification

#### Announcement banner improvements
- Markdown now renders when announcement banners are expanded
- When enabled by the System Admin, users can now dismiss announcement banners until their next session

### Improvements

 - Combined consecutive messages from the same user.
 - Added experimental support for certificate-based authentication (CBA) for iOS to identify a user or a device before granting access to Mattermost. See [documentation](https://docs.mattermost.com/deployment/certificate-based-authentication.html) to learn more.
 - Added support for the experimental automatic direct message replies feature.
 - Added support for the experimental timezone feature.
 - Changed post textbox to not be a connected component.
 - Allow connecting to mattermost instances hosted at subpaths.
 - Added support for starting YouTube videos at a given time.
 - Added support for keeping messages if slash command fails.

### Bug Fixes

 - Fixed an issue where the unread badge background was always white.
 - Fixed an issue where a username repeated in system message if user was added to a channel more than once.
 - Fixed an issue where Android Sharing from Microsoft apps failed.
 - Fixed an issue where YouTube crashed the app if link did not have a time set.
 - Fixed an issue where System Admins did not see all teams available to join on mobile.
 - Fixed an issue where users were unable to share from Files app.
 - Fixed an issue where viewing a non-existent permalink didn't show an error message.
 - Fixed an issue where jumping to a channel search did not bold unread channels.
 - Fixed an issue with being able to add own user to a Group Message channel.
 - Fixed an issue with not being able to reply from a push notification on iOS.
 - Fixed an issue where the app did not display Brazilian language.
 
## 1.9.3 Release
- Release Date: July 04, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes

- Fixed multiple issues causing app crashes
- Fixed an issue on iOS devices with typing non-english characters in the post input box

## 1.9.2 Release
- Release Date: June 27, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes

- Fixed an issue where attached videos did not play for the poster
- Fixed an issue where "Jump to recent messages" from the permalink view did not direct the user to the bottom of the channel
- Fixed an issue where post comments did not identify which parent post they belonged to
- Fixed multiple issues with typing non-english characters in the post input box
- Fixed multiple issues causing random app crashes
- Fixed an issue where files from the Android Files app failed to upload
- Fixed an issue where the iOS share extension crashed when switching the team or channel
- Fixed an issue where files from the Microsoft app failed to upload
- Fixed an issue on Android devices where sharing files changed the file extension of the attachment

## 1.9.1 Release
- Release Date: June 23, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes
- Fixed an issue with typing lag on Android devices
- Fixed an issue causing users to be logged out after upgrading to v1.9.0
- Fixed an issue where the ``in:`` and ``from:`` modifiers were not being added to the search field

## v1.9.0 Release
- Release Date: June 16, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights

#### Improved first load time on Android
 - Significantly decreased first load time on Android devices from cold start.
 
#### iOS Files app support
- Added support for attaching files from the iOS Files app from within Mattermost.

#### Improved styling of push notification
- Improved the layout of message content, channel name and sender name in push notifications.

### Improvements

 - Combined join/leave system messages.
 - Added splash screen and channel loader improvements.
 - Removed the desktop notification duration setting.
 - Added cache team icon and set background to always be white if using a PNG file.
 - Added whitelabel for icons and splash screen.

### Bug Fixes

 - Fixed an issue where other user's display name did not render in combined system messages after joining the channel.
 - Fixed an issue where posts incorrectly had "Commented on Someone's message" above them.
 - Fixed an issue where deleting a post or its parent in permalink view left permalink view blank.
 - Fixed an issue where "User is typing" message cut was off.
 - Fixed an issue where `More New Messages Above` appeared at the top of new channel on joining.
 - Fixed an issue where a user was not directed to Town Square when leaving a channel.
 - Fixed an issue where long post were not collapsed on Android.
 - Fixed an issue where a user's name was initially shown as "someone" when opening a direct message with the user.
 - Fixed an issue where an error was received when trying to change the team or channel from the share extension.
 - Fixed an issue where switching to a newly created channel from a push notification redirected a user to Town Square.
 - Fixed an issue where a public channel made private did not disappear automatically from clients not part of the channel.

## v1.8.0 Release
- Release Date: April 27, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights

#### Image performance
- Images are now downloaded and stored locally for better performance

#### Flagged Posts and Recent Mentions
- Access all your flagged posts and recent mentions from the buttons in the sidebar

#### Muted Channels
- Added support for Muted Channels released with Mattermost server v4.9 

### Improvements
- Date separators now appear between each posts in the search view
- Deactivated users are now filtered out of the channel members lists
- Direct Messages user list is now sorted by username first
- Added the option to Direct Message yourself from your user profile screen
- Improved performance on the post list
- Improved matching and display when searching for users in the Direct Message user list

### Bug Fixes
- Fixed an issue where emoji reactions could be added from the search view but did not appear
- Fixed an issue causing the app to crash when trying to share content from a custom keyboard
- Fixed an issue where team names were being sorted based on letter case
- Fixed an issue where username would not be inserted to the post draft when using experimental configuration settings
- Fixed an issue with nested bullet lists being cut off in the user interface
- Fixed an issue where private channels were listed in the public channels section of the channel autocomplete list
- Fixed an issue where a profile images could not be updated from the app

## v1.7.1 Release
- Release Date: April 3, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes
- Fixed an issue where the iOS share extension sometimes crashed the Mattermost app
- Fixed an issue preventing Markdown tables from rendering with some international characters 

## v1.7.0 Release
- Release Date: March 26, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights

#### iOS File Sharing
- Share files and images from other applications as attached files in Mattermost

#### Markdown Tables
- Tables created using markdown formatting can now be viewed in the app

#### Permalinks
- Permalinks now open in the app instead of launching a browser window 

### Improvements
- Increased the tappable area of various icons for improved usability
- Announcement banners now display in the app
- Added "+" button to add emoji reactions to a post
- Minor performance improvements for app launch time
- Text files can now be viewed in the app
- Support for email autolinking into the app

### Bugs
- Fixed an issue causing some devices to hang at the splash screen on app launch
- Fixed an issue causing some letters to be hidden in the Android search input box
- Fixed an issue causing some Direct Message channels to show date stamps below the most recent message
- Fixed an issue where users weren't able to join open teams they've never been a member of
- Fixed an issue so double tapping buttons can no longer cause UI issues
- Fixed an issue where changing the channel display name wasn't being updated in the UI appropriately
- Fixed an issue where searhing for public channels sometimes showed no results
- Fixed an issue where the post menu could remain open while scrolling in the post list
- Fixed an issue where the system message to add users to a channel was missing the execution link
- Fixed an issue where bulleted lists cut off text if nested deeper than two levels
- Fixed an issue where logging into an account that is not on any team freezes the app
- Fixed an issue on iOS causing the app to crash when taking a photo then attaching it to a post

## v1.6.1 Release
- Release Date: February 13, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes
- Fixed an issue preventing the app from going to the correct channel when opened from a push notification
- Fixed an issue on Android devices where the app could sometimes freeze on the launch screen
- Fixed an issue on Samsung devices causing extra letters to be insterted when typing to filter user lists

## v1.6.0 Release
- Release Date: February 6, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights

#### Android File Sharing
- Share files and images from other applications as attached files in Mattermost 

### Improvements
- Added a right drawer to access settings, edit profile information, change online status and logout
- Added support for opening a Direct Message channel with yourself

### Bugs
- Fixed a number of issues causing crashes on Android devices
- Fixed an issue with auto capitalization on Android keyboards
- Fixed an issue where the GitLab SSO login button sometimes didn't appear
- Fixed an issue with link previews not appearing on some accounts
- Fixed an issue where logging out of the app didn't clear the notification badge on the homescreen icon
- Fixed an issue where interactive message buttons would not wrap to a new line
- Fixed an issue where the keyboard would sometimes overlap the text input box
- Fixed an issue where the Direct Message channel wouldn't open from the profile page
- Fixed an issue where posts would sometimes overlap
- Fixed an issue where the app sometimes hangs on logout

## v1.5.3 Release
- Release Date: February 1, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported
- Fixed a login issue when connecting to servers running a Data Retention policy 

## v1.5.2 Release
- Release Date: January 12, 2018
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes
- Fixed an issue causing some Android devices to crash on launch
- Fixed an issue with the app occasionally crashing when receiving push notifications in a new channel 
- Channel footer area is now refreshed when switching between Group and Direct Message channels
- Fixed an issue on some Android devices so Mattermost verifies it has permissions to access ringtones
- Fixed an issue where the text box overlapped the keyboard on some iOS devices using multiple keyboard layouts
- Fixed an issue with video uploads on Android devices
- Fixed an issue with GIF uploads on iOS devices
- Fixed an issue with the mention badge flickering on the channel drawer icon when there were over 10 unread mentions
- Fixed an issue with the app occasionally freezing when requesting the RefreshToken

## v1.5.1 Release

- Release Date: December 7, 2017
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes
- Fixed an issue with the upgrade app screen showing with a transparent background
- Fixed an issue with clearing or replying to notifications sometimes crashing the app on Android
- Fixed an issue with the app sometimes crashing due to a missing function in the swiping control

## v1.5 Release 

- Release Date: December 6, 2017
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights 

#### File Viewer
- Preview videos, RTF,  PDFs, Word, Excel, and Powerpoint files 

#### iPhone X Compatibility
- Added support for iPhone X

#### Slash Commands
- Added support for using custom slash commands
- Added support for built-in slash commands /away, /online, /offline, /dnd, /header, /purpose, /kick, /me, /shrug

### Improvements
- In iOS, 3D touch can now be used to peek into a channel to view the contents, and quickly mark it as read
- Markdown images in posts now render 
- Copy posts, URLs, and code blocks
- Opening a channel with Unread messages takes you to the "New Messages" indicator 
- Support for data retention, interactive message buttons, and viewing Do Not Disturb statuses depending on the server version
- (Edited) indicator now shows up beside edited posts 
- Added a "Recently Used" section for emoji reactions

### Bug Fixes 
- Android notifications now follow the default system setting for vibration 
- Fixed app crashing when opening notification settings on Android 
- Fixed an issue where the "Proceed" button on sign in screen stopped working after pressing logout multiple times
- HEIC images posted from iPhones now get converted to JPEG before uploading

## v1.4.1 Release

Release Date: Nov 15, 2017
Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Bug Fixes

- Fixed network detection issue causing some people to be unable to access the app
- Fixed issue with lag when pressing send button 
- Fixed app crash when opening notification settings
- Fixed various other bugs to reduce app crashes

## v1.4 Release 

- Release Date: November 6, 2017
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights 

#### Performance improvements
- Various performance improvements to decrease channel load times 

### Bug Fixes
- Fixed issue with Android app sometimes showing a white screen when re-opening the app
- Fixed an issue with orientation lock not working on Android 

## v1.3 Release 

- Release Date: October 5, 2017
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights 

#### Tablet Support (Beta) 
- Added support for landscape view, so the app may be used on tablets
- Note: Tablet support is in beta, and further improvements are planned for a later date

#### Link Previews 
- Added support for image, GIF, and youtube link previews

#### Notifications
- Android: Added the ability to set light, vibrate, and sound settings
- Android: Improved notification stacking so most recent notification shows first 
- Updated the design for Notification settings to improve usability 
- Added the ability to reply from a push notification without opening the app (requires Android v7.0+, iOS 10+) 
- Increased speed when opening app from a push notification

#### Download Files 
- Added the ability to download all files on Android and images on iOS

### Improvements
- Using `+` shortcut for emoji reactions is now supported 
- Improved emoji formatting (alignment and rendering of non-square aspect ratios)
- Added support for error tracking with Sentry
- Only show the "Connecting..." bar after two connection attempts 

### Bug Fixes
- Fixed link rendering not working in certain cases
- Fixed theme color issue with status bar on Android

## v1.2 Release

- Release Date: September 5, 2017 
- Server Versions Supported: Server v4.0+ is required, Self-Signed SSL Certificates are not supported

### Highlights 

#### AppConfig Support for EMM solutions
- Added [AppConfig](https://www.appconfig.org/) support, to make it easier to integrate with a variety of EMM solutions

#### Code block viewer
- Tap on a code block to open a viewer for easier reading 

### Improvements
- Updated formatting for markdown lists and code blocks
- Updated formatting for `in:` and `from:` search autocomplete 

### Emoji Picker for Emoji Reactions
- Added an emoji picker for selecting a reaction 

### Bug Fixes
- Fixed issue where if only LDAP and GitLab login were enabled, LDAP did not show up on the login page
- Fixed issue with 3 digit mention count UI in channel drawer

### Known Issues
- Using `+:emoji:` to react to a message is not yet supported 

## v1.1 Release

- Release Date: August 2017 
- Server Versions Supported: Server v3.10+ is required, Self-Signed SSL Certificates are not supported

### Highlights 

#### Search
- Search posts and tap to preview the result
- Click "Jump" to open the channel the search result is from 

#### Emoji Reactions
- View Emoji Reactions on a post

#### Group Messages
- Start Direct and Group Messages from the same screen

#### Improved Performance on Poor Connections
- Added auto-retry to automatically reattempt to get posts if the network connection is intermittent
- Added manual loading option if auto-retry fails to retrieve new posts

### Improvements
- Android: Added Big Text support for Android notifications, so they expand to show more details
- Added a Reset Cache option
- Improved "Jump to conversation" filter so it matches on nickname, full name, or username 
- Tapping on an @username mention opens the user's profile
- Disabled the send button while attachments upload
- Adjusted margins on icons and elsewhere to make spacing more consistent
- iOS URL scheme: mattermost:// links now open the new app
- About Mattermost page now includes a link to NOTICES.txt for platform and the mobile app
- Various UI improvements

### Bug Fixes
- Fixed an issue where sometimes an unmounted badge caused app to crash on start up 
- Group Direct Messages now show the correct member count 
- Hamburger icon does not break after swiping to close sidebar
- Fixed an issue with some image thumbnails appearing out of focus
- Uploading a file and then leaving the channel no longer shows the file in a perpetual loading state
- For private channels, the last member can no longer delete the channel if the EE server permissions do not allow it
- Error messages are now shown when SSO login fails
- Android: Leaving a channel now redirects to Town Square instead of the Town Square info page
- Fixed create new public channel screen shown twice when trying to create a channel
- Tapping on a post will no longer close the keyboard

## v1.0.1 Release 

- Release Date: July 20, 2017 
- Server Versions Supported: Server v3.8+ is required, Self-Signed SSL Certificates are not yet supported

### Bug Fixes
- Huawei devices can now load messages
- GitLab SSO now works if there is a trailing `/` in the server URL
- Unsupported server versions now show a prompt clarifying that a server upgrade is necessary

## v1.0 Release 

- Release Date: July 10, 2017 
- Server Versions Supported: Server v3.8+ is required, Self-Signed SSL Certificates are not supported

### Highlights 

#### Authentication (Requires v3.10+ [Mattermost server](https://github.com/mattermost/platform))
- GitLab login 

#### Offline Support
- Added offline support, so already loaded portions of the app are accessible without a connection
- Retry mechanism for posts sent while offline 
- See [FAQ](https://github.com/mattermost/mattermost-mobile#frequently-asked-questions) for information on how data is handled for deactivated users

#### Notifications (Requires v3.10+ [push proxy server](https://github.com/mattermost/mattermost-push-proxy)) 
- Notifications are cleared when read on another device
- Notification sent just before session expires to let people know login is required to continue receiving notifications

#### Channel and Team Sidebar
- Unreads section to easily access channels with new messages
- Search filter to jump to conversations quickly 
- Improved team switching design for better cross-team notifications 
- Added ability to join open teams on the server 

#### Posts
- Emojis now render
- Integration attachments now render 
- ~channel links now render 

#### Navigation
- Updated navigation to have smoother transitions 

### Known Issues
- [Android: Swipe to close in-app notifications does not work](https://mattermost.atlassian.net/browse/RN-45)
- Apps are not yet at feature parity for desktop, so features not mentioned in the changelog are not yet supported

### Contributors

Many thanks to all our contributors. In alphabetical order:
- asaadmahmood, cpanato, csduarte, enahum, hmhealey, jarredwitt, JeffSchering, jasonblais, lfbrock, omar-dev, rthill

## Beta Release

- Release Date: March 29, 2017
- Server Versions Supported: Server v3.7+ is required, Self-Signed SSL Certificates are not yet supported

Note: If you need an SSL certificate, consider using [Let's Encrypt](https://docs.mattermost.com/install/config-ssl-http2-nginx.html) instead of a self-signed one.

### Highlights

The Beta apps are a work in progress, supported features are listed below. You can become a beta tester by [downloading the Android app](https://play.google.com/store/apps/details?id=com.mattermost.react.native&hl=en) or [signing up to test iOS](https://mattermost-fastlane.herokuapp.com/). 

#### Authentication
- Email login
- LDAP/AD login
- Multi-factor authentication 
- Logout

#### Messaging
- View and send posts in the center channel
- Automatically load more posts in the center channel when scrolling
- View and send replies in thread view
- "New messages" line in center channel (app does not yet scroll to the line)
- Date separators 
- @mention autocomplete
- ~channel autocomplete
- "User is typing" message
- Edit and delete posts
- Flag/Unflag posts
- Basic markdown (lists, headers, bold, italics, links)

#### Notifications
- Push notifications
- In-app notifications when you receive a message in another channel while the app is open
- Clicking on a push notification takes you to the channel

#### User profiles
- Status indicators
- View profile information by clicking on someone's username or profile picture

#### Files
- File thumbnails for posts with attachments
- Upload up to 5 images
- Image previewer to view images when clicked on

#### Channels
- Channel drawer for selecting channels
- Bolded channel names for Unreads, and mention jewel for Mentions
- (iOS only) Unread posts above/below indicator
- Favorite channels (Section in sidebar, and ability to favorite/unfavorite from channel menu)
- Create new public or private channels
- Create new Direct Messages (Group Direct Messages are not yet supported) 
- View channel info (name, header, purpose) 
- Join public channels
- Leave channel
- Delete channel
- View people in a channel
- Add/remove people from a channel
- Loading screen when opening channels 

#### Settings
- Account Settings > Notifications page
- About Mattermost info dialog
- Report a problem link that opens an email for bug reports

#### Teams
- Switch between teams using "Team Selection" in the main menu (viewing which teams have notifications is not yet supported) 

### Contributors

Many thanks to all our contributors. In alphabetical order:
- csduarte, dmeza, enahum, hmhealey, it33, jarredwitt, jasonblais, lfbrock, mfpiccolo, saturninoabril, thomchop
