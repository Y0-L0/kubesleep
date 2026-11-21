package k8s

import (
	"maps"
	"slices"
)

func (s *Integrationtest) TestGetSuspendablesByLabelSelector() {
	namespace := "get-suspendables-by-label-selector"
	deleteNamespace, err := testNamespace(s.ctx, namespace, s.k8s, false)
	s.Require().NoError(err)
	defer deleteNamespace()

	deleteDeployment, err := CreateDeployment(s.ctx, *s.k8s, namespace, "web", int32(1))
	s.Require().NoError(err)
	defer deleteDeployment()
	deleteStatefulSet, err := CreateStatefulSet(s.ctx, *s.k8s, namespace, "db", int32(1))
	s.Require().NoError(err)
	defer deleteStatefulSet()
	deleteCronJob, err := CreateCronJob(s.ctx, *s.k8s, namespace, "backup", false)
	s.Require().NoError(err)
	defer deleteCronJob()
	deleteKruiseStatefulSet, err := CreateKruiseStatefulSet(s.ctx, *s.k8s, namespace, "mail", int32(1))
	s.Require().NoError(err)
	defer deleteKruiseStatefulSet()

	tests := []struct {
		name      string
		selectors []string
		expected  []string
	}{
		{"no selector", nil, []string{"0:web", "1:db", "2:backup", "3:mail"}},
		{"one selector", []string{"app=web"}, []string{"0:web"}},
		{"selectors match any", []string{"app=web", "app=mail"}, []string{"0:web", "3:mail"}},
		{"set-based selector", []string{"app in (db,backup)"}, []string{"1:db", "2:backup"}},
		{"overlapping selectors", []string{"app=db", "app in (db,backup)"}, []string{"1:db", "2:backup"}},
		{"no match", []string{"app=missing"}, []string{}},
	}
	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			suspendables, err := s.k8s.GetSuspendables(s.ctx, namespace, testCase.selectors)
			s.Require().NoError(err)
			s.Require().ElementsMatch(testCase.expected, slices.Collect(maps.Keys(suspendables)))
		})
	}
}
