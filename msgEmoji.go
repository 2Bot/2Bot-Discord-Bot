package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

//Thanks to iopred
func emojiFile(s string) string {
	found := ""
	filename := ""
	for _, r := range s {
		if filename != "" {
			filename = fmt.Sprintf("%s-%x", filename, r)
		} else {
			filename = fmt.Sprintf("%x", r)
		}

		if _, err := os.Stat(fmt.Sprintf("emoji/%s.png", filename)); err == nil {
			found = filename
		} else if found != "" {
			return found
		}
	}
	return found
}

func msgEmoji(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	submatch := emojiRegex.FindStringSubmatch(msglist[0])

	/* if len(submatch) != 0 {
		emojiID := submatch[1]
	} */

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

				if m != nil {
					s.ChannelMessageDelete(m.ChannelID, m.ID)
				}
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

					if m != nil {
						s.ChannelMessageDelete(m.ChannelID, m.ID)
					}
				}
			}
			//not bigMoji
		} else if len(msglist) > 0 {
			if len(submatch) != 0 {
				emojiID := submatch[1]

				resp, err := http.Get(fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", emojiID))
				if err != nil {
					errorLog.Println("BM custom emoji err:", err.Error())
					return
				}
				defer resp.Body.Close()

				s.ChannelFileSend(m.ChannelID, "emoji.png", resp.Body)

				if m != nil {
					s.ChannelMessageDelete(m.ChannelID, m.ID)
				}
			} else {
				emoji := emojiFile(msglist[0])
				if emoji != "" {
					file, err := os.Open(fmt.Sprintf("emoji/%s.png", emoji))
					if err != nil {
						errorLog.Println("BM in-built emoji err:", err.Error())
						return
					}
					defer file.Close()

					s.ChannelFileSend(m.ChannelID, "emoji.png", file)

					if m != nil {
						s.ChannelMessageDelete(m.ChannelID, m.ID)
					}
				}
			}
		}
	}
	return
}
