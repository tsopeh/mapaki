package packer

import (
	"log"
	"os"
	"reflect"
	"testing"
)

func TestGetNaturallyOrderedImagePathsForSomeSimpleManga(t *testing.T) {
	log.Println(os.Getwd())
	images, err := discoverMangaChapters("./test_assets/Some Simple Manga")
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(images, []string{
		"test_assets/Some Simple Manga/Chapter 1/1.jpg",
		"test_assets/Some Simple Manga/Chapter 1/2.jpg",
		"test_assets/Some Simple Manga/Chapter 1/10.jpg",
		"test_assets/Some Simple Manga/Chapter 2/1.jpg",
		"test_assets/Some Simple Manga/Chapter 2/2.jpg",
		"test_assets/Some Simple Manga/Chapter 2/10.jpg",
		"test_assets/Some Simple Manga/Chapter 10/1.jpg",
		"test_assets/Some Simple Manga/Chapter 10/2.jpg",
		"test_assets/Some Simple Manga/Chapter 10/10.jpg",
	}) {
		t.Fail()
	}
}

func TestGetNaturallyOrderedImagePathsForSomeMangaWithTitlePageAndVolumes(t *testing.T) {
	log.Println(os.Getwd())
	images, err := discoverMangaChapters("./test_assets/Some Manga With Title Page And Volumes")
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(images, []string{
		"test_assets/Some Manga With Title Page And Volumes/title.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 1/vol1.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 1/Chapter 1/1.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 1/Chapter 1/2.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 1/Chapter 1/10.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 1/Chapter 2/1.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 1/Chapter 2/2.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 1/Chapter 2/10.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 2/vol2.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 2/Chapter 10/1.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 2/Chapter 10/2.jpg",
		"test_assets/Some Manga With Title Page And Volumes/Volume 2/Chapter 10/10.jpg",
	}) {
		t.Fail()
	}
}
