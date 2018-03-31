package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func membPresChange(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	guild, ok := sMap.Server[m.GuildID]
	if !ok || (guild.Kicked || !guild.Log) {
		return
	}

	memberStruct, err := memberDetails(m.GuildID, m.User.ID, s)
	if err != nil {
		errorLog.Println("Error")
	}

	s.ChannelMessageSend(guild.LogChannel, fmt.Sprintf("`%s is now %s`", memberStruct.User, status[m.Status]))
}

func membJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	guild, ok := sMap.Server[m.GuildID]
	if !ok || guild.Kicked || len(guild.JoinMessage) != 3 {
		return
	}

	isBool, err := strconv.ParseBool(guild.JoinMessage[0])
	if err != nil {
		errorLog.Println("couldnt parse bool", err)
		return
	}

	if !isBool || guild.JoinMessage[1] == "" {
		return
	}

	guildDetails, err := guildDetails("", m.GuildID, s)
	if err != nil {
		errorLog.Println("error getting guild details")
		return
	}

	membStruct, err := s.User(m.User.ID)
	if err != nil {
		errorLog.Println(guildDetails.Name, m.GuildID, err)
		return
	}

	s.ChannelMessageSend(guild.JoinMessage[2], strings.Replace(guild.JoinMessage[1], "%s", membStruct.Mention(), -1))
}
