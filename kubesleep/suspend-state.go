package kubesleep

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type StateFileActions interface {
	Update(*SuspendStateFile) error
	Delete() error
}

type suspendStateFileDto struct {
	Suspendables []suspendableDto `json:"suspendables"`
	Finished     *bool            `json:"finished"`
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
	for _, s := range stateFileDto.Suspendables {
		sus := s.fromDto()
		suspendables[sus.Identifier()] = sus
	}
	stateFile := SuspendStateFile{
		suspendables: suspendables,
		finished:     *stateFileDto.Finished,
	}
	slog.Debug("Read state file from json", "json", data, "SuspendStateFile", stateFile)
	return &stateFile
}

func (s *SuspendStateFile) ToJson() string {
	suspendables := []suspendableDto{}
	for _, s := range s.suspendables {
		suspendables = append(suspendables, s.toDto())
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

func (s *SuspendStateFile) merge(other *SuspendStateFile) *SuspendStateFile {
	result := SuspendStateFile{
		suspendables: make(map[string]Suspendable, len(other.suspendables)),
		finished:     s.finished && other.finished,
	}

	for k, v := range other.suspendables {
		if sv, ok := s.suspendables[k]; ok {
			v = sv
		}
		result.suspendables[k] = v
	}
	slog.Debug("Merged two statefiles together", "mergedStateFile", result)
	return &result
}
