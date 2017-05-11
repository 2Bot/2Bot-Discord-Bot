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
	"encoding/base64"
	"encoding/hex"
	"crypto/md5"	
	"golang.org/x/crypto/bcrypt"
	"crypto/sha256"
	"io/ioutil"
	"github.com/bwmarrin/discordgo"
	"time"
	"math/rand"
	"encoding/xml"
	"strconv"
	"github.com/PuerkitoBio/goquery"
    "net/url"
	"runtime"	
)

type server struct {
	Nsfw bool `json:"nsfw"`
	ID string `json:"server_id"`
	LogChannel string `json:"log_channel"`
}

type config struct {
	Game string `json:"game"`
	Prefix string `json:"prefix"`
	Servers []*server
} 

type rule34 struct {
	PostCount  int `xml:"count,attr"`
	Posts	   []struct {
		URL string `xml:"file_url,attr"`
	} `xml:"post"`	
}

var c *config
var r34 *rule34

var (
	m runtime.MemStats	
	lastReboot string
	token string
	emojiRegex = regexp.MustCompile("<:.*?:(.*?)>")
 	userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")
	servers []string
	commandList = []string{"help","info","bigMoji","userStats","r34","ibsearch","encode", "setNsfw","purge","logChannel"}
	helpText = map[string]string{
		"bigMoji":"Args: [emoji]\nSends a large version of the emoji as an image.\nShorthand available by excluding 'bigMoji'",
		"userStats":"Args: [@user]\nSends some basic details of the given user. \nIf no [user] is supplied, the command callers details are shown instead.",	
		"help":"Prints this useful help text :D",
		"r34":"Args: [search term]\nNSFW\nSearches rule34.xxx for all your saucy needs",
		"info":"Sends some basic details of 2Bot: Creator, Library, RAM Usage, Language etc",
		"ibsearch":"Args: [search term] [filter(s)]\nNSFW\nSearches ibsearch.xxx for an even more \"stuff\" to satisfy your needs.\nExtra search parameters supported are: rating, format.\nExample: `!owo ibsearch Pokemon | rating=s | format=png`\nEach parameter must be seperated by a |\nformats=gif, png, jpg | rating=e (explicit), s (safe), q (questionable)",
		"purge":"Args: [number]\nADMIN\nPurges the n last messages in the channel, max 100 and cannot be older than 14 days",
		"encode": "Args: [text] [method]\nencodes [text] to/using [method]. Supported methods: MD5, Bcrypt, SHA256, Base64",
		"setNsfw": "ADMIN\nEnables or disables NSFW commands such as r34 and ibsearch",
		"logChannel": "Args: [channel ID]\nADMIN\nChanges the log channel to the channel with the given ID. Default is main channel",
	//	"yt": "Args: [url]\nPlays the given youtube video in a voice channel",
	}
	status = map[discordgo.Status]string{"dnd":"busy","online":"Online","idle":"Idle","offline":"Offline"}
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
	timeNow := time.Now()
	lastReboot = timeNow.Format(time.RFC1123)[:22]
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	//Register event handlers
	dg.AddHandler(messageCreate)
	dg.AddHandler(joined)
	dg.AddHandler(online)
	dg.AddHandler(membPresChange)
	dg.AddHandler(kicked)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func loadConfig(s *discordgo.Session) {
	file, err := ioutil.ReadFile("config.json"); if err != nil { log(err.Error()); return }
	json.Unmarshal(file, &c)

	for _, guild := range c.Servers {
		if guild.LogChannel == "" {
			guild.LogChannel = guild.ID
			saveConfig()
		}
	}
	s.UpdateStatus(0, c.Game)
	return
}

func saveConfig() error {
	out, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("config.json", out, 0777); if err != nil { return err }
	return nil
}

func randRange(min, max int) int {
    rand.Seed(time.Now().Unix())
	if max == 0 {
		return 0
	}
    return rand.Intn(max - min) + min
}

func log(s ...string) {
	fmt.Println(time.Now().Format(time.RFC822)[:15], strings.Join(s, " "))
}

func findIndex(s []string, f string) int {
	for i,j := range s {
		if(j == f){
			return i
		}
	}
	return -1
}

func remove(s []string, i int) []string {
    s[i] = s[len(s)-1]
    return s[:len(s)-1]
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

func isIn(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

//https://stackoverflow.com/questions/19374219/how-to-find-the-difference-between-two-slices-of-strings-in-golang
func difference(slice1 []string, slice2 []string) []string {
    var diff []string

    // Loop two times, first to find slice1 strings not in slice2,
    // second loop to find slice2 strings not in slice1
    for i := 0; i < 2; i++ {
        for _, s1 := range slice1 {
            found := false
            for _, s2 := range slice2 {
                if s1 == s2 {
                    found = true
                    break
                }
            }
            // String not found. We add it to return slice
            if !found {
                diff = append(diff, s1)
            }
        }
        // Swap the slices, only if it was the first loop
        if i == 0 {
            slice1, slice2 = slice2, slice1
        }
    }

    return diff
}

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
		runtime.ReadMemStats(&m)
		msgList := strings.Fields(strings.TrimPrefix(event.Content, c.Prefix))

		if len(msgList) > 0 {
			command := strings.TrimSpace(msgList[0])
			channelInGuild, err := s.Channel(event.ChannelID); if err != nil { log(err.Error()); return}
			guildDetails, err   := s.Guild(channelInGuild.GuildID); if err != nil { log(err.Error()); return}
			submatch := emojiRegex.FindStringSubmatch(msgList[0])

			nsfw := false
			for _,guild := range c.Servers {
				if(guild.ID==guildDetails.ID){
					nsfw = guild.Nsfw
				}
			}
			
			if command == "bigMoji" || len(submatch) != 0 || emojiFile(msgList[0]) != "" { //EMOJI 
				msgEmoji(msgList, submatch, command, s, event)
			}else if command == "userStats" { //USER STATS			
				msgUserStats(msgList, channelInGuild, guildDetails, s, event)
			}else if command == "help" { //HELP
				msgHelp(s, event)
			}else if command == "r34" && nsfw { //RULE34
				msgRule34(msgList, s, event)
			}else if command == "info" { //INFO
				msgInfo(s, event)			
			}else if command =="ibsearch" && nsfw { //IBSEARCH
				msgIbsearch(s, event)
			}else if command == "purge" && len(msgList) == 2 { //PURGE
				msgPurge(msgList, s, event)
			} else if command == "encode" && len(msgList) > 2 { //ENCODE
				msgEncode(msgList, s, event)
			}else if command == "yt" && len(msgList) == 2 {
				msgYoutube(msgList, s, event)

			//ADMIN OR PERSONAL SPECIFIC COMMANDS
		}else if command == "announce" && event.Author.ID == "149612775587446784" && len(msgList) > 1 { //ANNOUNCE
				//Discord Bots, cool kidz only, social experiment, discord go			
				blacklist := []string{"110373943822540800", "272873324705742848", "244133074328092673",  "118456055842734083"}
				for _, guild := range s.State.Guilds {
					if !isIn(guild.ID,  blacklist) {
						s.ChannelMessageSend(guild.ID, strings.Join(msgList[1:], " "))
					}
				}
			}else if command == "setGame" && (event.Author.ID == "149612775587446784") { //SET GAME
				msgSetGame(s, event)
			}else if command == "haf" { //HARAM AF
				img, err := os.Open("images/haram.jpg"); if err != nil { log(err.Error()); return }
				defer img.Close()
				s.ChannelFileSend(event.ChannelID, "haramaf.jpg", img)
			}else if command == "setPrefix" && event.Author.ID == "149612775587446784" { //SET PREFIX
				msgPrefix(msgList, s, event)
			}else if command == "logChannel" && event.Author.ID == guildDetails.OwnerID && len(msgList) == 2 {
				for _, guild := range c.Servers {
					if guildDetails.ID == guild.ID {
						for _, channel := range guildDetails.Channels {
							if msgList[1] == channel.ID {
								guild.LogChannel = msgList[1]
								s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("Log channel changed to %s", channel.Name))
								saveConfig()
							}
						}

					}
				}
			}else if command == "setNsfw" {
				if event.Author.ID == guildDetails.OwnerID {
					nsfw = !nsfw
					for _,guild := range c.Servers {
						if(guild.ID==guildDetails.ID){
							guild.Nsfw = nsfw
						}
					}
					s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("NSFW enabled: %t", nsfw))
					saveConfig()
				} else {
					s.ChannelMessageSend(event.ChannelID, "Sorry, only the owner can do this")
				}
			}
		}
	}
	return
}

