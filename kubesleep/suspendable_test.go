package kubesleep

var TEST_SUSPENDABLE = Suspendable{
	manifestType: Deplyoment,
	name:         "test-deployment",
	Replicas:     int32(2),
}

var TEST_SUSPENDABLES = map[string]Suspendable{
	TEST_SUSPENDABLE.Identifier(): TEST_SUSPENDABLE,
}

func (s *Unittest) TestScaleStatefulSetBrokenK8S() {
	k8s, _ := NewMockK8S()
	k8s.On("ScaleStatefulSet", "foo", "test-statefulset", int32(2)).Return(errExpected)

	err := NewSuspendable(
		StatefulSet,
		"test-statefulset",
		int32(2),
		nil,
	).wake("foo", k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Error(err)
}

func (s *Unittest) TestScaleStatefulSet() {
	k8s, _ := NewMockK8S()
	k8s.On("ScaleStatefulSet", "foo", "test-statefulset", int32(2)).Return(nil)

	err := NewSuspendable(
		StatefulSet,
		"test-statefulset",
		int32(2),
		nil,
	).wake("foo", k8s)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}
