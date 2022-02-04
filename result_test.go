package do_test

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/topi-team/do"
)

func ExampleWithReturn() {
	f := do.WithReturn(os.Open("missing_file"))
	limit := do.Map(f, func(f *os.File) io.Reader { return io.LimitReader(f, 100) })
	lines := do.MapOrErr(limit, func(r io.Reader) ([]string, error) {
		scanner := bufio.NewScanner(r)
		var l []string
		for scanner.Scan() {
			l = append(l, scanner.Text())
		}
		return l, scanner.Err()
	})
	do.Fold(
		lines,
		func(l []string) {
			fmt.Println(len(l))
		},
		func(err error) {
			fmt.Println(err)
		},
	)
	// Output:
	// open missing_file: no such file or directory
}

func ExampleReturn() {
	fileLines := func(file string) ([]string, error) {
		f := do.WithReturn(os.Open(file))
		limit := do.Map(f, func(r *os.File) io.Reader { return io.LimitReader(r, 10000) })
		lines := do.MapOrErr(limit, func(r io.Reader) ([]string, error) {
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
		res := do.WithReturn(fileLines(file))
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
	printResult("result_test.go")
	// Output:
	// error: open missing: no such file or directory
	// value: package do_test
}
