package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	activeCommands   = make(map[string]command)
	disabledCommands = make(map[string]command)
)

type command struct {
	Name string
	Help string

	OwnerOnly     bool
	RequiresPerms bool

	PermsRequired int

	Exec func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

/*
	go-chi style command adding
		- aliases
		- nicer help text
		- uh...MORE
	also more here!
*/

func parseCommand(s *discordgo.Session, m *discordgo.MessageCreate, guildDetails *discordgo.Guild, message string) {
	msglist := strings.Fields(message)
	if len(msglist) == 0 {
		return
	}

	log.Trace(fmt.Sprintf("%s %s#%s, %s %s: %s", m.Author.ID, m.Author.Username, m.Author.Discriminator, guildDetails.ID, guildDetails.Name, m.Content))

	commandName := strings.ToLower(func() string {
		if strings.HasPrefix(message, " ") {
			return " " + msglist[0]
		}
		return msglist[0]
	}())

	if command, ok := activeCommands[commandName]; ok && commandName == strings.ToLower(command.Name) {
		userPerms, err := permissionDetails(m.Author.ID, m.ChannelID, s)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error verifying permissions :(")
			return
		}

		isOwner := m.Author.ID == conf.OwnerID
		hasPerms := userPerms&command.PermsRequired > 0
		if (!command.OwnerOnly && !command.RequiresPerms) || (command.RequiresPerms && hasPerms) || isOwner {
			command.Exec(s, m, msglist)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "You don't have the correct permissions to run this!")
		return
	}

	activeCommands["bigmoji"].Exec(s, m, msglist)
}

func (c command) add() command {
	activeCommands[strings.ToLower(c.Name)] = c
	return c
}

func newCommand(name string, permissions int, needsPerms bool, f func(*discordgo.Session, *discordgo.MessageCreate, []string)) command {
	return command{
		Name:          name,
		PermsRequired: permissions,
		RequiresPerms: needsPerms,
		Exec:          f,
	}
}

func (c command) alias(a string) command {
	activeCommands[strings.ToLower(a)] = c
	return c
}

func (c command) setHelp(help string) command {
	c.Help = help
	return c
}

func (c command) ownerOnly() command {
	c.OwnerOnly = true
	return c
}
