package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
)

const BoxCnt uint = 3

var (
	logInfo  *log.Logger
	logError *log.Logger
	boxes    map[string]horcrux
	cfg      *clientConfig
)

type Message struct {
	GflickerImage    string
	GoogleDriveImage string
	FlickerImage     string
}

func main() {
	// setup logging
	file, err := os.OpenFile("ghorcrux.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	multi := io.MultiWriter(file, os.Stdout)

	initLogging(multi)

	boxes = make(map[string]horcrux, 3)

	cfg = loadClientConfig()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/link", linkHandler)
	http.HandleFunc("/redirect", redirectHandler)
	http.ListenAndServe(":9999", nil)
}

func initLogging(w io.Writer) {
	logInfo = log.New(w, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logError = log.New(w, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if !cfg.UsingGdrive && !cfg.UsingDropbox && !cfg.UsingFlickr {
			msg := &Message{}
			renderLinkPage(w, "templates/link.html", msg)
		} else {
			http.Redirect(w, r, "http://localhost:9999/redirect", http.StatusFound)
		}
	} else {
		logError.Println("invalid request: ", r.Method)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func getImageBytes(fname string) (string, error) {
	f, err := os.OpenFile(fname, os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("error while reading background image. err: %v", err)

	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return "", fmt.Errorf("error while decoding background image. err: %v", err)

	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return "", fmt.Errorf("error while encoding background image. err: %v", err)

	}

	str := base64.StdEncoding.EncodeToString(buffer.Bytes())

	return str, nil

}

type imgFile struct {
	img   string
	fname string
}

func renderLinkPage(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		logError.Printf("error while parsing template. file: %s, err: %v", filename, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}

	queue := make(chan imgFile)

	go func() {
		queue <- imgFile{"GflickerImage", "templates/gflicker.png"}
		queue <- imgFile{"GoogleDriveImage", "templates/googldrive.png"}
		queue <- imgFile{"FlickerImage", "templates/flickr.png"}
		close(queue)
	}()

	for elem := range queue {
		imgStr, err := getImageBytes(elem.fname)
		if err != nil {
			logError.Println(err.Error())
		} else {
			if _, ok := data.(*Message); ok {
				m := data.(*Message)
				if elem.img == "GflickerImage" {
					m.GflickerImage = imgStr
				} else if elem.img == "GoogleDriveImage" {
					m.GoogleDriveImage = imgStr
				} else if elem.img == "FlickerImage" {
					m.FlickerImage = imgStr
				}
				data = m
			} else {
				logError.Println("type check failed")
			}
		}
	}

	if err := tmpl.Execute(w, data); err != nil {
		logError.Printf("error while executing template. err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}
}

func linkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		logInfo.Println(r.Form.Get("gdrive"))
		logInfo.Println(r.Form.Get("dbox"))
		logInfo.Println(r.Form.Get("flickr"))

		if r.Form.Get("gdrive") != "" {
			gd := NewGDrive()
			boxes["gdrive"] = gd
			cfg.SetUsingGdrive()
		}

		for k, v := range boxes {
			logInfo.Printf("k: %s, v: %v", k, v)
			v.Link(w, r)
		}
	} else {
		logError.Println("invalid request: ", r.Method)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		logError.Printf("error while parsing template. file: %s, err: %v", filename, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}
	if err := tmpl.Execute(w, data); err != nil {
		logError.Printf("error while executing template. err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		gd := boxes["gdrive"]
		gd.SaveToken(r)
		gd.List()
		render(w, "templates/thanks.html", nil)
	} else {
		logError.Println("invalid request: ", r.Method)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
