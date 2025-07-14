package handlers

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"locals3/internal/auth"
	"locals3/internal/storage"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Config holds handler configuration
type Config struct {
	Storage     storage.Storage
	Auth        auth.AuthProvider
	Region      string
	BaseDomain  string
	DisableAuth bool
}

// Handler holds the HTTP handlers
type Handler struct {
	storage     storage.Storage
	auth        auth.AuthProvider
	region      string
	baseDomain  string
	disableAuth bool
}

// New creates a new handler instance
func New(cfg *Config) *Handler {
	return &Handler{
		storage:     cfg.Storage,
		auth:        cfg.Auth,
		region:      cfg.Region,
		baseDomain:  cfg.BaseDomain,
		disableAuth: cfg.DisableAuth,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// authenticate validates the request authentication
func (h *Handler) authenticate(r *http.Request) error {
	// Skip authentication for health check
	if r.URL.Path == "/health" {
		return nil
	}

	// Skip authentication if disabled
	if h.disableAuth {
		logrus.Debug("Authentication disabled")
		return nil
	}

	// For now, allow requests without authentication for testing
	// In production, you should always authenticate
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logrus.Warn("Request without authentication")
		return nil
	}

	return h.auth.Authenticate(r)
}

// setS3Headers sets common S3 response headers
func (h *Handler) setS3Headers(w http.ResponseWriter) {
	w.Header().Set("Server", "LocalS3")
	w.Header().Set("x-amz-request-id", fmt.Sprintf("%d", time.Now().UnixNano()))
}

// writeErrorResponse writes an S3 error response
func (h *Handler) writeErrorResponse(w http.ResponseWriter, code string, message string, statusCode int) {
	h.setS3Headers(w)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)

	errorResp := &S3Error{
		Code:      code,
		Message:   message,
		RequestID: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	xml.NewEncoder(w).Encode(errorResp)
}

// ListBuckets handles GET / - list all buckets
func (h *Handler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	buckets, err := h.storage.ListBuckets()
	if err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	response := &ListAllMyBucketsResult{
		Owner: Owner{
			ID:          h.auth.GetAccessKey(),
			DisplayName: h.auth.GetAccessKey(),
		},
		Buckets: Buckets{
			Bucket: make([]Bucket, len(buckets)),
		},
	}

	for i, bucket := range buckets {
		response.Buckets.Bucket[i] = Bucket{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate.Format(time.RFC3339),
		}
	}

	h.setS3Headers(w)
	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(response)
}

// CreateBucket handles PUT /{bucket} - create bucket
func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if bucket == "" {
		h.writeErrorResponse(w, "InvalidBucketName", "Bucket name is required", http.StatusBadRequest)
		return
	}

	if h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "BucketAlreadyExists", "Bucket already exists", http.StatusConflict)
		return
	}

	if err := h.storage.CreateBucket(bucket); err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	h.setS3Headers(w)
	w.WriteHeader(http.StatusOK)
}

// DeleteBucket handles DELETE /{bucket} - delete bucket
func (h *Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	if err := h.storage.DeleteBucket(bucket); err != nil {
		if strings.Contains(err.Error(), "not empty") {
			h.writeErrorResponse(w, "BucketNotEmpty", "Bucket is not empty", http.StatusConflict)
		} else {
			h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		}
		return
	}

	h.setS3Headers(w)
	w.WriteHeader(http.StatusNoContent)
}

// ListObjects handles GET /{bucket} - list objects in bucket
func (h *Handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	// Parse query parameters
	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")
	marker := r.URL.Query().Get("marker")
	maxKeysStr := r.URL.Query().Get("max-keys")

	maxKeys := 1000 // Default
	if maxKeysStr != "" {
		if mk, err := strconv.Atoi(maxKeysStr); err == nil && mk > 0 {
			maxKeys = mk
		}
	}

	result, err := h.storage.ListObjects(bucket, prefix, delimiter, marker, maxKeys)
	if err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	response := &ListBucketResult{
		Name:           bucket,
		Prefix:         prefix,
		Marker:         marker,
		MaxKeys:        maxKeys,
		IsTruncated:    result.IsTruncated,
		NextMarker:     result.NextMarker,
		Contents:       make([]Object, len(result.Objects)),
		CommonPrefixes: make([]CommonPrefix, len(result.CommonPrefixes)),
	}

	for i, obj := range result.Objects {
		response.Contents[i] = Object{
			Key:          obj.Key,
			LastModified: obj.LastModified.Format(time.RFC3339),
			ETag:         obj.ETag,
			Size:         obj.Size,
			StorageClass: "STANDARD",
			Owner: Owner{
				ID:          h.auth.GetAccessKey(),
				DisplayName: h.auth.GetAccessKey(),
			},
		}
	}

	for i, prefix := range result.CommonPrefixes {
		response.CommonPrefixes[i] = CommonPrefix{
			Prefix: prefix,
		}
	}

	h.setS3Headers(w)
	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(response)
}

