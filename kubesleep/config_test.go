package kubesleep

import (
	"context"
	"github.com/stretchr/testify/mock"
	"io"
)

var brokenK8SFactory = func() (K8S, error) { return nil, errExpected }

var placeholderK8S = func() (K8S, error) { return nil, nil }

func (s *Unittest) TestSuspendBrokenK8SFactory() {
	err := cliConfig{namespaces: []string{"foo"}, outWriter: io.Discard}.suspend(context.TODO(), brokenK8SFactory)

	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestSuspendBrokenK8S() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", mock.Anything, "foo").Return(NewSuspendableNamespace("foo", false), nil)
	k8s.On("GetSuspendables", mock.Anything, "foo").Return(map[string]Suspendable{}, errExpected)

	err := cliConfig{namespaces: []string{"foo"}, outWriter: io.Discard}.suspend(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestSuspendSkip() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", mock.Anything, "foo").Return(NewSuspendableNamespace("foo", true), nil)

	err := cliConfig{namespaces: []string{"foo"}, outWriter: io.Discard}.suspend(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestSuspendEmptyNamespace() {
	k8s, factory := NewMockK8S()
	actions := MockStateFileActions{}
	k8s.On("GetSuspendableNamespace", mock.Anything, "foo").Return(NewSuspendableNamespace("foo", false), nil)
	k8s.On("GetSuspendables", mock.Anything, "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("CreateStateFile", mock.Anything, "foo", mock.Anything).Return(&actions, nil)
	actions.On("Update", mock.Anything, mock.Anything).Return(nil)

	err := cliConfig{namespaces: []string{"foo"}, outWriter: io.Discard}.suspend(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestSuspendAllNamespacesError() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespaces", mock.Anything).Return([]SuspendableNamespace{}, errExpected)

	err := cliConfig{allNamespaces: true, outWriter: io.Discard}.suspend(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(err, errExpected)
}

func (s *Unittest) TestSuspendAllNamespaces() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespaces", mock.Anything).Return([]SuspendableNamespace{NewSuspendableNamespace("bar", true)}, nil)

	err := cliConfig{allNamespaces: true, outWriter: io.Discard}.suspend(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestDontSuspendAutoprotected() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespaces", mock.Anything).Return([]SuspendableNamespace{NewSuspendableNamespace("kube-system", false)}, nil)

	err := cliConfig{allNamespaces: true, outWriter: io.Discard}.suspend(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestSuspendConfigCollision() {
	s.Require().Panics(func() {
		_ = cliConfig{allNamespaces: true, force: true, outWriter: io.Discard}.suspend(context.TODO(), placeholderK8S)
	})
}

func (s *Unittest) TestSuspendNoNamespace() {
	s.Require().Panics(func() {
		_ = cliConfig{outWriter: io.Discard}.suspend(context.TODO(), placeholderK8S)
	})
}

func (s *Unittest) TestSuspendEmptyNamespaceList() {
	s.Require().Panics(func() {
		_ = cliConfig{namespaces: []string{}, outWriter: io.Discard}.suspend(context.TODO(), placeholderK8S)
	})
}

func (s *Unittest) TestWakeNoNamespace() {
	s.Require().Panics(func() {
		_ = cliConfig{outWriter: io.Discard}.wake(context.TODO(), placeholderK8S)
	})
}

func (s *Unittest) TestWakeInvalidNamespace() {
	s.Require().Panics(func() {
		_ = cliConfig{namespaces: []string{""}, outWriter: io.Discard}.wake(context.TODO(), placeholderK8S)
	})
}

func (s *Unittest) TestWakeBrokenK8SFactory() {
	err := cliConfig{namespaces: []string{"foo"}, outWriter: io.Discard}.wake(context.TODO(), brokenK8SFactory)

	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestWakeBrokenK8S() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", mock.Anything, "foo").Return(NewSuspendableNamespace("foo", false), nil)
	k8s.On("GetStateFile", mock.Anything, "foo").Return((*SuspendState)(nil), (*MockStateFileActions)(nil), errExpected)

	err := cliConfig{namespaces: []string{"foo"}, outWriter: io.Discard}.wake(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestWakeEmptyNamespace() {
	k8s, factory := NewMockK8S()
	actions := MockStateFileActions{}
	k8s.On("GetSuspendableNamespace", mock.Anything, "foo").Return(NewSuspendableNamespace("foo", false), nil)
	k8s.On("GetStateFile", mock.Anything, "foo").Return(&SuspendState{finished: true}, &actions, nil)
	actions.On("Delete", mock.Anything).Return(nil)

	err := cliConfig{namespaces: []string{"foo"}, outWriter: io.Discard}.wake(context.TODO(), factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}
