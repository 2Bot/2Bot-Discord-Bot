package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

// Thanks to iopred
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

func sendEmojiFromFile(s *discordgo.Session, m *discordgo.MessageCreate, e string) {
	emoji := emojiFile(e)
	if emoji == "" {
		return
	}

	file, err := os.Open(fmt.Sprintf("emoji/%s.png", emoji))
	if err != nil {
		errorLog.Println("BM in-built emoji err:", err)
		return
	}
	defer file.Close()

	s.ChannelFileSend(m.ChannelID, "emoji.png", file)

	deleteMessage(m, s)
}

func msgEmoji(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	submatch := emojiRegex.FindStringSubmatch(msglist[1])

	if len(submatch) == 0 {
		sendEmojiFromFile(s, m, msglist[1])
		return
	}

	var url string
	file := "emoji"

	switch submatch[1] {
	case "":
		url = fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", submatch[2])
		file += ".png"
	case "a":
		url = fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.gif", submatch[2])
		file += ".gif"
	}

	resp, err := http.Get(url)
	if err != nil {
		errorLog.Println("BM custom emoji err:", err)
		return
	}
	defer resp.Body.Close()

	s.ChannelFileSend(m.ChannelID, file, resp.Body)

	deleteMessage(m, s)
}
