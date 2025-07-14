package storage

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func setupTestStorage(t *testing.T) (*FileSystemStorage, string) {
	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "locals3-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	storage := NewFileSystemStorage(tempDir)
	return storage, tempDir
}

func cleanupTestStorage(tempDir string) {
	os.RemoveAll(tempDir)
}

func TestNewFileSystemStorage(t *testing.T) {
	basePath := "/tmp/test-path"
	fs := NewFileSystemStorage(basePath)

	if fs.basePath != basePath {
		t.Errorf("Expected basePath %s, got %s", basePath, fs.basePath)
	}
}

func TestBucketOperations(t *testing.T) {
	fs, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(tempDir)

	// Test bucket creation
	bucketName := "test-bucket"
	err := fs.CreateBucket(bucketName)
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	// Test bucket exists
	if !fs.BucketExists(bucketName) {
		t.Errorf("Expected bucket %s to exist", bucketName)
	}

	// Test list buckets
	buckets, err := fs.ListBuckets()
	if err != nil {
		t.Fatalf("Failed to list buckets: %v", err)
	}
	if len(buckets) != 1 {
		t.Errorf("Expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].Name != bucketName {
		t.Errorf("Expected bucket name %s, got %s", bucketName, buckets[0].Name)
	}

	// Test bucket deletion
	err = fs.DeleteBucket(bucketName)
	if err != nil {
		t.Fatalf("Failed to delete bucket: %v", err)
	}

	if fs.BucketExists(bucketName) {
		t.Errorf("Expected bucket %s to not exist after deletion", bucketName)
	}
}

func TestObjectOperations(t *testing.T) {
	fs, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(tempDir)

	// Create a bucket first
	bucketName := "test-bucket"
	err := fs.CreateBucket(bucketName)
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	// Test PutObject
	objectKey := "test-key"
	objectContent := []byte("test content")
	objectSize := int64(len(objectContent))
	metadata := map[string]string{"Content-Type": "text/plain"}

	reader := bytes.NewReader(objectContent)
	objInfo, err := fs.PutObject(bucketName, objectKey, reader, objectSize, metadata)
	if err != nil {
		t.Fatalf("Failed to put object: %v", err)
	}
	if objInfo.Key != objectKey {
		t.Errorf("Expected object key %s, got %s", objectKey, objInfo.Key)
	}
	if objInfo.Size != objectSize {
		t.Errorf("Expected object size %d, got %d", objectSize, objInfo.Size)
	}

	// Test ObjectExists
	if !fs.ObjectExists(bucketName, objectKey) {
		t.Errorf("Expected object %s to exist", objectKey)
	}

	// Test HeadObject
	headInfo, err := fs.HeadObject(bucketName, objectKey)
	if err != nil {
		t.Fatalf("Failed to head object: %v", err)
	}
	if headInfo.Key != objectKey {
		t.Errorf("Expected object key %s, got %s", objectKey, headInfo.Key)
	}
	if headInfo.Size != objectSize {
		t.Errorf("Expected object size %d, got %d", objectSize, headInfo.Size)
	}

	// Test GetObject
	objReader, getInfo, err := fs.GetObject(bucketName, objectKey)
	if err != nil {
		t.Fatalf("Failed to get object: %v", err)
	}
	defer objReader.Close()

	data, err := io.ReadAll(objReader)
	if err != nil {
		t.Fatalf("Failed to read object data: %v", err)
	}

	if !bytes.Equal(data, objectContent) {
		t.Errorf("Object content mismatch. Expected %s, got %s", objectContent, data)
	}

	if getInfo.Key != objectKey {
		t.Errorf("Expected object key %s, got %s", objectKey, getInfo.Key)
	}

	// Test ListObjects
	listResult, err := fs.ListObjects(bucketName, "", "", "", 1000)
	if err != nil {
		t.Fatalf("Failed to list objects: %v", err)
	}

	if len(listResult.Objects) != 1 {
		t.Errorf("Expected 1 object, got %d", len(listResult.Objects))
	}

	if listResult.Objects[0].Key != objectKey {
		t.Errorf("Expected object key %s, got %s", objectKey, listResult.Objects[0].Key)
	}

	// Test DeleteObject
	err = fs.DeleteObject(bucketName, objectKey)
	if err != nil {
		t.Fatalf("Failed to delete object: %v", err)
	}

	if fs.ObjectExists(bucketName, objectKey) {
		t.Errorf("Expected object %s to not exist after deletion", objectKey)
	}
}

func TestMultipartOperations(t *testing.T) {
	fs, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(tempDir)

	// Create a bucket first
	bucketName := "test-bucket"
	err := fs.CreateBucket(bucketName)
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	// Test InitiateMultipartUpload
	objectKey := "multipart-key"
	metadata := map[string]string{"Content-Type": "application/octet-stream"}

	uploadID, err := fs.InitiateMultipartUpload(bucketName, objectKey, metadata)
	if err != nil {
		t.Fatalf("Failed to initiate multipart upload: %v", err)
	}

	if uploadID == "" {
		t.Error("Expected non-empty upload ID")
	}

	// Test UploadPart
	partContent1 := []byte("part 1 content")
	partContent2 := []byte("part 2 content")

	part1, err := fs.UploadPart(bucketName, objectKey, uploadID, 1, bytes.NewReader(partContent1), int64(len(partContent1)))
	if err != nil {
		t.Fatalf("Failed to upload part 1: %v", err)
	}

	part2, err := fs.UploadPart(bucketName, objectKey, uploadID, 2, bytes.NewReader(partContent2), int64(len(partContent2)))
	if err != nil {
		t.Fatalf("Failed to upload part 2: %v", err)
	}

	// Test CompleteMultipartUpload
	parts := []CompletePart{
		{PartNumber: 1, ETag: part1.ETag},
		{PartNumber: 2, ETag: part2.ETag},
	}

	objInfo, err := fs.CompleteMultipartUpload(bucketName, objectKey, uploadID, parts)
	if err != nil {
		t.Fatalf("Failed to complete multipart upload: %v", err)
	}

	if objInfo.Key != objectKey {
		t.Errorf("Expected object key %s, got %s", objectKey, objInfo.Key)
	}

	expectedSize := int64(len(partContent1) + len(partContent2))
	if objInfo.Size != expectedSize {
		t.Errorf("Expected object size %d, got %d", expectedSize, objInfo.Size)
	}

	// Verify content
	objReader, _, err := fs.GetObject(bucketName, objectKey)
	if err != nil {
		t.Fatalf("Failed to get object after multipart upload: %v", err)
	}
	defer objReader.Close()

	data, err := io.ReadAll(objReader)
	if err != nil {
		t.Fatalf("Failed to read object data: %v", err)
	}

	expectedContent := append(partContent1, partContent2...)
	if !bytes.Equal(data, expectedContent) {
		t.Errorf("Object content mismatch after multipart upload")
	}

	// Test AbortMultipartUpload
	// First initiate a new upload to abort
	newUploadID, err := fs.InitiateMultipartUpload(bucketName, "abort-key", metadata)
	if err != nil {
		t.Fatalf("Failed to initiate multipart upload for abort test: %v", err)
	}

	err = fs.AbortMultipartUpload(bucketName, "abort-key", newUploadID)
	if err != nil {
		t.Errorf("Failed to abort multipart upload: %v", err)
	}
}
