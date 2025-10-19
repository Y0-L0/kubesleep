package k8s

import (
	"context"
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s K8Simpl) getCronJobs(ctx context.Context, namespace string) (map[string]kubesleep.Suspendable, error) {
	cronJobs, err := k8s.clientset.BatchV1().
		CronJobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	suspendables := map[string]kubesleep.Suspendable{}

	for _, job := range cronJobs.Items {
		var suspend func(context.Context) error
		if job.Spec.Suspend != nil && *job.Spec.Suspend {
			suspend = k8s.noopSuspendCronJob(namespace, job.Name)
		} else {
			suspend = k8s.suspendCronJob(namespace, job.Name)
		}

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

func (k8s K8Simpl) noopSuspendCronJob(namespace, name string) func(context.Context) error {
	return func(ctx context.Context) error {
		slog.Debug("CronJob already suspended; skipping suspend", "namespace", namespace, "name", name)
		return nil
	}
}

func (k8s K8Simpl) suspendCronJob(namespace, name string) func(context.Context) error {
	return func(ctx context.Context) error {
		result := true
		return k8s.setCronJobSuspended(ctx, namespace, name, &result)
	}
}

func (k8s K8Simpl) scaleCronJob(ctx context.Context, namespace, name string, replicas int32) error {
	return k8s.setCronJobSuspended(ctx, namespace, name, replicasToSuspended(replicas))
}

func (k8s K8Simpl) setCronJobSuspended(ctx context.Context, namespace, name string, suspended *bool) error {
	cj, err := k8s.clientset.BatchV1().CronJobs(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}

	cj.Spec.Suspend = suspended

	_, err = k8s.clientset.BatchV1().CronJobs(namespace).Update(
		ctx,
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
