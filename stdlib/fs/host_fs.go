package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// File represents an open capability-checked stream file
type File struct {
	path     string
	rights   CapabilityRights
	osFile   *os.File
	isClosed bool
}

// IssueCapability creates a cryptographic file capability token for local storage access
func IssueCapability(issuer, path string, rights CapabilityRights, ttl time.Duration) *FileCapability {
	cleanPath := filepath.Clean(path)
	return &FileCapability{
		TokenID:   "cap_tok_" + issuer + "_" + cleanPath,
		Path:      cleanPath,
		Rights:    rights,
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Issuer:    issuer,
	}
}

// OpenFile opens a local disk file governed by an explicit capability
func OpenFile(cap *FileCapability, flag int, perm os.FileMode) (*File, error) {
	if cap == nil {
		return nil, errors.New("E_CAPABILITY_REQUIRED: nil capability provided")
	}

	// Verify required rights
	var required CapabilityRights
	if (flag & os.O_WRONLY) != 0 || (flag & os.O_RDWR) != 0 || (flag & os.O_CREATE) != 0 {
		required |= RightWrite
	}
	if (flag & os.O_APPEND) != 0 {
		required |= RightAppend
	}
	if (flag&os.O_WRONLY) == 0 || (flag&os.O_RDWR) != 0 {
		required |= RightRead
	}

	if !cap.IsValid(required) {
		return nil, errors.New("E_CAPABILITY_DENIED: caller lacks required file rights")
	}

	f, err := os.OpenFile(cap.Path, flag, perm)
	if err != nil {
		return nil, err
	}

	return &File{
		path:   cap.Path,
		rights: cap.Rights,
		osFile: f,
	}, nil
}

// ReadFile reads entire file from disk under RightRead capability
func ReadFile(cap *FileCapability) ([]byte, error) {
	if cap == nil || !cap.IsValid(RightRead) {
		return nil, errors.New("E_CAPABILITY_DENIED: caller lacks RightRead")
	}
	return os.ReadFile(cap.Path)
}

// WriteFile writes data atomically to disk under RightWrite capability
func WriteFile(cap *FileCapability, data []byte, perm os.FileMode) error {
	if cap == nil || !cap.IsValid(RightWrite) {
		return errors.New("E_CAPABILITY_DENIED: caller lacks RightWrite")
	}
	return os.WriteFile(cap.Path, data, perm)
}

// Read reads up to len(b) bytes from stream
func (f *File) Read(b []byte) (int, error) {
	if f.isClosed {
		return 0, errors.New("E_FILE_CLOSED")
	}
	if (f.rights & RightRead) == 0 {
		return 0, errors.New("E_CAPABILITY_DENIED: read permission not granted")
	}
	return f.osFile.Read(b)
}

// Write writes bytes to stream
func (f *File) Write(b []byte) (int, error) {
	if f.isClosed {
		return 0, errors.New("E_FILE_CLOSED")
	}
	if (f.rights & RightWrite) == 0 && (f.rights & RightAppend) == 0 {
		return 0, errors.New("E_CAPABILITY_DENIED: write/append permission not granted")
	}
	return f.osFile.Write(b)
}

// Seek sets the offset for the next Read or Write on file
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.isClosed {
		return 0, errors.New("E_FILE_CLOSED")
	}
	return f.osFile.Seek(offset, whence)
}

// Sync commits the current contents of the file to stable storage
func (f *File) Sync() error {
	if f.isClosed {
		return errors.New("E_FILE_CLOSED")
	}
	return f.osFile.Sync()
}

// Close closes the underlying OS file handle
func (f *File) Close() error {
	if f.isClosed {
		return nil
	}
	f.isClosed = true
	return f.osFile.Close()
}

// Copy copies from src to dst verifying capabilities on both streams
func Copy(dst *File, src *File) (int64, error) {
	if (dst.rights & RightWrite) == 0 {
		return 0, errors.New("E_CAPABILITY_DENIED: dst lacks RightWrite")
	}
	if (src.rights & RightRead) == 0 {
		return 0, errors.New("E_CAPABILITY_DENIED: src lacks RightRead")
	}
	return io.Copy(dst.osFile, src.osFile)
}
