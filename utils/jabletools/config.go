package jabletools

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func getConfigPath(defaultPATH string) (pwd string, chromePath string) {
	chromePath = `/bin/chromedriver`
	pwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return ``, chromePath
	}
	if runtime.GOOS == "darwin" {
		chromePath = pwd + "/bin/chromedriver_mac"
		if runtime.GOARCH == "arm64" {
			chromePath = pwd + "/bin/chromedriver_mac_arm"
		}
	} else if runtime.GOOS == "windows" {
		chromePath = pwd + "/bin/chromedriver.exe"
	} else if runtime.GOOS == "linux" {
		chromePath = pwd + "/bin/chromedriver_linux"
	}
	if len(defaultPATH) > 0 {
		return defaultPATH, chromePath
	}
	return
}
func getSeleniumCapabilitiess() selenium.Capabilities {
	return selenium.Capabilities{
		"chrome": chrome.Capabilities{
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
				"--disable-extensions",
				"--headless=new",
				"--disable-gpu",
				"--hide-scrollbars",
				"blink-settings=imagesEnabled=false",
				"--disable-software-rasterizer",
				fmt.Sprintf(`--adb-port=%d `, CHROME_PORT),
				"--silent",
				"user-agent=Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.5672.92 Safari/537.36",
			},
		},
	}
}
