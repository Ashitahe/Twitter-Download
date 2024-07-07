package collector

import (
	"runtime"
	"twitterDownload/pkg/config"
	"twitterDownload/pkg/utils"

	"github.com/gocolly/colly"
)

func setHeaders(r *colly.Request) {
	cookie := config.SettingConfig.Cookie
	const fieldName = "ct0"
	token := utils.ExtractValueFromCookie(cookie, fieldName)
	r.Headers.Set("X-Csrf-Token", token)
	r.Headers.Set("Cookie", cookie)
	r.Headers.Set("Content-Type", "application/json")
	r.Headers.Set("Accept", "*/*")
	r.Headers.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
}

func NewCollector() *colly.Collector {
	c := colly.NewCollector(colly.Async(true))

	// 设置并发限制
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: runtime.NumCPU(),
	})

	c.OnRequest(setHeaders)

	return c
}
