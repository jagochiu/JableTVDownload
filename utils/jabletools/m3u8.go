package jabletools

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/tebeka/selenium"
)

const (
	DOWNLOAD_PATH   = `downloads`
	INPROGRESS_PATH = `inprogress`
	COMPLETED_PATH  = `completed`
	CHROME_PORT     = 59515
)

func M3U8(driverPath, url string) (string, string) {
	// 配置Selenium參數
	wd, err := selenium.NewChromeDriverService(driverPath, CHROME_PORT)
	if err != nil {
		log.Fatalf("Failed to start the ChromeDriver service: %v \n", err)
	}
	caps := getSeleniumCapabilitiess()
	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", CHROME_PORT))
	if err != nil {
		log.Fatalf("Failed to create new remote chrome driver: %v \n", err)
	}
	defer driver.Quit()
	err = driver.Get(url)
	if err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}
	// 得到 m3u8 網址
	pageSource, err := driver.PageSource()
	if err != nil {
		log.Fatalf("Failed to get page source: %v", err)
	}
	wd.Stop()
	result := regexp.MustCompile("https://.+m3u8").FindStringSubmatch(pageSource)

	return pageSource, result[0]
}
func getIntoFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("[FILE] unable to download file: %v \n ", err)
	}
	defer resp.Body.Close()
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("[FILE] unable to create file: %v \n ", err)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("[FILE] unable to write file: %v \n ", err)
	}
	return nil
}
func exportDir(pwd, id string) (string, string) {
	downloadPATH := filepath.Join(pwd, DOWNLOAD_PATH)
	inprogressPATH := filepath.Join(downloadPATH, INPROGRESS_PATH)
	completedPATH := filepath.Join(downloadPATH, COMPLETED_PATH)
	exportPATH := filepath.Join(inprogressPATH, id)
	segmentsPATH := filepath.Join(exportPATH, `segments`)
	_, err := os.Stat(filepath.Join(completedPATH, id))
	if errors.Is(err, os.ErrExist) {
		log.Fatalln(`番號資料夾已存在, 跳過... `)
	}
	_, err = os.Stat(exportPATH)
	if errors.Is(err, os.ErrNotExist) {
		createDir(downloadPATH, true)
		createDir(inprogressPATH, true)
		createDir(completedPATH, true)
		createDir(exportPATH, false)
		createDir(segmentsPATH, false)
	} else if err != nil {
		log.Fatalln("    Schrodinger: file may or may not exist. See err for details.\n    Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence")
	} else {
		log.Fatalln(`番號資料夾已存在, 跳過... `)
	}
	return exportPATH, segmentsPATH
}
func createDir(path string, skip bool) {
	err := os.Mkdir(path, 0775)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			if !skip {
				log.Fatalf("[ERROR] %v \n ", err)
			}
			return
		}
		log.Fatalf("[ERROR] %v \n ", err)
	}
}
