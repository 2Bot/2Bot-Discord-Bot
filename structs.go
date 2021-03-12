package main

type config struct {
	Game    string `json:"game"`
	Prefix  string `json:"prefix"`
	Token   string `json:"token"`
	OwnerID string `json:"owner_id"`
	URL     string `json:"url"`

	InDev bool `json:"indev"`

	DiscordPWKey string `json:"discord.pw_key"`

	CurrImg int `json:"curr_img_id"`
	MaxProc int `json:"maxproc"`

	Blacklist []string `json:"blacklist"`
}
