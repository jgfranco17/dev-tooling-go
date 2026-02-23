package commandline

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CLI is a struct that represents the command-line interface of the application.
type CLI struct {
	root     *cobra.Command
	cleanups []func()
}

// ContextModifiers is a function type that takes a context and returns
// a modified context. This can be used to add additional values to the
// context for downstream consumption.
type ContextModifiers func(ctx context.Context) context.Context

// RootCommandOptions defines the options for creating a new CLI instance.
type RootCommandOptions struct {
	// Name is the name of the root command, i.e. the namespace used to invoke the CLI.
	Name string

	// Description is a brief description of the root command.
	// This will be displayed in the help message.
	Description string

	// Version is the version of the root command.
	// This will be displayed in the --version flag.
	Version string

	// Modifiers are functions that can modify the context before executing.
	// This can be used to add additional values to the context, such as a logger
	// or other dependencies.
	Modifiers []ContextModifiers

	// CleanupFuncs are functions that will be called when the CLI is cleaned up.
	// This can be used to clean up resources, such as closing database connections
	// or stopping background goroutines.
	CleanupFuncs []func()
}

// validate checks if the required fields in RootCommandOptions are set.
func (options RootCommandOptions) validate() error {
	if options.Name == "" {
		return fmt.Errorf("root command must have name")
	}
	if options.Version == "" {
		return fmt.Errorf("root command must have version")
	}
	return nil
}

// New creates a new instance of CLI with the provided options.
func New(options RootCommandOptions) (*CLI, error) {
	if err := options.validate(); err != nil {
		return nil, err
	}

	// Create a signal-aware context for the entire CLI lifecycle once here.
	// This ensures signal handling is never duplicated regardless of how many
	// commands are executed or how many times PersistentPreRunE fires.
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	var verbosity int
	root := &cobra.Command{
		Use:     options.Name,
		Version: options.Version,
		Short:   options.Description,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			verbosity, _ = cmd.Flags().GetCount("verbose")
			var level logrus.Level
			switch verbosity {
			case 1:
				level = logrus.InfoLevel
			case 2:
				level = logrus.DebugLevel
			case 3:
				level = logrus.TraceLevel
			default:
				level = logrus.WarnLevel
			}

			logger := logging.New(cmd.ErrOrStderr(), level)
			ctx := logging.AddToContext(cmd.Context(), logger)

			if options.Modifiers != nil {
				for _, modifier := range options.Modifiers {
					ctx = modifier(ctx)
				}
			}

			cmd.SetContext(ctx)
			return nil
		},
	}

	// Prime the root command with the signal-aware context so every derived
	// command context inherits cancellation on SIGTERM/SIGINT.
	root.SetContext(signalCtx)
	root.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity (up to -vvv)")

	// Internal cleanup (stop) is prepended so signal notifications are released
	// before user-provided cleanup functions run.
	allCleanups := make([]func(), 0, 1+len(options.CleanupFuncs))
	allCleanups = append(allCleanups, stop)
	allCleanups = append(allCleanups, options.CleanupFuncs...)

	return &CLI{
		root:     root,
		cleanups: allCleanups,
	}, nil
}

// RegisterCommands registers new commands with the CLI.
func (cr *CLI) RegisterCommands(commands []*cobra.Command) {
	cr.root.AddCommand(commands...)
}

// Cleanup releases resources held by the CLI.
func (cr *CLI) Cleanup() {
	for _, cleanup := range cr.cleanups {
		cleanup()
	}
}

// Execute executes the root command.
func (cr *CLI) Execute() error {
	return cr.root.Execute()
}
