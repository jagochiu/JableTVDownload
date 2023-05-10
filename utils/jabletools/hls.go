package jabletools

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/grafov/m3u8"
	"gopkg.in/cheggaaa/pb.v1"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type HLS struct {
	client    *http.Client
	headers   map[string]string
	dir       string
	URL       string
	workers   int
	bar       *pb.ProgressBar
	enableBar bool
	filename  string
	Key       []byte
	IV        []byte
}

type Segment struct {
	*m3u8.MediaSegment
	Path string
}
type DownloadResult struct {
	Err   error
	SeqId uint64
}

func NewHLS(url string, headers map[string]string, dir string, workers int, enableBar bool, filename string) *HLS {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if filename == "" {
		filename = time.Now().Format("20060102150405") + ".MP4"
	}
	return &HLS{
		URL:       url,
		dir:       dir,
		client:    &http.Client{},
		workers:   workers,
		enableBar: enableBar,
		headers:   headers,
		filename:  filename,
	}
}
func printStruct(v interface{}) {
	d, _ := json.Marshal(v)
	fmt.Println(string(d))
}

/*
TODO - Operations
*/
func wait(wg *sync.WaitGroup) chan bool {
	c := make(chan bool, 1)
	go func() {
		wg.Wait()
		c <- true
	}()
	return c
}
func (h *HLS) downloadSegments(segments []*Segment) error {

	wg := &sync.WaitGroup{}
	wg.Add(h.workers)

	finishedChan := wait(wg)
	quitChan := make(chan bool)
	segmentChan := make(chan *Segment)
	downloadResultChan := make(chan *DownloadResult, h.workers)

	for i := 0; i < h.workers; i++ {
		go func() {
			defer wg.Done()
			for segment := range segmentChan {
				tried := 0
			DOWNLOAD:
				tried++
				select {
				case <-quitChan:
					return
				default:
				}
				if err := h.downloadSegment(segment); err != nil {
					if strings.Contains(err.Error(), "connection reset by peer") && tried < 3 {
						time.Sleep(time.Second)
						log.Println("retry download segment ", segment.SeqId)
						goto DOWNLOAD
					}
					downloadResultChan <- &DownloadResult{Err: err, SeqId: segment.SeqId}
					return
				}
				downloadResultChan <- &DownloadResult{SeqId: segment.SeqId}
			}
		}()
	}
	go func() {
		defer close(segmentChan)
		for _, segment := range segments {
			segName := fmt.Sprintf("seg%d.ts", segment.SeqId)
			segment.Path = filepath.Join(h.dir, `segments`, segName)
			select {
			case segmentChan <- segment:
			case <-quitChan:
				return
			}
		}

	}()
	if h.enableBar {
		h.bar = pb.New(len(segments)).SetMaxWidth(100).Prefix("    Downloading... ")
		h.bar.ShowElapsedTime = true
		h.bar.Start()
	}
	defer func() {
		if h.enableBar {
			h.bar.Finish()
		}
	}()
	for {
		select {
		case <-finishedChan:
			return nil
		case result := <-downloadResultChan:
			if result.Err != nil {
				close(quitChan)
				return result.Err
			}
			if h.enableBar {
				h.bar.Increment()
			}
		}
	}
}
func (h *HLS) downloadSegment(segment *Segment) error {
	req, err := newRequest(segment.URI, h.headers)
	if err != nil {
		return err
	}

	res, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if segment.Key != nil {
		h.IV = []byte(segment.Key.IV)
		if len(h.IV) == 0 {
			h.IV = defaultIV(segment.SeqId)
		}
		if h.Key == nil || len(h.Key) <= 0 {
			h.Key, h.IV, err = h.getKey(segment)
			if err != nil {
				return err
			}
		}
		data, err = decryptAES128(data, h.Key, h.IV)
		if err != nil {
			return err
		}
	}
	for j := 0; j < len(data); j++ {
		if data[j] == syncByte {
			data = data[j:]
			break
		}
	}
	return os.WriteFile(segment.Path, data, 0755)
}

func (h *HLS) Download(m3u8uri, m3u8iv string) ([]*Segment, error) {
	segs, err := parseHlsSegments(h.URL, h.headers)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(h.dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = h.downloadSegments(segs)
	if err != nil {
		return nil, err
	}
	return segs, nil
}
