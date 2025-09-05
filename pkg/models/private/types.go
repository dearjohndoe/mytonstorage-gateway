package private

import (
	"io"

	v1 "mytonstorage-gateway/pkg/models/api/v1"
)

type FolderInfo struct {
	StreamFile     *StreamFile
	Files          []v1.File
	TotalSize      uint64
	PeersCount     int
	Description    string
	BagID          string
	SingleFilePath string
	IsValid        bool
}

type StreamFile struct {
	FileStream io.ReadCloser
	Size       uint64
}
