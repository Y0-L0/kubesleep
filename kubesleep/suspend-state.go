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

type SuspendStateFile struct {
	suspendables map[string]Suspendable
	finished     bool
}

func NewSuspendStateFile(suspendables map[string]Suspendable, finished bool) SuspendStateFile {
	return SuspendStateFile{
		suspendables: suspendables,
		finished:     false,
	}
}

func NewSuspendStateFileFromJson(data string) *SuspendStateFile {
	var stateFileDto suspendStateFileDto
	err := json.Unmarshal(
		[]byte(data),
		&stateFileDto,
	)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshall JSON into SuspendStateFile struct. %w", err))
	}
	if stateFileDto.Suspendables == nil || stateFileDto.Finished == nil {
		panic(fmt.Errorf("missing field in state file json string. json: %s, stateFileDto: %+v", data, stateFileDto))
	}

	suspendables := map[string]Suspendable{}
	for i, s := range stateFileDto.Suspendables {
		suspendables[i] = s.fromDto()
	}
	stateFile := SuspendStateFile{
		suspendables: suspendables,
		finished:     *stateFileDto.Finished,
	}
	slog.Debug("Read state file from json", "json", data, "SuspendStateFile", stateFile)
	return &stateFile
}

func (s SuspendStateFile) ToJson() string {
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
