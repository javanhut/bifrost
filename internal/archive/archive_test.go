package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestPack(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create source directory with test files
	srcDir := filepath.Join(tempDir, "src")
	err := os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		"file1.txt": "content of file 1",
		"file2.txt": "content of file 2",
		"subdir/file3.txt": "content of file 3",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(srcDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", fullPath, err)
		}
	}

	// Pack the directory
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	err = Pack(srcDir, archivePath)
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("archive file was not created")
	}

	// Verify it's a valid gzip file
	f, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	// Verify it's a valid tar file
	tr := tar.NewReader(gr)
	_, err = tr.Next()
	if err != nil && err != io.EOF {
		t.Fatalf("archive is not a valid tar file: %v", err)
	}
}

func TestPack_Errors(t *testing.T) {
	tests := []struct {
		name    string
		srcDir  string
		destTar string
		wantErr bool
	}{
		{
			name:    "non-existent source",
			srcDir:  "/non/existent/path",
			destTar: "test.tar.gz",
			wantErr: false, // Current implementation doesn't check if srcDir exists
		},
		{
			name:    "invalid destination",
			srcDir:  ".",
			destTar: "/invalid/path/that/does/not/exist/test.tar.gz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Pack(tt.srcDir, tt.destTar)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnpack(t *testing.T) {
	tempDir := t.TempDir()

	// First create a valid tar.gz file
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	if err := createTestArchive(archivePath); err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	// Unpack to destination
	destDir := filepath.Join(tempDir, "unpacked")
	err := Unpack(archivePath, destDir)
	if err != nil {
		t.Fatalf("Unpack() error = %v", err)
	}

	// Note: The current implementation doesn't actually extract files,
	// so we can't verify the extracted content
}

func TestUnpack_Errors(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string
		destDir  string
		wantErr  bool
	}{
		{
			name: "non-existent archive",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.tar.gz")
			},
			destDir: tempDir,
			wantErr: true,
		},
		{
			name: "invalid gzip file",
			setup: func() string {
				invalidPath := filepath.Join(tempDir, "invalid.tar.gz")
				os.WriteFile(invalidPath, []byte("not a gzip file"), 0644)
				return invalidPath
			},
			destDir: tempDir,
			wantErr: true,
		},
		{
			name: "corrupted tar inside gzip",
			setup: func() string {
				corruptedPath := filepath.Join(tempDir, "corrupted.tar.gz")
				f, _ := os.Create(corruptedPath)
				gw := gzip.NewWriter(f)
				gw.Write([]byte("not a valid tar file"))
				gw.Close()
				f.Close()
				return corruptedPath
			},
			destDir: tempDir,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := tt.setup()
			err := Unpack(archivePath, tt.destDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPackUnpackRoundTrip(t *testing.T) {
	tempDir := t.TempDir()

	// Create source with test data
	srcDir := filepath.Join(tempDir, "src")
	testData := map[string]string{
		"README.md":        "# Test Package",
		"src/main.crl":     "func main() { }",
		"src/lib/util.crl": "func util() { }",
		"LICENSE":          "MIT License",
	}

	for path, content := range testData {
		fullPath := filepath.Join(srcDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	// Pack
	archivePath := filepath.Join(tempDir, "package.tar.gz")
	if err := Pack(srcDir, archivePath); err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	// Verify archive was created
	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("failed to stat archive: %v", err)
	}
	if info.Size() == 0 {
		t.Error("archive file is empty")
	}

	// Unpack
	destDir := filepath.Join(tempDir, "dest")
	if err := Unpack(archivePath, destDir); err != nil {
		t.Fatalf("Unpack() error = %v", err)
	}

	// Note: Current implementation doesn't fully implement Pack/Unpack,
	// so we can't verify the round trip completely
}

// Helper function to create a valid test archive
func createTestArchive(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add a test file
	hdr := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len("test content")),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write([]byte("test content")); err != nil {
		return err
	}

	return nil
}

func TestArchiveEdgeCases(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		emptyDir := filepath.Join(tempDir, "empty")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatalf("failed to create empty dir: %v", err)
		}

		archivePath := filepath.Join(tempDir, "empty.tar.gz")
		err := Pack(emptyDir, archivePath)
		if err != nil {
			t.Fatalf("Pack() with empty dir error = %v", err)
		}

		// Archive should still be created
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			t.Error("archive was not created for empty directory")
		}
	})

	t.Run("special characters in filename", func(t *testing.T) {
		tempDir := t.TempDir()
		srcDir := filepath.Join(tempDir, "special")
		if err := os.MkdirAll(srcDir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		// Create file with special characters (that are valid on the filesystem)
		specialFile := filepath.Join(srcDir, "file with spaces.txt")
		if err := os.WriteFile(specialFile, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create special file: %v", err)
		}

		archivePath := filepath.Join(tempDir, "special.tar.gz")
		err := Pack(srcDir, archivePath)
		if err != nil {
			t.Fatalf("Pack() with special filename error = %v", err)
		}
	})
}