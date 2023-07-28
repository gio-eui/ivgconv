package ivgconv

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type SVG struct {
	Width   float32 `xml:"width,attr"`
	Height  float32 `xml:"height,attr"`
	ViewBox ViewBox `xml:"viewBox,attr"`
	Paths   []Path  `xml:"path"`
	// Some of the SVG files contain <circle> elements, not just <path>
	// elements. IconVG doesn't have circles per se. Instead, we convert such
	// circles to be paired arcTo commands, tacked on to the first path.
	//
	// In general, this isn't correct if the circles and the path overlap, but
	// that doesn't happen in the specific case of the Material Design icons.
	Circles []Circle `xml:"circle"`
}

func (s *SVG) Validate() error {
	// Update the viewbox if it is not set.
	if s.ViewBox.Width == 0 && s.ViewBox.Height == 0 {
		s.ViewBox.MinX = 0
		s.ViewBox.MinY = 0
		s.ViewBox.Width = s.Width
		s.ViewBox.Height = s.Height
	}
	// Check if the viewbox has a valid paths or circles.
	if len(s.Paths) == 0 && len(s.Circles) == 0 {
		return fmt.Errorf("no path or circle found in the SVG file")
	}
	// For each path, replace ',' with ' ' in the D attribute.
	for i := range s.Paths {
		s.Paths[i].D = strings.Replace(s.Paths[i].D, ",", " ", -1)
	}
	return nil
}

type ViewBox struct {
	MinX   float32 `xml:"min-x,attr"`
	MinY   float32 `xml:"min-y,attr"`
	Width  float32 `xml:"width,attr"`
	Height float32 `xml:"height,attr"`
}

type Path struct {
	D           string   `xml:"d,attr"`
	Fill        string   `xml:"fill,attr"`
	FillOpacity *float32 `xml:"fill-opacity,attr"`
	Opacity     *float32 `xml:"opacity,attr"`
}

type Circle struct {
	Cx float32 `xml:"cx,attr"`
	Cy float32 `xml:"cy,attr"`
	R  float32 `xml:"r,attr"`
}

// UnmarshalXMLAttr implements the xml.UnmarshalerAttr interface.
func (vb *ViewBox) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Name.Local != "viewBox" {
		return nil
	}
	if _, err := fmt.Sscanf(attr.Value, "%f %f %f %f", &vb.MinX, &vb.MinY, &vb.Width, &vb.Height); err != nil {
		return err
	}
	return nil
}
