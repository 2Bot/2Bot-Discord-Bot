package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	newCommand("r34", 0, false, false, msgRule34).setHelp("Args: [search]\n\nReturns a random image from rule34 for the given search term.\n\nExample:\n`!owo r34 lewds`").add()
}

func msgRule34(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) < 2 {
		return
	}

	channel, err := channelDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem getting some details :( Please try again!")
		return
	}

	guild, err := guildDetails("", channel.GuildID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem getting some details :( Please try again!")
		return
	}

	if !sMap.Server[guild.ID].Nsfw && (!strings.HasPrefix(channel.Name, "nsfw") && !channel.NSFW) {
		s.ChannelMessageSend(m.ChannelID, "NSFW is disabled on this server~")
		return
	}

	var r34 rule34
	var query string

	s.ChannelTyping(m.ChannelID)

	for _, word := range msglist[1:] {
		query += "+" + word
	}

	page, err := http.Get(fmt.Sprintf("https://rule34.xxx/index.php?page=dapi&s=post&q=index&tags=%s", query))
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "error getting data from Rule34 :(")
		errorLog.Println("R34 response err:", err)
		return
	}
	defer page.Body.Close()

	if page.StatusCode != 200 {
		s.ChannelMessageSend(m.ChannelID, "Rule34 didn't respond :(")
		return
	}

	if err = xml.NewDecoder(page.Body).Decode(&r34); err != nil {
		errorLog.Println("R34 xml unmarshal err:", err)
		return
	}

	if r34.PostCount == 0 {
		s.ChannelMessageSend(m.ChannelID, "No results ¯\\_(ツ)_/¯")
		return
	}

	url := r34.Posts[randRange(0, len(r34.Posts)-1)].URL

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s searched for `%s` \n%s", m.Author.Username, strings.Replace(query, "+", " ", -1), url))
}
