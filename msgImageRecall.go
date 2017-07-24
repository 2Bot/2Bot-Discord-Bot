package main

import (
	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"strings"

	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"strconv"
	"time"
)

//From Necroforger's dgwidgets
func nextMessageReactionAddC(s *discordgo.Session) chan *discordgo.MessageReactionAdd {
	out := make(chan *discordgo.MessageReactionAdd)
	s.AddHandlerOnce(func(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
		out <- e
	})
	return out
}

func nextMessageCreateC(s *discordgo.Session) chan *discordgo.MessageCreate {
	out := make(chan *discordgo.MessageCreate)
	s.AddHandlerOnce(func(_ *discordgo.Session, e *discordgo.MessageCreate) {
		out <- e
	})
	return out
}

func msgImageRecall(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {

	if len(msglist) < 2 {
		prefix := c.Prefix
		guild, err := guildDetails(m.ChannelID, s)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "There was an issue recalling your image :( Try again please~")
			errorLog.Println("image recall guild details error", err.Error())
			return
		} else if sMap.Server[guild.ID].Prefix != "" {
			prefix = sMap.Server[guild.ID].Prefix
		}

		s.ChannelMessageSend(m.ChannelID, "Available sub-commands for `image`:\n`save`, `delete`, `recall`, `list`, `status`\n"+
			"Type `"+prefix+"help image` to see more info about this command")
		return
	}

	switch msglist[1] {
	case "recall":
		fimageRecall(s, m, msglist[2:])
	case "save":
		fimageSave(s, m, msglist[2:])
	case "delete":
		fimageDelete(s, m, msglist[2:])
	case "list":
		fimageList(s, m, nil)
	case "status":
		fimageInfo(s, m, nil, false)
	}
}

func fimageRecall(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	var filename string
	if val, ok := u.User[m.Author.ID]; ok {
		if isInMap(strings.Join(msglist, " "), val.Images) {
			filename = val.Images[strings.Join(msglist, " ")]
		} else {
			s.ChannelMessageSend(m.ChannelID, "You dont have an image under that name saved with me <:2BThink:333694872802426880>")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "You've no saved images! Get storin'!")
		return
	}

	escapedFile := url.PathEscape(filename)
	imgURL, err := url.Parse("https://sushishader.eu/2Bot/images/" + escapedFile)
	if err != nil {
		errorLog.Println("Error parsing img url", err.Error())
		s.ChannelMessageSend(m.ChannelID, "Error getting the image :( Please pester Strum355#1180 about this")
		return
	}

	resp, err := http.Head(imgURL.String())
	if err != nil {
		errorLog.Println("Error recalling image", err.Error())
		s.ChannelMessageSend(m.ChannelID, "Error getting the image :( Please pester Strum355#1180 about this")
		return
	} else if resp.StatusCode != http.StatusOK {
		errorLog.Println("Non 200 status code")
		s.ChannelMessageSend(m.ChannelID, "Error getting the image :( Please pester Strum355#1180 about this")
		return
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Description: strings.Join(msglist, " "),

		Color: 0x000000,

		Image: &discordgo.MessageEmbedImage{
			URL: imgURL.String(),
		},
	})
}

