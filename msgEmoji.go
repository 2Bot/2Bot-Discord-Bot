package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
)

func msgEmoji(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	s.ChannelTyping(m.ChannelID)
	submatch := emojiRegex.FindStringSubmatch(msglist[0])
	if msglist[0] == "bigMoji" || len(submatch) != 0 || emojiFile(msglist[0]) != "" {
		//bigMoji
		if msglist[0] == "bigMoji" && len(msglist) > 1 {
			submatch := emojiRegex.FindStringSubmatch(msglist[1])
			if len(submatch) != 0 {
				emojiID := submatch[1]

				resp, err := http.Get(fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", emojiID))
				if err != nil {
					errorLog.Println("BM custom emoji err:", err.Error())
					return
				}
				defer resp.Body.Close()

				s.ChannelFileSend(m.ChannelID, "emoji.png", resp.Body)
				s.ChannelMessageDelete(m.ChannelID, m.ID)
			} else {
				emoji := emojiFile(msglist[1])
				if emoji != "" {
					file, err := os.Open(fmt.Sprintf("emoji/%s.png", emoji))
					if err != nil {
						errorLog.Println("BM in-built emoji err:", err.Error())
						return
					}
					defer file.Close()

					s.ChannelFileSend(m.ChannelID, "emoji.png", file)
					s.ChannelMessageDelete(m.ChannelID, m.ID)
				}
			}
			//not bigMoji
		} else if len(msglist) > 0 {
			if len(submatch) != 0 {
				emojiID := submatch[1]
				resp, err := http.Get(fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", emojiID))
				if err != nil {
					errorLog.Println("!BM custom emoji err:", err.Error())
					return
				}
				defer resp.Body.Close()

				s.ChannelFileSend(m.ChannelID, "emoji.png", resp.Body)
				s.ChannelMessageDelete(m.ChannelID, m.ID)
			} else {
				emoji := emojiFile(msglist[0])
				if emoji != "" {
					file, err := os.Open(fmt.Sprintf("emoji/%s.png", emoji))
					if err != nil {
						errorLog.Println("!BM in-built emoji err:", err.Error())
						return
					}
					defer file.Close()

					s.ChannelFileSend(m.ChannelID, "emoji.png", file)
					s.ChannelMessageDelete(m.ChannelID, m.ID)
				}
			}
		}
	}
	return
}
