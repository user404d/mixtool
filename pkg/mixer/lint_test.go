// Copyright 2018 mixtool authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mixer

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLintLokiAlerts(t *testing.T) {
	filename, delete := writeTempFile(t, "alerts.jsonnet", lokiAlerts)
	defer delete()

	e := NewDefaultEvaluator()
	opts := &RulesAlertsOptions{DataSource: Loki, ImportPath: filename}
	errs := make(chan error)
	go lintRulesAlerts(e, opts, NewLokiLinter(), errs)
	for err := range errs {
		t.Errorf("linting wrote unexpected output: %v", err)
	}
}

func TestLintLokiRules(t *testing.T) {
	filename, delete := writeTempFile(t, "rules.jsonnet", lokiRules)
	defer delete()

	e := NewDefaultEvaluator()
	opts := &RulesAlertsOptions{DataSource: Loki, ImportPath: filename}
	errs := make(chan error)
	go lintRulesAlerts(e, opts, NewLokiLinter(), errs)
	for err := range errs {
		t.Errorf("linting wrote unexpected output: %v", err)
	}
}

func TestLintPrometheusAlerts(t *testing.T) {
	const testAlerts = promAlerts + `+
{
  _config+:: {
     kubeStateMetricsSelector: 'job="ksm"',
  }
}`
	filename, delete := writeTempFile(t, "alerts.jsonnet", testAlerts)
	defer delete()

	e := NewDefaultEvaluator()
	opts := &RulesAlertsOptions{DataSource: Prometheus, ImportPath: filename}
	errs := make(chan error)
	go lintRulesAlerts(e, opts, NewPrometheusLinter(), errs)
	for err := range errs {
		t.Errorf("linting wrote unexpected output: %v", err)
	}
}

func TestLintPrometheusRules(t *testing.T) {
	filename, delete := writeTempFile(t, "rules.jsonnet", promRules)
	defer delete()

	e := NewDefaultEvaluator()
	opts := &RulesAlertsOptions{DataSource: Prometheus, ImportPath: filename}
	errs := make(chan error)
	go lintRulesAlerts(e, opts, NewPrometheusLinter(), errs)
	for err := range errs {
		t.Errorf("linting wrote unexpected output: %v", err)
	}
}

func TestLintGrafana(t *testing.T) {
	e := NewDefaultEvaluator()
	opts := &DashboardsOptions{ImportPath: "lint_test_dashboard.json"}
	errs := make(chan error)
	go lintGrafanaDashboards(e, opts, errs)
	for err := range errs {
		t.Errorf("linting wrote unexpected output: %v", err)
	}
}

func writeTempFile(t *testing.T, pattern string, contents string) (filename string, delete func()) {
	f, err := ioutil.TempFile("", pattern)
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
	}

	if _, err := f.WriteString(contents); err != nil {
		t.Errorf("failed to write rules.jsonnet to disk: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Errorf("failed to close temp file: %v", err)
	}

	return f.Name(), func() { os.Remove(f.Name()) }
}
