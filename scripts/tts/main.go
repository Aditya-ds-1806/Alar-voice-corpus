package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"context"

	"alar-corpus/shared"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)


func getSubFolderName(audioId string) string {
	id, err := strconv.Atoi(audioId)
	if err != nil {
		return ""
	}

	const dirSize = 10_000
	startId := id - (id % dirSize)
	endId := startId + dirSize - 1

	if startId == 0 {
		startId = 1
	}

	return fmt.Sprintf("%d-%d", startId, endId)
}

func performTTS(ctx context.Context, client *texttospeech.Client, entry* shared.Entry, wg* sync.WaitGroup) {
	defer wg.Done()

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: entry.Word},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "kn-IN",
			Name: "kn-IN-Standard-A",
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate: 0.85,
			Pitch: 0,
			VolumeGainDb: 0,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		fmt.Printf("Error synthesizing speech for entry %s: %v\n", entry.ID, err)
		return
	}

	filename := fmt.Sprintf("../audio/%s/%s.mp3", getSubFolderName(entry.ID), entry.ID)
	outFile, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file for entry %s: %v\n", entry.ID, err)
		return
	}
	defer outFile.Close()

	_, err = outFile.Write(resp.AudioContent)

	if err != nil {
		fmt.Printf("Error writing to file for entry %s: %v\n", entry.ID, err)
		return
	}
	
	fmt.Printf("Audio content written to file: %v\n", filename)
}

func initDirectories(filesCount int) {
	err := os.MkdirAll("../audio", 0644)

	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	dirSize := 10_000


	for i := 0; i < filesCount; i += dirSize {
		startId := i
		endId := startId + dirSize - 1
		if startId == 0 {
			startId = 1
		}
		
		dir := fmt.Sprintf("../audio/%d-%d", startId, endId)
		err = os.MkdirAll(dir, 0644)
		if err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("Initialized dir: %s\n", dir)
		}
	}
}

func getEntriesToProcess(entries []shared.Entry) []shared.Entry {
	var entriesToProcess []shared.Entry

	dirEntries, err := os.ReadDir("../audio")
	if err != nil {
		fmt.Println("Unable to read audio dir")
		os.Exit(1)
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}

		startId, endId, found := strings.Cut(dirEntry.Name(), "-")
		if !found {
			continue
		}

		parsedStartId, err := strconv.Atoi(startId)
		if err != nil {
			continue
		}

		parsedEndId, err := strconv.Atoi(endId)
		if err != nil {
			continue
		}

		for audioId := parsedStartId; audioId <= int(math.Min(float64(parsedEndId), float64(len(entries)))); audioId++ {
			_, err := os.Stat(fmt.Sprintf("../audio/%d-%d/%d.mp3", parsedStartId, parsedEndId, audioId))
			if err != nil {
				entriesToProcess = append(entriesToProcess, entries[audioId - 1])
			}
		}
	}

	return entriesToProcess
}

func main() {
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)

	if err != nil {
		fmt.Println("Error initializing google-tts client", err)
		return
	}

	defer client.Close()

	allEntries, err := shared.ReadYamlDataFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	initDirectories(len(allEntries))
	entries := getEntriesToProcess(allEntries)

	fmt.Printf("%d entries to process out of %d\n", len(entries), len(allEntries))

	const skip int = 0
	const concurrency int = 20
	const pageSize int = 1000
	pages := int(math.Ceil(float64(len(entries)) / float64(pageSize)))
	startPage := int(math.Ceil(float64(skip) / float64(pageSize)))

	for pg := startPage; pg < pages; pg++ {
		var wg sync.WaitGroup
		startIdx := pg * pageSize
		endIdx := int(math.Min(float64(startIdx + pageSize), float64(len(entries))))
		wordsPerWorker := int(math.Ceil(float64(endIdx - startIdx) / float64(concurrency)))

		startTime := time.Now()
		fmt.Printf("Processing page %d of %d\n", pg + 1, pages)

		for j := 0; j < concurrency; j++ {
			var workerEndIdx int
			workerStartIdx := startIdx + (j * wordsPerWorker)

			if j == concurrency - 1 {
				workerEndIdx = endIdx
			} else {
				workerEndIdx = int(math.Min(float64(workerStartIdx + wordsPerWorker), float64(len(entries))))
			}

			if workerStartIdx < len(entries) && workerEndIdx <= len(entries) {
				for _, entry := range entries[workerStartIdx:workerEndIdx] {
					wg.Add(1)
					go performTTS(ctx, client, &entry, &wg)
				}
			}
		}

		wg.Wait()

		endTime := time.Now()
		diff := endTime.Sub(startTime)
		buffer := 5 * time.Second
		sleepTime := time.Minute - diff + buffer

		if sleepTime > 0 && pg < pages - 1 {
			fmt.Printf("Sleeping for %v\n", sleepTime)
			time.Sleep(sleepTime)
		} else {
			fmt.Println("No sleep required, skipping sleep!")
		}
	}

	fmt.Println("Finished performing TTS on data!")
}
