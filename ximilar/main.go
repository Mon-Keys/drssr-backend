package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"ximilar/client"
)

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	xc := client.New("https://api.ximilar.com", "3e16ad7dfc2e13da0ffcc5deb15f7b0c4eedadc9", time.Second*5)

	filePath := "./test.jpeg"
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %w", err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Failed to read file: %w", err)
	}

	encodedData := base64.RawStdEncoding.EncodeToString(data)

	resp, err := xc.RemoveBGPrecise(context.Background(), &client.RemoveBGReq{
		Records: []client.RecordsReqStruct{
			{
				Base64: encodedData,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to send request: %w", err)
	}

	resFileName := "test_res.png"
	err = downloadFile(resp.Records[0].OutputURL, resFileName)
	if err != nil {
		log.Fatalf("Failed to save mask file: %w", err)
	}

	respColor, err := xc.DominantColorProduct(context.Background(), &client.DominantColotReq{
		ColorNames: true,
		Records: []client.DominantColotRecordReqStruct{
			{
				Base64: encodedData,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to send color request: %w", err)
	}

	fmt.Println(respColor.Records[0].DominantColors.Percentages)
	fmt.Println(respColor.Records[0].DominantColors.ColorNames)
	fmt.Println(respColor.Records[0].DominantColors.ColorNamesPantone)
}
