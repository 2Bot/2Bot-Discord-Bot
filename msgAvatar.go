package main

import (
	"github.com/bwmarrin/discordgo"
	"net/http"
)

func msgAvatar(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	var userID string

	if len(msglist) > 1 {
		submatch := userIDRegex.FindStringSubmatch(msglist[1])
		if len(submatch) != 0 {
			userID = submatch[1]
		}

		user, err := s.User(userID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "User not found :/")
			return
		}

		resp, err := http.Get(discordgo.EndpointUserAvatar(user.ID, user.Avatar))
		if err != nil {
			errorLog.Println("Error getting user avatar", err.Error())
			return
		}
		defer resp.Body.Close()

		s.ChannelFileSend(m.ChannelID, user.Username+"'s_avatar.png", resp.Body)
	} else {
		userID = m.Author.ID
		user, err := s.User(userID)
		if err != nil {
			errorLog.Println("Avatar user struct", err.Error())
			return
		}

		resp, err := http.Get(discordgo.EndpointUserAvatar(m.Author.ID, m.Author.Avatar))
		if err != nil {
			errorLog.Println(err.Error())
			return
		}
		defer resp.Body.Close()

		s.ChannelFileSend(m.ChannelID, user.Username+"'s_avatar.png", resp.Body)
	}
	return
}
