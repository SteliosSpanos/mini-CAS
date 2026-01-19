package objects

import (
	"bytes"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "simple string",
			data: []byte("hello"),
			want: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name: "empty data",
			data: []byte{},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob := Blob{Data: tt.data}
			got := Hash(blob)

			if got != tt.want {
				t.Errorf("Hash() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewBlob(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "simple string",
			data: []byte("hello"),
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "nil data",
			data: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob := NewBlob(tt.data)

			if blob == nil {
				t.Fatal("NewBlob() returned nil")
			}

			if !bytes.Equal(blob.Data, tt.data) {
				t.Errorf("NewBlob().Data = %v, want %v", blob.Data, tt.data)
			}
		})
	}
}
