package utils

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"
	"unsafe"
)

func CopyDirectory(source, destination string) error {
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destination, relativePath)
		if info.IsDir() {
			err = os.MkdirAll(destPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(path, destPath, info.Mode())
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func CopyFile(source, destination string, mode os.FileMode) error {
	dirPath, _ := filepath.Split(destination)
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if _, err = os.Stat(dirPath); os.IsNotExist(err) {
		// mkdir 创建目录，mkdirAll 可创建多层级目录
		os.MkdirAll(dirPath, os.ModePerm)
	}

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Chmod(mode)
	if err != nil {
		return err
	}

	return nil
}

func MakeTar(src, dst string) error {
	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()

	tw := tar.NewWriter(gw)

	defer tw.Close()

	return filepath.Walk(src, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, fileName)
		if err != nil {
			return err
		}
		hdr.Name = rel

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		fr, err := os.Open(fileName)
		defer fr.Close()
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, fr)
		if err != nil {
			return err
		}

		return nil
	})
}

// UnpackTar 解压时可选是否启用gzip解压。
func UnpackTar(tarFileName string, destDir string, useGZip bool) (err error) {
	// 打开tar文件
	fr, err := os.Open(tarFileName)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := fr.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()
	var tarReader *tar.Reader
	// 使用gzip解压
	if useGZip {
		gr, err := gzip.NewReader(fr)
		if err != nil {
			return err
		}
		defer func() {
			if err2 := gr.Close(); err2 != nil && err == nil {
				err = err2
			}
		}()
		tarReader = tar.NewReader(gr)
	} else {
		tarReader = tar.NewReader(fr)
	}

	// 循环读取
	for {
		header, err := tarReader.Next()
		switch {
		// 读取结束
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		targetFullPath := filepath.Join(destDir, header.Name)

		// 根据文件类型做处理，这里只处理目录和普通文件，如果需要处理其他类型文件，添加case即可
		switch header.Typeflag {
		case tar.TypeDir:
			// 是目录，不存在则创建
			if exists := DirExists(targetFullPath); !exists {
				if err = os.MkdirAll(targetFullPath, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			// 是普通文件，创建并将内容写入
			file, err := os.OpenFile(targetFullPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			// 循环内不能用defer，先关闭文件句柄
			if err2 := file.Close(); err2 != nil {
				return err2
			}
			// 这里再对文件copy的结果做判断
			if err != nil {
				return err
			}
		}
	}
}

func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	return (err == nil || os.IsExist(err)) && info.IsDir()
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func SaveFile(path string, content []byte, cover bool) error {
	yes, _ := FileExists(path)
	if yes { //已经存在
		if cover {
			_ = os.Remove(path)
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			_ = f.Chmod(os.ModePerm)
			defer f.Close()
			_, err = f.Write(content)
			if err != nil {
				return err
			}
			return nil
		} else {
			return errors.New("Create Failed:'" + path + "' exists")
		}
	}
	//不存在
	dir, fileName := filepath.Split(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errors.New("Create Failed:" + err.Error() + "dir:" + dir)
	}
	if fileName != "" {
		file, err := os.Create(path)
		if err != nil {
			return errors.New("Create Failed:" + err.Error())
		}
		_ = file.Chmod(os.ModePerm)
		defer file.Close()
		_, err = file.Write(content)
		if err != nil {
			return errors.New("Create Failed:" + err.Error())
		}
	} else {
		return errors.New("invalid path")
	}
	return nil
}

func RandStr(n int) string {
	// 6 bits to represent a letter index
	letterIdBits := 6
	// All 1-bits as many as letterIdBits
	letterIdMask := 1<<letterIdBits - 1
	letterIdMax := 63 / letterIdBits
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(int(cache) & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}
