package kubesleep

var TEST_SUSPENDABLES = map[string]Suspendable{
	"Deploymenttest-deployment": Suspendable{
		manifestType: "Deployment",
		name:         "test-deployment",
		Replicas:     int32(2),
	},
}
