package kubesleep

import (
	"github.com/stretchr/testify/mock"
)

var brokenK8SFactory = func() (K8S, error) { return nil, errExpected }

var placeholderK8S = func() (K8S, error) { return nil, nil }

func (s *Unittest) TestSuspendBrokenK8SFactory() {
	err := cliConfig{namespaces: []string{"foo"}}.suspend(brokenK8SFactory)

	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestSuspendBrokenK8S() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", true), nil)
	k8s.On("GetDeployments", "foo").Return(map[string]Suspendable{}, errExpected)

	err := cliConfig{namespaces: []string{"foo"}}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestSuspendSkip() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", false), nil)

	err := cliConfig{namespaces: []string{"foo"}}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestSuspendEmptyNamespace() {
	k8s, factory := NewMockK8S()
	actions := MockStateFileActions{}
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", true), nil)
	k8s.On("GetDeployments", "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("GetStatefulSets", "foo").Return(map[string]Suspendable{}, nil)
	k8s.On("CreateStateFile", "foo", mock.Anything).Return(&actions, nil)
	actions.On("Update", mock.Anything).Return(nil)

	err := cliConfig{namespaces: []string{"foo"}}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestSuspendAllNamespacesError() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespaces").Return([]SuspendableNamespace{}, errExpected)

	err := cliConfig{allNamespaces: true}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(err, errExpected)
}

func (s *Unittest) TestSuspendAllNamespaces() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespaces").Return([]SuspendableNamespace{NewSuspendableNamespace("bar", false)}, nil)

	err := cliConfig{allNamespaces: true}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestDontSuspendAutoprotected() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespaces").Return([]SuspendableNamespace{NewSuspendableNamespace("kube-system", true)}, nil)

	err := cliConfig{allNamespaces: true}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}

func (s *Unittest) TestSuspendConfigCollision() {
	s.Require().Panics(func() {
		_ = cliConfig{allNamespaces: true, force: true}.suspend(placeholderK8S)
	})
}

func (s *Unittest) TestSuspendNoNamespace() {
	s.Require().Panics(func() {
		_ = cliConfig{}.suspend(placeholderK8S)
	})
}

func (s *Unittest) TestSuspendEmptyNamespaceList() {
	s.Require().Panics(func() {
		_ = cliConfig{namespaces: []string{}}.suspend(placeholderK8S)
	})
}

func (s *Unittest) TestWakeNoNamespace() {
	s.Require().Panics(func() {
		_ = cliConfig{}.wake(placeholderK8S)
	})
}

func (s *Unittest) TestWakeInvalidNamespace() {
	s.Require().Panics(func() {
		_ = cliConfig{namespaces: []string{""}}.wake(placeholderK8S)
	})
}

func (s *Unittest) TestWakeBrokenK8SFactory() {
	err := cliConfig{namespaces: []string{"foo"}}.wake(brokenK8SFactory)

	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestWakeBrokenK8S() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", true), nil)
	k8s.On("GetStateFile", "foo").Return((*SuspendStateFile)(nil), (*MockStateFileActions)(nil), errExpected)

	err := cliConfig{namespaces: []string{"foo"}}.wake(factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestWakeEmptyNamespace() {
	k8s, factory := NewMockK8S()
	actions := MockStateFileActions{}
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", true), nil)
	k8s.On("GetStateFile", "foo").Return(&SuspendStateFile{finished: true}, &actions, nil)
	actions.On("Delete").Return(nil)

	err := cliConfig{namespaces: []string{"foo"}}.wake(factory)

	k8s.AssertExpectations(s.T())
	s.Require().NoError(err)
}
