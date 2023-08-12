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

func PackMangaForKindle(rootDir string) error {
	mangaChapters, err := discoverMangaChapters(rootDir)
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

	book := mobi.Book{
		Title:       "Manga title TODO", // TODO: Use title from arguments or fallback to root dir name.
		CSSFlows:    []string{basePageCSS},
		Chapters:    bookChapters,
		Images:      allImages,
		CoverImage:  allImages[0],
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
