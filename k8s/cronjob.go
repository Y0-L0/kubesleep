package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) GetCronJobs(namespace string) (map[string]kubesleep.Suspendable, error) {
	cronJobs, err := k8s.clientset.BatchV1().
		CronJobs(namespace).
		List(k8s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]kubesleep.Suspendable{}

	for _, job := range cronJobs.Items {
		name := job.Name
		suspend := k8s.suspendCronJob(namespace, name)

		s := kubesleep.NewSuspendable(
			kubesleep.CronJob,
			job.Name,
			suspendedToReplicas(*job.Spec.Suspend),
			suspend,
		)
		slog.Debug("parsed Suspendable", "Suspendable", s, "namespace", namespace)
		suspendables[s.Identifier()] = s
	}

	return suspendables, nil
}

func (k8s K8Simpl) suspendCronJob(namespace, name string) func() error {
	return func() error {
		result := true
		return k8s.setCronJobSuspended(namespace, name, &result)
	}
}

func (k8s K8Simpl) ScaleCronJob(namespace, name string, replicas int32) error {
	return k8s.setCronJobSuspended(namespace, name, replicasToSuspended(replicas))
}

func (k8s K8Simpl) setCronJobSuspended(namespace, name string, suspended *bool) error {
	cj, err := k8s.clientset.BatchV1().CronJobs(namespace).Get(
		k8s.ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}

	cj.Spec.Suspend = suspended

	_, err = k8s.clientset.BatchV1().CronJobs(namespace).Update(
		k8s.ctx,
		cj,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	if *suspended {
		slog.Info("Suspended CronJob", "name", name, "namespace", namespace)
	} else {
		slog.Info("Woke up CronJob", "name", name, "namespace", namespace)
	}
	return nil
}

func suspendedToReplicas(b bool) int32 {
	if b {
		return int32(0)
	}
	return int32(1)
}

func replicasToSuspended(i int32) *bool {
	var result bool
	switch i {
	case 0:
		result = true
	case 1:
		result = false
	default:
		panic("can't scale CronJob beyond 1=on")
	}
	return &result
}
