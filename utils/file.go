package utils

import (
	"encoding/json"
	"errors"
	"io"
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

func SaveMediaFile(savePath string, saveFileName string, saveContent []byte) error {
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

// LoadURLsFromJSON 从指定的JSON文件中读取URLs
func LoadURLsFromJSON(filePath string) ([]string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取文件内容
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 解析JSON到字符串切片
	var urls []string
	if err := json.Unmarshal(bytes, &urls); err != nil {
		return nil, err
	}

	return urls, nil
}
