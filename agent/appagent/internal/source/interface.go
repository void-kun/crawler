package source

import (
	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/spider"
)

type WebSource interface {
	ExtractSourceSession(browser *rod.Browser, spider spider.TaskSpider) error
	ExtractSession(url string, page *rod.Page, spider spider.TaskSpider) error
	ExtractChapter(url string, page *rod.Page, spider spider.TaskSpider) error
	ExtractBookInfo(url string, page *rod.Page, spider spider.TaskSpider) error
}
