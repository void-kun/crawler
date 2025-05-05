package spider

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
)

// CaptchaHandler defines the interface for handling captchas
type CaptchaHandler interface {
	HandleCaptcha(page *rod.Page) error
}

// ManualCaptchaHandler implements CaptchaHandler by waiting for manual user intervention
type ManualCaptchaHandler struct {
	WaitTime time.Duration
}

// NewManualCaptchaHandler creates a new ManualCaptchaHandler with default settings
func NewManualCaptchaHandler() *ManualCaptchaHandler {
	return &ManualCaptchaHandler{
		WaitTime: 3 * time.Second,
	}
}

// HandleCaptcha implements the CaptchaHandler interface for manual intervention
func (h *ManualCaptchaHandler) HandleCaptcha(page *rod.Page) error {
	fmt.Println("\n==================================================")
	fmt.Println("CAPTCHA DETECTED - MANUAL INTERVENTION REQUIRED")
	fmt.Println("==================================================")

	// Wait for user to press Enter
	fmt.Scanln() // This will block until the user presses Enter

	log.Println("Continuing after manual captcha resolution...")

	// Wait for any redirects or page changes after captcha resolution
	time.Sleep(h.WaitTime)
	return nil
}

// DetectCaptcha checks if a captcha is present on the page
func DetectCaptcha(page *rod.Page) bool {
	captchaFound := false

	// Check for Google reCAPTCHA
	_, err := page.Element("iframe[src*='recaptcha']")
	if err == nil {
		log.Println("Google reCAPTCHA detected")
		return true
	}

	// Check for hCaptcha
	_, err = page.Element("iframe[src*='hcaptcha']")
	if err == nil {
		log.Println("hCaptcha detected")
		return true
	}

	// Check for generic captcha input fields
	_, err = page.Element("input[name*='captcha'], input[id*='captcha'], .captcha-input")
	if err == nil {
		log.Println("Captcha input field detected")
		return true
	}

	// Check for captcha images
	_, err = page.Element("img[src*='captcha'], img[alt*='captcha'], .captcha-image")
	if err == nil {
		log.Println("Captcha image detected")
		return true
	}

	return captchaFound
}
