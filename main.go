package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"myTwitterDownload/utils"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
)

type userInfo struct {
	userId         string
	userName       string
	displayName    string
	followersCount int
	followingCount int
	tweetCount     int
	nextPageToken  string
}

var (
	userInfoCache = &userInfo{}
)

func extractValueFromCookie(cookie string, fieldName string) string {
	parts := strings.Split(cookie, ";")
	for _, part := range parts {
		pair := strings.Split(strings.TrimSpace(part), "=")
		if pair[0] == fieldName {
			return pair[1]
		}
	}
	return ""
}

func createCollector() *colly.Collector {
	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
	return c
}

func processString(input string) string {
	return strings.TrimSuffix(input, ":orig")
}

func downloadMedia(mediaUrls string) {
	c := createCollector()

	c.OnResponse(func(r *colly.Response) {
		fileName := strings.Split(r.Request.URL.Path, "/")
		dir := "media/"
		resName := fileName[len(fileName)-1]
		utils.SaveFile(dir, processString(resName), r.Body)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(mediaUrls)
	c.Wait()
}

func getTwitterMediaUrl() string {

	createVariables := func(userId string, cursor string) string {
		variables := map[string]interface{}{
			"userId":                 userId,
			"count":                  20,
			"cursor":                 cursor,
			"includePromotedContent": false,
			"withClientEventToken":   false,
			"withBirdwatchNotes":     false,
			"withVoice":              true,
			"withV2Timeline":         true,
		}

		variablesData, err := json.Marshal(variables)
		if err != nil {
			log.Println(err)
			return ""
		}

		return string(variablesData)
	}

	createFeatures := func() string {
		features := map[string]interface{}{
			"rweb_tipjar_consumption_enabled":                                         true,
			"responsive_web_graphql_exclude_directive_enabled":                        true,
			"verified_phone_label_enabled":                                            false,
			"creator_subscriptions_tweet_preview_api_enabled":                         true,
			"responsive_web_graphql_timeline_navigation_enabled":                      true,
			"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
			"communities_web_enable_tweet_community_results_fetch":                    true,
			"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
			"articles_preview_enabled":                                                true,
			"tweetypie_unmention_optimization_enabled":                                true,
			"responsive_web_edit_tweet_api_enabled":                                   true,
			"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
			"view_counts_everywhere_api_enabled":                                      true,
			"longform_notetweets_consumption_enabled":                                 true,
			"responsive_web_twitter_article_tweet_consumption_enabled":                true,
			"tweet_awards_web_tipping_enabled":                                        false,
			"creator_subscriptions_quote_tweet_preview_enabled":                       false,
			"freedom_of_speech_not_reach_fetch_enabled":                               true,
			"standardized_nudges_misinfo":                                             true,
			"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
			"tweet_with_visibility_results_prefer_gql_media_interstitial_enabled":     true,
			"rweb_video_timestamps_enabled":                                           true,
			"longform_notetweets_rich_text_read_enabled":                              true,
			"longform_notetweets_inline_media_enabled":                                true,
			"responsive_web_enhance_cards_enabled":                                    false,
		}

		featuresData, err := json.Marshal(features)
		if err != nil {
			log.Println(err)
			return ""
		}

		return string(featuresData)
	}

	qureyParams := url.Values{}

	fmt.Println("nextPageToken 2: ", userInfoCache.nextPageToken)

	qureyParams.Add("variables", createVariables(userInfoCache.userId, userInfoCache.nextPageToken))
	qureyParams.Add("features", createFeatures())
	getTwitterMediaUrl := &url.URL{
		Scheme:   "https",
		Host:     "twitter.com",
		Path:     "/i/api/graphql/aQQLnkexAl5z9ec_UgbEIA/UserMedia",
		RawQuery: qureyParams.Encode(),
	}
	return getTwitterMediaUrl.String()
}

func flattenArray(input gjson.Result) []string {
	var result []string

	input.ForEach(func(key, value gjson.Result) bool {
		if value.IsArray() {
			result = append(result, flattenArray(value)...)
		} else {
			result = append(result, value.String())
		}
		return true
	})

	return result
}

func extractNextPageTokenValue(json, keyword string) string {
	var result string
	gjson.Parse(json).ForEach(func(_, outer gjson.Result) bool {
		outer.ForEach(func(_, inner gjson.Result) bool {
			entryId := inner.Get("entryId").String()
			if strings.Contains(entryId, keyword) {
				result = inner.Get("content.value").String()
			}
			return true // keep iterating
		})
		return true // keep iterating
	})
	return result
}

func getTwitterMedia() {
	c := createCollector()

	setHeaders := func(r *colly.Request) {
		settings, _ := utils.ReadSettings()
		cookie := settings["cookie"].(string)
		const fieldName = "ct0"
		token := extractValueFromCookie(cookie, fieldName)
		r.Headers.Set("X-Csrf-Token", token)
		r.Headers.Set("Cookie", cookie)
		r.Headers.Set("Content-Type", "application/json")
		r.Headers.Set("Accept", "*/*")
		r.Headers.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	}

	c.OnRequest(setHeaders)

	extractMediaUrl := func(r *colly.Response) {
		const tmp = "item.itemContent.tweet_results.result.legacy.extended_entities.media.#.media_url_https"
		mediaUrls := gjson.Get(string(r.Body), "data.user.result.timeline_v2.timeline")

		extractUrls := func(path string) gjson.Result {
			return gjson.Get(mediaUrls.String(), path)
		}

		imgUrls := extractUrls("instructions.#.entries.#.content.items.#." + tmp)

		tmpStr, _ := json.Marshal(flattenArray(imgUrls))

		imgUrls = gjson.ParseBytes(tmpStr)

		if len(imgUrls.Array()) == 0 || !imgUrls.Exists() {
			fmt.Println("imgUrls is empty")
			imgUrls = extractUrls("instructions.#.moduleItems.#." + tmp)
		}

		getInfo := extractUrls("instructions.#.entries")

		userInfoCache.nextPageToken = extractNextPageTokenValue(getInfo.String(), "bottom")

		fmt.Println("nextPageToken: ", userInfoCache.nextPageToken)

		flattenedArray := flattenArray(imgUrls)
		// fmt.Println("imgUrls: ", flattenedArray)

		var wg sync.WaitGroup
		for _, url := range flattenedArray {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				downloadMedia(url + ":orig") // 添加原圖後綴，下載原圖
			}(url)
		}
		wg.Wait()

		if len(imgUrls.Array()) == 0 {
			fmt.Println("no more media")
			os.Exit(0)
		}
	}
	c.OnResponse(extractMediaUrl)

	handleError := func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	}
	c.OnError(handleError)

	c.Visit(getTwitterMediaUrl())
	c.Wait()
}

