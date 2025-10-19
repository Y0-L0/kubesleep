package kubesleep

import (
	"context"
	"log/slog"

	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test don't wake an unfinished namespace (aborted suspend) Including valid error statement.
// Optionally add a `--force` option to recover a half-suspended namespace.
func (s *Unittest) TestNamespaceWake() {
	k8s, _ := NewMockK8S()
	stateFile := TEST_SUSPEND_STATE_FILE
	stateFile.finished = true
	k8s.On("GetStateFile", mock.Anything, "foo").Return(&stateFile, (*MockStateFileActions)(nil), nil)
	k8s.On("ScaleSuspendable", mock.Anything, "foo", mock.Anything, mock.Anything, mock.Anything).Return(errExpected)

	err := NewSuspendableNamespace("foo", true).wake(context.TODO(), k8s)

	k8s.AssertExpectations(s.T())
	s.Require().ErrorIs(err, errExpected)
}

func (s *Unittest) TestNamespaceWakeFailsWhenSuspendInProgress() {
	k8s, _ := NewMockK8S()
	k8s.On("GetStateFile", mock.Anything, "foo").Return(&TEST_SUSPEND_STATE_FILE, (*MockStateFileActions)(nil), nil)

	ns := &suspendableNamespaceImpl{name: "foo"}
	err := ns.wake(context.TODO(), k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Error(err)
	s.Require().ErrorContains(err, "cannot wake the namespace")
}

func (s *Unittest) TestNamespaceSuspendStatefulSetError() {
	k8s, _ := NewMockK8S()
	k8s.On("GetSuspendables", mock.Anything, "foo").Return(map[string]Suspendable{}, errExpected)

	err := NewSuspendableNamespace("foo", true).suspend(context.TODO(), k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestNamespaceSuspendCreateStatefileFailed() {
	k8s, _ := NewMockK8S()
	k8s.On("GetSuspendables", mock.Anything, "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("CreateStateFile", mock.Anything, "foo", mock.Anything).Return((*MockStateFileActions)(nil), errExpected)

	err := NewSuspendableNamespace("foo", true).suspend(context.TODO(), k8s)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestSuspendConflict() {
	k8s, _ := NewMockK8S()
	actions := MockStateFileActions{}
	conflictErr := &apierrors.StatusError{
		ErrStatus: metav1.Status{Reason: metav1.StatusReasonConflict},
	}
	sus := TEST_SUSPENDABLE
	sus.Suspend = func(context.Context) error {
		slog.Debug("Mock suspend returning conflictErr")
		return conflictErr
	}
	k8s.On("GetSuspendables", mock.Anything, "foo").Return(map[string]Suspendable{sus.Identifier(): sus}, nil)
	k8s.On("CreateStateFile", mock.Anything, "foo", mock.Anything).Return(&actions, nil)

	err := NewSuspendableNamespace("foo", true).suspend(context.TODO(), k8s)

	k8s.AssertExpectations(s.T())
	s.Require().ErrorIs(err, conflictErr)
	s.Require().Contains(err.Error(), "operation failed after")
}

func (s *Unittest) TestNamespaceSuspend() {
	k8s, _ := NewMockK8S()
	actions := MockStateFileActions{}
	sus := TEST_SUSPENDABLE
	sus.Suspend = func(context.Context) error { return nil }
	k8s.On("GetSuspendables", mock.Anything, "foo").Return(map[string]Suspendable{sus.Identifier(): sus}, nil)
	k8s.On("CreateStateFile", mock.Anything, "foo", mock.Anything).Return(&actions, nil)
	actions.On("Update", mock.Anything, mock.Anything).Return(nil)

	err := NewSuspendableNamespace("foo", true).suspend(context.TODO(), k8s)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestNamespaceEnsureStateFileGetError() {
	k8s, _ := NewMockK8S()
	k8s.On("CreateStateFile", mock.Anything, "foo", mock.Anything).Return((*MockStateFileActions)(nil), StatefileAlreadyExistsError("foobar"))
	k8s.On("GetStateFile", mock.Anything, "foo").Return((*SuspendState)(nil), (*MockStateFileActions)(nil), errExpected)

	stateFile := NewSuspendState(map[string]Suspendable{}, false)
	namespace := &suspendableNamespaceImpl{"foo", true}
	_, _, err := namespace.ensureStateFile(context.TODO(), k8s, &stateFile)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestNamespaceEnsureStateFile() {
	k8s, _ := NewMockK8S()
	existingStateFile := TEST_SUSPEND_STATE_FILE
	k8s.On("CreateStateFile", mock.Anything, "foo", mock.Anything).Return((*MockStateFileActions)(nil), StatefileAlreadyExistsError("foobar"))
	k8s.On("GetStateFile", mock.Anything, "foo").Return(&existingStateFile, (*MockStateFileActions)(nil), nil)

	stateFile := NewSuspendState(map[string]Suspendable{}, false)
	namespace := &suspendableNamespaceImpl{"foo", true}
	_, _, err := namespace.ensureStateFile(context.TODO(), k8s, &stateFile)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}
