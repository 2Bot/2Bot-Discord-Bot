package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var mem runtime.MemStats

func init() {
	newCommand("setGame", 0, true, false, msgSetGame).add()
	newCommand("listUsers", 0, true, false, msgListUsers).add()
	newCommand("reloadConfig", 0, true, false, msgReloadConfig)

	newCommand("help", 0, false, false, msgHelp).setHelp("ok").add()
	newCommand("info", 0, false, false, msgInfo).setHelp("Args: none\n\nSome info about 2Bot.\n\nExample:\n`!owo info`").add()
	newCommand("invite", 0, false, false, msgInvite).setHelp("Args: none\n\nSends an invite link for 2Bot!\n\nExample:\n`!owo invite`").add()
	newCommand("git", 0, false, false, msgGit).setHelp("Args: none\n\nLinks 2Bots github page.\n\nExample:\n`!owo git`").add()

	newCommand("setNSFW",
		discordgo.PermissionAdministrator|discordgo.PermissionManageServer,
		false, true, msgNSFW).setHelp("Args: none\n\nToggles NSFW commands in NSFW channels.\nAdmin only.\n\nExample:\n`!owo setNSFW`").add()

	newCommand("joinMessage",
		discordgo.PermissionAdministrator|discordgo.PermissionManageServer,
		false, true, msgJoinMessage).setHelp("Args: [true,false] | [message] | [channelID]\n\nEnables or disables join messages.\nthe message and channel that the bot welcomes new people in.\n" +
		"To mention the user in the message, put `%s` where you want the user to be mentioned in the message.\nLeave message \n\nExample to set message:\n" +
		"`!owo joinMessage true | Hey there %s! | 312294858582654978`\n>On member join\n`Hey there [@new member]`\n\n" +
		"Example to disable:\n`!owo joinMessage false`").add()

}

/*
	These are usually short commands that dont warrant their own file
	or are only for me, the creator..usually
*/

func msgSetGame(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if m.Author.ID != noah && len(msglist) < 2 {
		return
	}

	game := strings.Join(msglist[1:], " ")

	if err := s.UpdateStatus(0, game); err != nil {
		errorLog.Println("Game change error", err)
		return
	}

	c.Game = game
	saveConfig()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(":ok_hand: | Game changed to %s!", game))
	return
}

func msgHelp(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) == 2 {
		if val, ok := commMap[l(msglist[1])]; ok {
			val.helpCommand(s, m)
			return
		}
	}

	var commands []string
	for _, val := range commMap {
		if !val.NoahOnly {
			commands = append(commands, "`"+val.Name+"`")
		}
	}

	prefix := c.Prefix
	if guild, err := guildDetails(m.ChannelID, s); err != nil {
		prefix = sMap.Server[guild.ID].Prefix
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "2Bot help", Value: strings.Join(commands, ", ") + "\n\nUse `" + prefix + "help [command]` for detailed info about a command."},
		},
	})
}

func (c command) helpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  c.Name,
				Value: c.Help,
			},
		},
	})
}

func msgInfo(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) {
	ct1, _ := getCreationTime(s.State.User.ID)
	creationTime := ct1.Format(time.UnixDate)[:10]
	runtime.ReadMemStats(&mem)
	var prefix string
	guild, err := guildDetails(m.ChannelID, s)
	if err == nil {
		if val, ok := sMap.Server[guild.ID]; ok {
			prefix = val.Prefix
		}
	}
	if prefix == "" {
		prefix = "None"
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    s.State.User.Username,
			IconURL: discordgo.EndpointUserAvatar(s.State.User.ID, s.State.User.Avatar),
		},
		Footer: footer,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Bot Name:", Value: codeBlock(s.State.User.Username), Inline: true},
			{Name: "Creator:", Value: codeBlock("Strum355#0554"), Inline: true},
			{Name: "Creation Date:", Value: codeBlock(creationTime), Inline: true},
			{Name: "Global Prefix:", Value: codeBlock(c.Prefix), Inline: true},
			{Name: "Local Prefix", Value: codeBlock(prefix), Inline: true},
			{Name: "Programming Language:", Value: codeBlock("Go"), Inline: true},
			{Name: "Library:", Value: codeBlock("Discordgo"), Inline: true},
			{Name: "Server Count:", Value: codeBlock(strconv.Itoa(len(s.State.Guilds))), Inline: true},
			{Name: "Memory Usage:", Value: codeBlock(strconv.Itoa(int(mem.Alloc/1024/1024)) + "MB"), Inline: true},
			{Name: "My Server:", Value: "https://discord.gg/9T34Y6u\nJoin here for support amongst other things!", Inline: false},
		},
	})
	return
}

