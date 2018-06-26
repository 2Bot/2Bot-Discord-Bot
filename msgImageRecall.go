package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"

	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"strconv"
	"time"

	"golang.org/x/crypto/blake2b"
)

var imageQueue = make(map[string]*queuedImage)

func init() {
	newCommand("image", 0, false, msgImageRecall).setHelp("Args: [save,recall,delete,list,status] [name]\n\nSave images and recall them at anytime! Everyone gets 8MB of image storage. Any name counts so long theres no `/` in it." +
		"Only you can 'recall' your saved images. There's a review process to make sure nothing illegal is being uploaded but we're fairly relaxed for the most part\n\n" +
		"Example:\n`!owo image save 2B Happy`\n2Bot downloads the image and sends it off for reviewing\n\n" +
		"`!owo image recall 2B Happy`\nIf your image was confirmed, 2Bot will send the image named `2B Happy`\n\n" +
		"`!owo image delete 2B Happy`\nThis will delete the image you saved called `2B Happy`\n\n" +
		"`!owo image list`\nThis will list your saved images along with a preview!\n\n" +
		"`!owo image status`\nShows some details on your saved images and quota").add()
}

func msgImageRecall(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if len(msglist) < 2 {
		prefix, err := activePrefix(m.ChannelID, s)
		if err != nil {
			return
		}

		s.ChannelMessageSend(m.ChannelID,
			"Available sub-commands for `image`:\n`save`, `delete`, `recall`, `list`, `status`\n"+
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
		fimageInfo(s, m, nil)
	}
}

func httpImageRecall(w http.ResponseWriter, r *http.Request) {
	// 404 for user not found, 410 for image not found
	defer r.Body.Close()

	id := chi.URLParam(r, "id")
	img := chi.URLParam(r, "img")

	log.Info(fmt.Sprintf("image request from %s for %s", id, img))

	if val, ok := u[id]; ok {
		for _, val := range val.Images {
			if strings.HasPrefix(val, img) {
				w.WriteHeader(http.StatusOK)
				log.Trace(fmt.Sprintf("user %s has image %s", id, img))
				fmt.Fprint(w, "https://noahsc.xyz/2Bot/images/"+val)
				return
			}
		}
		w.WriteHeader(http.StatusGone)
		log.Trace(fmt.Sprintf("user %s doesn't have image %s", id, img))
		return
	}

	log.Trace(fmt.Sprintf("user %s not in map", id))
	w.WriteHeader(http.StatusNotFound)
}

func fimageRecall(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	var filename string
	if val, ok := u[m.Author.ID]; ok {
		if val, ok := val.Images[strings.Join(msglist, " ")]; ok {
			filename = val
		} else {
			s.ChannelMessageSend(m.ChannelID, "You dont have an image under that name saved with me <:2BThink:333694872802426880>")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "You've no saved images! Get storin'!")
		return
	}

	imgURL := "http://noahsc.xyz/2Bot/images/" + url.PathEscape(filename)

	resp, err := http.Head(imgURL)
	if err != nil {
		log.Error("error recalling image", err)
		s.ChannelMessageSend(m.ChannelID, "Error getting the image :( Please pester my creator about this")
		return
	} else if resp.StatusCode != http.StatusOK {
		log.Error("non 200 status code " + imgURL)
		s.ChannelMessageSend(m.ChannelID, "Error getting the image :( Please pester my creator about this")
		return
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Description: strings.Join(msglist, " "),

		Color: 0x000000,

		Image: &discordgo.MessageEmbedImage{
			URL: imgURL,
		},
	})

	return
}

