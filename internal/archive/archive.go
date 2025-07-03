// internal/archive/archive.go
package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

func Pack(srcDir, destTarGz string) error {
	out, err := os.Create(destTarGz)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// walk srcDir and add files to tw…
	// (omitted for brevity; use filepath.Walk)
	return nil
}

func Unpack(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		// create files/directories based on hdr and tr…
	}
	return nil
}
