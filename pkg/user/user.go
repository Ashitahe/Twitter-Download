package user

import (
	"encoding/json"
	"log"
	"net/url"

	"twitterDownload/pkg/collector"

	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
)

type UserInfo struct {
	UserId         string
	UserName       string
	DisplayName    string
	FollowersCount int
	FollowingCount int
	TweetCount     int
	NextPageToken  string
	SaveDir        string
}

func GenerateTwitterUserInfoUrl(name string) string {
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

	queryParams := url.Values{}
	queryParams.Add("variables", createVarParams(name))
	queryParams.Add("features", createFeatParams())
	reqTwitterUserInfoUrl := &url.URL{
		Scheme:   "https",
		Host:     "twitter.com",
		Path:     "/i/api/graphql/qW5u-DAuXpMEG0zA1F7UGQ/UserByScreenName",
		RawQuery: queryParams.Encode(),
	}
	return reqTwitterUserInfoUrl.String()
}

func FetchUserInfo(userName string) (UserInfo, error) {
	c := collector.NewCollector()
	var userInfo UserInfo
	var err error

	c.OnResponse(func(r *colly.Response) {
			result := string(r.Body)
			userInfo = UserInfo{
					UserId:         gjson.Get(result, "data.user.result.rest_id").String(),
					UserName:       gjson.Get(result, "data.user.result.legacy.screen_name").String(),
					DisplayName:    gjson.Get(result, "data.user.result.legacy.name").String(),
					FollowersCount: int(gjson.Get(result, "data.user.result.legacy.followers_count").Int()),
					FollowingCount: int(gjson.Get(result, "data.user.result.legacy.friends_count").Int()),
					TweetCount:     int(gjson.Get(result, "data.user.result.legacy.media_count").Int()),
			}
	})

	c.OnError(func(r *colly.Response, e error) {
			log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", e)
			err = e
	})

	c.Visit(GenerateTwitterUserInfoUrl(userName))
	c.Wait()

	return userInfo, err
}

