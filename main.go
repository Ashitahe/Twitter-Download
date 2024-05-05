package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"myTwitterDownload/store"
	"myTwitterDownload/utils"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"

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
	saveDir        string
}

var (
	task              sync.WaitGroup
	userInfoCache     = &userInfo{}
	downloadedCount   int32
	logStore          = store.NewURLStore()
	settingsConfig, _ = utils.ReadSettings()
	csvList           = []utils.CSV{}
)

func createCollector() *colly.Collector {
	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
	return c
}

func downloadMedia(mediaUrls string) {

	extractSuffixAfterColon := func(url string) string {
		suffix := ""
		lastColonIndex := strings.LastIndex(url, ":")

		if !(lastColonIndex == -1) {
			suffix = url[lastColonIndex:]
		}

		return suffix
	}

	c := createCollector()

	c.OnResponse(func(r *colly.Response) {
		atomic.AddInt32(&downloadedCount, 1)
		reqUrl := r.Request.URL.String()
		fileName := strings.Split(reqUrl, "/")
		dir := userInfoCache.saveDir
		lastPath := fileName[len(fileName)-1]
		utils.SaveMediaFile(dir, strings.TrimSuffix(lastPath, extractSuffixAfterColon(lastPath)), r.Body)
		logStore.AddURL(strings.TrimSuffix(reqUrl, extractSuffixAfterColon(reqUrl)))
	})

	retryCount := 0
	c.OnError(func(r *colly.Response, err error) {
		retryUrl := r.Request.URL.String()
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		log.Panicln("Retry download media: ", retryUrl, "retry count: ", retryCount)
		if retryCount < 3 {
			c.Visit(retryUrl)
			retryCount++
		} else {
			log.Println("Retry download media failed: ", retryUrl)
		}
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

func extractNextPageTokenValue(json, keyword string) string {
	var result string
	gjson.Parse(json).ForEach(func(_, outer gjson.Result) bool {
		outer.ForEach(func(_, inner gjson.Result) bool {
			entryId := inner.Get("entryId").String()
			if strings.Contains(entryId, keyword) {
				result = inner.Get("content.value").String()
			}
			return true
		})
		return true
	})
	return result
}

func processUrl(originUrl string, cachedUrls *int32) {
	url := utils.TrimURLQueryAndHash(originUrl)
	if logStore.URLExists(url) {
		fmt.Println("media already downloaded: ", url)
		atomic.AddInt32(cachedUrls, 1)
		return
	}

	switch utils.FileType(url) {
	case "image":
		downloadMedia(url + ":orig")
	case "audio", "video":
		downloadMedia(url)
	default:
		fmt.Println("media type not supported: ", url)
	}
}

func downloadMediaUrls(urls []string) bool {
	var cachedUrls int32
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			processUrl(url, &cachedUrls)
		}(url)
	}

	wg.Wait()

	return int(cachedUrls) == len(urls)
}

func setHeaders(r *colly.Request) {
	cookie := settingsConfig["cookie"].(string)
	const fieldName = "ct0"
	token := utils.ExtractValueFromCookie(cookie, fieldName)
	r.Headers.Set("X-Csrf-Token", token)
	r.Headers.Set("Cookie", cookie)
	r.Headers.Set("Content-Type", "application/json")
	r.Headers.Set("Accept", "*/*")
	r.Headers.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
}

func getTwitterMedia() {
	fmt.Println("download process: ", downloadedCount)
	c := createCollector()

	c.OnRequest(setHeaders)

	handleMediaInfoResp := func(r *colly.Response) {

		processMediaInfo := func(jsonContentStr string) []utils.Legacy {
			const prefixJsonPath = "data.user.result.timeline_v2.timeline."
			const firstPageInfixJsonPath = "instructions.#.entries.#.content.items.#."
			const infixJsonPath = "instructions.#.moduleItems.#."
			const suffixJsonPath = "item.itemContent.tweet_results.result.legacy"

			mediaListJsonRes := gjson.Get(jsonContentStr, prefixJsonPath+firstPageInfixJsonPath+suffixJsonPath)

			flattenedSlice := utils.Flatten(mediaListJsonRes.Value())

			arrayLikeString, _ := utils.SliceToJSONString(flattenedSlice)

			mediaInfoList, _ := utils.ExtractMedias(arrayLikeString)

			if len(mediaInfoList) == 0 {
				mediaListJsonRes = gjson.Get(jsonContentStr, prefixJsonPath+infixJsonPath+suffixJsonPath)
				flattenedSlice = utils.Flatten(mediaListJsonRes.Value())
				arrayLikeString, _ = utils.SliceToJSONString(flattenedSlice)
			}

			legacys, _ := utils.ExtractLegacys(arrayLikeString)

			return legacys
		}

		legacyList := processMediaInfo(string(r.Body))

		var flattenedArray []string
		for _, legacyItm := range legacyList {

			log.Printf("Created at: %s\n", utils.ParseTwitterTime(legacyItm.CreatedAt))

			for _, media := range legacyItm.Extended.Media {

				csvList = append(csvList, utils.CSV{
					TweetDate:   utils.ParseTwitterTime(legacyItm.CreatedAt),
					TweetId:     legacyItm.TweetID,
					Username:    "@" + userInfoCache.userName,
					DisplayName: userInfoCache.displayName,
					TweetText:   legacyItm.TweetText,
					TweetURL:    media.ExpandedUrl,
					MediaType:   media.Type,
					MediaURL:    media.MediaURL,
				})

				if media.IsVideo {
					flattenedArray = append(flattenedArray, utils.FindMaxBitrateURL(media))
				} else {
					flattenedArray = append(flattenedArray, media.MediaURL)
				}
			}
		}

		isPageDownloaded := downloadMediaUrls(flattenedArray)

		if len(flattenedArray) == 0 || isPageDownloaded {
			fmt.Println("no more media. task completed.")
			logStore.SaveToFile("log.json")
			utils.SaveToCSV(csvList, "record.csv")
			userInfoCache = &userInfo{}
			csvList = []utils.CSV{}
			menu()
			return
		}

		const prefixJsonPath = "data.user.result.timeline_v2.timeline."
		const nextTokenJsonPath = "instructions.#.entries"

		nextTokenJsonData := gjson.Get(string(r.Body), prefixJsonPath+nextTokenJsonPath)

		userInfoCache.nextPageToken = extractNextPageTokenValue(nextTokenJsonData.String(), "bottom")

		task.Add(1)
		go func() {
			defer task.Done()
			getTwitterMedia()
		}()
	}
	c.OnResponse(handleMediaInfoResp)

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

	c.OnRequest(setHeaders)

	c.OnResponse(func(r *colly.Response) {
		result := string(r.Body)
		userInfoCache.userId = gjson.Get(result, "data.user.result.rest_id").String()
		userInfoCache.userName = gjson.Get(result, "data.user.result.legacy.screen_name").String()
		userInfoCache.displayName = gjson.Get(result, "data.user.result.legacy.name").String()
		userInfoCache.followersCount = int(gjson.Get(result, "data.user.result.legacy.followers_count").Int())
		userInfoCache.followingCount = int(gjson.Get(result, "data.user.result.legacy.friends_count").Int())
		userInfoCache.tweetCount = int(gjson.Get(result, "data.user.result.legacy.media_count").Int())
		printUserInfo()

		task.Add(1)
		go func() {
			defer task.Done()
			getTwitterMedia()
		}()
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
}

func menu() {
	fmt.Println("1. Get user info")
	fmt.Println("2. Get media from file")
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
		userInfoCache.saveDir = scanner.Text() + "/"
		getUserInfo()

	case 2:

		urls, err := utils.LoadURLsFromJSON("urls.json")
		if err != nil {
			log.Println("Error loading urls: ", err)
		} else {
			task.Add(1)
			go func() {
				defer task.Done()
				downloadMediaUrls(urls)
				menu()
			}()
		}

	case 3:
		fmt.Println("User info: ")
		printUserInfo()

	case 4:
		task.Done()
	default:
		fmt.Println("Invalid choice")
		menu()
	}
}

func main() {

	logStore.LoadFromFile("log.json")

	menu()

	task.Add(1)
	task.Wait()
}
