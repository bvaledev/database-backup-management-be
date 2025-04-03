package backup

import (
	"compress/gzip"
	"io"
	"os"
	"strings"
)

func CompressToGzip(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	_, err = io.Copy(gw, in)
	if err != nil {
		return err
	}

	return os.Remove(source)
}

func DecompressGzip(gzFile string) (string, error) {
	f, err := os.Open(gzFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	outputFile := strings.TrimSuffix(gzFile, ".gz")
	out, err := os.Create(outputFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, gr); err != nil {
		return "", err
	}

	return outputFile, nil
}
