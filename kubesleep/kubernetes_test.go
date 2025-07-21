package kubesleep

import (
	"errors"

	"github.com/stretchr/testify/mock"
)

var errExpected = errors.New("broken k8s factory")

type mockK8S struct{ mock.Mock }

func (m *mockK8S) GetNamespaces() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockK8S) GetSuspendableNamespace(ns string) (SuspendableNamespace, error) {
	args := m.Called(ns)
	return args.Get(0).(SuspendableNamespace), args.Error(1)
}

func (m *mockK8S) GetDeployments(ns string) (map[string]Suspendable, error) {
	args := m.Called(ns)
	return args.Get(0).(map[string]Suspendable), args.Error(1)
}

func (m *mockK8S) ScaleDeployment(ns, name string, replicas int32) error {
	args := m.Called(ns, name, replicas)
	return args.Error(0)
}

func (m *mockK8S) GetStatefulSets(ns string) (map[string]Suspendable, error) {
	args := m.Called(ns)
	return args.Get(0).(map[string]Suspendable), args.Error(1)
}

func (m *mockK8S) ScaleStatefulSet(ns, name string, replicas int32) error {
	args := m.Called(ns, name, replicas)
	return args.Error(0)
}

func (m *mockK8S) GetStateFile(ns string) (*SuspendStateFile, StateFileActions, error) {
	args := m.Called(ns)
	return args.Get(0).(*SuspendStateFile), args.Get(1).(StateFileActions), args.Error(2)
}

func (m *mockK8S) CreateStateFile(ns string, file *SuspendStateFile) (StateFileActions, error) {
	args := m.Called(ns, file)
	return args.Get(0).(StateFileActions), args.Error(1)
}

func (m *mockK8S) UpdateStateFile(ns string, file *SuspendStateFile) (*SuspendStateFile, error) {
	args := m.Called(ns, file)
	return args.Get(0).(*SuspendStateFile), args.Error(1)
}

func (m *mockK8S) DeleteStateFile(ns string) error {
	args := m.Called(ns)
	return args.Error(0)
}

func NewMockK8S() (*mockK8S, K8SFactory) {
	k8s := &mockK8S{}
	return k8s, func() (K8S, error) { return k8s, nil }
}

// func mockK8SFactory() (K8S, error) {
// 	k8s := mockK8S{}
// 	return &k8s, nil
// }
