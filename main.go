package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Parameters struct {
	Mutex            *sync.RWMutex
	LastRequestTime  time.Time
	Timer            time.Time
	BaseUrl          string
	UploadFilename   string
	DownloadFilename string
	ETag             string
}

var parameters = Parameters{
	Mutex:            &sync.RWMutex{},
	LastRequestTime:  time.Time{},
	BaseUrl:          "http://108.61.245.170",
	DownloadFilename: "image1.jpg",
	UploadFilename:   "image2.jpg",
	ETag:             "",
}

func main() {

	requestImage()
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.HandleFunc("/", RootImageHandler)

	// run scheduler
	go ScheduleRequest()

	// start server
	err := http.ListenAndServe(":9000", nil) // задаем слушать порт
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Handler
func RootImageHandler(w http.ResponseWriter, r *http.Request) {
	var content string
	parameters.Mutex.RLock()
	defer parameters.Mutex.RUnlock()
	content = fmt.Sprintf(`<html><body><img src="/images/%s" /></body></html>`, parameters.UploadFilename)
	etag := r.Header.Get("If-None-Match")

	if etag == parameters.ETag {
		w.Header().Add("Refresh", "1")
		w.Header().Add("Cache-Control", "max-age=10")
		w.Header().Add("Etag", parameters.ETag)
		fmt.Print(".")
		w.WriteHeader(304)
	} else {
		w.Header().Add("Cache-Control", "max-age=10")
		w.Header().Add("Etag", parameters.ETag)
		w.Header().Add("Refresh", "1")
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(content)))
		fmt.Println("request Etag  ", etag, " ==   ", parameters.ETag)
		fmt.Fprintf(w, content)
	}
}

// Scheduler
func ScheduleRequest() {

	timer := time.NewTimer(10 * time.Second)
	go func(timer *time.Timer) {
		//fmt.Println("wait  timer ")
		parameters.Timer = <-timer.C
		requestImage()
		ScheduleRequest()
	}(timer)
}
