package utils

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

type Variant struct {
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Bitrate     int    `json:"bitrate,omitempty"`
}

type VideoInfo struct {
	AspectRatio    []int `json:"aspect_ratio"`
	DurationMillis int   `json:"duration_millis"`
	Variants       []Variant
}

// Media 结构体用于存储视频信息，包括一个额外的布尔值标识是否为视频
type Media struct {
	Type        string    `json:"type"`
	ExpandedUrl string    `json:"expanded_url"`
	MediaURL    string    `json:"media_url_https,omitempty"` // 使用omitempty标签，当字段为空时不输出到JSON
	IsVideo     bool      `json:"-"`                         // 使用"-"忽略此字段的JSON序列化和反序列化
	VideoInfo   VideoInfo `json:"video_info,omitempty"`
}

type Extended struct {
	Media []Media `json:"media"`
}

type Legacy struct {
	CreatedAt string   `json:"created_at,omitempty"` // 使用omitempty标签，当字段为空时不输出到JSON
	Extended  Extended `json:"extended_entities"`
	TweetID   string   `json:"id_str"`
	TweetText string   `json:"full_text"`
}

// ExtractMedias 从JSON数组字符串中提取视频信息
func ExtractMedias(jsonArrayStr string) ([]Media, error) {
	var mediaInfos []Media
	err := json.Unmarshal([]byte(jsonArrayStr), &mediaInfos)
	if err != nil {
		return nil, err
	}

	// 遍历每个视频信息对象，根据type字段的值设置IsVideo
	for i, mediaInfo := range mediaInfos {
		mediaInfos[i].IsVideo = mediaInfo.Type == "video"
	}

	return mediaInfos, nil
}

// ExtractLegacyList 从JSON数组字符串中提取Legacy对象List
func ExtractLegacyList(jsonArrayStr string) ([]Legacy, error) {
	var legacyList []Legacy
	err := json.Unmarshal([]byte(jsonArrayStr), &legacyList)
	if err != nil {
		return nil, err
	}

	return legacyList, nil
}

// FindMaxBitrateURL 遍历variants数组，找到bitrate最大的元素并返回其URL
func FindMaxBitrateURL(mediaInfo Media) string {
	maxBitrate := 0
	maxBitrateURL := ""

	for _, variant := range mediaInfo.VideoInfo.Variants {
		if variant.Bitrate > maxBitrate {
			maxBitrate = variant.Bitrate
			maxBitrateURL = variant.URL
		}
	}

	return maxBitrateURL
}

// Flatten 接受一个任意深度嵌套的切片，并返回一个扁平化的切片
func Flatten(input interface{}) []interface{} {
	// 初始化结果切片，不预分配容量以避免过度分配
	var result []interface{}

	// 定义一个递归函数，用于内部处理嵌套逻辑
	var flattenRec func(item interface{})
	flattenRec = func(item interface{}) {
		// 尝试将item断言为切片
		if slice, ok := item.([]interface{}); ok {
			// 如果是切片，递归扁平化每个元素
			for _, elem := range slice {
				flattenRec(elem)
			}
		} else {
			// 如果不是切片，直接添加到结果切片
			result = append(result, item)
		}
	}

	// 开始处理输入
	flattenRec(input)

	// 返回扁平化后的结果
	return result
}

// SliceToJSONString 将一个 []interface{} 类型的切片转换为一个类似数组的JSON字符串
func SliceToJSONString(slice []interface{}) (string, error) {
	// 预估结果字符串的大小，这里假设每个元素转换后平均长度为1024字节
	estimatedSize := len(slice) * 1024
	var sb strings.Builder
	sb.Grow(estimatedSize) // 预分配容量以减少扩容次数

	// 开始数组字符串
	sb.WriteString("[")

	// 遍历切片，将每个元素转换为JSON字符串并添加到结果中
	for i, v := range slice {
		// 使用 json.Marshal 将元素转换为 JSON 字符串
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", err // 返回错误
		}

		// 将 JSON 字符串添加到结果字符串中
		sb.Write(jsonBytes)

		// 如果不是最后一个元素，添加逗号分隔符
		if i < len(slice)-1 {
			sb.WriteString(", ")
		}
	}

	// 结束数组字符串
	sb.WriteString("]")

	// 返回最终的类似数组的字符串
	return sb.String(), nil
}

// TrimURLQueryAndHash 从URL中除去查询参数和哈希
func TrimURLQueryAndHash(url string) string {
	// 查找查询参数的开始位置
	queryStart := strings.Index(url, "?")
	// 查找哈希的开始位置
	hashStart := strings.Index(url, "#")

	// 如果没有找到查询参数和哈希，返回原URL
	if queryStart == -1 && hashStart == -1 {
		return url
	}

	// 如果找到了查询参数，但没有找到哈希，或者查询参数在哈希之前出现
	if queryStart != -1 && (hashStart == -1 || queryStart < hashStart) {
		return url[:queryStart]
	}

	// 如果找到了哈希，但没有找到查询参数，或者哈希在查询参数之前出现
	return url[:hashStart]
}

// ParseTwitterTime 将Twitter时间格式转换为ISO日期格式
func ParseTwitterTime(inputTime string) string {
	twitterTimeLayout := "Mon Jan 2 15:04:05 -0700 2006"
	isoDateLayout := "2006-01-02"
	parsedTime, err := time.Parse(twitterTimeLayout, inputTime)
	if err != nil {
		log.Printf("Error parsing time: %v", err)
		return ""
	}
	return parsedTime.Format(isoDateLayout)
}

// ExtractValueFromCookie 从cookie字符串中提取指定字段的值
func ExtractValueFromCookie(cookie string, fieldName string) string {
	parts := strings.Split(cookie, ";")
	for _, part := range parts {
		pair := strings.Split(strings.TrimSpace(part), "=")
		if pair[0] == fieldName {
			return pair[1]
		}
	}
	return ""
}
