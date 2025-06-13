package stv

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/zrik/agent/appagent/pkg/logger"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractChapter(chapterUrl string, page *rod.Page, hSpider spider.TaskSpider) (any, error) {
	_, err := AsHeadSpider(hSpider)
	if err != nil {
		return nil, fmt.Errorf("spider is not of type *spider.HeadSpider")
	}
	defer page.MustClose()

	hSpider.ApplySessionData(page)

	page.MustWaitLoad()

	// Wait for chapter data loaded
	loopTime := 0
	var chapterContent *rod.Element
	for {
		spider.CircleMoveMouse(page)
		logger.Info().Msg("Waiting for chapter data to load...")

		if has, ele, _ := page.Has("div#content-container div.contentbox i"); has {
			logger.Info().Str("chapter", ele.MustText()).Msg("Chapter data loaded")

			chapterContent = page.MustElement("div#content-container div.contentbox")
			break
		}

		// Click the element to load the chapter

		loopTime++
		if loopTime > 3 {
			page.MustReload().MustWaitLoad()
			time.Sleep(1*time.Second + time.Duration(loopTime)*time.Second)
		}

		time.Sleep(2 * time.Second)
	}

	if chapterContent == nil {
		return nil, fmt.Errorf("failed to locate chapter content")
	}

	result, _ := ExtractTextFromHTML(chapterContent.MustHTML())
	chapterBytes, _ := ConvertToRawMessage(result)
	return chapterBytes, nil
}
