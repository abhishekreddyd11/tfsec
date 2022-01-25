package custom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquasecurity/defsec/rules"

	"github.com/aquasecurity/tfsec/internal/pkg/block"
	"github.com/aquasecurity/tfsec/internal/pkg/parser"
	"github.com/aquasecurity/tfsec/internal/pkg/scanner"
	"github.com/stretchr/testify/assert"
)

func init() {
	givenCheck(`{
  "checks": [
    {
      "code": "DP006",
      "description": "VPC flow logs must be enabled",
      "requiredTypes": [
        "resource"
      ],
      "requiredLabels": [
        "aws_vpc"
      ],
      "severity": "HIGH",
      "matchSpec": {
        "action": "requiresPresence",
		"name": "aws_flow_log",
		"subMatch": {
		  "action": "isPresent",
		  "name": "log_destination"
		}
      },
      "errorMessage": "VPCs should have an aws_flow_log associated with them",
      "relatedLinks": []
    }
  ]
}
`)
}

var testOrMatchSpec = MatchSpec{
	Action: "or",
	PredicateMatchSpec: []MatchSpec{
		{
			Name:   "name",
			Action: "isPresent",
		},
		{
			Name:   "description",
			Action: "isPresent",
		},
	},
}

var testAndMatchSpec = MatchSpec{
	Action: "and",
	PredicateMatchSpec: []MatchSpec{
		{
			Name:   "name",
			Action: "isPresent",
		},
		{
			Name:   "description",
			Action: "isPresent",
		},
	},
}

var testNestedMatchSpec = MatchSpec{
	Action: "and",
	PredicateMatchSpec: []MatchSpec{
		{
			Name:       "virtualization_type",
			Action:     "equals",
			MatchValue: "paravirtual",
		},
		{
			Action: "or",
			PredicateMatchSpec: []MatchSpec{
				{
					Name:   "image_location",
					Action: "isPresent",
				},
				{
					Name:   "kernel_id",
					Action: "isPresent",
				},
			},
		},
	},
}

var testNotMatchSpec = MatchSpec{
	Action: "not",
	PredicateMatchSpec: []MatchSpec{
		{
			Name:       "virtualization_type",
			Action:     "equals",
			MatchValue: "paravirtual",
		},
	},
}

var testPreConditionCheck = MatchSpec{
	Action:     "equals",
	Name:       "name",
	MatchValue: "test-name",
	PreConditions: []MatchSpec{
		{
			Action:     "equals",
			Name:       "id",
			MatchValue: "test-id",
		},
	},
}

var testAssignVariableMatchSpec = MatchSpec{
	Action: "and",
	PredicateMatchSpec: []MatchSpec{
		{
			Name:           "bucket",
			Action:         "isPresent",
			AssignVariable: "TFSEC_VAR_BUCKET_NAME",
		},
		{
			Name:   "lifecycle_rule",
			Action: "isPresent",
			SubMatch: &MatchSpec{
				Name:       "id",
				Action:     "startsWith",
				MatchValue: "TFSEC_VAR_BUCKET_NAME",
			},
		},
	},
}

func TestRequiresPresenceWithResourcePresent(t *testing.T) {
	scanResults := scanTerraform(t, `
resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"

  tags = {
    Name = "main"
  }
}

resource "aws_flow_log" "example" {
  iam_role_arn    = aws_iam_role.example.arn
  log_destination = aws_cloudwatch_log_group.example.arn
  traffic_type    = "ALL"
  vpc_id          = aws_vpc.example.id
}
`)
	assert.Len(t, scanResults, 0)
}

func TestRequiresPresenceWithResourceMissing(t *testing.T) {
	scanResults := scanTerraform(t, `
resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"

  tags = {
    Name = "main"
  }
}
`)
	assert.Len(t, scanResults, 1)
}

func TestRequiresPresenceWithSubMatchFailing(t *testing.T) {
	scanResults := scanTerraform(t, `
resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"

  tags = {
    Name = "main"
  }
}

resource "aws_flow_log" "example" {
  iam_role_arn    = aws_iam_role.example.arn
  traffic_type    = "ALL"
  vpc_id          = aws_vpc.example.id
}
`)
	assert.Len(t, scanResults, 1)
}

