package k8s

import (
	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateCronJob(k8s K8Simpl, namespace string, name string, suspended bool) (func() error, error) {
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "*/1 * * * *",
			Suspend:  func(b bool) *bool { v := b; return &v }(suspended),
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  name,
									Image: "k8s.gcr.io/pause:3.9",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := k8s.clientset.BatchV1().CronJobs(namespace).Create(k8s.ctx, cj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	delete := func() error {
		return k8s.clientset.BatchV1().CronJobs(namespace).Delete(k8s.ctx, name, metav1.DeleteOptions{})
	}
	return delete, nil
}

func (s *Integrationtest) TestCreateDeleteCronJob() {
	deleteNamespace, err := testNamespace("create-delete-cronjob", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateCronJob(*s.k8s, "create-delete-cronjob", "test-cronjob", false)
	s.Require().NoError(err)
	defer delete()
}

func (s *Integrationtest) TestGetCronJobs() {
	deleteNamespace, err := testNamespace("get-cronjobs", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateCronJob(*s.k8s, "get-cronjobs", "test-cronjob", false)
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.GetCronJobs("get-cronjobs")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)

	// simplify for easier comparison
	actual := sus["2:test-cronjob"]
	actual.Suspend = nil

	s.Require().Equal(
		kubesleep.NewSuspendable(
			kubesleep.CronJob,
			"test-cronjob",
			int32(1),
			nil,
		),
		actual,
	)
}

func (s *Integrationtest) TestSuspendCronJob() {
	deleteNamespace, err := testNamespace("suspend-cronjobs", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateCronJob(*s.k8s, "suspend-cronjobs", "test-cronjob", false)
	s.Require().NoError(err)
	defer delete()

	sus, err := s.k8s.GetCronJobs("suspend-cronjobs")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(1), sus["2:test-cronjob"].Replicas)

	err = sus["2:test-cronjob"].Suspend()
	s.Require().NoError(err)

	sus, err = s.k8s.GetCronJobs("suspend-cronjobs")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(0), sus["2:test-cronjob"].Replicas)
}

func (s *Integrationtest) TestScaleCronJob() {
	deleteNamespace, err := testNamespace("scale-cronjobs", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	delete, err := CreateCronJob(*s.k8s, "scale-cronjobs", "test-cronjob", true)
	s.Require().NoError(err)
	defer delete()

	err = s.k8s.ScaleCronJob("scale-cronjobs", "test-cronjob", 1)
	s.Require().NoError(err)

	sus, err := s.k8s.GetCronJobs("scale-cronjobs")
	s.Require().NoError(err)
	s.Require().NotEmpty(sus)
	s.Require().Equal(int32(1), sus["2:test-cronjob"].Replicas)
}
