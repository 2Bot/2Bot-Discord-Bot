package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
	"github.com/Necroforger/dgwidgets"
)

func msgYoutube(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	switch msglist[1] {
	case "play":
		addToQueue(s, m, msglist[2:])
	case "stop":
		stop(s, m, msglist[2:])
	case "list":
		queue(s, m)
	}
}

func queue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		fmt.Println(err)
		return
	}

	val := sMap.Server[guild.ID]
	p := dgwidgets.NewPaginator(s, m.ChannelID)
	p.Add(&discordgo.MessageEmbed{
		Title: guild.Name + "'s queue",

		Fields: func() (out []*discordgo.MessageEmbedField) {
			for i, song := range val.VoiceInst.Queue {
				out = append(out, &discordgo.MessageEmbedField{
					Name:  fmt.Sprintf("%d - %s", i, song.Name),
					Value: song.Duration.String(),
				})
			}
			return
		}(),
	})

	
	for _, song := range val.VoiceInst.Queue {
		p.Add(func() (out *discordgo.MessageEmbed) {
		out = &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Title: %s\nDuration: %s", song.Name, song.Duration),

			Image: &discordgo.MessageEmbedImage {
				URL: song.Image,
			},
		}
		return
		}())
	}

	p.SetPageFooters()
	p.Loop = true
	p.DeleteReactionsWhenDone = true
	p.Spawn()
}

func addToQueue(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) == 0 {
		return
	}

	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println(err)
		return
	}

	srvr := sMap.Server[guild.ID]

	if srvr.VoiceInst.Mutex == nil {
		srvr.VoiceInst.Mutex = &sync.Mutex{}
	}

	srvr.VoiceInst.Mutex.Lock()

	if srvr.VoiceInst.Done == nil {
		srvr.VoiceInst.Done = make(chan error)
	}

	if !strings.HasPrefix(msglist[0], "https://www.youtube.com/watch?v") {
		s.ChannelMessageSend(m.ChannelID, "Please make sure the URL starts with `https://www.youtube.com/watch?v`")
		srvr.VoiceInst.Mutex.Unlock()
		return
	}

	// TODO error messages boolean logic end of func
	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			if (vs.ChannelID == srvr.VoiceInst.ChannelID) || !srvr.VoiceInst.Playing {
				vc, err := s.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
				if err != nil {
					errorLog.Println(err)
					srvr.VoiceInst.Mutex.Unlock()
					return
				}

				vid, err := ytdl.GetVideoInfo(msglist[0])
				if err != nil {
					errorLog.Println("1", err)
					srvr.VoiceInst.Mutex.Unlock()
					return
				}

				srvr.VoiceInst.Queue = append(srvr.VoiceInst.Queue, song{
					URL:      msglist[0],
					Name:     vid.Title,
					Duration: vid.Duration,
					Image: vid.GetThumbnailURL(ytdl.ThumbnailQualityMedium).String(),
				})
				
				s.ChannelMessageSend(m.ChannelID, "Added "+vid.Title+" to the queue!")

				if !srvr.VoiceInst.Playing {
					go play(s, m, sMap.Server[guild.ID], vc)
				} else {
					srvr.VoiceInst.Mutex.Unlock()
				}
				return
			}

			if vs.ChannelID != srvr.VoiceInst.ChannelID {
				s.ChannelMessageSend(m.ChannelID, "Already playing in a different voice channel :(")
				srvr.VoiceInst.Mutex.Unlock()
				return
			}
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Need to be in a voice channel!")
	srvr.VoiceInst.Mutex.Unlock()
}

func play(s *discordgo.Session, m *discordgo.MessageCreate, srvr *server, vc *discordgo.VoiceConnection) {
	if len(srvr.VoiceInst.Queue) == 0 {
		vc.Disconnect()
		srvr.VoiceInst.ChannelID = ""
		srvr.VoiceInst.Playing = false
		s.ChannelMessageSend(m.ChannelID, "🔇 Done queue!")
		srvr.VoiceInst.Mutex.Unlock()
		return
	}

	vid, err := ytdl.GetVideoInfo(srvr.VoiceInst.Queue[0].URL)
	if err != nil {
		errorLog.Println("1", err)
		return
	}

	srvr.VoiceInst.Mutex.Unlock()

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
		fmt.Println(err)
		fmt.Println(encSesh.FFMPEGMessages())
		if err != nil && err != io.EOF && err.Error() != "stop" {
			errorLog.Println("Music stream error", err)

		} else if err.Error() == "stop" {
			vc.Disconnect()
			srvr.VoiceInst.Mutex.Lock()
			srvr.VoiceInst.Playing = false
			srvr.VoiceInst.Queue = []song{}
			srvr.VoiceInst.ChannelID = ""	
			srvr.VoiceInst.Mutex.Unlock()			
			s.ChannelMessageSend(m.ChannelID, "🔇 Stopped")
			return
		}
		break
	}

	srvr.VoiceInst.Mutex.Lock()
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

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Done <- errors.New("stop")
}