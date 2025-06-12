package kubesleep

import (
	"fmt"
	"log/slog"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const STATE_FILE_NAME = "kubesleep-suspend-state"
const STATE_FILE_KEY = "kubesleep.json"

func (k8s k8simpl) getStateFile(namespace string) (*suspendStateFile, error) {
	result, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Get(
		k8s.ctx,
		STATE_FILE_NAME,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, err
	}
	return newSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
}

func (k8s *k8simpl) createStateFile(namespace string, data *suspendStateFile) (*suspendStateFile, error) {
	// Create a new state file as a kubernetes configmap
	// Will return the current state of the configmap inside the cluster.
	result, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Get(
		k8s.ctx,
		STATE_FILE_NAME,
		metav1.GetOptions{},
	)
	if err == nil {
		slog.Info("statefile configmap already exists indicating a previously aborted suspend operation.", "configmap name", STATE_FILE_NAME)
		return newSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("Problem when reading configmap with name: %s error: %w", STATE_FILE_NAME, err)
	}

	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      STATE_FILE_NAME,
			Namespace: namespace,
		},
		Data: map[string]string{
			STATE_FILE_KEY: data.toJson(),
		},
	}

	result, err = k8s.clientset.CoreV1().ConfigMaps(namespace).Create(
		k8s.ctx,
		configmap,
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, err
	}
	return newSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
}

func (k8s *k8simpl) updateStateFile(namespace string, data *suspendStateFile) (*suspendStateFile, error) {
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

	configmap.Data[STATE_FILE_KEY] = data.toJson()
	result, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Update(
		k8s.ctx,
		configmap,
		metav1.UpdateOptions{},
	)
	return newSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
}

func (k8s *k8simpl) deleteStateFile(namespace string) error {
	return k8s.clientset.CoreV1().ConfigMaps(namespace).Delete(k8s.ctx, STATE_FILE_NAME, metav1.DeleteOptions{})
}
