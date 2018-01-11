package main

import (
	"encoding/json"
	"os"
)

var (
	c    = new(config)
	u    = new(users)
	q    = new(imageQueue)
	sMap = new(servers)
)

func saveJSON(path string, data interface{}) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		errorLog.Printf("Error saving %s: %s\n", path, err)
		return err
	}

	if err = json.NewEncoder(f).Encode(data); err != nil {
		errorLog.Printf("Error saving %s: %s\n", path, err)
		return err
	}
	return nil
}

func loadJSON(path string, v interface{}) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		errorLog.Printf("Error loading %s: %s\n", path, err)
		return err
	}

	if err := json.NewDecoder(f).Decode(v); err != nil {
		errorLog.Printf("Error loading %s: %s\n", path, err)
		return err
	}
	return nil
}

func cleanup() {
	saveConfig()
	saveQueue()
	saveServers()
	saveUsers()
}

func loadConfig() error {
	return loadJSON("config.json", c)
}

func saveConfig() error {
	return saveJSON("config.json", c)
}

func loadServers() error {
	sMap.Server = make(map[string]*server)
	return loadJSON("servers.json", sMap)
}

func saveServers() error {
	return saveJSON("servers.json", sMap)
}

func loadUsers() error {
	u.User = make(map[string]*user)
	return loadJSON("users.json", u)
}

func saveUsers() error {
	return saveJSON("users.json", u)
}

func loadQueue() error {
	q.QueuedMsgs = make(map[string]*queuedImage)
	return loadJSON("queue.json", q)
}

func saveQueue() error {
	return saveJSON("queue.json", q)
}
