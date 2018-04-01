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
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
)

const (
	happyEmoji string = "https://cdn.discordapp.com/emojis/332968429210435585.png"
	thinkEmoji string = "https://cdn.discordapp.com/emojis/333694872802426880.png"
	reviewChan string = "334092230845267988"
	noah       string = "149612775587446784"
	logChan    string = "312352242504040448"
	serverID   string = "312292616089894924"
	xmark      string = "<:xmark:314349398824058880>"
)

var (
	dg           *discordgo.Session
	errorLog     *log.Logger
	infoLog      *log.Logger
	lastReboot   string
	emojiRegex   = regexp.MustCompile("<(a)?:.*?:(.*?)>")
	userIDRegex  = regexp.MustCompile("<@!?([0-9]+)>")
	channelRegex = regexp.MustCompile("<#([0-9]+)>")
	status       = map[discordgo.Status]string{"dnd": "busy", "online": "online", "idle": "idle", "offline": "offline"}
	footer       = new(discordgo.MessageEmbedFooter)
)

func init() {
	footer.Text = "Brought to you by 2Bot :)\nLast Bot reboot: " + time.Now().Format("Mon, 02-Jan-06 15:04:05 MST")
}

func start() {
	var err error
	dg, err = discordgo.New("Bot " + c.Token)
	if err != nil {
		errorLog.Fatalln("Error creating Discord session,", err)
	}

	infoLog.Println("session created")

	dg.AddHandler(messageCreateEvent)
	dg.AddHandler(presenceChangeEvent)
	dg.AddHandler(guildKickedEvent)
	dg.AddHandler(memberJoinEvent)
	dg.AddHandler(readyEvent)
	dg.AddHandler(guildJoinEvent)

	if err := dg.Open(); err != nil {
		errorLog.Fatalln("Error opening connection,", err)
	}

	infoLog.Println("connection opened")

	sMap.Count = len(sMap.Server)
}

func dailyJobs() {
	for {
		postServerCount()
		time.Sleep(time.Hour * 24)
	}
}

func postServerCount() {
	url := "https://bots.discord.pw/api/bots/301819949683572738/stats"

	count := sMap.getCount()

	jsonStr := []byte(`{"server_count":` + strconv.Itoa(count) + `}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errorLog.Println("error making bots.discord.pw request", err)
		return
	}

	req.Header.Set("Authorization", c.DiscordPWKey)
	req.Header.Set("Content-Type", "application/json")

	if _, err := new(http.Client).Do(req); err != nil {
		errorLog.Println("bots.discord.pw error", err)
		return
	}

	infoLog.Println("POSTed " + strconv.Itoa(count) + " to bots.discord.pw")
}

func setBotGame(s *discordgo.Session) {
	if err := s.UpdateStatus(0, c.Game); err != nil {
		errorLog.Println("Update status err:", err)
		return
	}
	infoLog.Println("set initial game to", c.Game)
}

// Set all handlers for queued images, in case the bot crashes with images still in queue
func setQueuedImageHandlers() {
	for imgNum := range q.QueuedMsgs {
		imgNumInt, err := strconv.Atoi(imgNum)
		if err != nil {
			errorLog.Println("Error converting string to num for queue:", err)
			continue
		}
		go fimageReview(dg, q, imgNumInt)
	}
}

func main() {
	runtime.GOMAXPROCS(c.MaxProc)

	infoLog = log.New(os.Stdout, "INFO:  ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Println("/*********BOT RESTARTING*********\\")

	for i, f := range []func() error{loadConfig, loadUsers, loadServers, loadQueue} {
		if err := f(); err != nil {
			switch i {
			case 0:
				errorLog.Fatalln(err)
			default:
				errorLog.Println(err)
			}
		}
	}
	defer cleanup()

	infoLog.Println("files loaded")

	start()
	defer dg.Close()

	go setQueuedImageHandlers()

	if !c.InDev {
		go dailyJobs()
	}

	// Setup http server for selfbots
	router := chi.NewRouter()
	router.Get("/image/{id:[0-9]{18}}/recall/{img:[0-9a-z]{64}}", httpImageRecall)
	router.Get("/inServer/{id:[0-9]{18}}", isInServer)

	go func() { errorLog.Println(http.ListenAndServe("0.0.0.0:8080", router)) }()

	infoLog.Println("Bot is now running. Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGHUP)
	<-sc
}
