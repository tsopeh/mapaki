package packer

import (
	"facette.io/natsort"
	"fmt"
	"github.com/leotaku/mobi"
	"github.com/leotaku/mobi/records"
	"image"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

func PackMangaForKindle(rootDir string) error {
	orderedImagePaths, err := getNaturallyOrderedImagePaths(rootDir)
	if err != nil {
		return err
	}
	if len(orderedImagePaths) == 0 {
		return fmt.Errorf("no manga pages were found")
	}

	pages := make([]string, 0)
	chapters := make([]mobi.Chapter, 0)
	images := make([]image.Image, 0)
	pageImageIndex := 1

	for _, imagePath := range orderedImagePaths {
		img, err := readImageFromFilePath(imagePath)
		if err != nil {
			return fmt.Errorf(`cannot load image on path "%w"`, err)
		}
		images = append(images, img)
		pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
		pageImageIndex++
	}

	chapters = append(chapters, mobi.Chapter{
		Title:  "Chapter title TODO",
		Chunks: mobi.Chunks(pages...),
	})

	book := mobi.Book{
		Title:       "Manga title TODO", // TODO: Use title from arguments or fallback to root dir name.
		CSSFlows:    []string{basePageCSS},
		Chapters:    chapters,
		Images:      images,
		CoverImage:  images[0],
		FixedLayout: true,
		RightToLeft: true,
		CreatedDate: time.Unix(0, 0),
		UniqueID:    uint32(time.Unix(0, 0).UnixMilli()),
	}

	outFileAbsPath := path.Join(rootDir, "out.azw3")
	writer, err := os.Create(outFileAbsPath)
	if err != nil {
		return fmt.Errorf(`could not create output file: "%v" %w`, outFileAbsPath, err)
	}
	err = book.Realize().Write(writer)
	if err != nil {
		return fmt.Errorf(`could not write output file: "%v" %w`, outFileAbsPath, err)
	}

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
	natsort.Sort(images)  // in-place sort
	natsort.Sort(subDirs) // in-place sort
	for _, subDir := range subDirs {
		if subImages, err := getNaturallyOrderedImagePaths(subDir); err != nil {
			return nil, err
		} else {
			images = append(images, subImages...)
		}
	}
	return images, nil
}

func readImageFromFilePath(imgPath string) (image.Image, error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
