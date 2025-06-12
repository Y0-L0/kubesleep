package kubesleep

var TEST_SUSPEND_STATE_FILE = suspendStateFile{
	suspendables: TEST_SUSPENDABLES,
	finished:     false,
}

const TEST_SUSPEND_STATE_FILE_JSON = "{\"suspendables\":[{\"ManifestType\":\"Deployment\",\"Name\":\"testDeployment\",\"Replicas\":2}],\"finished\":false}"

func (s *Unittest) TestSerializeStatefile() {
	json := TEST_SUSPEND_STATE_FILE.toJson()
	s.Require().Equal(TEST_SUSPEND_STATE_FILE_JSON, json)
}

func (s *Unittest) TestDeserializeStatefile() {
	stateFile := newSuspendStateFileFromJson(TEST_SUSPEND_STATE_FILE_JSON)
	s.Require().Equal(&TEST_SUSPEND_STATE_FILE, stateFile)
}
