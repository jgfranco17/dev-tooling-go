package commandline

import (
	"bytes"
	"context"
	"testing"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommandOptions_validate(t *testing.T) {
	tests := []struct {
		name    string
		options RootCommandOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			options: RootCommandOptions{
				Name:        "",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "",
			},
			wantErr: true,
		},
		{
			name: "empty description allowed",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		options RootCommandOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "invalid options - no name",
			options: RootCommandOptions{
				Name:        "",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "invalid options - no version",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := New(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cli)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cli)
				assert.NotNil(t, cli.root)
				assert.Equal(t, tt.options.Name, cli.root.Use)
				assert.Equal(t, tt.options.Version, cli.root.Version)
				assert.Equal(t, tt.options.Description, cli.root.Short)
			}
		})
	}
}

func TestNew_WithModifiers(t *testing.T) {
	var modifierCalled bool
	var cleanupCalled bool

	modifier := func(ctx context.Context) context.Context {
		modifierCalled = true
		return context.WithValue(ctx, "test", "value")
	}

	options := RootCommandOptions{
		Name:        "testcli",
		Description: "A test CLI application",
		Version:     "1.0.0",
		Modifiers:   []ContextModifiers{modifier},
		CleanupFuncs: []func(){
			func() { cleanupCalled = true },
		},
	}

	cli, err := New(options)
	require.NoError(t, err)
	assert.NotNil(t, cli)

	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			value := cmd.Context().Value("test")
			assert.Equal(t, "value", value)
		},
	}
	cli.RegisterCommands([]*cobra.Command{testCmd})

	var buf bytes.Buffer
	cli.root.SetOut(&buf)
	cli.root.SetErr(&buf)
	cli.root.SetArgs([]string{"test"})

	err = cli.Execute()
	assert.NoError(t, err)
	assert.True(t, modifierCalled)

	cli.Cleanup()
	assert.True(t, cleanupCalled)
}

func TestRegisterCommands(t *testing.T) {
	options := RootCommandOptions{
		Name:        "testcli",
		Description: "A test CLI application",
		Version:     "1.0.0",
	}

	cli, err := New(options)
	require.NoError(t, err)

	assert.Empty(t, cli.root.Commands())

	cmd1 := &cobra.Command{Use: "command1"}
	cmd2 := &cobra.Command{Use: "command2"}
	cli.RegisterCommands([]*cobra.Command{cmd1, cmd2})

	commands := cli.root.Commands()
	assert.Len(t, commands, 2)

	commandNames := make([]string, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Use
	}
	assert.Contains(t, commandNames, "command1")
	assert.Contains(t, commandNames, "command2")
}

func TestVerbosityLevels(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedLevel logrus.Level
	}{
		{
			name:          "no verbosity flag",
			args:          []string{"test"},
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "single verbose flag",
			args:          []string{"-v", "test"},
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "double verbose flag",
			args:          []string{"-vv", "test"},
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "triple verbose flag",
			args:          []string{"-vvv", "test"},
			expectedLevel: logrus.TraceLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanupCalled bool

			options := RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "1.0.0",
				CleanupFuncs: []func(){
					func() { cleanupCalled = true },
				},
			}

			cli, err := New(options)
			require.NoError(t, err)

			var actualLevel logrus.Level
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {
					logger := logging.FromContext(cmd.Context())
					actualLevel = logger.GetLevel()
				},
			}
			cli.RegisterCommands([]*cobra.Command{testCmd})

			var buf bytes.Buffer
			cli.root.SetOut(&buf)
			cli.root.SetErr(&buf)
			cli.root.SetArgs(tt.args)

			err = cli.Execute()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLevel, actualLevel)

			cli.Cleanup()
			assert.True(t, cleanupCalled)
		})
	}
}

func TestExecute_HelpFlag(t *testing.T) {
	options := RootCommandOptions{
		Name:        "testcli",
		Description: "A test CLI application",
		Version:     "1.0.0",
	}

	cli, err := New(options)
	require.NoError(t, err)

	var buf bytes.Buffer
	cli.root.SetOut(&buf)
	cli.root.SetErr(&buf)
	cli.root.SetArgs([]string{"--help"})

	err = cli.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "A test CLI application")
}

func TestExecute_VersionFlag(t *testing.T) {
	options := RootCommandOptions{
		Name:        "testcli",
		Description: "A test CLI application",
		Version:     "1.0.0",
	}

	cli, err := New(options)
	require.NoError(t, err)

	var buf bytes.Buffer
	cli.root.SetOut(&buf)
	cli.root.SetErr(&buf)
	cli.root.SetArgs([]string{"--version"})

	err = cli.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "1.0.0")
}

func TestContextModifiers(t *testing.T) {
	modifier1Called := false
	modifier2Called := false
	cleanupCalled := false

	modifier1 := func(ctx context.Context) context.Context {
		modifier1Called = true
		return context.WithValue(ctx, "key1", "value1")
	}

	modifier2 := func(ctx context.Context) context.Context {
		modifier2Called = true
		return context.WithValue(ctx, "key2", "value2")
	}

	options := RootCommandOptions{
		Name:        "testcli",
		Description: "A test CLI application",
		Version:     "1.0.0",
		Modifiers:   []ContextModifiers{modifier1, modifier2},
		CleanupFuncs: []func(){
			func() { cleanupCalled = true },
		},
	}

	cli, err := New(options)
	require.NoError(t, err)

	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			assert.Equal(t, "value1", ctx.Value("key1"))
			assert.Equal(t, "value2", ctx.Value("key2"))
		},
	}
	cli.RegisterCommands([]*cobra.Command{testCmd})

	var buf bytes.Buffer
	cli.root.SetOut(&buf)
	cli.root.SetErr(&buf)
	cli.root.SetArgs([]string{"test"})

	err = cli.Execute()
	require.NoError(t, err)
	assert.True(t, modifier1Called)
	assert.True(t, modifier2Called)

	cli.Cleanup()
	assert.True(t, cleanupCalled)
}

func TestCleanup_WithoutExecute(t *testing.T) {
	var cleanupCalled bool
	options := RootCommandOptions{
		Name:    "testcli",
		Version: "1.0.0",
		CleanupFuncs: []func(){
			func() { cleanupCalled = true },
		},
	}

	cli, err := New(options)
	require.NoError(t, err)

	assert.NotPanics(t, func() { cli.Cleanup() })
	assert.True(t, cleanupCalled)
}

func TestCleanup_Order(t *testing.T) {
	var order []string

	options := RootCommandOptions{
		Name:    "testcli",
		Version: "1.0.0",
		CleanupFuncs: []func(){
			func() { order = append(order, "user1") },
			func() { order = append(order, "user2") },
		},
	}

	cli, err := New(options)
	require.NoError(t, err)

	original := cli.cleanups[0]
	cli.cleanups[0] = func() {
		order = append(order, "internal")
		original()
	}

	cli.Cleanup()
	assert.Equal(t, []string{"internal", "user1", "user2"}, order)
}
