package utils

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexmullins/zip"
)

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()

}

func Zip(src, dst, passwd string) error {
	zipfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if IsDir(path) {
			return nil
		}
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(path, filepath.Dir(src)+"/")
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		} // 设置密码
		if passwd != "" {
			header.SetPassword(passwd)
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err = io.Copy(writer, file); err != nil {
				return err
			}
		}
		return err
	})
	return err
}
