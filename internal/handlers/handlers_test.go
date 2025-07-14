package handlers

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"locals3/internal/storage"
)

// MockStorage is a mock implementation of the Storage interface for testing
type MockStorage struct {
	Buckets map[string]storage.BucketInfo
	Objects map[string]map[string]*storage.ObjectInfo
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		Buckets: make(map[string]storage.BucketInfo),
		Objects: make(map[string]map[string]*storage.ObjectInfo),
	}
}

// Implement Storage interface methods for testing
func (m *MockStorage) CreateBucket(bucket string) error {
	m.Buckets[bucket] = storage.BucketInfo{
		Name:         bucket,
		CreationDate: time.Now(),
	}
	m.Objects[bucket] = make(map[string]*storage.ObjectInfo)
	return nil
}

func (m *MockStorage) DeleteBucket(bucket string) error {
	delete(m.Buckets, bucket)
	delete(m.Objects, bucket)
	return nil
}

func (m *MockStorage) ListBuckets() ([]storage.BucketInfo, error) {
	buckets := make([]storage.BucketInfo, 0, len(m.Buckets))
	for _, bucket := range m.Buckets {
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}

func (m *MockStorage) BucketExists(bucket string) bool {
	_, ok := m.Buckets[bucket]
	return ok
}

// Implement remaining methods with minimal functionality needed for tests
func (m *MockStorage) PutObject(bucket, key string, data io.Reader, size int64, metadata map[string]string) (*storage.ObjectInfo, error) {
	if !m.BucketExists(bucket) {
		return nil, fmt.Errorf("bucket does not exist")
	}

	objInfo := &storage.ObjectInfo{
		Key:          key,
		Size:         size,
		ETag:         "test-etag",
		LastModified: time.Now(),
		ContentType:  metadata["Content-Type"],
		Metadata:     metadata,
	}

	m.Objects[bucket][key] = objInfo
	return objInfo, nil
}

func (m *MockStorage) GetObject(bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
	if !m.BucketExists(bucket) {
		return nil, nil, fmt.Errorf("bucket does not exist")
	}

	objInfo, ok := m.Objects[bucket][key]
	if !ok {
		return nil, nil, fmt.Errorf("object does not exist")
	}

	// Return empty reader and object info
	return io.NopCloser(strings.NewReader("test content")), objInfo, nil
}

func (m *MockStorage) DeleteObject(bucket, key string) error {
	if !m.BucketExists(bucket) {
		return fmt.Errorf("bucket does not exist")
	}

	delete(m.Objects[bucket], key)
	return nil
}

func (m *MockStorage) ListObjects(bucket, prefix, delimiter, marker string, maxKeys int) (*storage.ListObjectsResult, error) {
	if !m.BucketExists(bucket) {
		return nil, fmt.Errorf("bucket does not exist")
	}

	objects := make([]storage.ObjectInfo, 0)
	for k, v := range m.Objects[bucket] {
		if strings.HasPrefix(k, prefix) {
			objects = append(objects, *v)
		}
	}

	return &storage.ListObjectsResult{
		Objects:     objects,
		IsTruncated: false,
	}, nil
}

func (m *MockStorage) HeadObject(bucket, key string) (*storage.ObjectInfo, error) {
	if !m.BucketExists(bucket) {
		return nil, fmt.Errorf("bucket does not exist")
	}

	objInfo, ok := m.Objects[bucket][key]
	if !ok {
		return nil, fmt.Errorf("object does not exist")
	}

	return objInfo, nil
}

func (m *MockStorage) ObjectExists(bucket, key string) bool {
	if !m.BucketExists(bucket) {
		return false
	}

	_, ok := m.Objects[bucket][key]
	return ok
}

// Implement multipart methods (minimal stubs for now)
func (m *MockStorage) InitiateMultipartUpload(bucket, key string, metadata map[string]string) (string, error) {
	return "test-upload-id", nil
}

func (m *MockStorage) UploadPart(bucket, key, uploadID string, partNumber int, data io.Reader, size int64) (*storage.PartInfo, error) {
	return &storage.PartInfo{
		PartNumber: partNumber,
		ETag:       "test-etag",
		Size:       size,
	}, nil
}

func (m *MockStorage) CompleteMultipartUpload(bucket, key, uploadID string, parts []storage.CompletePart) (*storage.ObjectInfo, error) {
	objInfo := &storage.ObjectInfo{
		Key:          key,
		Size:         100,
		ETag:         "test-etag",
		LastModified: time.Now(),
	}

	m.Objects[bucket][key] = objInfo
	return objInfo, nil
}

func (m *MockStorage) AbortMultipartUpload(bucket, key, uploadID string) error {
	return nil
}

// MockAuth is a mock auth provider for testing
type MockAuth struct {
	AccessKey string
	SecretKey string
	Region    string
}

func NewMockAuth() *MockAuth {
	return &MockAuth{
		AccessKey: "test",
		SecretKey: "test123456789",
		Region:    "test-region",
	}
}

func (m *MockAuth) Authenticate(r *http.Request) error {
	return nil // Always succeed in tests
}

func (m *MockAuth) GetAccessKey() string {
	return m.AccessKey
}

func (m *MockAuth) GetSecretKey() string {
	return m.SecretKey
}

func (m *MockAuth) GetRegion() string {
	return m.Region
}

// Test functions
func TestNew(t *testing.T) {
	storage := NewMockStorage()
	auth := NewMockAuth()

	cfg := &Config{
		Storage:     storage,
		Auth:        auth,
		Region:      "test-region",
		BaseDomain:  "localhost",
		DisableAuth: false,
	}

	handler := New(cfg)

	if handler.region != "test-region" {
		t.Errorf("Expected region 'test-region', got '%s'", handler.region)
	}

	if handler.baseDomain != "localhost" {
		t.Errorf("Expected base domain 'localhost', got '%s'", handler.baseDomain)
	}

	if handler.disableAuth != false {
		t.Errorf("Expected disableAuth false, got %t", handler.disableAuth)
	}
}

func TestHealthCheck(t *testing.T) {
	storage := NewMockStorage()
	auth := NewMockAuth()

	cfg := &Config{
		Storage:     storage,
		Auth:        auth,
		Region:      "test-region",
		BaseDomain:  "localhost",
		DisableAuth: true,
	}

	handler := New(cfg)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(handler.HealthCheck).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	expected := `{"status":"ok"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Handler returned unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestListBuckets(t *testing.T) {
	storage := NewMockStorage()
	auth := NewMockAuth()

	// Create test buckets
	storage.CreateBucket("test-bucket-1")
	storage.CreateBucket("test-bucket-2")

	cfg := &Config{
		Storage:     storage,
		Auth:        auth,
		Region:      "test-region",
		BaseDomain:  "localhost",
		DisableAuth: true, // Disable auth for testing
	}

	handler := New(cfg)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(handler.ListBuckets).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Parse XML response
	var result ListAllMyBucketsResult
	err = xml.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(result.Buckets.Bucket) != 2 {
		t.Errorf("Expected 2 buckets, got %d", len(result.Buckets.Bucket))
	}

	// Check bucket names
	bucketNames := make(map[string]bool)
	for _, b := range result.Buckets.Bucket {
		bucketNames[b.Name] = true
	}

	if !bucketNames["test-bucket-1"] || !bucketNames["test-bucket-2"] {
		t.Errorf("Expected buckets test-bucket-1 and test-bucket-2, got %v", bucketNames)
	}
}
