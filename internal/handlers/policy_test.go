package handlers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"invalid json", "not json", ""},
		{"already compact", `{"a":1}`, `{"a":1}`},
		{"with whitespace", `{  "a" : 1  }`, `{"a":1}`},
		{"sorts keys", `{"b":2,"a":1}`, `{"a":1,"b":2}`},
		{"nested sorting", `{"z":{"b":2,"a":1},"a":0}`, `{"a":0,"z":{"a":1,"b":2}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, canonicalJSON(tt.input))
		})
	}
}

func TestDetectPolicyType(t *testing.T) {
	bucket := "my-bucket"

	publicReadPolicy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket)

	publicReadWritePolicy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket)

	tests := []struct {
		name       string
		policyJSON string
		bucketName string
		expected   string
	}{
		{
			name:       "empty policy is private",
			policyJSON: "",
			bucketName: bucket,
			expected:   "private",
		},
		{
			name:       "invalid json is custom",
			policyJSON: "not valid json",
			bucketName: bucket,
			expected:   "custom",
		},
		{
			name:       "public-read preset",
			policyJSON: publicReadPolicy,
			bucketName: bucket,
			expected:   "public-read",
		},
		{
			name:       "public-read-write preset",
			policyJSON: publicReadWritePolicy,
			bucketName: bucket,
			expected:   "public-read-write",
		},
		{
			name:       "public-read with different formatting",
			policyJSON: fmt.Sprintf(`{"Statement":[{"Action":["s3:GetObject"],"Effect":"Allow","Principal":{"AWS":["*"]},"Resource":["arn:aws:s3:::%s/*"]}],"Version":"2012-10-17"}`, bucket),
			bucketName: bucket,
			expected:   "public-read",
		},
		{
			name:       "wrong bucket name is custom",
			policyJSON: publicReadPolicy,
			bucketName: "other-bucket",
			expected:   "custom",
		},
		{
			name: "custom policy with s3:GetObject should not match public-read",
			policyJSON: fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    },
    {
      "Effect": "Deny",
      "Principal": {"AWS": ["arn:aws:iam::123456:user/bad"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/secret/*"]
    }
  ]
}`, bucket, bucket),
			bucketName: bucket,
			expected:   "custom",
		},
		{
			name: "custom policy with extra actions is custom",
			policyJSON: fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject", "s3:ListBucket"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket),
			bucketName: bucket,
			expected:   "custom",
		},
		{
			name: "custom policy with restricted principal is custom",
			policyJSON: fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["arn:aws:iam::123456:root"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket),
			bucketName: bucket,
			expected:   "custom",
		},
		{
			name: "custom policy with condition is custom",
			policyJSON: fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"],
      "Condition": {"IpAddress": {"aws:SourceIp": "192.168.1.0/24"}}
    }
  ]
}`, bucket),
			bucketName: bucket,
			expected:   "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectPolicyType(tt.policyJSON, tt.bucketName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
