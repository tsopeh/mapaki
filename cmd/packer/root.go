package packer

import (
	"fmt"
	"github.com/leotaku/mobi"
	"github.com/leotaku/mobi/records"
	"image"
	"os"
	"path"
	"time"
)

type PackForKindleParams struct {
	RootDir        string
	AutoCrop       bool
	RightToLeft    bool
	DoublePage     string
	Title          string
	OutputFilePath string
}

func PackMangaForKindle(params PackForKindleParams) error {

	mangaChapters, err := discoverMangaChapters(params.RootDir)
	if err != nil {
		return err
	}
	if len(mangaChapters) == 0 {
		return fmt.Errorf("no manga chapters were found")
	}

	bookChapters := []mobi.Chapter{}
	allImages := []image.Image{}
	pageImageIndex := 1

	for _, chapter := range mangaChapters {
		pages := []string{}
		for _, img := range chapter.images {
			allImages = append(allImages, img)
			pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
			pageImageIndex++
		}
		bookChapters = append(bookChapters, mobi.Chapter{
			Title:  chapter.title,
			Chunks: mobi.Chunks(pages...),
		})
	}

	mangaDirName := path.Base(params.RootDir)
	mangaTitle := params.Title
	if mangaTitle == "" {
		mangaTitle = mangaDirName
	}

	book := mobi.Book{
		Title:       mangaTitle, // TODO: Use title from arguments or fallback to root dir name.
		CSSFlows:    []string{basePageCSS},
		Chapters:    bookChapters,
		Images:      allImages,
		CoverImage:  allImages[0],
		FixedLayout: true,
		RightToLeft: params.RightToLeft,
		CreatedDate: time.Unix(0, 0),
		UniqueID:    uint32(time.Unix(0, 0).UnixMilli()),
	}

	outputFilePath := params.OutputFilePath
	if outputFilePath == "" {
		outputFilePath = path.Join(params.RootDir, "../", mangaDirName+".azw3")
	}
	writer, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf(`could not create output file: "%v" %w`, outputFilePath, err)
	}
	err = book.Realize().Write(writer)
	if err != nil {
		return fmt.Errorf(`could not write output file: "%v" %w`, outputFilePath, err)
	}

	return nil
}
