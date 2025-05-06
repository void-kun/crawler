package stv

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/zrik/agent/appagent/pkg/spider"
)

func (s *Sangtacviet) ExtractSourceSession(browser *rod.Browser, hsType spider.TaskSpider) (any, error) {
	_, err := AsHeadSpider(hsType)
	if err != nil {
		return nil, fmt.Errorf("spider is not of type *spider.HeadSpider")
	}

	page := browser.MustPage()
	fmt.Println("=============================== Extract source session ===============================")
	err = page.Navigate(s.origin)
	if err != nil {
		log.Printf("Error navigating to website: %v\n", err)
		return nil, err
	}

	err = page.WaitLoad()
	if err != nil {
		log.Printf("Error waiting for page to load: %v\n", err)
	}

	time.Sleep(1 * time.Second)
	languageModal, err := page.Element(".modal-content:has(.seloption[value='vi'])")
	if err != nil {
		return nil, err
	}

	vietnameseOption, err := languageModal.Element(".seloption[value='vi']")
	if err == nil {
		err = vietnameseOption.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			log.Printf("Error clicking Vietnamese option: %v\n", err)
		} else {
			time.Sleep(1 * time.Second)

			err = page.WaitLoad()
			if err != nil {
				log.Printf("Error waiting for page to load after language selection: %v\n", err)
			}
		}
	} else {
		log.Println("Vietnamese language option not found, continuing...")
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
		log.Println("No form found on the page, trying to find inputs directly...")
	} else {
		log.Println("Form found on the page")
	}

	usernameInput, err := loginForm.Element("input[name*='user']")
	if err != nil {
		usernameInput, err = page.Element("input[name*='user']")
		if err != nil {
			log.Printf("Username field not found: %v\n", err)
			return nil, err
		}
	}

	err = usernameInput.Input(s.username)
	if err != nil {
		log.Printf("Error entering username: %v\n", err)
		return nil, err
	}

	passwordInput, err := loginForm.Element("input[type='password']")
	if err != nil {
		passwordInput, err = page.Element("input[type='password']")
		if err != nil {
			log.Printf("Password field not found: %v\n", err)
			return nil, err
		}
	}

	err = passwordInput.Input(s.password)
	if err != nil {
		log.Printf("Error entering password: %v\n", err)
		return nil, err
	}

	page.MustEval(`async () => {
		await loginstv()
		return true;
	}`)

	err = page.WaitLoad()
	if err != nil {
		log.Printf("Error waiting for login to complete: %v\n", err)
		return nil, err
	}

	if spider.DetectCaptcha(page) {
		handler := spider.NewManualCaptchaHandler()
		handler.WaitTime = 5 * time.Second

		err = handler.HandleCaptcha(page)
		if err != nil {
			log.Printf("Error handling captcha: %v\n", err)
			return nil, err
		}
	} else {
		log.Println("No captcha detected, continuing...")
	}

	return nil, nil
}
