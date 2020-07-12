package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ExtractTarGz(filename string, extractDir string) error {

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
	return nil
}
