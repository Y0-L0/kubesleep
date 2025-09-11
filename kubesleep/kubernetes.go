package kubesleep

type K8S interface {
	GetSuspendableNamespace(string) (SuspendableNamespace, error)
	GetSuspendableNamespaces() ([]SuspendableNamespace, error)

	GetSuspendables(string) (map[string]Suspendable, error)
	ScaleSuspendable(namespace string, manifestType ManifestType, name string, replicas int32) error

	GetStateFile(string) (*SuspendStateFile, StateFileActions, error)
	CreateStateFile(string, map[string]string) (StateFileActions, error)
	DeleteStateFile(string) error
}

type K8SFactory func() (K8S, error)

type StatefileAlreadyExistsError string

func (e StatefileAlreadyExistsError) Error() string { return string(e) }