func msgEmoji(msgList, submatch []string, command string, s *discordgo.Session, event *discordgo.MessageCreate) {
	//if custom emoji
	if len(submatch) != 0 {
		var emojiID string
		if command == commandList[0]{
			emojiID = emojiRegex.FindStringSubmatch(msgList[1])[1]
		}else{
			emojiID = submatch[1]
		}
		resp, err := http.Get(fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", emojiID)); if err != nil { log(err.Error()); return }
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
			file, err := os.Open(fmt.Sprintf("emoji/%s.png", name)); if err != nil { log(err.Error()); return }
			defer file.Close()
			s.ChannelFileSend(event.ChannelID, "emoji.png", file)
			s.ChannelMessageDelete(event.ChannelID, event.ID)
		}
	}
	return
}

func msgUserStats(msgList []string, channelInGuild *discordgo.Channel, guildDetails *discordgo.Guild,s *discordgo.Session, event *discordgo.MessageCreate) {
	var userID string
	var nick string
	roleStruct := guildDetails.Roles

	if len(msgList) > 1 {
		submatch := userIDRegex.FindStringSubmatch(msgList[1])
		if len(submatch) != 0 { 
			userID = submatch[1] 
		}
	} else {
		userID = event.Author.ID
	}

	user, err := s.User(userID); if err != nil { log(err.Error()); return }

	memberStruct, err := s.State.Member(channelInGuild.GuildID, user.ID); if err != nil { log(err.Error()); return }
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
				Text: 	 	"Brought to you by 2Bot :)", 
			},
			Fields: 	 []*discordgo.MessageEmbedField {
							&discordgo.MessageEmbedField{Name: "Username:", Value: user.Username, Inline: true},
							&discordgo.MessageEmbedField{Name: "Nickname:", Value: nick, Inline: true},
							&discordgo.MessageEmbedField{Name: "Joined Server:", Value: memberStruct.JoinedAt[:10], Inline: false},
							&discordgo.MessageEmbedField{Name: "Roles:", Value: strings.Join(roleNames, ", "), Inline: true},
					//		&discordgo.MessageEmbedField{Name: "ID Number:", Value: user.ID, Inline: true},
						},
		})

	return
}

