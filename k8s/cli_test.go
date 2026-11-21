package k8s

import (
	"log/slog"

	kubesleep "github.com/Y0-L0/kubesleep/kubesleep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// runCli runs one kubesleep command against the test cluster.
func (s *Integrationtest) runCli(args ...string) {
	command, _ := kubesleep.NewParser(
		append([]string{"kubesleep"}, args...),
		func() (kubesleep.K8S, error) { return s.k8s, nil },
		func(slog.Level) {},
	)
	s.Require().NoError(command.Execute(), "kubesleep %v", args)
}

// deploymentReplicas maps each Deployment name in the namespace to its replica count.
func (s *Integrationtest) deploymentReplicas(namespace string) map[string]int32 {
	deployments, err := s.k8s.clientset.AppsV1().Deployments(namespace).List(s.ctx, metav1.ListOptions{})
	s.Require().NoError(err)
	replicas := map[string]int32{}
	for _, deployment := range deployments.Items {
		replicas[deployment.Name] = *deployment.Spec.Replicas
	}
	return replicas
}

func (s *Integrationtest) TestWakeSkipsWorkloadDeletedWhileSuspended() {
	namespace := "wake-skips-deleted-workload"
	deleteNamespace, err := testNamespace(s.ctx, namespace, s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	deleteKept, err := CreateDeployment(s.ctx, *s.k8s, namespace, "kept", int32(2))
	s.Require().NoError(err)
	defer deleteKept()
	deleteGone, err := CreateDeployment(s.ctx, *s.k8s, namespace, "gone", int32(3))
	s.Require().NoError(err)

	s.runCli("suspend", "-n", namespace)
	s.Require().NoError(deleteGone())

	s.runCli("wake", "-n", namespace)
	s.Require().Equal(map[string]int32{"kept": 2}, s.deploymentReplicas(namespace))
	_, _, err = s.k8s.GetStateFile(s.ctx, namespace)
	s.Require().ErrorAs(err, new(kubesleep.StatefileNotFoundError))
}
