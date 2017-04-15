package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"regexp"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
)
var emojiRegex = regexp.MustCompile("<:.*?:(.*?)>")
var userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")
var reply = map[string]string{
	S"kiddo": "mulia vs lulia :thinking:",
	"jizz": ":weary: :sweat_drops:",
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(joined)
	dg.AddHandler(online)
	dg.AddHandler(membPresChange)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, event *discordgo.MessageCreate) {
	if strings.HasPrefix(event.Content, "!owo ") {
		if event.Author.ID == s.State.User.ID {
			return
		}
		msgList := strings.Split(strings.TrimPrefix(event.Content, "!owo "), " ")
		command := strings.TrimSpace(msgList[0])

		if command == "bigMoji" {
			submatch := emojiRegex.FindStringSubmatch(msgList[1])
			if len(submatch) != 0 {
				h, err := http.Get("https://cdn.discordapp.com/emojis/" + submatch[1] + ".png")
				if err != nil {
					return
				}
				s.ChannelFileSend(event.ChannelID, "emoji.png", h.Body)
				h.Body.Close()

				return
			}
		}else if command == "userStats" {
			submatch := userIDRegex.FindStringSubmatch(msgList[1])

			var userID string
			var nick string

			if len(submatch) != 0 {
				userID = submatch[1]
			}

			user, error := s.User(userID)
			if error != nil {
				return
			}

			channelInGuild, _ := s.Channel(event.ChannelID)
			memberStruct, _   := s.State.Member(channelInGuild.GuildID, user.ID)
			guildDetails, _   := s.Guild(channelInGuild.GuildID)
			roleStruct 		  := guildDetails.Roles

			var roleNames []string

			for _, role := range memberStruct.Roles {
				for _, guildRole := range roleStruct {
					if guildRole.ID == role{
						roleNames = append(roleNames, guildRole.Name)
					}
				}
			}

			if memberStruct.Nick == "" {nick = "None"}else{nick = memberStruct.Nick}
			if len(roleNames) == 0 {roleNames = append(roleNames, "None")}

			s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
					Color:       s.State.UserColor(userID, event.ChannelID),
					Description: fmt.Sprintf("%s is a loyal member of %s", user.Username, guildDetails.Name),
					Author: 	 &discordgo.MessageEmbedAuthor{
						Name:    	user.Username,
						IconURL: 	discordgo.EndpointUserAvatar(userID, user.Avatar),
					},
					Footer: 	 &discordgo.MessageEmbedFooter{
						Text: 	 	"Brought to you by 2Bot", 
					},
					Fields: 	 []*discordgo.MessageEmbedField {
									&discordgo.MessageEmbedField{Name: "Username", Value: user.Username, Inline: true},
									&discordgo.MessageEmbedField{Name: "Nickname", Value: nick, Inline: true},
									&discordgo.MessageEmbedField{Name: "Joined", Value: memberStruct.JoinedAt[:10], Inline: false},
									&discordgo.MessageEmbedField{Name: "Roles", Value: strings.Join(roleNames, ", "), Inline: false},
								 },
				})

		}else{
			keyWord := reply[strings.TrimSpace(msgList[0])]
			if keyWord != "" {
				s.ChannelMessageSend(event.ChannelID, keyWord)
			}
		}
	}
}

func joined(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			//s.ChannelMessageSend(channel.ID, "Waddup mah niggas. Its ya boi Pepe Cheeze on the line!")
			return
		}
	}
}

func online(s *discordgo.Session, event *discordgo.Ready){
	s.UpdateStatus(0, "with your mom xd")
}

func membPresChange(s *discordgo.Session, event *discordgo.PresenceUpdate) {
	for _, guild := range s.State.Guilds {
		for _, channel := range guild.Channels {
			if channel.ID == guild.ID && event.GuildID == guild.ID{
			//	memberStruct, _ := s.State.Member(guild.ID, event.User.ID)
				if event.Presence.Nick != "" {
				//	s.ChannelMessageSend(channel.ID, fmt.Sprintf("`%s is now %s`", event.Presence.Nick, event.Status))
				}else{
				//	s.ChannelMessageSend(channel.ID, fmt.Sprintf("`%s is now %s`", memberStruct.User, event.Status))
				}
			}
		}
	}
}