/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resource

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"k8s.io/kubernetes/pkg/util/exec"
)

type ExecTransformer struct {
	runner  exec.Interface
	Command string
	Args    []string
}

func NewExecTransformer(spec string) (StreamTransform, error) {
	parts := strings.Split(spec, " ")
	if len(parts) == 0 {
		return nil, fmt.Errorf("expected at least one command")
	}
	transform := &ExecTransformer{Command: parts[0], Args: parts[1:]}
	return transform.Transform, nil
}

func (e *ExecTransformer) Transform(in io.Reader) (io.Reader, error) {
	runner := e.runner
	if runner == nil {
		runner = exec.New()
	}
	cmd := runner.Command(e.Command, e.Args...)
	cmd.SetStdin(in)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(data), nil
}