func getTwitterUserInfoUrl() string {
	createVarParams := func(userName string) string {
		variables := map[string]interface{}{
			"screen_name":              userName,
			"withSafetyModeUserFields": true,
		}
		variablesData, err := json.Marshal(variables)
		if err != nil {
			log.Println(err)
			return ""
		}
		return string(variablesData)
	}

	createFeatParams := func() string {
		features := map[string]interface{}{"hidden_profile_likes_enabled": true, "hidden_profile_subscriptions_enabled": true, "rweb_tipjar_consumption_enabled": true, "responsive_web_graphql_exclude_directive_enabled": true, "verified_phone_label_enabled": false, "subscriptions_verification_info_is_identity_verified_enabled": true, "subscriptions_verification_info_verified_since_enabled": true, "highlights_tweets_tab_ui_enabled": true, "responsive_web_twitter_article_notes_tab_enabled": true, "creator_subscriptions_tweet_preview_api_enabled": true, "responsive_web_graphql_skip_user_profile_image_extensions_enabled": false, "responsive_web_graphql_timeline_navigation_enabled": true}
		featuresData, err := json.Marshal(features)
		if err != nil {
			log.Println(err)
			return ""
		}
		return string(featuresData)
	}

	qureyParams := url.Values{}
	qureyParams.Add("variables", createVarParams(userInfoCache.userName))
	qureyParams.Add("features", createFeatParams())
	reqTwitterUserInfoUrl := &url.URL{
		Scheme:   "https",
		Host:     "twitter.com",
		Path:     "/i/api/graphql/qW5u-DAuXpMEG0zA1F7UGQ/UserByScreenName",
		RawQuery: qureyParams.Encode(),
	}
	return reqTwitterUserInfoUrl.String()
}

func getUserInfo() {
	c := createCollector()

	c.OnRequest(func(r *colly.Request) {
		settings, _ := utils.ReadSettings()
		cookie := settings["cookie"].(string)
		const fieldName = "ct0"
		token := extractValueFromCookie(cookie, fieldName)
		r.Headers.Set("X-Csrf-Token", token)
		r.Headers.Set("Cookie", cookie)
		r.Headers.Set("Content-Type", "application/json")
		r.Headers.Set("Accept", "*/*")
		r.Headers.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	})

	c.OnResponse(func(r *colly.Response) {
		result := string(r.Body)
		userInfoCache.userId = gjson.Get(result, "data.user.result.rest_id").String()
		userInfoCache.userName = gjson.Get(result, "data.user.result.legacy.screen_name").String()
		userInfoCache.displayName = gjson.Get(result, "data.user.result.legacy.name").String()
		userInfoCache.followersCount = int(gjson.Get(result, "data.user.result.legacy.followers_count").Int())
		userInfoCache.followingCount = int(gjson.Get(result, "data.user.result.legacy.friends_count").Int())
		userInfoCache.tweetCount = int(gjson.Get(result, "data.user.result.legacy.media_count").Int())
		getTwitterMedia()
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(getTwitterUserInfoUrl())
	c.Wait()
}

func printUserInfo() {
	fmt.Println("userId: ", userInfoCache.userId)
	fmt.Println("userName: ", userInfoCache.userName)
	fmt.Println("displayName: ", userInfoCache.displayName)
	fmt.Println("followersCount: ", userInfoCache.followersCount)
	fmt.Println("followingCount: ", userInfoCache.followingCount)
	fmt.Println("tweetCount: ", userInfoCache.tweetCount)
	fmt.Println("nextPageToken: ", userInfoCache.nextPageToken)
}

func menu() {
	fmt.Println("1. Get user info")
	fmt.Println("2. Get media")
	fmt.Println("3. Print user info")
	fmt.Println("4. Exit")

	var choice int = 0
	fmt.Scanln(&choice)

	scanner := bufio.NewScanner(os.Stdin)

	switch choice {
	case 1:
		fmt.Println("Enter user name: ")
		scanner.Scan()
		userInfoCache.userName = scanner.Text()
		getUserInfo()

		ticker := time.NewTicker(5 * time.Second)

		go func() {
			for range ticker.C {
				getUserInfo()
			}
		}()

	case 2:
		getTwitterMedia()

	case 3:
		fmt.Println("User info: ")
		printUserInfo()
	default:
		fmt.Println("Invalid choice")
	}
}

func main() {
	menu()

	select {}
}
