/*
 * +===============================================
 * | Author:        Parham Alvani (parham.alvani@gmail.com)
 * |
 * | Creation Date: 24-11-2015
 * |
 * | File Name:     bing.go
 * +===============================================
 */
package bing

import (
	"encoding/json"
	"fmt"
	"github.com/franela/goreq"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

func getBingImage(path string, image Image, w sync.WaitGroup) {
	fmt.Printf("Getting %s\n", image.StartDate)

	if _, err := os.Stat(fmt.Sprintf("%s/%s.jpg", path, image.FullStartDate)); err == nil {
		fmt.Printf("%s is already exists\n", image.StartDate)
		w.Done()
		return
	}

	resp, err := goreq.Request{
		Uri: fmt.Sprintf("http://www.bing.com/%s", image.URL),
	}.Do()
	if err != nil {
		glog.Errorf("Net.HTTP: %v\n", err)
		w.Done()
		return
	}

	defer resp.Body.Close()

	dest_file, err := os.Create(fmt.Sprintf("%s/%s.jpg", path, image.FullStartDate))
	if err != nil {
		glog.Errorf("OS: %v\n", err)
		w.Done()
		return
	}

	defer dest_file.Close()

	io.Copy(dest_file, resp.Body)

	fmt.Printf("%s was gotten\n", image.StartDate)

	w.Done()
}

func GetBingDesktop(path string, idx int, n int) error {
	goreq.SetConnectTimeout(1 * time.Minute)
	// Create HTTP GET request
	resp, err := goreq.Request{
		Uri: "http://www.bing.com/HPImageArchive.aspx",
		QueryString: BingRequest{
			Format: "js",
			Index:  idx,
			Number: n,
			Mkt:    "en-US",
		},
		UserAgent: "GoSiMac",
	}.Do()
	if err != nil {
		glog.Errorf("Net.HTTP: %v\n", err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("IO.IOUtil: %v\n", err)
	}
	var bing_resp BingResponse
	json.Unmarshal(body, &bing_resp)

	var w sync.WaitGroup
	// Create spreate thread for each image
	for _, image := range bing_resp.Images {
		w.Add(1)
		go getBingImage(path, image, w)
	}

	// Waiting for getting all the images
	w.Wait()

	return nil
}