func msgSetGame(s *discordgo.Session, event *discordgo.MessageCreate) {
	game := strings.TrimPrefix(event.Content, c.Prefix+"setGame ")
	s.UpdateStatus(0, fmt.Sprintf("%s", game))

	s.ChannelMessageSend(event.ChannelID, ":ok_hand: | Game changed successfully!")

/*	time.Sleep(time.Second*5)

	s.ChannelMessageDelete(event.ChannelID, event.ID)
	s.ChannelMessageDelete(event.ChannelID, msg.ID)
*/
	c.Game = game
	err := saveConfig(); if err != nil { log(err.Error()); return}

	return
}

func msgHelp(s *discordgo.Session, event *discordgo.MessageCreate) {
	var output []*discordgo.MessageEmbedField
	for _,item := range commandList{
		output = append(output, &discordgo.MessageEmbedField{Name: codeBlock(c.Prefix+" "+item), Value: codeBlock(helpText[item]), Inline: false})
	}
	s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
			Color:       0,
			Author: 	 &discordgo.MessageEmbedAuthor{
				Name:    	s.State.User.Username,
				IconURL: 	discordgo.EndpointUserAvatar(s.State.User.ID, s.State.User.Avatar),
			},
			Footer: 	 &discordgo.MessageEmbedFooter{
				Text: 	 	"Brought to you by 2Bot :)", 
			},
			Fields: 	 output,
		})

	return
}

