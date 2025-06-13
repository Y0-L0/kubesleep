package kubesleep

import (
	"errors"

	"github.com/stretchr/testify/mock"
)

var mockK8SError = errors.New("mockK8S error")

type mockK8S struct{ mock.Mock }

func (m *mockK8S) GetSuspendableNamespace(string) (SuspendableNamespace, error) {
	return nil, mockK8SError
}

func (m *mockK8S) GetDeployments(string) (map[string]Suspendable, error) {
	return nil, mockK8SError
}
func (m *mockK8S) ScaleDeployment(string, string, int32) error {
	return mockK8SError
}

func (m *mockK8S) GetStatefulSets(string) (map[string]Suspendable, error) {
	return nil, mockK8SError
}
func (m *mockK8S) ScaleStatefulSet(string, string, int32) error {
	return mockK8SError
}

func (m *mockK8S) GetStateFile(string) (*SuspendStateFile, error) {
	return nil, mockK8SError
}

func (m *mockK8S) CreateStateFile(string, *SuspendStateFile) (*SuspendStateFile, error) {
	return nil, mockK8SError
}

func (m *mockK8S) UpdateStateFile(string, *SuspendStateFile) (*SuspendStateFile, error) {
	return nil, mockK8SError
}

func (m *mockK8S) DeleteStateFile(string) error {
	return mockK8SError
}

func mockK8SFactory() (K8S, error) {
	k8s := mockK8S{}
	return &k8s, nil
}
