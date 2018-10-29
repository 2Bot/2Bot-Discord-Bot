package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	newCommand("setGlobalPrefix", 0, false, msgGlobalPrefix).ownerOnly().add()
	newCommand("setPrefix",
		discordgo.PermissionAdministrator|discordgo.PermissionManageServer,
		true, msgPrefix).setHelp("Args: [prefix]\n\nSets the servers prefix to 'prefix'\nAdmin only.\n\nExample:\n`!owo setPrefix .`\nNew Example command:\n`.help`").add()
}

func prefixWorker(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) (prefix string, ok bool) {
	prefix = strings.Join(msglist[1:], " ")

	for {
		next := <-nextMessageCreate(s)
		if next.ChannelID != m.ChannelID || next.Author.ID != m.Author.ID {
			continue
		}

		response := strings.ToLower(next.Content)
		if response != "yes" && response != "no" {
			s.ChannelMessageSend(m.ChannelID, "Invalid response. Command cancelled.")
			return
		}

		f := func() string {
			if response == "yes" {
				return " "
			}
			return ""
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Prefix changed to %s %s a trailing space", codeSeg(prefix), func() string {
			if f() == "" {
				return "without"
			}
			return "with"
		}()))

		return prefix + f(), true
	}
}

func msgPrefix(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) < 2 {
		s.ChannelMessageSend(m.ChannelID, "No prefix given :/")
		return
	}

	guildDetails, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem changing the prefix :( Try again please~")
		return
	}

	guild, ok := sMap.server(guildDetails.ID)
	if !ok || guild.Kicked {
		return
	}

	prefix := strings.Join(msglist[1:], " ")

	s.ChannelMessageSend(m.ChannelID, "Do you want trailing space? (yes/no)"+
		"```"+
		prefix+" help -> with trailing space\n"+
		prefix+"help -> without trailing space```")

	if prefix, ok = prefixWorker(s, m, msglist); ok {
		guild.Prefix = prefix
		saveServers()
	}
}

func msgGlobalPrefix(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) < 2 {
		return
	}

	if prefix, ok := prefixWorker(s, m, msglist); ok {
		conf.Prefix = prefix
		saveConfig()
	}
}
