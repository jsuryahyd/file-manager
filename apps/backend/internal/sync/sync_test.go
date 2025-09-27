package sync

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"file-manager-backend/internal/db"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDb(t *testing.T) *sql.DB {
	dbConn, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Copy init.sql to a temporary file in the in-memory filesystem
	initSQL, err := afero.Afero{Fs: afero.NewOsFs()}.ReadFile("../../database/init.sql")
	require.NoError(t, err)

	tmpDir, err := afero.TempDir(db.AppFs, "", "testdb")
	require.NoError(t, err)
	sqlPath := filepath.Join(tmpDir, "init.sql")
	afero.WriteFile(db.AppFs, sqlPath, initSQL, 0644)

	err = db.Migrate(dbConn, sqlPath, "")
	require.NoError(t, err)
	return dbConn
}

func TestSyncPeekMode(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	db.AppFs = fs // Use in-memory fs for tests
	dbConn := setupTestDb(t)
	defer dbConn.Close()

	srcDir := "/src"
	dstDir := "/dst"
	require.NoError(t, fs.MkdirAll(srcDir, 0755))
	require.NoError(t, fs.MkdirAll(dstDir, 0755))

	// Create files
	// file1: exists in src, not in dst -> should be copied
	afero.WriteFile(fs, "/src/file1.txt", []byte("file1"), 0644)
	// file2: exists in both, identical -> should be skipped
	afero.WriteFile(fs, "/src/file2.txt", []byte("file2"), 0644)
	afero.WriteFile(fs, "/dst/file2.txt", []byte("file2"), 0644)
	// file3: exists in both, different -> should be copied (overwrite=true)
	afero.WriteFile(fs, "/src/file3.txt", []byte("file3_src"), 0644)
	afero.WriteFile(fs, "/dst/file3.txt", []byte("file3_dst"), 0644)
	// file4: exists in src, but a dir with same name in dst -> should be copied with new name
	afero.WriteFile(fs, "/src/file4.txt", []byte("file4"), 0644)
	require.NoError(t, fs.Mkdir("/dst/file4.txt", 0755))

	// Create sync pair
	pairID, err := db.CreateSyncPair(dbConn, srcDir, dstDir)
	require.NoError(t, err)

	opts := Options{
		Recursive:       true,
		PeekMode:        true,
		Overwrite:       true,
		CheckDuplicates: true,
	}

	job, err := NewSyncJob(dbConn, srcDir, dstDir, pairID, opts)
	require.NoError(t, err)

	// Execute
	result, err := job.Run()
	require.NoError(t, err)

	// Assert
	assert.Empty(t, result.Errors)
	assert.ElementsMatch(t, []string{"file1.txt", "file3.txt", "file4.txt"}, result.FilesCopied)
	assert.ElementsMatch(t, []string{"file2.txt"}, result.FilesSkipped)

	// Assert that no files were actually copied
	_, err = fs.Stat("/dst/file1.txt")
	assert.True(t, os.IsNotExist(err))

	dstFile3Content, err := afero.ReadFile(fs, "/dst/file3.txt")
	require.NoError(t, err)
	assert.Equal(t, "file3_dst", string(dstFile3Content))
}

func TestSyncNameConflict(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	db.AppFs = fs // Use in-memory fs for tests
	dbConn := setupTestDb(t)
	defer dbConn.Close()

	srcDir := "/src"
	dstDir := "/dst"
	require.NoError(t, fs.MkdirAll(srcDir, 0755))
	require.NoError(t, fs.MkdirAll(dstDir, 0755))

	// Create a file in source
	afero.WriteFile(fs, "/src/file.txt", []byte("file content"), 0644)
	// Create a directory with the same name in destination
	require.NoError(t, fs.Mkdir("/dst/file.txt", 0755))

	// Create sync pair
	pairID, err := db.CreateSyncPair(dbConn, srcDir, dstDir)
	require.NoError(t, err)

	opts := Options{
		Recursive:       true,
		PeekMode:        false, // We want to actually copy the file
		Overwrite:       true,
		CheckDuplicates: true,
	}

	job, err := NewSyncJob(dbConn, srcDir, dstDir, pairID, opts)
	require.NoError(t, err)

	// Execute
	result, err := job.Run()
	require.NoError(t, err)

	// Assert
	assert.Empty(t, result.Errors)
	assert.Contains(t, result.FilesCopied, "file.txt")

	// Assert that the file was copied with a new name
	_, err = fs.Stat("/dst/file (1).txt")
	assert.NoError(t, err)
}
