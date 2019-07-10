package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"time"
)

func requestImage() {
	response, err := http.Get(parameters.BaseUrl)
	if err != nil {
		fmt.Printf("%s", err)
	} else {
		defer response.Body.Close()
		node, err := html.Parse(response.Body)

		if err != nil {
			fmt.Printf("%s", err)
		}

		if imageNode, err := getImgNode(node); err != nil {
			fmt.Printf("%s", err)
		} else {
			for _, v := range imageNode.Attr {
				if v.Key == "src" {
					if err := verifyImage(parameters.BaseUrl + v.Val); err != nil {
						fmt.Println("Request error : ", err.Error())
						return
					}
					break
				}
			}
		}
		fmt.Printf("%s\n", string(node.Data))
	}
}

func verifyImage(url string) error {
	if headerRequest, err := http.Head(url); err != nil {
		return err
	} else {
		defer headerRequest.Body.Close()
		if lt, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", headerRequest.Header.Get("Last-Modified")); err != nil {
			return downloadImage(url, time.Now())
		} else {
			if parameters.LastRequestTime.IsZero() || !lt.Equal(parameters.LastRequestTime) {
				return downloadImage(url, lt)
			} else {
				return nil
			}
		}
	}
}

func downloadImage(url string, t time.Time) error {
	if response, err := http.Get(url); err != nil {
		return err
	} else {
		defer response.Body.Close()
		file, err := os.Create("images/" + parameters.DownloadFilename)
		if err != nil {
			return err
		}
		defer file.Close()

		parameters.Mutex.Lock()
		defer func() {
			s := parameters.UploadFilename
			parameters.UploadFilename = parameters.DownloadFilename
			parameters.DownloadFilename = s
			parameters.LastRequestTime = t
			parameters.ETag = fmt.Sprintf(`"%s"`, t.Format("20060102150405"))
			parameters.Mutex.Unlock()
		}()
		_, err = io.Copy(file, response.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func getImgNode(doc *html.Node) (*html.Node, error) {
	var b *html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			b = n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	if b != nil {
		return b, nil
	}
	return nil, errors.New("Missing <img> in the node tree")
}
