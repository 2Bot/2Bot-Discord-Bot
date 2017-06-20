package main 

import (	
	"github.com/bwmarrin/discordgo"
	"strings"
)

var commMap = make(map[string]Command)


var setGame 	  = Command{"setGame", "", true, msgSetGame}
var announceComm  = Command{"announce", "", true, msgAnnounce}
var listUsersComm = Command{"listUsers","", true, msgListUsers}
var globalPrefix  = Command{"setGlobalPrefix","", true, msgGlobalPrefix}	
var avatar 		  = Command{"avatar","avatar help", false, msgAvatar}
var encodeComm 	  = Command{"encode","encode help", false, msgEncode}
var ibsearch 	  = Command{"ibsearch","ibsearch help", false, msgIbsearch}
var r34Comm 	  = Command{"r34","r34 help", false, msgRule34}
var userStats 	  = Command{"whois","whois help", false, msgUserStats}
var infoComm 	  = Command{"info","info help", false, msgInfo}
var bigMoji 	  = Command{"bigMoji","big moji help", false, msgEmoji}
var logChannel 	  = Command{"logChannel", "log channel help", false, msgLogChannel}
var nsfwComm 	  = Command{"setNSFW","nsfw help", false, msgNSFW}
var joinMsg 	  = Command{"joinMessage", "join message help", false, msgJoinMessage}
var loggingComm   = Command{"logging", "logging help", false, msgLogging}
var gitComm 	  = Command{"git", "git help", false, msgGit}
var purgeComm 	  = Command{"purge","purge help", false, msgPurge}
var prefixComm 	  = Command{"setPrefix","prefix help", false, msgPrefix}

//Command is the command struct from which all commands
//are made
type Command struct {
	name string 
	help string
	noahOnly bool
	exec func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

//Small wrapper function to reduce clutter
func l(s string) (r string) {
	return strings.ToLower(s)
}

func loadCommands(){
	commMap[l(avatar.name)] 		= avatar
	commMap[l(listUsersComm.name)]  = listUsersComm
	commMap[l(encodeComm.name)] 	= encodeComm
	commMap[l(purgeComm.name)] 		= purgeComm
	commMap[l(prefixComm.name)] 	= prefixComm
	commMap[l(ibsearch.name)] 		= ibsearch
	commMap[l(setGame.name)] 		= setGame
	commMap[l(globalPrefix.name)] 	= globalPrefix
	commMap[l(userStats.name)] 		= userStats
	commMap[l(infoComm.name)] 		= infoComm
	commMap[l(r34Comm.name)] 		= r34Comm
	commMap[l(logChannel.name)]		= logChannel
	commMap[l(gitComm.name)] 		= gitComm
	commMap[l(announceComm.name)]	= announceComm
	commMap[l(bigMoji.name)]		= bigMoji
	commMap[l(nsfwComm.name)]		= nsfwComm
	commMap[l(joinMsg.name)] 		= joinMsg
	commMap[l(loggingComm.name)]	= loggingComm


}

func parseCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, msgList []string) {
	command = strings.ToLower(strings.TrimSpace(command))
	submatch := emojiRegex.FindStringSubmatch(msgList[0])
	if len(submatch) != 0 || emojiFile(msgList[0]) != ""{

	}
	if command == "help" {
		if len(msgList)== 2{
			if val, ok := commMap[l(msgList[1])]; ok {
				val.helpCommand(s, m, msgList)
				return
			}
		}
		s.ChannelMessageSend(m.ChannelID, "type `!owo help [command]` for detailed help for a command. List of commands soon :) soz <3")
		return
	}
	if command == strings.ToLower(commMap[command].name) {
		commMap[command].exec(s, m, msgList)
	}
	return
}

func (c Command) helpCommand(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	s.ChannelMessageSend(m.ChannelID, c.help)
	return
}