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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/monitoring-mixins/mixtool/pkg/mixer"
	"github.com/urfave/cli"
)

func generateCommand() cli.Command {
	flags := []cli.Flag{
		cli.StringSliceFlag{
			Name: "jpath, J",
		},
		cli.BoolTFlag{
			Name: "yaml, y",
		},
	}

	return cli.Command{
		Name:  "generate",
		Usage: "Generate manifests from jsonnet input",
		Flags: append(flags,
			cli.StringSliceFlag{
				Name:  "data-sources, s",
				Usage: "The sources used when evaluating rules and alerts (loki,prometheus)",
			},
			cli.StringFlag{
				Name:  "directory, d",
				Usage: "The directory where generated outputs are written to",
				Value: "out",
			},
		),
		Subcommands: cli.Commands{
			cli.Command{
				Name:  "alerts",
				Usage: "Generate Prometheus alerts based on the mixins",
				Flags: append(flags,
					cli.StringFlag{
						Name:  "pattern, p",
						Usage: "Suffix of the file where Prometheus alerts are written",
						Value: "alerts",
					},
				),
				Action: generateAction(func(cfg *GenerateConfig) error {
					mixed, err := generateRulesAlerts(cfg.RulesAlertsCfgs, mixer.NewAlertsMixin)
					if err != nil {
						return err
					}
					return writeMixed(cfg.Dir, mixed)
				}),
			},
			cli.Command{
				Name:  "rules",
				Usage: "Generate Prometheus rules based on the mixins",
				Flags: append(flags,
					cli.StringFlag{
						Name:  "pattern, p",
						Usage: "Suffix of the file where Prometheus rules are written",
						Value: "rules",
					},
				),
				Action: generateAction(func(cfg *GenerateConfig) error {
					mixed, err := generateRulesAlerts(cfg.RulesAlertsCfgs, mixer.NewRulesMixin)
					if err != nil {
						return err
					}
					return writeMixed(cfg.Dir, mixed)
				}),
			},
			cli.Command{
				Name:  "dashboards",
				Usage: "Generate Grafana dashboards based on the mixins",
				Action: generateAction(func(cfg *GenerateConfig) error {
					mixed, err := generateDashboards(cfg.DashCfg)
					if err != nil {
						return err
					}
					return writeMixed(cfg.Dir, mixed)
				}),
			},
			cli.Command{
				Name:  "all",
				Usage: "Generate all resources - alerts, rules, and Grafana dashboards",
				Flags: append(flags,
					cli.StringFlag{
						Name:  "pattern, p",
						Usage: "Suffix of the file where alerts are written",
						Value: "rules-alerts",
					},
				),
				Action: generateAction(func(cfg *GenerateConfig) error {
					mixed, err := generateAll(cfg)
					if err != nil {
						return nil
					}
					return writeMixed(cfg.Dir, mixed)
				}),
			},
		},
	}
}

type RulesAlertsConfig struct {
	Dst       string
	MixinOpts *mixer.RulesAlertsOptions
	GenOpts   *mixer.GeneratorOptions
	Formatter mixer.Formatter
}

type DashboardsConfig struct {
	MixinOpts *mixer.DashboardsOptions
	GenOpts   *mixer.GeneratorOptions
	Formatter mixer.Formatter
}

type GenerateConfig struct {
	Dir             string
	RulesAlertsCfgs []*RulesAlertsConfig
	DashCfg         *DashboardsConfig
}

type GenerateAction func(cfg *GenerateConfig) error

