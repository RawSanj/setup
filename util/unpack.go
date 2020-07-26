package util

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func ExtractTarGz(filename, extractDir string) error {

	gzipStream, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error Opening File", filename)
		return err
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		fmt.Println("ExtractTarGz: NewReader failed")
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			fmt.Println("Archive Extracted at", extractDir)
			break
		}

		if err != nil {
			fmt.Println("ExtractTarGz: Next() failed:", err.Error())
			return err
		}

		switch header.Typeflag {

		case tar.TypeDir:
			if err := os.MkdirAll(filepath.FromSlash(extractDir+"/"+header.Name), 0755); err != nil {
				fmt.Println("ExtractTarGz: Mkdir() failed:", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.FromSlash(extractDir + "/" + header.Name))
			if err != nil {
				fmt.Println("ExtractTarGz: Create() failed:", err.Error())
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				fmt.Println("ExtractTarGz: Copy() failed:", err.Error())
				return err
			}
			_ = outFile.Close()

		default:
			fmt.Println(fmt.Sprintf("ExtractTarGz: uknown type: %s in %s",
				header.Typeflag,
				header.Name))
		}
	}

	return moveSingleDirToParent(extractDir)
}

func Unzip(filename, extractDir string) error {

	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(extractDir, f.Name)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, f.Mode())
			if err != nil {
				return err
			}
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				fmt.Println(err)
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return moveSingleDirToParent(extractDir)
}

func moveSingleDirToParent(extractDir string) error {

	files, err := ioutil.ReadDir(extractDir)
	if err != nil {
		return err
	}

	if len(files) == 1 && files[0].IsDir() {

		parentDirPath := filepath.Dir(extractDir)
		extractDirName := filepath.Base(extractDir)
		tempPath := filepath.FromSlash(parentDirPath + "/tmp/" + extractDirName)
		oldPath := filepath.FromSlash(extractDir + "/" + files[0].Name())

		// create parent/tmp directory
		err := os.MkdirAll(filepath.FromSlash(parentDirPath + "/tmp"), 0755)
		if err != nil {
			return err
		}

		// move parent/extractDir to parent/tmp/extractDir
		err = os.Rename(oldPath, tempPath)
		if err != nil {
			return err
		}

		// delete parent/extractDir
		err = os.Remove(filepath.FromSlash(parentDirPath + "/" + extractDirName))
		if err != nil {
			return err
		}

		// move parent/tmp/extractDir to parent/extractDir
		err = os.Rename(tempPath, extractDir)
		if err != nil {
			return err
		}

		// delete parent/tmp
		err = os.Remove(filepath.FromSlash(parentDirPath + "/tmp"))
		if err != nil {
			return err
		}
	}
	return nil
}
