package main

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
)

func init() {
	newCommand("yt", 0, false, false, msgYoutube).setHelp("Args: [play,stop] [url]\n\nWork In Progress!!! Play music from Youtube straight to your Discord Server!\n\n" +
		"Example 1: `!owo yt play https://www.youtube.com/watch?v=MvLdxtICOIY`\n" +
		"Example 2: `!owo yt stop`").add()
}

func msgYoutube(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) == 1 {
		return
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
		srvr.VoiceInst.Mutex = new(sync.Mutex)
	}

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
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
			if vs.ChannelID == srvr.VoiceInst.ChannelID || !srvr.VoiceInst.Playing {
				vc, err := s.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error joining voice channel")
					errorLog.Println("Error joining voice channel", err)
					return
				}

				vid, err := getVideoInfo(msglist[0], s, m)
				if err != nil {
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

func getVideoInfo(url string, s *discordgo.Session, m *discordgo.MessageCreate) (*ytdl.VideoInfo, error) {
	vid, err := ytdl.GetVideoInfo(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting video info")
		errorLog.Println("Error getting video info", err)
		return nil, err
	}
	return vid, nil
}

func play(s *discordgo.Session, m *discordgo.MessageCreate, srvr *server, vc *discordgo.VoiceConnection) {
	if len(srvr.VoiceInst.Queue) == 0 {
		srvr.VoiceInst.Lock()
		defer srvr.VoiceInst.Unlock()
		srvr.youtubeCleanup()
		s.ChannelMessageSend(m.ChannelID, "🔇 Done queue!")
		return
	}

	srvr.VoiceInst.Lock()

	vid, err := getVideoInfo(srvr.VoiceInst.Queue[0].URL, s, m)
	if err != nil {
		srvr.VoiceInst.Unlock()
		return
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	formats := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)
	if len(formats) > 0 {
		go func() {
			defer writer.Close()
			//Do i have to send an error down the `down` channel here? investigate
			if err := vid.Download(formats[0], writer); err != nil && err != io.ErrClosedPipe {
				s.ChannelMessageSend(m.ChannelID, xmark+" Error downloading the music")
				errorLog.Println("Youtube download error", err)
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
		s.ChannelMessageSend(m.ChannelID, xmark+" Error starting the stream")
		srvr.youtubeCleanup()
		vc.Disconnect()
		srvr.VoiceInst.Unlock()
		errorLog.Println("Encode mem error", err)
		return
	}
	defer encSesh.Cleanup()

	srvr.VoiceInst.StreamingSession = dca.NewStream(encSesh, vc, srvr.VoiceInst.Done)

	s.ChannelMessageSend(m.ChannelID, "🔊 Playing: "+vid.Title)

	srvr.VoiceInst.Playing = true
	srvr.VoiceInst.ChannelID = vc.ChannelID

	srvr.VoiceInst.Unlock()

Done:
	for {
		err := <-srvr.VoiceInst.Done
		srvr.VoiceInst.Lock()
		defer srvr.VoiceInst.Unlock()
		switch {
		case err.Error() == "stop":
			vc.Disconnect()
			srvr.youtubeCleanup()
			s.ChannelMessageSend(m.ChannelID, "🔇 Stopped")
			return
		case err.Error() == "skip":
			srvr.VoiceInst.Queue = srvr.VoiceInst.Queue[1:]
			s.ChannelMessageSend(m.ChannelID, "⏩ Skipping")
			break Done
		case err != nil && err != io.EOF:
			vc.Disconnect()
			srvr.youtubeCleanup()
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
		s.ChannelMessageSend(m.ChannelID, "Error stopping queue. Please try again.")
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	srvr.VoiceInst.Done <- errors.New("stop")
}

func pauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error pausing video. Please try again.")
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⏸ Paused. To unpause, use the command %s unpause", func() string {
		if srvr.Prefix == "" {
			return c.Prefix
		}
		return srvr.Prefix
	}()))

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	srvr.VoiceInst.StreamingSession.SetPaused(true)
}

func unpauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	srvr.VoiceInst.StreamingSession.SetPaused(false)
}

func skipSong(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	srvr.VoiceInst.Done <- errors.New("skip")
}
