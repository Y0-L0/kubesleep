package kubesleep

var TEST_SUSPENDABLE = Suspendable{
	manifestType: Deplyoment,
	name:         "test-deployment",
	Replicas:     int32(2),
}

var TEST_SUSPENDABLES = map[string]Suspendable{
	TEST_SUSPENDABLE.Identifier(): TEST_SUSPENDABLE,
}
