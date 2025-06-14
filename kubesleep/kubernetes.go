package kubesleep

type K8S interface {
	GetSuspendableNamespace(string) (SuspendableNamespace, error)

	GetDeployments(string) (map[string]Suspendable, error)
	ScaleDeployment(string, string, int32) error

	GetStatefulSets(string) (map[string]Suspendable, error)
	ScaleStatefulSet(string, string, int32) error

	GetStateFile(string) (*SuspendStateFile, error)
	CreateStateFile(string, *SuspendStateFile) (*SuspendStateFile, error)
	UpdateStateFile(string, *SuspendStateFile) (*SuspendStateFile, error)
	DeleteStateFile(string) error
}

type K8SFactory func() (K8S, error)
