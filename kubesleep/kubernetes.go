package kubesleep

import "context"

type K8S interface {
	GetSuspendableNamespace(ctx context.Context, namespace string) (SuspendableNamespace, error)
	GetSuspendableNamespaces(ctx context.Context) ([]SuspendableNamespace, error)

	GetSuspendables(ctx context.Context, namespace string) (map[string]Suspendable, error)
	ScaleSuspendable(ctx context.Context, namespace string, manifestType ManifestType, name string, replicas int32) error

	GetStateFile(ctx context.Context, namespace string) (*SuspendState, SuspendStateActions, error)
	CreateStateFile(ctx context.Context, namespace string, data map[string]string) (SuspendStateActions, error)
	DeleteStateFile(ctx context.Context, namespace string) error
}

type K8SFactory func() (K8S, error)

type StatefileAlreadyExistsError string

func (e StatefileAlreadyExistsError) Error() string { return string(e) }

type StatefileNotFoundError string

func (e StatefileNotFoundError) Error() string { return string(e) }

type NamespaceTerminatingError string

func (e NamespaceTerminatingError) Error() string { return string(e) }
