package kubesleep

var TEST_SUSPENDABLES = []suspendable{
	suspendable{
		manifestType: "Deployment",
		name:         "testDeployment",
		replicas:     int32(2),
	},
}
