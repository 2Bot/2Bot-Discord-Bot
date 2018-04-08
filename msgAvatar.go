package main

import (
	"github.com/bwmarrin/discordgo"
)

func init() {
	newCommand("avatar", 0, false, false, msgAvatar).setHelp("Args: [@user]\n\nReturns the given users avatar.\nIf no user ID is given, your own avatar is sent.\n\nExample:\n`!owo avatar @Strum355#2298`").add()
}

func msgAvatar(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) > 1 {
		getAvatar(m.Author.ID, m, s)
		return
	}

	submatch := userIDRegex.FindStringSubmatch(msglist[1])
	if len(submatch) != 0 {
		getAvatar(submatch[1], m, s)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "User not found :(")
}

func getAvatar(userID string, m *discordgo.MessageCreate, s *discordgo.Session) {
	user, err := userDetails(userID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error finding the user :( Please try again")
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
