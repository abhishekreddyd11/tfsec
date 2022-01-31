package iam

import (
	"testing"

	"github.com/aquasecurity/defsec/provider/aws/iam"
	"github.com/aquasecurity/tfsec/internal/pkg/adapter/testutil"
)

func Test_adaptUsers(t *testing.T) {
	t.SkipNow()
	tests := []struct {
		name      string
		terraform string
		expected  []iam.User
	}{
		{
			name: "basic",
			terraform: `
resource "" "example" {
    
}
`,
			expected: []iam.User{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modules := testutil.CreateModulesFromSource(test.terraform, ".tf", t)
			adapted := adaptUsers(modules)
			testutil.AssertDefsecEqual(t, test.expected, adapted)
		})
	}
}