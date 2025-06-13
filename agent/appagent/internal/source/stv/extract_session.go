package stv

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractSession(url string, page *rod.Page, spider spider.TaskSpider) (any, error) {
	hs, err := AsHeadSpider(spider)
	if err != nil {
		return nil, fmt.Errorf("spider is not of type *spider.HeadSpider")
	}

	defer page.MustClose()

	spider.ApplySessionData(page)
	page.MustReload().MustWaitLoad()

	_ = page.MustEval(`
		() => {
			// Set cookies before sending the request
			const hstamp = Math.floor(Date.now() / 1000);
			document.cookie = 'hstamp=' + hstamp + '; path=/';
			document.cookie = 'lang=vi; path=/';
		}
	`)

	fmt.Println("==================================================================================")
	fmt.Println("=============================== Extracting session ===============================")
	page.Mouse.Click(proto.InputMouseButtonLeft, 1)
	time.Sleep(3 * time.Second)

	page.Activate()
	page.Reload()
	page.MustWaitLoad()

	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("CHOOSE ONE CHAPTER TO EXTRACT SESSION DATA")
	fmt.Println("==================================================")

	fmt.Scanln()

	hs.ExtractSessionData(page)
	hs.SaveSessionDataToJSON()

	result, _ := ConvertToRawMessage(nil)
	return result, nil
}
