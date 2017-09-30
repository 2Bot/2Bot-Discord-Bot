package main

import (
	"strings"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
	"io"
)

func msgYoutube(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	switch msglist[1] {
	case "play":
		addToQueue(s, m, msglist[2:])
	case "stop":
		stop(s, m, msglist[2:])
	}
}

func addToQueue(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println(err)
		return
	}

	srvr := sMap.Server[guild.ID]
	if srvr.VoiceInst.Done == nil {
		srvr.VoiceInst.Done = make(chan error)
	}

	// TODO error messages boolean logic end of func
	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			if (vs.ChannelID == srvr.VoiceInst.ChannelID) || !srvr.VoiceInst.Playing {
				vc, err := s.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
				if err != nil {
					errorLog.Println(err)
					return
				}
				if !strings.HasPrefix(msglist[0], "https://www.youtube.com/watch?v") {
					s.ChannelMessageSend(m.ChannelID, "Please make sure the URL starts with `https://www.youtube.com/watch?v`")
				}
				srvr.VoiceInst.Queue = append(srvr.VoiceInst.Queue, song{
					URL: msglist[0],
				})
				vid, err := ytdl.GetVideoInfo(srvr.VoiceInst.Queue[0].URL)
				if err != nil {
					errorLog.Println("1", err)
					return
				}

				s.ChannelMessageSend(m.ChannelID, "Added "+vid.Title+" to the queue!")

				if !srvr.VoiceInst.Playing {
					go play(s, m, sMap.Server[guild.ID], vc)
				}
				return
			}

			if vs.ChannelID != srvr.VoiceInst.ChannelID {
				s.ChannelMessageSend(m.ChannelID, "Already playing in a different voice channel :(")
			}
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Need to be in a voice channel!")
}

func play(s *discordgo.Session, m *discordgo.MessageCreate, srvr *server, vc *discordgo.VoiceConnection) {
	if len(srvr.VoiceInst.Queue) == 0 {
		vc.Disconnect()
		srvr.VoiceInst.ChannelID = ""
		srvr.VoiceInst.Playing = false
		s.ChannelMessageSend(m.ChannelID, "🔇 Done queue!")
		return
	}

	vid, err := ytdl.GetVideoInfo(srvr.VoiceInst.Queue[0].URL)
	if err != nil {
		errorLog.Println("1", err)
		return
	}

	format := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	videoURL, err := vid.GetDownloadURL(format)
	if err != nil {
		errorLog.Println("2", err)
		return
	}

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 64
	options.Application = "lowdelay"
	options.Volume = 200
	options.Threads = 4

	encSesh, err := dca.EncodeFile(videoURL.String(), options)
	if err != nil {
		errorLog.Println("3", err)
		return
	}
	defer encSesh.Cleanup()

	dca.NewStream(encSesh, vc, srvr.VoiceInst.Done)
	s.ChannelMessageSend(m.ChannelID, "🔊 Playing: "+vid.Title)

	srvr.VoiceInst.Playing = true
	srvr.VoiceInst.ChannelID = vc.ChannelID

	for {
		err = <-srvr.VoiceInst.Done
		if err != nil && err != io.EOF && err.Error() != "stop" {
			errorLog.Println("Music stream error", err)
		}
		break
	}

	srvr.VoiceInst.Queue = srvr.VoiceInst.Queue[1:]
	go play(s, m, srvr, vc)

	return
}

func stop(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	voiceInst := &(sMap.Server[guild.ID].VoiceInst)
	voiceInst.Done <- errors.New("stop")
}
