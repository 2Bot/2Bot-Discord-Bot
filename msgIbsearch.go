package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	newCommand("ibsearch", 0, false, false, msgIbsearch).setHelp("Args: [search] | rating=[e,s,q] | format=[gif,png,jpg]\n\nReturns a random image from ibsearch for the given search term with the given filters applied.\n\n" +
		"Example:\n`!owo ibsearch lewds | rating=e | format=gif`")
}

func msgIbsearch(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("ibsearch guild details err", err)
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		errorLog.Println("Channel error", err)
		return
	}

	if !sMap.Server[guild.ID].Nsfw && (!strings.HasPrefix(channel.Name, "nsfw") && !channel.NSFW) {
		s.ChannelMessageSend(m.ChannelID, "NSFW is disabled on this server~")
		return
	}

	if len(msglist) < 2 {
		return
	}

	ibsearchStruct := ibStruct{}
	queryList := strings.Split(strings.Join(remove(msglist, 0), " "), "|")
	finalQuery := " "
	filters := []string{"rating", "format"}
	queries := []string{}
	URL, err := url.Parse("https://ibsearch.xxx")
	if err != nil {
		errorLog.Println("IBSearch query error", err)
		return
	}

	s.ChannelTyping(m.ChannelID)

	for i, item := range queryList {
		if strings.Contains(item, "=") {
			queries = append(queries, strings.TrimSpace(strings.Split(queryList[i], "=")[0]))
		}
	}

	for _, item1 := range filters {
		for i, item2 := range queries {
			if strings.Contains(item1, item2) {
				finalQuery += strings.Replace(queryList[i+1], " ", "", -1) + " "
			}
		}
	}

	//Assemble the URL
	URL.Path += "/api/v1/images.json"
	par := url.Values{}
	par.Add("q", strings.TrimSpace(queryList[0])+finalQuery+"random:")
	par.Add("limit", "1")
	//Public key that is for free, worst that can happen is that
	//i hit the ratelimit, but please dont do that to me
	par.Add("key", "2480CFA681A7A882CB33C0E4BA00A812C6F906A6")
	URL.RawQuery = par.Encode()

	page, err := http.Get(URL.String())
	if err != nil {
		errorLog.Println("Ibsearch http error", err)
	}
	if page.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "IBSearch didn't respond :(")
		return
	}
	defer page.Body.Close()

	body, err := ioutil.ReadAll(page.Body)
	if err != nil {
		errorLog.Println("IBSearch response body err:", err)
		return
	}

	err = json.Unmarshal([]byte(strings.TrimPrefix(strings.TrimSuffix(string(body), "]"), "[")), &ibsearchStruct)
	if err != nil {
		errorLog.Println("IBSearch json unmarshal err:", err)
		s.ChannelMessageSend(m.ChannelID, "No results ¯\\_(ツ)_/¯")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s searched for `%s` \nhttps://%s.ibsearch.xxx/%s", m.Author.Username, queryList[0], ibsearchStruct.Server, ibsearchStruct.Path))
	return
}