func generateAction(generate GenerateAction) cli.ActionFunc {
	return func(c *cli.Context) error {
		jPathFlag := c.StringSlice("jpath")
		filename := c.Args().First()
		if filename == "" {
			return fmt.Errorf("no jsonnet file given")
		}

		jPathFlag, err := availableVendor(filename, jPathFlag)
		if err != nil {
			return err
		}

		dataSources := c.StringSlice("data-source")
		if len(dataSources) == 0 {
			dataSources = []string{"loki", "prometheus"}
		}

		formatter := mixer.NoFormatter
		if c.BoolT("yaml") {
			formatter = mixer.JSONtoYaml
		}

		directory := c.String("directory")
		if directory == "" {
			directory = "out"
		}

		pattern := c.String("pattern")
		if pattern == "" || pattern == "-" || pattern == "stdout" {
			pattern = "/dev/stdout"
		} else {
			pattern = "%s-" + pattern
			if c.BoolT("yaml") {
				pattern += ".yml"
			} else {
				pattern += ".json"
			}
		}

		raCfgs := make([]*RulesAlertsConfig, 0)

		for _, dataSource := range dataSources {
			raCfg := &RulesAlertsConfig{
				GenOpts: &mixer.GeneratorOptions{
					Eval: mixer.NewEvaluator(jPathFlag),
				},
				Formatter: formatter,
			}

			switch mixer.DataSource(dataSource) {
			case mixer.Loki:
				if pattern != "/dev/stdout" {
					raCfg.Dst = fmt.Sprintf(pattern, "loki")
				} else {
					raCfg.Dst = pattern
				}
				raCfg.MixinOpts = &mixer.RulesAlertsOptions{
					DataSource: mixer.Loki,
					ImportPath: filename,
				}
			case mixer.Prometheus:
				if pattern != "/dev/stdout" {
					raCfg.Dst = fmt.Sprintf(pattern, "prom")
				} else {
					raCfg.Dst = pattern
				}
				raCfg.MixinOpts = &mixer.RulesAlertsOptions{
					DataSource: mixer.Prometheus,
					ImportPath: filename,
				}
			default:
				continue
			}
			raCfgs = append(raCfgs, raCfg)
		}

		return generate(&GenerateConfig{
			Dir:             directory,
			RulesAlertsCfgs: raCfgs,
			DashCfg: &DashboardsConfig{
				MixinOpts: &mixer.DashboardsOptions{ImportPath: filename},
				GenOpts: &mixer.GeneratorOptions{
					Eval: mixer.NewEvaluator(jPathFlag),
				},
				Formatter: formatter,
			},
		})
	}
}

func generateRulesAlerts(cfgs []*RulesAlertsConfig, mFactory mixer.MixinBuilder) (map[string]mixer.Mixin, error) {
	mixed := make(map[string]mixer.Mixin, 0)
	for _, cfg := range cfgs {
		mixin := mFactory(cfg.MixinOpts)
		gen := mixer.NewGenerator(cfg.GenOpts)
		out, err := gen.Generate(mixin)
		if err != nil {
			return nil, err
		}

		formatted, err := mixer.Mixin(out).ApplyFormatter(cfg.Formatter)
		if err != nil {
			return nil, err
		}

		// deal with stdout as dst
		if val, ok := mixed[cfg.Dst]; !ok {
			mixed[cfg.Dst] = formatted
		} else {
			mixed[cfg.Dst] = bytes.Join([][]byte{val, formatted}, []byte("----\n"))
		}
	}
	return mixed, nil
}

func generateDashboards(cfg *DashboardsConfig) (map[string]mixer.Mixin, error) {
	mixin := mixer.NewDashboardsMixin(cfg.MixinOpts)
	gen := mixer.NewGenerator(cfg.GenOpts)
	out, err := gen.Generate(mixin)
	if err != nil {
		return nil, err
	}

	var dashboards map[string]json.RawMessage
	if err := json.Unmarshal(out, &dashboards); err != nil {
		return nil, err
	}

	mixed := make(map[string]mixer.Mixin, len(dashboards))
	for dst, dashboard := range dashboards {
		mixin, err := mixer.Mixin(dashboard).ApplyFormatter(cfg.Formatter)
		if err != nil {
			return nil, err
		}
		mixed[dst] = mixin
	}

	return mixed, nil
}

func writeMixed(dir string, mixed map[string]mixer.Mixin) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	for dst, out := range mixed {
		if dst != "/dev/stdout" {
			full := path.Join(dir, dst)
			err := os.MkdirAll(path.Dir(full), 0755)
			if err != nil {
				return err
			}
			err = os.WriteFile(full, out, 0644)
			if err != nil {
				return err
			}
		} else {
			fmt.Print(string(out))
		}
	}

	return nil
}

func generateAll(cfg *GenerateConfig) (map[string]mixer.Mixin, error) {
	mixed, err := generateRulesAlerts(cfg.RulesAlertsCfgs, mixer.NewRulesAlertsMixin)
	if err != nil {
		return nil, err
	}
	dashboards, err := generateDashboards(cfg.DashCfg)
	if err != nil {
		return nil, err
	}

	for dst, val := range dashboards {
		if other, ok := mixed[dst]; !ok {
			mixed[path.Join("dashboard", dst)] = val
		} else {
			mixed[dst] = bytes.Join([][]byte{val, other}, []byte("----\n"))
		}
	}

	return mixed, nil
}
