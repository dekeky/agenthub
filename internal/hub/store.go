package hub

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agenthub/internal/archive"
)

const (
	agentsDirName = "agents"
	metaFileName  = "meta.json"
)

type AgentMeta struct {
	AgentName     string    `json:"agentName"`
	Category      string    `json:"category"`
	DisplayName   string    `json:"displayName"`
	Summary       string    `json:"summary"`
	LatestVersion string    `json:"latestVersion"`
	Versions      []string  `json:"versions"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type AgentDetail struct {
	AgentMeta
	Files []FileEntry `json:"files"`
}

type FileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Dir  bool   `json:"dir"`
}

type Store struct {
	root string
}

func NewStore(root string) (*Store, error) {
	agentsRoot := filepath.Join(root, agentsDirName)
	if err := os.MkdirAll(agentsRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &Store{root: agentsRoot}, nil
}

func (s *Store) agentDir(agentName string) string {
	return filepath.Join(s.root, agentName)
}

func (s *Store) versionDir(agentName, version string) string {
	return filepath.Join(s.agentDir(agentName), "versions", version)
}

func (s *Store) metaPath(agentName string) string {
	return filepath.Join(s.agentDir(agentName), metaFileName)
}

func (s *Store) ListAgents() ([]AgentMeta, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, err
	}
	out := make([]AgentMeta, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		meta, err := s.readMeta(e.Name())
		if err != nil {
			continue
		}
		out = append(out, meta)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].AgentName < out[j].AgentName
	})
	return out, nil
}

func (s *Store) GetAgent(agentName, version string) (AgentDetail, error) {
	meta, err := s.readMeta(agentName)
	if err != nil {
		return AgentDetail{}, err
	}
	version = resolveListingVersion(meta, version)
	files, err := s.listFiles(agentName, version)
	if err != nil {
		return AgentDetail{}, err
	}
	return AgentDetail{AgentMeta: meta, Files: files}, nil
}

func (s *Store) GetFile(agentName, version, relPath string) ([]byte, error) {
	root, err := s.resolveVersionRoot(agentName, version)
	if err != nil {
		return nil, err
	}
	safe, err := archive.SafeRelativePath(root, relPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(safe)))
}

func (s *Store) SavePackage(agentName, version, category, sourceDir string) (AgentMeta, error) {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return AgentMeta{}, err
	}
	category, err := NormalizeCategory(category)
	if err != nil {
		return AgentMeta{}, err
	}
	version = strings.TrimSpace(version)
	if version == "" {
		version = time.Now().UTC().Format("20060102.150405")
	}
	if err := archive.ValidateIdentifier(version); err != nil {
		return AgentMeta{}, fmt.Errorf("invalid version: %w", err)
	}

	target := s.versionDir(agentName, version)
	if err := os.RemoveAll(target); err != nil {
		return AgentMeta{}, err
	}
	if err := copyTree(sourceDir, target); err != nil {
		_ = os.RemoveAll(target)
		return AgentMeta{}, err
	}

	meta, err := s.readMeta(agentName)
	if err != nil {
		meta = AgentMeta{AgentName: agentName, Category: category}
	}
	if meta.DisplayName == "" {
		meta.DisplayName = agentName
	}
	meta.Category = category
	meta.LatestVersion = version
	meta.UpdatedAt = time.Now().UTC()
	meta.Versions = appendVersion(meta.Versions, version)
	if err := s.writeMeta(meta); err != nil {
		return AgentMeta{}, err
	}
	return meta, nil
}

func (s *Store) UpdateMeta(agentName, displayName, summary, category string) (AgentMeta, error) {
	meta, err := s.readMeta(agentName)
	if err != nil {
		return AgentMeta{}, err
	}
	if displayName != "" {
		meta.DisplayName = displayName
	}
	if summary != "" {
		meta.Summary = summary
	}
	if category != "" {
		cat, err := NormalizeCategory(category)
		if err != nil {
			return AgentMeta{}, err
		}
		meta.Category = cat
	}
	meta.UpdatedAt = time.Now().UTC()
	if err := s.writeMeta(meta); err != nil {
		return AgentMeta{}, err
	}
	return meta, nil
}

func (s *Store) WriteFile(agentName, version, relPath string, content []byte) error {
	root, err := s.resolveVersionRoot(agentName, version)
	if err != nil {
		return err
	}
	safe, err := archive.SafeRelativePath(root, relPath)
	if err != nil {
		return err
	}
	fullPath := filepath.Join(root, filepath.FromSlash(safe))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return err
	}
	// Update meta timestamp
	meta, err := s.readMeta(agentName)
	if err != nil {
		return err
	}
	meta.UpdatedAt = time.Now().UTC()
	return s.writeMeta(meta)
}

func (s *Store) DeleteAgent(agentName string) error {
	if err := archive.ValidateIdentifier(agentName); err != nil {
		return err
	}
	agentDir := s.agentDir(agentName)
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		return fmt.Errorf("agent %q not found", agentName)
	}
	return os.RemoveAll(agentDir)
}

func (s *Store) OpenVersionZip(agentName, version string) (string, func(), error) {
	root, err := s.resolveVersionRoot(agentName, version)
	if err != nil {
		return "", nil, err
	}
	tmp, err := os.CreateTemp("", "agenthub-*.zip")
	if err != nil {
		return "", nil, err
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	zipWriter := zip.NewWriter(tmp)
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			_, err = zipWriter.Create(rel + "/")
			return err
		}
		w, err := zipWriter.Create(rel)
		if err != nil {
			return err
		}
		if err := func() error {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(w, f)
			return err
		}(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = zipWriter.Close()
		_ = tmp.Close()
		cleanup()
		return "", nil, err
	}
	if err := zipWriter.Close(); err != nil {
		_ = tmp.Close()
		cleanup()
		return "", nil, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", nil, err
	}
	return tmpPath, cleanup, nil
}

func (s *Store) resolveVersionRoot(agentName, version string) (string, error) {
	meta, err := s.readMeta(agentName)
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(version)
	if v == "" || v == "latest" {
		v = meta.LatestVersion
	}
	if v == "" {
		return "", fmt.Errorf("agent %q has no versions", agentName)
	}
	root := s.versionDir(agentName, v)
	if _, err := os.Stat(root); err != nil {
		return "", fmt.Errorf("version %q not found for agent %q", v, agentName)
	}
	return archive.StripSingleRootDir(root)
}

func (s *Store) readMeta(agentName string) (AgentMeta, error) {
	data, err := os.ReadFile(s.metaPath(agentName))
	if err != nil {
		return AgentMeta{}, fmt.Errorf("agent %q not found", agentName)
	}
	var meta AgentMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return AgentMeta{}, err
	}
	if meta.AgentName == "" {
		var legacy struct {
			Slug string `json:"slug"`
		}
		if err := json.Unmarshal(data, &legacy); err == nil && legacy.Slug != "" {
			meta.AgentName = legacy.Slug
		}
	}
	if meta.AgentName == "" {
		meta.AgentName = agentName
	}
	if meta.Category == "" {
		meta.Category = CategoryPicoClaw
	}
	return meta, nil
}

func (s *Store) writeMeta(meta AgentMeta) error {
	if err := os.MkdirAll(s.agentDir(meta.AgentName), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(meta.AgentName), data, 0o644)
}

func resolveListingVersion(meta AgentMeta, version string) string {
	v := strings.TrimSpace(version)
	if v == "" || v == "latest" {
		v = meta.LatestVersion
	}
	if v == "" && len(meta.Versions) > 0 {
		v = meta.Versions[len(meta.Versions)-1]
	}
	return v
}

func isExcludedListPath(rel string) bool {
	return filepath.Base(rel) == metaFileName
}

func ancestorDirs(filePath string) []string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) <= 1 {
		return nil
	}
	out := make([]string, 0, len(parts)-1)
	for i := 1; i < len(parts); i++ {
		out = append(out, strings.Join(parts[:i], "/"))
	}
	return out
}

func (s *Store) listFiles(agentName, version string) ([]FileEntry, error) {
	root, err := s.resolveVersionRoot(agentName, version)
	if err != nil {
		return nil, err
	}

	var fileEntries []FileEntry
	var emptyDirCandidates []string
	impliedDirs := make(map[string]struct{})

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if isExcludedListPath(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			emptyDirCandidates = append(emptyDirCandidates, rel)
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		fileEntries = append(fileEntries, FileEntry{
			Path: rel,
			Size: info.Size(),
			Dir:  false,
		})
		for _, dir := range ancestorDirs(rel) {
			impliedDirs[dir] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := make([]FileEntry, 0, len(fileEntries)+len(emptyDirCandidates))
	if len(fileEntries) > 0 {
		for _, dir := range emptyDirCandidates {
			if _, implied := impliedDirs[dir]; implied {
				continue
			}
			out = append(out, FileEntry{Path: dir, Dir: true})
		}
	}
	for _, f := range fileEntries {
		out = append(out, f)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Dir != out[j].Dir {
			return out[i].Dir
		}
		return out[i].Path < out[j].Path
	})
	if out == nil {
		out = []FileEntry{}
	}
	return out, nil
}

func appendVersion(existing []string, version string) []string {
	for _, v := range existing {
		if v == version {
			return existing
		}
	}
	return append(existing, version)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
