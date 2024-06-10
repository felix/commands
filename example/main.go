package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/felix/commands"
)

func main() {
	bah.RegisterChild(printer)
	foo.RegisterChild(bah)
	commands.RegisterChild(foo)
	commands.Short("This is the root short description")

	if i, err := commands.Run(); err != nil {
		os.Exit(i)
	}
}

var (
	capitalize bool
)

var printer = &commands.Command{
	Name:  "print",
	Short: "Print args to stdout",
	Usage: "[-capitalize] <some text>",
	Long:  "This is the long text",
	Flags: func() *flag.FlagSet {
		fs := flag.NewFlagSet("print", flag.ExitOnError)
		fs.BoolVar(&capitalize, "capitalize", false, "capitalize output")
		return fs
	}(),
	Func: func(_ context.Context, stdout io.Writer, fs *flag.FlagSet, stdin io.Reader, stderr io.Writer) (int, error) {
		// if verbose := ctx.Value(verboseKey); verbose != nil {
		// 	fmt.Println("Executing print")
		// }
		for _, arg := range fs.Args() {
			if capitalize {
				arg = strings.ToUpper(arg)
			}
			fmt.Fprintf(stdout, "%s ", arg)
		}
		fmt.Println()
		return 0, nil
	},
}

var foo = &commands.Command{
	Name:  "foo",
	Short: "Nothing",
}
var bah = &commands.Command{
	Name:  "bah",
	Short: "Nothing else",
}
