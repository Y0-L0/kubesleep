package kubesleep

import "github.com/stretchr/testify/mock"

var TEST_SUSPEND_STATE_FILE = SuspendStateFile{
	suspendables: TEST_SUSPENDABLES,
	finished:     false,
}

const TEST_SUSPEND_STATE_FILE_JSON = `{
  "suspendables": {
    "0:test-deployment": {
      "ManifestType": 0,
      "Name": "test-deployment",
      "Replicas": 2
    }
  },
  "finished": false
}`

func (s *Unittest) TestSerializeStatefile() {
	json := TEST_SUSPEND_STATE_FILE.ToJson()
	s.Require().Equal(TEST_SUSPEND_STATE_FILE_JSON, json)
}

func (s *Unittest) TestDeserializeStatefile() {
	stateFile := NewSuspendStateFileFromJson(TEST_SUSPEND_STATE_FILE_JSON)
	s.Require().Equal(&TEST_SUSPEND_STATE_FILE, stateFile)
}

type MockStateFileActions struct {
	mock.Mock
}

func (m *MockStateFileActions) Update(data *SuspendStateFile) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockStateFileActions) Delete() error {
	args := m.Called()
	return args.Error(0)
}
