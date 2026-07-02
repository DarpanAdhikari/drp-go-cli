package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultReleaseRepo = "DarpanAdhikari/drp-go-cli"

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Download the latest drp release and replace this binary",
	RunE: func(c *cobra.Command, args []string) error {
		repo, _ := c.Flags().GetString("repo")
		binDir, _ := c.Flags().GetString("bin-dir")
		assetURL, _ := c.Flags().GetString("asset-url")

		if binDir == "" {
			if exe, err := os.Executable(); err == nil && filepath.Base(exe) == executableName("drp") {
				binDir = filepath.Dir(exe)
			} else {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				binDir = filepath.Join(home, ".local", "bin")
			}
		}

		if assetURL == "" {
			asset, version, err := latestReleaseAsset(repo)
			if err != nil {
				output.Fail("%v", err)
				return err
			}
			assetURL = asset
			output.Info("Latest release: %s", version)
		}

		dest := filepath.Join(binDir, executableName("drp"))
		tmp, err := downloadAsset(assetURL)
		if err != nil {
			output.Fail("%v", err)
			return err
		}
		defer os.Remove(tmp)

		binary, cleanup, err := extractBinary(tmp)
		if err != nil {
			output.Fail("%v", err)
			return err
		}
		defer cleanup()

		if err := copyExecutable(binary, dest); err != nil {
			output.Fail("%v", err)
			return err
		}

		output.Success("Upgraded drp at %s", dest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().String("repo", defaultReleaseRepo, "GitHub repository in owner/name form")
	upgradeCmd.Flags().String("bin-dir", "", "Directory containing drp (default: current drp dir or ~/.local/bin)")
	upgradeCmd.Flags().String("asset-url", "", "Direct release asset URL (mostly for testing)")
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func latestReleaseAsset(repo string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "drp-upgrade/"+Version)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("upgrade: fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("upgrade: GitHub returned %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("upgrade: decode release: %w", err)
	}
	wantOS := runtime.GOOS
	wantArch := runtime.GOARCH
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, wantOS) &&
			strings.Contains(name, wantArch) &&
			!strings.Contains(name, "sha256") &&
			!strings.Contains(name, "checksums") {
			return asset.BrowserDownloadURL, release.TagName, nil
		}
	}
	return "", "", fmt.Errorf("upgrade: no release asset found for %s/%s in %s", wantOS, wantArch, repo)
}

func downloadAsset(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "drp-upgrade/"+Version)

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upgrade: download asset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upgrade: asset download returned %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "drp-upgrade-*"+assetExt(url))
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", fmt.Errorf("upgrade: save asset: %w", err)
	}
	return tmp.Name(), nil
}

func assetExt(url string) string {
	name := strings.ToLower(filepath.Base(strings.Split(url, "?")[0]))
	switch {
	case strings.HasSuffix(name, ".tar.gz"):
		return ".tar.gz"
	case strings.HasSuffix(name, ".tgz"):
		return ".tgz"
	case strings.HasSuffix(name, ".zip"):
		return ".zip"
	default:
		return ""
	}
}

func extractBinary(path string) (string, func(), error) {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".zip") {
		return extractZip(path)
	}
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return extractTarGz(path)
	}

	if err := os.Chmod(path, 0o755); err != nil {
		return "", func() {}, err
	}
	return path, func() {}, nil
}

func extractZip(path string) (string, func(), error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", func() {}, err
	}
	defer r.Close()

	dir, err := os.MkdirTemp("", "drp-upgrade-extract-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	for _, file := range r.File {
		if file.FileInfo().IsDir() || !looksLikeDRPBinary(file.Name) {
			continue
		}
		out := filepath.Join(dir, filepath.Base(file.Name))
		if err := unzipFile(file, out); err != nil {
			cleanup()
			return "", cleanup, err
		}
		return out, cleanup, os.Chmod(out, 0o755)
	}
	cleanup()
	return "", cleanup, fmt.Errorf("upgrade: archive does not contain a drp binary")
}

func unzipFile(file *zip.File, dest string) error {
	in, err := file.Open()
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func extractTarGz(path string) (string, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return "", func() {}, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", func() {}, err
	}
	defer gz.Close()

	dir, err := os.MkdirTemp("", "drp-upgrade-extract-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			cleanup()
			return "", cleanup, err
		}
		if header.Typeflag != tar.TypeReg || !looksLikeDRPBinary(header.Name) {
			continue
		}
		out := filepath.Join(dir, filepath.Base(header.Name))
		if err := untarFile(tr, out); err != nil {
			cleanup()
			return "", cleanup, err
		}
		return out, cleanup, os.Chmod(out, 0o755)
	}
	cleanup()
	return "", cleanup, fmt.Errorf("upgrade: archive does not contain a drp binary")
}

func untarFile(in io.Reader, dest string) error {
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func looksLikeDRPBinary(name string) bool {
	base := strings.ToLower(filepath.Base(name))
	if strings.Contains(base, "sha256") || strings.Contains(base, "checksum") {
		return false
	}
	return base == "drp" || base == "drp.exe" || strings.HasPrefix(base, "drp-") || strings.HasPrefix(base, "drp_") || strings.HasPrefix(base, "drp-go-cli")
}
