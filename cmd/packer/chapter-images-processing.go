package packer

import (
	"fmt"
	"github.com/tsopeh/mapaki/cmd/crop"
	"image"
)

type ProcessingOptions struct {
	DisableAutoCrop bool
	DoublePage      string
}

type ChapterProcessingInput struct {
	chapterIndex int
	chapter      Chapter
	options      ProcessingOptions
}

type ImageProcessingInput struct {
	chapterIndex int
	pageIndex    int
	image        image.Image
	options      ProcessingOptions
}

type ImageProcessingResult struct {
	chapterIndex int
	pageIndex    int
	images       []image.Image
	err          error
}

func processChapters(chapters []Chapter, options ProcessingOptions) ([]Chapter, error) {
	chapterProcessingInputCh := make(chan ChapterProcessingInput)
	imageProcessingInputCh := make(chan ImageProcessingInput)
	imageProcessingResultCh := make(chan ImageProcessingResult)

	//for i := 0; i < 2; i++ {
	go processChapter(chapterProcessingInputCh, imageProcessingInputCh)
	//}
	//for i := 0; i < 5; i++ {
	go processImage(imageProcessingInputCh, imageProcessingResultCh)
	//}

	for chapterIndex, chapter := range chapters {
		chapterProcessingInputCh <- ChapterProcessingInput{
			chapterIndex: chapterIndex,
			chapter:      chapter,
			options: ProcessingOptions{
				DisableAutoCrop: options.DisableAutoCrop,
				DoublePage:      options.DoublePage,
			},
		}
	}

	return nil, nil
}

func processChapter(inputCh <-chan ChapterProcessingInput, outputCh chan<- ImageProcessingInput) {
	for input := range inputCh {
		for pageIndex, img := range input.chapter.images {
			outputCh <- ImageProcessingInput{
				chapterIndex: input.chapterIndex,
				pageIndex:    pageIndex,
				image:        img,
				options:      input.options,
			}
		}
	}
}

func processImage(inputCh <-chan ImageProcessingInput, outputCh chan<- ImageProcessingResult) {
	for input := range inputCh {
		croppedImage := input.image
		if !input.options.DisableAutoCrop {
			if cropped, err := crop.Crop(input.image, crop.Limited(input.image, 0.1)); err != nil {
				outputCh <- ImageProcessingResult{
					err: fmt.Errorf(`failed to crop an image in chapter %v, at index %v. %w`, input.chapterIndex, input.pageIndex, err),
				}
				return
			} else {
				croppedImage = cropped
			}
		}

		outputImages := []image.Image{}
		bounds := croppedImage.Bounds()
		isDoublePage := bounds.Dx() >= bounds.Dy()
		if isDoublePage && input.options.DoublePage != "only-double" {
			leftImage, rightImage, err := crop.SplitVertically(croppedImage)
			if err != nil {
				outputCh <- ImageProcessingResult{
					err: fmt.Errorf(`could not split the image at chapter %v, at index %v`, input.chapterIndex, input.pageIndex, err),
				}
				return
			}
			switch input.options.DoublePage {
			case "only-split":
				outputImages = append(outputImages, rightImage, leftImage)
			case "split-then-double":
				outputImages = append(outputImages, rightImage, leftImage, croppedImage)
			case "double-then-split":
				outputImages = append(outputImages, croppedImage, rightImage, leftImage)
			default:
				outputCh <- ImageProcessingResult{
					err: fmt.Errorf(`unknown double-page flag value %v`, input.options.DoublePage),
				}
				return
			}
		} else {
			outputImages = append(outputImages, croppedImage)
		}
		outputCh <- ImageProcessingResult{
			chapterIndex: input.chapterIndex,
			pageIndex:    input.pageIndex,
			images:       outputImages,
			err:          nil,
		}
	}
}
