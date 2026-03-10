package sessionstore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type stagedDelete struct {
	OriginalPath string
	StagedPath   string
}

type copiedBackup struct {
	OriginalPath string
	BackupPath   string
}

func stageDeletePaths(paths []string) ([]stagedDelete, error) {
	staged := make([]stagedDelete, 0, len(paths))
	for _, path := range paths {
		if !pathExists(path) {
			continue
		}
		item, err := stageDeletePath(path)
		if err != nil {
			_ = restoreStagedDeletes(staged)
			return nil, err
		}
		staged = append(staged, item)
	}
	return staged, nil
}

func stageDeletePath(path string) (stagedDelete, error) {
	path = cleanPath(path)
	if !filepath.IsAbs(path) {
		return stagedDelete{}, fmt.Errorf("refusing to stage non-absolute path %q", path)
	}
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	stagedPath := filepath.Join(dir, fmt.Sprintf(".%s.oac-stage-%d", base, time.Now().UnixNano()))
	if err := os.Rename(path, stagedPath); err != nil {
		return stagedDelete{}, err
	}
	return stagedDelete{
		OriginalPath: path,
		StagedPath:   stagedPath,
	}, nil
}

func restoreStagedDeletes(items []stagedDelete) error {
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if !pathExists(item.StagedPath) {
			continue
		}
		if pathExists(item.OriginalPath) {
			return fmt.Errorf("cannot restore %s because %s already exists", item.StagedPath, item.OriginalPath)
		}
		if err := os.Rename(item.StagedPath, item.OriginalPath); err != nil {
			return err
		}
	}
	return nil
}

func cleanupStagedDeletes(items []stagedDelete) error {
	for _, item := range items {
		if err := deletePath(item.StagedPath); err != nil {
			return err
		}
	}
	return nil
}

func backupFileIfExists(path string) ([]copiedBackup, error) {
	if strings.TrimSpace(path) == "" || !pathExists(path) {
		return nil, nil
	}
	backupPath := filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s.oac-backup-%d", filepath.Base(path), time.Now().UnixNano()))
	if err := copyFile(path, backupPath); err != nil {
		return nil, err
	}
	return []copiedBackup{{
		OriginalPath: path,
		BackupPath:   backupPath,
	}}, nil
}

func backupSQLiteFiles(path string) ([]copiedBackup, error) {
	paths := []string{path, path + "-wal", path + "-shm"}
	backups := []copiedBackup{}
	for _, item := range paths {
		copied, err := backupFileIfExists(item)
		if err != nil {
			_ = cleanupCopiedBackups(backups)
			return nil, err
		}
		backups = append(backups, copied...)
	}
	return backups, nil
}

func restoreCopiedBackups(items []copiedBackup) error {
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if !pathExists(item.BackupPath) {
			continue
		}
		if pathExists(item.OriginalPath) {
			if err := deletePath(item.OriginalPath); err != nil {
				return err
			}
		}
		if err := os.Rename(item.BackupPath, item.OriginalPath); err != nil {
			return err
		}
	}
	return nil
}

func cleanupCopiedBackups(items []copiedBackup) error {
	for _, item := range items {
		if err := deletePath(item.BackupPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	info, err := source.Stat()
	if err != nil {
		return err
	}
	target, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}

	if _, err := io.Copy(target, source); err != nil {
		target.Close()
		_ = deletePath(dst)
		return err
	}
	if err := target.Close(); err != nil {
		_ = deletePath(dst)
		return err
	}
	return nil
}