func fimageSave(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	c.CurrImg++
	saveConfig()
	currentImageNumber := c.CurrImg

	if len(m.Attachments) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No image sent. Please send me an image to save for you!")
		return
	}

	if len(msglist) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Gotta name your image!")
		return
	}

	if m.Attachments[0].Height == 0 {
		s.ChannelMessageSend(m.ChannelID, "Either your image is corrupted or you didn't send me an image <:2BThink:333694872802426880> I can only save images for you~")
		return
	}

	imgName := strings.Join(msglist, " ")
	isLegalFileName := fileNameRegex.MatchString(imgName)

	if isLegalFileName {
		s.ChannelMessageSend(m.ChannelID, "I can't use that as a file name <:2BThink:333694872802426880> Please no forward-slash in my christian file names!")
		return
	}

	prefixedImgName := m.Author.ID + "_" + imgName
	fileExtension := strings.ToLower(path.Ext(m.Attachments[0].URL))
	imgFileName := prefixedImgName + fileExtension

	if _, ok := u.User[m.Author.ID]; !ok {
		u.User[m.Author.ID] = &user{
			Images:     map[string]string{},
			TempImages: []string{},
			DiskQuota:  8000000,
			QueueSize:  0,
		}
	}

	currUser := u.User[m.Author.ID]

	//if named image is in queue or already saved, abort
	if isIn(imgName, currUser.TempImages) || isInMap(imgName, currUser.Images) {
		s.ChannelMessageSend(m.ChannelID, "You've already saved an image under that name! Delete it first~")
		return
	}

	fileSize := m.Attachments[0].Size

	//if the image + current used space > quota
	if fileSize+currUser.CurrDiskUsed > currUser.DiskQuota {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The image file size is too big by %.2fMB :(",
			float32(fileSize+currUser.CurrDiskUsed-currUser.DiskQuota)/1000/1000))
		return
	}

	//if when the image is added to the queue, the queue size + current used space > quota
	if fileSize+currUser.QueueSize+currUser.CurrDiskUsed > currUser.DiskQuota {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The image file size is too big by %.2fMB :(\nNote, this only takes your queued (aka unconfirmed) images into account, so if one of them gets rejected, you can try adding this image again!",
			float32(fileSize+currUser.QueueSize+currUser.CurrDiskUsed-currUser.DiskQuota)/1000/1000))
		return
	}

	dlMsg, _ := s.ChannelMessageSend(m.ChannelID, "<:update:264184209617321984> Downloading your image~")

	resp, err := http.Get(m.Attachments[0].URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "Error downloading the image :( Please pester Strum355#1180 about this")
		errorLog.Println("Error downloading image ", err.Error())
		return
	}
	defer resp.Body.Close()

	guild, err := guildDetails(m.ChannelID, s)
	if err != nil {
		errorLog.Println("image save guild details error", err.Error())
	}

	tempFilepath := "../../public_html/2Bot/images/temp/" + imgFileName

	//create temp file in temp path
	tempFile, err := os.Create(tempFilepath)
	if err != nil {
		errorLog.Println("Error creating temp file", err.Error())
		s.ChannelMessageSend(m.ChannelID, "There was an error saving the image :( Please pester Strum355#1180 about this")
		return
	}
	defer tempFile.Close()

	bodyImg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorLog.Println("Error parsing body", err.Error())
		return
	}

	_, err = tempFile.Write(bodyImg)
	if err != nil {
		errorLog.Println("Error writing image to file", err.Error())
		return
	}

	_, err = s.ChannelMessageEdit(m.ChannelID, dlMsg.ID, fmt.Sprintf("%s Thanks for the submission! Your image is being reviewed by our ~~lazy~~ hard-working review team! You'll get a PM from either my master himself or from me once its been confirmed or rejected :) Sit tight!", m.Author.Mention()))
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Thanks for the submission! Your image is being reviewed by our ~~lazy~~ hard-working review team! You'll get a PM from either my master himself or from me once its been confirmed or rejected :) Sit tight!", m.Author.Mention()))
		s.ChannelMessageDelete(m.ChannelID, dlMsg.ID)
	}

	reviewMsg, _ := s.ChannelMessageSendEmbed(reviewChan, &discordgo.MessageEmbed{
		Description: fmt.Sprintf("Image ID: %d\nNew image from:\n`%s#%s` ID: %s\nfrom server `%s` `%s`\nnamed `%s`",
			currentImageNumber,
			m.Author.Username,
			m.Author.Discriminator,
			m.Author.ID,
			guild.Name,
			guild.ID,
			imgName),

		Color: 0x000000,

		Image: &discordgo.MessageEmbedImage{
			URL: m.Attachments[0].URL,
		},
	})

	err = s.MessageReactionAdd(reviewMsg.ChannelID, reviewMsg.ID, "✅")
	if err != nil {
		s.ChannelMessageSend(reviewChan, "Couldn't add ✅ to message")
		errorLog.Println("Error attaching reaction", err.Error())
	}
	err = s.MessageReactionAdd(reviewMsg.ChannelID, reviewMsg.ID, "❌")
	if err != nil {
		s.ChannelMessageSend(reviewChan, "Couldn't add ❌ to message")
		errorLog.Println("Error attaching reaction", err.Error())
	}

	currUser.TempImages = append(currUser.TempImages, imgName)
	currUser.QueueSize += fileSize

	q.QueuedMsgs[strconv.Itoa(currentImageNumber)] = &queuedImage{
		ReviewMsgID:   reviewMsg.ID,
		AuthorID:      m.Author.ID,
		AuthorDiscrim: m.Author.Discriminator,
		AuthorName:    m.Author.Username,
		ImageName:     imgName,
		ImageURL:      m.Attachments[0].ProxyURL,
		FileSize:      fileSize,
	}

	saveQueue()
	saveUsers()

	fimageReview(s, q, currentImageNumber)
}

