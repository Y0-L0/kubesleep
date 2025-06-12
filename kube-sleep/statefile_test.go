package kubesleep

import (
	"log/slog"
)

func (s *Integrationtest) TestCreateDeleteStatefile() {
	deleteNamespace, err := testNamespace("create-delete-statefile", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	_, err = s.k8s.createStateFile("create-delete-statefile", &suspendStateFile{})
	s.Require().NoError(err)

	err = s.k8s.deleteStateFile("create-delete-statefile")
	s.Require().NoError(err)
}

func (s *Integrationtest) TestUpdateStatefile() {
	deleteNamespace, err := testNamespace("update-statefile", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	sus := []suspendable{
		suspendable{
			manifestType: "Deployment",
			name:         "testDeployment",
			replicas:     int32(2),
		},
	}

	_, err = s.k8s.createStateFile("update-statefile", &suspendStateFile{
		suspendables: []suspendable{},
		finished:     false,
	})
	s.Require().NoError(err)
	defer s.k8s.deleteStateFile("update-statefile")

	_, err = s.k8s.updateStateFile("update-statefile", &suspendStateFile{
		suspendables: sus,
		finished:     true,
	})

	stateFile, err := s.k8s.getStateFile("update-statefile")
	slog.Debug("Read updated state file from cluster", "stateFile", stateFile)
	s.Require().True(stateFile.finished)
	s.Require().Equal(sus, stateFile.suspendables)
}
