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
	title string
	pages []image.Image
}

type MangaDirContent struct {
	subDirs []os.FileInfo
	images  []image.Image
}

func discoverMangaChapters(rootDirPath string) ([]Chapter, error) {
	rootContent, err := getMangaDirContent(rootDirPath)
	if err != nil {
		return nil, err
	}
	hasSubDirs := len(rootContent.subDirs) > 0
	hasRootImages := len(rootContent.images) > 0
	isRootChapter := !hasSubDirs && hasRootImages
	if hasSubDirs {
		chapters := []Chapter{}
		for _, subDir := range rootContent.subDirs {
			subDirPath := path.Join(rootDirPath, subDir.Name())
			subContent, err := getMangaDirContent(subDirPath)
			if err != nil {
				return nil, err
			}
			isVolume := len(subContent.subDirs) > 0
			hasImages := len(subContent.images) > 0
			isChapter := !isVolume && hasImages
			if isVolume {
				volumeChapters := []Chapter{}
				for _, chapterDir := range subContent.subDirs {
					chapterPath := path.Join(subDirPath, chapterDir.Name())
					chapterContent, err := getMangaDirContent(chapterPath)
					if err != nil {
						return nil, err
					}
					hasSubDirs := len(chapterContent.subDirs) > 0
					if hasSubDirs {
						return nil, fmt.Errorf(`detected subdirectories within a chapter directory "%v"`, chapterPath)
					}
					chapter := Chapter{
						title: chapterDir.Name(),
						pages: chapterContent.images,
					}
					volumeChapters = append(volumeChapters, chapter)
				}
				if hasImages {
					for _, img := range subContent.images {
						volumeChapters[0].pages = append([]image.Image{img}, volumeChapters[0].pages...)
					}
				}
				chapters = append(chapters, volumeChapters...)
			} else if isChapter {
				chapter := Chapter{
					title: subDir.Name(),
					pages: subContent.images,
				}
				chapters = append(chapters, chapter)
			} else {
				// An empty directory will produce an empty chapter
				chapters = append(chapters, Chapter{
					title: subDir.Name(),
					pages: []image.Image{},
				})
			}
		}
		if hasRootImages {
			for _, img := range rootContent.images {
				chapters[0].pages = append([]image.Image{img}, chapters[0].pages...)
			}
		}
		return chapters, nil
	} else if isRootChapter {
		return []Chapter{
			Chapter{
				title: path.Base(rootDirPath),
				pages: rootContent.images,
			},
		}, err
	} else {
		return nil, fmt.Errorf(`unknown manga directory structure at path "%v"`, rootDirPath)
	}
}

func getMangaDirContent(dirPath string) (MangaDirContent, error) {
	items, _ := os.ReadDir(dirPath)
	subDirNames := []string{}
	imageNames := []string{}
	for _, item := range items {
		itemName := item.Name()
		if item.IsDir() {
			subDirNames = append(subDirNames, itemName)
		} else {
			itemPath := path.Join(dirPath, itemName)
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
				imageNames = append(imageNames, itemName)
			default:
				log.Println(fmt.Sprintf(`ignored file. reason: file type is not of an image "%v", for file "%v"`, filetype, itemPath))
			}
		}
	}

	natsort.Sort(subDirNames) // in-place sort
	subDirs := []os.FileInfo{}
	for _, name := range subDirNames {
		subDir, err := os.Stat(path.Join(dirPath, name))
		if err != nil {
			return MangaDirContent{}, err
		}
		subDirs = append(subDirs, subDir)
	}

	natsort.Sort(imageNames) // in-place sort
	images := []image.Image{}
	for _, name := range imageNames {
		img, err := readImageFromPath(path.Join(dirPath, name))
		if err != nil {
			return MangaDirContent{}, err
		}
		images = append(images, img)
	}

	return MangaDirContent{
		subDirs: subDirs,
		images:  images,
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
