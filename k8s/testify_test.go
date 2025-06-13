package k8s

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type Integrationtest struct {
	suite.Suite
	logBuf    bytes.Buffer
	oldLogger *slog.Logger

	stopCluster func() error
	k8s         *K8Simpl
}

func (s *Integrationtest) SetupTest() {
	s.logBuf.Reset()

	handler := slog.NewTextHandler(&s.logBuf, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func (s *Integrationtest) TearDownTest() {
	if !s.T().Failed() || !testing.Verbose() {
		return
	}
	s.T().Log("=== Captured Production Logs ===\n")
	s.T().Log(s.logBuf.String())
}

func (s *Integrationtest) SetupSuite() {
	var err error
	s.k8s, s.stopCluster, err = testCluster()
	s.Require().NoError(err)
}

func (s *Integrationtest) TearDownSuite() {
	s.stopCluster()
}

func testCluster() (*K8Simpl, func() error, error) {
	k8s := &K8Simpl{}

	slog.Debug("Starting a testing kubernetes control plane")
	testEnv := &envtest.Environment{}
	cfg, err := testEnv.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start test cluster %w", err)
	}

	k8s.clientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create client config for the test cluster %w", err)
	}

	k8s.ctx, k8s.cancel = context.WithCancel(context.TODO())

	stop := func() error {
		k8s.cancel()
		return testEnv.Stop()
	}

	return k8s, stop, nil
}

func testNamespace(name string, k8s *K8Simpl) (func() error, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	namespace, err := k8s.clientset.CoreV1().Namespaces().Create(k8s.ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create testing namespace: %s %w", name, err)
	}

	delete := func() error {
		return k8s.clientset.CoreV1().Namespaces().Delete(k8s.ctx, name, metav1.DeleteOptions{})
	}
	return delete, nil
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(Integrationtest))
}
