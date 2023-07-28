package ivgconv

import (
	"bytes"
	"github.com/gio-eui/ivgconv/testdata"

	"testing"
)

func TestFromFile(t *testing.T) {
	// Encode the SVG file as IconVG.
	ivgData, err := FromFile("testdata/close.svg")
	if err != nil {
		t.Fatal(err)
	}

	// Check that the IconVG data matches the expected output.
	if !bytes.Equal(ivgData, testdata.Close) {
		t.Fatalf("ivgData != Close")
	}

	// Encode the SVG file as IconVG.
	ivgData, err = FromFile("testdata/StarHalf.svg")
	if err != nil {
		t.Fatal(err)
	}

	// Check that the IconVG data matches the expected output.
	if !bytes.Equal(ivgData, testdata.StarHalf) {
		t.Fatalf("ivgData != StarHalf")
	}
}
