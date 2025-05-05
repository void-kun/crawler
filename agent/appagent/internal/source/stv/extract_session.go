package stv

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractSession(url string, page *rod.Page, spider spider.TaskSpider) error {
	hs, err := AsHeadSpider(spider)
	if err != nil {
		return fmt.Errorf("spider is not of type *spider.HeadSpider")
	}

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
	return nil
}
