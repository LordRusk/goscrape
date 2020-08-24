package main

import (
	"os"
	"io"
	"fmt"
	"regexp"
	"strings"
	"net/http"

	"github.com/monaco-io/request"
)

func download(imageUrls []string, urlNum int, urlNumChan chan int) {
	pimageUrl := strings.Split(imageUrls[urlNum], "/") /* looks like [ i.4cdn.org g 244211.jpg] */
	filename := pimageUrl[2]

	if _, err := os.Stat(filename); err == nil {
		fmt.Println(filename, "exists! Skipping...")
		urlNumChan <- urlNum
		return
	}

	response, err := http.Get("https://"+imageUrls[urlNum])
	if err != nil {
		fmt.Println("Error downloading", filename)
		urlNumChan <- urlNum
		return
	}
	defer response.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error downloading", filename)
		urlNumChan <- urlNum
		return
	}
	defer file.Close()

	io.Copy(file, response.Body)

	urlNumChan <- urlNum
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Error! No url(s) specified!\n\nFirst arg: Url(s) *must be in quotes if multiple urls*\nSecond arg: Custom download directory\n")
		os.Exit(0)
	}

	urls := strings.Split(os.Args[1], " ")

	origDir, _ := os.Getwd()

	dlc := make(chan int)

	/* loop through all urls */
	for i := 0; i < len(urls); i++ {
		os.Chdir(origDir)

		url := urls[i]

		/* grab page info */
		client := request.Client {
			URL: url,
			Method: "GET",
		}
		resp, err := client.Do()
		if err != nil {
			fmt.Println("Unable to reach", url + "!\nCheck your connection and make sure the url is correct.")
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
					fmt.Println("Error! Cannot create custom directory! Check permissions!")
					os.Exit(0)
				} else {
					os.Chdir(string(os.Args[2])+"/")
				}
			}
		} else {
			purl := strings.Split(url, "/") /* url looks like [https:  4chan.org g 4532123] */
			os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm)
			os.Chdir(purl[3]+"/"+purl[5])
		}

		fmt.Println("Downloading", url, i+1, "of", len(urls))

		/* download all the images */
		for i := 0; i < len(imageUrls); i++ {
			go download(imageUrls, i, dlc)
		}

		for i := 0; i < len(imageUrls); i++ {
			urlNum := <-dlc
			pimageUrl := strings.Split(imageUrls[urlNum], "/")
			fmt.Println("Finished downloading", pimageUrl[2], i+1, "of", len(imageUrls))
		}
	}
	close(dlc)
}
