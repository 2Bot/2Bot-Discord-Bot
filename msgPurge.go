package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"time"
)

func msgPurge(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {	
	guildDetails, err := guildDetails(m.ChannelID, s)
	if err != nil {
		return
	}

	if m.Author.ID != guildDetails.OwnerID || m.Author.ID != noah {
		s.ChannelMessageSend(m.ChannelID, "Sorry, only the owner can do this")
		return
	}
	
	if len(msglist) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Gotta specify a number of messages to delete~")
		return
	}

/*	var userToPurge string
	if len(msglist) == 3 {
		userToPurge = msglist[2]
	}*/

	purgeAmount, err := strconv.Atoi(msglist[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("How do i delete %s messages O.o Please only give numbers!", msglist[1]))
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.Message.ID)

	loop := purgeAmount / 100
	for i := 0; i <= loop; i++ {
		if purgeAmount > 0 {
			del := min(purgeAmount, 100)
			list, err := s.ChannelMessages(m.ChannelID, del, "", "", "")						
			if err != nil {
				log(true, guildDetails.Name, guildDetails.ID, "Purge populate message list err:", err.Error())
				s.ChannelMessageSend(m.ChannelID, "There was a problem purging the chat :(")
				return
			}

			if len(list) == 0 {
				break
			}
			purgeList := []string{}
			for _, msg := range list {
				purgeList = append(purgeList, msg.ID)
			}

			err = s.ChannelMessagesBulkDelete(m.ChannelID, purgeList)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Dont have permissions or messages are older than 14 days :(")
				return
			}
			purgeAmount -= 100
		}
	}
	msg, _ := s.ChannelMessageSend(m.ChannelID, "Successfully deleted :ok_hand:")
	time.Sleep(time.Second * 5)
	s.ChannelMessageDelete(m.ChannelID, msg.ID)

	return
}