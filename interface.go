package main

import (
	"io"
	"net/http"
)

type horcrux interface {
	Link(http.ResponseWriter, *http.Request)
	Unlink()
	HasToken() bool
	SaveToken(*http.Request) error
	RefreshToken()
	List()
	UploadFile(string, io.Reader) error
	UploadFiles()
	UploadFolder()
	Delete()
}
