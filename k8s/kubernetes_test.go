package k8s

func (s *Integrationtest) TestDeploymentStatefulsetNameConflict() {
	deleteNamespace, err := testNamespace("deployment-statefulset-name-conflict", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	deleteStatefulSet, err := CreateStatefulSet(*s.k8s, "deployment-statefulset-name-conflict", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer deleteStatefulSet()

	deleteDeployment, err := CreateDeployment(*s.k8s, "deployment-statefulset-name-conflict", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer deleteDeployment()
}
