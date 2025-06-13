package k8s

import "github.com/Y0-L0/kubesleep/kubesleep"

func (s *Integrationtest) TestDeploymentStatefulsetNameConflict() {
	deleteNamespace, err := testNamespace("deployment-statefulset-name-conflict", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	deleteStatefulSet, err := CreateStatefulSet(*s.k8s, "deployment-statefulset-name-conflict", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer deleteStatefulSet()

	deleteDeployment, err := CreateDeployment(*s.k8s, "deployment-statefulset-name-conflict", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer deleteDeployment()
}

func (s *Integrationtest) TestGetSuspendableNamespace() {
	deleteNamespace, err := testNamespace("get-suspendable-namespace", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	namespace, err := s.k8s.GetSuspendableNamespace("get-suspendable-namespace")
	s.Require().NoError(err)

	s.Require().Equal(
		kubesleep.NewSuspendableNamespace("get-suspendable-namespace", true),
		namespace,
	)
}

func (s *Integrationtest) TestGetNonSuspendableNamespace() {
	deleteNamespace, err := testNamespace("get-non-suspendable-namespace", s.k8s, true)
	s.Require().NoError(err)
	defer deleteNamespace()

	namespace, err := s.k8s.GetSuspendableNamespace("get-non-suspendable-namespace")
	s.Require().NoError(err)

	s.Require().Equal(
		kubesleep.NewSuspendableNamespace("get-non-suspendable-namespace", false),
		namespace,
	)
}
