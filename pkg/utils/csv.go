package utils

import (
	"encoding/csv"
	"os"
)

type CSV struct {
	TweetDate   string
	TweetId     string
	Username    string
	DisplayName string
	TweetText   string
	TweetURL    string
	MediaType   string
	MediaURL    string
}

func SaveToCSV(records []CSV, filePath string) error {
	// 检查文件是否存在来决定是否需要写入头部
	var writeHeaders bool
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		writeHeaders = true
	}

	// 使用os.OpenFile以追加模式打开或创建文件
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果需要，写入头部
	if writeHeaders {
		headers := []string{"TweetDate", "TweetId", "Username", "DisplayName", "TweetText", "TweetURL", "MediaType", "MediaURL"}
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	for _, record := range records {
		row := []string{
			record.TweetDate,
			record.TweetId,
			record.Username,
			record.DisplayName,
			record.TweetText,
			record.TweetURL,
			record.MediaType,
			record.MediaURL,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
