package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func msgPurge(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem purging :( Please try again~")
		errorLog.Println("purge guild details error", err)
		return
	}

	if m.Author.ID != guild.OwnerID && m.Author.ID != noah {
		s.ChannelMessageSend(m.ChannelID, "Sorry, only the owner can do this~")
		return
	}

	if len(msglist) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Gotta specify a number of messages to delete~")
		return
	}

	if perm, err := s.State.UserChannelPermissions(s.State.User.ID, m.ChannelID); err == nil {
		if perm&discordgo.PermissionManageMessages <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Dont have permissions :(")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "Couldn't determine permissions :(")
		errorLog.Println("error getting permissions", err)
		return
	}

	purgeAmount, err := strconv.Atoi(msglist[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("How do i delete %s messages O.o Please only give numbers!", msglist[1]))
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

	deleteMessage(m, s)

	if userToPurge == "" {
		standardPurge(purgeAmount, s, m)
	} else {
		err = userPurge(purgeAmount, s, m, userToPurge)
	}

	if err == nil {
		msg, _ := s.ChannelMessageSend(m.ChannelID, "Successfully deleted :ok_hand:")
		time.Sleep(time.Second * 5)
		if msg != nil {
			s.ChannelMessageDelete(m.ChannelID, msg.ID)
		}
	}
	return
}

func standardPurge(purgeAmount int, s *discordgo.Session, m *discordgo.MessageCreate) error {
	outOfDate := false
	for purgeAmount > 0 {
		list, err := s.ChannelMessages(m.ChannelID, purgeAmount%100, "", "", "")
		if err != nil {
			errorLog.Println("Purge populate message list err:", err)
			s.ChannelMessageSend(m.ChannelID, "There was an issue deleting messages :(")
			return err
		}

		//if more was requested to be deleted than exists
		if len(list) == 0 {
			break
		}

		var purgeList []string
		for _, msg := range list {
			timeSince, err := getMessageAge(msg, s, m)
			if err != nil {
				//if the time is malformed for whatever reason, we'll try the next message
				continue
			}

			if timeSince.Hours()/24 >= 14 {
				outOfDate = true
				break
			}

			purgeList = append(purgeList, msg.ID)
		}

		if err := massDelete(purgeList, s, m); err != nil {
			return err
		}

		if outOfDate {
			break
		}

		purgeAmount -= 100
	}

	return nil
}

func userPurge(purgeAmount int, s *discordgo.Session, m *discordgo.MessageCreate, userToPurge string) error {
	for purgeAmount > 0 {
		del := min(purgeAmount, 100)
		var purgeList []string

	OutOfDate:
		for len(purgeList) < del {
			list, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "There was an issue deleting messages :(")
				errorLog.Println("Purge populate message list err:", err)
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
					timeSince, err := getMessageAge(msg, s, m)
					if err != nil {
						//if the time is malformed for whatever reason, we'll try the next message
						continue
					}

					if timeSince.Hours()/24 >= 14 {
						break OutOfDate
					}

					purgeList = append(purgeList, msg.ID)
				}
			}
		}

		if err := massDelete(purgeList, s, m); err != nil {
			return err
		}

		purgeAmount -= 100
	}

	return nil
}

func massDelete(list []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	err := s.ChannelMessagesBulkDelete(m.ChannelID, list)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an issue deleting messages :(")
		errorLog.Println("error purging", err)
	}
	return err
}

func getMessageAge(msg *discordgo.Message, s *discordgo.Session, m *discordgo.MessageCreate) (time.Duration, error) {
	then, err := msg.Timestamp.Parse()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an issue deleting messages :(")
		errorLog.Println("time parse error", err)
		return time.Duration(0), err
	}
	return time.Since(then), nil
}
