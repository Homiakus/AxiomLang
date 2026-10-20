package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CapabilityRights defines explicit fine-grained token permissions.
type CapabilityRights uint32

const (
	RightRead CapabilityRights = 1 << iota
	RightWrite
	RightAppend
	RightExecute
	RightDelegate
)

// FileCapability represents a cryptographic capability token granting file access.
type FileCapability struct {
	TokenID   string
	Path      string
	Rights    CapabilityRights
	ExpiresAt int64
	Issuer    string
}

func (fc *FileCapability) IsValid(required CapabilityRights) bool {
	if time.Now().Unix() > fc.ExpiresAt {
		return false
	}
	return (fc.Rights & required) == required
}

// Inode represents a capability-governed immutable state block.
type Inode struct {
	ID        uint64
	Path      string
	Data      []byte
	Hash      string
	Version   uint64
	Modified  int64
	mu        sync.RWMutex
}

// GovernedFileSystem provides transactional, zero-ambient-authority storage operations.
type GovernedFileSystem struct {
	mu     sync.RWMutex
	inodes map[string]*Inode
	nextID uint64
}

func NewGovernedFileSystem() *GovernedFileSystem {
	return &GovernedFileSystem{
		inodes: make(map[string]*Inode),
		nextID: 1,
	}
}

// Write transactional write requiring explicit RightWrite capability.
func (gfs *GovernedFileSystem) Write(cap *FileCapability, data []byte) error {
	if cap == nil || !cap.IsValid(RightWrite) {
		return errors.New("E_CAPABILITY_DENIED: caller lacks valid RightWrite authority")
	}

	gfs.mu.Lock()
	defer gfs.mu.Unlock()

	hasher := sha256.New()
	hasher.Write(data)
	hashStr := hex.EncodeToString(hasher.Sum(nil))

	inode, exists := gfs.inodes[cap.Path]
	if !exists {
		gfs.inodes[cap.Path] = &Inode{
			ID:       gfs.nextID,
			Path:     cap.Path,
			Data:     data,
			Hash:     hashStr,
			Version:  1,
			Modified: time.Now().Unix(),
		}
		gfs.nextID++
		return nil
	}

	inode.mu.Lock()
	defer inode.mu.Unlock()
	inode.Data = data
	inode.Hash = hashStr
	inode.Version++
	inode.Modified = time.Now().Unix()
	return nil
}

// Read transactional read requiring explicit RightRead capability.
func (gfs *GovernedFileSystem) Read(cap *FileCapability) ([]byte, string, error) {
	if cap == nil || !cap.IsValid(RightRead) {
		return nil, "", errors.New("E_CAPABILITY_DENIED: caller lacks valid RightRead authority")
	}

	gfs.mu.RLock()
	inode, exists := gfs.inodes[cap.Path]
	gfs.mu.RUnlock()

	if !exists {
		return nil, "", fmt.Errorf("E_NOT_FOUND: resource %s does not exist", cap.Path)
	}

	inode.mu.RLock()
	defer inode.mu.RUnlock()
	dataCopy := make([]byte, len(inode.Data))
	copy(dataCopy, inode.Data)
	return dataCopy, inode.Hash, nil
}
