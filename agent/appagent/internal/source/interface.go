package source

import (
	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/spider"
)

type WebSource interface {
	ExtractSourceSession(browser *rod.Browser, hs *spider.HeadSpider) error
	ExtractSession(url string, page *rod.Page, hs *spider.HeadSpider) error
	ExtractChapter(url string, page *rod.Page, hs *spider.HeadSpider) error
	ExtractBookInfo(url string, page *rod.Page, hs *spider.HeadSpider) error
}
