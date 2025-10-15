package kubesleep

import (
	"context"
	"github.com/stretchr/testify/mock"
)

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
	sus := NewSuspendable(StatefulSet, "test-statefulset", int32(2), nil)
	k8s.On("ScaleSuspendable", mock.Anything, "foo", StatefulSet, "test-statefulset", int32(2)).Return(errExpected)

	err := sus.wake(context.TODO(), "foo", k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Error(err)
}

func (s *Unittest) TestScaleStatefulSet() {
	k8s, _ := NewMockK8S()
	sus := NewSuspendable(StatefulSet, "test-statefulset", int32(2), nil)
	k8s.On("ScaleSuspendable", mock.Anything, "foo", StatefulSet, "test-statefulset", int32(2)).Return(nil)

	err := sus.wake(context.TODO(), "foo", k8s)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}
