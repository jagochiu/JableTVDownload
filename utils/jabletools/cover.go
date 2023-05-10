package jabletools

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getCover(html, path string) {
	// 取得封面
	previewIMG := ``
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Printf("[ERROR] html Document - %v \n", err)
	}
	doc.Find(`meta`).Each(func(i int, el *goquery.Selection) {
		if property, ok := el.Attr(`property`); ok && property == `og:image` {
			if content, ok := el.Attr(`content`); ok {
				previewIMG = content
			}
		}
	})
	err = getIntoFile(previewIMG, path+`.jpg`)
	if err != nil {
		log.Printf("[Cover Image] %v \n", err)
	}
}
