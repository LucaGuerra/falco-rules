package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path"
)

func tarGzSingleFile(outputPath string, fileName string) error {
	var file *os.File
	var err error
	var writer *gzip.Writer

	if file, err = os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return err
	}
	defer file.Close()

	if writer, err = gzip.NewWriterLevel(file, gzip.DefaultCompression); err != nil {
		return err
	}
	defer writer.Close()

	tw := tar.NewWriter(writer)
	defer tw.Close()

	body, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	hdr := &tar.Header{
		Name: path.Base(fileName),
		Mode: int64(0644),
		Size: int64(len(body)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(body); err != nil {
		return err
	}

	return nil
}
