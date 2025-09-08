package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

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
