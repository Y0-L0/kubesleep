package k8s

import (
	"maps"
	"slices"

	kubesleep "github.com/Y0-L0/kubesleep/kube-sleep"
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
	deleteNamespace, err := testNamespace("create-delete-statefulset", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "create-delete-statefulset", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetStatefulSet() {
	deleteNamespace, err := testNamespace("get-statefulsets", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "get-statefulsets", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.GetStatefulSets("get-statefulsets")
	s.Require().NoError(err)

	s.Require().Equal([]string{"StatefulSettest-statefulset"}, slices.Collect(maps.Keys(sus)))

	// simplify for easier comparison
	actual := sus["StatefulSettest-statefulset"]
	actual.suspend = nil
	s.Require().Equal(
		kubesleep.NewSuspendable(
			"StatefulSet",
			"test-statefulset",
			int32(2),
			nil,
		),
		actual,
	)
}

func (s *Integrationtest) TestSuspendStatefulSet() {
	deleteNamespace, err := testNamespace("suspend-statefulsets", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "suspend-statefulsets", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.GetStatefulSets("suspend-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(2), sus["StatefulSettest-statefulset"].Replicas)

	err = sus["StatefulSettest-statefulset"].Suspend()
	s.Require().NoError(err)

	sus, err = s.k8s.GetStatefulSets("suspend-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(0), sus["StatefulSettest-statefulset"].Replicas)
}

func (s *Integrationtest) TestScaleStatefulSet() {
	deleteNamespace, err := testNamespace("scale-statefulsets", s.k8s)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(*s.k8s, "scale-statefulsets", "test-statefulset", int32(0))
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.ScaleStatefulSet("scale-statefulsets", "test-statefulset", int32(2))

	sus, err := s.k8s.GetStatefulSets("scale-statefulsets")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(2), sus["StatefulSettest-statefulset"].Replicas)
}
