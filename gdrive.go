package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type GDrive struct {
	clientSecret string
	cacheFile    string
	config       *oauth2.Config
	ctx          context.Context
	client       *http.Client
}

func NewGDrive() *GDrive {
	var err error

	usr, err := user.Current()
	if err != nil {
		return nil
	}

	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	tokenCacheFile := filepath.Join(tokenCacheDir,
		url.QueryEscape("drive-go-quickstart.json"))

	b, err := ioutil.ReadFile("gdrive_client_secret.json")
	if err != nil {
		logError.Printf("Unable to read client secret file: %v", err)
		return nil
	}

	cfg, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		logError.Printf("Unable to parse client secret file to config: %v", err)
		return nil
	}

	return &GDrive{
		clientSecret: "gdrive_client_secret.json",
		cacheFile:    tokenCacheFile,
		config:       cfg,
	}
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

func (gd *GDrive) Link(w http.ResponseWriter, r *http.Request) {
	tok, err := tokenFromFile(gd.cacheFile)
	if err != nil {
		authURL := gd.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		http.Redirect(w, r, authURL, http.StatusFound)
	} else {
		logInfo.Println("token: ", tok)
	}
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) error {
	logInfo.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		logError.Printf("Unable to cache oauth token: %v", err)
		return err
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}

func (gd *GDrive) SaveToken(r *http.Request) error {
	logInfo.Println("got: ", r.URL.Query())

	values := r.URL.Query()
	if _, ok := values["code"]; !ok {
		err := fmt.Errorf("code is not present")
		logError.Println(err.Error())
		return err
	}

	code := r.URL.Query()["code"][0]
	logInfo.Println("got code: ", code)

	tok, err := gd.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		logError.Printf("Unable to retrieve token from web %v", err)
		return err
	}

	saveToken(gd.cacheFile, tok)

	gd.ctx = context.Background()
	gd.client = gd.config.Client(gd.ctx, tok)

	return nil
}

func (gd *GDrive) Unlink() {
}

func (gd *GDrive) RefreshToken() {
}

func (gd *GDrive) List() {
	srv, err := drive.New(gd.client)
	if err != nil {
		logError.Printf("Unable to retrieve drive Client %v", err)
		return
	}

	r, err := srv.Files.List().Do()
	if err != nil {
		logError.Printf("Unable to retrieve files.", err)
		return
	}

	logInfo.Println("Files:")
	if len(r.Files) > 0 {
		for _, i := range r.Files {
			logInfo.Printf("%s (%s)\n", i.Name, i.Id)
		}
	} else {
		logInfo.Print("No files found.")
	}
}

func (gd *GDrive) UploadFile() {
}

func (gd *GDrive) UploadFiles() {
}

func (gd *GDrive) UploadFolder() {
}

func (gd *GDrive) Delete() {
}