func TestOrMatchFunction(t *testing.T) {

	var tests = []struct {
		name               string
		source             string
		predicateMatchSpec MatchSpec
		expected           bool
	}{
		{
			name: "check `or` match function with no true evaluation",
			source: `
resource "aws_ami" "example" {
}
`,
			predicateMatchSpec: testOrMatchSpec,
			expected:           false,
		},
		{
			name: "check `or` match function with a single true evaluation",
			source: `
resource "aws_ami" "example" {
	name = "placeholder-name"
}
`,
			predicateMatchSpec: testOrMatchSpec,
			expected:           true,
		},
		{
			name: "check `or` match function with all true evaluation",
			source: `
resource "aws_ami" "example" {
	name = "placeholder-name"
	description = "this is a description."
}
`,
			predicateMatchSpec: testOrMatchSpec,
			expected:           true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseFromSource(test.source)[0].GetBlocks()[0]
			result := evalMatchSpec(block, &test.predicateMatchSpec, NewEmptyCustomContext())
			assert.Equal(t, test.expected, result, "`Or` match function evaluating incorrectly.")
		})
	}
}

func TestAndMatchFunction(t *testing.T) {
	var tests = []struct {
		name               string
		source             string
		predicateMatchSpec MatchSpec
		expected           bool
	}{
		{
			name: "check `and` match function with no true evaluation",
			source: `
resource "aws_ami" "example" {
}
`,
			predicateMatchSpec: testAndMatchSpec,
			expected:           false,
		},
		{
			name: "check `and` match function with a single true evaluation",
			source: `
resource "aws_ami" "example" {
	name = "placeholder-name"
}
`,
			predicateMatchSpec: testAndMatchSpec,
			expected:           false,
		},
		{
			name: "check `and` match function with all true evaluation",
			source: `
resource "aws_ami" "example" {
	name = "placeholder-name"
	description = "this is a description."
}
`,
			predicateMatchSpec: testAndMatchSpec,
			expected:           true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseFromSource(test.source)[0].GetBlocks()[0]
			result := evalMatchSpec(block, &test.predicateMatchSpec, NewEmptyCustomContext())
			assert.Equal(t, test.expected, result, "`And` match function evaluating incorrectly.")
		})
	}
}

func TestNestedMatchFunction(t *testing.T) {
	var tests = []struct {
		name               string
		source             string
		predicateMatchSpec MatchSpec
		expected           bool
	}{
		{
			name: "check nested match function with only inner true evaluation",
			source: `
resource "aws_ami" "example" {
	virtualization_type = "hvm"
	image_location = "image-XXXX"
	kernel_id = "XXXXXXXXXX"
}
`,
			predicateMatchSpec: testNestedMatchSpec,
			expected:           false,
		},
		{
			name: "check nested match function with no true evaluation",
			source: `
resource "aws_ami" "example" {
	virtualization_type = "hvm"
}
`,
			predicateMatchSpec: testNestedMatchSpec,
			expected:           false,
		},
		{
			name: "check nested match function with all true evaluation",
			source: `
resource "aws_ami" "example" {
	virtualization_type = "paravirtual"
	image_location = "image-XXXX"
	kernel_id = "XXXXXXXXXX"
}
`,
			predicateMatchSpec: testNestedMatchSpec,
			expected:           true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseFromSource(test.source)[0].GetBlocks()[0]
			result := evalMatchSpec(block, &test.predicateMatchSpec, NewEmptyCustomContext())
			assert.Equal(t, test.expected, result, "Nested match functions evaluating incorrectly.")
		})
	}
}

func TestNotFunction(t *testing.T) {
	var tests = []struct {
		name      string
		source    string
		matchSpec MatchSpec
		expected  bool
	}{
		{
			name: "check that not correctly inverts the outcome of a given predicateMatchSpec",
			source: `
resource "aws_ami" "example" {
	virtualization_type = "paravirtual"
}
`,
			matchSpec: testNotMatchSpec,
			expected:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseFromSource(test.source)[0].GetBlocks()[0]
			result := evalMatchSpec(block, &test.matchSpec, NewEmptyCustomContext())
			assert.Equal(t, test.expected, result, "Not match functions evaluating incorrectly.")
		})
	}
}

func TestPreCondition(t *testing.T) {
	var tests = []struct {
		name      string
		source    string
		matchSpec MatchSpec
		expected  bool
	}{
		{
			name: "check that precondition check prevents check being performed",
			source: `
resource "aws_ami" "testing" {
	name = "something else"
}
`,
			matchSpec: testPreConditionCheck,
			expected:  true,
		},
		{
			name: "check that precondition which passes allows check to be performed which fails",
			source: `
resource "aws_ami" "testing" {
	name = "something else"
	id   = "test-id"
}
`,
			matchSpec: testPreConditionCheck,
			expected:  false,
		},
		{
			name: "check that precondition which passes allows check to be performed which passes",
			source: `
resource "aws_ami" "testing" {
	name = "test-name"
	id   = "test-id"
}
`,
			matchSpec: testPreConditionCheck,
			expected:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseFromSource(test.source)[0].GetBlocks()[0]
			result := evalMatchSpec(block, &test.matchSpec, NewEmptyCustomContext())
			assert.Equal(t, test.expected, result, "precondition functions evaluating incorrectly.")
		})
	}
}

