package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
)

const (
	ExitCodeSuccess int = iota
	ExitCodeFailure
	ExitCodeSerious
)

// CommandFunc is a command's function. It runs the command and returns the
// proper exit code along with any error that occurred.
type CommandFunc func(
	ctx context.Context,
	stdout io.Writer,
	fs *flag.FlagSet,
	getenv func(string) string,
	stdin io.Reader,
	stderr io.Writer,
) (int, error)

// Command represents a subcommand. Name, Func, and Short are required.
type Command struct {
	// The name of the subcommand. Must conform to the format described by
	// the RegisterCommand() godoc.
	// Required.
	Name string

	// Func is a function that executes a subcommand using the parsed
	// flags. It returns an exit code and any associated error.
	// Required.
	Func CommandFunc

	// Usage is a brief message describing the syntax of the subcommand's
	// flags and args. Use [] to indicate optional parameters and <> to
	// enclose literal values intended to be replaced by the user. Only
	// include the actual parameters for this command.
	Usage string

	// Short is a one-line message explaining what the command does. Should
	// not end with punctuation.
	// Required.
	Short string

	// Long is the full help text shown to the user. Will be trimmed of
	// whitespace on both ends before being printed.
	Long string

	// Flags is the flagset for command.
	Flags *flag.FlagSet

	parent      *Command
	children    map[string]*Command
	longestName int

	// stdin defaults to os.stdin unless overridden
	stdin io.Reader
	// stdout defaults to os.stdout unless overridden
	stdout io.Writer
	// stderr defaults to os.stderr unless overridden
	stderr io.Writer
}

var root = &Command{}

func Short(s string) { root.Short = s }
func Long(s string)  { root.Long = s }

// Commands returns a list of commands initialised by RegisterCommand
// func Commands() map[string]Command {
// 	return commands
// }

func init() {
	root.Name = path.Base(os.Args[0])
	root.stdin = os.Stdin
	root.stdout = os.Stdout
	root.stderr = os.Stderr
}

func Stdin(r io.Reader) {
	root.stdin = r
}
func Stdout(w io.Writer) {
	root.stdout = w
}
func Stderr(w io.Writer) {
	root.stderr = w
}

// Run the command set.
// If no argument strings are passed then os.Args is used.
func Run(args ...string) (int, error) {
	if len(args) == 0 {
		args = os.Args
	}
	return root.Execute(args)
}

// args first item is the command name, program name for the root command.
func (cmd *Command) Execute(args []string) (int, error) {
	if len(args) == 0 {
		panic(fmt.Sprintf("invalid execute %q", cmd.Name))
	}

	if len(cmd.children) == 0 && cmd.Func == nil {
		panic(fmt.Sprintf("invalid command %q", cmd.Name))
	}

	if cmd.parent != nil {
		cmd.stdin = cmd.parent.stdin
		cmd.stdout = cmd.parent.stdout
		cmd.stderr = cmd.parent.stderr
	}

	var err error

	fs := cmd.Flags
	if fs == nil {
		fs = flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	}
	fs.Usage = func() {
		showHelp(cmd)
	}
	if err = fs.Parse(args[1:]); err != nil {
		return ExitCodeSerious, err
	}

	if len(args) == 1 {
		if cmd.Func != nil {
			return cmd.Func(context.TODO(), cmd.stdout, fs, os.Getenv, cmd.stdin, cmd.stderr)
		}
		return helpForCommand(cmd, fs)
	}

	if args[1] == "help" {
		return helpForCommand(cmd, fs)
	}

	subcommandName := args[1]
	subcommand, ok := cmd.children[subcommandName]
	if !ok {
		// if strings.HasPrefix(args[1], "-") {
		// 	// user probably forgot to type the subcommand
		// 	fmt.Fprintf(os.Stderr, "[ERROR] first argument must be a subcommand; see 'help'")
		// }
		// fmt.Fprintf(os.Stderr, "[ERROR] '%s' is not a recognized subcommand; see 'help'\n", args[1])
		// os.Exit(ExitCodeSerious)
		if cmd.Func != nil {
			return cmd.Func(context.TODO(), cmd.stdout, fs, os.Getenv, cmd.stdin, cmd.stderr)
		}
		return helpForCommand(cmd, fs)
	}
	return subcommand.Execute(args[1:])
}

// RegisterChild registers the command cmd. cmd.Name must be unique and
// conform to the following format:
//
//   - lowercase
//   - alphanumeric and hyphen characters only
//   - cannot start or end with a hyphen
//   - hyphen cannot be adjacent to another hyphen
//
// This function panics if the name is already registered, if the name does not
// meet the described format, or if any of the fields are missing from cmd.
//
// This function should be used in init().
func (c *Command) RegisterChild(cmd *Command) {
	if cmd.Name == "" {
		panic("command name is required")
	}
	// if cmd.Func == nil {
	// 	panic("command function missing")
	// }
	if cmd.Short == "" {
		panic("command short string is required")
	}
	if c.children == nil {
		c.children = make(map[string]*Command)
	}
	if _, exists := c.children[cmd.Name]; exists {
		panic("command already registered: " + cmd.Name)
	}
	if !commandNameRegex.MatchString(cmd.Name) {
		panic("invalid command name")
	}
	cmd.parent = c
	c.children[cmd.Name] = cmd
	if len(cmd.Name) > c.longestName {
		c.longestName = len(cmd.Name)
	}
}

// RegisterChild registers the command cmd. cmd.Name must be unique and
// conform to the following format:
//
//   - lowercase
//   - alphanumeric and hyphen characters only
//   - cannot start or end with a hyphen
//   - hyphen cannot be adjacent to another hyphen
//
// This function panics if the name is already registered, if the name does not
// meet the described format, or if any of the fields are missing from cmd.
//
// This function should be used in init().
func RegisterChild(cmd *Command) {
	root.RegisterChild(cmd)
}

var commandNameRegex = regexp.MustCompile(`^[a-z0-9]$|^([a-z0-9]+-?[a-z0-9]*)+[a-z0-9]$`)
