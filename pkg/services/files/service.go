package files

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	remotes "mytonstorage-gateway/pkg/clients/remote-ton-storage"
	tonstorageClient "mytonstorage-gateway/pkg/clients/ton-storage"
	"mytonstorage-gateway/pkg/models"
	v1 "mytonstorage-gateway/pkg/models/api/v1"
	"mytonstorage-gateway/pkg/models/private"
)

type service struct {
	reports          reportsDb
	tonstorage       storage
	remoteTonStorage remotes.Client
	logger           *slog.Logger
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

func (s *service) GetPathInfo(ctx context.Context, bagID, path string) (private.FolderInfo, error) {
	log := s.logger.With(
		slog.String("method", "GetPathInfo"),
		slog.String("bagID", bagID),
		slog.String("path", path),
	)
	isBanned, err := s.reports.HasBan(ctx, bagID)
	if err != nil {
		log.Error("failed to check ban status", slog.String("error", err.Error()))
		return private.FolderInfo{}, models.NewAppError(models.InternalServerErrorCode, "")
	}

	if isBanned {
		log.Warn("bag is banned", slog.String("bagID", bagID))
		return private.FolderInfo{}, models.NewAppError(models.NotAcceptableErrorCode, "bag is banned")
	}

	if info, err := s.getFromLocalStorage(ctx, bagID, path, log); err == nil {
		return info, nil
	}

	return s.getFromRemoteStorage(ctx, bagID, path, log)
}

func (s *service) getFromLocalStorage(ctx context.Context, bagID, path string, log *slog.Logger) (private.FolderInfo, error) {
	bag, err := s.tonstorage.GetBag(ctx, bagID)
	if err != nil {
		log.Info("failed to get bag from local storage", slog.String("error", err.Error()))
		return private.FolderInfo{}, err
	}

	info := private.FolderInfo{
		BagID:       bagID,
		PeersCount:  len(bag.Peers),
		TotalSize:   bag.Size,
		Description: bag.Description,
		Files:       ls(bag.Files, path),
	}

	if s.isSingleFile(info.Files, path) {
		info.SingleFilePath = filepath.Join(bag.Path, bag.DirName, path)
	}

	return info, nil
}

func (s *service) getFromRemoteStorage(ctx context.Context, bagID, path string, log *slog.Logger) (private.FolderInfo, error) {
	if s.remoteTonStorage == nil {
		log.Error("remote ton storage client is not configured")
		return private.FolderInfo{}, models.NewAppError(models.NotFoundErrorCode, "bag not found")
	}

	files, err := s.remoteTonStorage.ListFiles(ctx, bagID)
	if err != nil {
		if errors.Is(err, remotes.ErrTimeout) {
			log.Error("failed to list files from remote", slog.String("error", err.Error()))
			return private.FolderInfo{
				BagID:      bagID,
				PeersCount: files.PeersCount,
			}, models.NewAppError(models.NotFoundErrorCode, "bag not found")
		}

		log.Error("remote-ton-storage ListFiles failed", slog.String("error", err.Error()))
		return private.FolderInfo{}, models.NewAppError(models.NotFoundErrorCode, "bag not found")
	}

	if len(files.Files) == 0 {
		log.Warn("no files found in remote", slog.String("bagID", bagID))
		return private.FolderInfo{}, models.NewAppError(models.NotFoundErrorCode, "bag doesn't contain any files")
	}

	info := private.FolderInfo{
		BagID:       bagID,
		Description: files.Description,
		TotalSize:   files.TotalSize,
		PeersCount:  files.PeersCount,
		Files:       ls(files.Files, path),
	}

	if s.isSingleFile(info.Files, path) {
		info.StreamFile, err = s.streamRemoteFile(ctx, bagID, path, log)
		if err != nil {
			log.Error("failed to stream file from remote", slog.String("error", err.Error()))
			return private.FolderInfo{}, err
		}
	}

	return info, nil
}

func (s *service) isSingleFile(files []v1.File, path string) bool {
	return len(files) == 1 && !files[0].IsFolder && strings.HasSuffix(path, files[0].Name)
}

func (s *service) streamRemoteFile(ctx context.Context, bagID, path string, log *slog.Logger) (*private.StreamFile, error) {
	fs, err := s.remoteTonStorage.StreamFile(ctx, bagID, path)
	if err != nil {
		if errors.Is(err, remotes.ErrTimeout) {
			return &private.StreamFile{
				PeersCount: fs.PeersCount,
			}, models.NewAppError(models.NotFoundErrorCode, "bag not found")
		}

		log.Error("failed to stream file from remote", slog.String("error", err.Error()))
		return nil, models.NewAppError(models.InternalServerErrorCode, "")
	}

	return &private.StreamFile{
		FileStream: fs.FileStream,
		Size:       fs.Size,
		PeersCount: fs.PeersCount,
	}, nil
}

func ls(files []tonstorageClient.File, path string) []v1.File {
	normalizedPath := strings.Trim(path, string(filepath.Separator))

	var result []v1.File
	foundDirs := make(map[string]bool)

	for _, file := range files {
		fileName := strings.Trim(file.Name, string(filepath.Separator))

		if normalizedPath == "" {
			parts := strings.Split(fileName, string(filepath.Separator))
			dirName := parts[0]

			if strings.Contains(fileName, string(filepath.Separator)) {
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
		} else if subDir, ok := strings.CutPrefix(fileName, normalizedPath+string(filepath.Separator)); ok {
			parts := strings.Split(subDir, string(filepath.Separator))
			dirName := parts[0]

			if strings.Contains(subDir, string(filepath.Separator)) {
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
	rstorage remotes.Client,
	logger *slog.Logger,
) Files {
	return &service{
		reports:          reports,
		tonstorage:       tonstorage,
		remoteTonStorage: rstorage,
		logger:           logger,
	}
}
