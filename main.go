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
	"sort"
	"math/rand"
	"encoding/xml"
	"strconv"
	"github.com/PuerkitoBio/goquery"
    "net/url"
)

type Config struct {
	Game string `json:"game"`
	Prefix string `json:"prefix"`
} 

type Rule34 struct {
	PostCount  int `xml:"count,attr"`
	Posts	   []struct {
		URL string `xml:"file_url,attr"`
	} `xml:"post"`	
}

var c *Config
var r34 Rule34

var (
	lastReboot string
	Token string
	emojiRegex = regexp.MustCompile("<:.*?:(.*?)>")
 	userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")

	commandList = []string{"bigMoji","userStats","help","r34","info","ibsearch", "purge"}
	helpText = map[string]string{
		"bigMoji [emoji]":"Sends a large version of the emoji as an image.\nShorthand available by excluding 'bigMoji'",
		"userStats [user]":"Sends some basic details of the given user. \nIf no [user] is supplied, the command callers details are shown instead.",	
		"help":"Prints this useful help text :D",
		"r34 [search]":"Searches rule34.xxx for all your saucy needs",
		"info":"Sends some basic details of 2Bot",
		"ibsearch":"Searches ibsearch.xxx for an even large amount of \"stuff\" to satisfy your needs.\nExtra search parameters supported are: rating, format.\nExample: `!owo ibsearch Pokemon | rating=s | format=png`\nEach parameter must be seperated by a |\nAny amount of spacing between = works",
		"purge [n]":"Purges the n last messages in the channel",
	}
	helpKeys = []string{}
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
	timeNow := time.Now()
	lastReboot = timeNow.Format(time.RFC1123)[:22]
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	for k := range helpText {
		helpKeys = append(helpKeys, k)
	}
	sort.Strings(helpKeys)

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
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func loadConfig(s *discordgo.Session) {
	file, err := ioutil.ReadFile("config.json"); if err != nil { fmt.Println("Open config file error"); return }
	json.Unmarshal(file, &c)

	if c.Prefix == ""{
		c.Prefix = "!owo"
		err := saveConfig(); if err != nil { fmt.Println("Save config file error"); return }
	}
	s.UpdateStatus(0, c.Game)
}

func randRange(min, max int) int {
    rand.Seed(time.Now().Unix())
	if max == 0 {
		return 0
	}
    return rand.Intn(max - min) + min
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

func saveConfig() error {
	out, err := json.Marshal(&c)
	if err != nil {
		return err
	}
	err1 := ioutil.WriteFile("config.json", out, 0777); if err != nil {	return err1	}
	return nil
}

func codeSeg(s ...string) string {
	ret := "`"
	for _, i := range s {
		ret += i
	}
	return ret+"`"
}

func codeBlock(s ...string) string {
	ret := "```"
	for _, i := range s {
		ret += i
	}
	return ret+"```"
}

//
func emojiFile(s string) string {
	found := ""
	filename := ""
	for _, r := range s {
		if filename != "" {
			filename = fmt.Sprintf("%s-%x", filename, r)
		} else {
			filename = fmt.Sprintf("%x", r)
		}

		if _, err := os.Stat(fmt.Sprintf("emoji/%s.png", filename)); err == nil {
			found = filename
		} else if found != "" {
			return found
		}
	}
	return found
}

func messageCreate(s *discordgo.Session, event *discordgo.MessageCreate) {
	if strings.HasPrefix(event.Content, c.Prefix) {
		if event.Author.ID == s.State.User.ID || event.Author.Bot {
			return
		}

		msgList := strings.Fields(strings.TrimPrefix(event.Content, c.Prefix))

		if len(msgList) > 0 {
			command := strings.TrimSpace(msgList[0])
			channelInGuild, _ := s.Channel(event.ChannelID)
			guildDetails, _   := s.Guild(channelInGuild.GuildID)
			roleStruct 		  := guildDetails.Roles
			submatch := emojiRegex.FindStringSubmatch(msgList[0])

			//EMOJI 
			if command == commandList[0] || len(submatch) != 0 || emojiFile(msgList[0]) != ""{
				//if custom emoji
				if len(submatch) != 0 {
					var emojiID string
					if command == commandList[0]{
						emojiID = emojiRegex.FindStringSubmatch(msgList[1])[1]
					}else{
						emojiID = submatch[1]
					}
					resp, err := http.Get(fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", emojiID)); if err != nil { return }
					defer resp.Body.Close()					
					s.ChannelFileSend(event.ChannelID, "emoji.png", resp.Body)
					s.ChannelMessageDelete(event.ChannelID, event.ID)
				//elif not custom emoji
				}else{
					var name string
					if command == commandList[0] && len(msgList) > 1 {
						name = emojiFile(msgList[1])
					}else{
						name = emojiFile(msgList[0])
					}
					if name != "" {
						file, err := os.Open(fmt.Sprintf("emoji/%s.png", name)); if err != nil { fmt.Println(err); return }
						defer file.Close()
						s.ChannelFileSend(event.ChannelID, "emoji.png", file)
						s.ChannelMessageDelete(event.ChannelID, event.ID)
					}
				}

			//USER STATS			
			}else if command == commandList[1] {
				var userID string
				var nick string

				if len(msgList) > 1 {
					submatch := userIDRegex.FindStringSubmatch(msgList[1])
					if len(submatch) != 0 { 
						userID = submatch[1] 
					}
				} else {
					userID = event.Author.ID
				}

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
			}else if command == "setGame" && (event.Author.ID == "149612775587446784") {
				game := strings.TrimPrefix(event.Content, c.Prefix+" setGame ")
				s.UpdateStatus(0, fmt.Sprintf("%s", game))

				msg, _ := s.ChannelMessageSend(event.ChannelID, ":ok_hand: | Game changed successfully!")

				time.Sleep(time.Second*5)

				s.ChannelMessageDelete(event.ChannelID, event.ID)
				s.ChannelMessageDelete(event.ChannelID, msg.ID)

				c.Game = game
				err1 := saveConfig(); if err1 != nil {
					fmt.Println(err1)
				}

			//HELP
			}else if command == commandList[2] {
				var output []*discordgo.MessageEmbedField
				for _,item := range helpKeys{
					output = append(output, &discordgo.MessageEmbedField{Name: c.Prefix+" "+item, Value: helpText[item], Inline: false})
				}
				s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
						Color:       0,
						Author: 	 &discordgo.MessageEmbedAuthor{
							Name:    	s.State.User.Username,
							IconURL: 	discordgo.EndpointUserAvatar(s.State.User.ID, s.State.User.Avatar),
						},
						Footer: 	 &discordgo.MessageEmbedFooter{
							Text: 	 	"Brought to you by 2Bot", 
						},
						Fields: 	 output,
					})

			//HARAM AF
			}else if command == "haf" {
				img, err := os.Open("images/haram.jpg"); if err != nil { fmt.Println(err) }
				defer img.Close()
				s.ChannelFileSend(event.ChannelID, "haramaf.jpg", img)

			//SET PREFIX
			}else if command == "setPrefix" && (event.Author.ID == "149612775587446784") && len(msgList) > 1{
				c.Prefix = msgList[1]
				msg,_ := s.ChannelMessageSend(event.ChannelID, ":ok_hand: | All done! Prefix changed!")
				err := saveConfig(); if err != nil { fmt.Println(err); return }
				time.Sleep(time.Second*5)

				s.ChannelMessageDelete(event.ChannelID, event.ID)
				s.ChannelMessageDelete(event.ChannelID, msg.ID)

			} else if command == commandList[3] {
				if len(msgList) > 1 {
					s.ChannelTyping(event.ChannelID)
					var query string
					for _, word := range msgList[1:] {
						query += "+"+word
					}
					resp, err := http.Get(fmt.Sprintf("https://rule34.xxx/index.php?page=dapi&s=post&q=index&tags=%s",query)); if err != nil { return }
					defer resp.Body.Close()
					body, err1 := ioutil.ReadAll(resp.Body)	
					if err1 != nil {
						fmt.Println(err1)
						return
					}

					err2 := xml.Unmarshal(body, &r34)
					if err2 != nil {
						fmt.Println(err2);
						return
					}

					var url string
					if r34.PostCount == 0 {
						s.ChannelMessageSend(event.ChannelID, "No results ¯\\_(ツ)_/¯")
					} else {	
						url = "https:"+r34.Posts[randRange(0,len(r34.Posts)-1)].URL
						resp1, err2 := http.Get(url)
						if err2 != nil { 
							return 
						}
						
						defer resp1.Body.Close()						
						s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("%s searched for `%s` \n%s", event.Author.Username, strings.Replace(query, "+"," ",-1), url))
						//resets list of URLs
						r34.Posts = r34.Posts[:0]
					}
				}
				return

			}else if command == commandList[4]{
				ct1,_ := getCreationTime(s.State.User.ID)
				creationTime := ct1.Format(time.UnixDate)[:10]

				s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
					Color:       0,
					Author: 	 &discordgo.MessageEmbedAuthor{
						Name:    	s.State.User.Username,
						IconURL: 	discordgo.EndpointUserAvatar(s.State.User.ID, s.State.User.Avatar),
					},
					Footer: 	 &discordgo.MessageEmbedFooter{
						Text: 	 	"Brought to you by 2Bot.\nLast Bot reboot: " + lastReboot+ " GMT", 
					},
					Fields: 	 []*discordgo.MessageEmbedField {
									&discordgo.MessageEmbedField{Name: "Bot Name:", Value: codeBlock(s.State.User.Username), Inline: true},
									&discordgo.MessageEmbedField{Name: "Creation Date:", Value: codeBlock(creationTime), Inline: true},
									&discordgo.MessageEmbedField{Name: "Creator:", Value: "```Strum355#1180```", Inline: true},
									&discordgo.MessageEmbedField{Name: "Programming Language", Value: "```Go```", Inline: true},
									&discordgo.MessageEmbedField{Name: "Library", Value: "```Discordgo```", Inline: true},	
									&discordgo.MessageEmbedField{Name: "Server Count", Value: codeBlock(strconv.Itoa(len(s.State.Guilds))), Inline: true},								
								},					
					})

			//IBSEARCH
			}else if command == commandList[5]{

				queryList := strings.Split(strings.TrimPrefix(event.Content, c.Prefix+" ibsearch"), "|")

				var queries []string
				for i,item := range queryList {
					//queryList[i] = strings.TrimSpace(item)
					if strings.Contains(item,"=") {
						queries = append(queries,strings.TrimSpace(strings.Split(queryList[i],"=")[0]))
					}
				}

				finalQuery := " "
				commands := []string{"rating","format"}
					
				for _,item1 := range commands{
					for i,item2 := range queries {
						if strings.Contains(item1, item2){
							finalQuery += strings.Replace(queryList[i+1], " ", "",-1)+" "
						}
					}
				}
				
				URL, err := url.Parse("https://ibsearch.xxx")
				if err != nil {
					fmt.Println("IBSearch error", err)
					return
				}
				URL.Path += "/api/v1/images.html"
				par := url.Values{}
				par.Add("q", strings.TrimSpace(queryList[0])+finalQuery+"random:")
				par.Add("limit", "1")
				par.Add("key", "2480CFA681A7A882CB33C0E4BA00A812C6F906A6")
				URL.RawQuery = par.Encode()

				doc, err := goquery.NewDocument(URL.String())
				if err != nil {
					fmt.Println(err)
				}
				
				//this count is so naive..has to be a better way
				count := 0
				doc.Find("table tr").Each(func(_ int, tr *goquery.Selection)  {
					//For each <tr> found, find the <td>s inside
					tr.Find("td[colspan*=\"3\"]").Each(func(_ int, td *goquery.Selection){
						if (strings.HasSuffix(td.Text(), ".gif") || strings.HasSuffix(td.Text(), ".png") || strings.HasSuffix(td.Text(), ".jpg")) {	
							count++
							s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("https://im1.ibsearch.xxx/%s", td.Text()))	
						}
					})
				})
				//yuk
				if count == 0 {
					s.ChannelMessageSend(event.ChannelID, "No results ¯\\_(ツ)_/¯")
				}
			}else if command == commandList[6] && len(msgList) == 2 {
				purgeAmount,err1:= strconv.Atoi(msgList[1])
				fmt.Println(err1, purgeAmount)
				if (purgeAmount > 100 || purgeAmount < 1) || err1 != nil {
					msg,_ := s.ChannelMessageSend(event.ChannelID, "Number has to be between 1 and 100 inclusive :rolling_eyes:")
					time.Sleep(time.Second*5)				
					s.ChannelMessageDelete(event.ChannelID, event.Message.ID)
					s.ChannelMessageDelete(event.ChannelID, msg.ID)
					return
				}
				
					fmt.Println("is a number below 100 and above 1")
					list,_ := s.ChannelMessages(event.ChannelID, purgeAmount,"","","")
					purgeList := []string{}
					for _,msg := range list {
						purgeList = append(purgeList, msg.ID)
					}

					err := s.ChannelMessagesBulkDelete(event.ChannelID, purgeList)
					if err == nil {
						msg,_ := s.ChannelMessageSend(event.ChannelID, "Successfully deleted :ok_hand:")
						time.Sleep(time.Second*5)				
						s.ChannelMessageDelete(event.ChannelID, msg.ID)					
					}
				
			}
		}
	}
	return
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
	loadConfig(s)
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