package merkle

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func testHashFunc(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func TestBuildTree(t *testing.T) {
	tree := NewTree(testHashFunc)

	hashes := []string{"hash1", "hash2", "hash3", "hash4"}
	err := tree.Build(hashes)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	root, err := tree.RootHash()
	if err != nil {
		t.Fatalf("RootHash() failed: %v", err)
	}

	if root == "" {
		t.Fatal("Root hash is empty")
	}

	t.Logf("Root hash: %s", root)
}

func TestBuildEmptyTree(t *testing.T) {
	tree := NewTree(testHashFunc)
	err := tree.Build([]string{})

	if err != ErrEmptyLeaves {
		t.Fatalf("Expected ErrEmptyLeaves, got: %v", err)
	}
}

func TestSingleLeaf(t *testing.T) {
	tree := NewTree(testHashFunc)
	err := tree.Build([]string{"onlyhash"})
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	root, _ := tree.RootHash()
	if root != "onlyhash" {
		t.Fatalf("Single leaf root should equal leaf hash")
	}
}

func TestProofGeneration(t *testing.T) {
	tree := NewTree(testHashFunc)
	hashes := []string{"h0", "h1", "h2", "h3"}
	tree.Build(hashes)

	for i := 0; i < len(hashes); i++ {
		proof, err := tree.GenerateProof(i)
		if err != nil {
			t.Fatalf("GenerateProof(%d) failed: %v", i, err)
		}

		if proof.LeafHash != hashes[i] {
			t.Fatal("Wrong leaf hash in proof")
		}

		if len(proof.Siblings) != 2 {
			t.Fatalf("Expected 2 siblings, got %d", len(proof.Siblings))
		}
	}
}

func TestProofVerification(t *testing.T) {
	tree := NewTree(testHashFunc)
	hashes := []string{"h0", "h1", "h2", "h3"}
	tree.Build(hashes)

	for i := 0; i < len(hashes); i++ {
		proof, _ := tree.GenerateProof(i)

		if !proof.Verify(testHashFunc) {
			t.Fatalf("Proof for index %d failed verification", i)
		}
	}
}

func TestTamperedProof(t *testing.T) {
	tree := NewTree(testHashFunc)
	hashes := []string{"h0", "h1", "h2", "h3"}
	tree.Build(hashes)

	proof, _ := tree.GenerateProof(0)

	proof.LeafHash = "tampered"

	if proof.Verify(testHashFunc) {
		t.Fatal("Tampered proof should fail verification")
	}
}

func TestOddNumberOfLeaves(t *testing.T) {
	tree := NewTree(testHashFunc)
	hashes := []string{"h0", "h1", "h2"}
	err := tree.Build(hashes)

	if err != nil {
		t.Fatalf("Build with odd leaves failed: %v", err)
	}

	for i := 0; i < len(hashes); i++ {
		proof, _ := tree.GenerateProof(i)
		if !proof.Verify(testHashFunc) {
			t.Fatalf("Proof for index %d failed with odd leaves", i)
		}
	}
}
