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

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/modern-go/concurrent"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
	"github.com/stretchr/testify/assert"
)

// Try to install every mixin from the mixin repository
// verify that each package generated has the yaml files
func TestInstallMixin(t *testing.T) {
	body, err := queryWebsite(defaultWebsite)
	if err != nil {
		t.Errorf("failed to query website %v", err)
	}
	mixins, err := parseMixinJSON(body)
	if err != nil {
		t.Errorf("failed to parse mixin body %v", err)
	}

	// download each mixin in turn
	for _, m := range mixins {
		t.Run(m.Name, func(t *testing.T) {
			t.Parallel()
			m := m
			testInstallMixin(t, m)
		})
	}
}

func testInstallMixin(t *testing.T, m mixin) {
	tmpdir := t.TempDir()

	mixinURL := path.Join(m.URL, m.Subdir)

	fmt.Printf("installing %v\n", mixinURL)
	dldir := path.Join(tmpdir, m.Name+"mixin-test")

	err := os.Mkdir(dldir, 0755)
	assert.NoError(t, err)

	jsonnetHome := "vendor"

	err = downloadMixin(mixinURL, jsonnetHome, dldir)
	assert.NoError(t, err)

	importPath, err := locateImportFile(path.Join(dldir, jsonnetHome), mixinURL)
	assert.NoError(t, err)
	deps := []string{importPath}

	results := path.Join(tmpdir, "out")
	cfg := &GenerateConfig{
		Dir: results,
		RulesAlertsCfgs: []*RulesAlertsConfig{
			{
				Dst: "loki-rules-alerts.yml",
				MixinOpts: &mixer.RulesAlertsOptions{
					DataSource: mixer.Loki,
					ImportPath: importPath,
				},
				GenOpts: &mixer.GeneratorOptions{
					Eval: mixer.NewEvaluator(deps),
				},
				Formatter: mixer.JSONtoYaml,
			},
			{
				Dst: "prom-rules-alerts.yml",
				MixinOpts: &mixer.RulesAlertsOptions{
					DataSource: mixer.Prometheus,
					ImportPath: importPath,
				},
				GenOpts: &mixer.GeneratorOptions{
					Eval: mixer.NewEvaluator(deps),
				},
				Formatter: mixer.JSONtoYaml,
			},
		},
		DashCfg: &DashboardsConfig{
			MixinOpts: &mixer.DashboardsOptions{ImportPath: importPath},
			GenOpts:   &mixer.GeneratorOptions{Eval: mixer.NewEvaluator(deps)},
			Formatter: mixer.JSONtoYaml,
		},
	}

	_, err = generateMixin(dldir, jsonnetHome, mixinURL, cfg)
	assert.NoError(t, err)

	// verify that alerts, rules, dashboards exist
	contents := concurrent.NewMap()
	err = filepath.WalkDir(results, func(path string, d fs.DirEntry, err error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		contents.Store(path, d)
		return nil
	})
	assert.NoError(t, err)
	for _, raCfg := range cfg.RulesAlertsCfgs {
		full := path.Join(results, raCfg.Dst)
		_, ok := contents.Load(full)
		assert.True(t, ok, full+" not found")
	}

	dashboards := concurrent.NewMap()
	err = filepath.WalkDir(path.Join(results, "dashboards"), func(path string, d fs.DirEntry, err error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		t.Log(path)
		dashboards.Store(path, d)
		return nil
	})
	assert.NoError(t, err)
	fail := true
	dashboards.Range(func(key, value interface{}) bool {
		fail = false
		return fail
	})
	assert.False(t, fail, "no dashboards found")

	// verify that the output of alerts and rules matches using jsonnet
}
