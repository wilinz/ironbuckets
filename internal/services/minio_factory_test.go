package services

import "testing"

func TestShouldUseSSL_Localhost(t *testing.T) {
	tests := []struct {
		endpoint string
		want     bool
	}{
		{"localhost:9000", false},
		{"127.0.0.1:9000", false},
		{"minio:9000", false},
		{"play.minio.io:9000", true},
		{"s3.amazonaws.com", true},
		{"minio.example.com:9000", true},
		{"localhost:9001", true}, // Different port
		{"192.168.1.100:9000", true},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			got := shouldUseSSL(tt.endpoint)
			if got != tt.want {
				t.Errorf("shouldUseSSL(%q) = %v, want %v", tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestCredentials_Struct(t *testing.T) {
	creds := Credentials{
		Endpoint:     "localhost:9000",
		AccessKey:    "admin",
		SecretKey:    "password",
		SessionToken: "token",
	}

	if creds.Endpoint != "localhost:9000" {
		t.Errorf("unexpected Endpoint: %s", creds.Endpoint)
	}
	if creds.AccessKey != "admin" {
		t.Errorf("unexpected AccessKey: %s", creds.AccessKey)
	}
	if creds.SecretKey != "password" {
		t.Errorf("unexpected SecretKey: %s", creds.SecretKey)
	}
	if creds.SessionToken != "token" {
		t.Errorf("unexpected SessionToken: %s", creds.SessionToken)
	}
}

func TestRealMinioFactory_Implements_Interface(t *testing.T) {
	// Compile-time check that RealMinioFactory implements MinioClientFactory
	var _ MinioClientFactory = (*RealMinioFactory)(nil)
}
