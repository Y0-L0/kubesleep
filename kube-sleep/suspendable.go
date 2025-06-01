package kubesleep

type suspendable struct {
	manifestType string
	name         string
	replicas     int32
  suspend func() error
}

func (s suspendable)toDto() suspendableDto {
  return suspendableDto{
    ManifestType: s.manifestType,
    Name: s.name,
    Replicas: s.replicas,
  }
}

type suspendableDto struct {
  ManifestType string
  Name string
  Replicas int32
}

func (s suspendableDto)fromDto() suspendable {
  return suspendable{
    manifestType: s.ManifestType,
    name: s.Name,
    replicas: s.Replicas,
  }
}
