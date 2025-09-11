package kubesleep

import (
	"fmt"
	"github.com/stretchr/testify/mock"
)

type MockStateFileActions struct {
	mock.Mock
}

func (m *MockStateFileActions) Update(data map[string]string) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockStateFileActions) Delete() error {
	args := m.Called()
	return args.Error(0)
}

var TEST_SUSPEND_STATE_FILE = SuspendStateFile{
	suspendables: TEST_SUSPENDABLES,
	finished:     false,
}

const TEST_SUSPEND_STATE_FILE_JSON = `{
  "suspendables": [
    {
      "ManifestType": 0,
      "Name": "test-deployment",
      "Replicas": 2
    }
  ],
  "finished": false
}`

func (s *Unittest) TestStatefileJson() {
	json := TEST_SUSPEND_STATE_FILE.ToJson()
	stateFile := NewSuspendStateFileFromJson(json)
	s.Require().Equal(&TEST_SUSPEND_STATE_FILE, stateFile)
}

func (s *Unittest) TestEmptyStatefileJson() {
	expectedStateFile := &SuspendStateFile{suspendables: map[string]Suspendable{}}
	json := expectedStateFile.ToJson()
	stateFile := NewSuspendStateFileFromJson(json)
	s.Require().Equal(expectedStateFile, stateFile)
}

func (s *Unittest) TestSerializeStatefile() {
	json := TEST_SUSPEND_STATE_FILE.ToJson()
	s.Require().Equal(TEST_SUSPEND_STATE_FILE_JSON, json)
}

func (s *Unittest) TestDeserializeStatefile() {
	stateFile := NewSuspendStateFileFromJson(TEST_SUSPEND_STATE_FILE_JSON)
	s.Require().Equal(&TEST_SUSPEND_STATE_FILE, stateFile)
}

func (s *Unittest) TestDeserializeStatefileInvalidJson() {
	s.Require().Panics(func() {
		_ = NewSuspendStateFileFromJson("{\"")
	})
}

func (s *Unittest) TestDeserializeStatefileIncompleteJson() {
	s.Require().Panics(func() {
		_ = NewSuspendStateFileFromJson(`{"finished": false}`)
	})
}

func (s *Unittest) TestMergeStateFiles() {
	a := Suspendable{1, "a", 1, nil}
	b := Suspendable{2, "b", 2, nil}
	c := Suspendable{1, "c", 3, nil}
	c2 := Suspendable{1, "c", 30, nil}
	d := Suspendable{2, "d", 4, nil}
	e := Suspendable{1, "e", 5, nil}
	existing := SuspendStateFile{
		map[string]Suspendable{
			a.Identifier(): a,
			b.Identifier(): b,
			c.Identifier(): c,
		},
		true,
	}
	new := SuspendStateFile{
		map[string]Suspendable{
			b.Identifier():  b,
			c2.Identifier(): c2,
			d.Identifier():  d,
			e.Identifier():  e,
		},
		false,
	}
	expected := &SuspendStateFile{
		map[string]Suspendable{
			b.Identifier(): b,
			c.Identifier(): c,
			d.Identifier(): d,
			e.Identifier(): e,
		},
		false,
	}

	actual := existing.merge(&new)

	s.Require().Equal(expected, actual)
}

// makeTestStateFile creates a state file that always contains a Deployment "d1"
// and conditionally includes a CronJob "cj1" when includeCronJob is true.
func makeTestStateFile(includeCronJob bool) SuspendStateFile {
	suspendables := map[string]Suspendable{}
	d1 := NewSuspendable(Deplyoment, "d1", 1, nil)
	suspendables[d1.Identifier()] = d1
	if includeCronJob {
		cj1 := NewSuspendable(CronJob, "cj1", 1, nil)
		suspendables[cj1.Identifier()] = cj1
	}
	return NewSuspendStateFile(suspendables, false)
}

func (s *Unittest) TestWriteSuspendStateWithoutCronJobs() {
	state := makeTestStateFile(false)
	data := WriteSuspendState(&state)

	// Both v1 and v2 exist and are identical
	s.Require().Contains(data, STATE_FILE_KEY_V1)
	s.Require().Contains(data, STATE_FILE_KEY_V2)
	s.Require().Equal(data[STATE_FILE_KEY_V1], data[STATE_FILE_KEY_V2])
}

func (s *Unittest) TestWriteSuspendStateWithCronJobs() {
	state := makeTestStateFile(true)
	data := WriteSuspendState(&state)

	// v2 contains real data; v1 contains an upgrade message
	s.Require().Contains(data, STATE_FILE_KEY_V1)
	s.Require().Contains(data, STATE_FILE_KEY_V2)
	s.Require().NotEqual(data[STATE_FILE_KEY_V1], data[STATE_FILE_KEY_V2])
	s.Require().Contains(data[STATE_FILE_KEY_V1], "please upgrade kubesleep")
}

func (s *Unittest) TestOldVersionParseFails() {
	state := makeTestStateFile(true)
	data := WriteSuspendState(&state)
	msgJson := data[STATE_FILE_KEY_V1]

	// Old versions attempting to parse v1 should panic with a descriptive error
	expected := fmt.Sprintf(
		"missing field in state file json string. json: %s, stateFileDto: %+v",
		msgJson,
		suspendStateFileDto{},
	)
	s.Require().PanicsWithError(expected, func() { _ = NewSuspendStateFileFromJson(msgJson) })
}
