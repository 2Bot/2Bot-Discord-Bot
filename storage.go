package main

import (
	"github.com/2Bot/2Bot-Discord-Bot/metrics"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/2Bot/2Bot-Discord-Bot/config"
)

var (
	u    = make(users)
	sMap = newServers()
)

func saveJSON(path string, data interface{}) error {
	f, err := os.OpenFile("json/"+path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Error("error saving", path, err)
		return err
	}
	defer f.Close()

	var b []byte
	if _, ok := data.(*config.Config); ok {
		b, err = json.MarshalIndent(&data, "", "  ")
	} else {
		b, err = json.Marshal(&data)
	}

	if err != nil {
		log.Error("error saving", path, err)
		return err
	}

	if _, err = f.Write(b); err != nil {
		log.Error("error saving", path, err)
		return err
	}

	return nil
}

func loadJSON(path string, v interface{}) error {
	f, err := os.OpenFile("json/"+path, os.O_RDONLY, 0600)
	if err != nil {
		log.Error("error loading", path, err)
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error("error loading", path, err)
		return err
	}

	if err = json.Unmarshal(b, &v); err != nil {
		log.Error("error loading", path, err)
		return err
	}

	return nil
}

func cleanup() {
	for _, f := range []func() error{saveConfig, saveQueue, saveServers, saveUsers} {
		if err := f(); err != nil {
			log.Error("error cleaning up files", err)
		}
	}

	metrics.InfluxClient.Close()

	log.Info("Done cleanup. Exiting.")
}

func loadConfig() error {
	return loadJSON("config.json", conf)
}

func saveConfig() error {
	return saveJSON("config.json", conf)
}

func loadServers() error {
	sMap = newServers()
	return loadJSON("servers.json", &sMap.serverMap)
}

func saveServers() error {
	return saveJSON("servers.json", &sMap.serverMap)
}

func loadUsers() error {
	u = make(map[string]*user)
	return loadJSON("users.json", &u)
}

func saveUsers() error {
	return saveJSON("users.json", &u)
}

func loadQueue() error {
	return loadJSON("queue.json", &imageQueue)
}

func saveQueue() error {
	return saveJSON("queue.json", &imageQueue)
}
