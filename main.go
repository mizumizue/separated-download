package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"separated-download/downloader"
)

func main() {
	url := "https//yahoo.co.jp" // TODO Change url is acceptable `Range` Request
	dl := downloader.NewDownloader(downloader.NewHttpClient(&http.Client{}))
	f, err := dl.Download(url)
	if err != nil {
		log.Fatalln(err)
	}
	if err := ioutil.WriteFile("tmp/yahoo"+f.Type, f.Data, os.ModePerm); err != nil {
		log.Fatalln(err)
	}
	log.Printf("download is done. content length: %d", len(f.Data))
}
