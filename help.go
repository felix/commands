package commands

import (
	"flag"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func getNames(cmd *Command) []string {
	if cmd.parent == nil {
		return []string{cmd.Name}
	}
	return append(getNames(cmd.parent), cmd.Name)
}

func showHelp(cmd *Command) {
	names := strings.Join(getNames(cmd), " ")
	result := fmt.Sprintf("%s\n\nUsage:\n  %s", cmd.Short, names)

	if cmd.Usage != "" {
		result += fmt.Sprintf(" %s", cmd.Usage)
	}

	keys := make([]string, 0, len(cmd.children))
	for k := range cmd.children {
		keys = append(keys, k)
	}
	if len(keys) > 0 {
		result += " <command>"
	}
	result += "\n"
	//result += " [<args...>]\n"

	if help := flagHelp(cmd.Flags); help != "" {
		result += fmt.Sprintf("\nFlags:\n%s", help)
	}

	if len(keys) > 0 {
		result += "\nCommands:\n"
		format := fmt.Sprintf("  %%-%ds %%s\n", cmd.longestName)
		sort.Strings(keys)
		for _, k := range keys {
			child := cmd.children[k]
			short := strings.TrimSuffix(child.Short, ".")
			result += fmt.Sprintf(format, child.Name, short)
		}
		result += fmt.Sprintf("\nUse '%s help <command>' for more information about a command.\n", names)
	}

	helpText := strings.TrimSpace(cmd.Long)
	if helpText != "" {
		result += fmt.Sprintf("\n%s\n", helpText)
	}

	fmt.Fprint(cmd.stdout, result)
}

func helpForCommand(cmd *Command, fl *flag.FlagSet) (int, error) {
	args := fl.Args()

	// Drop help if present
	if len(args) > 0 && args[0] == "help" {
		args = args[1:]
	}

	if len(args) > 2 {
		showHelp(cmd)
		return ExitCodeSerious, fmt.Errorf("can only give help with one command")
	}

	if len(args) < 1 {
		// Called when nothing else matched
		showHelp(cmd)
		return ExitCodeSuccess, nil
	}

	subcommand, ok := cmd.children[args[0]]
	if !ok {
		showHelp(cmd)
		return ExitCodeSerious, fmt.Errorf("unknown command: %s", args[0])
	}
	showHelp(subcommand)

	return ExitCodeSuccess, nil
}

func flagHelp(f *flag.FlagSet) string {
	if f == nil {
		return ""
	}
	var b strings.Builder
	f.VisitAll(func(fl *flag.Flag) {
		fmt.Fprintf(&b, "  -%s", fl.Name) // Two spaces before -; see next two comments.
		name, usage := flag.UnquoteUsage(fl)
		if len(name) > 0 {
			b.WriteString(" ")
			b.WriteString(name)
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if b.Len() <= 4 { // space, space, '-', 'x'.
			b.WriteString("\t")
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			b.WriteString("    \t")
		}
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

		if !isZeroValue(fl, fl.DefValue) {
			if _, ok := fl.Value.(*stringValue); ok {
				// put quotes on the value
				fmt.Fprintf(&b, " (default %q)", fl.DefValue)
			} else {
				fmt.Fprintf(&b, " (default %v)", fl.DefValue)
			}
		}
		b.WriteByte('\n')
	})
	return b.String()
}

type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) String() string { return string(*s) }

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(fl *flag.Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(fl.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	return value == z.Interface().(flag.Value).String()
}
