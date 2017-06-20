package main 

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

func msgUserStats(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	channelInGuild, err := s.State.Channel(m.ChannelID)
	if err != nil {
		log(true, "channelInGuild err:", m.Content, err.Error()) 
		return
	}

	guildDetails, err := s.State.Guild(channelInGuild.GuildID)
	if err != nil { 
		log(true, "guildDetails err:", m.Content, err.Error())
		return
	}	
	
	var userID string
	var nick string
	roleStruct := guildDetails.Roles

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
		log(true, "user struct err:", m.Content, err.Error())
		return 
	}
	memberStruct, err := s.State.Member(channelInGuild.GuildID, user.ID)
	if err != nil { 
		log(true, "memberStruct err:", m.Content, err.Error())
		return 
	}
	
	var roleNames []string

	for _, role := range memberStruct.Roles {
		for _, guildRole := range roleStruct {
			if guildRole.ID == role{
				roleNames = append(roleNames, guildRole.Name)
			}
		}
	}

	if memberStruct.Nick == "" {
		nick = "None"
	}else{
		nick = memberStruct.Nick
	}
	
	if len(roleNames) == 0 {
		roleNames = append(roleNames, "None")
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Color:       s.State.UserColor(userID, m.ChannelID),
			Description: fmt.Sprintf("%s is a loyal member of %s", user.Username, guildDetails.Name),
			Author: 	 &discordgo.MessageEmbedAuthor{
				Name:    	user.Username,
				IconURL: 	discordgo.EndpointUserAvatar(userID, user.Avatar),
			},
			Footer: 	 &discordgo.MessageEmbedFooter{
				Text: 	 	"Brought to you by 2Bot :)", 
			},
			Fields: 	 []*discordgo.MessageEmbedField {
							&discordgo.MessageEmbedField{Name: "Username:", Value: user.Username, Inline: true},
							&discordgo.MessageEmbedField{Name: "Nickname:", Value: nick, Inline: true},
							&discordgo.MessageEmbedField{Name: "Joined Server:", Value: memberStruct.JoinedAt[:10], Inline: false},
							&discordgo.MessageEmbedField{Name: "Roles:", Value: strings.Join(roleNames, ", "), Inline: true},
					//		&discordgo.MessageEmbedField{Name: "ID Number:", Value: user.ID, Inline: true},
						},
		})

	return
}