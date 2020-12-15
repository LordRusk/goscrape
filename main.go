package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mtarnawa/godesu"
	"github.com/pborman/getopt"
)

type finishState struct {
	filename string
	err      error
}

var (
	// opts
	help              = getopt.BoolLong("help", 'h', "Help")
	useOrigFilename   = getopt.BoolLong("useOrigFilename", 'o', "Download with the original filename")
	customDownloadDir = getopt.StringLong("customDownloadDir", 'c', "", "Set a custom directory for the images to download to")
)

func main() {
	getopt.Parse()
	if *help {
		getopt.Usage()
		return
	}
	args := getopt.Args()
	if len(args) < 1 {
		getopt.Usage()
		return
	}
	urls := strings.Split(args[0], " ")
	origDir, _ := os.Getwd()
	Gochan := godesu.New()

	// loop through all urls
	for urlNum, url := range urls {
		purl := strings.Split(url, "/")
		ThreadNum, _ := strconv.Atoi(purl[5])
		err, Thread := Gochan.Board(purl[3]).GetThread(ThreadNum)
		if err != nil {
			fmt.Printf("Could not fetch thread! | %v\n", err)
			return
		}
		images := Thread.Images()

		// make the download channel with proper buffer size
		finishStateChan := make(chan finishState, len(images))

		if *customDownloadDir != "" {
			if err := os.Chdir(*customDownloadDir + "/"); err != nil {
				if err := os.MkdirAll(*customDownloadDir+"/", os.ModePerm); err != nil {
					fmt.Printf("Cannot create directory! | %v\n", err)
					return
				} else {
					os.Chdir(*customDownloadDir + "/")
				}
			}
		} else {
			if err := os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm); err != nil {
				fmt.Printf("Cannot create directory! | %v\n", err)
				return
			}
			os.Chdir(purl[3] + "/" + purl[5])
		}

		fmt.Printf("Downloading '%v' | %v of %v\n", url, urlNum+1, len(urls))

		// get the images downloading
		for _, image := range images {
			go func(image godesu.Image) {
				var filename string
				if *useOrigFilename {
					filename = image.OriginalFilename
				} else {
					filename = image.Filename + image.Extension
				}

				fs := finishState{filename: filename}

				if _, err := os.Stat(filename); err == nil {
					fs.err = fmt.Errorf("'%v' exists! Skipping...", filename)
					finishStateChan <- fs
					return
				}

				resp, err := http.Get(image.URL)
				if err != nil {
					fs.err = fmt.Errorf("Error downloading '%v'! | %v", image.URL, err)
					finishStateChan <- fs
					return
				} else if resp.StatusCode != http.StatusOK {
					fs.err = fmt.Errorf("Error downloading '%v'! Http status not ok: %v", image.URL, resp.StatusCode)
					finishStateChan <- fs
					return
				}
				defer resp.Body.Close()

				tmpFilename := filename + ".part"

				file, err := os.Create(tmpFilename)
				if err != nil {
					fs.err = fmt.Errorf("Cannot create '%v'! | %v", filename, err)
					finishStateChan <- fs
					return
				}

				io.Copy(file, resp.Body)

				if err := os.Rename(tmpFilename, filename); err != nil {
					fs.err = fmt.Errorf("Unable to rename '%v' to '%v'! | %v", tmpFilename, filename, err)
				}

				finishStateChan <- fs
				return
			}(image)
		}

		for i := 0; i < len(images); i++ {
			fs := <-finishStateChan
			if fs.err != nil {
				fmt.Printf("%v | %v of %v\n", fs.err, i+1, len(images))
			} else {
				fmt.Printf("Finished downloading '%v' | %v of %v\n", fs.filename, i+1, len(images))
			}
		}

		close(finishStateChan)
		os.Chdir(origDir)
	}
}
