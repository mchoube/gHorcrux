package main

import (
	"io"
	"net/http"
)

type FileList struct {
	Icon         string
	Name         string
	Extension    string
	CreatedTime  string
	ModifiedTime string
}

type horcrux interface {
	Link(http.ResponseWriter, *http.Request)
	Unlink()
	HasToken() bool
	SaveToken(*http.Request) error
	RefreshToken()
	List() []FileList
	UploadFile(string, io.Reader) error
	UploadFiles()
	UploadFolder()
	Delete()
}
