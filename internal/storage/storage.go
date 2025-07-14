package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Storage defines the interface for storage backends
type Storage interface {
	// Bucket operations
	CreateBucket(bucket string) error
	DeleteBucket(bucket string) error
	ListBuckets() ([]BucketInfo, error)
	BucketExists(bucket string) bool

	// Object operations
	PutObject(bucket, key string, data io.Reader, size int64, metadata map[string]string) (*ObjectInfo, error)
	GetObject(bucket, key string) (io.ReadCloser, *ObjectInfo, error)
	DeleteObject(bucket, key string) error
	ListObjects(bucket, prefix, delimiter, marker string, maxKeys int) (*ListObjectsResult, error)
	HeadObject(bucket, key string) (*ObjectInfo, error)
	ObjectExists(bucket, key string) bool

	// Multipart operations
	InitiateMultipartUpload(bucket, key string, metadata map[string]string) (string, error)
	UploadPart(bucket, key, uploadID string, partNumber int, data io.Reader, size int64) (*PartInfo, error)
	CompleteMultipartUpload(bucket, key, uploadID string, parts []CompletePart) (*ObjectInfo, error)
	AbortMultipartUpload(bucket, key, uploadID string) error
}

// BucketInfo represents bucket information
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// ObjectInfo represents object information
type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	ContentType  string
	Metadata     map[string]string
}

// ListObjectsResult represents the result of listing objects
type ListObjectsResult struct {
	Objects        []ObjectInfo
	CommonPrefixes []string
	IsTruncated    bool
	NextMarker     string
}

// PartInfo represents multipart upload part information
type PartInfo struct {
	PartNumber int
	ETag       string
	Size       int64
}

// CompletePart represents a part in complete multipart upload
type CompletePart struct {
	PartNumber int
	ETag       string
}

// FileSystemStorage implements Storage interface using local filesystem
type FileSystemStorage struct {
	basePath string
}

// NewFileSystemStorage creates a new filesystem storage backend
func NewFileSystemStorage(basePath string) *FileSystemStorage {
	return &FileSystemStorage{
		basePath: basePath,
	}
}

func (fs *FileSystemStorage) CreateBucket(bucket string) error {
	bucketPath := filepath.Join(fs.basePath, bucket)
	return os.MkdirAll(bucketPath, 0755)
}

func (fs *FileSystemStorage) DeleteBucket(bucket string) error {
	bucketPath := filepath.Join(fs.basePath, bucket)

	// Check if bucket is empty
	entries, err := os.ReadDir(bucketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("bucket does not exist")
		}
		return err
	}

	if len(entries) > 0 {
		return fmt.Errorf("bucket is not empty")
	}

	return os.Remove(bucketPath)
}

func (fs *FileSystemStorage) ListBuckets() ([]BucketInfo, error) {
	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []BucketInfo{}, nil
		}
		return nil, err
	}

	var buckets []BucketInfo
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			buckets = append(buckets, BucketInfo{
				Name:         entry.Name(),
				CreationDate: info.ModTime(),
			})
		}
	}

	return buckets, nil
}

func (fs *FileSystemStorage) BucketExists(bucket string) bool {
	bucketPath := filepath.Join(fs.basePath, bucket)
	info, err := os.Stat(bucketPath)
	return err == nil && info.IsDir()
}

