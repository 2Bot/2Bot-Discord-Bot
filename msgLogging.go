package main 

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

func msgLogChannel(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string){	
	guildDetails, err := guildDetails(m.ChannelID, s)
	if err != nil {
		return
	}

	if len(msglist) < 2{
		return
	} 
	
	if m.Author.ID != guildDetails.OwnerID {
		s.ChannelMessageSend(m.ChannelID, "Sorry, only the owner can do this")
		return
	}
	
	if guild, ok := c.Servers[guildDetails.ID]; ok && !guild.Kicked {
		guild.LogChannel = msglist[1]
		saveConfig()
		channel, err := s.Channel(msglist[1])
		if err != nil {
			log(true, "Channel error", err.Error())
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Log channel changed to %s", channel.Name))
	}	
	return
}

func msgLogging(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string){
	guildDetails, err := guildDetails(m.ChannelID, s)
	if err != nil { 
		return
	}	

	if len(msglist) < 2 && m.Author.ID != guildDetails.OwnerID {
		return
	}

	if guild, ok := c.Servers[guildDetails.ID]; ok && !guild.Kicked {
		guild.Log = !guild.Log
		saveConfig()
	}
}