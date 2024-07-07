package utils

import (
	"strings"
)

// isImageUrl 判断给定的URL是否可能指向一个图片
func IsImageUrl(url string) bool {
	// 定义一个图片扩展名的列表
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp"}

	// 将URL转换为小写，以便进行不区分大小写的比较
	urlLower := strings.ToLower(url)

	// 遍历所有已知的图片扩展名，检查URL是否以其中任一扩展名结尾
	for _, ext := range imageExtensions {
		if strings.HasSuffix(urlLower, ext) {
			return true
		}
	}

	// 如果没有匹配到任何已知的图片扩展名，返回false
	return false
}

// IsAudioUrl 判断给定的URL是否可能指向一个音频
func IsAudioUrl(url string) bool {
	// 定义一个音频扩展名的列表
	audioExtensions := []string{".mp3", ".wav", ".wma", ".ogg", ".m4a"}

	// 将URL转换为小写，以便进行不区分大小写的比较
	urlLower := strings.ToLower(url)

	// 遍历所有已知的音频扩展名，检查URL是否以其中任一扩展名结尾
	for _, ext := range audioExtensions {
		if strings.HasSuffix(urlLower, ext) {
			return true
		}
	}

	// 如果没有匹配到任何已知的音频扩展名，返回false
	return false
}

// IsVideoUrl 判断给定的URL是否可能指向一个视频
func IsVideoUrl(url string) bool {
	// 定义一个视频扩展名的列表
	videoExtensions := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"}

	// 将URL转换为小写，以便进行不区分大小写的比较
	urlLower := strings.ToLower(url)

	// 遍历所有已知的视频扩展名，检查URL是否以其中任一扩展名结尾
	for _, ext := range videoExtensions {
		if strings.HasSuffix(urlLower, ext) {
			return true
		}
	}

	// 如果没有匹配到任何已知的视频扩展名，返回false
	return false
}

func FileType(url string) string {
	if IsImageUrl(url) {
		return "image"
	} else if IsAudioUrl(url) {
		return "audio"
	} else if IsVideoUrl(url) {
		return "video"
	} else {
		return "unknown"
	}
}
