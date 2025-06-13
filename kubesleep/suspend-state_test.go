package kubesleep

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
