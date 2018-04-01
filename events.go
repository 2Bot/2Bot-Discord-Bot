package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageCreateEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	guildDetails, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		return
	}

	var prefix string
	if val, ok := sMap.Server[guildDetails.ID]; ok {
		if prefix = val.Prefix; prefix == "" {
			prefix = c.Prefix
		}
	}

	parseCommand(s, m, func() string {
		if strings.HasPrefix(m.Content, c.Prefix) {
			return strings.TrimPrefix(m.Content, c.Prefix)
		}
		return strings.TrimPrefix(m.Content, prefix)
	}())

	return
}

func readyEvent(s *discordgo.Session, m *discordgo.Ready) {
	s.ChannelMessageSendEmbed(logChan, &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Info:", Value: "Received ready payload"},
		},
	})
	setBotGame(s)
}

func guildJoinEvent(s *discordgo.Session, m *discordgo.GuildCreate) {
	if m.Guild.Unavailable {
		return
	}

	guildDetails, err := guildDetails("", m.Guild.ID, s)
	if err != nil {
		return
	}

	user, err := s.User(guildDetails.OwnerID)
	if err != nil {
		errorLog.Println("error getting guild owner", err)
		user = &discordgo.User{
			Username:      "error",
			Discriminator: "error",
		}
	}

	embed := &discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: discordgo.EndpointGuildIcon(m.Guild.ID, m.Guild.Icon),
		},

		Footer: footer,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Name:", Value: m.Guild.Name, Inline: true},
			{Name: "User Count:", Value: strconv.Itoa(m.Guild.MemberCount), Inline: true},
			{Name: "Region:", Value: m.Guild.Region, Inline: true},
			{Name: "Channel Count:", Value: strconv.Itoa(len(m.Guild.Channels)), Inline: true},
			{Name: "ID:", Value: m.Guild.ID, Inline: true},
			{Name: "Owner:", Value: user.Username + "#" + user.Discriminator, Inline: true},
		},
	}

	if _, ok := sMap.Server[m.Guild.ID]; !ok {
		//if newly joined
		embed.Color = 65280
		s.ChannelMessageSendEmbed(logChan, embed)
		infoLog.Println("Joined server", m.Guild.ID, m.Guild.Name)

		sMap.Server[m.Guild.ID] = &server{
			LogChannel:  m.Guild.ID,
			Log:         false,
			Nsfw:        false,
			JoinMessage: [3]string{"false", "", ""},
		}
	} else if val := sMap.Server[m.Guild.ID]; val.Kicked == true {
		//If previously kicked and then readded
		embed.Color = 16751104
		s.ChannelMessageSendEmbed(logChan, embed)
		infoLog.Println("Rejoined server", m.Guild.ID, m.Guild.Name)
	}

	sMap.Server[m.Guild.ID].Kicked = false
	sMap.Mutex.Lock()
	defer sMap.Mutex.Unlock()
	sMap.Count++
	saveServers()
}

func guildKickedEvent(s *discordgo.Session, m *discordgo.GuildDelete) {
	if m.Unavailable {
		return
	}

	s.ChannelMessageSendEmbed(logChan, &discordgo.MessageEmbed{
		Color:  16711680,
		Footer: footer,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Name:", Value: m.Name, Inline: true},
			{Name: "ID:", Value: m.Guild.ID, Inline: true},
		},
	})

	infoLog.Println("Kicked from", m.Guild.ID, m.Name)

	sMap.Server[m.Guild.ID].Kicked = true
	sMap.Mutex.Lock()
	defer sMap.Mutex.Unlock()
	sMap.Count--
	saveServers()
}

func presenceChangeEvent(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	guild, ok := sMap.Server[m.GuildID]
	if !ok || (guild.Kicked || !guild.Log) {
		return
	}

	memberStruct, err := memberDetails(m.GuildID, m.User.ID, s)
	if err != nil {
		return
	}

	s.ChannelMessageSend(guild.LogChannel, fmt.Sprintf("`%s is now %s`", memberStruct.User, status[m.Status]))
}

func memberJoinEvent(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	guild, ok := sMap.Server[m.GuildID]
	if !ok || guild.Kicked || len(guild.JoinMessage) != 3 {
		return
	}

	isBool, err := strconv.ParseBool(guild.JoinMessage[0])
	if err != nil {
		errorLog.Println("couldnt parse bool", err)
		return
	}

	if !isBool || guild.JoinMessage[1] == "" {
		return
	}

	guildDetails, err := guildDetails("", m.GuildID, s)
	if err != nil {
		return
	}

	membStruct, err := s.User(m.User.ID)
	if err != nil {
		errorLog.Println(guildDetails.Name, m.GuildID, err)
		return
	}

	s.ChannelMessageSend(guild.JoinMessage[2], strings.Replace(guild.JoinMessage[1], "%s", membStruct.Mention(), -1))
}
