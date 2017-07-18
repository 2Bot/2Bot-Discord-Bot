package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

var (
	commMap = make(map[string]command)

	setGame       = command{"setGame", "", true, false, msgSetGame}
	announceComm  = command{"announce", "", true, false, msgAnnounce}
	listUsersComm = command{"listUsers", "", true, false, msgListUsers}
	globalPrefix  = command{"setGlobalPrefix", "", true, false, msgGlobalPrefix}
	//reloadConf    = command{"reload", "", true, false, msgReloadConfig}
	//printJSON     = command{"printJSON", "", true, false, msgPrintJSON}

	avatar = command{"avatar",
		"Args: [@user]\n\nReturns the given users avatar.\nIf no user ID is given, your own avatar is sent.\n\nExample:\n`!owo avatar @Strum355#2298`",
		false, false, msgAvatar}
	encodeComm = command{"encode",
		"Args: [base] [text]\n\nBases: `base64`, `bcrypt`, `md5`, `sh256`\nEncodes the given text in the given base.\n\nExample:\n`!owo encode md5 some text`",
		false, false, msgEncode}
	ibsearch = command{"ibsearch",
		"Args: [search] | rating=[e,s,q] | format=[gif,png,jpg]\n\nReturns a random image from ibsearch for the given search term with the given filters applied.\n\n" +
			"Example:\n`!owo ibsearch lewds | rating=e | format=gif`",
		false, false, msgIbsearch}
	r34Comm = command{"r34",
		"Args: [search]\n\nReturns a random image from rule34 for the given search term.\n\nExample:\n`!owo r34 lewds`",
		false, false, msgRule34}
	userStats = command{"whois",
		"Args: [@user]\n\nSome info about the given user.\n\nExample:\n`!owo whois @Strum355#2298`",
		false, false, msgUserStats}
	infoComm = command{"info",
		"Args: none\n\nSome info about 2Bot.\n\nExample:\n`!owo info`",
		false, false, msgInfo}
	bigMoji = command{"bigMoji",
		"Args: [emoji]\n\nSends a large image of the given emoji.\nCommand 'bigMoji' can be excluded for shorthand.\n\nExample:\n`!owo :smile:`\nor\n`!owo bigMoji :smile:`",
		false, false, msgEmoji}
	logChannel = command{"logChannel",
		"Args: [channelID]\n\nSets the log channel to the given channel.\nAdmin only.\n\nExample:\n`!owo logChannel 312292616089894924`",
		false, true, msgLogChannel}
	nsfwComm = command{"setNSFW",
		"Args: none\n\nToggles NSFW commands in NSFW channels.\nAdmin only.\n\nExample:\n`!owo setNSFW`",
		false, true, msgNSFW}
	joinMsg = command{"joinMessage",
		"Args: [true,false] | [message] | [channelID]\n\nEnables or disables join messages.\nthe message and channel that the bot welcomes new people in.\n" +
			"To mention the user in the message, put `%s` where you want the user to be mentioned in the message.\nLeave message \n\nExample to set message:\n`!owo joinMessage true | Hey there %s! | 312294858582654978`\n>On member join\n`Hey there [@new member]`\n\n" +
			"Example to disable:\n`!owo joinMessage false`",
		false, true, msgJoinMessage}
	loggingComm = command{"logging",
		"Args: none\n\nToggles user presence logging.\n\nExample:\n`!owo logging`",
		false, true, msgLogging}
	gitComm = command{"git",
		"Args: none\n\nLinks 2Bots github page.\n\nExample:\n`!owo git`",
		false, false, msgGit}
	purgeComm = command{"purge",
		"Args: [number]\n\nPurges 'number' amount of messages.\nAdmin only\n\nExample:\n!owo purge 300",
		false, true, msgPurge}
	prefixComm = command{"setPrefix",
		"Args: [prefix] | [whitespace?]\n\nSets the servers prefix to 'prefix'\nAdmin only.\n\nExample:\n`!owo setPrefix . | false`\nNew Example command:\n`.help`",
		false, true, msgPrefix}
	imageRecall = command{"image",
		"Args: [save,recall,delete,list,status] [name]\n\nSave images and recall them at anytime! Everyone gets 8MB of image storage. Any name counts so long theres no `/` in it." +
			"Only you can 'recall' your saved images. There's a review process to make sure nothing illegal is being uploaded but we're fairly relaxed for the most part\n\n" +
			"Example:\n`!owo image save 2B Happy`\n2Bot downloads the image and sends it off for reviewing\n\n" +
			"`!owo image recall 2B Happy`\nIf your image was confirmed, 2Bot will send the image named `2B Happy`\n\n" +
			"`!owo image delete 2B Happy`\nThis will delete the image you saved called `2B Happy`\n\n" +
			"`!owo image list`\nThis will list your saved images along with a preview!\n\n" +
			"`!owo image status`\nShows some details on your saved images and quota",
		false, false, msgImageRecall}
	inviteLink = command{"invite",
		"Args: none\n\nSends an invite link for 2Bot!\n\nExample:\n`!owo invite`",
		false, false, msgInvite}
)

//Small wrapper function to reduce clutter
func l(s string) (r string) {
	return strings.ToLower(s)
}

func loadCommands() {
	commMap[l(avatar.Name)] = avatar
	commMap[l(listUsersComm.Name)] = listUsersComm
	commMap[l(encodeComm.Name)] = encodeComm
	commMap[l(purgeComm.Name)] = purgeComm
	commMap[l(prefixComm.Name)] = prefixComm
	commMap[l(ibsearch.Name)] = ibsearch
	commMap[l(setGame.Name)] = setGame
	commMap[l(globalPrefix.Name)] = globalPrefix
	commMap[l(userStats.Name)] = userStats
	commMap[l(infoComm.Name)] = infoComm
	commMap[l(r34Comm.Name)] = r34Comm
	commMap[l(logChannel.Name)] = logChannel
	commMap[l(gitComm.Name)] = gitComm
	commMap[l(announceComm.Name)] = announceComm
	commMap[l(bigMoji.Name)] = bigMoji
	commMap[l(nsfwComm.Name)] = nsfwComm
	commMap[l(joinMsg.Name)] = joinMsg
	commMap[l(loggingComm.Name)] = loggingComm
	commMap[l(imageRecall.Name)] = imageRecall
	//	commMap[l(reloadConf.Name)] = reloadConf
	commMap[l(inviteLink.Name)] = inviteLink
	//commMap[l(printJSON.Name)] = printJSON
}

func parseCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, msgList []string) {
	command = strings.ToLower(strings.TrimSuffix(command, " "))

	submatch := emojiRegex.FindStringSubmatch(msgList[0])
	if len(submatch) > 0 {
		commMap[l(bigMoji.Name)].Exec(s, m, msgList)
	}

	if command == "help" {
		if len(msgList) == 2 {
			if val, ok := commMap[l(msgList[1])]; ok {
				val.helpCommand(s, m, msgList)
				return
			}
		}
		var commands []string
		for _, val := range commMap {
			if !val.NoahOnly {
				commands = append(commands, "`"+val.Name+"`")
			}
		}
		s.ChannelMessageSend(m.ChannelID, strings.Join(commands, ", ")+
			"\n\nUse `[prefix] help [command]` for detailed info about a command.")
		return
	}

	if command == strings.ToLower(commMap[command].Name) {
		commMap[command].Exec(s, m, msgList)
	}
	return
}

func (c command) helpCommand(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	s.ChannelMessageSend(m.ChannelID, c.Help)
	return
}
