package main

type queuedImage struct {
	ReviewMsgID   string `json:"reviewMsgID"`
	AuthorID      string `json:"author_id"`
	AuthorDiscrim string `json:"author_discrim"`
	AuthorName    string `json:"author_name"`
	ImageName     string `json:"image_name"`
	ImageURL      string `json:"image_url"`

	FileSize int `json:"file_size"`
}

type users map[string]*user

type user struct {
	Images map[string]string `json:"images"`

	DiskQuota    int `json:"quota"`
	CurrDiskUsed int `json:"curr_used"`
	QueueSize    int `json:"queue_size"`

	TempImages []string `json:"temp_images"`
}
