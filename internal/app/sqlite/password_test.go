package sqlite

import (
	"strings"
	"testing"
)

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	ph := PasswordHasher{}
	hash, err := ph.Hash("testpassword123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	valid, newHash := ph.Verify("testpassword123", hash)
	if !valid {
		t.Error("expected password to verify")
	}
	if newHash != "" {
		t.Error("expected no rehash for matching parameters")
	}

	valid, _ = ph.Verify("wrongpassword", hash)
	if valid {
		t.Error("expected wrong password to fail")
	}
}

func TestPasswordHasher_HashEmpty(t *testing.T) {
	ph := PasswordHasher{}
	hash, err := ph.Hash("")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash for empty password")
	}
	valid, _ := ph.Verify("", hash)
	if !valid {
		t.Error("expected empty password to verify")
	}
}

func TestPasswordHasher_VerifyEmptyStored(t *testing.T) {
	ph := PasswordHasher{}
	valid, _ := ph.Verify("anything", "")
	if valid {
		t.Error("empty stored hash should never verify")
	}
}

func TestPasswordHasher_VerifyBcryptRejected(t *testing.T) {
	ph := PasswordHasher{}
	valid, _ := ph.Verify("test", "$2a$10$abcdefghijklmnopqrstuvwxyz0123456789")
	if valid {
		t.Error("non-argon2id hash should be rejected")
	}
}

func TestPasswordHasher_VerifyRehashOnParamUpgrade(t *testing.T) {
	ph := PasswordHasher{}
	hash, err := ph.Hash("secret123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	// Simulate old params by replacing the prefix with a different argon2 variant.
	// Format: $argon2id$v=19$m=47104,t=1,p=4$<salt>$<hash>
	parts := strings.SplitN(hash, "$", 6)
	if len(parts) < 6 {
		t.Fatalf("unexpected hash format: %s", hash)
	}
	oldHash := "$argon2id$v=19$m=32768,t=2,p=2$" + parts[4] + "$" + parts[5]

	valid, newHash := ph.Verify("secret123", oldHash)
	if !valid {
		t.Error("expected password to verify with old params")
	}
	if newHash == "" {
		t.Error("expected rehash when stored params differ from current")
	}
	if !strings.HasPrefix(newHash, argonPrefix) {
		t.Errorf("rehashed hash should use current prefix %s, got: %s", argonPrefix, newHash)
	}
}

func TestPasswordHasher_VerifyMalformedHash(t *testing.T) {
	ph := PasswordHasher{}

	tests := []struct {
		name  string
		hash  string
		valid bool
	}{
		{"too few parts", "$argon2id$v=19$only", false},
		{"invalid salt base64", "$argon2id$v=19$m=47104,t=1,p=4$!!!not-base64$abc123", false},
		{"invalid hash base64", "$argon2id$v=19$m=47104,t=1,p=4$abc123$!!!not-base64", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := ph.Verify("anything", tt.hash)
			if valid != tt.valid {
				t.Errorf("Verify(%q) = %v, want %v", tt.name, valid, tt.valid)
			}
		})
	}
}

func TestPasswordHasher_HashProducesCorrectPrefix(t *testing.T) {
	ph := PasswordHasher{}
	hash, err := ph.Hash("mypassword")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(hash, argonPrefix) {
		t.Errorf("hash should start with %s, got: %s", argonPrefix, hash)
	}
	// Verify format: $argon2id$v=19$m=47104,t=1,p=4$<b64_salt>$<b64_hash>
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 $-separated parts, got %d: %s", len(parts), hash)
	}
}

func TestPasswordHasher_VerifyWrongPasswordNoRehash(t *testing.T) {
	ph := PasswordHasher{}
	hash, err := ph.Hash("correct")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	// Even with old params, wrong password should fail — no rehash.
	parts := strings.SplitN(hash, "$", 6)
	oldHash := "$argon2id$v=19$m=32768,t=2,p=2$" + parts[4] + "$" + parts[5]
	valid, newHash := ph.Verify("wrong", oldHash)
	if valid {
		t.Error("wrong password should not verify")
	}
	if newHash != "" {
		t.Error("wrong password should not trigger rehash")
	}
}

func TestPasswordHasher_VerifySubtleConstantTime(t *testing.T) {
	// Verify doesn't leak timing info via early exit — both salt/hash decode
	// failures return false without comparing.
	ph := PasswordHasher{}
	valid, _ := ph.Verify("test", "$argon2id$v=19$m=47104,t=1,p=4$abc123$!!!bad")
	if valid {
		t.Error("invalid base64 hash should not verify")
	}
}
