package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mtarnawa/godesu"
)

type finishState struct {
	filename string
	err      error
}

var ( // opts
	help              = flag.Bool("h", false, "Show help menu")
	useOrigFilename   = flag.Bool("o", false, "Download with the original filename")
	customDownloadDir = flag.String("c", "", "Set a custom directory for the images to download to")
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return
	}

	urls := strings.Split(args[0], " ")
	origDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Could not grab working directory! | %v\n", err)
		return
	}
	Gochan := godesu.New() // intialize godesu

	for urlNum, url := range urls { // loop through all urls
		purl := strings.Split(url, "/")
		ThreadNum, err := strconv.Atoi(purl[5])
		if err != nil {
			fmt.Printf("Make sure the URL is correct! | %v\n", err)
			return
		}

		err, Thread := Gochan.Board(purl[3]).GetThread(ThreadNum)
		if err != nil {
			fmt.Printf("Could not fetch thread! |  %v\n", err)
			return
		}

		images := Thread.Images()
		finishStateChan := make(chan finishState, len(images)) // make the download channel with proper buffer size

		if *customDownloadDir != "" {
			if err := os.Chdir(*customDownloadDir + "/"); err != nil {
				if err := os.MkdirAll(*customDownloadDir+"/", os.ModePerm); err != nil {
					fmt.Printf("Cannot create directory! %v\n", err)
					return
				}

				os.Chdir(*customDownloadDir + "/")
			}
		} else {
			if err := os.Chdir(purl[3] + "/" + purl[5]); err != nil {
				if err := os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm); err != nil {
					fmt.Printf("Cannot create directory! %v\n", err)
					return
				}

				os.Chdir(purl[3] + "/" + purl[5])
			}
		}

		fmt.Printf("Downloading '%v' %v of %v\n", url, urlNum+1, len(urls))

		for _, image := range images { // get the images downloading
			go func(image godesu.Image) {
				var fs finishState
				if *useOrigFilename {
					fs.filename = image.OriginalFilename
				} else {
					fs.filename = image.Filename + image.Extension
				}

				if _, err := os.Stat(fs.filename); err == nil {
					fs.err = fmt.Errorf("'%v' exists! Skipping...", fs.filename)
					finishStateChan <- fs
					return
				}

				resp, err := http.Get(image.URL)
				if err != nil {
					fs.err = fmt.Errorf("Error downloading '%v'! %v", image.URL, err)
					finishStateChan <- fs
					return
				} else if resp.StatusCode != http.StatusOK {
					fs.err = fmt.Errorf("Error downloading '%v'! Http status not ok: %d", image.URL, resp.StatusCode)
					finishStateChan <- fs
					return
				}
				defer resp.Body.Close()

				tmpFilename := fs.filename + ".part"

				file, err := os.Create(tmpFilename)
				if err != nil {
					fs.err = fmt.Errorf("Cannot create '%v'! %v", tmpFilename, err)
					finishStateChan <- fs
					return
				}
				defer file.Close()

				io.Copy(file, resp.Body)

				if err := os.Rename(tmpFilename, fs.filename); err != nil {
					fs.err = fmt.Errorf("Unable to rename '%v' to '%v'! %v", tmpFilename, fs.filename, err)
				}

				finishStateChan <- fs
				return
			}(image)
		}

		for i := 0; i < len(images); i++ { // watch for images to finish
			fs := <-finishStateChan
			if fs.err != nil {
				fmt.Printf("%v %v of %v\n", fs.err, i+1, len(images))
			} else {
				fmt.Printf("Finished downloading '%v' %v of %v\n", fs.filename, i+1, len(images))
			}
		}

		close(finishStateChan)
		os.Chdir(origDir)
	}
}
