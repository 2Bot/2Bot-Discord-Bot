package main

import (
	"github.com/bwmarrin/discordgo"
)

func msgAvatar(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	s.ChannelTyping(m.ChannelID)
	if len(msglist) > 1 {
		submatch := userIDRegex.FindStringSubmatch(msglist[1])
		if len(submatch) != 0 {
			getAvatar(submatch[1], m, s)
		}
	} else {
		getAvatar(m.Author.ID, m, s)
	}
	return
}

func getAvatar(userID string, m *discordgo.MessageCreate, s *discordgo.Session) {
	user, err := s.User(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error finding the user :( Please try again")
		errorLog.Println("Avatar user struct: ", err)
		return
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Description: user.Username + "'s Avatar",

		Color: 0x000000,

		Image: &discordgo.MessageEmbedImage{
			URL: user.AvatarURL("2048"),
		},
	})
}
