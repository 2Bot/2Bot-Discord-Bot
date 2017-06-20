package main

type server struct {
	Nsfw bool `json:"nsfw"`
	//ID string `json:"server_id"`
	LogChannel string `json:"log_channel"`
	Prefix string `json:"server_prefix"`
	Log bool `json:"log_active"`
	Kicked bool `json:"kicked"`
	//Enabled, Message, Channel
	JoinMessage []string `json:"join"`
}

type ibStruct struct {
	Path string `json:"path"`
	Server string `json:"server"`
}

type config struct {
	Game string `json:"game"`
	Prefix string `json:"prefix"`
	Servers map[string]*server 
} 

type rule34 struct {
	PostCount  int `xml:"count,attr"`
	Posts	   []struct {
		URL string `xml:"file_url,attr"`
	} `xml:"post"`	
}