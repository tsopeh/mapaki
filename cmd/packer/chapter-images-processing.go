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

func processChapters(input []Chapter, options ProcessingOptions) ([]Chapter, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	chapterProcessingInputCh := make(chan ChapterProcessingInput)

	imageProcessingResultCh := processImage(doneCh, processChapter(doneCh, chapterProcessingInputCh))

	go func() {
		defer close(chapterProcessingInputCh)
		for chapterIndex, chapter := range input {
			select {
			case chapterProcessingInputCh <- ChapterProcessingInput{
				chapterIndex: chapterIndex,
				chapter:      chapter,
				options: ProcessingOptions{
					DisableAutoCrop: options.DisableAutoCrop,
					DoublePage:      options.DoublePage,
				},
			}:
			case <-doneCh:
				return
			}
		}
	}()

	chapterMap := make(map[int]map[int][]image.Image)
	for result := range imageProcessingResultCh {
		if result.err != nil {
			return nil, fmt.Errorf(`error occured while processing an image. %w`, result.err)
		} else {
			pageMap, ok := chapterMap[result.chapterIndex]
			if !ok {
				chapterMap[result.chapterIndex] = make(map[int][]image.Image)
				pageMap = chapterMap[result.chapterIndex]
			}
			pageMap[result.pageIndex] = result.images
		}
	}

	output := []Chapter{}
	for chapterIndex := 0; chapterIndex < len(chapterMap); chapterIndex++ {
		pageMap, _ := chapterMap[chapterIndex]
		pages := []image.Image{}
		for pageIndex := 0; pageIndex < len(pageMap); pageIndex++ {
			images, _ := pageMap[pageIndex]
			pages = append(pages, images...)
		}
		output = append(output, Chapter{
			title:  input[chapterIndex].title,
			images: pages,
		})
	}
	return output, nil
}

func processChapter(doneCh <-chan struct{}, inputCh <-chan ChapterProcessingInput) <-chan ImageProcessingInput {
	outputCh := make(chan ImageProcessingInput)
	go func() {
		defer close(outputCh)
		for input := range inputCh {
			for pageIndex, img := range input.chapter.images {
				select {
				case outputCh <- ImageProcessingInput{
					chapterIndex: input.chapterIndex,
					pageIndex:    pageIndex,
					image:        img,
					options:      input.options,
				}:
				case <-doneCh:
					return
				}
			}
		}
	}()
	return outputCh
}

func processImage(doneCh chan struct{}, inputCh <-chan ImageProcessingInput) <-chan ImageProcessingResult {
	outputCh := make(chan ImageProcessingResult)
	go func() {
		defer close(outputCh)
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
			select {
			case outputCh <- ImageProcessingResult{
				chapterIndex: input.chapterIndex,
				pageIndex:    input.pageIndex,
				images:       outputImages,
				err:          nil,
			}:
			case <-doneCh:
				return
			}
		}
	}()
	return outputCh
}
