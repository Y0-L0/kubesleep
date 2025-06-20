package k8s

import (
	"fmt"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const STATE_FILE_NAME = "kubesleep-suspend-state"
const STATE_FILE_KEY = "kubesleep.json"

type StateFileActionsImpl struct {
	k8s       *K8Simpl
	configmap *corev1.ConfigMap
}

func (s *StateFileActionsImpl) Update(data *kubesleep.SuspendStateFile) error {
	var err error
	s.configmap.Data[STATE_FILE_KEY] = data.ToJson()
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
	return kubesleep.NewSuspendStateFileFromJson(configmap.Data[STATE_FILE_KEY]), &StateFileActionsImpl{k8s, configmap}, nil
}

func (k8s *K8Simpl) CreateStateFile(namespace string, data *kubesleep.SuspendStateFile) (kubesleep.StateFileActions, error) {
	// Create a new state file as a kubernetes configmap

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      STATE_FILE_NAME,
			Namespace: namespace,
		},
		Data: map[string]string{
			STATE_FILE_KEY: data.ToJson(),
		},
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

func (k8s *K8Simpl) UpdateStateFile(namespace string, data *kubesleep.SuspendStateFile) (*kubesleep.SuspendStateFile, error) {
	// Write a state file to a kubernetes configmap
	// Will return the current state of the configmap.
	configmap, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Get(
		k8s.ctx,
		STATE_FILE_NAME,
		metav1.GetOptions{},
	)

	if apierrors.IsNotFound(err) {
		panic(fmt.Errorf("Updating an existing statefile only works if it has been created before, %w", err))
	}
	if err != nil {
		return nil, err
	}

	configmap.Data[STATE_FILE_KEY] = data.ToJson()
	result, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Update(
		k8s.ctx,
		configmap,
		metav1.UpdateOptions{},
	)
	return kubesleep.NewSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
}

func (k8s *K8Simpl) DeleteStateFile(namespace string) error {
	return k8s.clientset.CoreV1().ConfigMaps(namespace).Delete(k8s.ctx, STATE_FILE_NAME, metav1.DeleteOptions{})
}
