package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kube-sleep"
)

var TEST_SUSPENDABLES = map[string]kubesleep.Suspendable{
	"1:test-deployment": kubesleep.NewSuspendable(
		kubesleep.StatefulSet,
		"test-deployment",
		int32(2),
		nil,
	),
}

func (s *Integrationtest) TestCreateDeleteStatefile() {
	deleteNamespace, err := testNamespace("create-delete-statefile", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	_, err = s.k8s.CreateStateFile("create-delete-statefile", &kubesleep.SuspendStateFile{})
	s.Require().NoError(err)

	err = s.k8s.DeleteStateFile("create-delete-statefile")
	s.Require().NoError(err)
}

func (s *Integrationtest) TestUpdateStatefile() {
	deleteNamespace, err := testNamespace("update-statefile", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	stateFile := kubesleep.NewSuspendStateFile(
		map[string]kubesleep.Suspendable{},
		true,
	)
	_, err = s.k8s.CreateStateFile("update-statefile", &stateFile)
	s.Require().NoError(err)
	defer s.k8s.DeleteStateFile("update-statefile")

	stateFile = kubesleep.NewSuspendStateFile(
		TEST_SUSPENDABLES,
		true,
	)
	_, err = s.k8s.UpdateStateFile("update-statefile", &stateFile)

	actualStateFile, err := s.k8s.GetStateFile("update-statefile")
	slog.Debug("Read updated state file from cluster", "stateFile", stateFile)

	expected := kubesleep.NewSuspendStateFile(TEST_SUSPENDABLES, true)
	s.Require().Equal(
		&expected,
		actualStateFile,
	)
}
