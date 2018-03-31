package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/Strum355/ytdl"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

const (
	stdURL   = "https://www.youtube.com/watch"
	shortURL = "https://youtu.be/"
	embedURL = "https://www.youtube.com/embed/"
)

func init() {
	newCommand("yt", 0, false, false, msgYoutube).setHelp("Args: [play,stop] [url]\n\nWork In Progress!!! Play music from Youtube straight to your Discord Server!\n\n" +
		"Example 1: `!owo yt play https://www.youtube.com/watch?v=MvLdxtICOIY`\n" +
		"Example 2: `!owo yt stop`\n\nSubCommands:\nplay\nstop\nlist, queue, songs\npause\nresume, unpause\nskip, next").add()
}

func msgYoutube(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) == 1 {
		return
	}

	switch msglist[1] {
	case "play":
		addToQueue(s, m, msglist[2:])
	case "stop":
		stopQueue(s, m)
	case "list", "queue", "songs":
		listQueue(s, m)
	case "pause":
		pauseQueue(s, m)
	case "resume", "unpause":
		unpauseQueue(s, m)
	case "skip", "next":
		skipSong(s, m)
	}
}

func addToQueue(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) == 0 {
		return
	}

	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		errorLog.Println(err)
		return
	}

	srvr, ok := sMap.Server[guild.ID]
	if !ok {
		s.ChannelMessageSend(m.ChannelID, "An error occured that really shouldn't have happened...")
		errorLog.Println(guild.ID, "not in server map?")
		return
	}

	if srvr.VoiceInst == nil {
		srvr.newVoiceInstance()
	}

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()

	url := msglist[0]

	if !strings.HasPrefix(url, stdURL) && !strings.HasPrefix(url, shortURL) && !strings.HasPrefix(url, embedURL) {
		s.ChannelMessageSend(m.ChannelID, "Please make sure the URL is a valid YouTube URL. If I got this wrong, please let my creator know~")
		return
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			if vs.ChannelID == srvr.VoiceInst.ChannelID || !srvr.VoiceInst.Playing {
				vid, err := getVideoInfo(url, s, m)
				if err != nil {
					return
				}

				vc, err := s.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error joining voice channel")
					errorLog.Println("Error joining voice channel", err)
					return
				}

				srvr.addSong(song{
					URL:      url,
					Name:     vid.Title,
					Duration: vid.Duration,
					Image:    vid.GetThumbnailURL(ytdl.ThumbnailQualityMedium).String(),
				})

				s.ChannelMessageSend(m.ChannelID, "Added "+vid.Title+" to the queue!")

				if !srvr.VoiceInst.Playing {
					srvr.VoiceInst.VoiceCon = vc
					srvr.VoiceInst.Playing = true
					srvr.VoiceInst.ChannelID = vc.ChannelID
					go play(s, m, srvr, vc)
				}
				return
			}
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
	if srvr.queueLength() == 0 {
		srvr.youtubeCleanup()
		s.ChannelMessageSend(m.ChannelID, "🔇 Done queue!")
		return
	}

	srvr.VoiceInst.Lock()

	vid, err := getVideoInfo(srvr.nextSong().URL, s, m)
	if err != nil {
		srvr.VoiceInst.Unlock()
		return
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	formats := vid.Formats.Best(ytdl.FormatAudioBitrateKey)
	if len(formats) > 0 {
		go func() {
			defer writer.Close()
			if err := vid.Download(formats[0], writer); err != nil && err != io.ErrClosedPipe {
				s.ChannelMessageSend(m.ChannelID, xmark+" Error downloading the music")
				errorLog.Println("Youtube download error", err)
				srvr.VoiceInst.Done <- err
				return
			}
		}()
	}

	encSesh, err := dca.EncodeMem(reader, dca.StdEncodeOptions)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, xmark+" Error starting the stream")
		srvr.youtubeCleanup()
		srvr.VoiceInst.Unlock()
		errorLog.Println("Encode mem error", err)
		return
	}
	defer encSesh.Cleanup()

	srvr.VoiceInst.StreamingSession = dca.NewStream(encSesh, vc, srvr.VoiceInst.Done)

	s.ChannelMessageSend(m.ChannelID, "🔊 Playing: "+vid.Title)

	srvr.VoiceInst.Unlock()

Outer:
	for {
		err = <-srvr.VoiceInst.Done

		done, _ := srvr.VoiceInst.StreamingSession.Finished()

		switch {
		case err.Error() == "stop":
			srvr.youtubeCleanup()
			s.ChannelMessageSend(m.ChannelID, "🔇 Stopped")
			return
		case err.Error() == "skip":
			s.ChannelMessageSend(m.ChannelID, "⏩ Skipping")
			break Outer
		case !done && err != io.EOF:
			srvr.youtubeCleanup()
			s.ChannelMessageSend(m.ChannelID, "There was an error streaming music :(")
			errorLog.Println("Music stream error", err)
			return
		}
	}

	go play(s, m, srvr, vc)
}

func listQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		errorLog.Println(err)
		return
	}

	srvr, ok := sMap.Server[guild.ID]
	if !ok {
		s.ChannelMessageSend(m.ChannelID, "An error occured that really shouldn't have happened...")
		errorLog.Println(guild.ID, "not in server map?")
		return
	}

	if srvr.queueLength() == 0 {
		s.ChannelMessageSend(m.ChannelID, "No songs in queue!")
		return
	}

	p := dgwidgets.NewPaginator(s, m.ChannelID)
	p.Add(&discordgo.MessageEmbed{
		Title: guild.Name + "'s queue",

		Fields: func() (out []*discordgo.MessageEmbedField) {
			for i, song := range srvr.iterateQueue() {
				out = append(out, &discordgo.MessageEmbedField{
					Name:  fmt.Sprintf("%d - %s", i, song.Name),
					Value: song.Duration.String(),
				})
			}
			return
		}(),
	})

	for _, song := range srvr.iterateQueue() {
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

func stopQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, "", s)
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
	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error pausing video. Please try again.")
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⏸ Paused. To unpause, use the command `%sunpause`", func() string {
		if srvr.Prefix == "" {
			return c.Prefix
		}
		return srvr.Prefix
	}()))

	srvr.VoiceInst.StreamingSession.SetPaused(true)
}

func unpauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, "", s)
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
	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		errorLog.Println("Guild details error", err)
		return
	}

	srvr := sMap.Server[guild.ID]
	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	srvr.VoiceInst.Done <- errors.New("skip")
}

func (s *server) youtubeCleanup() {
	s.VoiceInst.Lock()
	defer s.VoiceInst.Unlock()
	s.VoiceInst.VoiceCon.Disconnect()
	s.newVoiceInstance()
	//sMap.VoiceInsts--
}
