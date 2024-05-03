package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

var (
	settingsCache map[string]interface{}
	settingsMutex sync.Mutex
)

func ReadSettings() (map[string]interface{}, error) {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()

	if settingsCache != nil {
		return settingsCache, nil
	}

	file, err := os.Open("setting.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	settings := make(map[string]interface{})
	err = decoder.Decode(&settings)
	if err != nil {
		return nil, err
	}

	settingsCache = settings
	return settings, nil
}

func SaveFile(savePath string, saveFileName string, saveContent []byte) error {
	fileName := saveFileName
	dir := savePath

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Println("Error creating directory:", err)
			return errors.New("error creating directory")
		}
	}
	file, err := os.Create(dir + fileName)
	if err != nil {
		log.Println("Error creating file:", err)
		return errors.New("error creating file")
	}
	defer file.Close()

	_, err = file.Write(saveContent)
	if err != nil {
		log.Println("Error saving file:", err)
		return errors.New("error saving file")
	}
	return nil
}

func ReadLog() (map[string]interface{}, error) {
	file, err := os.Open("login.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	settings := make(map[string]interface{})
	err = decoder.Decode(&settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func SaveLog(saveContent map[string]interface{}) error {
	file, err := os.Create("login.json")
	if err != nil {
		log.Println("Error creating file:", err)
		return errors.New("error creating file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(saveContent)
	if err != nil {
		log.Println("Error saving file:", err)
		return errors.New("error saving file")
	}
	return nil
}

func WriteBytesToFile(filename string, data []byte) {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
	}
}