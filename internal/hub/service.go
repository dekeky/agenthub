package hub

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/agenthub/internal/archive"
)

const maxUploadSize = 200 << 20 // 200MB

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

type UploadInput struct {
	AgentName string
	Category  string
	Version   string
	File      *multipart.FileHeader
}

func (s *Service) Upload(in UploadInput) (AgentMeta, error) {
	if in.File == nil {
		return AgentMeta{}, fmt.Errorf("file is required")
	}
	if in.File.Size > maxUploadSize {
		return AgentMeta{}, fmt.Errorf("file exceeds max size (%d bytes)", maxUploadSize)
	}

	tmpDir, err := os.MkdirTemp("", "agenthub-upload-*")
	if err != nil {
		return AgentMeta{}, err
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "upload.zip")
	if err := saveUpload(in.File, zipPath); err != nil {
		return AgentMeta{}, err
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := archive.ExtractZipFile(zipPath, extractDir); err != nil {
		return AgentMeta{}, fmt.Errorf("invalid zip: %w", err)
	}
	packageRoot, err := archive.StripSingleRootDir(extractDir)
	if err != nil {
		return AgentMeta{}, err
	}

	agentName := strings.TrimSpace(in.AgentName)
	if agentName == "" {
		return AgentMeta{}, fmt.Errorf("agentName is required")
	}
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return AgentMeta{}, err
	}

	if err := ValidatePackage(in.Category, packageRoot); err != nil {
		return AgentMeta{}, err
	}

	return s.store.SavePackage(agentName, in.Version, in.Category, packageRoot)
}

func saveUpload(fh *multipart.FileHeader, dest string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func (s *Service) List(category string) ([]AgentMeta, error) {
	agents, err := s.store.ListAgents()
	if err != nil {
		return nil, err
	}
	filter, err := NormalizeCategory(category)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(category) == "" {
		return agents, nil
	}
	out := make([]AgentMeta, 0, len(agents))
	for _, a := range agents {
		if a.Category == filter {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *Service) Get(agentName, version string) (AgentDetail, error) {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return AgentDetail{}, err
	}
	return s.store.GetAgent(agentName, version)
}

func (s *Service) GetFile(agentName, version, path string) ([]byte, error) {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return nil, err
	}
	return s.store.GetFile(agentName, version, path)
}

func (s *Service) UpdateMeta(agentName, displayName, summary, category string) (AgentMeta, error) {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return AgentMeta{}, err
	}
	return s.store.UpdateMeta(agentName, displayName, summary, category)
}

func (s *Service) WriteFile(agentName, version, relPath string, content []byte) error {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return err
	}
	return s.store.WriteFile(agentName, version, relPath, content)
}

func (s *Service) Delete(agentName string) error {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return err
	}
	return s.store.DeleteAgent(agentName)
}

func (s *Service) OpenDownload(agentName, version string) (path string, cleanup func(), err error) {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return "", nil, err
	}
	return s.store.OpenVersionZip(agentName, version)
}