func fimageSave(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	conf.CurrImg++
	saveConfig()

	currentImageNumber := conf.CurrImg

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
	prefixedImgName := m.Author.ID + "_" + imgName

	fileExtension := strings.ToLower(path.Ext(m.Attachments[0].URL))
	hash := blake2b.Sum256([]byte(prefixedImgName))
	imgFileName := hex.EncodeToString(hash[:]) + fileExtension

	currUser, ok := u[m.Author.ID]
	if !ok {
		u[m.Author.ID] = &user{
			Images:     map[string]string{},
			TempImages: []string{},
			DiskQuota:  8000000,
			QueueSize:  0,
		}
		currUser = u[m.Author.ID]
	}

	_, ok = currUser.Images[imgName]
	//if named image is in queue or already saved, abort
	if isIn(imgName, currUser.TempImages) || ok {
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
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The image file size is too big by %.2fMB :(\n"+
			"Note, this only takes your queued (aka unconfirmed) images into account, so if one of them gets rejected, you can try adding this image again!",
			float32(fileSize+currUser.QueueSize+currUser.CurrDiskUsed-currUser.DiskQuota)/1000/1000))
		return
	}

	dlMsg, _ := s.ChannelMessageSend(m.ChannelID, "<:update:264184209617321984> Downloading your image~")

	resp, err := http.Get(m.Attachments[0].URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "Error downloading the image :( Please pester creator about this")
		log.Error("error downloading image ", err)
		return
	}
	defer resp.Body.Close()

	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		guild = &discordgo.Guild{
			Name: "error",
			ID:   "error",
		}
	}

	tempFilepath := "images/temp/" + imgFileName

	//create temp file in temp path
	tempFile, err := os.Create(tempFilepath)
	if err != nil {
		log.Error("error creating temp file", err)
		s.ChannelMessageSend(m.ChannelID, "There was an error saving the image :( Please pester my creator about this")
		return
	}
	defer tempFile.Close()

	bodyImg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error parsing body", err)
		return
	}

	if _, err = tempFile.Write(bodyImg); err != nil {
		log.Error("error writing image to file", err)
		return
	}

	_, err = s.ChannelMessageEdit(m.ChannelID, dlMsg.ID, m.Author.Mention()+" Thanks for the submission! "+
		"Your image is being reviewed by our ~~lazy~~ hard-working review team! You'll get a PM from either my master himself or from me once its been confirmed or rejected :) Sit tight!")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Thanks for the submission! "+
			"Your image is being reviewed by our ~~lazy~~ hard-working review team! You'll get a PM from either my master himself or from me once its been confirmed or rejected :) Sit tight!")
		deleteMessage(dlMsg, s)
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
		log.Error("error attaching reaction", err)
	}
	err = s.MessageReactionAdd(reviewMsg.ChannelID, reviewMsg.ID, "❌")
	if err != nil {
		s.ChannelMessageSend(reviewChan, "Couldn't add ❌ to message")
		log.Error("error attaching reaction", err)
	}

	currUser.TempImages = append(currUser.TempImages, imgName)
	currUser.QueueSize += fileSize

	imageQueue[strconv.Itoa(currentImageNumber)] = &queuedImage{
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

	fimageReview(s, currentImageNumber)
}

