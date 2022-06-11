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

// Ignore deprecated warnings until we fix https://github.com/monitoring-mixins/mixtool/issues/22
//nolint:staticcheck
package mixer

import (
	"github.com/google/go-jsonnet"
	"github.com/grafana/tanka/pkg/jsonnet/native"
)

type Evaluator interface {
	Exec(mixin Mixin) ([]byte, error)
}

type eval struct {
	vm *jsonnet.VM
}

func NewDefaultEvaluator() Evaluator {
	return &eval{
		vm: jsonnet.MakeVM(),
	}
}

func NewEvaluator(jpath []string) Evaluator {
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: jpath,
	})
	for _, nf := range native.Funcs() {
		vm.NativeFunction(nf)
	}
	return &eval{
		vm: vm,
	}
}

func (e eval) Exec(mixin Mixin) ([]byte, error) {
	out, err := e.vm.EvaluateSnippet("", string(mixin))
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}
