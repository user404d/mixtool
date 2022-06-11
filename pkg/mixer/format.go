package mixer

import "sigs.k8s.io/yaml"

type Formatter func(content Mixin) (Mixin, error)

func NoFormatter(content Mixin) (Mixin, error) {
	return content, nil
}

func JSONtoYaml(content Mixin) (Mixin, error) {
	return yaml.JSONToYAML(content)
}
