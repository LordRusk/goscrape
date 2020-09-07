# Simple 4chan image scraper written in GO
Simple 4chan scraper written in go that supports multiple treads and custom download directories.

## How to install
`go get -u github.com/lordrusk/goscrape`

## How to use
run `goscrape` with the first argument being the link to the thread. Put in quotes for multiple links. The second argument can be a custom downloads directory, if none is given it will download it to `board/postid/`.

## Features
* Goscrape is upwards of 4x faster then other scrapers, goscrape does this by using go's concurrency to download multiple images at the same time, taking advantage of more bandwidth. You won't find download speeds like this anywhere else.
* Goscrape uses the [godesu](https://github.com/mtarnawa/godesu) 4chan read-only api to interact with 4chan. The old system used a request library, regex, maps, etc, but this seems to fit more with the go style.
* Goscrape is cross-platform, works everywhere that go does out of the box! (Plan9, etc, etc)

## Why?
Because *all* 4chan scrapers I've seen and used were written in python, I dislike python and wanted to make something in GO.

### License
Goscrape is in the public domain.
