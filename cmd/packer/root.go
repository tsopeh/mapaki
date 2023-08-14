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
	RootDir         string
	DisableAutoCrop bool
	LeftToRight     bool
	DoublePage      string
	Title           string
	OutputFilePath  string
}

func PackMangaForKindle(params PackForKindleParams) error {

	discoveredChapters, err := discoverMangaChapters(params.RootDir)
	if err != nil {
		return err
	}
	if len(discoveredChapters) == 0 {
		return fmt.Errorf("no manga chapters were found")
	}

	processedChapters, err := processChapters(discoveredChapters, ProcessingOptions{
		DisableAutoCrop: params.DisableAutoCrop,
		DoublePage:      params.DoublePage,
	})

	bookChapters := []mobi.Chapter{}
	allImages := []image.Image{}
	pageImageIndex := 1
	for _, chapter := range processedChapters {
		pages := []string{}
		if len(chapter.pages) > 0 {
			for _, img := range chapter.pages {
				allImages = append(allImages, img)
				pages = append(pages, templateToString(imagePageTemplate, records.To32(pageImageIndex)))
				pageImageIndex++
			}
		} else {
			pages = append(pages, templateToString(emptyPageTemplate, nil))
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
		Title:       mangaTitle,
		CSSFlows:    []string{basePageCSS},
		Chapters:    bookChapters,
		Images:      allImages,
		CoverImage:  allImages[0],
		FixedLayout: true,
		RightToLeft: !params.LeftToRight,
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
