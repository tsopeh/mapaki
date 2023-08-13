package packer

import (
	"fmt"
	"github.com/leotaku/mobi"
	"github.com/leotaku/mobi/records"
	"github.com/tsopeh/mapaki/cmd/crop"
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

	for chapterIndex, chapter := range mangaChapters {
		pages := []string{}
		for imageIndex, img := range chapter.images {
			croppedImage := img
			if !params.DisableAutoCrop {
				if cropped, err := crop.Crop(img, crop.Limited(img, 0.1)); err != nil {
					return fmt.Errorf(`failed to crop an image in chapter %v, at index %v. %w`, chapterIndex, imageIndex, err)
				} else {
					croppedImage = cropped
				}
			}

			bounds := croppedImage.Bounds()
			isDoublePage := bounds.Dx() >= bounds.Dy()
			if isDoublePage && params.DoublePage != "only-double" {
				leftImage, rightImage, err := crop.SplitVertically(croppedImage)
				if err != nil {
					return fmt.Errorf(`could not split the image at chapter %v, at index %v`, chapterIndex, imageIndex, err)
				}
				switch params.DoublePage {
				case "only-split":
					allImages = append(allImages, rightImage, leftImage)
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++
				case "split-then-double":
					allImages = append(allImages, rightImage, leftImage)
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++

					allImages = append(allImages, croppedImage)
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++
				case "double-then-split":
					allImages = append(allImages, croppedImage)
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++

					allImages = append(allImages, rightImage, leftImage)
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++
					pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
					pageImageIndex++
				default:
					return fmt.Errorf(`unknown double-page flag value %v`, params.DoublePage)
				}
			} else {
				allImages = append(allImages, croppedImage)
				pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
				pageImageIndex++
			}
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
