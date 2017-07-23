package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"time"
)

func msgPurge(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		return
	}

	if m.Author.ID != guild.OwnerID && m.Author.ID != noah {
		s.ChannelMessageSend(m.ChannelID, "Sorry, only the owner can do this")
		return
	}

	if len(msglist) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Gotta specify a number of messages to delete~")
		return
	}

	var userToPurge string
	if len(msglist) == 3 {
		submatch := userIDRegex.FindStringSubmatch(msglist[2])
		if len(submatch) == 0 {
			s.ChannelMessageSend(m.ChannelID, "Couldn't find that user :(")
			return
		}
		userToPurge = submatch[1]
	}

	purgeAmount, err := strconv.Atoi(msglist[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("How do i delete %s messages O.o Please only give numbers!", msglist[1]))
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.Message.ID)

	if userToPurge == "" {
		standardPurge(purgeAmount, s, m)
	} else {

		err = userPurge(purgeAmount, s, m, userToPurge)
	}

	if err == nil {
		msg, _ := s.ChannelMessageSend(m.ChannelID, "Successfully deleted :ok_hand:")
		time.Sleep(time.Second * 5)
		s.ChannelMessageDelete(m.ChannelID, msg.ID)
	}
	return
}

func standardPurge(purgeAmount int, s *discordgo.Session, m *discordgo.MessageCreate) {
	outOfDate := false
	for purgeAmount > 0 {
		del := min(purgeAmount, 100)
		list, err := s.ChannelMessages(m.ChannelID, del, "", "", "")
		if err != nil {
			log(true, "Purge populate message list err:", err.Error())
			s.ChannelMessageSend(m.ChannelID, "There was a problem purging the chat :(")
			return
		}

		if len(list) == 0 {
			break
		}

		var purgeList []string
		for _, msg := range list {
			then, _ := msg.Timestamp.Parse()
			timeSince := time.Since(then)

			if timeSince.Hours()/24 >= 14 {
				outOfDate = true
				break
			}

			purgeList = append(purgeList, msg.ID)
		}

		err = s.ChannelMessagesBulkDelete(m.ChannelID, purgeList)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Dont have permissions to delete messages :(")
			return
		}

		if outOfDate {
			break
		}

		purgeAmount -= 100
	}
}

func userPurge(purgeAmount int, s *discordgo.Session, m *discordgo.MessageCreate, userToPurge string) error {
	for purgeAmount > 0 {
		del := min(purgeAmount, 100)
		var purgeList []string

	IfOutOfDate:
		for len(purgeList) < del {
			list, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
			if err != nil {
				log(true, "Purge populate message list err:", err.Error())
				s.ChannelMessageSend(m.ChannelID, "There was a problem purging the chat :(")
				return err
			}

			if len(list) == 0 {
				break
			}

			for _, msg := range list {
				if len(purgeList) >= del {
					break
				}

				if msg.Author.ID == userToPurge {
					then, _ := msg.Timestamp.Parse()
					timeSince := time.Since(then)

					if timeSince.Hours()/24 >= 14 {
						break IfOutOfDate
					}

					purgeList = append(purgeList, msg.ID)
				}
			}
		}

		err := s.ChannelMessagesBulkDelete(m.ChannelID, purgeList)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Dont have permissions to delete messages :( \n"+err.Error())
			return err
		}

		purgeAmount -= 100
	}

	return nil
}
