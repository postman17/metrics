package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSubscriber_getID(t *testing.T) {
	s := &FileSubscriber{ID: "file-1"}
	assert.Equal(t, "file-1", s.getID())
}

func TestFileSubscriber_send(t *testing.T) {
	t.Run("writes to file successfully", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "audit.jsonl")
		s := &FileSubscriber{ID: "file-1", FilePath: fp}

		err := s.send("192.168.1.1", []string{"cpu", "mem"})
		require.NoError(t, err)

		data, err := os.ReadFile(fp)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		assert.Len(t, lines, 1)
		assert.Contains(t, lines[0], `"ip_address":"192.168.1.1"`)
		assert.Contains(t, lines[0], `"metrics":["cpu","mem"]`)
		assert.Contains(t, lines[0], `"ts":`)
	})

	t.Run("creates directory if not exists", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "subdir", "nested", "audit.jsonl")
		s := &FileSubscriber{ID: "file-2", FilePath: fp}

		err := s.send("10.0.0.1", []string{"disk"})
		require.NoError(t, err)

		_, err = os.Stat(fp)
		assert.NoError(t, err)
	})

	t.Run("appends multiple entries", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "audit.jsonl")
		s := &FileSubscriber{ID: "file-3", FilePath: fp}

		err := s.send("10.0.0.1", []string{"cpu"})
		require.NoError(t, err)
		err = s.send("10.0.0.2", []string{"mem"})
		require.NoError(t, err)

		data, err := os.ReadFile(fp)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		assert.Len(t, lines, 2)
		assert.Contains(t, lines[0], `"ip_address":"10.0.0.1"`)
		assert.Contains(t, lines[1], `"ip_address":"10.0.0.2"`)
	})

	t.Run("invalid path returns error", func(t *testing.T) {
		s := &FileSubscriber{ID: "file-4", FilePath: "/dev/null/invalid/audit.jsonl"}
		err := s.send("10.0.0.1", []string{"cpu"})
		assert.Error(t, err)
	})
}
