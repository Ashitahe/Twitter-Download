package config

import (
	"encoding/json"
	"log"
	"os"

	"twitterDownload/pkg/storage"
)

type Settings struct {
	Cookie      string `json:"cookie"`
	UserList	 	[]string `json:"userList"`
}

var SettingConfig Settings

var LogRecord = storage.NewURLStore("log.json")

func init(){
	LogRecord.LoadFromFile()
	SettingConfig = LoadSettings("setting.json")
}

func LoadSettings(settingsFilePath string) Settings {

	var settings Settings

	data, err := os.ReadFile(settingsFilePath)
	if err != nil {
		log.Fatalf("Error reading settings file: %v", err)
	}
	err = json.Unmarshal(data, &settings)
	if err != nil {
		log.Fatalf("Error parsing settings JSON: %v", err)
	}

	return settings
}

