# Alar Voice Corpus

Voice corpus for the Alar Kannada-English Dictionary generated using [Google TTS API](https://cloud.google.com/text-to-speech/custom-voice/docs). Data for this corpus comes from [Alar-dict/data](https://github.com/alar-dict/data).

## Manifest File

The `manifest.csv` file serves as a structured reference to all the audio files in the dataset. It provides metadata and additional information needed for processing or analysis.

### File Structure

The `manifest.csv` file is a comma-separated values (CSV) file with the following columns:

| Column Name | Description                                                  | Example Value                   |
| ----------- | ------------------------------------------------------------ | ------------------------------- |
| `File Path` | Relative path to the audio file.                             | `./audio/30000-39999/35112.mp3` |
| `Duration`  | Duration of the audio file in seconds.                       | `0.96`                          |
| `Word`      | Text transcription or label corresponding to the audio file. | `ಕನ್ನಡ`                         |
| `ID`        | Unique identifier for the word.                              | `35112`                         |
| `Head`      | First character of the word.                                 | `ಕ`                             |

## Scripts

There are a couple of scripts: one to generate the manifest file and the other to generate the dataset yourself.

To generate the dataset yourself, setup gcloud cli as detailed out [here](https://cloud.google.com/text-to-speech/docs/create-audio-text-client-libraries).

```bash
cd scripts

# Generating the manifest file
go run ./manifest.go

# Generate the dataset yourself
gcloud init
gcloud auth application-default login
mkdir data
curl https://raw.githubusercontent.com/alar-dict/data/refs/heads/master/alar.yml -o data/data.yml
go run ./tts
```

## Audio File Organization

The audio files are grouped into folders based on their ID ranges, with each folder containing files named after their unique numeric ID. For example, `audio/1-9999/1234.mp3` contains the audio file for `ID 1234`, and `10000-19999/15000.mp3` contains the audio file for `ID 15000`. Each folder covers a range of 10,000 IDs.
