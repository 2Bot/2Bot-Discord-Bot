package config

type Config struct {
	Game    string `json:"game"`
	Prefix  string `json:"prefix"`
	Token   string `json:"token"`
	OwnerID string `json:"owner_id"`
	URL     string `json:"url"`
	Influx  struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"influx"`

	InDev bool `json:"indev"`

	DiscordPWKey string `json:"discord.pw_key"`

	CurrImg int `json:"curr_img_id"`
	MaxProc int `json:"maxproc"`

	Blacklist []string `json:"blacklist"`
}

var Conf = New()

// New creates a default config with some values set, including game, prefix and indev
func New() *Config {
	return &Config{
		Game:   "!owo help",
		Prefix: "!owo ",
		InDev:  true,
		OwnerID: "149612775587446784",
	}
}
