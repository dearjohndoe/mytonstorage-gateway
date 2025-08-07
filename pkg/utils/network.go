package utils

import (
	"strconv"
)

func FormatSize(size uint64) string {
	if size == 0 {
		return "folder"
	}

	if size < 1024 {
		return strconv.FormatUint(size, 10) + " B"
	}

	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for size >= 1024 && i < len(sizes)-1 {
		size /= 1024
		i++
	}

	return strconv.FormatUint(size, 10) + " " + sizes[i]
}