func msgPrefix(msgList []string, s *discordgo.Session, event *discordgo.MessageCreate) {
	c.Prefix = msgList[1]
	s.ChannelMessageSend(event.ChannelID, ":ok_hand: | All done! Prefix changed!")
	err := saveConfig(); if err != nil { log(err.Error()); return }
/*	time.Sleep(time.Second*5)

	s.ChannelMessageDelete(event.ChannelID, event.ID)
	s.ChannelMessageDelete(event.ChannelID, msg.ID)
*/
	return
}

func msgRule34(msgList []string, s *discordgo.Session, event *discordgo.MessageCreate) {
	if len(msgList) > 1 {
		s.ChannelTyping(event.ChannelID)
		var query string
		for _, word := range msgList[1:] {
			query += "+"+word
		}
		resp, err := http.Get(fmt.Sprintf("https://rule34.xxx/index.php?page=dapi&s=post&q=index&tags=%s",query)); if err != nil { log(err.Error()); return }
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body); if err != nil { log(err.Error()); return }

		err = xml.Unmarshal(body, &r34); if err != nil { log(err.Error()); return }

		var url string
		if r34.PostCount == 0 {
			s.ChannelMessageSend(event.ChannelID, "No results ¯\\_(ツ)_/¯")
		} else {	
			url = "https:"+r34.Posts[randRange(0,len(r34.Posts)-1)].URL
			resp, err := http.Get(url); if err != nil { log(err.Error()); return }
			defer resp.Body.Close()	

			s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("%s searched for `%s` \n%s", event.Author.Username, strings.Replace(query, "+"," ",-1), url))
			//resets list of URLs
			r34.Posts = r34.Posts[:0]
		}
	}
	return
}

func msgInfo(s *discordgo.Session, event *discordgo.MessageCreate) {
	ct1,_ := getCreationTime(s.State.User.ID)
	creationTime := ct1.Format(time.UnixDate)[:10]

	s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
		Color:       0,
		Author: 	 &discordgo.MessageEmbedAuthor{
			Name:    	s.State.User.Username,
			IconURL: 	discordgo.EndpointUserAvatar(s.State.User.ID, s.State.User.Avatar),
		},
		Footer: 	 &discordgo.MessageEmbedFooter{
			Text: 	 	"Brought to you by 2Bot :)\nLast Bot reboot: " + lastReboot+ " GMT", 
		},
		Fields: 	 []*discordgo.MessageEmbedField {
						&discordgo.MessageEmbedField{Name: "Bot Name:", Value: codeBlock(s.State.User.Username), Inline: true},
						&discordgo.MessageEmbedField{Name: "Creator:", Value: codeBlock("Strum355#1180"), Inline: true},
						&discordgo.MessageEmbedField{Name: "Creation Date:", Value: codeBlock(creationTime), Inline: true},
						&discordgo.MessageEmbedField{Name: "Prefix:", Value: codeBlock(c.Prefix), Inline: true},									
						&discordgo.MessageEmbedField{Name: "Programming Language:", Value: codeBlock("Go"), Inline: true},
						&discordgo.MessageEmbedField{Name: "Library:", Value: codeBlock("Discordgo"), Inline: true},	
						&discordgo.MessageEmbedField{Name: "Server Count:", Value: codeBlock(strconv.Itoa(len(s.State.Guilds))), Inline: true},
						&discordgo.MessageEmbedField{Name: "Memory Usage:", Value: codeBlock(strconv.Itoa(int(m.Alloc /1024/1024))+"MB"), Inline: true },
						&discordgo.MessageEmbedField{Name: "My Server:", Value: "https://discord.gg/9T34Y6u\nJoin here for support amongst other things!", Inline: false},
					},					
		})
		
	return
}

