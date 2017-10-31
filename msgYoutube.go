package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
)

func msgYoutube(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	switch msglist[1] {
	case "play":
		addToQueue(s, m, msglist[2:])
	case "stop":
		stop(s, m, msglist[2:])
	case "list":
		listQueue(s, m)
	case "pause":
		pauseQueue(s, m)
	case "unpause":
		unpauseQueue(s, m)
	}
}

func listQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println(err)
		return
	}

	val := sMap.Server[guild.ID]

	if len(val.VoiceInst.Queue) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No songs in queue!")
		return
	}

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
		p.Add(&discordgo.MessageEmbed{
			Title: fmt.Sprintf("Title: %s\nDuration: %s", song.Name, song.Duration),

			Image: &discordgo.MessageEmbedImage{
				URL: song.Image,
			},
		})
	}

	p.SetPageFooters()
	p = &dgwidgets.Paginator{
		Loop:                    true,
		ColourWhenDone:          0xff0000,
		DeleteReactionsWhenDone: true,
	}
	p.Widget.Timeout = time.Minute * 2
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

	srvr.VoiceInst.Mutex = &sync.Mutex{}
	srvr.VoiceInst.Mutex.Lock()

	srvr.VoiceInst.Done = make(chan error)

	if !strings.HasPrefix(msglist[0], "https://www.youtube.com/watch?v") {
		s.ChannelMessageSend(m.ChannelID, "Please make sure the URL starts with `https://www.youtube.com/watch?v`")
		srvr.VoiceInst.Mutex.Unlock()
		return
	}

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
					Image:    vid.GetThumbnailURL(ytdl.ThumbnailQualityMedium).String(),
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
		srvr.VoiceInst = voiceInst{
			ChannelID: "",
			Playing:   false,
		}
		s.ChannelMessageSend(m.ChannelID, "🔇 Done queue!")
		srvr.VoiceInst.Mutex.Unlock()
		return
	}

	vid, err := ytdl.GetVideoInfo(srvr.VoiceInst.Queue[0].URL)
	if err != nil {
		srvr.VoiceInst.Mutex.Unlock()
		errorLog.Println(err)
		return
	}

	srvr.VoiceInst.Mutex.Unlock()

	reader, writer := io.Pipe()
	defer reader.Close()

	formats := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)
	if len(formats) > 0 {
		go func() {
			defer writer.Close()
			if err := vid.Download(formats[0], writer); err != nil {
				fmt.Println(err)
				return
			}
		}()
	}

	encSesh, err := dca.EncodeMem(reader, &dca.EncodeOptions{
		RawOutput:        true,
		Bitrate:          96,
		Application:      "lowdelay",
		Volume:           200,
		Threads:          4,
		FrameDuration:    20,
		FrameRate:        48000,
		Channels:         2,
		VBR:              true,
		BufferedFrames:   100,
		PacketLoss:       1,
		CompressionLevel: 10,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer encSesh.Cleanup()

	srvr.VoiceInst.StreamingSession = dca.NewStream(encSesh, vc, srvr.VoiceInst.Done)
	
	s.ChannelMessageSend(m.ChannelID, "🔊 Playing: "+vid.Title)

	srvr.VoiceInst.Playing = true
	srvr.VoiceInst.ChannelID = vc.ChannelID

	for {
		err = <-srvr.VoiceInst.Done
		if err != nil && err != io.EOF && err.Error() != "stop" {
			errorLog.Println("Music stream error", err)
		} else if err.Error() == "stop" {
			vc.Disconnect()
			srvr.VoiceInst.Mutex.Lock()
			srvr.VoiceInst = voiceInst{
				Playing:   false,
				Queue:     []song{},
				ChannelID: "",
			}
			srvr.VoiceInst.Mutex.Unlock()
			s.ChannelMessageSend(m.ChannelID, "🔇 Stopped")
			return
		}
		break
	}

	srvr.VoiceInst.Mutex.Lock()
	srvr.VoiceInst.Queue = srvr.VoiceInst.Queue[1:]
	go play(s, m, srvr, vc)
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

func pauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.StreamingSession.SetPaused(true)
}

func unpauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.StreamingSession.SetPaused(false)
}
