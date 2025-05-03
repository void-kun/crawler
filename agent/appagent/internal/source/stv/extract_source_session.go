package sangtacviet

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractSourceSession(browser *rod.Browser, hs *spider.HeadSpider) error {
	page := browser.MustPage()

	fmt.Println("\n====== Extract source session ======")
	err := page.Navigate("https://sangtacviet.app/")
	if err != nil {
		fmt.Printf("Error navigating to website: %v\n", err)
		return err
	}

	err = page.WaitLoad()
	if err != nil {
		fmt.Printf("Error waiting for page to load: %v\n", err)
	}

	time.Sleep(1 * time.Second)
	languageModal, err := page.Element(".modal-content:has(.seloption[value='vi'])")
	if err != nil {
		return err
	}

	vietnameseOption, err := languageModal.Element(".seloption[value='vi']")
	if err == nil {
		err = vietnameseOption.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			fmt.Printf("Error clicking Vietnamese option: %v\n", err)
		} else {
			time.Sleep(1 * time.Second)

			err = page.WaitLoad()
			if err != nil {
				fmt.Printf("Error waiting for page to load after language selection: %v\n", err)
			}
		}
	} else {
		fmt.Println("Vietnamese language option not found, continuing...")
	}

	page.MustEval(`() => {
		const loginLink = document.querySelector("#tm-nav-search-top-right a");
		if (loginLink) {
			loginLink.click();
			return true;
		}
		return false;
	}`)

	loginForm, err := page.Element("form")
	if err != nil {
		fmt.Println("No form found on the page, trying to find inputs directly...")
	} else {
		fmt.Println("Form found on the page")
	}

	usernameInput, err := loginForm.Element("input[name*='user']")
	if err != nil {
		usernameInput, err = page.Element("input[name*='user']")
		if err != nil {
			fmt.Printf("Username field not found: %v\n", err)
			return err
		}
	}

	err = usernameInput.Input(s.username)
	if err != nil {
		fmt.Printf("Error entering username: %v\n", err)
		return err
	}

	passwordInput, err := loginForm.Element("input[type='password']")
	if err != nil {
		passwordInput, err = page.Element("input[type='password']")
		if err != nil {
			fmt.Printf("Password field not found: %v\n", err)
			return err
		}
	}

	err = passwordInput.Input(s.password)
	if err != nil {
		fmt.Printf("Error entering password: %v\n", err)
		return err
	}

	page.MustEval(`async () => {
		await loginstv()
		return true;
	}`)

	err = page.WaitLoad()
	if err != nil {
		fmt.Printf("Error waiting for login to complete: %v\n", err)
		return err
	}

	if spider.DetectCaptcha(page) {
		handler := spider.NewManualCaptchaHandler()
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
