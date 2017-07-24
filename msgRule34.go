package main

import (
	"encoding/xml"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"net/http"
	"strings"
)

func msgRule34(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem getting some details :( Please try again!")
		errorLog.Println("rule34 guild details error", err.Error())
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		errorLog.Println("Channel error", err.Error())
		return
	}

	if !sMap.Server[guild.ID].Nsfw && !strings.HasPrefix(channel.Name, "nsfw") {
		s.ChannelMessageSend(m.ChannelID, "NSFW is disabled on this server~")
		return
	}

	if len(msglist) < 2 {
		return
	}

	var r34 = rule34{}
	var query string

	s.ChannelTyping(m.ChannelID)

	for _, word := range msglist[1:] {
		query += "+" + word
	}
	page, err := http.Get(fmt.Sprintf("https://rule34.xxx/index.php?page=dapi&s=post&q=index&tags=%s", query))
	if err != nil {
		errorLog.Println("R34 response err:", err.Error())
		return
	}
	if page.StatusCode != 200 {
		s.ChannelMessageSend(m.ChannelID, "Rule34 didn't respond :(")
		return
	}
	defer page.Body.Close()

	body, err := ioutil.ReadAll(page.Body)
	if err != nil {
		errorLog.Println("R34 response body err:", err.Error())
		return
	}

	err = xml.Unmarshal(body, &r34)
	if err != nil {
		errorLog.Println("R34 xml unmarshal err:", err.Error())
		return
	}

	var url string
	if r34.PostCount == 0 {
		s.ChannelMessageSend(m.ChannelID, "No results ¯\\_(ツ)_/¯")
	} else {
		url = "https:" + r34.Posts[randRange(0, len(r34.Posts)-1)].URL
		resp, err := http.Get(url)
		if err != nil {
			errorLog.Println("R34 image response err:", err.Error())
			return
		}
		defer resp.Body.Close()

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s searched for `%s` \n%s", m.Author.Username, strings.Replace(query, "+", " ", -1), url))
	}

	return
}
