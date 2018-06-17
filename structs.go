package main

import (
	"github.com/bwmarrin/discordgo"
)

type config struct {
	Game   string `json:"game"`
	Prefix string `json:"prefix"`
	Token  string `json:"token"`

	InDev bool `json:"indev"`

	DiscordPWKey string `json:"discord.pw_key"`

	CurrImg int `json:"curr_img_id"`
	MaxProc int `json:"maxproc"`

	Blacklist []string `json:"blacklist"`
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

	NoahOnly      bool
	RequiresPerms bool

	PermsRequired int

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
