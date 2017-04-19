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
	"encoding/json"
	"io/ioutil"
	"github.com/bwmarrin/discordgo"
	"time"
)

var (
	Token string
	emojiRegex = regexp.MustCompile("<:.*?:(.*?)>")
 	userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")
 	reply = map[string]string{
		"kiddo": "mulia vs lulia :thinking:",
		"censored": "stuff",}
	commandList = []string{"bigMoji","userStats","setGame","help"}
)

type Config struct {
	Game string `json:"game"`
} 

var c *Config

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

	//Register event handlers
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

func LoadConfig(s *discordgo.Session) {
	_,err1 := s.UserUpdate("","","2Bot",s.State.User.Avatar,""); if err1 != nil {
		fmt.Println(err1)
	}
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("File error")
	}
	json.Unmarshal(file, &c)
	s.UpdateStatus(0, c.Game)
}

func SaveConfig() error {
	out, err := json.Marshal(&c)
	if err != nil {
		return err
	}
	err1 := ioutil.WriteFile("config.json", out, 0777)
	if err != nil {
		return err1
	}
	return nil
}

func messageCreate(s *discordgo.Session, event *discordgo.MessageCreate) {
	if strings.HasPrefix(event.Content, "!xd ") {

		if event.Author.ID == s.State.User.ID || event.Author.Bot {
			return
		}

		msgList := strings.Fields(strings.TrimPrefix(event.Content, "!xd "))
		command := strings.TrimSpace(msgList[0])
		channelInGuild, _ := s.Channel(event.ChannelID)
		guildDetails, _   := s.Guild(channelInGuild.GuildID)
		roleStruct 		  := guildDetails.Roles
		submatch := emojiRegex.FindStringSubmatch(msgList[0])

		//EMOJI 
		if command == commandList[0] || len(submatch) != 0 {
			var emojiID string
			if command == commandList[0] {
				emojiID = emojiRegex.FindStringSubmatch(msgList[1])[1]
			}else if len(submatch) != 0 {
				emojiID = submatch[1]
			}else {
				return
			}
			h, err := http.Get(fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", emojiID)); if err != nil { return }
			s.ChannelFileSend(event.ChannelID, "emoji.png", h.Body)
			h.Body.Close()

		//USER STATS			
		}else if command == commandList[1] {
			submatch := userIDRegex.FindStringSubmatch(msgList[1])

			var userID string
			var nick string

			if len(submatch) != 0 { userID = submatch[1] }

			user, error := s.User(userID); if error != nil { return }

			memberStruct, _ := s.State.Member(channelInGuild.GuildID, user.ID)

			var roleNames []string

			for _, role := range memberStruct.Roles {
				for _, guildRole := range roleStruct {
					if guildRole.ID == role{
						roleNames = append(roleNames, guildRole.Name)
					}
				}
			}

			if memberStruct.Nick == "" {
				nick = "None"
			}else{
				nick = memberStruct.Nick
			}
			
			if len(roleNames) == 0 {
				roleNames = append(roleNames, "None")
			}

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
									&discordgo.MessageEmbedField{Name: "Username:", Value: user.Username, Inline: true},
									&discordgo.MessageEmbedField{Name: "Nickname:", Value: nick, Inline: true},
									&discordgo.MessageEmbedField{Name: "Joined Server:", Value: memberStruct.JoinedAt[:10], Inline: false},
									&discordgo.MessageEmbedField{Name: "Roles:", Value: strings.Join(roleNames, ", "), Inline: true},
							//		&discordgo.MessageEmbedField{Name: "ID Number:", Value: user.ID, Inline: true},
								 },
				})

		//SET GAME
		}else if command == commandList[2] && (event.Author.ID == "149612775587446784" || event.Author.ID == guildDetails.OwnerID) {
			game := strings.TrimPrefix(event.Content, "!xd setGame ")
			s.UpdateStatus(0, fmt.Sprintf("%s", game))

			msg, _ := s.ChannelMessageSend(event.ChannelID, "Game changed successfully!")

			time.Sleep(time.Second*3)

			err := s.ChannelMessageDelete(event.ChannelID, event.ID); if err != nil{
        		fmt.Printf("Permission Error: %v\n", err)
			}
			s.ChannelMessageDelete(event.ChannelID, msg.ID)

			c.Game = game
			err1 := SaveConfig(); if err1 != nil {
				fmt.Println(err1)
			}

		//HELP
		}else if command == commandList[3] {
			var output []*discordgo.MessageEmbedField
			for _,item := range commandList{
				output = append(output, &discordgo.MessageEmbedField{Name: "!xd "+item, Value: item, Inline: false},
)
			}
			s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
					Color:       000,
					Author: 	 &discordgo.MessageEmbedAuthor{
						Name:    	s.State.User.Username,
						IconURL: 	discordgo.EndpointUserAvatar(s.State.User.ID, s.State.User.Avatar),
					},
					Footer: 	 &discordgo.MessageEmbedFooter{
						Text: 	 	"Brought to you by 2Bot", 
					},
					Fields: 	 output,
				})
		//RESPONSES
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
			//s.ChannelMessageSend(channel.ID, "Waddup mah bois. Its ya boi Pepe Cheeze on the line!")
			return
		}
	}
}

func online(s *discordgo.Session, event *discordgo.Ready){
	LoadConfig(s)
}

func membPresChange(s *discordgo.Session, event *discordgo.PresenceUpdate) {
	for _, guild := range s.State.Guilds {
		for _, channel := range guild.Channels {
			if channel.ID == guild.ID && event.GuildID == guild.ID{
				//memberStruct, _ := s.State.Member(guild.ID, event.User.ID)
				if event.Presence.Nick != "" {
					//s.ChannelMessageSend(channel.ID, fmt.Sprintf("`%s is now %s`", event.Presence.Nick, event.Status))
				}else{
					//s.ChannelMessageSend(channel.ID, fmt.Sprintf("`%s is now %s`", memberStruct.User, event.Status))
				}
			}
		}
	}
}