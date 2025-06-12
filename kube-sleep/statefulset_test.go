package kubesleep

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createStatefulSet(k8s k8simpl, namespace string, name string, replicas int32) (func() error, error) {
	labels := map[string]string{"app": name}

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "k8s.gcr.io/pause:3.9",
						},
					},
				},
			},
		},
	}

	_, err := k8s.clientset.AppsV1().StatefulSets(namespace).Create(k8s.ctx, statefulset, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	delete := func() error {
		return k8s.clientset.AppsV1().StatefulSets(namespace).Delete(k8s.ctx, name, metav1.DeleteOptions{})
	}
	return delete, nil
}

func (s *Integrationtest) TestCreateDeleteStatefulSet() {
	deleteNamespace, err := testNamespace("create-delete-statefulset", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createStatefulSet(*s.k8s, "create-delete-statefulset", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetStatefulSet() {
	deleteNamespace, err := testNamespace("get-statefulsets", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createStatefulSet(*s.k8s, "get-statefulsets", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.getStatefulSets("get-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)

	// simplify for easier comparison
	sus[0].suspend = nil
	s.Require().Equal(
		[]suspendable{suspendable{
			manifestType: "StatefulSet",
			name:         "test-statefulset",
			replicas:     int32(2),
		}},
		sus,
	)
}

func (s *Integrationtest) TestSuspendStatefulSet() {
	deleteNamespace, err := testNamespace("suspend-statefulsets", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createStatefulSet(*s.k8s, "suspend-statefulsets", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.getStatefulSets("suspend-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(2), sus[0].replicas)

	err = sus[0].suspend()
	s.Require().NoError(err)

	sus, err = s.k8s.getStatefulSets("suspend-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(0), sus[0].replicas)
}

func (s *Integrationtest) TestScaleStatefulSet() {
	deleteNamespace, err := testNamespace("scale-statefulsets", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := createStatefulSet(*s.k8s, "scale-statefulsets", "test-statefulset", int32(0))
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.scaleStatefulSet("scale-statefulsets", "test-statefulset", int32(2))

	sus, err := s.k8s.getStatefulSets("scale-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(2), sus[0].replicas)
}
