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

	if guild, ok := sMap.Server[guild.ID]; ok && !guild.Kicked {
		guild.LogChannel = msglist[1]
		saveConfig()
		channel, err := s.Channel(msglist[1])
		if err != nil {
			errorLog.Println("Channel error", err.Error())
			channel = &discordgo.Channel{
				Name: msglist[1],
			}
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Log channel changed to %s", channel.Name))
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

	fmt.Println(msglist)
	if len(msglist) < 2 {
		return
	}

	if guild, ok := sMap.Server[guild.ID]; ok && !guild.Kicked {
		guild.Log = !guild.Log
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Logging? %t", guild.Log))
		saveConfig()
	}
}
