package spider

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// TaskSpider defines the interface for a spider that can process tasks
type TaskSpider interface {
	// Browser management
	InitBrowser() error
	CreatePage() (*rod.Page, error)
	CloseBrowser()
	SetHeadless(isHeadless bool)

	// Session management
	GetCookies() []*proto.NetworkCookie
	SetCookies(cookies []*proto.NetworkCookie)
	LoadCookiesFromJSON(filePath string) error
	SaveCookiesToJSON(filePath string) error

	// Session data
	ExtractSessionData(page *rod.Page) error
	GetSessionData() *SessionData
	SaveSessionDataToJSON() error
	LoadSessionDataFromJSON() error
	ApplySessionData(page *rod.Page) error

	// Preparation steps
	AddPrepStep(step func(*rod.Browser, *HeadSpider) error)
	ExecutePrepSteps() error

	// Task processing
	ProcessURL(url string) error
	ProcessBookURL(bookURL string, bookID string, bookHost string) error
	ProcessChapterURL(chapterURL string, bookID string, chapterID string, bookHost string, bookSty string) error
	ProcessSessionURL(url string) error
	ProcessPageWithCallback(url string, callback func(url string, page *rod.Page, spider TaskSpider) error) error
}
