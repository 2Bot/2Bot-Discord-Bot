package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"regexp"
	"encoding/json"
	"io/ioutil"
	"github.com/bwmarrin/discordgo"
	"time"
	"math/rand"
	"strconv"
	"net/http"
	"bytes"
)

var c = &config{}
//var buffer = make([][]byte, 0)

var (
	lastReboot string
	token string
	emojiRegex = regexp.MustCompile("<:.*?:(.*?)>")
 	userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")
	status = map[discordgo.Status]string{"dnd":"busy","online":"online","idle":"idle","offline":"offline"}
	//Discord Bots, cool kidz only, social experiment, discord go			
	blacklist = []string{"110373943822540800", "272873324705742848", "244133074328092673",  "118456055842734083"}
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	timeNow := time.Now()
	lastReboot = timeNow.Format(time.RFC1123)[:22]
}

func main() {
	loadConfig()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	//Register event handlers
	dg.AddHandler(messageCreate)
	// dg.AddHandler(messageDelete)
	dg.AddHandler(joined)
	dg.AddHandler(membPresChange)
	dg.AddHandler(kicked)
	dg.AddHandler(membJoin)

	loadCommands()
	setInitialGame(dg)
	//go postServerCount()

	log(false, "\n",`/*********BOT RESTARTED*********\`)	

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc 
}

func loadConfig() {
	c.Servers = make(map[string]*server)
	file, err := ioutil.ReadFile("config.json")
	if err != nil { 
		log(true,"Config open err", err.Error())
		return 
	}

	json.Unmarshal(file, c)
	for gID, guild := range c.Servers {
		if guild.LogChannel == "" {
			guild.LogChannel = gID
			saveConfig()
		}
	}
	return
}

func postServerCount(){
	for {
		sCount := activeServerCount()
		url 	 := "https://bots.discord.pw/api/bots/301819949683572738/stats"
		jsonStr  := []byte(`{"server_count":22}`)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

		req.Header.Set("Authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiIxNDk2MTI3NzU1ODc0NDY3ODQiLCJyYW5kIjo4NiwiaWF0IjoxNDk3Nzk5MTkyfQ.chZmx9j84Yr0k22C46ftY8f_N2xS880KeXYFNLs3Dgs")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		_, err  = client.Do(req)
		if err != nil {
			log(true, "bots.discord.pw error", err.Error())
		}
		log(true, "POSTed"+strconv.Itoa(sCount)+"to bots.discord.pw")
		time.Sleep(time.Hour*24)
	}
}

func activeServerCount() (sCount int) {
	for _,g := range c.Servers {
		if !g.Kicked {
			sCount++
		}
	}
	return
}

func setInitialGame(s *discordgo.Session) {
	err := s.UpdateStatus(0, c.Game)
	if err != nil {
		log(true, "Update status err:", err.Error())
	}
	return
}

func saveConfig(){
	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil { 
		log(true, "Config marshall err:", err.Error())
		return
	}
	
	err = ioutil.WriteFile("config.json", out, 0777)
	if err != nil { 
		log(true, "Save config err:", err.Error()) 
	}
	return
}

func randRange(min, max int) int {
    rand.Seed(time.Now().Unix())
	if max == 0 {
		return 0
	}
    return rand.Intn(max - min) + min
}

func log(timed bool, s ...string) {
	var f *os.File
	var out []byte
	var time1 string

	if _, err := os.Stat("err.log"); err == nil {
		f, err = os.OpenFile("err.log", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			fmt.Println(err)
			return 
		}
		defer f.Close()
	} else {
		f, err = os.Create("err.log")
		if err != nil { 
			fmt.Println(err)
			return 
		}
		defer f.Close()
	}

	if timed {
		time1 = time.Now().Format(time.RFC822)[:15]+" "
	}

	out = []byte(time1 + strings.Join(s, " ")+"\n")

	_, err := f.Write(out)
	if err != nil { 
		fmt.Println(err)
		return 
	}
	return
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
        if b == a { return true }
    }
    return false
}

func trimSlice(s []string) (ret []string) {
	for _,i := range s {
		ret = append(ret, strings.TrimSpace(i))
	}
	return
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

func guildDetails(id string, s *discordgo.Session) (*discordgo.Guild, error){
	channelInGuild, err := s.State.Channel(id)
	if err != nil {
		log(true, "channelInGuild err:", err.Error()) 
		return nil, err
	}
	guildDetails, err := s.State.Guild(channelInGuild.GuildID)
	if err != nil { 
		log(true, "guildDetails err:", err.Error())
		return nil, err
	}
	return guildDetails, nil
}

func messageCreate(s *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.ID == s.State.User.ID || event.Author.Bot {
		return
	}

	guildDetails, err := guildDetails(event.ChannelID, s)
	if err != nil {
		return
	}

	var prefix string
	if prefix = c.Servers[guildDetails.ID].Prefix; prefix == "" {
		prefix = c.Prefix
	}

	if strings.HasPrefix(event.Content, prefix) {
		//code seg checks if extra whitespace is between prefix and command. Not allowed, nope :} 
		//would break prefixes without trailing whitespace otherwise
		var command string

		//casted to rune to index, cant index strings :(
		if string([]rune(strings.TrimPrefix(event.Content, prefix))[0]) == " " {
			command += " "
		}

		msgList := strings.Fields(strings.TrimPrefix(event.Content, prefix))
		//fmt.Println(strings.Join(msgList, ", "))
		if len(msgList) > 0 {
			command += msgList[0]
			parseCommand(s, event, command, msgList)
		}
	}
	return
}


func joined(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	if guild, ok := c.Servers[event.Guild.ID]; !ok && !guild.Kicked {
		c.Servers[event.Guild.ID] = &server {
			LogChannel: event.Guild.ID,
			Log: false,
			Nsfw: false,
			JoinMessage: []string{"false", "Welcome %s", event.Guild.ID},
		}
	}

	c.Servers[event.Guild.ID].Kicked = false

	saveConfig()

	log(true, "Joined server", event.Guild.ID, event.Guild.Name)
	return
}

func kicked(s *discordgo.Session, event *discordgo.GuildDelete) {
	if !event.Unavailable {
		log(true, "Kicked from", event.Guild.ID, event.Name)
		c.Servers[event.Guild.ID].Kicked = true;
		saveConfig() 
	}
	return
}

func membPresChange(s *discordgo.Session, event *discordgo.PresenceUpdate) {
	if guild, ok := c.Servers[event.GuildID]; ok && !guild.Kicked{
		if guild.Log {
			guildDetails, err   := s.State.Guild(event.GuildID)
			if err != nil { 
				 log(true, "guildDetails err:", err.Error())
				 return
			}

			memberStruct, err := s.State.Member(event.GuildID, event.User.ID)
			if err != nil { 
				log(true, guildDetails.Name, event.GuildID, err.Error())
				return 
			}

			s.ChannelMessageSend(guild.LogChannel, fmt.Sprintf("`%s is now %s`", memberStruct.User, status[event.Status]))
		}
	}
	return
}

func membJoin(s *discordgo.Session, event *discordgo.GuildMemberAdd){
	if guild, ok := c.Servers[event.GuildID]; ok && !guild.Kicked{
		if len(guild.JoinMessage) == 3 {
			if isBool,_ := strconv.ParseBool(guild.JoinMessage[0]); isBool && guild.JoinMessage[1] != "" {
				guildDetails, err := s.State.Guild(event.GuildID)
				if err != nil { 
					log(true, "guildDetails err:", err.Error())
					return
				}			

				membStruct, err := s.User(event.User.ID)
				if err != nil { 
					log(true,  guildDetails.Name, event.GuildID, err.Error())
					return 
				}

				message := strings.Replace(guild.JoinMessage[1], "%s", membStruct.Mention(), -1)
				s.ChannelMessageSend(guild.JoinMessage[2], message)
			}
		}
	}
	return
}