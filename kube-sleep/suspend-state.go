package kubesleep

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type suspendStateFileDto struct {
	Suspendables map[string]suspendableDto `json:"suspendables"`
	Finished     *bool                     `json:"finished"`
}

type suspendStateFile struct {
	suspendables map[string]suspendable
	finished     bool
}

func newSuspendStateFile(suspendables map[string]suspendable) suspendStateFile {
	return suspendStateFile{
		suspendables: suspendables,
		finished:     false,
	}
}

func newSuspendStateFileFromJson(data string) *suspendStateFile {
	var stateFileDto suspendStateFileDto
	err := json.Unmarshal(
		[]byte(data),
		&stateFileDto,
	)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshall JSON into suspendStateFile struct. %w", err))
	}
	if stateFileDto.Suspendables == nil || stateFileDto.Finished == nil {
		panic(fmt.Errorf("Missing field in state file json string. json: %s, stateFileDto: %+v", data, stateFileDto))
	}

	suspendables := map[string]suspendable{}
	for i, s := range stateFileDto.Suspendables {
		suspendables[i] = s.fromDto()
	}
	stateFile := suspendStateFile{
		suspendables: suspendables,
		finished:     *stateFileDto.Finished,
	}
	slog.Debug("Read state file from json", "json", data, "suspendStateFile", stateFile)
	return &stateFile
}

func (s suspendStateFile) toJson() string {
	suspendables := map[string]suspendableDto{}
	for i, s := range s.suspendables {
		suspendables[i] = s.toDto()
	}
	stateFileDto := suspendStateFileDto{
		suspendables,
		&s.finished,
	}
	jsonData, err := json.MarshalIndent(stateFileDto, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal Data to JSON: %w", err))
	}
	slog.Debug("Serialized state file to json", "jsonString", jsonData)
	return string(jsonData)
}
