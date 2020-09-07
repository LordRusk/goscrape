package main

import (
	"os"
	"io"
	"log"
	"errors"
	"strings"
	"strconv"
	"net/http"

	"github.com/mtarnawa/godesu"
)

type finishState struct {
	image godesu.Image
	err error
}

func download(image godesu.Image, finishStateChan chan finishState) {
	var filename string
	filename = (image.Filename+image.Extension)

	if _, err := os.Stat(filename); err == nil {
		var err strings.Builder
		err.WriteString(filename)
		err.WriteString(" exists! Skipping...")

		fs := finishState {
			image: image,
			err: errors.New(err.String()),
		}

		finishStateChan <- fs
		return
	}

	response, err := http.Get(image.URL)
	if err != nil {
		var err strings.Builder
		err.WriteString("Error downloading ")
		err.WriteString(filename)

		fs := finishState {
			image: image,
			err: errors.New(err.String()),
		}

		finishStateChan <- fs
		return
	}

	/* Use a temp file name to avoid half downloaded
	 * images if goscrape were to be killed then restarted
	 * on the same Thread(s) caused by goscrape skipping
	 * images that already exist.
	 */
	var tmpFilename strings.Builder
	tmpFilename.WriteString(filename)
	tmpFilename.WriteString(".part")

	file, err := os.Create(tmpFilename.String())
	if err != nil {
		var err strings.Builder
		err.WriteString("Error downloading ")
		err.WriteString(filename)

		fs := finishState {
			image: image,
			err: errors.New(err.String()),
		}
		finishStateChan <- fs
		return
	}

	io.Copy(file, response.Body)

	if err := os.Rename(tmpFilename.String(), filename); err != nil {
		log.Println(err)
	}

	fs := finishState {
		image: image,
		err: nil,
	}

	finishStateChan <- fs
	return
}

func main() {
	if len(os.Args) < 2 {
		log.Printf("Error! No url(s) specified!\n\nFirst arg: Url(s) *must be in quotes if multiple urls*\nSecond arg: Custom download directory\n")
		os.Exit(1)
	}

	urls := strings.Split(os.Args[1], " ")
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
		}
		images := Thread.Images()

		/* directory stuff */
		if len(os.Args) > 2 {
			if err := os.Chdir(string(os.Args[2])+"/"); err != nil {
				if err := os.MkdirAll(string(os.Args[2]+"/"), os.ModePerm); err != nil {
					log.Println("Error! Cannot create custom directory! Check permissions!")
					os.Exit(1)
				} else {
					os.Chdir(string(os.Args[2])+"/")
				}
			}
		} else {
			if err := os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm); err != nil {
				log.Println("Error! Cannot create directory! Check permissions")
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
				log.Println("Finished downloading", fs.image.Filename+fs.image.Extension, i+1, "of", len(images))
			}
		}
	}
	close(dlc)
}
