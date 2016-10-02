package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/GeertJohan/go.rice"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	conf := rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateEmbedded, rice.LocateAppended, rice.LocateFS},
	}
	box, err := conf.FindBox("example-files")
	if err != nil {
		log.Fatalf("error opening rice.Box: %s\n", err)
	}
	// spew.Dump(box)

	contentString, err := box.String("file.txt")
	if err != nil {
		log.Fatalf("could not read file contents as string: %s\n", err)
	}
	log.Printf("Read some file contents as string:\n%s\n", contentString)

	contentBytes, err := box.Bytes("file.txt")
	if err != nil {
		log.Fatalf("could not read file contents as byteSlice: %s\n", err)
	}
	log.Printf("Read some file contents as byteSlice:\n%s\n", hex.Dump(contentBytes))

	file, err := box.Open("file.txt")
	if err != nil {
		log.Fatalf("could not open file: %s\n", err)
	}
	spew.Dump(file)

	// find/create a rice.Box
	templateBox, err := rice.FindBox("example-templates")
	if err != nil {
		log.Fatal(err)
	}
	// get file contents as string
	templateString, err := templateBox.String("message.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	// parse and execute the template
	tmplMessage, err := template.New("message").Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}
	tmplMessage.Execute(os.Stdout, map[string]string{"Message": "Hello, world!"})

	http.Handle("/", http.FileServer(box.HTTPBox()))
	go func() {
		fmt.Println("Serving files on :8080, press ctrl-C to exit")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("error serving files: %v", err)
		}
	}()
	select {}
}
