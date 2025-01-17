package shared

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	ID string `yaml:"id"`
	Head string `yaml:"head"`
	Word string `yaml:"entry"`
}

func ReadYamlDataFile() ([]Entry, error) {
	fmt.Println("Reading yaml file...")

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

	fmt.Printf("Decoded %d entries\n", len(entries))

	return entries, nil
}
