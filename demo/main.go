package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	s := http.FileServer(http.Dir(dir))
	if err := http.ListenAndServe(":8080", s); err != nil {
		log.Println(err)
	}
}
