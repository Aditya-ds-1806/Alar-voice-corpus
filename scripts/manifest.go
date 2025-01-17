package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/go-mp3"
	"gopkg.in/yaml.v3"
)

type Entry struct {
	ID string `yaml:"id"`
	Head string `yaml:"head"`
	Word string `yaml:"entry"`
}

func readYamlDataFile() ([]Entry, error) {
	var entries []Entry
	file, err := os.Open("./data/data.yml")

	if err != nil {
		fmt.Printf("Error opening data file\n")
		return entries, err
	}

	defer file.Close()
	
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&entries)

	if err != nil {
		fmt.Println("Error reading yaml data file")
		return entries, err
	}

	return entries, nil
}

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

func generateManifest(dirPath string, entries* []Entry, files* []string, outFile* os.File) *[]string {
	dirEntries, err :=	os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading dir: %v\n", dirPath)
		return files
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

			if err == nil {
				*files = append(*files, filePath)
				duration := getAudioDuration(filePath)
				entry := (*entries)[audioId - 1]
				outFile.WriteString(fmt.Sprintf("%v, %v, %v, %v, %v\n", audioId, entry.Head, entry.Word, duration, filePath))
			}
		}
	}

	return files
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
	}

	csv.WriteString("ID, Head, Word, Duration, File Path\n")

	entries, err := readYamlDataFile()
	if err != nil {
		fmt.Println(err)
		return
	}

	generateManifest(*dirPath, &entries, &filePaths, csv)
}
