package checksum

import (
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "Empty data",
			data: []byte{},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "Hello World",
			data: []byte("Hello World"),
			want: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name: "Test data",
			data: []byte("The quick brown fox jumps over the lazy dog"),
			want: "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSHA256(tt.data)
			if got != tt.want {
				t.Errorf("CalculateSHA256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("Hello World")
	correctSum := "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"
	wrongSum := "0000000000000000000000000000000000000000000000000000000000000000"

	tests := []struct {
		name     string
		data     []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "Correct checksum",
			data:     data,
			expected: correctSum,
			wantErr:  false,
		},
		{
			name:     "Correct checksum uppercase",
			data:     data,
			expected: "A591A6D40BF420404A011733CFB7B190D62C65BF0BCDA32B57B277D9AD9F146E",
			wantErr:  false,
		},
		{
			name:     "Wrong checksum",
			data:     data,
			expected: wrongSum,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyChecksum(tt.data, tt.expected)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyChecksum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseChecksumFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
	}{
		{
			name: "Standard sha256sum format",
			content: `a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e  file1.tar.gz
d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592  file2.deb`,
			want: map[string]string{
				"file1.tar.gz": "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
				"file2.deb":    "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592",
			},
		},
		{
			name:    "Binary mode (asterisk)",
			content: `a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e *file1.tar.gz`,
			want: map[string]string{
				"file1.tar.gz": "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
			},
		},
		{
			name: "With comments and empty lines",
			content: `# SHA256 checksums
a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e  file1.tar.gz

d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592  file2.deb`,
			want: map[string]string{
				"file1.tar.gz": "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
				"file2.deb":    "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592",
			},
		},
		{
			name:    "Empty content",
			content: "",
			want:    map[string]string{},
		},
		{
			name:    "Invalid format",
			content: "not a valid checksum file",
			want:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseChecksumFile(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("parseChecksumFile() returned %d items, want %d", len(got), len(tt.want))
			}
			for filename, checksum := range tt.want {
				if got[filename] != checksum {
					t.Errorf("parseChecksumFile()[%q] = %q, want %q", filename, got[filename], checksum)
				}
			}
		})
	}
}
