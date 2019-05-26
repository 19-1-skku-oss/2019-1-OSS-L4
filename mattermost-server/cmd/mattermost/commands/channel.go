// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/app"
	"github.com/mattermost/mattermost-server/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var ChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Management of channels",
}

var ChannelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a channel",
	Long:  `Create a channel.`,
	Example: `  channel create --team myteam --name mynewchannel --display_name "My New Channel"
  channel create --team myteam --name mynewprivatechannel --display_name "My New Private Channel" --private`,
	RunE: createChannelCmdF,
}

var ChannelRenameCmd = &cobra.Command{
	Use:     "rename",
	Short:   "Rename a channel",
	Long:    `Rename a channel.`,
	Example: `"  channel rename myteam:mychannel newchannelname --display_name "New Display Name"`,
	Args:    cobra.MinimumNArgs(2),
	RunE:    renameChannelCmdF,
}

var RemoveChannelUsersCmd = &cobra.Command{
	Use:   "remove [channel] [users]",
	Short: "Remove users from channel",
	Long:  "Remove some users from channel",
	Example: `  channel remove myteam:mychannel user@example.com username
  channel remove myteam:mychannel --all-users`,
	RunE: removeChannelUsersCmdF,
}

var AddChannelUsersCmd = &cobra.Command{
	Use:     "add [channel] [users]",
	Short:   "Add users to channel",
	Long:    "Add some users to channel",
	Example: "  channel add myteam:mychannel user@example.com username",
	Args:    cobra.MinimumNArgs(2),
	RunE:    addChannelUsersCmdF,
}

var ArchiveChannelsCmd = &cobra.Command{
	Use:   "archive [channels]",
	Short: "Archive channels",
	Long: `Archive some channels.
Archive a channel along with all related information including posts from the database.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel archive myteam:mychannel",
	Args:    cobra.MinimumNArgs(1),
	RunE:    archiveChannelsCmdF,
}

var DeleteChannelsCmd = &cobra.Command{
	Use:   "delete [channels]",
	Short: "Delete channels",
	Long: `Permanently delete some channels.
Permanently deletes a channel along with all related information including posts from the database.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel delete myteam:mychannel",
	Args:    cobra.MinimumNArgs(1),
	RunE:    deleteChannelsCmdF,
}

var ListChannelsCmd = &cobra.Command{
	Use:   "list [teams]",
	Short: "List all channels on specified teams.",
	Long: `List all channels on specified teams.
Archived channels are appended with ' (archived)'.`,
	Example: "  channel list myteam",
	Args:    cobra.MinimumNArgs(1),
	RunE:    listChannelsCmdF,
}

var MoveChannelsCmd = &cobra.Command{
	Use:   "move [team] [channels] --username [user]",
	Short: "Moves channels to the specified team",
	Long: `Moves the provided channels to the specified team.
Validates that all users in the channel belong to the target team. Incoming/Outgoing webhooks are moved along with the channel.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel move newteam oldteam:mychannel --username myusername",
	Args:    cobra.MinimumNArgs(2),
	RunE:    moveChannelsCmdF,
}

var RestoreChannelsCmd = &cobra.Command{
	Use:   "restore [channels]",
	Short: "Restore some channels",
	Long: `Restore a previously deleted channel
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel restore myteam:mychannel",
	Args:    cobra.MinimumNArgs(1),
	RunE:    restoreChannelsCmdF,
}

