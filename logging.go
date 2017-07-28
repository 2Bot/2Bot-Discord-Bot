package main 

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strconv"
	"strings"
)

func membPresChange(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	if guild, ok := sMap.Server[m.GuildID]; ok && !guild.Kicked {
		if guild.Log {
			memberStruct, err := s.State.Member(m.GuildID, m.User.ID)
			if err != nil {
				errorLog.Println("Member struct error", err)
				return
			}

			s.ChannelMessageSend(guild.LogChannel, fmt.Sprintf("`%s is now %s`", memberStruct.User, status[m.Status]))
		}
	}
	return
}

func membJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if guild, ok := sMap.Server[m.GuildID]; ok && !guild.Kicked {
		if len(guild.JoinMessage) == 3 {
			isBool, err := strconv.ParseBool(guild.JoinMessage[0])
			if err != nil {
				errorLog.Println("Config join msg bool err", err)
				return
			}

			if isBool && guild.JoinMessage[1] != "" {
				guildDetails, err := s.State.Guild(m.GuildID)
				if err != nil {
					errorLog.Println("(membJoin) guildDetails err:", err)
					return
				}

				membStruct, err := s.User(m.User.ID)
				if err != nil {
					errorLog.Println(guildDetails.Name, m.GuildID, err)
					return
				}

				message := strings.Replace(guild.JoinMessage[1], "%s", membStruct.Mention(), -1)
				s.ChannelMessageSend(guild.JoinMessage[2], message)
			}
		}
	}
	return
}