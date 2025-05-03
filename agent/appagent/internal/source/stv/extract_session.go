package sangtacviet

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractSession(url string, page *rod.Page, hs *spider.HeadSpider) error {
	if !strings.HasPrefix(url, hs.SessionPrefix) {
		return nil
	}

	fmt.Println("\n====== Extract session ======")
	page.Mouse.Click(proto.InputMouseButtonLeft, 1)
	time.Sleep(3 * time.Second)

	page.Activate()
	page.Reload()
	page.MustWaitLoad()

	for {
		contentElements := page.MustElements("div#content-container > div i")
		if len(contentElements) > 0 {
			break
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Println("\n==================================================")
	fmt.Println("CHOOSE ONE CHAPTER TO EXTRACT SESSION DATA")
	fmt.Println("==================================================")

	fmt.Scanln()

	hs.ExtractSessionData(page)
	hs.SaveSessionDataToJSON()
	return nil
}
