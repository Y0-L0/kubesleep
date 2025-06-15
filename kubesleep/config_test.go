package kubesleep

var brokenK8SFactory = func() (K8S, error) { return nil, errExpected }

func (s *Unittest) TestSuspendBrokenK8SFactory() {
	err := cliConfig{}.suspend(brokenK8SFactory)

	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestSuspendBrokenK8S() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", true), nil)
	k8s.On("GetDeployments", "foo").Return(map[string]Suspendable{}, errExpected)

	err := cliConfig{namespace: "foo"}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestSuspendSkip() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", false), nil)

	err := cliConfig{namespace: "foo"}.suspend(factory)

	k8s.AssertExpectations(s.T())
	s.Require().Nil(err)
}

func (s *Unittest) TestWakeBrokenK8SFactory() {
	err := cliConfig{namespace: "foo"}.wake(brokenK8SFactory)

	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestWakeInvalidNamespace() {
	s.Require().Panics(func() {
		_ = cliConfig{}.wake(nil)
	})
}

func (s *Unittest) TestWakeBrokenK8S() {
	k8s, factory := NewMockK8S()
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", true), nil)
	k8s.On("GetStateFile", "foo").Return((*SuspendStateFile)(nil), (*MockStateFileActions)(nil), errExpected)

	err := cliConfig{namespace: "foo"}.wake(factory)

	k8s.AssertExpectations(s.T())
	s.Require().Equal(errExpected, err)
}

func (s *Unittest) TestWakeEmptyNamespace() {
	k8s, factory := NewMockK8S()
	actions := MockStateFileActions{}
	k8s.On("GetSuspendableNamespace", "foo").Return(NewSuspendableNamespace("foo", true), nil)
	k8s.On("GetStateFile", "foo").Return(&SuspendStateFile{finished: true}, &actions, nil)
	actions.On("Delete").Return(nil)

	err := cliConfig{namespace: "foo"}.wake(factory)

	k8s.AssertExpectations(s.T())
	s.Require().Nil(err)
}
