/*
Copyright 2026 The Kubernetes Authors.

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

package main

import (
	"os"
	"strings"
	"testing"

	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
)

func newTestCDIHandler(t *testing.T) (*CDIHandler, string) {
	t.Helper()
	dir := t.TempDir()
	cdi, err := NewCDIHandler(&Config{flags: &Flags{cdiRoot: dir}})
	if err != nil {
		t.Fatalf("NewCDIHandler: %v", err)
	}
	return cdi, dir
}

func TestGetClaimDevices(t *testing.T) {
	cdi, _ := newTestCDIHandler(t)

	got := cdi.GetClaimDevices("claim-uid", []string{"dev0", "dev1"})

	want := []string{
		cdiKind + "=" + cdiCommonDeviceName,
		cdiKind + "=claim-uid-dev0",
		cdiKind + "=claim-uid-dev1",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d devices, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("device[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCreateAndDeleteClaimSpecFile(t *testing.T) {
	cdi, dir := newTestCDIHandler(t)

	devices := PreparedDevices{
		{Device: kubeletplugin.Device{DeviceName: "dev0"}, ContainerEdits: &cdiapi.ContainerEdits{}},
	}
	if err := cdi.CreateClaimSpecFile("claim-uid", devices); err != nil {
		t.Fatalf("CreateClaimSpecFile: %v", err)
	}

	if countSpecFiles(t, dir) == 0 {
		t.Fatal("expected a spec file to be written")
	}

	if err := cdi.DeleteClaimSpecFile("claim-uid"); err != nil {
		t.Fatalf("DeleteClaimSpecFile: %v", err)
	}
	if n := countSpecFiles(t, dir); n != 0 {
		t.Errorf("expected spec files removed, found %d", n)
	}
}

func TestCreateCommonSpecFile(t *testing.T) {
	cdi, dir := newTestCDIHandler(t)

	if err := cdi.CreateCommonSpecFile(map[string]string{"FOO": "bar"}, nil); err != nil {
		t.Fatalf("CreateCommonSpecFile: %v", err)
	}
	if countSpecFiles(t, dir) == 0 {
		t.Error("expected a common spec file to be written")
	}
}

func countSpecFiles(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	n := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") || strings.HasSuffix(e.Name(), ".yaml") {
			n++
		}
	}
	return n
}
