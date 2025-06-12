package kubesleep

var TEST_SUSPENDABLES = map[string]suspendable{
	"Deploymenttest-deployment": suspendable{
		manifestType: "Deployment",
		name:         "test-deployment",
		replicas:     int32(2),
	},
}
