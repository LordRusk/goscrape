package main

import (
	"os"
	"io"
	"log"
	"errors"
	"strings"
	"strconv"
	"net/http"

	"github.com/pborman/getopt"
	"github.com/mtarnawa/godesu"
)

type finishState struct {
	filename string
	err error
}

var (
	/* opts */
	help = getopt.BoolLong("help", 0, "Help")
	useOrigFilename = getopt.BoolLong("useOrigFilename", 'o', "Download with the original filename")
	customDownloadDir = getopt.StringLong("customDownloadDir", 'c', "", "Set a custom directory for the images to download to")
)

func download(image godesu.Image, finishStateChan chan finishState) {
	var filename string
	if *useOrigFilename {
		filename = image.OriginalFilename
	} else {
		filename = (image.Filename+image.Extension)
	}

	fs := finishState { filename: filename }

	if _, err := os.Stat(filename); err == nil {
		var err strings.Builder
		err.WriteString("'")
		err.WriteString(filename)
		err.WriteString("'")
		err.WriteString(" exists! Skipping...")

		fs.err = errors.New(err.String()) }
		finishStateChan <- fs
		return
	}

	response, err := http.Get(image.URL)
	if err != nil {
		var err strings.Builder
		err.WriteString("Error downloading ")
		err.WriteString("'")
		err.WriteString(filename)
		err.WriteString("'")

		fs := finishState { err: errors.New(err.String()) }
		finishStateChan <- fs
		return
	}

	var tmpFilename strings.Builder
	tmpFilename.WriteString(filename)
	tmpFilename.WriteString(".part")

	file, err := os.Create(tmpFilename.String())
	if err != nil {
		var err strings.Builder
		err.WriteString("Error downloading ")
		err.WriteString("'")
		err.WriteString(filename)
		err.WriteString("'")

		fs.err = errors.New(err.String())
		finishStateChan <- fs
		return
	}

	io.Copy(file, response.Body)

	if err := os.Rename(tmpFilename.String(), filename); err != nil {
		log.Println(err)
	}

	fs.err = nil
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
	if len(args) == 0 {
		getopt.Usage()
		os.Exit(1)
	}
	urls := strings.Split(args[0], " ")
	origDir, _ := os.Getwd()
	dlc := make(chan finishState)
	Gochan := godesu.New()

	/* loop through all urls */
	for urlNum, url := range urls {
		os.Chdir(origDir)

		/* godesu suff */
		purl := strings.Split(url, "/")
		ThreadNum, _ := strconv.Atoi(purl[5])
		err, Thread := Gochan.Board(purl[3]).GetThread(ThreadNum)
		if err != nil {
			log.Fatal(errors.New("Error! Could not fetch Thread!"))
			os.Exit(1)
		}
		images := Thread.Images()

		/* directory stuff */
		if *customDownloadDir != "" {
			if err := os.Chdir(*customDownloadDir+"/"); err != nil {
				if err := os.MkdirAll(*customDownloadDir+"/", os.ModePerm); err != nil {
					log.Fatal(errors.New("Error! Cannot create directory! Check permissions"))
					os.Exit(1)
				} else {
					os.Chdir(*customDownloadDir+"/")
				}
			}
		} else {
			if err := os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm); err != nil {
				log.Fatal(errors.New("Error! Cannot create directory! Check permissions"))
				os.Exit(1)
			}
			os.Chdir(purl[3]+"/"+purl[5])
		}

		log.Println("Downloading", url, urlNum+1, "of", len(urls))

		/* download all the images */
		for _, image := range images {
			go download(image, dlc)
		}

		for i := 0; i < len(images); i++ {
			fs := <-dlc
			if fs.err != nil {
				log.Println(fs.err, i+1, "of", len(images))
			} else {
				log.Println("Finished downloading", "'"+fs.filename+"'", i+1, "of", len(images))
			}
		}
	}
	close(dlc)
}
