package private

import v1 "mytonstorage-gateway/pkg/models/api/v1"

type FolderInfo struct {
	IsValid  bool      `json:"is_valid"`
	BagID    string    `json:"bagid"`
	DiskPath string    `json:"root_dir"`
	Files    []v1.File `json:"files"`
}
