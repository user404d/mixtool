package mixer

import "fmt"

const (
	importFormat = `
local mixin = (import %q);
`

	alertsFormat = `
if std.objectHasAll(mixin, "%[1]sAlerts")
then mixin.%[1]sAlerts
else {}
`

	rulesFormat = `
if std.objectHasAll(mixin, "%[1]sRules")
then mixin.%[1]sRules
else {}
`

	rulesAlertsFormat = `
if std.objectHasAll(mixin, "%[1]sRules") && std.objectHasAll(mixin, "%[1]sAlerts")
then mixin.%[1]sRules + mixin.%[1]sAlerts
else if std.objectHasAll(mixin, "%[1]sRules")
then mixin.%[1]sRules 
else if std.objectHasAll(mixin, "%[1]sAlerts")
then mixin.%[1]sAlerts
else {}
`
	dashboards = `
if std.objectHasAll(mixin, "grafanaDashboards")
then mixin.grafanaDashboards
else {}
`
)

type DataSource string

const (
	Loki       DataSource = "loki"
	Prometheus DataSource = "prometheus"
)

type RulesAlertsOptions struct {
	DataSource DataSource
	ImportPath string
}

type Mixin []byte

type MixinBuilder func(opts *RulesAlertsOptions) Mixin

func NewAlertsMixin(opts *RulesAlertsOptions) Mixin {
	return Mixin(fmt.Sprintf(importFormat, opts.ImportPath) + fmt.Sprintf(alertsFormat, opts.DataSource))
}

func NewRulesMixin(opts *RulesAlertsOptions) Mixin {
	return Mixin(fmt.Sprintf(importFormat, opts.ImportPath) + fmt.Sprintf(rulesFormat, opts.DataSource))
}

func NewRulesAlertsMixin(opts *RulesAlertsOptions) Mixin {
	return Mixin(fmt.Sprintf(importFormat, opts.ImportPath) + fmt.Sprintf(rulesAlertsFormat, opts.DataSource))
}

type DashboardsOptions struct {
	ImportPath string
}

func NewDashboardsMixin(opts *DashboardsOptions) Mixin {
	return Mixin(fmt.Sprintf(importFormat, opts.ImportPath) + dashboards)
}

func (m Mixin) ApplyFormatter(f Formatter) (Mixin, error) {
	return f(m)
}