func (fs *FileSystemStorage) PutObject(bucket, key string, data io.Reader, size int64, metadata map[string]string) (*ObjectInfo, error) {
	if !fs.BucketExists(bucket) {
		return nil, fmt.Errorf("bucket does not exist")
	}

	objectPath := filepath.Join(fs.basePath, bucket, key)

	// Create directory if needed
	dir := filepath.Dir(objectPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create file
	file, err := os.Create(objectPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Copy data
	written, err := io.Copy(file, data)
	if err != nil {
		return nil, err
	}

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Store metadata if provided
	if len(metadata) > 0 {
		if err := fs.storeMetadata(objectPath, metadata); err != nil {
			// Log error but don't fail
		}
	}

	return &ObjectInfo{
		Key:          key,
		Size:         written,
		ETag:         fmt.Sprintf("\"%x\"", info.ModTime().Unix()),
		LastModified: info.ModTime(),
		ContentType:  getContentType(key),
		Metadata:     metadata,
	}, nil
}

func (fs *FileSystemStorage) GetObject(bucket, key string) (io.ReadCloser, *ObjectInfo, error) {
	if !fs.BucketExists(bucket) {
		return nil, nil, fmt.Errorf("bucket does not exist")
	}

	objectPath := filepath.Join(fs.basePath, bucket, key)

	file, err := os.Open(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("object does not exist")
		}
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	metadata := fs.loadMetadata(objectPath)

	objectInfo := &ObjectInfo{
		Key:          key,
		Size:         info.Size(),
		ETag:         fmt.Sprintf("\"%x\"", info.ModTime().Unix()),
		LastModified: info.ModTime(),
		ContentType:  getContentType(key),
		Metadata:     metadata,
	}

	return file, objectInfo, nil
}

func (fs *FileSystemStorage) DeleteObject(bucket, key string) error {
	if !fs.BucketExists(bucket) {
		return fmt.Errorf("bucket does not exist")
	}

	objectPath := filepath.Join(fs.basePath, bucket, key)

	// Remove metadata file if exists
	fs.removeMetadata(objectPath)

	return os.Remove(objectPath)
}

func (fs *FileSystemStorage) ListObjects(bucket, prefix, delimiter, marker string, maxKeys int) (*ListObjectsResult, error) {
	if !fs.BucketExists(bucket) {
		return nil, fmt.Errorf("bucket does not exist")
	}

	bucketPath := filepath.Join(fs.basePath, bucket)

	var objects []ObjectInfo
	var commonPrefixes []string
	prefixMap := make(map[string]bool)

	err := filepath.Walk(bucketPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path from bucket
		relPath, err := filepath.Rel(bucketPath, path)
		if err != nil {
			return nil
		}

		// Convert to forward slashes for S3 compatibility
		key := strings.ReplaceAll(relPath, "\\", "/")

		// Skip metadata files
		if strings.HasSuffix(key, ".metadata") {
			return nil
		}

		// Apply prefix filter
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}

		// Apply marker filter
		if marker != "" && key <= marker {
			return nil
		}

		// Handle delimiter
		if delimiter != "" {
			remaining := strings.TrimPrefix(key, prefix)
			if idx := strings.Index(remaining, delimiter); idx >= 0 {
				commonPrefix := prefix + remaining[:idx+len(delimiter)]
				if !prefixMap[commonPrefix] {
					commonPrefixes = append(commonPrefixes, commonPrefix)
					prefixMap[commonPrefix] = true
				}
				return nil
			}
		}

		metadata := fs.loadMetadata(path)

		objects = append(objects, ObjectInfo{
			Key:          key,
			Size:         info.Size(),
			ETag:         fmt.Sprintf("\"%x\"", info.ModTime().Unix()),
			LastModified: info.ModTime(),
			ContentType:  getContentType(key),
			Metadata:     metadata,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Apply maxKeys limit
	isTruncated := false
	nextMarker := ""
	if maxKeys > 0 && len(objects) > maxKeys {
		isTruncated = true
		if len(objects) > maxKeys {
			nextMarker = objects[maxKeys-1].Key
			objects = objects[:maxKeys]
		}
	}

	return &ListObjectsResult{
		Objects:        objects,
		CommonPrefixes: commonPrefixes,
		IsTruncated:    isTruncated,
		NextMarker:     nextMarker,
	}, nil
}

func (fs *FileSystemStorage) HeadObject(bucket, key string) (*ObjectInfo, error) {
	if !fs.BucketExists(bucket) {
		return nil, fmt.Errorf("bucket does not exist")
	}

	objectPath := filepath.Join(fs.basePath, bucket, key)

	info, err := os.Stat(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object does not exist")
		}
		return nil, err
	}

	metadata := fs.loadMetadata(objectPath)

	return &ObjectInfo{
		Key:          key,
		Size:         info.Size(),
		ETag:         fmt.Sprintf("\"%x\"", info.ModTime().Unix()),
		LastModified: info.ModTime(),
		ContentType:  getContentType(key),
		Metadata:     metadata,
	}, nil
}

func (fs *FileSystemStorage) ObjectExists(bucket, key string) bool {
	objectPath := filepath.Join(fs.basePath, bucket, key)
	_, err := os.Stat(objectPath)
	return err == nil
}

// Multipart upload methods (simplified implementation)
func (fs *FileSystemStorage) InitiateMultipartUpload(bucket, key string, metadata map[string]string) (string, error) {
	if !fs.BucketExists(bucket) {
		return "", fmt.Errorf("bucket does not exist")
	}

	uploadID := fmt.Sprintf("%d", time.Now().UnixNano())
	uploadDir := filepath.Join(fs.basePath, bucket, ".uploads", uploadID)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	return uploadID, nil
}

func (fs *FileSystemStorage) UploadPart(bucket, key, uploadID string, partNumber int, data io.Reader, size int64) (*PartInfo, error) {
	uploadDir := filepath.Join(fs.basePath, bucket, ".uploads", uploadID)
	partPath := filepath.Join(uploadDir, fmt.Sprintf("part-%d", partNumber))

	file, err := os.Create(partPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	written, err := io.Copy(file, data)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &PartInfo{
		PartNumber: partNumber,
		ETag:       fmt.Sprintf("\"%x\"", info.ModTime().Unix()),
		Size:       written,
	}, nil
}

func (fs *FileSystemStorage) CompleteMultipartUpload(bucket, key, uploadID string, parts []CompletePart) (*ObjectInfo, error) {
	uploadDir := filepath.Join(fs.basePath, bucket, ".uploads", uploadID)
	objectPath := filepath.Join(fs.basePath, bucket, key)

	// Create directory if needed
	dir := filepath.Dir(objectPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create final file
	file, err := os.Create(objectPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var totalSize int64

	// Concatenate parts
	for _, part := range parts {
		partPath := filepath.Join(uploadDir, fmt.Sprintf("part-%d", part.PartNumber))
		partFile, err := os.Open(partPath)
		if err != nil {
			return nil, err
		}

		written, err := io.Copy(file, partFile)
		partFile.Close()
		if err != nil {
			return nil, err
		}
		totalSize += written
	}

	// Clean up upload directory
	os.RemoveAll(uploadDir)

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &ObjectInfo{
		Key:          key,
		Size:         totalSize,
		ETag:         fmt.Sprintf("\"%x\"", info.ModTime().Unix()),
		LastModified: info.ModTime(),
		ContentType:  getContentType(key),
	}, nil
}

func (fs *FileSystemStorage) AbortMultipartUpload(bucket, key, uploadID string) error {
	uploadDir := filepath.Join(fs.basePath, bucket, ".uploads", uploadID)
	return os.RemoveAll(uploadDir)
}

// Helper methods
func (fs *FileSystemStorage) storeMetadata(objectPath string, metadata map[string]string) error {
	if len(metadata) == 0 {
		return nil
	}

	metadataPath := objectPath + ".metadata"
	file, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	for key, value := range metadata {
		fmt.Fprintf(file, "%s=%s\n", key, value)
	}

	return nil
}

func (fs *FileSystemStorage) loadMetadata(objectPath string) map[string]string {
	metadata := make(map[string]string)
	metadataPath := objectPath + ".metadata"

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return metadata
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			metadata[parts[0]] = parts[1]
		}
	}

	return metadata
}

func (fs *FileSystemStorage) removeMetadata(objectPath string) {
	metadataPath := objectPath + ".metadata"
	os.Remove(metadataPath)
}

func getContentType(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
