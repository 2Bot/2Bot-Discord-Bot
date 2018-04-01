package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	activeCommands   = make(map[string]command)
	disabledCommands = make(map[string]command)
)

//Small wrapper function to reduce clutter
func l(s string) string {
	return strings.ToLower(s)
}

func parseCommand(s *discordgo.Session, m *discordgo.MessageCreate, message string) {
	msglist := strings.Fields(message)
	if len(msglist) == 0 {
		return
	}
	commandName := l(func() string {
		if strings.HasPrefix(message, " ") {
			return " " + msglist[0]
		}
		return msglist[0]
	}())

	submatch := emojiRegex.FindStringSubmatch(msglist[0])
	if len(submatch) > 0 {
		activeCommands["bigmoji"].Exec(s, m, append([]string{""}, msglist...))
		return
	}

	if command, ok := activeCommands[commandName]; ok && commandName == l(command.Name) {
		userPerms, err := permissionDetails(m.Author.ID, m.ChannelID, s)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error verifying permissions :(")
			return
		}

		isNoah := m.Author.ID == noah
		hasPerms := userPerms&command.PermsRequired > 0
		if (!command.NoahOnly && !command.RequiresPerms) || (command.RequiresPerms && hasPerms) || isNoah {
			command.Exec(s, m, msglist)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "You don't have the correct permissions to run this!")
		return
	}

	activeCommands["bigmoji"].Exec(s, m, append([]string{""}, msglist...))
}

func (c command) add() command {
	activeCommands[l(c.Name)] = c
	return c
}

func newCommand(name string, permissions int, noah, needsPerms bool, f func(*discordgo.Session, *discordgo.MessageCreate, []string)) command {
	return command{
		Name:          name,
		PermsRequired: permissions,
		RequiresPerms: needsPerms,
		NoahOnly:      noah,
		Exec:          f,
	}
}

func (c command) setHelp(help string) command {
	c.Help = help
	return c
}
