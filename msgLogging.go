package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func init() {
	newCommand("logging",
		discordgo.PermissionAdministrator|discordgo.PermissionManageServer,
		false, true, msgLogging).setHelp("Args: none\n\nToggles user presence logging.\n\nExample:\n`!owo logging`").add()
	newCommand("logChannel",
		discordgo.PermissionAdministrator|discordgo.PermissionManageServer,
		false, true, msgLogChannel).setHelp("Args: [channelID,channel tag]\n\nSets the log channel to the given channel.\nAdmin only.\n\nExample:\n`!owo logChannel 312292616089894924`\n`!owo logChannel #bot-channel`").add()
}

func msgLogChannel(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem setting the details :( Try again please~")
		errorLog.Println("log channel guild details error", err)
		return
	}

	if len(msglist) < 2 {
		return
	}

	var channelID string
	channelIDMatch := channelRegex.FindStringSubmatch(msglist[1])
	if len(channelIDMatch) != 2 {
		if len(msglist[1]) == 18 {
			channelID = msglist[1]
		} else {
			s.ChannelMessageSend(m.ChannelID, "Not a valid channel!")
			return
		}
	}

	var chanList []string
	for _, channel := range guild.Channels {
		chanList = append(chanList, channel.ID)
	}

	if !isIn(channelID, chanList) {
		s.ChannelMessageSend(m.ChannelID, "That channel isn't in this server <:2BThink:333694872802426880>")
		return
	}

	if guild, ok := sMap.Server[guild.ID]; ok && !guild.Kicked {
		guild.LogChannel = channelID
		saveServers()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Log channel changed to <#%s>", channelID))
	}
	return
}

func msgLogging(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem toggling logging :( Try again please~")
		errorLog.Println("logging guild details error", err)
		return
	}

	if guild, ok := sMap.Server[guild.ID]; ok && !guild.Kicked {
		guild.Log = !guild.Log
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Logging %t", guild.Log))
		saveServers()
	}
}
