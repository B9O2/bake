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
	// 创建文件
	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	// 将 tar 包使用 gzip 压缩，其实添加压缩功能很简单，
	// 只需要在 fw 和 tw 之前加上一层压缩就行了，和 Linux 的管道的感觉类似
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// 创建 Tar.Writer 结构
	tw := tar.NewWriter(gw)
	// 如果需要启用 gzip 将上面代码注释，换成下面的

	defer tw.Close()

	// 下面就该开始处理数据了，这里的思路就是递归处理目录及目录下的所有文件和目录
	// 这里可以自己写个递归来处理，不过 Golang 提供了 filepath.Walk 函数，可以很方便的做这个事情
	// 直接将这个函数的处理结果返回就行，需要传给它一个源文件或目录，它就可以自己去处理
	// 我们就只需要去实现我们自己的 打包逻辑即可，不需要再去路径相关的事情
	return filepath.Walk(src, func(fileName string, fi os.FileInfo, err error) error {
		// 因为这个闭包会返回个 error ，所以先要处理一下这个
		if err != nil {
			return err
		}

		// 这里就不需要我们自己再 os.Stat 了，它已经做好了，我们直接使用 fi 即可
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		// 这里需要处理下 hdr 中的 Name，因为默认文件的名字是不带路径的，
		// 打包之后所有文件就会堆在一起，这样就破坏了原本的目录结果
		// 例如： 将原本 hdr.Name 的 syslog 替换程 log/syslog
		// 这个其实也很简单，回调函数的 fileName 字段给我们返回来的就是完整路径的 log/syslog
		// strings.TrimPrefix 将 fileName 的最左侧的 / 去掉，
		// 熟悉 Linux 的都知道为什么要去掉这个
		rel, err := filepath.Rel(src, fileName)
		if err != nil {
			return err
		}
		hdr.Name = rel

		// 写入文件信息
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		// 判断下文件是否是标准文件，如果不是就不处理了，
		// 如： 目录，这里就只记录了文件信息，不会执行下面的 copy
		if !fi.Mode().IsRegular() {
			return nil
		}

		// 打开文件
		fr, err := os.Open(fileName)
		defer fr.Close()
		if err != nil {
			return err
		}

		// copy 文件数据到 tw
		_, err = io.Copy(tw, fr)
		if err != nil {
			return err
		}

		// 记录下过程，这个可以不记录，这个看需要，这样可以看到打包的过程
		//log.Printf("成功打包 %s ，共写入了 %d 字节的数据\n", fileName, n)

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
