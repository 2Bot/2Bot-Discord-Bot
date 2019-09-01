package config

import (
	"github.com/spf13/viper"
)

func loadDefaults() {
	viper.SetDefault("bot.game", "!xd help")
	viper.SetDefault("bot.prefix", "!xd ")
	viper.SetDefault("bot.token", "invalid")
	viper.SetDefault("bot.owner", "1234567890")

	viper.SetDefault("dev", true)
	viper.SetDefault("s3.url", "minio:9000")
	viper.SetDefault("s3.access_key", "sample_text")
	viper.SetDefault("s3.secret_key", "sample_text")
}