func fimageReview(s *discordgo.Session, queue *imageQueue, currentImageNumber int) {
	imgInQueue := queue.QueuedMsgs[strconv.Itoa(currentImageNumber)]

	fileSize := imgInQueue.FileSize
	prefixedImgName := imgInQueue.AuthorID + "_" + imgInQueue.ImageName
	fileExtension := strings.ToLower(path.Ext(imgInQueue.ImageURL))
	imgFileName := prefixedImgName + fileExtension
	tempFilepath := "../../public_html/2Bot/images/temp/" + imgFileName
	currUser := u.User[imgInQueue.AuthorID]

	//Wait here for a relevant reaction to the confirmation message
	for {
		confirm := <-nextMessageReactionAddC(s)
		if confirm.UserID == s.State.User.ID || confirm.MessageID != imgInQueue.ReviewMsgID {
			continue
		}

		if confirm.MessageReaction.Emoji.Name == "✅" {
			//IF CONFIRMED
			s.ChannelMessageSend(reviewChan, fmt.Sprintf("Confirmed image `%s` from `%s#%s` ID: `%s`",
				imgInQueue.ImageName,
				imgInQueue.AuthorName,
				imgInQueue.AuthorDiscrim,
				imgInQueue.AuthorID))

			break
		} else if confirm.MessageReaction.Emoji.Name == "❌" {
			//IF REJECTED
			s.ChannelMessageSend(reviewChan, fmt.Sprintf("Rejected image `%s` from `%s#%s` ID: `%s`\nGive a reason next! Enter `None` to give no reason",
				imgInQueue.ImageName,
				imgInQueue.AuthorName,
				imgInQueue.AuthorDiscrim,
				imgInQueue.AuthorID))

			var reason string
			for {
				rejectMsg := <-nextMessageCreateC(s)
				if rejectMsg.Author.ID == confirm.UserID {
					rejectMsgList := strings.Fields(rejectMsg.Content)
					if len(rejectMsgList) < 1 || rejectMsgList[0] != strconv.Itoa(currentImageNumber) {
						continue
					}
					if strings.Join(rejectMsgList[1:], " ") != "None" {
						reason = "Reason: " + strings.Join(rejectMsgList[1:], " ")
					}

					currUser.TempImages = remove(currUser.TempImages, findIndex(currUser.TempImages, imgInQueue.ImageName))
					currUser.QueueSize -= fileSize
					delete(q.QueuedMsgs, strconv.Itoa(currentImageNumber))

					saveUsers()
					saveQueue()

					err := os.Remove(tempFilepath)
					if err != nil {
						errorLog.Println("Error deleting temp image", err.Error())
					}

					//Make PM channel to inform user that image was rejected
					channel, err := s.UserChannelCreate(imgInQueue.AuthorID)
					//Couldnt make PM channel
					if err != nil {
						s.ChannelMessageSend(reviewChan, fmt.Sprintf("Couldn't inform %s#%s ID: %s about rejection\n%s", imgInQueue.AuthorName, imgInQueue.AuthorDiscrim, imgInQueue.AuthorID, err))
						return
					}

					//Try PMing
					_, err = s.ChannelMessageSend(channel.ID, "Your image got rejected :( Sorry\n"+reason)
					//Couldn't PM
					if err != nil {
						s.ChannelMessageSend(reviewChan, fmt.Sprintf("Couldn't inform %s#%s ID: %s about rejection\n%s", imgInQueue.AuthorName, imgInQueue.AuthorDiscrim, imgInQueue.AuthorID, err))
					}

					s.ChannelMessageSend(reviewChan, fmt.Sprintf("Reason for image `%s` from `%s#%s` ID: `%s`\n%s",
						imgInQueue.ImageName,
						imgInQueue.AuthorName,
						imgInQueue.AuthorDiscrim,
						imgInQueue.AuthorID,
						reason))
					return
				}
			}
		}
	}

	//If image has been reviewed and confirmed
	channel, err := s.UserChannelCreate(imgInQueue.AuthorID)
	if err != nil {
		s.ChannelMessageSend("334092230845267988", fmt.Sprintf("Couldn't inform %s#%s ID: %s about confirmation\n%s", imgInQueue.AuthorName, imgInQueue.AuthorDiscrim, imgInQueue.AuthorID, err))
	}

	filepath := "../../public_html/2Bot/images/" + imgFileName

	os.Rename(tempFilepath, filepath)
	err = os.Chmod(filepath, 655)
	if err != nil {
		s.ChannelMessageSend(reviewChan, "Can't chmod "+err.Error())
		errorLog.Println("Cant chmod", err.Error())
	}

	delete(q.QueuedMsgs, strconv.Itoa(currentImageNumber))
	currUser.TempImages = remove(currUser.TempImages, findIndex(currUser.TempImages, imgInQueue.ImageName))
	currUser.CurrDiskUsed += fileSize
	currUser.QueueSize -= fileSize
	currUser.Images[imgInQueue.ImageName] = imgFileName

	saveQueue()
	saveUsers()

	s.ChannelMessageSend(channel.ID, "Your image was confirmed and is now saved :D To \"recall\" it, type `[prefix] image recall "+imgInQueue.ImageName+"`")
}

