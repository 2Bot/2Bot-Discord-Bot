package main

import (
	"sync"
)

type servers struct {
	Count      int
	VoiceInsts int

	Mutex sync.RWMutex

	Server map[string]*server
}

func (s *servers) getCount() int {
	//not gonna mutex this because am i really gonna cry over
	//an inaccurate result for a single request? 
	return s.Count
}

type server struct {
	LogChannel string `json:"log_channel"`
	Prefix     string `json:"server_prefix"`

	Log    bool `json:"log_active"`
	Kicked bool `json:"kicked"`
	Nsfw   bool `json:"nsfw"`

	//Enabled, Message, Channel
	JoinMessage [3]string `json:"join"`

	VoiceInst voiceInst `json:"-"`

	Playlists map[string][]song `json:"playlists"`
}
