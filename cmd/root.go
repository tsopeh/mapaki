package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tsopeh/mapaki/cmd/packer"
	"math"
	"os"
	"runtime"
)

var rootCmd = &cobra.Command{
	Use:     "mapaki",
	Short:   "A no-brainer manga packer for Kindle.",
	Version: "1.3",
	RunE: func(cmd *cobra.Command, args []string) error {
		disableAutoCrop, _ := cmd.Flags().GetBool("disable-auto-crop")
		leftToRight, _ := cmd.Flags().GetBool("left-to-right")
		doublePage, _ := cmd.Flags().GetString("double-page")
		title, _ := cmd.Flags().GetString("title")
		inputDir, _ := cmd.Flags().GetString("input-dir")
		outputFilePath, _ := cmd.Flags().GetString("output-file-path")
		desiredCoresCount, _ := cmd.Flags().GetInt("cores")

		totalCoresCount := runtime.NumCPU()
		usedCoresCount := 1
		if desiredCoresCount > 1 {
			usedCoresCount = int(math.Min(float64(desiredCoresCount), float64(totalCoresCount)))
		} else {
			usedCoresCount = int(math.Max(math.Floor(float64(totalCoresCount)*0.8), 1))
		}
		runtime.GOMAXPROCS(usedCoresCount)
		fmt.Println(fmt.Sprintf(`mapaki is using %v out of %v CPU cores.`, usedCoresCount, totalCoresCount))

		err := packer.PackMangaForKindle(packer.PackForKindleParams{
			RootDir:         inputDir,
			DisableAutoCrop: disableAutoCrop,
			LeftToRight:     leftToRight,
			DoublePage:      doublePage,
			Title:           title,
			OutputFilePath:  outputFilePath,
			CoresCount:      usedCoresCount,
		})
		return err
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Bool("disable-auto-crop", false, `should disable auto cropping (default: false)`)
	rootCmd.Flags().Bool("left-to-right", false, `left to right reading direction (default: false)`)
	rootCmd.Flags().String("double-page", "double-then-split", `what to do with double pages. options: "only-double", "only-split", "split-then-double" and "double-then-split"`)
	rootCmd.Flags().String("title", "", `manga title. does not affect the output file path`)
	rootCmd.Flags().StringP("input-dir", "i", "", `path to the manga root directory (required)`)
	rootCmd.Flags().StringP("output-file-path", "o", "", `output path that includes the filename and '.azw3' extension (default: "../[manga dir name].azw3")`)
	rootCmd.Flags().Int("cores", -1, `number of CPU cores that will be used for processing manga pages (by default use 80% of all cores)`)
	rootCmd.MarkFlagRequired("input-dir")
	if err := rootCmd.ParseFlags(os.Args); err != nil {
		// Don't exit. `ParseFlags` will fail if `--version` flag gets passed.
		// log.Fatalf(`init: failed to parse the command input. %w`, err)
	}
}
