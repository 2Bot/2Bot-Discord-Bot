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
	"encoding/json"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/2Bot/2Bot-Discord-Bot/actors"
	"github.com/bwmarrin/discordgo"
)

const (
	happyEmoji string = "https://cdn.discordapp.com/emojis/332968429210435585.png"
	thinkEmoji string = "https://cdn.discordapp.com/emojis/333694872802426880.png"
	reviewChan string = "334092230845267988"
	logChan    string = "312352242504040448"
	serverID   string = "312292616089894924"
	xmark      string = "<:xmark:314349398824058880>"
	zerowidth  string = "​"
)

var (
	conf         = new(config)
	dg           *discordgo.Session
	lastReboot   string
	log          = newLog()
	emojiRegex   = regexp.MustCompile("<(a)?:.*?:(.*?)>")
	userIDRegex  = regexp.MustCompile("<@!?([0-9]+)>")
	channelRegex = regexp.MustCompile("<#([0-9]+)>")
	footer       = new(discordgo.MessageEmbedFooter)
)

func init() {
	footer.Text = "Created with ❤ by Strum355\nLast Bot reboot: " + time.Now().Format("Mon, 02-Jan-06 15:04:05 MST")
}

func setBotGame(s *discordgo.Session) {
	if err := s.UpdateStatus(0, conf.Game); err != nil {
		log.Error("Update status err:", err)
		return
	}
	log.Info("set initial game to", conf.Game)
}

func saveJSON(path string, data interface{}) error {
	f, err := os.OpenFile("json/"+path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Error("error saving", path, err)
		return err
	}

	if err = json.NewEncoder(f).Encode(data); err != nil {
		log.Error("error saving", path, err)
		return err
	}
	return nil
}

func loadJSON(path string, v interface{}) error {
	f, err := os.OpenFile("json/"+path, os.O_RDONLY, 0600)
	if err != nil {
		log.Error("error loading", path, err)
		return err
	}

	if err := json.NewDecoder(f).Decode(v); err != nil {
		log.Error("error loading", path, err)
		return err
	}
	return nil
}

func cleanup() {
	for _, f := range []func() error{saveConfig} {
		if err := f(); err != nil {
			log.Error("error cleaning up files", err)
		}
	}
	log.Info("Done cleanup. Exiting.")
}

func loadConfig() error {
	return loadJSON("config.json", conf)
}

func saveConfig() error {
	return saveJSON("config.json", conf)
}

func main() {
	log.Info("/*********BOT RESTARTING*********\\")

	names := []string{"config"}
	for i, f := range []func() error{loadConfig} {
		if err := f(); err != nil {
			switch i {
			case 0:
				os.Exit(404)
			default:
			}
			continue
		}
		log.Trace("loaded", names[i])
	}
	defer cleanup()

	log.Info("files loaded")

	// Initialize the actor model system
	system = actors.NewActorSystem()

	var err error
	dg, err = discordgo.New("Bot " + conf.Token)
	if err != nil {
		log.Error("Error creating Discord session,", err)
		return
	}

	log.Trace("session created")

	dg.AddHandler(messageCreateEvent)
	//dg.AddHandler(presenceChangeEvent)
	//dg.AddHandler(guildKickedEvent)
	//dg.AddHandler(memberJoinEvent)
	//dg.AddHandler(readyEvent)
	//dg.AddHandler(guildJoinEvent)

	if err := dg.Open(); err != nil {
		log.Error("Error opening connection,", err)
		return
	}
	defer dg.Close()

	log.Trace("connection opened")

	log.Info("Bot is now running. Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGHUP)
	<-sc
}