// PutObject handles PUT /{bucket}/{key} - upload object
func (h *Handler) PutObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	// Get content length
	contentLength := r.ContentLength
	if contentLength < 0 {
		contentLengthStr := r.Header.Get("Content-Length")
		if contentLengthStr != "" {
			if cl, err := strconv.ParseInt(contentLengthStr, 10, 64); err == nil {
				contentLength = cl
			}
		}
	}

	// Extract metadata from headers
	metadata := make(map[string]string)
	for name, values := range r.Header {
		if strings.HasPrefix(strings.ToLower(name), "x-amz-meta-") {
			metadata[name] = values[0]
		}
	}

	// Store object
	objInfo, err := h.storage.PutObject(bucket, key, r.Body, contentLength, metadata)
	if err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	h.setS3Headers(w)
	w.Header().Set("ETag", objInfo.ETag)
	w.WriteHeader(http.StatusOK)
}

// GetObject handles GET /{bucket}/{key} - download object
func (h *Handler) GetObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	reader, objInfo, err := h.storage.GetObject(bucket, key)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			h.writeErrorResponse(w, "NoSuchKey", "Object does not exist", http.StatusNotFound)
		} else {
			h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer reader.Close()

	h.setS3Headers(w)
	w.Header().Set("Content-Type", objInfo.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(objInfo.Size, 10))
	w.Header().Set("ETag", objInfo.ETag)
	w.Header().Set("Last-Modified", objInfo.LastModified.Format(http.TimeFormat))

	// Set metadata headers
	for key, value := range objInfo.Metadata {
		w.Header().Set(key, value)
	}

	io.Copy(w, reader)
}

// DeleteObject handles DELETE /{bucket}/{key} - delete object
func (h *Handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	if err := h.storage.DeleteObject(bucket, key); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			h.writeErrorResponse(w, "NoSuchKey", "Object does not exist", http.StatusNotFound)
		} else {
			h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		}
		return
	}

	h.setS3Headers(w)
	w.WriteHeader(http.StatusNoContent)
}

// HeadObject handles HEAD /{bucket}/{key} - get object metadata
func (h *Handler) HeadObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	objInfo, err := h.storage.HeadObject(bucket, key)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			h.writeErrorResponse(w, "NoSuchKey", "Object does not exist", http.StatusNotFound)
		} else {
			h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		}
		return
	}

	h.setS3Headers(w)
	w.Header().Set("Content-Type", objInfo.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(objInfo.Size, 10))
	w.Header().Set("ETag", objInfo.ETag)
	w.Header().Set("Last-Modified", objInfo.LastModified.Format(http.TimeFormat))

	// Set metadata headers
	for key, value := range objInfo.Metadata {
		w.Header().Set(key, value)
	}

	w.WriteHeader(http.StatusOK)
}

// Multipart upload handlers (simplified)
func (h *Handler) InitiateMultipartUpload(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	// Extract metadata from headers
	metadata := make(map[string]string)
	for name, values := range r.Header {
		if strings.HasPrefix(strings.ToLower(name), "x-amz-meta-") {
			metadata[name] = values[0]
		}
	}

	uploadID, err := h.storage.InitiateMultipartUpload(bucket, key, metadata)
	if err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	response := &InitiateMultipartUploadResult{
		Bucket:   bucket,
		Key:      key,
		UploadID: uploadID,
	}

	h.setS3Headers(w)
	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(response)
}

func (h *Handler) UploadPart(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	partNumberStr := r.URL.Query().Get("partNumber")
	uploadID := r.URL.Query().Get("uploadId")

	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 {
		h.writeErrorResponse(w, "InvalidPart", "Invalid part number", http.StatusBadRequest)
		return
	}

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	contentLength := r.ContentLength
	partInfo, err := h.storage.UploadPart(bucket, key, uploadID, partNumber, r.Body, contentLength)
	if err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	h.setS3Headers(w)
	w.Header().Set("ETag", partInfo.ETag)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CompleteMultipartUpload(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	uploadID := r.URL.Query().Get("uploadId")

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	// Parse request body to get parts
	var completeRequest CompleteMultipartUpload
	if err := xml.NewDecoder(r.Body).Decode(&completeRequest); err != nil {
		h.writeErrorResponse(w, "MalformedXML", "Invalid XML", http.StatusBadRequest)
		return
	}

	objInfo, err := h.storage.CompleteMultipartUpload(bucket, key, uploadID, completeRequest.Part)
	if err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	response := &CompleteMultipartUploadResult{
		Location: fmt.Sprintf("http://%s/%s/%s", h.baseDomain, bucket, key),
		Bucket:   bucket,
		Key:      key,
		ETag:     objInfo.ETag,
	}

	h.setS3Headers(w)
	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(response)
}

func (h *Handler) AbortMultipartUpload(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeErrorResponse(w, "AccessDenied", err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	uploadID := r.URL.Query().Get("uploadId")

	if !h.storage.BucketExists(bucket) {
		h.writeErrorResponse(w, "NoSuchBucket", "Bucket does not exist", http.StatusNotFound)
		return
	}

	if err := h.storage.AbortMultipartUpload(bucket, key, uploadID); err != nil {
		h.writeErrorResponse(w, "InternalError", err.Error(), http.StatusInternalServerError)
		return
	}

	h.setS3Headers(w)
	w.WriteHeader(http.StatusNoContent)
}
