package k8s

import (
	"context"
	"fmt"
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	"golang.org/x/sync/errgroup"
)

func mergeNoOverwrite[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			if _, exists := result[k]; exists {
				panic(fmt.Sprintf("Suspendable Identifier duplicate suspendable identifier: %v", k))
			}
			result[k] = v
		}
	}
	return result
}

func (k8s K8Simpl) GetSuspendables(ctx context.Context, namespace string) (map[string]kubesleep.Suspendable, error) {
	g, ctxGroup := errgroup.WithContext(ctx)

	var deployments, statefulSets, cronJobs map[string]kubesleep.Suspendable

	g.Go(func() error {
		var err error
		deployments, err = k8s.getDeployments(ctxGroup, namespace)
		return err
	})
	g.Go(func() error {
		var err error
		statefulSets, err = k8s.getStatefulSets(ctxGroup, namespace)
		return err
	})
	g.Go(func() error {
		var err error
		cronJobs, err = k8s.getCronJobs(ctxGroup, namespace)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return mergeNoOverwrite(deployments, statefulSets, cronJobs), nil
}

func (k8s K8Simpl) ScaleSuspendable(ctx context.Context, namespace string, manifestType kubesleep.ManifestType, name string, replicas int32) error {
	slog.Debug("Scaling suspendable", "namespace", namespace, "name", name, "manifestType", manifestType, "replicas", replicas)
	switch manifestType {
	case kubesleep.Deplyoment:
		return k8s.scaleDeployment(ctx, namespace, name, replicas)
	case kubesleep.StatefulSet:
		return k8s.scaleStatefulSet(ctx, namespace, name, replicas)
	case kubesleep.CronJob:
		return k8s.scaleCronJob(ctx, namespace, name, replicas)
	default:
		return fmt.Errorf("unknown manifest type: %d", manifestType)
	}
}
