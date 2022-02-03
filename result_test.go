package do_test

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/topi-team/do"
)

func ExampleResult() {
	fileLines := func(file string) ([]string, error) {
		f := do.Check(os.Open(file))
		limit := do.Map(f, func(r *os.File) io.Reader { return io.LimitReader(r, 10000) })
		lines := do.MapErr(limit, func(r io.Reader) ([]string, error) {
			scanner := bufio.NewScanner(r)
			var l []string
			for scanner.Scan() {
				l = append(l, scanner.Text())
			}
			return l, scanner.Err()
		})

		return lines.Return()
	}

	printResult := func(file string) {
		res := do.Check(fileLines(file))
		do.Fold(
			res,
			func(lines []string) {
				fmt.Printf("value: %s\n", lines[0])
			},
			func(err error) {
				fmt.Printf("error: %s\n", err)
			},
		)
	}

	printResult("missing")
	printResult("result.go")
	// Output:
	// error: open missing: no such file or directory
	// value: package do
}
