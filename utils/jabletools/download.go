package jabletools

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Download(url, defaultPATH string) int {
	status := 0
	pwd, driverPath := getConfigPath(defaultPATH)

	// 建立番號資料夾
	urlSplit := strings.Split(url, `/`)
	id := urlSplit[len(urlSplit)-2]
	log.Printf("[Jable.TV] %s : \n ", id)
	exportPATH, segmentsPATH := exportDir(pwd, id)
	exportFILE := filepath.Join(exportPATH, id)

	// 取得 m3u8 網址 及 文件
	pageSource, m3u8url := M3U8(driverPath, url)
	log.Println("[M3U8_URL] " + m3u8url)
	err := os.WriteFile(exportFILE+`.html`, []byte(pageSource), 0755)
	if err != nil {
		log.Printf(`[ERROR] %v`, err)
	}

	// 得到 m3u8 網址 list
	m3u8urlList := strings.Split(m3u8url, `/`)
	if len(m3u8urlList) > 0 {
		m3u8urlList = m3u8urlList[:len(m3u8urlList)-1]
	}
	downloadurl := strings.Join(m3u8urlList, `/`)
	log.Println("[DOWNLOAD_URL] " + downloadurl)

	// 儲存 m3u8 file 至資料夾
	err = getIntoFile(m3u8url, exportFILE+`.m3u8`)
	if err != nil {
		log.Printf("[M3U8] %v \n", err)
	}
	m3u8file, err := os.Open(exportFILE + `.m3u8`)
	if err != nil {
		log.Fatal(err)
	}

	// 得到 m3u8 file裡的 URI和 IV
	m3u8uri := ``
	m3u8iv := ``
	m3u8Scanner := bufio.NewScanner(m3u8file)
	m3u8Scanner.Split(bufio.ScanLines)
	for m3u8Scanner.Scan() {
		if strings.HasPrefix(m3u8Scanner.Text(), "#EXT-X-KEY:") {
			paramsLine := strings.ReplaceAll(m3u8Scanner.Text(), "#EXT-X-KEY:", "")
			params := strings.Split(paramsLine, ",")
			for _, v := range params {
				if strings.HasPrefix(strings.TrimSpace(v), "IV=") {
					m3u8iv = strings.ReplaceAll(strings.ReplaceAll(v, "IV=", ""), `"`, ``)
				} else if strings.HasPrefix(strings.TrimSpace(v), "URI=") {
					m3u8uri = strings.ReplaceAll(strings.ReplaceAll(v, "URI=", ""), `"`, ``)
				}
			}
			break
		}
	}
	m3u8file.Close()

	// 開始爬蟲並下載mp4片段至資料夾
	hlsDL := NewHLS(m3u8url, nil, exportPATH, runtime.NumCPU(), true, id+".MP4")
	segments, err := hlsDL.Download(m3u8uri, m3u8iv)
	if err != nil {
		os.RemoveAll(segmentsPATH)
		panic(err)
	}

	// 合併 mp4
	mergeBar(exportPATH, id, segments)

	// 刪除子mp4
	err = os.RemoveAll(segmentsPATH)
	if err != nil {
		fmt.Printf("[DELETE] %v \n ", err)
	}

	// 取得封面
	getCover(pageSource, exportFILE)

	//TODO - 轉檔

	// move completed file from inprogress folder to completed folder
	os.Rename(exportPATH, strings.ReplaceAll(exportPATH, INPROGRESS_PATH, COMPLETED_PATH))

	return status
}
