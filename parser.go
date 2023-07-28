package ivgconv

import (
	"fmt"
	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/image/math/f32"
	"io"
	"strings"
)

func parseSVG(svg SVG, opts ConverterOptions) ([]byte, error) {
	// Original svgSize.
	svgSize := svg.ViewBox.Width
	// Output svgSize.
	outSize := opts.OutputSize

	// iconVG encoder.
	var enc iconvg.Encoder
	enc.Reset(iconvg.Metadata{
		ViewBox: iconvg.Rectangle{
			//Min: f32.Vec2{-24, -24},
			//Max: f32.Vec2{+24, +24},
			Min: f32.Vec2{
				svg.ViewBox.MinX - svg.ViewBox.Width,
				svg.ViewBox.MinY - svg.ViewBox.Height,
			},
			Max: f32.Vec2{
				svg.ViewBox.Width,
				svg.ViewBox.Height,
			},
		},
		Palette: iconvg.DefaultPalette,
	})

	// Calculate the offset of the iconVG image.
	// The offset is the difference between the original
	// svgSize and the output svgSize.
	var vbx, vby float32
	vbx = svg.ViewBox.MinX
	vby = svg.ViewBox.MinY
	offset := f32.Vec2{
		vbx * outSize / svgSize,
		vby * outSize / svgSize,
	}

	// adjustment maps from opacity to a cReg adj value.
	adjs := map[float32]uint8{}

	// Generate the iconVG path.
	for _, p := range svg.Paths {
		// Skip the excluded paths.
		skip := false
		for _, ep := range opts.ExcludePath {
			if p.D == ep.D && p.Fill == ep.Fill {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		// Generate the iconVG path from the SVG path.
		if err := genPath(&enc, &p, adjs, svgSize, offset, outSize, svg.Circles); err != nil {
			return nil, err
		}
		svg.Circles = nil
	}

	// Generate the iconVG circle.
	if len(svg.Circles) != 0 {
		// Generate the iconVG path from the SVG circle.
		if err := genPath(&enc, &Path{}, adjs, svgSize, offset, outSize, svg.Circles); err != nil {
			return nil, err
		}
		svg.Circles = nil
	}

	// Return the iconVG data.
	return enc.Bytes()
}

// genPath generates the iconVG path.
func genPath(enc *iconvg.Encoder, p *Path, adjs map[float32]uint8, size float32, offset f32.Vec2, outSize float32, circles []Circle) error {
	// Adjustments.
	adj := uint8(0)
	// Opacity.
	opacity := float32(1)
	if p.Opacity != nil {
		opacity = *p.Opacity
	} else if p.FillOpacity != nil {
		opacity = *p.FillOpacity
	}
	// Blend the opacity.
	if opacity != 1 {
		var ok bool
		if adj, ok = adjs[opacity]; !ok {
			adj = uint8(len(adjs) + 1)
			adjs[opacity] = adj
			// Set CREG[0-adj] to be a blend of transparent (0x7f) and the
			// first custom palette color (0x80).
			enc.SetCReg(adj, false, iconvg.BlendColor(uint8(opacity*0xff), 0x7f, 0x80))
		}
	}

	// Generate the path.
	needStartPath := true
	if p.D != "" {
		needStartPath = false
		if err := genPathData(enc, adj, p.D, size, offset, outSize); err != nil {
			return err
		}
	}

	// Generate the circle.
	for _, c := range circles {
		// Normalize.
		cx := c.Cx * outSize / size
		cx -= outSize/2 + offset[0]
		cy := c.Cy * outSize / size
		cy -= outSize/2 + offset[1]
		r := c.R * outSize / size

		if needStartPath {
			needStartPath = false
			enc.StartPath(adj, cx-r, cy)
		} else {
			enc.ClosePathAbsMoveTo(cx-r, cy)
		}

		// Convert a circle to two relative arcTo ops, each of 180 degrees.
		// We can't use one 360 degree arcTo as the start and end point
		// would be coincident and the computation is degenerate.
		enc.RelArcTo(r, r, 0, false, true, +2*r, 0)
		enc.RelArcTo(r, r, 0, false, true, -2*r, 0)
	}

	enc.ClosePathEndPath()
	return nil
}

// genPathData generates the iconVG path data.
func genPathData(enc *iconvg.Encoder, adj uint8, pathData string, size float32, offset f32.Vec2, outSize float32) error {
	pathData = strings.TrimSuffix(pathData, "z")
	r := strings.NewReader(pathData)

	var args [7]float32
	op, relative, started := byte(0), false, false
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch {
		case b == ' ':
			continue
		case 'A' <= b && b <= 'Z':
			op, relative = b, false
		case 'a' <= b && b <= 'z':
			op, relative = b, true
		default:
			if err := r.UnreadByte(); err != nil {
				return err
			}
		}

		n := 0
		switch op {
		case 'Z', 'z':
			n = 0
		case 'H', 'h', 'V', 'v':
			n = 1
		case 'L', 'l', 'M', 'm', 'T', 't':
			n = 2
		case 'Q', 'q', 'S', 's':
			n = 4
		case 'C', 'c':
			n = 6
		case 'A', 'a':
			n = 7
			return fmt.Errorf("arcTo not supported")
		default:
			return fmt.Errorf("unknown opcode %c", b)
		}

		if err := scan(&args, r, n); err != nil {
			return err
		}
		normalize(&args, n, op, size, offset, outSize, relative)

		switch op {
		case 'H':
			enc.AbsHLineTo(args[0])
		case 'h':
			enc.RelHLineTo(args[0])
		case 'V':
			enc.AbsVLineTo(args[0])
		case 'v':
			enc.RelVLineTo(args[0])
		case 'L':
			enc.AbsLineTo(args[0], args[1])
		case 'l':
			enc.RelLineTo(args[0], args[1])
		case 'M':
			if !started {
				started = true
				enc.StartPath(adj, args[0], args[1])
			} else {
				enc.ClosePathAbsMoveTo(args[0], args[1])
			}
		case 'm':
			enc.ClosePathRelMoveTo(args[0], args[1])
		case 'T':
			enc.AbsSmoothQuadTo(args[0], args[1])
		case 't':
			enc.RelSmoothQuadTo(args[0], args[1])
		case 'Q':
			enc.AbsQuadTo(args[0], args[1], args[2], args[3])
		case 'q':
			enc.RelQuadTo(args[0], args[1], args[2], args[3])
		case 'S':
			enc.AbsSmoothCubeTo(args[0], args[1], args[2], args[3])
		case 's':
			enc.RelSmoothCubeTo(args[0], args[1], args[2], args[3])
		case 'C':
			enc.AbsCubeTo(args[0], args[1], args[2], args[3], args[4], args[5])
		case 'c':
			enc.RelCubeTo(args[0], args[1], args[2], args[3], args[4], args[5])
			//case 'A':
			//	enc.AbsArcTo(args[0], args[1], args[2], args[3] != 0, args[4] != 0, args[5], args[6])
			//case 'a':
			//	enc.RelArcTo(args[0], args[1], args[2], args[3] != 0, args[4] != 0, args[5], args[6])
		}
	}
	return nil
}

func scan(args *[7]float32, r *strings.Reader, n int) error {
	// Read the n arguments.
	for i := 0; i < n; i++ {
		for {
			if b, _ := r.ReadByte(); b != ' ' && b != ',' {
				if err := r.UnreadByte(); err != nil {
					return err
				}
				break
			}
		}
		if _, err := fmt.Fscanf(r, "%f", &args[i]); err != nil {
			return err
		}
	}
	// Clear the remaining arguments.
	for i := n; i < 7; i++ {
		args[i] = 0
	}
	return nil
}

func normalize(args *[7]float32, n int, op byte, size float32, offset f32.Vec2, outSize float32, relative bool) {
	for i := 0; i < n; i++ {
		args[i] *= outSize / size
		if relative {
			continue
		}
		args[i] -= outSize / 2
		switch {
		case n != 1:
			args[i] -= offset[i&0x01]
		case op == 'H':
			args[i] -= offset[0]
		case op == 'V':
			args[i] -= offset[1]
		}
	}
}
