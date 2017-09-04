package main

import (
	"io"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
)

func msgYoutube(s* discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	switch msglist[1] {
		case "play":
			play(s, m, msglist[2:])
		case "stop":
			stop(s, m, msglist[2:])
		case "playlist":
			playlist(s, m, msglist[2:])

	}
}

func play(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		fmt.Println(err)
		return
	}

	voiceInst := &(sMap.Server[guild.ID].VoiceInst)
	voiceInst.Done = make(chan error)

	vid, err := ytdl.GetVideoInfo(msglist[0])
	if err != nil {
	  fmt.Println(err)
	  return
	}

	format := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	videoURL, err := vid.GetDownloadURL(format)
	if err != nil {
		fmt.Println(err)
		return
	}

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			vc, err := s.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer vc.Disconnect()

			encSesh, err := dca.EncodeFile(videoURL.String(), options)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer encSesh.Cleanup()
			
			dca.NewStream(encSesh, vc, voiceInst.Done)
			for {
				select {
					case err := <- voiceInst.Done:
						if err != nil && err != io.EOF && err.Error() != "stop" {
							fmt.Println(err)
						}
						encSesh.Cleanup()
						vc.Disconnect()
						return
				}
			}
		}
	}
}

func stop(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		fmt.Println(err)
		return
	}

	voiceInst := &(sMap.Server[guild.ID].VoiceInst)
	voiceInst.Done <- errors.New("stop")
}

func playlist(s  *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch msglist[0]{
		case "create":
			sMap.Server[guild.ID].Playlists[msglist[1]] = []song{}
			s.ChannelMessageSend(m.ChannelID, "Created playlist "+msglist[1])
			return
			
	}
}