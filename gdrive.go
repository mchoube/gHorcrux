package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	"google.golang.org/api/googleapi"
)

type GDrive struct {
	clientSecret string
	cacheFile    string
	config       *oauth2.Config
	ctx          context.Context
	client       *http.Client
	w            http.ResponseWriter
	r            *http.Request
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
func (gd *GDrive) tokenFromFile(file string) (*oauth2.Token, error) {
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
	refreshToken := false

	tok, err := gd.tokenFromFile(gd.cacheFile)
	if err != nil {
		refreshToken = true
	}

	logInfo.Println("token: ", tok)

	if !tok.Valid() {
		logInfo.Println("token is expired.")
		refreshToken = true
	}

	if refreshToken {
		authURL := gd.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

func (gd *GDrive) HasToken() bool {
	tok, err := gd.tokenFromFile(gd.cacheFile)
	if err != nil {
		return false
	}
	logInfo.Println("token: ", tok)

	if !tok.Valid() {
		logInfo.Println("token is expired.")
		return false
	}

	if gd.client == nil {
		gd.ctx = context.Background()
		gd.client = gd.config.Client(gd.ctx, tok)
	}

	return true
}

// saveToken uses a file path to create a file and store the
// token in it.
func (gd *GDrive) saveToken(file string, token *oauth2.Token) error {
	logInfo.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		logError.Printf("Unable to cache oauth token: %v", err)
		return err
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	gd.ctx = context.Background()
	gd.client = gd.config.Client(gd.ctx, token)
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

	gd.saveToken(gd.cacheFile, tok)

	return nil
}

func (gd *GDrive) Unlink() {
}

func (gd *GDrive) RefreshToken() {
}

func (gd *GDrive) List() []FileList {
	srv, err := drive.New(gd.client)
	if err != nil {
		logError.Printf("Unable to retrieve drive Client %v", err)
		return nil
	}

	r, err := srv.Files.List().Do()
	if err != nil {
		logError.Printf("Unable to retrieve files. err: %v", err)
		return nil
	}

	if len(r.Files) <= 0 {
		logInfo.Print("No files found.")
		return nil
	}

	files := make([]FileList, len(r.Files))

	logInfo.Println("Files:")
	for i, f := range r.Files {
		logInfo.Printf("%v", f)
		logInfo.Printf("%s (%s)\n", f.Name, f.Id)
		fl := FileList{
			Icon:         "",
			Name:         f.Name,
			Extension:    f.FileExtension,
			CreatedTime:  f.CreatedTime,
			ModifiedTime: f.ModifiedTime,
		}
		files[i] = fl
	}

	logInfo.Printf("%v", files)

	return files
}

func (gd *GDrive) UploadFile(fname string, r io.Reader) error {
	srv, err := drive.New(gd.client)
	if err != nil {
		logError.Printf("Unable to retrieve drive Client %v", err)
		return err
	}

	mo := googleapi.ChunkSize(googleapi.DefaultUploadChunkSize)

	fc := srv.Files.Create(&drive.File{Name: fname})
	fc = fc.Media(r, mo)

	logInfo.Printf("%v", fc)
	f, err := fc.Do()
	if err != nil {
		logError.Printf("Unable to upload file: %v.", err)
		return err
	}
	logInfo.Printf("%s (%s)\n", f.Name, f.Id)

	return nil
}

func (gd *GDrive) UploadFolder() {
}

func (gd *GDrive) Delete() {
}
