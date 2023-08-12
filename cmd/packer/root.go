package packer

import (
	"facette.io/natsort"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func PackMangaForKindle(inputPath string) error {
	orderedImages, err := getNaturallyOrderedImagePaths(inputPath)
	if err != nil {
		return err
	}
	log.Println(orderedImages)
	return nil
}

func getNaturallyOrderedImagePaths(dirPath string) ([]string, error) {
	items, _ := os.ReadDir(dirPath)
	images := []string{}
	subDirs := []string{}
	for _, item := range items {
		absPath := path.Join(dirPath, item.Name())
		if item.IsDir() {
			subDirs = append(subDirs, absPath)
		} else {
			file, err := os.Open(absPath)
			if err != nil {
				return nil, err
			}
			buff := make([]byte, 512) // why 512 bytes? see http://golang.org/pkg/net/http/#DetectContentType
			bytesRead, err := file.Read(buff)
			if err != nil && err != io.EOF {
				return nil, err
			}
			// Slice to remove fill-up zero values which cause a wrong content type detection in the next step
			buff = buff[:bytesRead]
			filetype := http.DetectContentType(buff)
			switch filetype {
			case "image/jpeg", "image/jpg", "image/png":
				images = append(images, absPath)
			default:
				fmt.Println("unknown file type uploaded", filetype, "for file", absPath)
			}
		}
	}
	natsort.Sort(images)
	natsort.Sort(subDirs)
	for _, subDir := range subDirs {
		if subImages, err := getNaturallyOrderedImagePaths(subDir); err != nil {
			return nil, err
		} else {
			images = append(images, subImages...)
		}
	}
	return images, nil
}
