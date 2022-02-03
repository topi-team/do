package is_test

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/topi-team/is"
)

func ExampleEither() {
	fileLines := func(file string) ([]string, error) {
		f := is.Check(os.Open(file))
		limit := is.Map(f, func(r *os.File) io.Reader { return io.LimitReader(r, 10000) })
		lines := is.MapErr(limit, func(r io.Reader) ([]string, error) {
			scanner := bufio.NewScanner(r)
			var l []string
			for scanner.Scan() {
				l = append(l, scanner.Text())
			}
			return l, scanner.Err()
		})

		return lines.Fold()
	}

	fmt.Printf("error: %s\n", is.Check(fileLines("missing")).Err())
	fmt.Printf("value: %s\n", is.Check(fileLines("either.go")).Val()[0])
	// Output:
	// error: open missing: no such file or directory
	// value: package is
}