func msgIbsearch(s *discordgo.Session, event *discordgo.MessageCreate) {
	queryList := strings.Split(strings.TrimPrefix(event.Content, c.Prefix+"ibsearch"), "|")
	finalQuery := " "
	filters := []string{"rating","format"}
	queries := []string{}
	URL, err := url.Parse("https://ibsearch.xxx")

	s.ChannelTyping(event.ChannelID)

	for i,item := range queryList {
		//queryList[i] = strings.TrimSpace(item)
		if strings.Contains(item,"=") {
			queries = append(queries,strings.TrimSpace(strings.Split(queryList[i],"=")[0]))
		}
	}

	for _,item1 := range filters{
		for i,item2 := range queries {
			if strings.Contains(item1, item2) {
				finalQuery += strings.Replace(queryList[i+1], " ", "",-1)+" "
			}
		}
	}
	
	if err != nil { fmt.Println("IBSearch error", err); return }

	//Assemble the URL
	URL.Path += "/api/v1/images.html"
	par := url.Values{}
	par.Add("q", strings.TrimSpace(queryList[0])+finalQuery+"random:")
	par.Add("limit", "1")
	par.Add("key", "2480CFA681A7A882CB33C0E4BA00A812C6F906A6")
	URL.RawQuery = par.Encode()

	doc, err := goquery.NewDocument(URL.String()); if err != nil { log(err.Error()); return }

	found := false
	doc.Find("table tr").Each(func(_ int, tr *goquery.Selection) {
		//For each <tr> found, find the <td>s inside
		tr.Find("td[colspan*=\"3\"]").Each(func(_ int, td *goquery.Selection) {
			if (strings.HasSuffix(td.Text(), ".gif") || strings.HasSuffix(td.Text(), ".png") || strings.HasSuffix(td.Text(), ".jpg")) {
				found = true
				s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("%s searched for %s \nhttps://im1.ibsearch.xxx/%s", event.Author.Username, codeSeg(queryList[0]), td.Text()))
				return	
			}
		})
	})

	if !found {	s.ChannelMessageSend(event.ChannelID, "No results ¯\\_(ツ)_/¯") }

	return
}

func msgPurge(msgList []string, s *discordgo.Session, event *discordgo.MessageCreate) {
	purgeAmount,err := strconv.Atoi(msgList[1])
	if (purgeAmount > 100 || purgeAmount < 1) || err != nil {
		msg,_ := s.ChannelMessageSend(event.ChannelID, "Number has to be between 1 and 100 inclusive")
		time.Sleep(time.Second*5)				
		s.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		s.ChannelMessageDelete(event.ChannelID, msg.ID)
		return
	}
	list,_ := s.ChannelMessages(event.ChannelID, purgeAmount,"","","")
	purgeList := []string{}
	for _,msg := range list {
		purgeList = append(purgeList, msg.ID)
	}

	err = s.ChannelMessagesBulkDelete(event.ChannelID, purgeList)
	if err != nil {
		msg,_ := s.ChannelMessageSend(event.ChannelID, "Dont have permissions or messages are older than 14 days :(")
		time.Sleep(time.Second*5)				
		s.ChannelMessageDelete(event.ChannelID, msg.ID)		
		return					
	}
	msg,_ := s.ChannelMessageSend(event.ChannelID, "Successfully deleted :ok_hand:")
	time.Sleep(time.Second*5)				
	s.ChannelMessageDelete(event.ChannelID, msg.ID)

	return
}

