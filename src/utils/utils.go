package utils

import (
	"CIHunter/src/config"
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/mholt/archiver/v4"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
)

func Init() {
	localRepos, _ := filepath.Glob(path.Join(config.DEV_SHM, "*:*"))
	for _, p := range localRepos {
		os.RemoveAll(p)
	}
}

func RandomString() string {
	const bytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, rand.Intn(10)+10)
	for i := range b {
		b[i] = bytes[rand.Intn(len(bytes))]
	}
	return string(b)
}

// define the archive format
var format = archiver.CompressedArchive{
	Compression: archiver.Gz{},
	Archival:    archiver.Tar{},
}

// SerializeRepo serialize the git repository into bytea
func SerializeRepo(repoName string) ([]byte, error) {
	sourcePath := path.Join(config.DEV_SHM, repoName)

	files, err := archiver.FilesFromDisk(nil, map[string]string{
		sourcePath: "",
	})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	if err := format.Archive(context.Background(), out, files); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeRepo deserialize the git repositories from bytes
func DeserializeRepo(source []byte) error {
	input := bytes.NewReader(source)

	handler := func(ctx context.Context, f archiver.File) error {
		// do something with the file
		filePath := path.Join(config.DEV_SHM, f.NameInArchive)
		fileHeader := f.Header.(*tar.Header)

		switch fileHeader.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
			return nil
		case tar.TypeReg, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return err
			}

			out, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer out.Close()

			if err = out.Chmod(f.Mode()); err != nil {
				return err
			}

			in, err := f.Open()

			if _, err = io.Copy(out, in); err != nil {
				return err
			}
			return nil
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return err
			}

			if err := os.Symlink(fileHeader.Linkname, filePath); err != nil {
				return err
			}
			return nil

		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return err
			}
			if err := os.Link(filepath.Join(filePath, fileHeader.Linkname), filePath); err != nil {
				return err
			}
			return nil
		default:
			return fmt.Errorf("%s: unknown type flag: %c", fileHeader.Name, fileHeader.Typeflag)
		}
	}

	err := format.Extract(context.Background(), input, nil, handler)
	if err != nil {
		return err
	}

	return nil
}
