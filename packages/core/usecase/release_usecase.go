package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReleaseUseCase struct {
	repo       *repository.Queries
	storageDir string
}

func NewReleaseUseCase(repo *repository.Queries, storageDir string) *ReleaseUseCase {
	if storageDir == "" {
		storageDir = "uploads/releases"
	}
	return &ReleaseUseCase{
		repo:       repo,
		storageDir: storageDir,
	}
}

func (uc *ReleaseUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// UploadRelease saves a new application release binary, records metadata in the database,
// and automatically deletes older build files to prevent disk exhaustion.
func (uc *ReleaseUseCase) UploadRelease(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
	appName string,
	version string,
	channel string,
	releaseNotes string,
	expectedChecksum string,
	isMandatory bool,
	minSupportedVer string,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not initialized", nil)
	}

	appName = strings.ToLower(strings.TrimSpace(appName))
	if appName == "" {
		appName = "pos-client"
	}
	if channel == "" {
		channel = "stable"
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return utils.NewResponse(utils.CodeBadReq, "version is required", nil)
	}

	// Prepare release directory: uploads/releases/<app_name>/<channel>/
	targetDir := filepath.Join(uc.storageDir, appName, channel)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create release directory: %v", err), nil)
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		ext = ".exe"
	}
	destFileName := fmt.Sprintf("%s-%s%s", appName, version, ext)
	destPath := filepath.Join(targetDir, destFileName)

	// Save incoming multipart file to destination while calculating SHA-256
	src, err := fileHeader.Open()
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to open uploaded file: %v", err), nil)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create target file on disk: %v", err), nil)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(dst, hasher)
	writtenBytes, err := io.Copy(multiWriter, src)
	_ = dst.Close()
	if err != nil {
		_ = os.Remove(destPath)
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to save binary content: %v", err), nil)
	}

	computedHash := hex.EncodeToString(hasher.Sum(nil))
	if expectedChecksum != "" && !strings.EqualFold(expectedChecksum, computedHash) {
		_ = os.Remove(destPath)
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("checksum mismatch: expected %s, got %s", expectedChecksum, computedHash), nil)
	}

	// Record release in master DB
	releaseParams := repository.CreateReleaseParams{
		AppName:             appName,
		Version:             version,
		Channel:             channel,
		ReleaseNotes:        pgtype.Text{String: releaseNotes, Valid: releaseNotes != ""},
		FileName:            destFileName,
		FilePath:            destPath,
		FileSizeBytes:       writtenBytes,
		ChecksumSha256:      computedHash,
		IsMandatory:         isMandatory,
		MinSupportedVersion: pgtype.Text{String: minSupportedVer, Valid: minSupportedVer != ""},
		IsActive:            true,
	}

	newRelease, err := uc.repo.CreateRelease(ctx, releaseParams)
	if err != nil {
		_ = os.Remove(destPath)
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to record release in database: %v", err), nil)
	}

	// PRUNE OLD BUILDS: Remove older release binaries from disk and database
	go uc.cleanupOldReleases(context.Background(), appName, channel, newRelease.ID)

	return utils.NewResponse(utils.CodeCreated, "release published successfully", newRelease)
}

// cleanupOldReleases removes files of older releases from disk and database
func (uc *ReleaseUseCase) cleanupOldReleases(ctx context.Context, appName, channel string, currentReleaseID int32) {
	olderReleases, err := uc.repo.GetOlderReleasesToPrune(ctx, repository.GetOlderReleasesToPruneParams{
		AppName: appName,
		Channel: channel,
		ID:      currentReleaseID,
	})
	if err != nil {
		log.Printf("[RELEASE_CLEANUP] Failed to query older releases: %v", err)
		return
	}

	for _, rel := range olderReleases {
		if rel.FilePath != "" {
			if _, statErr := os.Stat(rel.FilePath); statErr == nil {
				if removeErr := os.Remove(rel.FilePath); removeErr != nil {
					log.Printf("[RELEASE_CLEANUP] Could not remove old binary '%s': %v", rel.FilePath, removeErr)
				} else {
					log.Printf("[RELEASE_CLEANUP] Successfully purged old build file '%s' (version %s)", rel.FilePath, rel.Version)
				}
			}
		}
		// Delete database record for pruned release
		if delErr := uc.repo.DeleteRelease(ctx, rel.ID); delErr != nil {
			log.Printf("[RELEASE_CLEANUP] Failed to delete release ID %d from database: %v", rel.ID, delErr)
		}
	}
}

// GetLatestReleaseInfo returns latest release info for client check
func (uc *ReleaseUseCase) GetLatestReleaseInfo(ctx context.Context, appName, channel string) (*repository.AppRelease, error) {
	if uc.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	if appName == "" {
		appName = "pos-client"
	}
	if channel == "" {
		channel = "stable"
	}
	release, err := uc.repo.GetLatestActiveRelease(ctx, repository.GetLatestActiveReleaseParams{
		AppName: appName,
		Channel: channel,
	})
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// GetReleaseByID retrieves a release by ID
func (uc *ReleaseUseCase) GetReleaseByID(ctx context.Context, id int32) (*repository.AppRelease, error) {
	if uc.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	release, err := uc.repo.GetReleaseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// IsVersionNewer compares semver strings without 'v' prefix
func IsVersionNewer(curr, latest string) bool {
	if curr == "" {
		return true
	}
	currParts := ParseVersion(curr)
	latestParts := ParseVersion(latest)

	for i := 0; i < len(currParts) || i < len(latestParts); i++ {
		var c, l int
		if i < len(currParts) {
			c = currParts[i]
		}
		if i < len(latestParts) {
			l = latestParts[i]
		}
		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}
	return false
}

// ParseVersion extracts integer components of version string
func ParseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		var num int
		fmt.Sscanf(p, "%d", &num)
		res = append(res, num)
	}
	return res
}

