package kubesleep

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type suspendStateFileDto struct {
	Suspendables *[]suspendableDto
	Finished     *bool
}

type suspendStateFile struct {
	suspendables []suspendable
	finished     bool
}

func newSuspendStateFile(suspendables []suspendable) suspendStateFile {
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

	suspendables := make([]suspendable, len(*stateFileDto.Suspendables))
	for i, s := range *stateFileDto.Suspendables {
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
	suspendables := make([]suspendableDto, len(s.suspendables))
	for i, s := range s.suspendables {
		suspendables[i] = s.toDto()
	}
	stateFileDto := suspendStateFileDto{
		&suspendables,
		&s.finished,
	}
	jsonData, err := json.Marshal(stateFileDto)
	if err != nil {
		panic(fmt.Errorf("failed to marshal Data to JSON: %w", err))
	}
	slog.Debug("Serialized state file to json", "jsonString", jsonData)
	return string(jsonData)
}
