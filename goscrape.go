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

func download(imageurls []string, urlnum int, c chan int) {
	pimageurl := strings.Split(imageurls[urlnum], "/") /* looks like [ i.4cdn.org g 244211] */
	filename := pimageurl[2]

	response, err := http.Get("https://"+imageurls[urlnum])
	if err != nil { fmt.Println("Error downloading", filename) }
	defer response.Body.Close()

	file, err := os.Create(filename)
	if err != nil { fmt.Println("Error downloading", filename) }
	defer file.Close()

	io.Copy(file, response.Body)

	c <- urlnum
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Error! No url(s) specified!\n\nFirst arg: Url(s) *must be in quotes if multiple urls*\nSecond arg: Custom download directory\n")
		os.Exit(0)
	}

	urls := strings.Split(os.Args[1], " ")

	origdir, _ := os.Getwd()

	dlc := make(chan int)

	/* loop through all urls */
	for i := 0; i < len(urls); i++ {
		os.Chdir(origdir)

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
		imageurls1 := lpat1.FindAllString(string(resp.Data), -1)
		imageurls2 := lpat2.FindAllString(string(resp.Data), -1)
		imageurls1 = append(imageurls1, imageurls2...)

		/* remove duplicate links caused by thumbnails */
		imagemap := make(map[string]bool)
		for _, item := range imageurls1 {
			if _, ok := imagemap[item]; !ok {
				imagemap[item] = true
			}
		}
      		var imageurls []string
      		for item, _ := range imagemap { imageurls = append(imageurls, item) }

		/* directory stuff */
		if len(os.Args) > 2 {
			if err := os.Chdir(string(os.Args[2])+"/"); err != nil {
				fmt.Println("Error! Directory does not exist!")
				os.Exit(0)
			}
		} else {
			purl := strings.Split(url, "/") /* url looks like [https:  4chan.org g 4532123] */
			os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm)
			os.Chdir(purl[3]+"/"+purl[5])
		}

		fmt.Println("Downloading", url)

		/* download all the images */
		for i := 0; i < len(imageurls); i++ {
			go download(imageurls, i, dlc)
		}

		for i := 0; i < len(imageurls); i++ {
			c := <- dlc
			pimageurl := strings.Split(imageurls[c], "/") /* looks like [ i.4cdn.org g 244211] */
			fmt.Println("Finished downloading", pimageurl[2])
		}
	}
	close(dlc)
}
