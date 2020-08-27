package main

import (
	"os"
	"io"
	"log"
	"regexp"
	"errors"
	"strings"
	"net/http"

	"github.com/monaco-io/request"
)

type finishState struct {
	filename string
	err error
}

func download(imageUrls []string, urlNum int, finishStateChan chan finishState) {
	pimageUrl := strings.Split(imageUrls[urlNum], "/") /* looks like [ i.4cdn.org g 244211.jpg] */
	filename := pimageUrl[2]

	if _, err := os.Stat(filename); err == nil {
		var err strings.Builder
		err.WriteString(filename)
		err.WriteString(" exists! Skipping...")

		fs := finishState {
			filename: filename,
			err: errors.New(err.String()),
		}

		finishStateChan <- fs
		return
	}

	response, err := http.Get("https://"+imageUrls[urlNum])
	if err != nil {
		var err strings.Builder
		err.WriteString("Error downloading ")
		err.WriteString(filename)

		fs := finishState {
			filename: filename,
			err: errors.New(err.String()),
		}

		finishStateChan <- fs
		return
	}
	defer response.Body.Close()

	/* Use a temp file name to avoid half downloaded
	 * images if goscrape were to be killed then restarted
	 * on the same thread(s) caused by goscrape skipping
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
			filename: filename,
			err: errors.New(err.String()),
		}
		finishStateChan <- fs
		return
	}
	defer file.Close()

	io.Copy(file, response.Body)

	if err := os.Rename(tmpFilename.String(), filename); err != nil {
		log.Println(err)
	}

	fs := finishState {
		filename: filename,
		err: nil,
	}

	finishStateChan <- fs
	return
}

func main() {
	if len(os.Args) < 2 {
		log.Printf("Error! No url(s) specified!\n\nFirst arg: Url(s) *must be in quotes if multiple urls*\nSecond arg: Custom download directory\n")
		os.Exit(0)
	}

	urls := strings.Split(os.Args[1], " ")

	origDir, _ := os.Getwd()

	dlc := make(chan finishState)

	/* loop through all urls */
	for urlNum, url := range urls {
		os.Chdir(origDir)

		/* grab page info */
		client := request.Client {
			URL: url,
			Method: "GET",
		}
		resp, err := client.Do()
		if err != nil {
			log.Println("Unable to reach", url + "!\nCheck your connection and make sure the url is correct.")
			os.Exit(0)
		}

		/* get links to the images */
		lpat1 := regexp.MustCompile(`i\.4cdn\.org/[a-z]+/[0-9]*\.(png|jpg|gif|webm)`)
		lpat2 := regexp.MustCompile(`is2\.4chan\.org/[a-z]+/[0-9]*\.(png|jpg|gif|webm)`)
		imageUrls1 := lpat1.FindAllString(string(resp.Data), -1)
		imageUrls2 := lpat2.FindAllString(string(resp.Data), -1)
		imageUrls1 = append(imageUrls1, imageUrls2...)

		/* remove duplicate links caused by thumbnails */
		imageMap := make(map[string]bool)
		for _, item := range imageUrls1 {
			if _, ok := imageMap[item]; !ok {
				imageMap[item] = true
			}
		}
      		var imageUrls []string
      		for item, _ := range imageMap { imageUrls = append(imageUrls, item) }

		/* directory stuff */
		if len(os.Args) > 2 {
			if err := os.Chdir(string(os.Args[2])+"/"); err != nil {
				if err := os.MkdirAll(string(os.Args[2]+"/"), os.ModePerm); err != nil {
					log.Println("Error! Cannot create custom directory! Check permissions!")
					os.Exit(0)
				} else {
					os.Chdir(string(os.Args[2])+"/")
				}
			}
		} else {
			purl := strings.Split(url, "/") /* url looks like [https:  4chan.org g 4532123] */
			if err := os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm); err != nil {
				log.Println("Error! Cannot create directory! Check permissions")
				os.Exit(0)
			}
			os.Chdir(purl[3]+"/"+purl[5])
		}

		log.Println("Downloading", url, urlNum+1, "of", len(urls))

		/* download all the images */
		for urlNum, _ := range imageUrls {
			go download(imageUrls, urlNum, dlc)
		}

		for i := 0; i < len(imageUrls); i++ {
			fs := <-dlc
			if fs.err != nil {
				log.Println(fs.err, fs.filename, i+1, "of", len(imageUrls))
			} else {
				log.Println("Finished downloading", fs.filename, i+1, "of", len(imageUrls))
			}
		}
	}
	close(dlc)
}
