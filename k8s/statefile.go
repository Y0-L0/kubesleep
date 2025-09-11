package k8s

import (
	"fmt"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const STATE_FILE_NAME = "kubesleep-suspend-state"

type StateFileActionsImpl struct {
	k8s       *K8Simpl
	configmap *corev1.ConfigMap
}

func (s *StateFileActionsImpl) Update(data map[string]string) error {
	var err error
	s.configmap.Data = data
	s.configmap, err = s.k8s.clientset.CoreV1().ConfigMaps(s.configmap.ObjectMeta.Namespace).Update(
		s.k8s.ctx,
		s.configmap,
		metav1.UpdateOptions{},
	)
	return err
}

func (s *StateFileActionsImpl) Delete() error {
	return s.k8s.clientset.CoreV1().ConfigMaps(s.configmap.ObjectMeta.Namespace).Delete(
		s.k8s.ctx,
		s.configmap.Name,
		metav1.DeleteOptions{},
	)
}

func (k8s *K8Simpl) GetStateFile(namespace string) (*kubesleep.SuspendStateFile, kubesleep.StateFileActions, error) {
	configmap, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Get(
		k8s.ctx,
		STATE_FILE_NAME,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, nil, err
	}
	return kubesleep.ReadSuspendState(configmap.Data), &StateFileActionsImpl{k8s, configmap}, nil
}

func (k8s *K8Simpl) CreateStateFile(namespace string, data map[string]string) (kubesleep.StateFileActions, error) {

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      STATE_FILE_NAME,
			Namespace: namespace,
		},
		Data: data,
	}

	configmap, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Create(
		k8s.ctx,
		configmap,
		metav1.CreateOptions{},
	)
	if apierrors.IsAlreadyExists(err) {
		return nil, kubesleep.StatefileAlreadyExistsError(
			fmt.Sprintf("statefile configmap %s already exists indicating an in-progress or aborted suspend operation.", STATE_FILE_NAME),
		)
	}
	if err != nil {
		return nil, err
	}

	return &StateFileActionsImpl{k8s, configmap}, nil
}

func (k8s *K8Simpl) DeleteStateFile(namespace string) error {
	return k8s.clientset.CoreV1().ConfigMaps(namespace).Delete(k8s.ctx, STATE_FILE_NAME, metav1.DeleteOptions{})
}
