package jabletools

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

func merge(dir, id string, segments []*Segment) (string, error) {
	log.Println("[MERGE] segments")
	start_time := time.Now()
	full_path := filepath.Join(dir, id+".MP4")
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].SeqId < segments[j].SeqId
	})
	f2, err := os.OpenFile(full_path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Printf("[ERROR] f2 %v\n", err)
		return ``, err
	}
	defer f2.Close()
	for _, seg := range segments {
		f1, err := os.Open(seg.Path)
		if err != nil {
			log.Printf("[ERROR] f1 %v\n", err)
			return ``, err
		}
		defer f1.Close()
		content, err := io.ReadAll(f1)
		if err != nil {
			log.Printf("[ERROR] f1 read %v\n", err)
			return ``, err
		}
		_, err = f2.Write(content)
		if err != nil {
			log.Printf("[ERROR] f2 write %v\n", err)
			return ``, err
		}
	}
	duration := time.Since(start_time)
	minutes := int(duration.Minutes())
	seconds := duration.Seconds() - float64(minutes*60)
	log.Printf("It took %dm %.2fs to render the video. \n", minutes, seconds)
	return full_path, nil
}

type MergeResult struct {
	Err   error
	SeqId uint64
}

func mergeBar(dir, id string, segments []*Segment) (string, error) {
	log.Println("[MERGE] segments")
	start_time := time.Now()
	full_path := filepath.Join(dir, id+".MP4")
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].SeqId < segments[j].SeqId
	})
	f2, err := os.OpenFile(full_path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Printf("[ERROR] f2 %v\n", err)
		return ``, err
	}
	defer f2.Close()

	var bar *pb.ProgressBar

	wg := &sync.WaitGroup{}
	wg.Add(1)

	finishedChan := wait(wg)
	quitChan := make(chan bool)
	segmentChan := make(chan *Segment)
	mergeResultChan := make(chan *MergeResult)

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
			if err := func(seg *Segment) error {
				f1, err := os.Open(seg.Path)
				if err != nil {
					log.Printf("[ERROR] %s open %v\n", seg.Path, err)
					return err
				}
				defer f1.Close()
				content, err := io.ReadAll(f1)
				if err != nil {
					log.Printf("[ERROR] %s read %v\n", seg.Path, err)
					return err
				}
				_, err = f2.Write(content)
				if err != nil {
					log.Printf("[ERROR] output file write %v\n", err)
					return err
				}
				return nil
			}(segment); err != nil {
				if tried < 3 {
					time.Sleep(time.Second)
					log.Println("retry merge segment ", segment.SeqId)
					goto DOWNLOAD
				}
				mergeResultChan <- &MergeResult{Err: err, SeqId: segment.SeqId}
				return
			}
			mergeResultChan <- &MergeResult{SeqId: segment.SeqId}
		}
	}()

	go func() {
		defer close(segmentChan)
		for _, segment := range segments {
			select {
			case segmentChan <- segment:
			case <-quitChan:
				return
			}
		}
	}()

	bar = pb.New(len(segments)).SetMaxWidth(100).Prefix("    Merging... ")
	bar.ShowElapsedTime = true
	bar.Start()
	for {
		select {
		case <-finishedChan:
			bar.Finish()
			duration := time.Since(start_time)
			minutes := int(duration.Minutes())
			seconds := duration.Seconds() - float64(minutes*60)
			log.Printf("It took %dm %.2fs to render the video. \n", minutes, seconds)
			return full_path, nil
		case result := <-mergeResultChan:
			if result.Err != nil {
				close(quitChan)
				return ``, result.Err
			}
			bar.Increment()

		}
	}
}
