package kubesleep

import "github.com/stretchr/testify/mock"

// Test don't wake an unfinished namespace (aborted suspend) Including valid error statement.
// Optionally add a `--force` option to recover a half-suspended namespace.
func (s *Unittest) TestNamespaceWake() {
	k8s, _ := NewMockK8S()
	stateFile := TEST_SUSPEND_STATE_FILE
	stateFile.finished = true
	k8s.On("GetStateFile", "foo").Return(&stateFile, (*MockStateFileActions)(nil), nil)
	k8s.On("ScaleDeployment", "foo", "test-deployment", int32(2)).Return(errExpected)

	err := NewSuspendableNamespace("foo", true).wake(k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestNamespaceWakeFailsWhenSuspendInProgress() {
	k8s, _ := NewMockK8S()
	k8s.On("GetStateFile", "foo").Return(&TEST_SUSPEND_STATE_FILE, (*MockStateFileActions)(nil), nil)

	ns := &suspendableNamespaceImpl{name: "foo"}
	err := ns.wake(k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Error(err)
	s.Require().ErrorContains(err, "cannot wake the namespace")
}

func (s *Unittest) TestNamespaceSuspendStatefulSetError() {
	k8s, _ := NewMockK8S()
	k8s.On("GetDeployments", "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("GetStatefulSets", "foo").Return(map[string]Suspendable{}, errExpected)

	err := NewSuspendableNamespace("foo", true).suspend(k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestNamespaceSuspendCreateStatefileFailed() {
	k8s, _ := NewMockK8S()
	k8s.On("GetDeployments", "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("GetStatefulSets", "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("CreateStateFile", "foo", mock.Anything).Return((*MockStateFileActions)(nil), errExpected)

	err := NewSuspendableNamespace("foo", true).suspend(k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestNamespaceSuspend() {
	k8s, _ := NewMockK8S()
	actions := MockStateFileActions{}
	sus := TEST_SUSPENDABLE
	sus.Suspend = func() error { return nil }
	k8s.On("GetDeployments", "foo").Return(map[string]Suspendable{sus.Identifier(): sus}, nil)
	k8s.On("GetStatefulSets", "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("CreateStateFile", "foo", mock.Anything).Return(&actions, nil)
	actions.On("Update", mock.Anything).Return(nil)

	err := NewSuspendableNamespace("foo", true).suspend(k8s)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestNamespaceEnsureStateFileGetError() {
	k8s, _ := NewMockK8S()
	k8s.On("CreateStateFile", "foo", mock.Anything).Return((*MockStateFileActions)(nil), StatefileAlreadyExistsError("foobar"))
	k8s.On("GetStateFile", "foo").Return((*SuspendStateFile)(nil), (*MockStateFileActions)(nil), errExpected)

	stateFile := NewSuspendStateFile(map[string]Suspendable{}, false)
	namespace := &suspendableNamespaceImpl{"foo", true}
	_, _, err := namespace.ensureStateFile(k8s, &stateFile)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestNamespaceEnsureStateFile() {
	k8s, _ := NewMockK8S()
	existingStateFile := TEST_SUSPEND_STATE_FILE
	k8s.On("CreateStateFile", "foo", mock.Anything).Return((*MockStateFileActions)(nil), StatefileAlreadyExistsError("foobar"))
	k8s.On("GetStateFile", "foo").Return(&existingStateFile, (*MockStateFileActions)(nil), nil)

	stateFile := NewSuspendStateFile(map[string]Suspendable{}, false)
	namespace := &suspendableNamespaceImpl{"foo", true}
	_, _, err := namespace.ensureStateFile(k8s, &stateFile)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}
