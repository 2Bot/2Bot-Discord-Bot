package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
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

func deleteMessage(m *discordgo.Message, s *discordgo.Session) {
	if m != nil {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

func channelDetails(channelID string, s *discordgo.Session) (channelDetails *discordgo.Channel, err error) {
	channelDetails, err = s.State.Channel(channelID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			channelDetails, err = s.Channel(channelID)
			if err != nil {
				log.Error("error getting channel details", channelID, err)
			}
		}
	}
	return
}

func permissionDetails(authorID, channelID string, s *discordgo.Session) (perms int, err error) {
	perms, err = s.State.UserChannelPermissions(authorID, channelID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			perms, err = s.UserChannelPermissions(authorID, channelID)
			if err != nil {
				log.Error("error getting perm details", err)
			}
		}
	}
	return
}

func userDetails(memberID string, s *discordgo.Session) (user *discordgo.User, err error) {
	user, err = s.User(memberID)
	if err != nil {
		log.Error("error getting user details", err)
	}
	return
}

func activePrefix(channelID string, s *discordgo.Session) (prefix string, err error) {
	prefix = conf.Prefix
	guild, err := guildDetails(channelID, "", s)
	if err != nil {
		s.ChannelMessageSend(channelID, "There was an issue executing the command :( Try again please~")
		return
	} else if val, ok := sMap.server(guild.ID); ok && val.Prefix != "" {
		prefix = val.Prefix
	}
	return prefix, nil
}

func memberDetails(guildID, memberID string, s *discordgo.Session) (member *discordgo.Member, err error) {
	member, err = s.State.Member(guildID, memberID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			member, err = s.GuildMember(guildID, memberID)
			if err != nil {
				log.Error("error getting member details", err)
			}
		}
	}
	return
}

func guildDetails(channelID, guildID string, s *discordgo.Session) (guildDetails *discordgo.Guild, err error) {
	if guildID == "" {
		var channel *discordgo.Channel
		channel, err = channelDetails(channelID, s)
		if err != nil {
			return
		}

		guildID = channel.GuildID
	}

	guildDetails, err = s.State.Guild(guildID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			guildDetails, err = s.Guild(guildID)
			if err != nil {
				log.Error("error getting guild details", guildID, err)
			}
		}
	}
	return
}

func isInServer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := chi.URLParam(r, "id")
	guild, err := guildDetails(serverID, "", dg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, member := range guild.Members {
		if member.User.ID == id {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
