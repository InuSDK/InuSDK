package candidate

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func Install(sdk, version, url, checksum, binPath string) error {
	baseDir := viper.GetString("base_dir")
	downloadsDir := filepath.Join(baseDir, "downloads")
	candidateDir := filepath.Join(baseDir, "candidates", sdk, version)

	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return fmt.Errorf("Could not create downloads directory: %w", err)
	}

	if err := os.MkdirAll(candidateDir, 0755); err != nil {
		return fmt.Errorf("Could not create candidate directory: %w", err)
	}

	// Download the sdk
	ext := resolveExt(url)
	tempFile := filepath.Join(downloadsDir, fmt.Sprintf("%s-%s%s", sdk, version, ext))

	fmt.Printf("Downloading %s %s. . .\n", sdk, version)
	if err := download(url, tempFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("Checksum mismatch: %w", err)
	}
	fmt.Println("Checksum verified")

	fmt.Println("Extracting . . .")
	if err := extract(tempFile, candidateDir); err != nil {
		os.RemoveAll(candidateDir)
		return fmt.Errorf("Extraction failed: %w", err)
	}

	os.Remove(tempFile)

	fmt.Printf("%s %s installed to %s\n", sdk, version, candidateDir)
	return nil
}

func download(url, destination string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer _file.Close()

	_, err = io.Copy(_file, resp.Body)
	return err
}

func verifyChecksum(filePath, expected string) error {
	parts := strings.SplitN(expected, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("Invalid checksum format, expected sha256:<hash>")
	}

	_file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer _file.Close()

	_hash := sha256.New()
	if _, err := io.Copy(_hash, _file); err != nil {
		return err
	}

	actual := hex.EncodeToString(_hash.Sum(nil))
	if actual != parts[1] {
		return fmt.Errorf("Expected %s, got %s", parts[1], actual)
	}

	return nil
}

func resolveExt(url string) string {
	if strings.HasSuffix(url, ".zip") {
		return ".zip"
	}
	return ".tar.gz"
}

func extract(src, destination string) error {
	if strings.HasSuffix(src, ".zip") {
		return extractZip(src, destination)
	}
	return extractTarGz(src, destination)
}

func extractZip(src, destination string) error {
	fileReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	for _, _file := range fileReader.File {
		// strip top-level directoy from path
		parts := strings.SplitN(_file.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		target := filepath.Join(destination, parts[1])

		if _file.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		out, err := os.Create(target)
		if err != nil {
			return err
		}

		ReadCloser, err := _file.Open()
		if err != nil {
			out.Close()
			return err
		}

		_, err = io.Copy(out, ReadCloser)
		ReadCloser.Close()
		out.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func extractTarGz(src, destination string) error {
	_file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer _file.Close()

	gz, err := gzip.NewReader(_file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tarFile := tar.NewReader(gz)

	for {
		header, err := tarFile.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		parts := strings.SplitN(header.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		target := filepath.Join(destination, parts[1])

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			out, err := os.Create(target)
			if err != nil {
				return err
			}

			_, err = io.Copy(out, tarFile)
			out.Close()

			if err != nil {
				return err
			}

			// Keep the executable permissions.
			os.Chmod(target, os.FileMode(header.Mode))
		}
	}

	return nil
}