func fimageDelete(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	var filename string
	if val, ok := u.User[m.Author.ID]; ok {
		if isInMap(strings.Join(msglist, " "), val.Images) {
			filename = val.Images[strings.Join(msglist, " ")]
		} else {
			s.ChannelMessageSend(m.ChannelID, "You dont have an image under that name saved with me <:2BThink:333694872802426880>")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "You've no saved images! Get storin'!")
	}

	err := os.Remove("../../public_html/2Bot/images/" + filename)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Image couldnt be deleted :( Please pester Strum355#1180 for me")
		errorLog.Println("Error deleting image", err.Error())
		return
	}

	delete(u.User[m.Author.ID].Images, strings.Join(msglist, " "))

	saveUsers()

	s.ChannelMessageSend(m.ChannelID, "Image deleted~")
}

func fimageList(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if val, ok := u.User[m.Author.ID]; ok {
		if len(u.User[m.Author.ID].Images) == 0 {
			s.ChannelMessageSend(m.ChannelID, "You've no saved images! Get storin'!")
			return
		}

		var out []string
		var files []string
		//var usedSpace float32
		for key, value := range val.Images {
			out = append(out, key)
			files = append(files, value)
		}

		s.ChannelMessageSend(m.ChannelID, "Your saved images are: `"+strings.Join(out, ", ")+
			"`\nAssemblin' a preview your images! **Don't click the reactions until all 5 are there!** Blame Discords rate-limit...")

		p := dgwidgets.NewPaginator(s, m.ChannelID)

		success := true
		for i, img := range files {
			escapedFile := url.PathEscape(img)
			imgURL, err := url.Parse("https://sushishader.eu/2Bot/images/" + escapedFile)
			if err != nil {
				errorLog.Println("Error parsing img url", err.Error())
				success = false
				continue
			}
			p.Add(&discordgo.MessageEmbed{
				Description: out[i],
				Image: &discordgo.MessageEmbedImage{
					URL: imgURL.String(),
				},
			})
		}

		p.SetPageFooters()

		p.ColourWhenDone = 0xff0000
		p.Loop = true
		p.DeleteReactionsWhenDone = true

		p.Widget.Timeout = time.Minute * 5

		err := p.Spawn()
		if err != nil {
			errorLog.Println("Error creating image list", err.Error())
			s.ChannelMessageSend(m.ChannelID, "Couldn't make the list :( Go pester Strum355#1180 about this")
			return
		}

		if !success {
			s.ChannelMessageSend(m.ChannelID, "I couldn't assemble all of your images, but here are the ones i could get!")
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "You've no saved images! Get storin'!")
	}
}

func fimageInfo(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string, called bool) {
	if val, ok := u.User[m.Author.ID]; ok {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```autohotkey\nTotal Images:%21d```"+
			"```autohotkey\nTotal Space Used:%20.2f/%.2fMB (%.2f/%.2fKB)```"+
			"```autohotkey\nQueued Images:%20d```"+
			"```autohotkey\nQueued Disk Space:%19.2f/%.2fMB (%.2f/%.2fKB)```"+
			"```autohotkey\nFree Space:%26.2fMB (%.2fKB)```",
			len(val.Images),
			float32(val.CurrDiskUsed)/1000/1000,
			float32(val.DiskQuota)/1000/1000,
			float32(val.CurrDiskUsed)/1000,
			float32(val.DiskQuota)/1000,
			len(val.TempImages),
			float32(val.QueueSize)/1000/1000,
			float32(val.DiskQuota)/1000/1000,
			float32(val.QueueSize)/1000,
			float32(val.DiskQuota)/1000,
			float32(val.DiskQuota-(val.QueueSize+val.CurrDiskUsed))/1000/1000,
			float32(val.DiskQuota-(val.QueueSize+val.CurrDiskUsed))/1000))
	} else {
		//If function has been called recursively but
		//u.User[m.Author.ID] doesnt exist yet,
		//something went wrong, so abort
		if called {
			return
		}

		u.User[m.Author.ID] = &user{
			Images:     map[string]string{},
			TempImages: []string{},
			DiskQuota:  8000000,
			QueueSize:  0,
		}

		saveUsers()

		fimageInfo(s, m, msglist, true)
	}
}
