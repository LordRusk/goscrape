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

// opts
var useOrigFilename = flag.Bool("o", false, "Download with the original filename")
var customDownloadDir = flag.String("c", "", "Set a custom directory for the images to download to")

func main() {
	flag.Parse()
	urls := flag.Args()
	if len(urls) < 1 {
		flag.Usage()
		os.Exit(0)
	}

	origDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Could not grab working directory: %s\n", err)
		os.Exit(1)
	}

	Gochan := godesu.New() // intialize godesu

	for urlNum, url := range urls { // loop through all urls
		purl := strings.Split(url, "/")
		if len(purl) < 6 { // check length to avoid runtime error
			fmt.Println("Invalid URL")
			os.Exit(1)
		}

		ThreadNum, err := strconv.Atoi(purl[5])
		if err != nil {
			fmt.Printf("Inavlid URL: %s\n", err)
			os.Exit(1)
		}

		err, Thread := Gochan.Board(purl[3]).GetThread(ThreadNum)
		if err != nil {
			fmt.Printf("Could not fetch thread: %s\n", err)
			os.Exit(1)
		}

		images := Thread.Images()
		finishStateChan := make(chan finishState, len(images)) // make the download channel with proper buffer size

		if *customDownloadDir != "" {
			if err := os.MkdirAll(*customDownloadDir+"/", os.ModePerm); err != nil {
				fmt.Printf("Cannot create '%s': %s\n", *customDownloadDir+"/", err)
				os.Exit(1)
			}

			if err := os.Chdir(*customDownloadDir + "/"); err != nil { // this should never return an error
				fmt.Printf("Cannot change into '%s': %s\n", *customDownloadDir+"/", err)
				os.Exit(2)
			}
		} else {
			if err := os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm); err != nil {
				fmt.Printf("Cannot create '%s': %s\n", purl[3]+"/"+purl[5], err)
				os.Exit(1)
			}

			if err := os.Chdir(purl[3] + "/" + purl[5]); err != nil { // this should never return an error
				fmt.Printf("Cannot change into '%s': %s\n", purl[3]+"/"+purl[5], err)
				os.Exit(2)
			}
		}

		fmt.Printf("Downloading '%s' %d of %d\n", url, urlNum+1, len(urls))
		for _, image := range images { // start downloading images
			go func(image godesu.Image) {
				var fs finishState
				if *useOrigFilename {
					fs.filename = image.OriginalFilename
				} else {
					fs.filename = image.Filename + image.Extension
				}

				if _, err := os.Stat(fs.filename); err == nil {
					fs.err = fmt.Errorf("'%s' exists: Skipping...", fs.filename)
					finishStateChan <- fs
					return
				}

				resp, err := http.Get(image.URL)
				if err != nil {
					fs.err = fmt.Errorf("Error downloading '%s': %s", image.URL, err)
					finishStateChan <- fs
					return
				} else if resp.StatusCode != http.StatusOK {
					fs.err = fmt.Errorf("Error downloading '%s': http status not ok: %d", image.URL, resp.StatusCode)
					finishStateChan <- fs
					return
				}
				defer resp.Body.Close()

				tmpFilename := fs.filename + ".part"

				file, err := os.Create(tmpFilename)
				if err != nil {
					fs.err = fmt.Errorf("Cannot create '%s': %s", tmpFilename, err)
					finishStateChan <- fs
					return
				}
				defer file.Close()

				io.Copy(file, resp.Body)

				if err := os.Rename(tmpFilename, fs.filename); err != nil {
					fs.err = fmt.Errorf("Unable to rename '%s' to '%s': %s", tmpFilename, fs.filename, err)
				}

				finishStateChan <- fs
				return
			}(image)
		}

		for i := 0; i < len(images); i++ { // watch for images to finish
			fs := <-finishStateChan
			if fs.err != nil {
				fmt.Printf("%s %d of %d\n", fs.err, i+1, len(images))
			} else {
				fmt.Printf("Finished downloading '%s' %d of %d\n", fs.filename, i+1, len(images))
			}
		}

		close(finishStateChan)
		os.Chdir(origDir)
	}
}
