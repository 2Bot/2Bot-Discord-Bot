package main

import (
	"errors"
	"fmt"
	"io"
	_ "strings"
	"sync"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
)

func msgYoutube(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) == 1 {

	}
	switch msglist[1] {
	case "play":
		addToQueue(s, m, msglist[2:])
	case "stop":
		stopQueue(s, m, msglist[2:])
	case "list":
		listQueue(s, m)
	case "pause":
		pauseQueue(s, m)
	case "unpause":
		unpauseQueue(s, m)
	case "skip":
		skipSong(s, m)
	}
}

func listQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println(err)
		return
	}

	val, ok := sMap.Server[guild.ID]
	if !ok {
		s.ChannelMessageSend(m.ChannelID, "Error (ill change this sometime)")
		return
	}

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
	p.Loop = true
	p.ColourWhenDone = 0xff0000
	p.DeleteReactionsWhenDone = true
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

	srvr, ok := sMap.Server[guild.ID]
	if !ok {
		s.ChannelMessageSend(m.ChannelID, "Error (ill change this sometime as well)")
		return
	}

	if srvr.VoiceInst.Mutex == nil {
		srvr.VoiceInst.Mutex = &sync.Mutex{}

	}
	srvr.VoiceInst.Mutex.Lock()
	defer srvr.VoiceInst.Mutex.Unlock()
	if srvr.VoiceInst.Done == nil {
		srvr.VoiceInst.Done = make(chan error)
	}

	/* 	if !strings.HasPrefix(msglist[0], "https://www.youtube.com/watch?v") {
		s.ChannelMessageSend(m.ChannelID, "Please make sure the URL starts with `https://www.youtube.com/watch?v`")
		srvr.VoiceInst.Mutex.Unlock()
		return
	} */

	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			if (vs.ChannelID == srvr.VoiceInst.ChannelID) || !srvr.VoiceInst.Playing {
				vc, err := s.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
				if err != nil {
					errorLog.Println(err)
					return
				}

				vid, err := ytdl.GetVideoInfo(msglist[0])
				if err != nil {
					errorLog.Println("1", err)
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
				}
				return
			}

			/* 			if vs.ChannelID != srvr.VoiceInst.ChannelID {
				s.ChannelMessageSend(m.ChannelID, "Already playing in a different voice channel :(")
				return
			} */
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Need to be in a voice channel!")
}

func play(s *discordgo.Session, m *discordgo.MessageCreate, srvr *server, vc *discordgo.VoiceConnection) {
	if len(srvr.VoiceInst.Queue) == 0 {
		srvr.VoiceInst.Mutex.Lock()
		defer srvr.VoiceInst.Mutex.Unlock()
		srvr.youtubeCleanup()
		s.ChannelMessageSend(m.ChannelID, "🔇 Done queue!")
		return
	}

	srvr.VoiceInst.Mutex.Lock()

	vid, err := ytdl.GetVideoInfo(srvr.VoiceInst.Queue[0].URL)
	if err != nil {
		errorLog.Println(err)
		return
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	formats := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)
	if len(formats) > 0 {
		go func() {
			defer writer.Close()
			if err := vid.Download(formats[0], writer); err != nil {
				errorLog.Println(err)
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
		errorLog.Println(err)
		return
	}
	defer encSesh.Cleanup()

	srvr.VoiceInst.StreamingSession = dca.NewStream(encSesh, vc, srvr.VoiceInst.Done)

	s.ChannelMessageSend(m.ChannelID, "🔊 Playing: "+vid.Title)

	srvr.VoiceInst.Playing = true
	srvr.VoiceInst.ChannelID = vc.ChannelID

	srvr.VoiceInst.Mutex.Unlock()

Done:
	for {
		err := <-srvr.VoiceInst.Done
		srvr.VoiceInst.Mutex.Lock()
		switch {
		case err.Error() == "stop":
			vc.Disconnect()
			srvr.youtubeCleanup()
			srvr.VoiceInst.Mutex.Unlock()
			s.ChannelMessageSend(m.ChannelID, "🔇 Stopped")
			return
		case err.Error() == "skip":
			srvr.VoiceInst.Queue = srvr.VoiceInst.Queue[1:]
			srvr.VoiceInst.Mutex.Unlock()
			s.ChannelMessageSend(m.ChannelID, "⏩ Skipping")
			break Done
		case err != nil && err != io.EOF:
			vc.Disconnect()
			srvr.youtubeCleanup()
			srvr.VoiceInst.Mutex.Unlock()
			s.ChannelMessageSend(m.ChannelID, "There was an error streaming music :(")
			errorLog.Println("Music stream error", err)
			return
		}
	}

	go play(s, m, srvr, vc)
}

func (s *server) youtubeCleanup() {
	s.VoiceInst.ChannelID = ""
	s.VoiceInst.Playing = false
	s.VoiceInst.Queue = []song{}
	sMap.VoiceInsts--
}

func stopQueue(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	errorLog.Println(srvr.VoiceInst.Mutex)
	srvr.VoiceInst.Mutex.Lock()
	defer srvr.VoiceInst.Mutex.Unlock()
	srvr.VoiceInst.Done <- errors.New("stop")
	srvr.VoiceInst.Mutex.Unlock()
}

func pauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Mutex.Lock()
	defer srvr.VoiceInst.Mutex.Unlock()
	srvr.VoiceInst.StreamingSession.SetPaused(true)
}

func unpauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Mutex.Lock()
	srvr.VoiceInst.StreamingSession.SetPaused(false)
	srvr.VoiceInst.Mutex.Unlock()
}

func skipSong(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Mutex.Lock()
	srvr.VoiceInst.Done <- errors.New("skip")
	srvr.VoiceInst.Mutex.Unlock()
}
