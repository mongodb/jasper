package generator

import "io"

type Generator interface {
	// Generate generates an evergreen configuration and writes it to the
	// output.
	Generate(output io.Writer) error
}
