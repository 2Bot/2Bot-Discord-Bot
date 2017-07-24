package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func msgEncode(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) < 3 {
		return
	}

	base := strings.ToLower(msglist[1])
	text := strings.Join(msglist[2:], " ")

	switch base {
	case "base64":
		s.ChannelTyping(m.ChannelID)
		output := base64.StdEncoding.EncodeToString([]byte(text))
		s.ChannelMessageSend(m.ChannelID, output)
	case "bcrypt":
		s.ChannelTyping(m.ChannelID)
		output, err := bcrypt.GenerateFromPassword([]byte(text), 14)
		if err != nil {
			errorLog.Println("Bcrypt err:", err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, string(output))
	case "md5":
		s.ChannelTyping(m.ChannelID)
		output := md5.Sum([]byte(text))
		s.ChannelMessageSend(m.ChannelID, hex.EncodeToString(output[:]))
	case "sha256":
		s.ChannelTyping(m.ChannelID)
		hash := sha256.Sum256([]byte(text))
		s.ChannelMessageSend(m.ChannelID, hex.EncodeToString(hash[:]))
	default:
		s.ChannelMessageSend(m.ChannelID, "Base not supported")
	}
	return
}
