package auth

import (
	"net/http"
	"testing"
)

func TestNewAWSV4Auth(t *testing.T) {
	accessKey := "test-access"
	secretKey := "test-secret"
	region := "test-region"

	auth := NewAWSV4Auth(accessKey, secretKey, region)

	if auth.GetAccessKey() != accessKey {
		t.Errorf("Expected access key %s, got %s", accessKey, auth.GetAccessKey())
	}

	if auth.GetSecretKey() != secretKey {
		t.Errorf("Expected secret key %s, got %s", secretKey, auth.GetSecretKey())
	}

	if auth.GetRegion() != region {
		t.Errorf("Expected region %s, got %s", region, auth.GetRegion())
	}
}

func TestParseAuthHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected map[string]string
	}{
		{
			name:   "Valid header",
			header: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-date, Signature=aabc1234",
			expected: map[string]string{
				"Credential":    "AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request",
				"SignedHeaders": "host;range;x-amz-date",
				"Signature":     "aabc1234",
			},
		},
		{
			name:     "Invalid header",
			header:   "Invalid Authorization Header",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAuthHeader(tt.header)

			// Check result matches expected
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestAuthenticate_MissingHeader(t *testing.T) {
	auth := NewAWSV4Auth("test", "test", "us-east-1")
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	err := auth.Authenticate(req)

	if err == nil {
		t.Error("Expected error for missing authorization header, got nil")
	}

	if err.Error() != "missing authorization header" {
		t.Errorf("Expected error message 'missing authorization header', got '%s'", err.Error())
	}
}

func TestAuthenticate_InvalidFormat(t *testing.T) {
	auth := NewAWSV4Auth("test", "test", "us-east-1")
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Authorization", "Invalid Format")

	err := auth.Authenticate(req)

	if err == nil {
		t.Error("Expected error for invalid header format, got nil")
	}

	if err.Error() != "invalid authorization header format" {
		t.Errorf("Expected error message 'invalid authorization header format', got '%s'", err.Error())
	}
}

func TestHmacSHA256(t *testing.T) {
	key := []byte("test-key")
	data := "test-data"

	// Expected HMAC-SHA256 of "test-data" with key "test-key"
	// This value was pre-calculated
	expected := []byte{0x5d, 0xde, 0xc1, 0xb9, 0x77, 0x8c, 0x9f, 0xed, 0xfa, 0xb8, 0x24, 0x65, 0x23, 0x92, 0x56, 0xe4, 0xc2, 0x54, 0xb6, 0xad, 0xf3, 0x76, 0x75, 0xc5, 0x39, 0x5, 0x36, 0xda, 0xdb, 0xc6, 0x73, 0x28}

	result := hmacSHA256(key, data)

	if len(result) != len(expected) {
		t.Errorf("Expected HMAC length %d, got %d", len(expected), len(result))
	}

	for i := 0; i < len(expected); i++ {
		if i < len(result) && result[i] != expected[i] {
			t.Errorf("HMAC mismatch at position %d: expected %x, got %x", i, expected[i], result[i])
		}
	}
}

func TestSha256Hex(t *testing.T) {
	data := "test-data"
	// Expected SHA256 hash of "test-data" in hex
	expected := "9e0e8a93105f51a967406ded7fb08f649a97376eb5f8ae4e2bec7d4e5b67feb8"

	result := sha256Hex(data)

	if result != expected {
		t.Errorf("Expected SHA256 hash %s, got %s", expected, result)
	}
}
