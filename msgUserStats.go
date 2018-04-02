package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	newCommand("whois", 0, false, false, msgUserStats).setHelp("Args: [@user]\n\nSome info about the given user.\n\nExample:\n`!owo whois @Strum355#2298`").add()
}

func msgUserStats(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	channel, err := channelDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error getting the data :(")
		return
	}

	guild, err := guildDetails("", channel.GuildID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error getting the data :(")
		return
	}

	var userID string
	var nick string

	if len(msglist) > 1 {
		submatch := userIDRegex.FindStringSubmatch(msglist[1])
		if len(submatch) != 0 {
			userID = submatch[1]
		}
	} else {
		userID = m.Author.ID
	}

	user, err := s.User(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error getting the data :(")
		errorLog.Println("user struct err:", m.Content, err)
		return
	}
	
	memberStruct, err := memberDetails(guild.ID, userID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error getting the data :(")
		return
	}

	var roleNames []string

	for _, role := range memberStruct.Roles {
		for _, guildRole := range guild.Roles {
			if guildRole.ID == role {
				roleNames = append(roleNames, guildRole.Name)
			}
		}
	}

	if len(roleNames) == 0 {
		roleNames = append(roleNames, "None")
	}

	if memberStruct.Nick == "" {
		nick = "None"
	} else {
		nick = memberStruct.Nick
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color:       s.State.UserColor(userID, m.ChannelID),
		Description: fmt.Sprintf("%s is a loyal member of %s", user.Username, guild.Name),
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.Username,
			IconURL: discordgo.EndpointUserAvatar(userID, user.Avatar),
		},
		Footer: footer,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Username:", Value: user.Username, Inline: true},
			{Name: "Nickname:", Value: nick, Inline: true},
			{Name: "Joined Server:", Value: memberStruct.JoinedAt[:10], Inline: false},
			{Name: "Roles:", Value: strings.Join(roleNames, ", "), Inline: true},
			{Name: "ID Number:", Value: user.ID, Inline: true},
		},
	})
}
