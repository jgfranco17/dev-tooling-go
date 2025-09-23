package logging

import (
	"bytes"
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestApplyToContext(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, logrus.TraceLevel)
	ctx := WithContext(context.Background(), logger)
	assert.Equal(t, logger, FromContext(ctx))
}

func TestFromContext(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, logrus.TraceLevel)
	ctx := WithContext(context.Background(), logger)
	assert.Equal(t, logger, FromContext(ctx))
}