func msgEncode(msgList []string, s *discordgo.Session, event *discordgo.MessageCreate) {
	base := msgList[1]		
	text := strings.TrimPrefix(event.Content, fmt.Sprintf("%s encode %s ", c.Prefix, base))
	switch base {
		case "base64":
			s.ChannelTyping(event.ChannelID)										
			output := base64.StdEncoding.EncodeToString([]byte(text))
			s.ChannelMessageSend(event.ChannelID, output)
		case "bcrypt":
			s.ChannelTyping(event.ChannelID)					
			output, err:= bcrypt.GenerateFromPassword([]byte(text), 14); if err != nil {log(err.Error()); return}
			s.ChannelMessageSend(event.ChannelID, string(output))
		case "md5":
			s.ChannelTyping(event.ChannelID)					
			output := md5.Sum([]byte(text))
			s.ChannelMessageSend(event.ChannelID, hex.EncodeToString(output[:]))
		case "sha256":
			s.ChannelTyping(event.ChannelID)										
			hash := sha256.Sum256([]byte(text))
			s.ChannelMessageSend(event.ChannelID, hex.EncodeToString(hash[:]))
		default:
			s.ChannelMessageSend(event.ChannelID, "Base not supported")
	}
	return
}

func msgYoutube(m []string, s *discordgo.Session, event *discordgo.MessageCreate) {
	return
}

func joined(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, guild := range c.Servers {
		servers = append(servers, guild.ID)
	}
	//servers = append(servers, event.Guild.ID)

	if(!isIn(event.Guild.ID, servers)) {
		c.Servers = append(c.Servers, &server {
			ID: event.Guild.ID,
			Nsfw: false,
		})
	}
	saveConfig()

	log("Joined server", event.Guild.ID, event.Guild.Name)
	return
}

func kicked(s *discordgo.Session, event *discordgo.GuildDelete) {
	if !event.Unavailable {
		fmt.Println("Kicked from", event.ID, event.Name)
		/*err := os.Truncate("servers.dat", 0); if err != nil { fmt.Println(err); return }
		file, err := os.OpenFile("servers.dat", os.O_RDWR, os.ModeAppend); if err != nil { fmt.Println(err); return }
		defer file.Close()
		id := findIndex(servers, event.ID)
		servers = remove(servers, id)
		for _, server := range servers {
			file.Write([]byte(server+"\n"))
		}*/
	}
	return
}

func online(s *discordgo.Session, event *discordgo.Ready) {
	loadConfig(s)
	saveConfig()

/*	currServers := []string{}
	file, err := os.OpenFile("servers.dat", os.O_APPEND|os.O_RDWR, os.ModeAppend); if err != nil { fmt.Println(err); return }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	//loads all previously stored server IDs
	for scanner.Scan() {
		servers = append(servers, scanner.Text())
	}

	for _,guild := range event.Guilds {
		//stores all servers 
		currServers = append(currServers, guild.ID)

		//if bot has been added to server while offline, add to list of stored server IDs
		//and write to file
		if !isIn(guild.ID, servers){
			servers = append(servers, guild.ID)
			_, err := file.Write([]byte(guild.ID+"\n")); if err != nil { fmt.Println(err); return }
			err = file.Sync(); if err != nil { fmt.Println(err); return }
		}
	}*/
	return	
}

func membPresChange(s *discordgo.Session, event *discordgo.PresenceUpdate) {
	for _, guild := range s.State.Guilds {
		for _, confGuild := range c.Servers {
			for _, channel := range guild.Channels {
				if channel.ID == confGuild.LogChannel && event.GuildID == guild.ID{
					memberStruct, _ := s.State.Member(guild.ID, event.User.ID)
					if event.Presence.Nick != "" {
						s.ChannelMessageSend(channel.ID, fmt.Sprintf("`%s is now %s`", event.Presence.Nick, status[event.Status]))
					}else{
						s.ChannelMessageSend(channel.ID, fmt.Sprintf("`%s is now %s`", memberStruct.User, status[event.Status]))
					}
				}
			}
		}
	}
	return
}