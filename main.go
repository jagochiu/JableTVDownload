package main

import (
	"bufio"
	"io/ioutil"
	"jable/utils/common"
	"jable/utils/jabletools"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	// 讀取 download list 文件
	resources_file_path := `./resources`
	resources_file, err := os.Open(resources_file_path)
	if err != nil {
		log.Fatal(err)
	}
	// 取得到 download list
	scan := bufio.NewScanner(resources_file)
	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		// println(scan.Text())
		url := strings.TrimSpace(scan.Text())
		if !strings.HasPrefix(url, "#") && strings.HasPrefix(url, "http") {
			start := time.Now()
			status := jabletools.Download(url, `/Volumes/Jago/Downloads/AVs/Jable.TV`)
			switch status {
			case 429:
				del_URL(url, resources_file_path, true)
			default:
				del_URL(url, resources_file_path, false)
			}
			common.DurationCheck(start)
		}
	}
	resources_file.Close()
}
func del_URL(url, path string, common bool) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("[DELETE] %v \n ", err)
	}
	output := strings.ReplaceAll(string(input), url+"\n", "")
	if common {
		output += "\n# # " + url
	}
	err = ioutil.WriteFile(path, []byte(output), 0777)
	if err != nil {
		log.Printf("[DELETE] %v \n ", err)
	}
}
