# Mapaki

Mapaki is a no-brainer **ma**nga **pa**cker for **ki**ndle.

## Workflow and usage

1. Use [HakuNeko](https://github.com/manga-download/hakuneko) to download manga to your computer.
2. Run the following command on the downloaded mangas' directory.
    ```shell
    mapaki -i "./Manga Name"
    ```
   This command will generate a single `Manga Name.azw3` file as its output.
3. Use [Calibre](https://github.com/kovidgoyal/calibre) to upload generated file to your Kindle device.

## Mapaki features

### Automatic chapter discovery

It will automatically figure out file system layout of the downloaded manga. It supports two organizational patterns:

* The first
    ```text
      Manga Name/             // root dir
          title_cover.jpg     // title cover image (optional)
          Chapter 1/          // chapter 1 dir
              1.jpg           // page 1
              2.jpg           // page 2
              ...             // etc.
          Chapter 2/
              1.jpg
              2.jpg
              ...
          ...
    ```
* The second
    ```text
      Manga Name/                 // root dir
          title_cover.jpg         // title cover image  (optional)
          Volume 1/               // volume 1 dir
              volume_cover.jpg    // volume cover image (optional)
              Chapter 1/          // chapter 1 dir
                  1.jpg           // page 1
                  2.jpg           // page 2
                  ...             // etc.
              Chapter 2/
                  1.jpg
                  2.jpg
                  ...
              ...
          Volume 2
              volume_cover.jpg
              Chapter 14/
                  1.jpg
                  2.jpg
                  ...
              Chapter 15/
                  1.jpg
                  2.jpg
                  ...
              ...
    ```

The good thing is that the names of the chapters, volumes, pages and covers does not follow any special convention. They
will be "naturally sorted" ([read more](https://github.com/facette/natsort)) while respecting the directory layout.

### Auto cropping

By default, Mapaki will crop out white space around all images. Auto cropping can be disabled via following
flag `--disable-auto-crop=true`.

### Auto double page handling

By default, Makapi will ensure that every double page is displayed firstly "as is", followed by the page's right side,
and lastly, followed by the page's left side. This behaviour can be changed via the following
flag `--double-page [mode]`:

* `--double-page only-double`
* `--double-page only-split`
* `--double-page split-then-double`
* `--double-page double-then-split` (default)

## Install

Mapaki can be installed from source easily if you already have access to a Go toolchain. Otherwise, follow
the [Go installation instructions](https://go.dev/doc/install) for your operating system, then execute
the following command.

``` shell
go install github.com/tsopeh/mapaki@latest
```

Afterward, verify your installation succeeded by executing the application on the command line.

``` shell
mapaki --version
```

On many systems, the Go binary directory is not added to the list of directories searched for executables by default.
If you get a "command not found" or similar error after the previous command, run the following command and try again.
If you are using Windows, please find out how to add directories to the lookup path yourself, as there does not seem to
be any quality documentation that I could link here.

``` shell
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Thanks

Many thanks to Leo Gaskin ([@leotaku](https://github.com/leotaku)) 🎉. This project was inspired and heavily influenced
by his work on [kojirou](https://github.com/leotaku/kojirou). Some code (e.g. the `crop` module, image template string,
the "install step" in this readme) has been copied directly from `kojirou`. Leo also developed
the [mobi](https://github.com/leotaku/mobi) library for Go, that handles packing of the images into the `.azw3` file.