func fimageReview(s *discordgo.Session, currentImageNumber int) {
	imgInQueue := imageQueue[strconv.Itoa(currentImageNumber)]

	fileSize := imgInQueue.FileSize
	prefixedImgName := imgInQueue.AuthorID + "_" + imgInQueue.ImageName

	fileExtension := strings.ToLower(path.Ext(imgInQueue.ImageURL))
	hash := blake2b.Sum256([]byte(prefixedImgName))
	imgFileName := hex.EncodeToString(hash[:]) + fileExtension
	tempFilepath := "images/temp/" + imgFileName
	currUser := u[imgInQueue.AuthorID]

	//Wait here for a relevant reaction to the confirmation message
	for {
		confirm := <-nextReactionAdd(s)
		if confirm.UserID == s.State.User.ID || confirm.MessageID != imgInQueue.ReviewMsgID {
			continue
		}

		user, err := userDetails(confirm.UserID, s)
		if err != nil {
			s.ChannelMessageSend(reviewChan, "Error getting user for image confirming")
			continue
		}

		if confirm.MessageReaction.Emoji.Name == "✅" {
			//IF CONFIRMED
			s.ChannelMessageSend(reviewChan, fmt.Sprintf("%s confirmed image `%s` from `%s#%s` ID: `%s`",
				func() string {
					if user != nil {
						return user.Username
					}
					return confirm.UserID
				}(),
				imgInQueue.ImageName,
				imgInQueue.AuthorName,
				imgInQueue.AuthorDiscrim,
				imgInQueue.AuthorID))

			break
		} else if confirm.MessageReaction.Emoji.Name == "❌" {
			//IF REJECTED
			s.ChannelMessageSend(reviewChan, fmt.Sprintf("%s rejected image `%s` from `%s#%s` ID: `%s`\nGive a reason next! Enter `None` to give no reason",
				func() string {
					if user != nil {
						return user.Username
					}
					return confirm.UserID
				}(),
				imgInQueue.ImageName,
				imgInQueue.AuthorName,
				imgInQueue.AuthorDiscrim,
				imgInQueue.AuthorID))

			var reason string
			for {
				rejectMsg := <-nextMessageCreate(s)
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
					delete(imageQueue, strconv.Itoa(currentImageNumber))

					saveUsers()
					saveQueue()

					if err := os.Remove(tempFilepath); err != nil {
						log.Error("error deleting temp image", err)
						s.ChannelMessageSend(reviewChan, "Error deleting temp image")
					}

					//Make PM channel to inform user that image was rejected
					channel, err := s.UserChannelCreate(imgInQueue.AuthorID)
					//Couldnt make PM channel
					if err != nil {
						s.ChannelMessageSend(reviewChan, fmt.Sprintf("Couldn't inform %s#%s ID: %s about rejection\n%s", imgInQueue.AuthorName, imgInQueue.AuthorDiscrim, imgInQueue.AuthorID, err))
						return
					}

					//Try PMing
					if _, err = s.ChannelMessageSend(channel.ID, "Your image got rejected :( Sorry\n"+reason); err != nil {
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
		s.ChannelMessageSend(reviewChan, fmt.Sprintf("Couldn't inform %s#%s ID: %s about confirmation\n%s", imgInQueue.AuthorName, imgInQueue.AuthorDiscrim, imgInQueue.AuthorID, err))
	}

	filepath := "images/" + imgFileName

	if err := os.Rename(tempFilepath, filepath); err != nil {
		s.ChannelMessageSend(reviewChan, "Error moving file from temp dir")
		log.Error("error moving file from temp dir", err)
	} else {
		if err := os.Chmod(filepath, 0755); err != nil {
			s.ChannelMessageSend(reviewChan, "Can't chmod "+err.Error())
			log.Error("cant chmod", err)
		}
	}

	delete(imageQueue, strconv.Itoa(currentImageNumber))
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
	if val, ok := u[m.Author.ID]; ok {
		if val, ok := val.Images[strings.Join(msglist, " ")]; ok {
			filename = val
		} else {
			s.ChannelMessageSend(m.ChannelID, "You dont have an image under that name saved with me <:2BThink:333694872802426880>")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "You've no saved images! Get storin'!")
		return
	}

	err := os.Remove("images/" + filename)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Image couldnt be deleted :( Please pester my creator for me")
		log.Error("error deleting image", err)
		return
	}

	delete(u[m.Author.ID].Images, strings.Join(msglist, " "))

	saveUsers()

	s.ChannelMessageSend(m.ChannelID, "Image deleted~")
}

func fimageList(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	val, ok := u[m.Author.ID]
	if (ok && len(u[m.Author.ID].Images) == 0) || !ok {
		s.ChannelMessageSend(m.ChannelID, "You've no saved images! Get storin'!")
		return
	}

	var out []string
	var files []string
	for key, value := range val.Images {
		out = append(out, key)
		files = append(files, value)
	}

	msg, err := s.ChannelMessageSend(m.ChannelID, "Assemblin' a preview your images!")

	p := dgwidgets.NewPaginator(s, m.ChannelID)

	success := true
	for i, img := range files {
		imgURL, err := url.Parse("http://noahsc.xyz/2Bot/images/" + url.PathEscape(img))
		if err != nil {
			log.Error("error parsing img url", err)
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
	p.Loop = true
	p.ColourWhenDone = 0xff0000
	p.DeleteReactionsWhenDone = true
	p.Widget.Timeout = time.Minute * 2

	if err != nil {
		s.ChannelMessageEdit(m.ChannelID, msg.ID, "Your saved images are: `"+strings.Join(out, ", "))
	}

	if err := p.Spawn(); err != nil {
		log.Error("error creating image list", err)
		s.ChannelMessageSend(m.ChannelID, "Couldn't make the list :( Go pester Strum355#1180 about this")
		return
	}

	if !success {
		s.ChannelMessageSend(m.ChannelID, "I couldn't assemble all of your images, but here are the ones i could get!")
	}
}

func fimageInfo(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	if val, ok := u[m.Author.ID]; ok {
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

		return
	}

	u[m.Author.ID] = &user{
		Images:     map[string]string{},
		TempImages: []string{},
		DiskQuota:  8000000,
		QueueSize:  0,
	}

	saveUsers()

	fimageInfo(s, m, msglist)
}
