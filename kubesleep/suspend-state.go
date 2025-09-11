package kubesleep

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

// Versioned statefile keys stored in the ConfigMap's data
const (
	STATE_FILE_KEY_V1 = "kubesleep.json"
	STATE_FILE_KEY_V2 = "kubesleep.v2.json"
)

type SuspendStateActions interface {
	Update(map[string]string) error
	Delete() error
}

type suspendStateDto struct {
	Suspendables []suspendableDto `json:"suspendables"`
	Finished     *bool            `json:"finished"`
}

type SuspendState struct {
	suspendables map[string]Suspendable
	finished     bool
}

func NewSuspendState(suspendables map[string]Suspendable, finished bool) SuspendState {
	return SuspendState{
		suspendables: suspendables,
		finished:     false,
	}
}

func (s *SuspendState) merge(other *SuspendState) *SuspendState {
	result := SuspendState{
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

func (s *SuspendState) toJson() string {
	suspendables := []suspendableDto{}
	for _, s := range s.suspendables {
		suspendables = append(suspendables, s.toDto())
	}
	stateFileDto := suspendStateDto{
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

// hasCronJobs returns true if any suspendable is of type CronJob
func (s *SuspendState) hasCronJobs() bool {
	for _, sus := range s.suspendables {
		if sus.manifestType == CronJob {
			return true
		}
	}
	return false
}

func (s *SuspendState) Write() map[string]string {
	json := s.toJson()

	if s.hasCronJobs() {
		return map[string]string{
			STATE_FILE_KEY_V2: json,
			STATE_FILE_KEY_V1: `{"message":"please upgrade kubesleep to v0.4.0 or higher to gain CronJob support"}`,
		}
	}
	return map[string]string{
		STATE_FILE_KEY_V2: json,
		STATE_FILE_KEY_V1: json,
	}
}

func ReadSuspendState(data map[string]string) *SuspendState {
	if v2, ok := data[STATE_FILE_KEY_V2]; ok && v2 != "" {
		return newSuspendStateFromJson(v2)
	}
	if v1, ok := data[STATE_FILE_KEY_V1]; ok && v1 != "" {
		return newSuspendStateFromJson(v1)
	}
	panic(fmt.Errorf("missing both %s and %s in statefile keys in configmap", STATE_FILE_KEY_V1, STATE_FILE_KEY_V2))
}

func newSuspendStateFromJson(data string) *SuspendState {
	var stateFileDto suspendStateDto
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
	stateFile := SuspendState{
		suspendables: suspendables,
		finished:     *stateFileDto.Finished,
	}
	slog.Debug("Read state file from json", "json", data, "SuspendStateFile", stateFile)
	return &stateFile
}
