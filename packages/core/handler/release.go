package handler

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-core/config"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
)

type ReleaseHandler struct {
	useCase *usecase.ReleaseUseCase
	cfg     *config.Config
}

func NewReleaseHandler(uc *usecase.ReleaseUseCase, cfg *config.Config) *ReleaseHandler {
	return &ReleaseHandler{
		useCase: uc,
		cfg:     cfg,
	}
}

// UploadRelease handles POST /api/v1/releases/upload
func (h *ReleaseHandler) UploadRelease(c *gin.Context) {
	// Authenticate via RELEASE_API_SECRET
	providedSecret := c.GetHeader("X-Release-Secret")
	if providedSecret == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			providedSecret = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	expectedSecret := h.cfg.ReleaseAPISecret
	if expectedSecret == "" || subtle.ConstantTimeCompare([]byte(providedSecret), []byte(expectedSecret)) != 1 {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(
			utils.CodeUnauthorized,
			"unauthorized: valid release secret required",
			nil,
		))
		return
	}

	appName := c.DefaultPostForm("app_name", "pos-client")
	version := c.PostForm("version")
	if version == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"version is required (e.g. v1.0.1)",
			nil,
		))
		return
	}

	channel := c.DefaultPostForm("channel", "stable")
	releaseNotes := c.PostForm("release_notes")
	checksum := c.PostForm("checksum_sha256")
	isMandatory := c.DefaultPostForm("is_mandatory", "true") == "true"
	minSupportedVer := c.PostForm("min_supported_version")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"release binary file is required",
			nil,
		))
		return
	}

	resp := h.useCase.UploadRelease(
		c.Request.Context(),
		file,
		appName,
		version,
		channel,
		releaseNotes,
		checksum,
		isMandatory,
		minSupportedVer,
	)

	c.JSON(resp.StatusCode, resp)
}

// CheckLatestUpdate handles GET /api/v1/updates/:app/latest
func (h *ReleaseHandler) CheckLatestUpdate(c *gin.Context) {
	// Authenticate client
	providedKey := c.GetHeader("X-Nembus-App-Key")
	if providedKey == "" {
		providedKey = c.Query("client_key")
	}

	expectedKey := h.cfg.AppUpdateClientKey
	if expectedKey != "" && subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(
			utils.CodeUnauthorized,
			"unauthorized: invalid app update client key",
			nil,
		))
		return
	}

	appName := c.Param("app")
	if appName == "" {
		appName = "pos-client"
	}
	channel := c.DefaultQuery("channel", "stable")
	currVer := c.Query("current_version")

	rel, err := h.useCase.GetLatestReleaseInfo(c.Request.Context(), appName, channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"hasUpdate":      false,
			"currentVersion": currVer,
			"message":        "No active releases available",
		})
		return
	}

	latestVer := strings.TrimPrefix(rel.Version, "v")
	cleanCurrVer := strings.TrimPrefix(currVer, "v")
	hasUpdate := isVersionNewer(cleanCurrVer, latestVer)

	releaseNotesStr := ""
	if rel.ReleaseNotes.Valid {
		releaseNotesStr = rel.ReleaseNotes.String
	}

	downloadURL := fmt.Sprintf("/api/v1/updates/%s/download/%d", appName, rel.ID)

	c.JSON(http.StatusOK, gin.H{
		"hasUpdate":           hasUpdate,
		"currentVersion":      currVer,
		"latestVersion":       rel.Version,
		"releaseNotes":        releaseNotesStr,
		"downloadUrl":         downloadURL,
		"releaseId":           rel.ID,
		"fileName":            rel.FileName,
		"fileSizeBytes":       rel.FileSizeBytes,
		"checksumSha256":      rel.ChecksumSha256,
		"isMandatory":         rel.IsMandatory,
		"minSupportedVersion": rel.MinSupportedVersion.String,
		"createdAt":           rel.CreatedAt.Time,
	})
}

// DownloadRelease handles GET /api/v1/updates/:app/download/:id
func (h *ReleaseHandler) DownloadRelease(c *gin.Context) {
	providedKey := c.GetHeader("X-Nembus-App-Key")
	if providedKey == "" {
		providedKey = c.Query("client_key")
	}

	expectedKey := h.cfg.AppUpdateClientKey
	if expectedKey != "" && subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(
			utils.CodeUnauthorized,
			"unauthorized: invalid app update client key",
			nil,
		))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid release id",
			nil,
		))
		return
	}

	rel, err := h.useCase.GetReleaseByID(c.Request.Context(), int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewResponse(
			utils.CodeNotFound,
			"release not found",
			nil,
		))
		return
	}

	if _, err := os.Stat(rel.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, utils.NewResponse(
			utils.CodeNotFound,
			"release binary not found on server storage",
			nil,
		))
		return
	}

	c.Header("X-Checksum-SHA256", rel.ChecksumSha256)
	c.Header("X-Release-Version", rel.Version)
	c.FileAttachment(rel.FilePath, rel.FileName)
}

func isVersionNewer(curr, latest string) bool {
	if curr == "" {
		return true
	}
	currParts := parseVersion(curr)
	latestParts := parseVersion(latest)

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

func parseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		var num int
		num, _ = strconv.Atoi(p)
		res = append(res, num)
	}
	return res
}
