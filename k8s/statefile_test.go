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
	deleteNamespace, err := testNamespace(s.ctx, "create-delete-statefile", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	_, err = s.k8s.CreateStateFile(s.ctx, "create-delete-statefile", map[string]string{})
	s.Require().NoError(err)

	err = s.k8s.DeleteStateFile(s.ctx, "create-delete-statefile")
	s.Require().NoError(err)
}

func (s *Integrationtest) TestCreateStatefileAlreadyExists() {
	deleteNamespace, err := testNamespace(s.ctx, "create-statefile-already-exists", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	actions, err := s.k8s.CreateStateFile(s.ctx, "create-statefile-already-exists", map[string]string{})
	s.Require().NoError(err)
	defer actions.Delete(s.ctx)

	_, err = s.k8s.CreateStateFile(s.ctx, "create-statefile-already-exists", map[string]string{})
	s.Require().ErrorAs(err, new(kubesleep.StatefileAlreadyExistsError))
}

func (s *Integrationtest) TestUpdateStatefile() {
	deleteNamespace, err := testNamespace(s.ctx, "update-statefile", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	stateFile := kubesleep.NewSuspendState(
		map[string]kubesleep.Suspendable{},
		true,
	)
	actions, err := s.k8s.CreateStateFile(s.ctx, "update-statefile", stateFile.Write())
	s.Require().NoError(err)
	defer s.k8s.DeleteStateFile(s.ctx, "update-statefile")

	stateFile = kubesleep.NewSuspendState(
		TEST_SUSPENDABLES,
		true,
	)
	err = actions.Update(s.ctx, stateFile.Write())

	actualStateFile, _, err := s.k8s.GetStateFile(s.ctx, "update-statefile")
	s.Require().NoError(err)
	slog.Debug("Read updated state file from cluster", "stateFile", stateFile)

	expected := kubesleep.NewSuspendState(TEST_SUSPENDABLES, true)
	s.Require().Equal(
		&expected,
		actualStateFile,
	)
}

func (s *Integrationtest) TestUpdateStatefileOptimisticConcurrency() {
	// arrange
	namespace := "uptate-statefile-optimistic-concurrency"
	deleteNamespace, err := testNamespace(s.ctx, namespace, s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	stateFile1 := kubesleep.NewSuspendState(
		map[string]kubesleep.Suspendable{},
		true,
	)
	stateFile2 := kubesleep.NewSuspendState(
		TEST_SUSPENDABLES,
		true,
	)
	actions, err := s.k8s.CreateStateFile(s.ctx, namespace, stateFile1.Write())
	s.Require().NoError(err)
	defer s.k8s.DeleteStateFile(s.ctx, namespace)

	// act
	_, initialActions, err := s.k8s.GetStateFile(s.ctx, namespace)
	s.Require().NoError(err)
	err = actions.Update(s.ctx, stateFile2.Write())
	s.Require().NoError(err)
	err = initialActions.Update(s.ctx, stateFile1.Write())
	s.Require().Error(err)

	actualStateFile, _, err := s.k8s.GetStateFile(s.ctx, namespace)
	s.Require().NoError(err)

	s.Require().Equal(
		&stateFile2,
		actualStateFile,
	)
}