var ModifyChannelCmd = &cobra.Command{
	Use:   "modify [channel] [flags] --username [user]",
	Short: "Modify a channel's public/private type",
	Long: `Change the public/private type of a channel.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel modify myteam:mychannel --private --username myusername",
	Args:    cobra.MinimumNArgs(1),
	RunE:    modifyChannelCmdF,
}

var SearchChannelCmd = &cobra.Command{
	Use:   "search [channel]\n  mattermost search --team [team] [channel]",
	Short: "Search a channel",
	Long: `Search a channel by channel name.
Channel can be specified by team. ie. --team myTeam myChannel or by team ID.`,
	Example: `  channel search myChannel
  channel search --team myTeam myChannel`,
	Args: cobra.ExactArgs(1),
	RunE: searchChannelCmdF,
}

func init() {
	ChannelCreateCmd.Flags().String("name", "", "Channel Name")
	ChannelCreateCmd.Flags().String("display_name", "", "Channel Display Name")
	ChannelCreateCmd.Flags().String("team", "", "Team name or ID")
	ChannelCreateCmd.Flags().String("header", "", "Channel header")
	ChannelCreateCmd.Flags().String("purpose", "", "Channel purpose")
	ChannelCreateCmd.Flags().Bool("private", false, "Create a private channel.")

	MoveChannelsCmd.Flags().String("username", "", "Required. Username who is moving the channel.")
	MoveChannelsCmd.Flags().Bool("remove-deactivated-users", false, "Automatically remove any deactivated users from the channel before moving it.")

	DeleteChannelsCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the channels.")

	ModifyChannelCmd.Flags().Bool("private", false, "Convert the channel to a private channel")
	ModifyChannelCmd.Flags().Bool("public", false, "Convert the channel to a public channel")
	ModifyChannelCmd.Flags().String("username", "", "Required. Username who changes the channel privacy.")

	ChannelRenameCmd.Flags().String("display_name", "", "Channel Display Name")
	SearchChannelCmd.Flags().String("team", "", "Team name or ID")

	RemoveChannelUsersCmd.Flags().Bool("all-users", false, "Remove all users from the indicated channel.")

	ChannelCmd.AddCommand(
		ChannelCreateCmd,
		RemoveChannelUsersCmd,
		AddChannelUsersCmd,
		ArchiveChannelsCmd,
		DeleteChannelsCmd,
		ListChannelsCmd,
		MoveChannelsCmd,
		RestoreChannelsCmd,
		ModifyChannelCmd,
		ChannelRenameCmd,
		SearchChannelCmd,
	)

	RootCmd.AddCommand(ChannelCmd)
}

func createChannelCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	name, errn := command.Flags().GetString("name")
	if errn != nil || name == "" {
		return errors.New("Name is required")
	}
	displayname, errdn := command.Flags().GetString("display_name")
	if errdn != nil || displayname == "" {
		return errors.New("Display Name is required")
	}
	teamArg, errteam := command.Flags().GetString("team")
	if errteam != nil || teamArg == "" {
		return errors.New("Team is required")
	}
	header, _ := command.Flags().GetString("header")
	purpose, _ := command.Flags().GetString("purpose")
	useprivate, _ := command.Flags().GetBool("private")

	channelType := model.CHANNEL_OPEN
	if useprivate {
		channelType = model.CHANNEL_PRIVATE
	}

	team := getTeamFromTeamArg(a, teamArg)
	if team == nil {
		return errors.New("Unable to find team: " + teamArg)
	}

	channel := &model.Channel{
		TeamId:      team.Id,
		Name:        name,
		DisplayName: displayname,
		Header:      header,
		Purpose:     purpose,
		Type:        channelType,
		CreatorId:   "",
	}

	createdChannel, errCreatedChannel := a.CreateChannel(channel, false)
	if errCreatedChannel != nil {
		return errCreatedChannel
	}

	CommandPrettyPrintln("Id: " + createdChannel.Id)
	CommandPrettyPrintln("Name: " + createdChannel.Name)
	CommandPrettyPrintln("Display Name: " + createdChannel.DisplayName)
	return nil
}

func removeChannelUsersCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	allUsers, _ := command.Flags().GetBool("all-users")

	if allUsers && len(args) != 1 {
		return errors.New("individual users must not be specified in conjunction with the --all-users flag")
	}

	if !allUsers && len(args) < 2 {
		return errors.New("you must specify some users to remove from the channel, or use the --all-users flag to remove them all")
	}

	channel := getChannelFromChannelArg(a, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	if allUsers {
		removeAllUsersFromChannel(a, channel)
	} else {
		users := getUsersFromUserArgs(a, args[1:])
		for i, user := range users {
			removeUserFromChannel(a, channel, user, args[i+1])
		}
	}

	return nil
}

func removeUserFromChannel(a *app.App, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		CommandPrintErrorln("Can't find user '" + userArg + "'")
		return
	}
	if err := a.RemoveUserFromChannel(user.Id, "", channel); err != nil {
		CommandPrintErrorln("Unable to remove '" + userArg + "' from " + channel.Name + ". Error: " + err.Error())
	}
}

func removeAllUsersFromChannel(a *app.App, channel *model.Channel) {
	if result := <-a.Srv.Store.Channel().PermanentDeleteMembersByChannel(channel.Id); result.Err != nil {
		CommandPrintErrorln("Unable to remove all users from " + channel.Name + ". Error: " + result.Err.Error())
	}
}

func addChannelUsersCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	channel := getChannelFromChannelArg(a, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	users := getUsersFromUserArgs(a, args[1:])
	for i, user := range users {
		addUserToChannel(a, channel, user, args[i+1])
	}

	return nil
}

func addUserToChannel(a *app.App, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		CommandPrintErrorln("Can't find user '" + userArg + "'")
		return
	}
	if _, err := a.AddUserToChannel(user, channel); err != nil {
		CommandPrintErrorln("Unable to add '" + userArg + "' from " + channel.Name + ". Error: " + err.Error())
	}
}

func archiveChannelsCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	channels := getChannelsFromChannelArgs(a, args)
	for i, channel := range channels {
		if channel == nil {
			CommandPrintErrorln("Unable to find channel '" + args[i] + "'")
			continue
		}
		if err := a.Srv.Store.Channel().Delete(channel.Id, model.GetMillis()); err != nil {
			CommandPrintErrorln("Unable to archive channel '" + channel.Name + "' error: " + err.Error())
		}
	}

	return nil
}

func deleteChannelsCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	confirmFlag, _ := command.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		CommandPrettyPrintln("Are you sure you want to delete the channels specified?  All data will be permanently deleted? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals.")
		}
	}

	channels := getChannelsFromChannelArgs(a, args)
	for i, channel := range channels {
		if channel == nil {
			CommandPrintErrorln("Unable to find channel '" + args[i] + "'")
			continue
		}
		if err := deleteChannel(a, channel); err != nil {
			CommandPrintErrorln("Unable to delete channel '" + channel.Name + "' error: " + err.Error())
		} else {
			CommandPrettyPrintln("Deleted channel '" + channel.Name + "'")
		}
	}

	return nil
}

func deleteChannel(a *app.App, channel *model.Channel) *model.AppError {
	return a.PermanentDeleteChannel(channel)
}

func moveChannelsCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	team := getTeamFromTeamArg(a, args[0])
	if team == nil {
		return errors.New("Unable to find destination team '" + args[0] + "'")
	}

	username, erru := command.Flags().GetString("username")
	if erru != nil || username == "" {
		return errors.New("Username is required.")
	}
	user := getUserFromUserArg(a, username)

	removeDeactivatedMembers, _ := command.Flags().GetBool("remove-deactivated-users")

	channels := getChannelsFromChannelArgs(a, args[1:])
	for i, channel := range channels {
		if channel == nil {
			CommandPrintErrorln("Unable to find channel '" + args[i+1] + "'")
			continue
		}
		originTeamID := channel.TeamId
		if err := moveChannel(a, team, channel, user, removeDeactivatedMembers); err != nil {
			CommandPrintErrorln("Unable to move channel '" + channel.Name + "' error: " + err.Error())
		} else {
			CommandPrettyPrintln("Moved channel '" + channel.Name + "' to " + team.Name + "(" + team.Id + ") from " + originTeamID + ".")
		}
	}

	return nil
}

func moveChannel(a *app.App, team *model.Team, channel *model.Channel, user *model.User, removeDeactivatedMembers bool) *model.AppError {
	oldTeamId := channel.TeamId

	if err := a.MoveChannel(team, channel, user, removeDeactivatedMembers); err != nil {
		return err
	}

	if incomingWebhooks, err := a.GetIncomingWebhooksForTeamPage(oldTeamId, 0, 10000000); err != nil {
		return err
	} else {
		for _, webhook := range incomingWebhooks {
			if webhook.ChannelId == channel.Id {
				webhook.TeamId = team.Id
				if _, err := a.Srv.Store.Webhook().UpdateIncoming(webhook); err != nil {
					CommandPrintErrorln("Failed to move incoming webhook '" + webhook.Id + "' to new team.")
				}
			}
		}
	}

	if outgoingWebhooks, err := a.GetOutgoingWebhooksForTeamPage(oldTeamId, 0, 10000000); err != nil {
		return err
	} else {
		for _, webhook := range outgoingWebhooks {
			if webhook.ChannelId == channel.Id {
				webhook.TeamId = team.Id
				if _, err := a.Srv.Store.Webhook().UpdateOutgoing(webhook); err != nil {
					CommandPrintErrorln("Failed to move outgoing webhook '" + webhook.Id + "' to new team.")
				}
			}
		}
	}

	return nil
}

func listChannelsCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	teams := getTeamsFromTeamArgs(a, args)
	for i, team := range teams {
		if team == nil {
			CommandPrintErrorln("Unable to find team '" + args[i] + "'")
			continue
		}
		if result := <-a.Srv.Store.Channel().GetAll(team.Id); result.Err != nil {
			CommandPrintErrorln("Unable to list channels for '" + args[i] + "'")
		} else {
			channels := result.Data.([]*model.Channel)

			for _, channel := range channels {
				if channel.DeleteAt > 0 {
					CommandPrettyPrintln(channel.Name + " (archived)")
				} else {
					CommandPrettyPrintln(channel.Name)
				}
			}
		}
	}

	return nil
}

func restoreChannelsCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	channels := getChannelsFromChannelArgs(a, args)
	for i, channel := range channels {
		if channel == nil {
			CommandPrintErrorln("Unable to find channel '" + args[i] + "'")
			continue
		}
		if err := a.Srv.Store.Channel().SetDeleteAt(channel.Id, 0, model.GetMillis()); err != nil {
			CommandPrintErrorln("Unable to restore channel '" + args[i] + "'")
		}
	}

	return nil
}

func modifyChannelCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	username, erru := command.Flags().GetString("username")
	if erru != nil || username == "" {
		return errors.New("Username is required.")
	}

	public, _ := command.Flags().GetBool("public")
	private, _ := command.Flags().GetBool("private")

	if public == private {
		return errors.New("You must specify only one of --public or --private")
	}

	channel := getChannelFromChannelArg(a, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	if !(channel.Type == model.CHANNEL_OPEN || channel.Type == model.CHANNEL_PRIVATE) {
		return errors.New("You can only change the type of public/private channels.")
	}

	channel.Type = model.CHANNEL_OPEN
	if private {
		channel.Type = model.CHANNEL_PRIVATE
	}

	user := getUserFromUserArg(a, username)
	if _, err := a.UpdateChannelPrivacy(channel, user); err != nil {
		return errors.Wrapf(err, "Failed to update channel ('%s') privacy", args[0])
	}

	return nil
}

func renameChannelCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	var newDisplayName, newChannelName string
	if err != nil {
		return err
	}
	defer a.Shutdown()

	channel := getChannelFromChannelArg(a, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	newChannelName = args[1]
	newDisplayName, errdn := command.Flags().GetString("display_name")
	if errdn != nil {
		return errdn
	}

	_, errch := a.RenameChannel(channel, newChannelName, newDisplayName)
	if errch != nil {
		return errors.Wrapf(errch, "Error in updating channel from %s to %s", channel.Name, newChannelName)
	}

	return nil
}

func searchChannelCmdF(command *cobra.Command, args []string) error {

	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return errors.Wrap(err, "failed to InitDBCommandContextCobra")
	}
	defer a.Shutdown()

	var channel *model.Channel

	if teamArg, _ := command.Flags().GetString("team"); teamArg != "" {
		team := getTeamFromTeamArg(a, teamArg)
		if team == nil {
			CommandPrettyPrintln(fmt.Sprintf("Team %s is not found", teamArg))
			return nil
		}

		var aErr *model.AppError
		channel, aErr = a.GetChannelByName(args[0], team.Id, true)
		if aErr != nil || channel == nil {
			CommandPrettyPrintln(fmt.Sprintf("Channel %s is not found in team %s", args[0], teamArg))
			return nil
		}
	} else {
		teams, aErr := a.GetAllTeams()
		if aErr != nil {
			return errors.Wrap(err, "failed to GetAllTeams")
		}

		for _, team := range teams {
			channel, _ = a.GetChannelByName(args[0], team.Id, true)
			if channel != nil && channel.Name == args[0] {
				break
			}
		}

		if channel == nil {
			CommandPrettyPrintln(fmt.Sprintf("Channel %s is not found in any team", args[0]))
			return nil
		}
	}

	if channel.DeleteAt > 0 {
		CommandPrettyPrintln(fmt.Sprintf(`Channel Name :%s, Display Name :%s, Channel ID :%s (archived)`, channel.Name, channel.DisplayName, channel.Id))
	} else {
		CommandPrettyPrintln(fmt.Sprintf(`Channel Name :%s, Display Name :%s, Channel ID :%s`, channel.Name, channel.DisplayName, channel.Id))
	}
	return nil
}
