package download

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"twitterDownload/pkg/collector"
	"twitterDownload/pkg/config"
	"twitterDownload/pkg/user"
	"twitterDownload/pkg/utils"

	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
)

func downloadMedia(mediaUrls string, userInfo *user.UserInfo) {

	extractSuffixAfterColon := func(url string) string {
		suffix := ""
		lastColonIndex := strings.LastIndex(url, ":")

		if !(lastColonIndex == -1) {
			suffix = url[lastColonIndex:]
		}

		return suffix
	}

	c := collector.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		reqUrl := r.Request.URL.String()
		fileName := strings.Split(reqUrl, "/")
		dir := userInfo.SaveDir
		lastPath := fileName[len(fileName)-1]
		utils.SaveMediaFile(dir, strings.TrimSuffix(lastPath, extractSuffixAfterColon(lastPath)), r.Body)
		config.LogRecord.AddURL(strings.TrimSuffix(reqUrl, extractSuffixAfterColon(reqUrl)))
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

func processUrl(originUrl string, cachedUrls *int32, userInfo *user.UserInfo) {
	url := utils.TrimURLQueryAndHash(originUrl)
	if config.LogRecord.URLExists(url) {
		fmt.Println("media already downloaded: ", url)
		atomic.AddInt32(cachedUrls, 1)
		return
	}

	switch utils.FileType(url) {
	case "image":
		downloadMedia(url + ":orig", userInfo)
	case "audio", "video":
		downloadMedia(url, userInfo)
	default:
		fmt.Println("media type not supported: ", url)
	}
}

func downloadMediaUrls(urls []string, userInfo *user.UserInfo) bool {
	var cachedUrls int32
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			processUrl(u, &cachedUrls, userInfo)
		}(url)
	}

	wg.Wait()

	return int(cachedUrls) == len(urls)
}

func extractMediaInfo(jsonContentStr string) []utils.Legacy {
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
	legacyList, _ := utils.ExtractLegacyList(arrayLikeString)
	return legacyList
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

func generateTwitterMediaUrl(userInfoCache *user.UserInfo) string {

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

	queryParams := url.Values{}

	queryParams.Add("variables", createVariables(userInfoCache.UserId, userInfoCache.NextPageToken))
	queryParams.Add("features", createFeatures())
	twitterMediaUrl := &url.URL{
		Scheme:   "https",
		Host:     "twitter.com",
		Path:     "/i/api/graphql/aQQLnkexAl5z9ec_UgbEIA/UserMedia",
		RawQuery: queryParams.Encode(),
	}
	return twitterMediaUrl.String()
}

func DownloadTwitterMedia(userInfoCache *user.UserInfo, csvList *[]utils.CSV) {
	c := collector.NewCollector()

	handleMediaInfoResp := func(r *colly.Response) {
			legacyList := extractMediaInfo(string(r.Body))
			var flattenedArray []string
			for _, legacyItm := range legacyList {
					for _, media := range legacyItm.Extended.Media {
							*csvList = append(*csvList, utils.CSV{
									TweetDate:   utils.ParseTwitterTime(legacyItm.CreatedAt),
									TweetId:     legacyItm.TweetID,
									Username:    "@" + userInfoCache.UserName,
									DisplayName: userInfoCache.DisplayName,
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

			isPageDownloaded := downloadMediaUrls(flattenedArray, userInfoCache)

			if len(flattenedArray) == 0 || isPageDownloaded {
				fmt.Println("no more media. task completed.")
				config.LogRecord.SaveToFile()
				utils.SaveToCSV(*csvList, "record.csv")
				return
			}

			const prefixJsonPath = "data.user.result.timeline_v2.timeline."
			const nextTokenJsonPath = "instructions.#.entries"
	
			nextTokenJsonData := gjson.Get(string(r.Body), prefixJsonPath+nextTokenJsonPath)
	
			userInfoCache.NextPageToken = extractNextPageTokenValue(nextTokenJsonData.String(), "bottom")
	
			go func() {
				DownloadTwitterMedia(userInfoCache, csvList)
			}()
	}
	c.OnResponse(handleMediaInfoResp)

	c.Visit(generateTwitterMediaUrl(userInfoCache))
	c.Wait()
}

