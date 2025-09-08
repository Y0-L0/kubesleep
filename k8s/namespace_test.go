package k8s

import (
	"github.com/Y0-L0/kubesleep/kubesleep"
)

var STANDARD_NAMESPACES = []kubesleep.SuspendableNamespace{
	kubesleep.NewSuspendableNamespace("default", false),
	kubesleep.NewSuspendableNamespace("kube-node-lease", false),
	kubesleep.NewSuspendableNamespace("kube-public", false),
	kubesleep.NewSuspendableNamespace("kube-system", false),
}

func (s *Integrationtest) TestGetSuspendableNamespace() {
	deleteNamespace, err := testNamespace("get-suspendable-namespace", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	namespace, err := s.k8s.GetSuspendableNamespace("get-suspendable-namespace")
	s.Require().NoError(err)

	s.Require().Equal(
		kubesleep.NewSuspendableNamespace("get-suspendable-namespace", false),
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
		kubesleep.NewSuspendableNamespace("get-non-suspendable-namespace", true),
		namespace,
	)
}

func (s *Integrationtest) TestGetNamespace() {
	expected := append(STANDARD_NAMESPACES, kubesleep.NewSuspendableNamespace("get-namespaces", false))
	deleteNamespace, err := testNamespace("get-namespaces", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	namespaces, err := s.k8s.GetSuspendableNamespaces()
	s.Require().NoError(err)
	for _, e := range expected {
		s.Require().Contains(namespaces, e)
	}
}
