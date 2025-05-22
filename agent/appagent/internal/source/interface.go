package source

import (
	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/spider"
)

type WebSource interface {
	ExtractSourceSession(browser *rod.Browser, spider spider.TaskSpider) (any, error)
	ExtractSession(url string, page *rod.Page, spider spider.TaskSpider) (any, error)
	ExtractChapter(url string, page *rod.Page, spider spider.TaskSpider) (any, error)
	ExtractBookInfo(url string, page *rod.Page, spider spider.TaskSpider) (any, error)
}
