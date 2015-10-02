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
	"io/ioutil"
	"testing"

	"k8s.io/kubernetes/pkg/util/exec"
)

func TestExecTransformer(t *testing.T) {
	tests := []struct {
		cmd  string
		args []string
	}{
		{"echo", []string{"a", "b", "c"}},
		{"echo", []string{}},
	}
	for _, test := range tests {
		fakeExec := exec.FakeExec{}
		transform := ExecTransformer{
			runner:  &fakeExec,
			Command: test.cmd,
			Args:    test.args,
		}

		output := "output"
		fakeCmd := exec.FakeCmd{
			CombinedOutputScript: []exec.FakeCombinedOutputAction{
				func() ([]byte, error) {
					return []byte(output), nil
				},
			},
		}

		fakeExec.CommandScript = []exec.FakeCommandAction{
			func(receivedCmd string, receivedArgs ...string) exec.Cmd {
				if test.cmd != receivedCmd {
					t.Errorf("unexpected command: %s", test.cmd)
				}
				if len(receivedArgs) != len(test.args) {
					t.Errorf("unexpected args: %v vs %v", receivedArgs, test.args)
				}
				for ix := range receivedArgs {
					if receivedArgs[ix] != test.args[ix] {
						t.Errorf("unexpected args: %v vs %v", receivedArgs, test.args)
					}
				}
				return &fakeCmd
			},
		}

		input := "foobar"
		stdin := bytes.NewBuffer([]byte(input))
		out, err := transform.Transform(stdin)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		outBytes, err := ioutil.ReadAll(out)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(outBytes) != output {
			t.Errorf("expected: %s, saw: %s", output, string(outBytes))
		}
		in, err := ioutil.ReadAll(fakeCmd.Stdin)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(in) != input {
			t.Errorf("expected: %s, saw: %s", input, string(in))
		}
	}
}
