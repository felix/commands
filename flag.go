package commands

import (
	"flag"
	"strconv"
)

// String returns the string representation of the
// flag given by name. It panics if the flag is not
// in the flag set.
func FlagString(f *flag.FlagSet, name string) string {
	return f.Lookup(name).Value.String()
}

// Bool returns the boolean representation of the
// flag given by name. It returns false if the flag
// is not a boolean type. It panics if the flag is
// not in the flag set.
func FlagBool(f *flag.FlagSet, name string) bool {
	val, _ := strconv.ParseBool(FlagString(f, name))
	return val
}

// Int returns the integer representation of the
// flag given by name. It returns 0 if the flag
// is not an integer type. It panics if the flag is
// not in the flag set.
func FlagInt(f *flag.FlagSet, name string) int {
	val, _ := strconv.ParseInt(FlagString(f, name), 0, strconv.IntSize)
	return int(val)
}

// Float64 returns the float64 representation of the
// flag given by name. It returns 0 if the flag
// is not a float64 type. It panics if the flag is
// not in the flag set.
func FlagFloat64(f *flag.FlagSet, name string) float64 {
	val, _ := strconv.ParseFloat(FlagString(f, name), 64)
	return val
}
