package k8s

import (
	"os"
	"path/filepath"

	"github.com/Y0-L0/kubesleep/kubesleep"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var STANDARD_NAMESPACES = []kubesleep.SuspendableNamespace{
	kubesleep.NewSuspendableNamespace("default", true),
	kubesleep.NewSuspendableNamespace("kube-node-lease", true),
	kubesleep.NewSuspendableNamespace("kube-public", true),
	kubesleep.NewSuspendableNamespace("kube-system", true),
}

func (s *Integrationtest) TestLoadKubeconfig() {
	apiConfig := clientcmdapi.NewConfig()
	apiConfig.CurrentContext = "default"
	apiConfig.Clusters["default"] = &clientcmdapi.Cluster{
		Server:                   s.restconfig.Host,
		CertificateAuthorityData: s.restconfig.CAData,
	}
	apiConfig.AuthInfos["default"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: s.restconfig.CertData,
		ClientKeyData:         s.restconfig.KeyData,
	}
	apiConfig.Contexts["default"] = &clientcmdapi.Context{
		Cluster:  "default",
		AuthInfo: "default",
	}

	tmp := s.T().TempDir()
	kubeconfigPath := filepath.Join(tmp, "config")
	err := clientcmd.WriteToFile(*apiConfig, kubeconfigPath)
	s.Require().NoError(err)

	err = os.Setenv("KUBECONFIG", kubeconfigPath)
	s.Require().NoError(err)

	_, err = NewK8S()
	s.Require().NoError(err)
}

func (s *Integrationtest) TestDeploymentStatefulsetNameConflict() {
	deleteNamespace, err := testNamespace("deployment-statefulset-name-conflict", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	deleteStatefulSet, err := CreateStatefulSet(*s.k8s, "deployment-statefulset-name-conflict", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer deleteStatefulSet()

	deleteDeployment, err := CreateDeployment(*s.k8s, "deployment-statefulset-name-conflict", "test-statefulset", int32(2))
	s.Require().NoError(err)
	defer deleteDeployment()
}

func (s *Integrationtest) TestGetSuspendableNamespace() {
	deleteNamespace, err := testNamespace("get-suspendable-namespace", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	namespace, err := s.k8s.GetSuspendableNamespace("get-suspendable-namespace")
	s.Require().NoError(err)

	s.Require().Equal(
		kubesleep.NewSuspendableNamespace("get-suspendable-namespace", true),
		namespace,
	)
}

func (s *Integrationtest) TestGetNonSuspendableNamespace() {
	deleteNamespace, err := testNamespace("get-non-suspendable-namespace", s.k8s, true)
	s.Require().NoError(err)
	defer deleteNamespace()

	namespace, err := s.k8s.GetSuspendableNamespace("get-non-suspendable-namespace")
	s.Require().NoError(err)

	s.Require().Equal(
		kubesleep.NewSuspendableNamespace("get-non-suspendable-namespace", false),
		namespace,
	)
}

func (s *Integrationtest) TestGetNamespace() {
	expected := append(STANDARD_NAMESPACES, kubesleep.NewSuspendableNamespace("get-namespaces", true))
	deleteNamespace, err := testNamespace("get-namespaces", s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	namespaces, err := s.k8s.GetSuspendableNamespaces()
	s.Require().NoError(err)
	for _, e := range expected {
		s.Require().Contains(namespaces, e)
	}
}
