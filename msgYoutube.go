package main

import (
	"io"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
)

func msgYoutube(s* discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	fmt.Println("Got song "+msglist[1])
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		fmt.Println(err)
		return
	}

	vid, err := ytdl.GetVideoInfo(msglist[1])
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

			encSesh, err := dca.EncodeFile(videoURL.String(), options)
			if err != nil {
				fmt.Println(err)
				return
			}
		
			done := make(chan error)
			dca.NewStream(encSesh, vc, done)
			for {
				select {
					case err := <- done:
						if err != nil && err != io.EOF {
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