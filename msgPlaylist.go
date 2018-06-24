package main

import (
	"strconv"
	"strings"

	"github.com/Strum355/ytdl"
	"github.com/bwmarrin/discordgo"
)

func init() {
	newCommand("playlist", 0, false, msgPlaylist).setHelp("dab on em").add() //TODO
}

func msgPlaylist(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) < 2 {
		return
	}

	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		return
	}

	server := sMap.Server[guild.ID]
	if server.Playlists == nil {
		server.Playlists = make(map[string][]song)
	}

	switch msglist[1] {
	case "create":
		createPlaylist(s, m, msglist, server)
	case "delete":
		deletePlaylist(s, m, msglist, server)
	case "add":
		addToPlaylist(s, m, msglist, server)
	case "remove":
		removeFromPlaylist(s, m, msglist, server)
	}

	saveServers()
}

func createPlaylist(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string, server *server) {
	playlist := strings.Join(msglist[2:], " ")
	if _, ok := server.Playlists[playlist]; ok {
		s.ChannelMessageSend(m.ChannelID, "Playlist `"+playlist+"` already exists!")
		return
	}

	server.Playlists[playlist] = []song{}
	s.ChannelMessageSend(m.ChannelID, "Created playlist `"+playlist+"`")
	return
}

func deletePlaylist(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string, server *server) {
	if _, ok := server.Playlists[msglist[1]]; !ok {
		s.ChannelMessageSend(m.ChannelID, "Playlist `"+msglist[1]+"` doesn't exist!")
		return
	}
	delete(server.Playlists, msglist[1])
	s.ChannelMessageSend(m.ChannelID, "Playlist `"+msglist[1]+"` was deleted")
}

func addToPlaylist(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string, server *server) {
	if len(msglist) < 3 {
		return
	}

	if _, ok := server.Playlists[msglist[1]]; !ok {
		s.ChannelMessageSend(m.ChannelID, "Playlist `"+msglist[1]+"` doesn't exist!")
		return
	}

	for _, song := range server.Playlists[msglist[1]] {
		if song.URL == msglist[2] {
			s.ChannelMessageSend(m.ChannelID, "That song is already in the playlist!")
			return
		}
	}

	vid, err := ytdl.GetVideoInfo(msglist[2])
	if err != nil {
		log.Error("error getting YouTube video info", err)
		s.ChannelMessageSend(m.ChannelID, "There was an error adding the song to the playlist :( Check the command and try again")
		return
	}

	format := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	_, err = vid.GetDownloadURL(format)
	if err != nil {
		log.Error("error getting download URL", err)
		s.ChannelMessageSend(m.ChannelID, "There was an error adding the song to the playlist :( Check the command and try again")
		return
	}

	server.Playlists[msglist[1]] = append(server.Playlists[msglist[1]], song{
		URL:      msglist[2],
		Name:     vid.Title,
		Duration: vid.Duration,
	})

	s.ChannelMessageSend(m.ChannelID, vid.Title+" added to playlist `"+msglist[1]+"`")
}

func removeFromPlaylist(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string, server *server) {
	index, err := strconv.Atoi(msglist[2])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Please give the index of the song to delete~")
		return
	}

	server.Playlists[msglist[1]] = append(server.Playlists[msglist[1]][:index], server.Playlists[msglist[1]][index+1:]...)
	s.ChannelMessageSend(m.ChannelID, "Song removed from `"+msglist[1]+"`")
}
