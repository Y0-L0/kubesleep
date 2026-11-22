package kubesleep

import (
	"context"
	"fmt"
	"log/slog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type ManifestType int

const (
	Deplyoment ManifestType = iota
	StatefulSet
	CronJob
	KruiseStatefulSet
)

type Suspendable struct {
	manifestType ManifestType
	name         string
	Replicas     int32
	Suspend      func(context.Context) error
}

func NewSuspendable(manifestType ManifestType, name string, Replicas int32, suspend func(context.Context) error) Suspendable {
	return Suspendable{
		manifestType: manifestType,
		name:         name,
		Replicas:     Replicas,
		Suspend:      suspend,
	}
}

func (s Suspendable) Identifier() string {
	return fmt.Sprintf("%d:%s", s.manifestType, s.name)
}

func (s Suspendable) wake(ctx context.Context, namespace string, k8s K8S) error {
	err := k8s.ScaleSuspendable(ctx, namespace, s.manifestType, s.name, s.Replicas)
	if apierrors.IsNotFound(err) {
		slog.Warn("Skipping workload that no longer exists", "namespace", namespace, "name", s.name, "manifestType", s.manifestType)
		return nil
	}
	if err != nil {
		return fmt.Errorf("Failed to scale resource: %s of type: %d in Namespace: %s, %w", s.name, s.manifestType, namespace, err)
	}
	return nil
}

func (s Suspendable) toDto() suspendableDto {
	return suspendableDto{
		ManifestType: s.manifestType,
		Name:         s.name,
		Replicas:     s.Replicas,
	}
}

type suspendableDto struct {
	ManifestType ManifestType
	Name         string
	Replicas     int32
}

func (s suspendableDto) fromDto() Suspendable {
	return Suspendable{
		manifestType: s.ManifestType,
		name:         s.Name,
		Replicas:     s.Replicas,
	}
}
