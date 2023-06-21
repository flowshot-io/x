package artifact

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type (
	Artifact interface {
		SaveToWriter(w io.Writer) error
		LoadFromReader(r io.ReadCloser) error
		ExtractToDirectory(dir string) error
		AddFile(virtualPath string, filePath string, content []byte) error
		ListFiles() ([]string, error)
		GetName() string
	}

	TarGzArtifact struct {
		name string
		vfs  afero.Fs
	}
)

func New(artifactName string) Artifact {
	if !strings.HasSuffix(artifactName, ".tar.gz") {
		artifactName = artifactName + ".tar.gz"
	}

	return &TarGzArtifact{
		name: artifactName,
		vfs:  afero.NewMemMapFs(),
	}
}

func NewFromTarGz(tarGzFilePath string) (Artifact, error) {
	if !strings.HasSuffix(tarGzFilePath, ".tar.gz") {
		return nil, fmt.Errorf("tar.gz file path must end with .tar.gz")
	}

	artifact := New(filepath.Base(tarGzFilePath))

	tarGzFile, err := os.Open(tarGzFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening tar.gz file: %w", err)
	}
	defer tarGzFile.Close()

	err = artifact.LoadFromReader(tarGzFile)
	if err != nil {
		return nil, fmt.Errorf("error loading artifact from tar.gz file: %w", err)
	}

	return artifact, nil
}

func NewWithPaths(artifactName string, paths []string) (Artifact, error) {
	artifact := New(artifactName)

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("error stating path: %s, error: %w", path, err)
		}

		if info.IsDir() {
			err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					content, err := os.ReadFile(subPath)
					if err != nil {
						return fmt.Errorf("error reading file: %s, error: %w", subPath, err)
					}
					relativePath, err := filepath.Rel(path, subPath)
					if err != nil {
						return fmt.Errorf("error creating relative file path: %s, error: %w", subPath, err)
					}
					err = artifact.AddFile(relativePath, subPath, content)
					if err != nil {
						return fmt.Errorf("error adding file: %s, error: %w", subPath, err)
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("error walking directory: %s, error: %w", path, err)
			}
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("error reading file: %s, error: %w", path, err)
			}
			err = artifact.AddFile("", path, content)
			if err != nil {
				return nil, fmt.Errorf("error adding file: %s, error: %w", path, err)
			}
		}
	}

	return artifact, nil
}

// AddFile adds a file to the artifact
func (a *TarGzArtifact) AddFile(virtualPath string, filePath string, content []byte) error {
	if virtualPath == "" {
		virtualPath = "/"
	}

	// Ensure the filePath is a directory and is clean
	virtualPath = path.Clean("/" + filepath.Dir(virtualPath))

	// Join the virtualPath and the file name
	virtualFilePath := path.Join(virtualPath, filepath.Base(filePath))

	// Ensure the parent directory structure exists
	dirPath := filepath.Dir(virtualFilePath)
	if err := a.vfs.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("error creating directories in the virtual file system: %w", err)
	}

	file, err := a.vfs.Create(virtualFilePath)
	if err != nil {
		return fmt.Errorf("error creating file in the virtual file system: %w", err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("error writing content to the virtual file: %w", err)
	}

	return nil
}

func (a *TarGzArtifact) ExtractToDirectory(outputDir string) error {
	return afero.Walk(a.vfs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		outPath := filepath.Join(outputDir, path)

		if info.IsDir() {
			return os.MkdirAll(outPath, os.ModePerm)
		}

		inFile, err := a.vfs.Open(path)
		if err != nil {
			return err
		}
		defer inFile.Close()

		outFile, err := os.Create(outPath)
		if err != nil {
			fmt.Println("error creating file: ", err)
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, inFile)

		return err
	})
}

func (a *TarGzArtifact) SaveToWriter(writer io.Writer) error {
	gzWriter := gzip.NewWriter(writer)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	err := afero.Walk(a.vfs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(path)

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		file, err := a.vfs.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})

	return err
}

func (a *TarGzArtifact) LoadFromReader(reader io.ReadCloser) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			err = a.vfs.MkdirAll(header.Name, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := a.vfs.Create(header.Name)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *TarGzArtifact) ListFiles() ([]string, error) {
	var fileList []string

	err := afero.Walk(a.vfs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileList = append(fileList, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileList, nil
}

func (a *TarGzArtifact) GetName() string {
	return a.name
}
