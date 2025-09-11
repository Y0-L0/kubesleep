package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
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
	deleteNamespace, err := testNamespace("create-delete-statefile", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	_, err = s.k8s.CreateStateFile("create-delete-statefile", map[string]string{})
	s.Require().NoError(err)

	err = s.k8s.DeleteStateFile("create-delete-statefile")
	s.Require().NoError(err)
}

func (s *Integrationtest) TestCreateStatefileAlreadyExists() {
	deleteNamespace, err := testNamespace("create-statefile-already-exists", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	actions, err := s.k8s.CreateStateFile("create-statefile-already-exists", map[string]string{})
	s.Require().NoError(err)
	defer actions.Delete()

	_, err = s.k8s.CreateStateFile("create-statefile-already-exists", map[string]string{})
	s.Require().ErrorAs(err, new(kubesleep.StatefileAlreadyExistsError))
}

func (s *Integrationtest) TestUpdateStatefile() {
	deleteNamespace, err := testNamespace("update-statefile", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	stateFile := kubesleep.NewSuspendStateFile(
		map[string]kubesleep.Suspendable{},
		true,
	)
	actions, err := s.k8s.CreateStateFile("update-statefile", kubesleep.WriteSuspendState(&stateFile))
	s.Require().NoError(err)
	defer s.k8s.DeleteStateFile("update-statefile")

	stateFile = kubesleep.NewSuspendStateFile(
		TEST_SUSPENDABLES,
		true,
	)
	err = actions.Update(kubesleep.WriteSuspendState(&stateFile))

	actualStateFile, _, err := s.k8s.GetStateFile("update-statefile")
	s.Require().NoError(err)
	slog.Debug("Read updated state file from cluster", "stateFile", stateFile)

	expected := kubesleep.NewSuspendStateFile(TEST_SUSPENDABLES, true)
	s.Require().Equal(
		&expected,
		actualStateFile,
	)
}

func (s *Integrationtest) TestUpdateStatefileOptimisticConcurrency() {
	// arrange
	namespace := "uptate-statefile-optimistic-concurrency"
	deleteNamespace, err := testNamespace(namespace, s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	stateFile1 := kubesleep.NewSuspendStateFile(
		map[string]kubesleep.Suspendable{},
		true,
	)
	stateFile2 := kubesleep.NewSuspendStateFile(
		TEST_SUSPENDABLES,
		true,
	)
	actions, err := s.k8s.CreateStateFile(namespace, kubesleep.WriteSuspendState(&stateFile1))
	s.Require().NoError(err)
	defer s.k8s.DeleteStateFile(namespace)

	// act
	_, initialActions, err := s.k8s.GetStateFile(namespace)
	s.Require().NoError(err)
	err = actions.Update(kubesleep.WriteSuspendState(&stateFile2))
	err = initialActions.Update(kubesleep.WriteSuspendState(&stateFile1))
	s.Require().Error(err)

	actualStateFile, _, err := s.k8s.GetStateFile(namespace)
	s.Require().NoError(err)

	s.Require().Equal(
		&stateFile2,
		actualStateFile,
	)
}
