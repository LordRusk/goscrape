package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"net/http"
	"io"

	"github.com/monaco-io/request"
)

func main() {
	/* check if there is a url given */
	if len(os.Args) < 2 {
		fmt.Printf("Error! No url(s) specified!\n\nFirst arg: Url(s) *must be in quotes if multiple urls*\nSecond arg: Custom download directory\n")
		os.Exit(0)
	}

	/* get urls from args */
	urls := strings.Split(os.Args[1], " ")

	/* save current directory */
	origdir, _ := os.Getwd()

	/* loop through all urls */
	for i := 0; i < len(urls); i++ {
		/* get back to original directory */
		os.Chdir(origdir)

		/* get url */
		url := urls[i]

		/* grab page info */
		client := request.Client {
			URL: url,
			Method: "GET",
		}
		resp, err := client.Do()
		if err != nil {
			fmt.Println("Invalid URL!")
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
		for _, item := range imageurls1 { if _, ok := imagemap[item]; !ok { imagemap[item] = true }}
      		var imageurls []string
      		for item, _ := range imagemap { imageurls = append(imageurls, item) }

		/* directory stuff */
		if len(os.Args) > 2 { if _, err := os.Stat(os.Args[2]); err != nil {
			fmt.Println("Error! Directory does not exist!")
			os.Exit(0)
		} else {
			os.Chdir(string(os.Args[2])+"/")
		}} else {
			purl := strings.Split(url, "/") /* url looks like [https:  4chan.org g 4532123] */
			os.MkdirAll(purl[3]+"/"+purl[5], os.ModePerm)
			os.Chdir(purl[3]+"/"+purl[5])
		}

		/* tell them what is downloading */
		fmt.Println("Downloading", url)

		/* download all the images */
		for i := 0; i < len(imageurls); i++ {
			/* parse link to get image name */
			pimageurl := strings.Split(imageurls[i], "/") /* looks like [ i.4cdn.org g 244211] */
			filename := pimageurl[2]

			/* tell them what is downloading */
			fmt.Println("Downloading", pimageurl[2]+"...", i+1, "of", len(imageurls))

			/* Get the response bytes from the url */
			response, err := http.Get("https://"+imageurls[i])
			if err != nil { fmt.Println("Error downloading", filename) }
			defer response.Body.Close()

			/* Create a empty file */
			file, err := os.Create(filename)
			if err != nil { fmt.Println("Error downloading", filename) }
			defer file.Close()

			/* Write the bytes to the file */
			_, err = io.Copy(file, response.Body)
		}
	fmt.Printf("\n")
	}
}
