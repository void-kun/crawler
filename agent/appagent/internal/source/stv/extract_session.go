package sangtacviet

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractSession(browser *rod.Browser, hs *spider.HeadSpider) error {
	page := browser.MustPage()

	fmt.Printf("credential: %s %s\n", s.username, s.password)
	fmt.Println("Navigating to website...")
	err := page.Navigate("https://sangtacviet.app/truyen/fanqie/1/7486080399878147134/7487925476573315608")
	if err != nil {
		fmt.Printf("Error navigating to website: %v\n", err)
		return err
	}

	err = page.WaitLoad()
	if err != nil {
		fmt.Printf("Error waiting for page to load: %v\n", err)
	}

	fmt.Println("Waiting for modals to appear...")
	time.Sleep(1 * time.Second)

	fmt.Println("Looking for language modal...")
	languageModal, err := page.Element(".modal-content:has(.seloption[value='vi'])")
	if err != nil {
		return err
	}
	fmt.Println("Language modal found, looking for Vietnamese option...")
	vietnameseOption, err := languageModal.Element(".seloption[value='vi']")
	if err == nil {
		fmt.Println("Clicking Vietnamese language option...")
		err = vietnameseOption.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			fmt.Printf("Error clicking Vietnamese option: %v\n", err)
		} else {
			fmt.Println("Waiting for page to update after language selection...")
			time.Sleep(1 * time.Second)

			err = page.WaitLoad()
			if err != nil {
				fmt.Printf("Error waiting for page to load after language selection: %v\n", err)
			}
		}
	} else {
		fmt.Println("Vietnamese language option not found, continuing...")
	}

	fmt.Println("Clicking login link...")
	page.MustEval(`() => {
		const loginLink = document.querySelector("#tm-nav-search-top-right a");
		if (loginLink) {
			loginLink.click();
			return true;
		}
		return false;
	}`)

	fmt.Println("Looking for login form...")

	loginForm, err := page.Element("form")
	if err != nil {
		fmt.Println("No form found on the page, trying to find inputs directly...")
	} else {
		fmt.Println("Form found on the page")
	}

	fmt.Println("Looking for username field...")
	usernameInput, err := loginForm.Element("input[name*='user']")
	if err != nil {
		usernameInput, err = page.Element("input[name*='user']")
		if err != nil {
			fmt.Printf("Username field not found: %v\n", err)
			return err
		}
	}

	fmt.Println("Entering username...")
	err = usernameInput.Input(s.username)
	if err != nil {
		fmt.Printf("Error entering username: %v\n", err)
		return err
	}

	fmt.Println("Looking for password field...")
	passwordInput, err := loginForm.Element("input[type='password']")
	if err != nil {
		passwordInput, err = page.Element("input[type='password']")
		if err != nil {
			fmt.Printf("Password field not found: %v\n", err)
			return err
		}
	}

	fmt.Println("Entering password...")
	err = passwordInput.Input(s.password)
	if err != nil {
		fmt.Printf("Error entering password: %v\n", err)
		return err
	}

	fmt.Println("Clicking login button...")
	page.MustEval(`async () => {
		await loginstv()
		return true;
	}`)

	fmt.Println("Waiting for login to complete...")
	err = page.WaitLoad()
	if err != nil {
		fmt.Printf("Error waiting for login to complete: %v\n", err)
		return err
	}

	fmt.Println("Checking for captcha...")
	if spider.DetectCaptcha(page) {
		handler := spider.NewManualCaptchaHandler()
		handler.ScreenshotPath = "login_captcha.png"
		handler.WaitTime = 5 * time.Second

		err = handler.HandleCaptcha(page)
		if err != nil {
			fmt.Printf("Error handling captcha: %v\n", err)
			return err
		}
	} else {
		fmt.Println("No captcha detected, continuing...")
	}

	return nil
}
