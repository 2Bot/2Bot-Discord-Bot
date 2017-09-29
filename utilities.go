package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

//From Necroforger's dgwidgets
func nextReactionAdd(s *discordgo.Session) chan *discordgo.MessageReactionAdd {
	out := make(chan *discordgo.MessageReactionAdd)
	s.AddHandlerOnce(func(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
		out <- e
	})
	return out
}

func nextMessageCreate(s *discordgo.Session) chan *discordgo.MessageCreate {
	out := make(chan *discordgo.MessageCreate)
	s.AddHandlerOnce(func(_ *discordgo.Session, e *discordgo.MessageCreate) {
		out <- e
	})
	return out
}

func randRange(min, max int) int {
	rand.Seed(time.Now().Unix())
	if max == 0 {
		return 0
	}
	return rand.Intn(max-min) + min
}

func findIndex(s []string, f string) int {
	for i, j := range s {
		if j == f {
			return i
		}
	}
	return -1
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getCreationTime(ID string) (t time.Time, err error) {
	i, err := strconv.ParseInt(ID, 10, 64)
	if err != nil {
		return
	}
	timestamp := (i >> 22) + 1420070400000
	t = time.Unix(timestamp/1000, 0)
	return
}

func codeSeg(s ...string) string {
	return "`" + strings.Join(s, " ") + "`"
}

func codeBlock(s ...string) string {
	return "```" + strings.Join(s, " ") + "```"
}

func isIn(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func trimSlice(s []string) (ret []string) {
	for _, i := range s {
		ret = append(ret, strings.TrimSpace(i))
	}
	return
}

func guildDetails(channelID string, s *discordgo.Session) (*discordgo.Guild, error) {
	channelInGuild, err := s.State.Channel(channelID)
	if err != nil {
		return nil, err
	}
	guildDetails, err := s.State.Guild(channelInGuild.GuildID)
	if err != nil {
		return nil, err
	}
	return guildDetails, nil
}

func isInServer(w http.ResponseWriter, r *http.Request) {
	authorID := r.FormValue("id")
	guild, err := guildDetails(serverID, dg)
	if err != nil {
		errorLog.Println(err)
		return
	}

	for _, member := range guild.Members {
		if member.User.ID == authorID {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
