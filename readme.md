[![Go Report Card](https://goreportcard.com/badge/github.com/lordrusk/goscrape)](https://goreportcard.com/report/github.com/lordrusk/goscrape)

# Simple and efficient 4chan media scraper written in GO
Goscrape is a very simple and efficient 4chan media scraper written in go that supports multiple treads, custom download directories, and more!

## How to install
`go get github.com/lordrusk/goscrape`

## How to use
`-h` for help menu.

`-o` for original filenames.

`-c` to set a custom directory.

Put in quotes for multiple threads.

## Features
* Goscrape is upwards of 4x faster then other scrapers, goscrape does this by using go's concurrency to download multiple images at the same time, taking advantage of more bandwidth. You won't find download speeds like this anywhere else.
* Goscrape uses the [godesu](https://github.com/mtarnawa/godesu) 4chan read-only api to interact with 4chan. The old system used a request library, regex, maps, etc, but this seems to fit more with the go style.
* Goscrape is cross-platform, works everywhere that go does out of the box! (Plan9, etc, etc)

## Why?
Because *all* 4chan scrapers I've seen and used were written in python, I dislike python and wanted to make something in GO.
