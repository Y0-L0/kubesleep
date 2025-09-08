package kubesleep

import "fmt"

type ManifestType int

const (
	Deplyoment ManifestType = iota
	StatefulSet
	CronJob
)

type Suspendable struct {
	manifestType ManifestType
	name         string
	Replicas     int32
	Suspend      func() error
}

func NewSuspendable(manifestType ManifestType, name string, Replicas int32, suspend func() error) Suspendable {
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

func (s Suspendable) wake(namespace string, k8s K8S) error {
	if err := k8s.ScaleSuspendable(namespace, s.manifestType, s.name, s.Replicas); err != nil {
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