func TestAssignVariable(t *testing.T) {
	var tests = []struct {
		name      string
		source    string
		matchSpec MatchSpec
		expected  bool
	}{
		{
			name: "check assignVariable handling in pass case",
			source: `
resource "aws_s3_bucket" "test-bucket" {
  bucket = "test-bucket"

  lifecycle_rule {
    id = "test-bucket-rule-1"
  }
}
`,
			matchSpec: testAssignVariableMatchSpec,
			expected:  true,
		},
		{
			name: "check assignVariable handling in fail case",
			source: `
resource "aws_s3_bucket" "test-bucket" {
  bucket = "test-bucket"

  lifecycle_rule {
    id = "not-bucket-name-rule-1"
  }
}
`,
			matchSpec: testAssignVariableMatchSpec,
			expected:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseFromSource(test.source)[0].GetBlocks()[0]
			result := evalMatchSpec(block, &test.matchSpec, NewEmptyCustomContext())
			assert.Equal(t, test.expected, result, "processing variable assignments incorrectly.")
		})
	}
}

func TestRegexMatches(t *testing.T) {
	var tests = []struct {
		name               string
		source             string
		predicateMatchSpec MatchSpec
		expected           bool
	}{
		{
			name: "check `regexMatches` in pass case",
			source: `
resource "google_compute_instance" "default" {
  name         = "test_instance_name"
  machine_type = "e2-medium"
  region       = "europe-west3"
  zone         = "europe-west3-a"
}
`,
			predicateMatchSpec: MatchSpec{
				Name:       "name",
				Action:     "regexMatches",
				MatchValue: "^test_.*$",
			},
			expected: true,
		},
		{
			name: "check `regexMatches` in regex-not-matching fail case",
			source: `
resource "google_compute_instance" "default" {
  name         = "wrong_test_instance_name"
  machine_type = "e2-medium"
  region       = "europe-west3"
  zone         = "europe-west3-a"
}
`,
			predicateMatchSpec: MatchSpec{
				Name:       "name",
				Action:     "regexMatches",
				MatchValue: "^test_.*$",
			},
			expected: false,
		},
		{
			name: "check `regexMatches` in attribute-not-found fail case",
			source: `
resource "google_compute_instance" "default" {
  name         = "wrong_test_instance_name"
  machine_type = "e2-medium"
  region       = "europe-west3"
  zone         = "europe-west3-a"
}
`,
			predicateMatchSpec: MatchSpec{
				Name:       "not-name",
				Action:     "regexMatches",
				MatchValue: "^test_.*$",
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseFromSource(test.source)[0].GetBlocks()[0]
			result := evalMatchSpec(block, &test.predicateMatchSpec, NewEmptyCustomContext())
			assert.Equal(t, result, test.expected, "`regexMatches` match function evaluating incorrectly.")
		})
	}
}

func givenCheck(jsonContent string) {
	var checksfile ChecksFile
	err := json.NewDecoder(strings.NewReader(jsonContent)).Decode(&checksfile)
	if err != nil {
		panic(err)
	}
	processFoundChecks(checksfile)
}

func scanTerraform(t *testing.T, mainTf string) []rules.Result {
	dirName, err := ioutil.TempDir("", "tfsec-testing-")
	assert.NoError(t, err)

	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", dirName, "main.tf"), []byte(mainTf), os.ModePerm)
	assert.NoError(t, err)

	blocks, err := parser.New(dirName, parser.OptionStopOnHCLError()).ParseDirectory()
	assert.NoError(t, err)

	res, _ := scanner.New(scanner.OptionStopOnErrors()).Scan(blocks)
	return res
}

// This function is copied from setup_test.go as it is not possible to import function from test files.
// TODO: Extract into a testing utility package once the amount of duplication justifies introducing an extra package.
func ParseFromSource(source string) block.Modules {
	path := createTestFile("test.tf", source)
	modules, err := parser.New(filepath.Dir(path), parser.OptionStopOnHCLError()).ParseDirectory()
	if err != nil {
		panic(err)
	}
	return modules
}

// This function is copied from setup_test.go as it is not possible to import function from test files.
// TODO: Extract into a testing utility package once the amount of duplication justifies introducing an extra package.
func createTestFile(filename, contents string) string {
	dir, err := ioutil.TempDir(os.TempDir(), "tfsec")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, filename)
	if err := ioutil.WriteFile(path, []byte(contents), 0600); err != nil {
		panic(err)
	}
	return path
}