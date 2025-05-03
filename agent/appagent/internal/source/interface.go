package source

import (
	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/spider"
)

type WebSource interface {
	ExtractSession(browser *rod.Browser, s *spider.HeadSpider) error
	ExtractChapter(url string, page *rod.Page, hs *spider.HeadSpider) error
	ExtractBookInfo(url string, page *rod.Page, hs *spider.HeadSpider) error
}