func msgListUsers(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if m.Author.ID != noah && len(msglist) < 2 {
		return
	}

	if guild, ok := sMap.Server[msglist[1]]; !ok || guild.Kicked {
		s.ChannelMessageSend(m.ChannelID, "2Bot isn't in that server")
		return
	}

	s.ChannelTyping(m.ChannelID)

	guild, err := guildDetails(msglist[1], s)
	if err != nil {
		return
	}

	var out []string

	for _, user := range guild.Members {
		//TODO limit check
		out = append(out, user.User.Username)
	}

	s.ChannelMessageSend(m.ChannelID, "Users in: "+guild.Name+"\n`"+strings.Join(out, ", ")+"`")
}

func msgGit(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) {
	s.ChannelMessageSend(m.ChannelID, "Check me out here https://github.com/Strum355/2Bot-Discord-Bot\nGive it star to make my creators day! ⭐")
}

func msgNSFW(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error toggling NSFW :( Try again please~")
		errorLog.Println("nsfw guild details error", err)
		return
	}

	onOrOff := map[bool]string{true: "enabled", false: "disabled"}

	if guild, ok := sMap.Server[guild.ID]; ok {
		guild.Nsfw = !guild.Nsfw
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("NSFW %s", onOrOff[guild.Nsfw]))
		saveServers()
	}
}

func msgJoinMessage(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an error with discord :( Try again please~")
		errorLog.Println("join message guild details error", err)
		return
	}

	split := trimSlice(strings.Split(strings.Join(msglist[1:], " "), "|"))

	if len(split) == 0 {
		split = append(split, msglist[1])
	}

	if len(split) > 0 {
		if guild, ok := sMap.Server[guild.ID]; ok {
			if split[0] != "false" && split[0] != "true" {
				s.ChannelMessageSend(m.ChannelID, "Please say either `true` or `false` for enabling or disabling join messages~")
				return
			}

			if split[0] == "false" {
				guild.JoinMessage = [3]string{split[0]}
				saveServers()
				s.ChannelMessageSend(m.ChannelID, "Join messages disabled! ")
				return
			}

			if len(split) != 3 {
				s.ChannelMessageSend(m.ChannelID, "Not enough info given! :/\nMake sure the command only has two `|` in it.")
				return
			}
			channelStruct, err := s.State.Channel(split[2])
			if err != nil {
				errorLog.Println("Join message channel struct or bad channel ID?", split[2], err)
				s.ChannelMessageSend(m.ChannelID, "Please give me a proper channel ID :(")
				return
			}

			if split[1] == "" {
				s.ChannelMessageSend(m.ChannelID, "No message given :/")
				return
			}

			guild.JoinMessage = [3]string{split[0], split[1], split[2]}
			saveServers()

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Join message set to:\n%s\nin %s", split[1], channelStruct.Name))
		}
	}
}

func msgReloadConfig(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if m.Author.ID != noah {
		return
	}

	if len(msglist) < 2 {
		return
	}

	var reloaded string
	switch msglist[1] {
	case "c":
		c = new(config)
		if err := loadConfig(); err != nil {
			errorLog.Println("Error reloading config", err)
			s.ChannelMessageSend(m.ChannelID, "Error reloading config")
			return
		}
		reloaded = "config"
	case "u":
		u = new(users)
		if err := loadUsers(); err != nil {
			errorLog.Println("Error reloading config", err)
			s.ChannelMessageSend(m.ChannelID, "Error reloading config")
			return
		}
		reloaded = "users"
	}

	s.ChannelMessageSend(m.ChannelID, "Reloaded "+reloaded)
}

func msgInvite(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) {
	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,
		Image: &discordgo.MessageEmbedImage{
			URL: happyEmoji,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Invite me with this link!", Value: "https://discordapp.com/oauth2/authorize?client_id=301819949683572738&scope=bot&permissions=3533824", Inline: true},
		},
	})
}
