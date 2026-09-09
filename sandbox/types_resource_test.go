//go:build unit

package sandbox

import (
	"encoding/json"
	"testing"

	"github.com/qiniu/go-sdk/v7/sandbox/internal/apis"
)

func TestSandboxResourceSpecToAPIKodoResource(t *testing.T) {
	prefix := "datasets/"
	readOnly := true

	resource, err := sandboxResourceSpecToAPI(SandboxResourceSpec{
		Kodo: &KodoResource{
			Bucket:    "test-bucket",
			MountPath: "/mnt/kodo",
			Prefix:    &prefix,
			ReadOnly:  &readOnly,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(mustMarshalJSON(t, resource), &got); err != nil {
		t.Fatalf("unexpected JSON error: %v", err)
	}
	if got["type"] != "kodo" {
		t.Errorf("expected type kodo, got %v", got["type"])
	}
	if got["bucket"] != "test-bucket" {
		t.Errorf("expected bucket test-bucket, got %v", got["bucket"])
	}
	if got["mount_path"] != "/mnt/kodo" {
		t.Errorf("expected mount_path /mnt/kodo, got %v", got["mount_path"])
	}
	if got["prefix"] != "datasets/" {
		t.Errorf("expected prefix datasets/, got %v", got["prefix"])
	}
	if got["read_only"] != true {
		t.Errorf("expected read_only true, got %v", got["read_only"])
	}
	if _, ok := got["access_key"]; ok {
		t.Errorf("expected access_key omitted, got %v", got["access_key"])
	}
	if _, ok := got["secret_key"]; ok {
		t.Errorf("expected secret_key omitted, got %v", got["secret_key"])
	}
}

func TestSandboxResourceSpecToAPIKodoResourceWithInlineCredentials(t *testing.T) {
	accessKey := "test-ak"
	secretKey := "test-sk"

	resource, err := sandboxResourceSpecToAPI(SandboxResourceSpec{
		Kodo: &KodoResource{
			AccessKey: &accessKey,
			Bucket:    "test-bucket",
			MountPath: "/mnt/kodo",
			SecretKey: &secretKey,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(mustMarshalJSON(t, resource), &got); err != nil {
		t.Fatalf("unexpected JSON error: %v", err)
	}
	if got["access_key"] != accessKey || got["secret_key"] != secretKey {
		t.Fatalf("expected inline credentials, got %#v", got)
	}
}

func TestSandboxResourceSpecToAPIKodoResourceValidation(t *testing.T) {
	_, err := sandboxResourceSpecToAPI(SandboxResourceSpec{
		Kodo: &KodoResource{MountPath: "/mnt/kodo"},
	})
	if err == nil {
		t.Fatal("expected error when KodoResource.Bucket is empty")
	}

	_, err = sandboxResourceSpecToAPI(SandboxResourceSpec{
		Kodo: &KodoResource{Bucket: "test-bucket"},
	})
	if err == nil {
		t.Fatal("expected error when KodoResource.MountPath is empty")
	}

	accessKey := "test-ak"
	_, err = sandboxResourceSpecToAPI(SandboxResourceSpec{
		Kodo: &KodoResource{AccessKey: &accessKey, Bucket: "test-bucket", MountPath: "/mnt/kodo"},
	})
	if err == nil {
		t.Fatal("expected error when only one Kodo credential is provided")
	}
}

func TestMaskedSandboxResourceFromAPIKodoResource(t *testing.T) {
	resourceID := "res_kodo"
	prefix := "datasets/"
	readOnly := true
	resource := apis.SandboxResource{}
	if err := resource.FromKodoResource(apis.KodoResource{
		Bucket:     "test-bucket",
		MountPath:  "/mnt/kodo",
		Prefix:     &prefix,
		ReadOnly:   &readOnly,
		ResourceID: &resourceID,
	}); err != nil {
		t.Fatal(err)
	}

	masked, err := maskedSandboxResourceFromAPI(resource)
	if err != nil {
		t.Fatal(err)
	}
	if masked.Kodo == nil || masked.Kodo.ResourceID != resourceID || masked.Kodo.Bucket != "test-bucket" || masked.Kodo.Prefix == nil || *masked.Kodo.Prefix != prefix || masked.Kodo.ReadOnly == nil || *masked.Kodo.ReadOnly != readOnly {
		t.Fatalf("unexpected masked Kodo resource: %#v", masked.Kodo)
	}
}

func mustMarshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected JSON marshal error: %v", err)
	}
	return data
}
