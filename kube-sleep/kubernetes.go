package kubesleep

import (
	"context"
	"fmt"
	"log/slog"

	v1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const STATE_FILE_NAME = "kubesleep-suspend-state"
const STATE_FILE_KEY = "kubesleep.json"

type k8simpl struct {
	clientset *kubernetes.Clientset
}

func NewK8simpl() (k8simpl, error) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
  	clientcmd.NewDefaultClientConfigLoadingRules(),
	  &clientcmd.ConfigOverrides{},
  )

	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return k8simpl{}, err
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return k8simpl{}, err
	}

	return k8simpl{clientset: clientset}, nil
}

func (k8s k8simpl) suspendableNamespace(namespace string) (suspendableNamespace, error) {
	kubernetesNamespace, err := k8s.clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	namespaceObj := suspendableNamespaceImpl{
		name:         namespace,
		_suspendable: true,
	}
	if kubernetesNamespace.ObjectMeta.Annotations != nil {
		_, found := kubernetesNamespace.ObjectMeta.Annotations["kubesleep.xyz/do-not-suspend"]
		namespaceObj._suspendable = !found
	} else {
		slog.Debug("Namespace has no annotations")
	}

	slog.Debug("namespace manifest", "kubernetesNamespace", kubernetesNamespace)
	slog.Info("parsed namespace", "namespace", namespaceObj)
	return &namespaceObj, err
}

func (k8s k8simpl) getDeployments(namespace string) ([]suspendable, error) {
	deployments, err := k8s.clientset.AppsV1().
		Deployments(namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var suspendables []suspendable

	for _, deployment := range deployments.Items {
    suspend := func() error {
      return k8s.suspendDeployment(&deployment)
    }

		s := suspendable{
			manifestType: "Deployment",
			name:         deployment.ObjectMeta.Name,
			replicas:     *deployment.Spec.Replicas,
      suspend: suspend,
		}
		slog.Debug("parsed suspendable", "suspendable", s)
		suspendables = append(suspendables, s)
	}

	return suspendables, nil
}

func (k8s k8simpl) suspendDeployment(deployment *appsv1.Deployment) error {
  replicas := int32(0)
  deployment.Spec.Replicas = &replicas
  _, err := k8s.clientset.AppsV1().Deployments(deployment.Namespace).Update(
    context.TODO(),
    deployment,
    metav1.UpdateOptions{},
  )
  if err != nil {
    return err
  }
  slog.Info("Suspended Deployment", "name", deployment.Name)
  return nil
}

func (k8s k8simpl)scaleDeployment(namespace string, name string, replicas int32)  error {
  deployment, err := k8s.clientset.AppsV1().Deployments(namespace).Get(
    context.TODO(),
    name,
    metav1.GetOptions{},
  )
  if err != nil {
    return err
  }
  deployment.Spec.Replicas = &replicas

  _, err = k8s.clientset.AppsV1().Deployments(deployment.Namespace).Update(
    context.TODO(),
    deployment,
    metav1.UpdateOptions{},
  )
  if err != nil {
    return err
  }

  slog.Info("Woke up Deployment", "name", name)
  return nil
}

func (k8s k8simpl) getStatefulsets(namespace string) ([]suspendable, error) {
	statefulSets, err := k8s.clientset.AppsV1().
		StatefulSets(namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var suspendables []suspendable

	for _, statefulSet := range statefulSets.Items {
    suspend := func() error {
      return k8s.suspendStatefulSet(&statefulSet)
    }

		s := suspendable{
			manifestType: "StatefulSet",
			name:         statefulSet.ObjectMeta.Name,
			replicas:     *statefulSet.Spec.Replicas,
      suspend: suspend,
		}
		slog.Debug("parsed suspendable", "suspendable", s)
		suspendables = append(suspendables, s)
	}

	return suspendables, nil
}

func (k8s k8simpl) suspendStatefulSet(statefulSet *appsv1.StatefulSet) error {
  replicas := int32(0)
  statefulSet.Spec.Replicas = &replicas
  _, err := k8s.clientset.AppsV1().StatefulSets(statefulSet.Namespace).Update(
    context.TODO(),
    statefulSet,
    metav1.UpdateOptions{},
  )
  if err != nil {
    return err
  }
  slog.Info("Suspended StatefulSet", "name", statefulSet.Name)
  return nil
}

func (k8s k8simpl)scaleStatefulSet(namespace string, name string, replicas int32)  error {
  statefulSet, err := k8s.clientset.AppsV1().StatefulSets(namespace).Get(
    context.TODO(),
    name,
    metav1.GetOptions{},
  )
  if err != nil {
    return err
  }
  statefulSet.Spec.Replicas = &replicas

  _, err = k8s.clientset.AppsV1().StatefulSets(statefulSet.Namespace).Update(
    context.TODO(),
    statefulSet,
    metav1.UpdateOptions{},
  )
  if err != nil {
    return err
  }

  slog.Info("Woke up StatefulSet", "name", name)
  return nil
}

func (k8s k8simpl) getStateFile(namespace string) (*suspendStateFile, error) {
  result, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Get(
		context.TODO(),
		STATE_FILE_NAME,
		metav1.GetOptions{},
	)
  if err != nil {
    return nil, err
  }
  return newSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
}

func (k8s k8simpl) createStateFile(namespace string, data *suspendStateFile) ( *suspendStateFile, error ) {
  // Create a new state file as a kubernetes configmap
  // Will return the current state of the configmap inside the cluster.
  result, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Get(
		context.TODO(),
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
    context.TODO(),
    configmap,
    metav1.CreateOptions{},
  )
  if err != nil {
    return nil, err
  }
  return newSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
}

func (k8s k8simpl) updateStateFile(namespace string, data *suspendStateFile) ( *suspendStateFile, error ) {
  // Write a state file to a kubernetes configmap
  // Will return the current state of the configmap.
  configmap, err := k8s.clientset.CoreV1().ConfigMaps(namespace).Get(
		context.TODO(),
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
    context.TODO(),
    configmap,
    metav1.UpdateOptions{},
  )
  return newSuspendStateFileFromJson(result.Data[STATE_FILE_KEY]), nil
}
