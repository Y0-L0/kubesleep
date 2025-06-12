package kubesleep

import "fmt"

type suspendable struct {
	manifestType string
	name         string
	replicas     int32
	suspend      func() error
}

func (s suspendable) wake(namespace string, k8s *k8simpl) error {
	if s.manifestType == "StatefulSet" {
		if err := k8s.scaleStatefulSet(namespace, s.name, s.replicas); err != nil {
			return err
		}
	} else if s.manifestType == "Deployment" {
		if err := k8s.scaleDeployment(namespace, s.name, s.replicas); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("suspendable: %s with invalid namifestType: %s", s.name, s.manifestType)
	}
	return nil
}

func (s suspendable) toDto() suspendableDto {
	return suspendableDto{
		ManifestType: s.manifestType,
		Name:         s.name,
		Replicas:     s.replicas,
	}
}

type suspendableDto struct {
	ManifestType string
	Name         string
	Replicas     int32
}

func (s suspendableDto) fromDto() suspendable {
	return suspendable{
		manifestType: s.ManifestType,
		name:         s.Name,
		replicas:     s.Replicas,
	}
}
