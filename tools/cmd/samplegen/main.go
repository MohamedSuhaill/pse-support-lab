// Command samplegen writes the customer sample files into a directory.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"invoicely/tools/internal/dataset"
	"invoicely/tools/internal/workbook"
)

func main() {
	out := flag.String("out", "samples", "output directory")
	flag.Parse()
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, s := range dataset.Samples {
		data, err := workbook.Build(s.Build(""))
		if err != nil {
			fmt.Fprintln(os.Stderr, s.Filename, err)
			os.Exit(1)
		}
		path := filepath.Join(*out, s.Filename)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s (%d bytes)\n", path, len(data))
	}
}
