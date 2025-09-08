package k8s

import (
	"maps"
	"slices"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateStatefulSet(k8s K8Simpl, namespace string, name string, Replicas int32) (func() error, error) {
	labels := map[string]string{"app": name}

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &Replicas,
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
	deleteNamespace, err := testNamespace("create-delete-statefulset", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "create-delete-statefulset", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetStatefulSet() {
	deleteNamespace, err := testNamespace("get-statefulsets", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "get-statefulsets", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	suspendables, err := s.k8s.GetSuspendables("get-statefulsets")
	s.Require().NoError(err)

	s.Require().Equal([]string{"1:test-statefulset"}, slices.Collect(maps.Keys(suspendables)))

	// simplify for easier comparison
	actual := suspendables["1:test-statefulset"]
	actual.Suspend = nil
	s.Require().Equal(
		kubesleep.NewSuspendable(
			kubesleep.StatefulSet,
			"test-statefulset",
			int32(2),
			nil,
		),
		actual,
	)
}

func (s *Integrationtest) TestSuspendStatefulSet() {
	deleteNamespace, err := testNamespace("suspend-statefulsets", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "suspend-statefulsets", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	suspendables, err := s.k8s.GetSuspendables("suspend-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(suspendables)
	s.Require().Equal(int32(2), suspendables["1:test-statefulset"].Replicas)

	err = s.k8s.ScaleSuspendable("suspend-statefulsets", kubesleep.StatefulSet, "test-statefulset", 0)
	s.Require().NoError(err)

	suspendables, err = s.k8s.GetSuspendables("suspend-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(suspendables)
	s.Require().Equal(int32(0), suspendables["1:test-statefulset"].Replicas)
}

func (s *Integrationtest) TestScaleStatefulSet() {
	deleteNamespace, err := testNamespace("scale-statefulsets", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "scale-statefulsets", "test-statefulset", int32(0))
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.ScaleSuspendable("scale-statefulsets", kubesleep.StatefulSet, "test-statefulset", int32(2))

	suspendables, err := s.k8s.GetSuspendables("scale-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(suspendables)
	s.Require().Equal(int32(2), suspendables["1:test-statefulset"].Replicas)
}
