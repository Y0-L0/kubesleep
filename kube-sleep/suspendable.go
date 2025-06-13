package kubesleep

import "fmt"

type Suspendable struct {
	manifestType string
	name         string
	Replicas     int32
	Suspend      func() error
}

func NewSuspendable(manifestType string, name string, Replicas int32, suspend func() error) Suspendable {
	return Suspendable{
		manifestType: manifestType,
		name:         name,
		Replicas:     Replicas,
		Suspend:      suspend,
	}
}

func (s Suspendable) Identifier() string {
	return s.manifestType + s.name
}

func (s Suspendable) wake(namespace string, k8s K8S) error {
	switch s.manifestType {
	case "StatefulSet":
		if err := k8s.ScaleStatefulSet(namespace, s.name, s.Replicas); err != nil {
			return err
		}
	case "Deployment":
		if err := k8s.ScaleDeployment(namespace, s.name, s.Replicas); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Suspendable: %s with invalid namifestType: %s", s.name, s.manifestType)
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
	ManifestType string
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
