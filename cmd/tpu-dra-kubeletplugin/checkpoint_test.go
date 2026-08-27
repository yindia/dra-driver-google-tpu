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
	"testing"
)

func TestNewCheckpoint(t *testing.T) {
	cp := newCheckpoint()
	if cp.Checksum != 0 {
		t.Errorf("expected zero checksum, got %d", cp.Checksum)
	}
	if cp.V1 == nil {
		t.Fatal("expected V1 to be initialized")
	}
	if cp.V1.PreparedClaims == nil {
		t.Error("expected PreparedClaims map to be initialized")
	}
}

func TestCheckpointMarshalRoundTrip(t *testing.T) {
	cp := newCheckpoint()
	cp.V1.PreparedClaims["claim-1"] = PreparedDevices{}

	data, err := cp.MarshalCheckpoint()
	if err != nil {
		t.Fatalf("MarshalCheckpoint: %v", err)
	}
	if cp.Checksum == 0 {
		t.Error("expected checksum to be set after marshal")
	}

	got := newCheckpoint()
	if err := got.UnmarshalCheckpoint(data); err != nil {
		t.Fatalf("UnmarshalCheckpoint: %v", err)
	}
	if _, ok := got.V1.PreparedClaims["claim-1"]; !ok {
		t.Error("expected claim-1 to survive round trip")
	}
	if err := got.VerifyChecksum(); err != nil {
		t.Errorf("VerifyChecksum on valid data: %v", err)
	}
}

func TestVerifyChecksumDetectsTampering(t *testing.T) {
	cp := newCheckpoint()
	data, err := cp.MarshalCheckpoint()
	if err != nil {
		t.Fatalf("MarshalCheckpoint: %v", err)
	}

	got := newCheckpoint()
	if err := got.UnmarshalCheckpoint(data); err != nil {
		t.Fatalf("UnmarshalCheckpoint: %v", err)
	}
	// Mutate after the checksum was computed.
	got.V1.PreparedClaims["injected"] = PreparedDevices{}
	if err := got.VerifyChecksum(); err == nil {
		t.Error("expected checksum verification to fail on tampered data")
	}
}

func TestVerifyChecksumPreservesStoredChecksum(t *testing.T) {
	cp := newCheckpoint()
	if _, err := cp.MarshalCheckpoint(); err != nil {
		t.Fatalf("MarshalCheckpoint: %v", err)
	}
	before := cp.Checksum
	_ = cp.VerifyChecksum()
	if cp.Checksum != before {
		t.Errorf("VerifyChecksum mutated stored checksum: before=%d after=%d", before, cp.Checksum)
	}
}
