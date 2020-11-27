package main

import (
	"errors"
	"fmt"
	"io"
	"log"
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
	help              = getopt.BoolLong("help", 0, "Help")
	useOrigFilename   = getopt.BoolLong("useOrigFilename", 'o', "Download with the original filename")
	customDownloadDir = getopt.StringLong("customDownloadDir", 'c', "", "Set a custom directory for the images to download to")
)

func download(image godesu.Image, finishStateChan chan<- finishState) {
	var filename string
	if *useOrigFilename {
		filename = image.OriginalFilename
	} else {
		filename = image.Filename + image.Extension
	}

	fs := finishState{filename: filename}

	if _, err := os.Stat(filename); err == nil {
		fs.err = errors.New("'" + filename + "' exists! Skipping...")
		finishStateChan <- fs
		return
	}

	resp, err := http.Get(image.URL)
	if err != nil {
		fs.err = errors.New("Error downloading '" + filename + "'")
		finishStateChan <- fs
		return
	}

	tmpFilename := filename + ".part"

	file, err := os.Create(tmpFilename)
	if err != nil {
		fs.err = errors.New("Error downloading '" + filename + "'")
		finishStateChan <- fs
		return
	}

	io.Copy(file, resp.Body)

	if err := os.Rename(tmpFilename, filename); err != nil {
		fmt.Println(err)
	}

	finishStateChan <- fs
	return
}

func main() {
	getopt.Parse()
	if *help {
		getopt.Usage()
		os.Exit(0)
	}
	args := getopt.Args()
	if len(args) < 1 {
		getopt.Usage()
		os.Exit(1)
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
			log.Fatal(errors.New("Error! Could not fetch Thread!"))
			os.Exit(1)
		}
		images := Thread.Images()

		// make the download chan with proper buffer size
		finishStateChan := make(chan finishState, len(images))

		if *customDownloadDir != "" {
			if err := os.Chdir(*customDownloadDir + "/"); err != nil {
				if err := os.MkdirAll(*customDownloadDir+"/", os.ModePerm); err != nil {
					log.Fatal(errors.New("Error! Cannot create directory! Check permissions"))
					os.Exit(1)
				} else {
					os.Chdir(*customDownloadDir + "/")
				}
			}
		} else {
			if err := os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm); err != nil {
				log.Fatal(errors.New("Error! Cannot create directory! Check permissions"))
				os.Exit(1)
			}
			os.Chdir(purl[3] + "/" + purl[5])
		}

		fmt.Println("Downloading", url, urlNum+1, "of", len(urls))

		// get the images downloading
		for _, image := range images {
			go download(image, finishStateChan)
		}

		for i := 0; i < len(images); i++ {
			fs := <-finishStateChan
			if fs.err != nil {
				fmt.Println(fs.err, i+1, "of", len(images))
			} else {
				fmt.Println("Finished downloading", "'"+fs.filename+"'", i+1, "of", len(images))
			}
		}
		close(finishStateChan)
		os.Chdir(origDir)
	}
}
