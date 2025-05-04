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
	fmt.Println("\n==================================================================================")
	fmt.Println("====== Extract session")
	page.Mouse.Click(proto.InputMouseButtonLeft, 1)
	time.Sleep(3 * time.Second)

	page.Activate()
	page.Reload()
	page.MustWaitLoad()

	fmt.Println("\n==================================================")
	fmt.Println("CHOOSE ONE CHAPTER TO EXTRACT SESSION DATA")
	fmt.Println("==================================================")

	fmt.Scanln()

	hs.ExtractSessionData(page)
	hs.SaveSessionDataToJSON()
	fmt.Println("====== Session data saved")
	return nil
}
