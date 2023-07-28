# ivgconv

[![GoDoc](https://godoc.org/github.com/gio-eui/ivgconv?status.svg)](https://godoc.org/github.com/gio-eui/ivgconv)
[![Go Report Card](https://goreportcard.com/badge/github.com/gio-eui/ivgconv)](https://goreportcard.com/report/github.com/gio-eui/ivgconv)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/go-mods/avatar/blob/master/LICENSE)

ivgconv implements conversion between SVG and [IconVG](https://github.com/golang/exp/tree/master/shiny/iconvg) format.


## Limitations

Only a subset of SVG is supported, and only a subset of IconVG is generated.

If the conversion fails, an error is returned.


## Credits

The IconVG converter is based on [gen.go](https://github.com/golang/exp/blob/master/shiny/materialdesign/icons/gen.go) which is licensed under the [APACHE LICENSE, VERSION 2.0](https://github.com/golang/exp/blob/master/shiny/materialdesign/icons/LICENSE)
