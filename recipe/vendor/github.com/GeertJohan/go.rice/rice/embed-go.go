package main

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const boxFilename = "rice-box.go"

func operationEmbedGo(pkg *build.Package) {

	boxMap := findBoxes(pkg)

	// notify user when no calls to rice.FindBox are made (is this an error and therefore os.Exit(1) ?
	if len(boxMap) == 0 {
		fmt.Println("no calls to rice.FindBox() found")
		return
	}

	verbosef("\n")
	var boxes []*boxDataType

	for boxname := range boxMap {
		// find path and filename for this box
		boxPath := filepath.Join(pkg.Dir, boxname)

		// Check to see if the path for the box is a symbolic link.  If so, simply
		// box what the symbolic link points to.  Note: the filepath.Walk function
		// will NOT follow any nested symbolic links.  This only handles the case
		// where the root of the box is a symbolic link.
		symPath, serr := os.Readlink(boxPath)
		if serr == nil {
			boxPath = symPath
		}

		// verbose info
		verbosef("embedding box '%s' to '%s'\n", boxname, boxFilename)

		// read box metadata
		boxInfo, ierr := os.Stat(boxPath)
		if ierr != nil {
			fmt.Printf("Error: unable to access box at %s\n", boxPath)
			os.Exit(1)
		}

		// create box datastructure (used by template)
		box := &boxDataType{
			BoxName: boxname,
			UnixNow: boxInfo.ModTime().Unix(),
			Files:   make([]*fileDataType, 0),
			Dirs:    make(map[string]*dirDataType),
		}

		if !boxInfo.IsDir() {
			fmt.Printf("Error: Box %s must point to a directory but points to %s instead\n",
				boxname, boxPath)
			os.Exit(1)
		}

		// fill box datastructure with file data
		filepath.Walk(boxPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("error walking box: %s\n", err)
				os.Exit(1)
			}

			filename := strings.TrimPrefix(path, boxPath)
			filename = strings.Replace(filename, "\\", "/", -1)
			filename = strings.TrimPrefix(filename, "/")
			if info.IsDir() {
				dirData := &dirDataType{
					Identifier: "dir" + nextIdentifier(),
					FileName:   filename,
					ModTime:    info.ModTime().Unix(),
					ChildFiles: make([]*fileDataType, 0),
					ChildDirs:  make([]*dirDataType, 0),
				}
				verbosef("\tincludes dir: '%s'\n", dirData.FileName)
				box.Dirs[dirData.FileName] = dirData

				// add tree entry (skip for root, it'll create a recursion)
				if dirData.FileName != "" {
					pathParts := strings.Split(dirData.FileName, "/")
					parentDir := box.Dirs[strings.Join(pathParts[:len(pathParts)-1], "/")]
					parentDir.ChildDirs = append(parentDir.ChildDirs, dirData)
				}
			} else {
				fileData := &fileDataType{
					Identifier: "file" + nextIdentifier(),
					FileName:   filename,
					ModTime:    info.ModTime().Unix(),
				}
				verbosef("\tincludes file: '%s'\n", fileData.FileName)
				fileData.Content, err = ioutil.ReadFile(path)
				if err != nil {
					fmt.Printf("error reading file content while walking box: %s\n", err)
					os.Exit(1)
				}
				box.Files = append(box.Files, fileData)

				// add tree entry
				pathParts := strings.Split(fileData.FileName, "/")
				parentDir := box.Dirs[strings.Join(pathParts[:len(pathParts)-1], "/")]
				if parentDir == nil {
					fmt.Printf("Error: parent of %s is not within the box\n", path)
					os.Exit(1)
				}
				parentDir.ChildFiles = append(parentDir.ChildFiles, fileData)
			}
			return nil
		})
		boxes = append(boxes, box)

	}

	embedSourceUnformated := bytes.NewBuffer(make([]byte, 0))

	// execute template to buffer
	err := tmplEmbeddedBox.Execute(
		embedSourceUnformated,
		embedFileDataType{pkg.Name, boxes},
	)
	if err != nil {
		log.Printf("error writing embedded box to file (template execute): %s\n", err)
		os.Exit(1)
	}

	// format the source code
	embedSource, err := format.Source(embedSourceUnformated.Bytes())
	if err != nil {
		log.Printf("error formatting embedSource: %s\n", err)
		os.Exit(1)
	}

	// create go file for box
	boxFile, err := os.Create(filepath.Join(pkg.Dir, boxFilename))
	if err != nil {
		log.Printf("error creating embedded box file: %s\n", err)
		os.Exit(1)
	}
	defer boxFile.Close()

	// write source to file
	_, err = io.Copy(boxFile, bytes.NewBuffer(embedSource))
	if err != nil {
		log.Printf("error writing embedSource to file: %s\n", err)
		os.Exit(1)
	}

}
