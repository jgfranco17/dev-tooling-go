package fileutils

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRootDirToContext(t *testing.T) {
	tests := []struct {
		name   string
		rootFS fs.FS
	}{
		{
			name: "add valid MapFS",
			rootFS: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("content")},
			},
		},
		{
			name:   "add empty MapFS",
			rootFS: fstest.MapFS{},
		},
		{
			name:   "add nil filesystem",
			rootFS: nil,
		},
		{
			name: "add MapFS with nested structure",
			rootFS: fstest.MapFS{
				"dir/file1.txt": &fstest.MapFile{Data: []byte("content1")},
				"dir/file2.txt": &fstest.MapFile{Data: []byte("content2")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := ApplyRootDirToContext(ctx, tt.rootFS)

			assert.NotNil(t, newCtx)
			assert.NotEqual(t, ctx, newCtx)

			retrieved := newCtx.Value(rootDirKey)
			assert.Equal(t, tt.rootFS, retrieved)
		})
	}
}

func TestRootDirFromContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectPanic bool
	}{
		{
			name: "retrieve existing filesystem",
			setupCtx: func() context.Context {
				rootFS := fstest.MapFS{
					"file.txt": &fstest.MapFile{Data: []byte("test")},
				}
				return ApplyRootDirToContext(context.Background(), rootFS)
			},
			expectPanic: false,
		},
		{
			name: "retrieve from empty context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectPanic: true,
		},
		{
			name: "retrieve from context with wrong key",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "wrongKey", "value")
			},
			expectPanic: true,
		},
		{
			name: "retrieve from context with wrong type",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), rootDirKey, "not a filesystem")
			},
			expectPanic: true,
		},
		{
			name: "retrieve nil filesystem",
			setupCtx: func() context.Context {
				return ApplyRootDirToContext(context.Background(), nil)
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()

			if tt.expectPanic {
				assert.Panics(t, func() {
					RootDirFromContext(ctx)
				})
			} else {
				assert.NotPanics(t, func() {
					rootFS := RootDirFromContext(ctx)
					assert.NotNil(t, rootFS)
				})
			}
		})
	}
}

func TestRootDirFromContext_Read(t *testing.T) {
	content := []byte("test content")
	rootFS := fstest.MapFS{
		"test.txt": &fstest.MapFile{Data: content},
	}

	ctx := ApplyRootDirToContext(context.Background(), rootFS)
	retrievedFS := RootDirFromContext(ctx)

	file, err := retrievedFS.Open("test.txt")
	require.NoError(t, err)
	defer file.Close()

	info, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name())

	data := make([]byte, len(content))
	n, err := file.Read(data)
	require.NoError(t, err)
	assert.Equal(t, len(content), n)
	assert.Equal(t, content, data)
}

func TestContextIsolation(t *testing.T) {
	rootFS1 := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("content1")},
	}
	rootFS2 := fstest.MapFS{
		"file2.txt": &fstest.MapFile{Data: []byte("content2")},
	}

	ctx1 := ApplyRootDirToContext(context.Background(), rootFS1)
	ctx2 := ApplyRootDirToContext(context.Background(), rootFS2)

	retrieved1 := RootDirFromContext(ctx1)
	retrieved2 := RootDirFromContext(ctx2)

	assert.Equal(t, rootFS1, retrieved1)
	assert.Equal(t, rootFS2, retrieved2)
	assert.NotEqual(t, retrieved1, retrieved2)

	file1, err := retrieved1.Open("file1.txt")
	require.NoError(t, err)
	file1.Close()

	file2, err := retrieved2.Open("file2.txt")
	require.NoError(t, err)
	file2.Close()

	_, err = retrieved1.Open("file2.txt")
	assert.Error(t, err)

	_, err = retrieved2.Open("file1.txt")
	assert.Error(t, err)
}

func TestContextChaining(t *testing.T) {
	rootFS := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("content")},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "otherKey", "otherValue")
	ctx = ApplyRootDirToContext(ctx, rootFS)

	retrieved := RootDirFromContext(ctx)
	require.NotNil(t, retrieved)
	assert.Equal(t, rootFS, retrieved)

	otherValue := ctx.Value("otherKey")
	assert.Equal(t, "otherValue", otherValue)
}

func TestContextOverwrite(t *testing.T) {
	rootFS1 := fstest.MapFS{
		"first.txt": &fstest.MapFile{Data: []byte("first")},
	}
	rootFS2 := fstest.MapFS{
		"second.txt": &fstest.MapFile{Data: []byte("second")},
	}

	ctx := context.Background()
	ctx = ApplyRootDirToContext(ctx, rootFS1)
	ctx = ApplyRootDirToContext(ctx, rootFS2)

	retrieved := RootDirFromContext(ctx)
	assert.Equal(t, rootFS2, retrieved)

	_, err := retrieved.Open("second.txt")
	assert.NoError(t, err)

	_, err = retrieved.Open("first.txt")
	assert.Error(t, err)
}

func TestContextKey(t *testing.T) {
	assert.Equal(t, contextKey("rootDir"), rootDirKey)
	assert.NotEqual(t, "rootDir", rootDirKey)
}
