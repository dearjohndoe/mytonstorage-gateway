package files

import (
	"context"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"mytonstorage-gateway/pkg/models"
	v1 "mytonstorage-gateway/pkg/models/api/v1"
	"mytonstorage-gateway/pkg/models/private"
	tonstorageClient "mytonstorage-gateway/pkg/ton-storage"
)

type service struct {
	reports    reportsDb
	tonstorage storage
	logger     *slog.Logger
}

type reportsDb interface {
	HasBan(ctx context.Context, bagID string) (bool, error)
}

type storage interface {
	GetBag(ctx context.Context, bagId string) (*tonstorageClient.BagDetailed, error)
}

type Files interface {
	GetPathInfo(ctx context.Context, bagID, path string) (private.FolderInfo, error)
}

func (s *service) GetPathInfo(ctx context.Context, bagID, path string) (info private.FolderInfo, err error) {
	log := s.logger.With(
		slog.String("method", "GetPathInfo"),
		slog.String("bagID", bagID),
		slog.String("path", path),
	)

	isBanned, err := s.reports.HasBan(ctx, bagID)
	if err != nil {
		log.Error("failed to check ban status", slog.String("error", err.Error()))
		return info, models.NewAppError(models.InternalServerErrorCode, "")
	}

	if isBanned {
		log.Warn("bag is banned", slog.String("bagID", bagID))
		return info, models.NewAppError(models.NotAcceptableErrorCode, "bag is banned")
	}

	bag, err := s.tonstorage.GetBag(ctx, bagID)
	if err != nil {
		if err != tonstorageClient.ErrNotFound {
			log.Error("failed to get bag", slog.String("error", err.Error()))
		}

		return info, models.NewAppError(models.NotFoundErrorCode, "bag not found")
	}

	info = private.FolderInfo{
		IsValid:  true,
		BagID:    strings.ToUpper(bagID),
		DiskPath: filepath.Join(bag.Path, bag.DirName),
		Files:    ls(bag.Files, path),
	}

	return
}

func ls(files []tonstorageClient.File, path string) []v1.File {
	normalizedPath := strings.Trim(path, "/")

	var result []v1.File
	foundDirs := make(map[string]bool)

	for _, file := range files {
		fileName := strings.Trim(file.Name, "/")

		if normalizedPath == "" {
			parts := strings.Split(fileName, "/")
			dirName := parts[0]

			if strings.Contains(fileName, "/") {
				if !foundDirs[dirName] {
					foundDirs[dirName] = true
					result = append(result, v1.File{
						Name:     dirName,
						Size:     0,
						IsFolder: true,
					})
				}
			} else {
				result = append(result, v1.File{
					Name:     filepath.Base(fileName),
					Size:     file.Size,
					IsFolder: false,
				})
			}
		} else if subDir, ok := strings.CutPrefix(fileName, normalizedPath+"/"); ok {
			parts := strings.Split(subDir, "/")
			dirName := parts[0]

			if strings.Contains(subDir, "/") {
				if !foundDirs[dirName] {
					foundDirs[dirName] = true
					result = append(result, v1.File{
						Name:     dirName,
						Size:     0,
						IsFolder: true,
					})
				}
			} else {
				result = append(result, v1.File{
					Name:     filepath.Base(fileName),
					Size:     file.Size,
					IsFolder: false,
				})
			}
		} else if normalizedPath == fileName {
			result = append(result, v1.File{
				Name:     filepath.Base(fileName),
				Size:     file.Size,
				IsFolder: false,
			})

			break
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsFolder && !result[j].IsFolder {
			return true
		}
		if !result[i].IsFolder && result[j].IsFolder {
			return false
		}
		return result[i].Name < result[j].Name
	})

	return result
}

func NewService(
	reports reportsDb,
	tonstorage storage,
	logger *slog.Logger,
) Files {
	return &service{
		reports:    reports,
		tonstorage: tonstorage,
		logger:     logger,
	}
}
