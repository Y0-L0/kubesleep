package kubesleep

import "fmt"

type ManifestType int

const (
	Deplyoment ManifestType = iota
	StatefulSet
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
	switch s.manifestType {
	case Deplyoment:
		if err := k8s.ScaleDeployment(namespace, s.name, s.Replicas); err != nil {
			return fmt.Errorf("Failed to scale Deployment: %s in Namespace: %s, %w", s.name, namespace, err)
		}
	case StatefulSet:
		if err := k8s.ScaleStatefulSet(namespace, s.name, s.Replicas); err != nil {
			return fmt.Errorf("Failed to scale StatefulSet: %s in Namespace: %s, %w", s.name, namespace, err)
		}
	default:
		return fmt.Errorf("Suspendable: %s in namespace: %s with invalid namifestType: %d", s.name, namespace, s.manifestType)
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
