package main

import "github.com/bwmarrin/discordgo"

type ibStruct struct {
	Path   string `json:"path"`
	Server string `json:"server"`
}

type rule34 struct {
	PostCount int `xml:"count,attr"`

	Posts []struct {
		URL string `xml:"file_url,attr"`
	} `xml:"post"`
}

type config struct {
	Game   string `json:"game"`
	Prefix string `json:"prefix"`

	InDev bool `json:"indev"`

	DiscordPWKey string `json:"discord.pw_key"`

	CurrImg int `json:"curr_img_id"`
}

type servers struct {
	Server map[string]*server
}

type server struct {
	LogChannel string `json:"log_channel"`
	Prefix     string `json:"server_prefix"`

	Log    bool `json:"log_active"`
	Kicked bool `json:"kicked"`
	Nsfw   bool `json:"nsfw"`

	//Enabled, Message, Channel
	JoinMessage [3]string `json:"join"`

	VoiceInst struct {
		ChannelID string
		Queue []song
		Done chan error
	}

	Playlist map[string][]song
}

type song struct {
	DownloadURL string `json:"url"`
	Name        string `json:"name"`
}

type imageQueue struct {
	QueuedMsgs map[string]*queuedImage
}

type queuedImage struct {
	ReviewMsgID   string `json:"reviewMsgID"`
	AuthorID      string `json:"author_id"`
	AuthorDiscrim string `json:"author_discrim"`
	AuthorName    string `json:"author_name"`
	ImageName     string `json:"image_name"`
	ImageURL      string `json:"image_url"`

	FileSize int `json:"file_size"`
}

type command struct {
	Name string
	Help string

	NoahOnly  bool
	AdminOnly bool

	Exec func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

type users struct {
	User map[string]*user
}

type user struct {
	Images map[string]string `json:"images"`

	DiskQuota    int `json:"quota"`
	CurrDiskUsed int `json:"curr_used"`
	QueueSize    int `json:"queue_size"`

	TempImages []string `json:"temp_images"`
}
