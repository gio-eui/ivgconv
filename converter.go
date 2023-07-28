package ivgconv

import (
	"encoding/xml"
	"os"
)

// ConverterOptions contains options for the SVG to IconVG converter.
type ConverterOptions struct {
	// OutputSize is the size of the IconVG output image.
	OutputSize float32
	// ExcludePath is a list of paths to exclude from the IconVG image.
	ExcludePath []Path
}

// Option is a function that configures a ConverterOptions.
type Option func(*ConverterOptions)

// FromFile encodes the SVG file as IconVG.
func FromFile(filepath string, options ...Option) ([]byte, error) {
	// Read the SVG file.
	svgData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	// Encode the SVG file content as IconVG.
	return FromContent(svgData, options...)
}

// FromContent encodes the SVG file content as IconVG.
func FromContent(content []byte, options ...Option) ([]byte, error) {
	// Set the default converter options.
	opts := ConverterOptions{
		OutputSize: 48,
		ExcludePath: []Path{
			// Matches <path d="M0 0h24v24H0z" fill="none"/>
			{D: "M0 0h24v24H0z", Fill: "none"},
			// Matches <path d="M0 0H24V24H0z" fill="none"/>
			{D: "M0 0H24V24H0z", Fill: "none"},
		},
	}
	// Set the converter options.
	for _, option := range options {
		option(&opts)
	}
	// Parse the SVG file.
	var svg SVG
	if err := xml.Unmarshal(content, &svg); err != nil {
		return nil, err
	}
	// Check if the SVG file is valid.
	if err := svg.Validate(); err != nil {
		return nil, err
	}
	// Encode the SVG file as IconVG.
	return parseSVG(svg, opts)
}

// WithOutputSize sets the size of the IconVG image.
func WithOutputSize(outputSize float32) Option {
	return func(opts *ConverterOptions) {
		opts.OutputSize = outputSize
	}
}

// WithExcludePath sets the list of paths to exclude from the IconVG image.
func WithExcludePath(excludePath []Path) Option {
	return func(opts *ConverterOptions) {
		opts.ExcludePath = excludePath
	}
}
