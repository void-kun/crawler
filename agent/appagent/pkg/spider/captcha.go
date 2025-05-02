package spider

import (
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
)

// CaptchaHandler defines the interface for handling captchas
type CaptchaHandler interface {
	HandleCaptcha(page *rod.Page) error
}

// ManualCaptchaHandler implements CaptchaHandler by waiting for manual user intervention
type ManualCaptchaHandler struct {
	ScreenshotPath string
	WaitTime       time.Duration
}

// NewManualCaptchaHandler creates a new ManualCaptchaHandler with default settings
func NewManualCaptchaHandler() *ManualCaptchaHandler {
	return &ManualCaptchaHandler{
		ScreenshotPath: "captcha.png",
		WaitTime:       3 * time.Second,
	}
}

// HandleCaptcha implements the CaptchaHandler interface for manual intervention
func (h *ManualCaptchaHandler) HandleCaptcha(page *rod.Page) error {
	// Take a screenshot to help the user see the captcha
	fmt.Println("Taking screenshot to help with captcha...")
	data, err := page.Screenshot(false, nil)
	if err != nil {
		fmt.Printf("Error taking captcha screenshot: %v\n", err)
	} else {
		err = os.WriteFile(h.ScreenshotPath, data, 0o644)
		if err != nil {
			fmt.Printf("Error saving captcha screenshot: %v\n", err)
		}
	}

	fmt.Println("\n==================================================")
	fmt.Println("CAPTCHA DETECTED - MANUAL INTERVENTION REQUIRED")
	fmt.Println("==================================================")
	fmt.Printf("1. Please check the '%s' screenshot\n", h.ScreenshotPath)
	fmt.Println("2. Go to the browser window and solve the captcha manually")
	fmt.Println("3. After solving the captcha, press Enter to continue...")
	fmt.Println("==================================================")

	// Wait for user to press Enter
	fmt.Scanln() // This will block until the user presses Enter

	fmt.Println("Continuing after manual captcha resolution...")

	// Wait for any redirects or page changes after captcha resolution
	time.Sleep(h.WaitTime)
	err = page.WaitLoad()
	if err != nil {
		fmt.Printf("Error waiting for page to load after captcha resolution: %v\n", err)
	}

	return nil
}

// DetectCaptcha checks if a captcha is present on the page
func DetectCaptcha(page *rod.Page) bool {
	captchaFound := false

	// Check for Google reCAPTCHA
	_, err := page.Element("iframe[src*='recaptcha']")
	if err == nil {
		fmt.Println("Google reCAPTCHA detected")
		return true
	}

	// Check for hCaptcha
	_, err = page.Element("iframe[src*='hcaptcha']")
	if err == nil {
		fmt.Println("hCaptcha detected")
		return true
	}

	// Check for generic captcha input fields
	_, err = page.Element("input[name*='captcha'], input[id*='captcha'], .captcha-input")
	if err == nil {
		fmt.Println("Captcha input field detected")
		return true
	}

	// Check for captcha images
	_, err = page.Element("img[src*='captcha'], img[alt*='captcha'], .captcha-image")
	if err == nil {
		fmt.Println("Captcha image detected")
		return true
	}

	return captchaFound
}
