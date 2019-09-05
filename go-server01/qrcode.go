package main

import (
	"io"
	"log"
	"net/http"
	"os"

	genqrcode "github.com/skip2/go-qrcode"
)

func createQrCode(content string) ([]byte, error) {
	var png []byte
	png, err := genqrcode.Encode(content, genqrcode.Medium, 256)
	if err != nil {
		log.Println("createQrCode Error:", err)
	}
	return png, err
}

func verifyQrCode(content string) bool {
	return true
}

func readQrCode(account string) string {
	/*fileName := fmt.Sprintf("qrcode/%s.jpg", account)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		url := fmt.Sprintf("%s/%s.jpg", CONFIGS.Person.ImageServer, account)
		log.Println("Qr Code not exsits. Downloading.", url)
		err = DownloadFile(fileName, url)
		if err != nil {
			log.Println("DownloadFile Error:", err)
			return ""
		}
	}
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		log.Println("DownloadFile Error Failed:", err)
		return ""
	}
	fin, err := os.Open(fileName)
	if err != nil {
		log.Println("Can't open qrcode image", err)
		return ""
	}
	defer fin.Close()

	src, err := jpeg.Decode(fin)
	if err != nil {
		log.Println("jpeg decode qrcode image", err)
		src, err = png.Decode(fin)
		if err != nil {
			log.Println("png decode qrcode image", err)
			return ""
		}
	}

	img := barcode.NewImage(src)
	scanner := barcode.NewScanner().SetEnabledAll(true)

	symbols, err := scanner.ScanImage(img)
	if err != nil {
		log.Println("ScanImage error", err)
		return ""
	}
	data := ""
	for _, s := range symbols {
		data = s.Data
		fmt.Printf("Name:%s Data:%s Quality:%v Boundary:%v\n", s.Type.Name(), s.Data, s.Quality, s.Boundary)
	}
	return data*/
	return ""
}

func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
