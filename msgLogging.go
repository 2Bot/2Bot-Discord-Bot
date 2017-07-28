package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func msgLogChannel(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem setting the details :( Try again please~")
		errorLog.Println("log channel guild details error", err.Error())
		return
	}

	if m.Author.ID != guild.OwnerID && m.Author.ID != noah {
		s.ChannelMessageSend(m.ChannelID, "Sorry, only the owner can do this!")
		return
	}

	if len(msglist) < 2 {
		return
	}

	channelID := channelRegex.FindStringSubmatch(msglist[1])
	if len(channelID) != 2 {
		s.ChannelMessageSend(m.ChannelID, "Not a valid channel!")
		return
	}

	var chanList []string
	for _, channel := range guild.Channels {
		chanList = append(chanList, channel.ID)
	}

	if !isIn(channelID[1], chanList) {
		s.ChannelMessageSend(m.ChannelID, "That channel isn't in this server <:2BThink:333694872802426880>")
		return
	}

	if guild, ok := sMap.Server[guild.ID]; ok && !guild.Kicked {
		guild.LogChannel = channelID[1]
		saveServers()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Log channel changed to %s", channelID[0]))
	}
	return
}

func msgLogging(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem toggling logging :( Try again please~")
		errorLog.Println("logging guild details error", err.Error())
		return
	}

	if m.Author.ID != guild.OwnerID && m.Author.ID != noah {
		s.ChannelMessageSend(m.ChannelID, "Sorry, only the owner can do this!")
		return
	}

	if guild, ok := sMap.Server[guild.ID]; ok && !guild.Kicked {
		guild.Log = !guild.Log
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Logging %t", guild.Log))
		saveServers()
	}
}
