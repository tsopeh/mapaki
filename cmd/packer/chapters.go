package packer

import (
	"facette.io/natsort"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

type Chapter struct {
	title  string
	images []image.Image
}

type MangaDirContent struct {
	subDirPaths []string
	imagePaths  []string
}

func discoverMangaChapters(rootDirPath string) ([]Chapter, error) {
	rootContent, err := getMangaDirContent(rootDirPath)
	if err != nil {
		return nil, err
	}
	hasSubDirs := len(rootContent.subDirPaths) > 0
	hasRootImages := len(rootContent.imagePaths) > 0
	isRootChapter := !hasSubDirs && hasRootImages
	if hasSubDirs {
		chapters := []Chapter{}
		for _, subDirPath := range rootContent.subDirPaths {
			subContent, err := getMangaDirContent(subDirPath)
			if err != nil {
				return nil, err
			}
			isVolume := len(subContent.subDirPaths) > 0
			hasImages := len(subContent.imagePaths) > 0
			isChapter := !isVolume && hasImages
			if isVolume {
				volumeChapters := []Chapter{}
				for _, chapterPath := range subContent.subDirPaths {
					chapterContent, err := getMangaDirContent(chapterPath)
					if err != nil {
						return nil, err
					}
					hasSubDirs := len(chapterContent.subDirPaths) > 0
					if hasSubDirs {
						return nil, fmt.Errorf(`detected subdirectories within a chapter directory "%v"`, chapterPath)
					}
					chapter, err := createChapter(chapterPath)
					if err != nil {
						return nil, err
					}
					volumeChapters = append(volumeChapters, chapter)
				}
				if hasImages {
					for _, imagePath := range subContent.imagePaths {
						img, err := readImageFromPath(imagePath)
						if err != nil {
							return nil, fmt.Errorf(`cannot load an image on path "%v" %w`, imagePath, err)
						}
						volumeChapters[0].images = append([]image.Image{img}, volumeChapters[0].images...)
					}
				}
				chapters = append(chapters, volumeChapters...)
			} else if isChapter {
				chapter, err := createChapter(subDirPath)
				if err != nil {
					return nil, err
				}
				chapters = append(chapters, chapter)
			} else {
				return nil, fmt.Errorf(`unknown manga sub-directory structure at path "%v"`, subDirPath)
			}
		}
		if hasRootImages {
			for _, imagePath := range rootContent.imagePaths {
				img, err := readImageFromPath(imagePath)
				if err != nil {
					return nil, fmt.Errorf(`cannot load an image on path "%v" %w`, imagePath, err)
				}
				chapters[0].images = append([]image.Image{img}, chapters[0].images...)
			}
		}
		return chapters, nil
	} else if isRootChapter {
		chapter, err := createChapter(rootDirPath)
		if err != nil {
			return nil, err
		}
		return []Chapter{chapter}, err
	} else {
		return nil, fmt.Errorf(`unknown manga directory structure at path "%v"`, rootDirPath)
	}
}

func getMangaDirContent(dirPath string) (MangaDirContent, error) {
	items, _ := os.ReadDir(dirPath)
	imagePaths := []string{}
	subDirPaths := []string{}
	for _, item := range items {
		itemPath := path.Join(dirPath, item.Name())
		if item.IsDir() {
			subDirPaths = append(subDirPaths, itemPath)
		} else {
			file, err := os.Open(itemPath)
			if err != nil {
				return MangaDirContent{}, err
			}
			defer file.Close()
			buff := make([]byte, 512) // why 512 bytes? see http://golang.org/pkg/net/http/#DetectContentType
			bytesRead, err := file.Read(buff)
			if err != nil && err != io.EOF {
				return MangaDirContent{}, err
			}
			// Slice to remove fill-up zero values which cause a wrong content type detection in the next step
			buff = buff[:bytesRead]
			filetype := http.DetectContentType(buff)
			switch filetype {
			case "image/jpeg", "image/jpg", "image/png":
				imagePaths = append(imagePaths, itemPath)
			default:
				log.Println(fmt.Sprintf(`file type is not of an image "%v", for file "%v"`, filetype, itemPath))
			}
		}
	}
	natsort.Sort(imagePaths)  // in-place sort
	natsort.Sort(subDirPaths) // in-place sort
	return MangaDirContent{
		subDirPaths: subDirPaths,
		imagePaths:  imagePaths,
	}, nil
}

func createChapter(chapterDirPath string) (Chapter, error) {
	items, _ := os.ReadDir(chapterDirPath)
	loadedImages := []image.Image{}
	for _, imageItem := range items {
		imagePath := path.Join(chapterDirPath, imageItem.Name())
		img, err := readImageFromPath(imagePath)
		if err != nil {
			return Chapter{}, fmt.Errorf(`cannot load an image on path "%v" %w`, imagePath, err)
		}
		loadedImages = append(loadedImages, img)
	}
	return Chapter{
		title:  path.Base(chapterDirPath),
		images: loadedImages,
	}, nil
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
