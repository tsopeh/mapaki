package packer

import (
	"facette.io/natsort"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/leotaku/mobi"
	"github.com/leotaku/mobi/records"
	"image"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"hash/fnv"
)

type PackForKindleParams struct {
	RootDir         string
	DisableAutoCrop bool
	LeftToRight     bool
	DoublePage      string
	Title           string
	Author          string
	OutputFilePath  string
	CoresCount      int
}

type ProcessingOptions struct {
	DisableAutoCrop bool
	DoublePage      string
	CoresCount      int
}

type ProcessedPage struct {
	imagePath string
	images    []image.Image
}

func PackMangaForKindle(params PackForKindleParams) error {
	commCh := CommCh{
		done: make(chan struct{}),
		err:  make(chan error),
	}
	defer close(commCh.done)
	defer close(commCh.err)

	processedPageCh := processFile(
		commCh,
		ProcessingOptions{
			DisableAutoCrop: params.DisableAutoCrop,
			DoublePage:      params.DoublePage,
			CoresCount:      params.CoresCount,
		},
		discoverFiles(commCh, params.RootDir),
	)

	go func() {
		for err := range commCh.err {
			if err != nil {
				log.Println(fmt.Errorf(`caught error: %w`, err))
				commCh.done <- struct{}{}
			}
		}
	}()

	processed := []ProcessedPage{}
	for page := range processedPageCh {
		processed = append(processed, page)
	}

	sort.SliceStable(processed, func(i, j int) bool {
		a := processed[i].imagePath
		b := processed[j].imagePath
		aDir, _ := filepath.Split(a)
		bDir, _ := filepath.Split(b)
		areInSameDir := aDir == bDir
		if areInSameDir {
			return natsort.Compare(a, b)
		} else {
			if strings.Contains(bDir, aDir) {
				return true
			} else if strings.Contains(aDir, bDir) {
				return false
			} else {
				return natsort.Compare(aDir, bDir)
			}
		}
	})
	if len(processed) == 0 {
		return fmt.Errorf(`nothing to output, no manga pages were found`)
	}

	allImages := []image.Image{}
	bookChapters := []mobi.Chapter{}
	chapterBuffer := []string{}
	pageImageIndex := 1
	prevPageChapterName := getChapterNameForImagePath(processed[0].imagePath)
	for _, page := range processed {
		currPageChapterName := getChapterNameForImagePath(page.imagePath)
		isSameChapterAsPrevPage := prevPageChapterName == currPageChapterName
		if !isSameChapterAsPrevPage {
			bookChapters = append(bookChapters, mobi.Chapter{
				Title:  prevPageChapterName,
				Chunks: mobi.Chunks(chapterBuffer...),
			})
			chapterBuffer = []string{}
			prevPageChapterName = currPageChapterName
		}
		for _, img := range page.images {
			allImages = append(allImages, img)
			chapterBuffer = append(chapterBuffer, templateToString(imagePageTemplate, records.To32(pageImageIndex)))
			pageImageIndex++
		}
	}
	bookChapters = append(bookChapters, mobi.Chapter{
		Title:  prevPageChapterName,
		Chunks: mobi.Chunks(chapterBuffer...),
	})

	mangaDirName := path.Base(params.RootDir)
	mangaTitle := params.Title
	if mangaTitle == "" {
		mangaTitle = mangaDirName
	}

	book := mobi.Book{
		Title:       mangaTitle,
		Authors:     []string{params.Author},
		CSSFlows:    []string{basePageCSS},
		Chapters:    bookChapters,
		Images:      allImages,
		CoverImage:  allImages[0],
		FixedLayout: true,
		RightToLeft: !params.LeftToRight,
		CreatedDate: time.Unix(0, 0),
		UniqueID:    getUniqueId(mangaTitle),
	}

	outputFilePath := params.OutputFilePath
	if outputFilePath == "" {
		outputFilePath = path.Join(params.RootDir, "../", mangaDirName+".azw3")
	}
	writer, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf(`could not create output file: "%v" %w`, outputFilePath, err)
	}
	bookExportPb := pb.ProgressBarTemplate(`Exporting manga {{(cycle . "↖" "↗" "↘" "↙" )}}`)
	bookExportPb.Start(0)
	err = book.Realize().Write(writer)
	if err != nil {
		return fmt.Errorf(`could not write output file: "%v" %w`, outputFilePath, err)
	}

	return nil
}

func getChapterNameForImagePath(imagePath string) string {
	return filepath.Base(filepath.Dir(imagePath))
}

func getUniqueId(title string) uint32 {
	hash := fnv.New32()
	hash.Write([]byte(title))
	return hash.Sum32()
}
