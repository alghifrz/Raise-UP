package password_test

import (
	"testing"

	"github.com/diuk/raiseup/pkg/password"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := password.Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" || hash == "secret-password" {
		t.Fatal("Hash() returned plaintext or empty value")
	}
	if !password.Verify(hash, "secret-password") {
		t.Fatal("Verify() = false, want true for matching password")
	}
	if password.Verify(hash, "wrong-password") {
		t.Fatal("Verify() = true, want false for mismatched password")
	}
}
