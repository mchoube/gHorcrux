# gHorcrux

Single pane to your personal cloud storage. User can manage various personal cloud storage providers like Google Drive, Dropbox and Flickr using a single web UI. Web UI supports drag and drop capabilities.

Developed by [Mehul Choube](https://github.com/mchoube) during the 2016 Gophergala global hackathon.

##Usage

### Getting Started
Install gHorcrux into your $GOPATH with `go get`
```Go
go get -u github.com/gophergala2016/gHorcrux
```
Navigate to gHorcrux directory and run  
```Go
go run *.go 
```
then open browser and type loclahost:9999 to use the web appliacation.

### Extensible
By just implementing following interface new storage provides can be added:
```Go
type horcrux interface {
	Link(http.ResponseWriter, *http.Request)
	Unlink()
	HasToken() bool
	SaveToken(*http.Request) error
	RefreshToken()
	List() []FileList
	UploadFile(string, io.Reader) error
	UploadFolder()
	Delete()
}
```
