package k8s

import (
	"context"
	"maps"
	"slices"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateStatefulSet(ctx context.Context, k8s K8Simpl, namespace string, name string, Replicas int32) (func() error, error) {
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

	_, err := k8s.clientset.AppsV1().StatefulSets(namespace).Create(ctx, statefulset, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	delete := func() error {
		return k8s.clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	}
	return delete, nil
}

func (s *Integrationtest) TestCreateDeleteStatefulSet() {
	deleteNamespace, err := testNamespace(s.ctx, "create-delete-statefulset", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(s.ctx, *s.k8s, "create-delete-statefulset", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetStatefulSet() {
	deleteNamespace, err := testNamespace(s.ctx, "get-statefulsets", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(s.ctx, *s.k8s, "get-statefulsets", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	suspendables, err := s.k8s.GetSuspendables(s.ctx, "get-statefulsets")
	s.Require().NoError(err)

	s.Require().Equal([]string{"1:test-statefulset"}, slices.Collect(maps.Keys(suspendables)))

	actual := s.getSuspendable("get-statefulsets", "1:test-statefulset")
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
	deleteNamespace, err := testNamespace(s.ctx, "suspend-statefulsets-via-suspendable", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(s.ctx, *s.k8s, "suspend-statefulsets-via-suspendable", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer delete()

	before := s.getSuspendable("suspend-statefulsets-via-suspendable", "1:test-statefulset")
	s.Require().Equal(int32(2), before.Replicas)

	s.Require().NoError(before.Suspend())

	actual := s.getSuspendable("suspend-statefulsets-via-suspendable", "1:test-statefulset")
	s.Require().Equal(int32(0), actual.Replicas)
}

func (s *Integrationtest) TestAlreadySuspendedStatefulSet() {
	deleteNamespace, err := testNamespace(s.ctx, "skip-already-suspended-statefulset", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(s.ctx, *s.k8s, "skip-already-suspended-statefulset", "test-statefulset", int32(0))
	s.Require().NoError(err)
	defer delete()

	beforeScale, err := s.k8s.clientset.AppsV1().StatefulSets("skip-already-suspended-statefulset").GetScale(s.ctx, "test-statefulset", metav1.GetOptions{})
	s.Require().NoError(err)

	before := s.getSuspendable("skip-already-suspended-statefulset", "1:test-statefulset")
	s.Require().Equal(int32(0), before.Replicas)

	s.Require().NoError(before.Suspend())

	afterScale, err := s.k8s.clientset.AppsV1().StatefulSets("skip-already-suspended-statefulset").GetScale(s.ctx, "test-statefulset", metav1.GetOptions{})
	s.Require().NoError(err)
	s.Require().Equal(beforeScale.ResourceVersion, afterScale.ResourceVersion)
}

func (s *Integrationtest) TestScaleStatefulSet() {
	deleteNamespace, err := testNamespace(s.ctx, "scale-statefulsets", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateStatefulSet(s.ctx, *s.k8s, "scale-statefulsets", "test-statefulset", int32(0))
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.ScaleSuspendable(s.ctx, "scale-statefulsets", kubesleep.StatefulSet, "test-statefulset", int32(2))

	actual := s.getSuspendable("scale-statefulsets", "1:test-statefulset")
	s.Require().Equal(int32(2), actual.Replicas)
}
