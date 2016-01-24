package main

import "net/http"

type horcrux interface {
	Link(http.ResponseWriter, *http.Request)
	Unlink()
	SaveToken(*http.Request) error
	RefreshToken()
	List()
	UploadFile()
	UploadFiles()
	UploadFolder()
	Delete()
}
