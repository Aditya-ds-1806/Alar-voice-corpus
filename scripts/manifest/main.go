package main

import (
	"alar-corpus/shared"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/go-mp3"
)

func getAudioDuration (filePath string) float64 {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error reading file %v\n", filePath)
		return 0
	}
	defer file.Close()
	
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		fmt.Printf("Error decoding file %v\n", filePath)
		return 0
	}

	duration := float64(decoder.Length()) / float64(decoder.SampleRate()) / 4

	return duration
}

func generateManifest(dirPath string, entries* []shared.Entry, files* []string, outFile* os.File) {
	fmt.Printf("Looking for files in dir: %s\n", dirPath)

	dirEntries, err :=	os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading dir: %v\n", dirPath)
		return
	}

	for _, dirEntry := range dirEntries {
		dirName := dirEntry.Name()

		if dirName[0] == '.' {
			continue
		}

		if dirEntry.IsDir() {
			newDirPath := fmt.Sprintf("%v/%v", dirPath, dirName)
			generateManifest(newDirPath, entries, files, outFile)
		} else {
			filePath := fmt.Sprintf("%v/%v", dirPath, dirName)
			audioId, err := strconv.Atoi(strings.Split(dirName, ".")[0])
			fmt.Printf("Generating manifest for: %s\n", filePath)

			if err == nil {
				*files = append(*files, filePath)
				duration := getAudioDuration(filePath)
				entry := (*entries)[audioId - 1]
				outFile.WriteString(fmt.Sprintf("%v, %v, %v, %v, %v\n", audioId, entry.Head, entry.Word, duration, filePath))
			}
		}
	}
}

func main() {
	var filePaths []string
	dirPath := flag.String("dir", "../audio", "Directory path")
	outFile := flag.String("out", "../manifest.csv", "Output file path")

	flag.Usage = func() {
		fmt.Println("Usage: go run manifest.go [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	csv, err := os.OpenFile(*outFile, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error creating out file: %v\n", *outFile)
		os.Exit(1)
	}

	csv.WriteString("ID, Head, Word, Duration, File Path\n")

	entries, err := shared.ReadYamlDataFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	generateManifest(*dirPath, &entries, &filePaths, csv)
	fmt.Println("Finished generating manifests for all files!")
}
