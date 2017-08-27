package main

import (
	"net/http"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func msgAvatar(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
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
		errorLog.Println("Avatar user struct", err.Error())
		return
	}

	resp, err := http.Head(fmt.Sprintf("%s/%s.gif", discordgo.EndpointCDNAvatars+userID, user.Avatar))
	if err != nil {
		errorLog.Println(err)
		return
	}
	defer resp.Body.Close()

	imgURL := discordgo.EndpointUserAvatar(user.ID, user.Avatar)
	if resp.StatusCode == http.StatusOK {
		imgURL = fmt.Sprintf("%s/%s.gif", discordgo.EndpointCDNAvatars+userID, user.Avatar)
	}
			
	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Description: user.Username+"'s Avatar",
		
		Color: 0x000000,

		Image: &discordgo.MessageEmbedImage{
			URL: imgURL,
		},
	})
}