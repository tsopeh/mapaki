package packer

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/tsopeh/mapaki/cmd/crop"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type DiscoveryInput struct {
	rootDir string
}

type CommCh struct {
	done chan struct{}
	err  chan error
}

type ProcessableFileInfo struct {
	filePath string
}

func discoverFiles(commCh CommCh, rootDir string) <-chan ProcessableFileInfo {
	outCh := make(chan ProcessableFileInfo, 100)
	go func() {
		defer close(outCh)
		err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf(`an error occurred in filepath.Walk callback: %w`, err)
			}

			if info.IsDir() {
				return nil
			}

			select {
			case outCh <- ProcessableFileInfo{
				filePath: path,
			}:
			case <-commCh.done:
				return nil
			}

			return nil
		})
		if err != nil {
			commCh.err <- fmt.Errorf(`an error occurred during file walk: %w`, err)
		}
	}()

	return outCh
}

func processFile(commCh CommCh, options ProcessingOptions, inCh <-chan ProcessableFileInfo) <-chan ProcessedPage {
	outCh := make(chan ProcessedPage, 100)

	var processingPb = pb.New(0)
	processingPb.Set("prefix", "Processing images")
	processingPb.SetMaxWidth(80)
	processingPb.Start()

	wg := sync.WaitGroup{}
	for i := 0; i < options.CoresCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for info := range inCh {
				processingPb.AddTotal(1)
				originalImg, err := readImageFromPath(info.filePath)
				if err != nil {
					// Just print the message, do not stop processing.
					log.Println(fmt.Errorf(`skip file: "%v": not an image or corrupted: %w`, info.filePath, err))
					processingPb.AddTotal(-1)
					continue
				}

				croppedImage := originalImg
				if !options.DisableAutoCrop {
					if cropped, err := crop.Crop(originalImg, crop.Limited(originalImg, 0.1)); err != nil {
						commCh.err <- fmt.Errorf(`failed to crop an image "%v": %w`, info.filePath, err)
						return
					} else {
						croppedImage = cropped
					}
				}

				outputImages := []image.Image{}
				bounds := croppedImage.Bounds()
				isDoublePage := bounds.Dx() >= bounds.Dy()
				if isDoublePage && options.DoublePage != "only-double" {
					leftImage, rightImage, err := crop.SplitVertically(croppedImage)
					if err != nil {
						commCh.err <- fmt.Errorf(`could not split and image "%v": %w`, info.filePath, err)
						return
					}
					switch options.DoublePage {
					case "only-split":
						outputImages = append(outputImages, rightImage, leftImage)
					case "split-then-double":
						outputImages = append(outputImages, rightImage, leftImage, croppedImage)
					case "double-then-split":
						outputImages = append(outputImages, croppedImage, rightImage, leftImage)
					default:
						if err != nil {
							commCh.err <- fmt.Errorf(`unknown double-page flag value "%v"`, options.DoublePage)
							return
						}
						return
					}
				} else {
					outputImages = append(outputImages, croppedImage)
				}

				select {
				case outCh <- ProcessedPage{
					imagePath: info.filePath,
					images:    outputImages,
				}:
					{
						processingPb.Add(1)
					}
				case <-commCh.done:
					return
				}
			}
		}()
	}

	go func() {
		defer close(outCh)
		defer processingPb.Finish()
		wg.Wait()
	}()

	return outCh
}

func readImageFromPath(imgPath string) (image.Image, error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
