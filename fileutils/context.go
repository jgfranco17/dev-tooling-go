package fileutils

import (
	"context"
	"io/fs"
)

type contextKey string

const (
	rootDirKey contextKey = "rootDir"
)

// ApplyRootDirToContext adds the provided fs.FS to the context under a key.
func ApplyRootDirToContext(ctx context.Context, files fs.FS) context.Context {
	ctx = context.WithValue(ctx, rootDirKey, files)
	return ctx
}

// RootDirFromContext retrieves the fs.FS from the context.
// It panics if the value is not found, is of the wrong type, or is nil.
func RootDirFromContext(ctx context.Context) fs.FS {
	rootDir, ok := ctx.Value(rootDirKey).(fs.FS)
	if !ok || rootDir == nil {
		panic("No root dir found in context, bad code path")
	}
	return rootDir
}
