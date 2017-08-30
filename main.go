/*
	Copyright (C) 2017  Noah Santschi-Cooney

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	happyEmoji string = "https://cdn.discordapp.com/emojis/332968429210435585.png"
	thinkEmoji string = "https://cdn.discordapp.com/emojis/333694872802426880.png"
	reviewChan string = "334092230845267988"
	noah       string = "149612775587446784"
	logChan    string = "312352242504040448"
)

var (
	c            = &config{}
	u            = &users{}
	q            = &imageQueue{}
	sMap         = &servers{}
	errorLog     *log.Logger
	infoLog      *log.Logger
	logF         *os.File
	lastReboot   string
	token        string
	emojiRegex   = regexp.MustCompile("<:.*?:(.*?)>")
	userIDRegex  = regexp.MustCompile("<@!?([0-9]{18})>")
	channelRegex = regexp.MustCompile("<#([0-9]{18})>")
	status       = map[discordgo.Status]string{"dnd": "busy", "online": "online", "idle": "idle", "offline": "offline"}
	//Discord Bots, cool kidz only, social experiment, discord go
	blacklist    = []string{"110373943822540800", "272873324705742848", "244133074328092673", "118456055842734083"}
	errEmptyFile = errors.New("file is empty")
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	timeNow := time.Now()
	lastReboot = timeNow.Format(time.RFC1123)[:22]
}

func main() {
	loadLog()
	defer logF.Close()

	log.SetOutput(logF)

	infoLog = log.New(logF, "INFO:  ", log.Ldate|log.Ltime)
	errorLog = log.New(logF, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	loadConfig()
	loadUsers()
	loadQueue()
	loadServers()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}
	defer dg.Close()

	//Register handlers
	dg.AddHandler(messageCreate)
	dg.AddHandler(membPresChange)
	dg.AddHandler(kicked)
	dg.AddHandler(membJoin)

	setInitialGame(dg)

	go setQueuedImageHandlers(dg)
	if !c.InDev {
		dg.AddHandler(joined)
		go dailyJobs(dg)
	}

	fmt.Fprintln(logF, "")
	infoLog.Println(`/*********BOT RESTARTED*********\`)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func loadLog() *os.File {
	var err error
	logF, err = os.OpenFile("log.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return logF
}

func dailyJobs(s *discordgo.Session) {
	for {
		postServerCount()
		setInitialGame(s)
		time.Sleep(time.Hour * 24)
	}
}

func postServerCount() {
	url := "https://bots.discord.pw/api/bots/301819949683572738/stats"

	sCount := activeServerCount()
	jsonStr := []byte(`{"server_count"` + strconv.Itoa(sCount) + `:}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	req.Header.Set("Authorization", c.DiscordPWKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errorLog.Println("bots.discord.pw error", err)
		return
	}

	infoLog.Println("POSTed " + strconv.Itoa(sCount) + " to bots.discord.pw. Resp: "+strconv.Itoa(resp.StatusCode)+" "+func() string {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return ""
		}
		return string(bytes)
	}())
}

func activeServerCount() (sCount int) {
	for _, g := range sMap.Server {
		if !g.Kicked {
			sCount++
		}
	}
	return
}

func setInitialGame(s *discordgo.Session) {
	err := s.UpdateStatus(0, c.Game)
	if err != nil {
		errorLog.Println("Update status err:", err)
		return
	}
	infoLog.Println("set initial game to ", c.Game)
	return
}

func loadConfig() error {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		errorLog.Println("Config open err", err)
		return err
	}

	if len(file) < 1 {
		infoLog.Println("config.json is empty")
		return errEmptyFile
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		errorLog.Println("Config unmarshal err", err)
		return err
	}

	return nil
}

func saveConfig() {
	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		errorLog.Println("Config marshall err:", err)
		return
	}

	err = ioutil.WriteFile("config.json", out, 0755)
	if err != nil {
		errorLog.Println("Save config err:", err)
	}
	return
}

func loadServers() error {
	sMap.Server = make(map[string]*server)
	file, err := ioutil.ReadFile("servers.json")
	if err != nil {
		fmt.Println(true, "Servers open err", err)
		return err
	}

	if len(file) < 1 {
		infoLog.Println("servers.json is empty")
		return errEmptyFile
	}

	err = json.Unmarshal(file, sMap)
	if err != nil {
		errorLog.Println("Servers unmarshal err", err)
		return err
	}

	for gID, guild := range sMap.Server {
		if guild.LogChannel == "" {
			guild.LogChannel = gID
			saveServers()
		}
	}

	return nil
}

func saveServers() {
	out, err := json.MarshalIndent(sMap, "", "  ")
	if err != nil {
		errorLog.Println("Servers marshall err:", err)
		return
	}

	err = ioutil.WriteFile("servers.json", out, 0755)
	if err != nil {
		errorLog.Println("Save servers err:", err)
	}

	return
}

func loadUsers() error {
	u.User = make(map[string]*user)
	file, err := ioutil.ReadFile("users.json")
	if err != nil {
		fmt.Println(true, "Users open err", err)
		return err
	}

	if len(file) < 1 {
		infoLog.Println("users.json is empty")
		return errEmptyFile
	}

	err = json.Unmarshal(file, u)
	if err != nil {
		errorLog.Println("Users unmarshal err", err)
		return err
	}

	return nil
}

func saveUsers() {
	out, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		errorLog.Println("Users marshall err:", err)
		return
	}

	err = ioutil.WriteFile("users.json", out, 0755)
	if err != nil {
		errorLog.Println("Save user err:", err)
	}

	return
}

func loadQueue() error {
	q.QueuedMsgs = make(map[string]*queuedImage)

	file, err := ioutil.ReadFile("queue.json")
	if err != nil {
		errorLog.Println("Queue open err", err)
		return err
	}

	if len(file) < 1 {
		infoLog.Println("queue.json is empty")
		return nil
	}

	err = json.Unmarshal(file, q)
	if err != nil {
		errorLog.Println("Queue unmarshal err", err)
		return err
	}

	return nil
}

func saveQueue() {
	out, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		errorLog.Println("Queue marshall err:", err)
		return
	}

	err = ioutil.WriteFile("queue.json", out, 0755)
	if err != nil {
		errorLog.Println("Save queue err:", err)
	}
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	guildDetails, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("Message create guild details err:")
		return
	}

	var prefix string
	if _, ok := sMap.Server[guildDetails.ID]; ok {
		if prefix = sMap.Server[guildDetails.ID].Prefix; prefix == "" {
			prefix = c.Prefix
		}
	}

	if strings.HasPrefix(m.Content, prefix) {
		//code to check if extra whitespace is between prefix and command. Not allowed, nope :}
		//would break prefixes without trailing whitespace otherwise
		var command string
		if string(strings.TrimPrefix(m.Content, prefix)[0]) == " " {
			command = " "
		}

		msgList := strings.Fields(strings.TrimPrefix(m.Content, prefix))

		if len(msgList) > 0 {
			command += msgList[0]
			parseCommand(s, m, command, msgList)
		}
	} else if prefix != c.Prefix && strings.HasPrefix(m.Content, c.Prefix) {
		msgList := strings.Fields(strings.TrimPrefix(m.Content, c.Prefix))

		if len(msgList) > 0 {
			parseCommand(s, m, msgList[0], msgList)
		}
	}
	return
}

// Set all handlers for queued images, in case the bot crashes
// with images still in queue
func setQueuedImageHandlers(s *discordgo.Session) {
	for imgNum := range q.QueuedMsgs {
		imgNumInt, err := strconv.Atoi(imgNum)
		if err != nil {
			errorLog.Println("Error converting string to num for queue:", err)
			continue
		}
		go fimageReview(s, q, imgNumInt)
	}
}

func joined(s *discordgo.Session, m *discordgo.GuildCreate) {
	if m.Guild.Unavailable {
		return
	}

	guildDetails, err := s.State.Guild(m.Guild.ID)
	if err != nil {
		errorLog.Println("Join guild struct", err)
	}

	user, err := s.User(guildDetails.OwnerID)
	if err != nil {
		errorLog.Println("Joined user struct err", err)
		user = &discordgo.User{
			Username:      "error",
			Discriminator: "error",
		}
	}

	if _, ok := sMap.Server[m.Guild.ID]; !ok {
		//if newly joined
		s.ChannelMessageSendEmbed(logChan, &discordgo.MessageEmbed{
			Color: 65280,

			Image: &discordgo.MessageEmbedImage{
				URL: discordgo.EndpointGuildIcon(m.Guild.ID, m.Guild.Icon),
			},

			Footer: &discordgo.MessageEmbedFooter{
				Text: "Brought to you by 2Bot :)\nLast Bot reboot: " + lastReboot + " GMT",
			},

			Fields: []*discordgo.MessageEmbedField{
				{Name: "Name:", Value: m.Guild.Name, Inline: true},
				{Name: "User Count:", Value: strconv.Itoa(m.Guild.MemberCount), Inline: true},
				{Name: "Region:", Value: m.Guild.Region, Inline: true},
				{Name: "Channel Count:", Value: strconv.Itoa(len(m.Guild.Channels)), Inline: true},
				{Name: "ID:", Value: m.Guild.ID, Inline: true},
				{Name: "Owner:", Value: user.Username + "#" + user.Discriminator, Inline: true},
			},
		})

		sMap.Server[m.Guild.ID] = &server{
			LogChannel:  m.Guild.ID,
			Log:         false,
			Nsfw:        false,
			JoinMessage: [3]string{"false", "", ""},
		}

		infoLog.Println("Joined server", m.Guild.ID, m.Guild.Name)
	} else if val := sMap.Server[m.Guild.ID]; val.Kicked == true {
		//If previously kicked and then readded
		s.ChannelMessageSendEmbed(logChan, &discordgo.MessageEmbed{
			Color: 16751104,

			Image: &discordgo.MessageEmbedImage{
				URL: discordgo.EndpointGuildIcon(m.Guild.ID, m.Guild.Icon),
			},

			Footer: &discordgo.MessageEmbedFooter{
				Text: "Brought to you by 2Bot :)\nLast Bot reboot: " + lastReboot + " GMT",
			},

			Fields: []*discordgo.MessageEmbedField{
				{Name: "Name:", Value: m.Guild.Name, Inline: true},
				{Name: "User Count:", Value: strconv.Itoa(m.Guild.MemberCount), Inline: true},
				{Name: "Region:", Value: m.Guild.Region, Inline: true},
				{Name: "Channel Count:", Value: strconv.Itoa(len(m.Guild.Channels)), Inline: true},
				{Name: "ID:", Value: m.Guild.ID, Inline: true},
				{Name: "Owner:", Value: user.Username + "#" + user.Discriminator, Inline: true},
			},
		})

		infoLog.Println("Rejoined server", m.Guild.ID, m.Guild.Name)
	}

	sMap.Server[m.Guild.ID].Kicked = false
	saveServers()

	return
}

func kicked(s *discordgo.Session, m *discordgo.GuildDelete) {
	if !m.Unavailable {
		s.ChannelMessageSendEmbed(logChan, &discordgo.MessageEmbed{
			Color: 16711680,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Brought to you by 2Bot :)\nLast Bot reboot: " + lastReboot + " GMT",
			},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Name:", Value: m.Name, Inline: true},
				{Name: "ID:", Value: m.Guild.ID, Inline: true},
			},
		})

		infoLog.Println("Kicked from", m.Guild.ID, m.Name)

		sMap.Server[m.Guild.ID].Kicked = true
		saveServers()
	}
	return
}
