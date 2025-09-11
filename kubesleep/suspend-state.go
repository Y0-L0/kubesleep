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

type StateFileActions interface {
	Update(map[string]string) error
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

// hasCronJobs returns true if any suspendable is of type CronJob
func (s *SuspendStateFile) hasCronJobs() bool {
	for _, sus := range s.suspendables {
		if sus.manifestType == CronJob {
			return true
		}
	}
	return false
}

// WriteSuspendState writes the state into the provided data map using versioned keys.
// If CronJobs are present, v2 contains the real data and v1 contains an upgrade message.
// Otherwise, both v1 and v2 contain identical real data.
func WriteSuspendState(state *SuspendStateFile) map[string]string {
	data := map[string]string{}
	json := state.ToJson()
	if state.hasCronJobs() {
		data[STATE_FILE_KEY_V2] = json
		data[STATE_FILE_KEY_V1] = `{"message":"please upgrade kubesleep to a version that supports CronJobs"}`
		return data
	}
	data[STATE_FILE_KEY_V1] = json
	data[STATE_FILE_KEY_V2] = json
	return data
}

// ReadSuspendState reads state from the provided data map preferring v2 and falling back to v1.
// Panics if neither key exists.
func ReadSuspendState(data map[string]string) *SuspendStateFile {
	if v2, ok := data[STATE_FILE_KEY_V2]; ok && v2 != "" {
		return NewSuspendStateFileFromJson(v2)
	}
	if v1, ok := data[STATE_FILE_KEY_V1]; ok && v1 != "" {
		return NewSuspendStateFileFromJson(v1)
	}
	panic(fmt.Errorf("missing both %s and %s in statefile configmap", STATE_FILE_KEY_V1, STATE_FILE_KEY_V2))
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
