# Simple 4chan image scraper written in GO
Simple 4chan scraper written in go that supports multiple treads and custom download directories.

## How to install
`go get -u github.com/lordrusk/goscrape`

## How to use
run `goscrape` with the first argument being the link to the thread. Put in quotes for multiple links. The second argument can be a custom downloads directory, if none is given it will download it to `board/postid/`.

## ToDo
I need to change the way I download the images from using `net/http` to using `github.com/monaco-io/request`.

## FQA - Frequently Questioned Answers
+ Why do you use `github.com/monaco-io/request` as the request library and not `net/http`?

Because `/net/http` doesn't give me the information I need from 4chan.
