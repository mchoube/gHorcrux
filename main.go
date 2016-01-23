package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"text/template"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var (
	logInfo  *log.Logger
	logError *log.Logger
)

func main() {
	// setup logging
	file, err := os.OpenFile("ghorcrux.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	multi := io.MultiWriter(file, os.Stdout)

	initLogging(multi)

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
		render(w, "templates/index.html", nil)
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

func linkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//ctx := context.Background()

		b, err := ioutil.ReadFile("gdrive_client_secret.json")
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)

		}

		config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)

		}

		cacheFile, err := tokenCacheFile()
		if err != nil {
			log.Fatalf("Unable to get path to cached credential file. %v", err)

		}
		tok, err := tokenFromFile(cacheFile)
		if err != nil {
			authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
			http.Redirect(w, r, authURL, http.StatusFound)
		} else {
			logInfo.Println("token: ", tok)
		}
		//	client := getClient(ctx, config, w, r)

		//	srv, err := drive.New(client)
		//	if err != nil {
		//		log.Fatalf("Unable to retrieve drive Client %v", err)
		//	}

		//	r, err := srv.Files.List().Do()
		//	if err != nil {
		//		log.Fatalf("Unable to retrieve files.", err)
		//	}

		//	fmt.Println("Files:")
		//	if len(r.Files) > 0 {
		//		for _, i := range r.Files {
		//			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		//		}
		//	} else {
		//		fmt.Print("No files found.")
		//	}
	} else {
		logError.Println("invalid request: ", r.Method)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config, w http.ResponseWriter, r *http.Request) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)

	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config, w, r)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)

}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config, w http.ResponseWriter, r *http.Request) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	//	if _, err := fmt.Scan(&code); err != nil {
	//		log.Fatalf("Unable to read authorization code %v", err)
	//
	//	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)

	}
	return tok

}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err

	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("drive-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	logInfo.Println("token file: ", file)
	f, err := os.Open(file)
	if err != nil {
		return nil, err

	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err

}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)

	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		render(w, "templates/thanks.html", nil)
	} else {
		logError.Println("invalid request: ", r.Method)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
