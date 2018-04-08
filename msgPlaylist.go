package main

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/rylio/ytdl"
)

func init() {
	newCommand("playlist", 0, false, false, msgPlaylist) //TODO
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

	if sMap.Server[guild.ID].Playlists == nil {
		sMap.Server[guild.ID].Playlists = make(map[string][]song)
	}

	switch msglist[0] {
	case "create":
		func() {
			if _, ok := server.Playlists[msglist[1]]; ok {
				s.ChannelMessageSend(m.ChannelID, "Playlist "+msglist[1]+" already exists!")
				return
			}

			server.Playlists[msglist[1]] = []song{}
			s.ChannelMessageSend(m.ChannelID, "Created playlist "+msglist[1])
			return
		}()
	case "delete":
		func() {
			if _, ok := server.Playlists[msglist[1]]; !ok {
				s.ChannelMessageSend(m.ChannelID, "Playlist "+msglist[1]+" doesn't exist!")
				return
			}
			delete(server.Playlists, msglist[1])
			s.ChannelMessageSend(m.ChannelID, "Playlist "+msglist[1]+" was deleted")
		}()
	case "add":
		func() {
			if len(msglist) < 3 {
				return
			}

			if _, ok := server.Playlists[msglist[1]]; !ok {
				s.ChannelMessageSend(m.ChannelID, "Playlist \""+msglist[1]+"\" doesn't exist!")
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

			s.ChannelMessageSend(m.ChannelID, vid.Title+" added to playlist "+msglist[1])
		}()
	case "remove":
		func() {
			index, err := strconv.Atoi(msglist[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Please give the index of the song to delete~")
				return
			}

			server.Playlists[msglist[1]] = append(server.Playlists[msglist[1]][:index], server.Playlists[msglist[1]][index+1:]...)
			s.ChannelMessageSend(m.ChannelID, "Song removed from "+msglist[1])
		}()
	}

	saveServers()
}
