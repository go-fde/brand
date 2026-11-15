// Package kitcheck holds the tests that keep the brand kit whole: every file
// the pipeline (logos/brandgen) writes for the organisation and each of its
// projects is present and non-empty, the SVGs parse as XML, and the PNGs and
// JPEGs decode at the size their directory names.
package kitcheck

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

var names = []string{
	"go-fde",
	"go-fde-bitlocker",
	"go-fde-filevault",
	"go-fde-luks",
	"go-fde-luks2",
	"go-fde-veracrypt",
}

var (
	sizes    = []int{16, 32, 48, 64, 128, 256, 512, 1024}
	variants = []string{"color", "white", "black"}
)

func root() string { return filepath.Join("..", "..") }

func read(t *testing.T, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root(), rel))
	if err != nil {
		t.Errorf("%s: %v", rel, err)
		return nil
	}
	if len(b) == 0 {
		t.Errorf("%s is empty", rel)
	}
	return b
}

func TestSVGsParse(t *testing.T) {
	for _, name := range names {
		for _, v := range variants {
			rel := filepath.Join("svg", v, name+".svg")
			b := read(t, rel)
			if b == nil {
				continue
			}
			d := xml.NewDecoder(bytes.NewReader(b))
			sawSVG := false
			for {
				tok, err := d.Token()
				if err != nil {
					if !errors.Is(err, io.EOF) {
						t.Errorf("%s: %v", rel, err)
					}
					break
				}
				if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "svg" {
					sawSVG = true
				}
			}
			if !sawSVG {
				t.Errorf("%s has no <svg> element", rel)
			}
		}
	}
}

func TestRastersDecodeAtTheirSize(t *testing.T) {
	for _, name := range names {
		for _, s := range sizes {
			for _, v := range variants {
				checkSize(t, filepath.Join("png", v, fmt.Sprint(s), name+".png"), s)
			}
			checkSize(t, filepath.Join("jpg", fmt.Sprint(s), name+".jpg"), s)
		}
	}
}

func TestOtherFormatsPresent(t *testing.T) {
	for _, name := range names {
		for _, rel := range []string{
			filepath.Join("avatar", name+".png"),
			filepath.Join("ico", name+".ico"),
			filepath.Join("icns", name+".icns"),
			filepath.Join("social", name+".png"),
		} {
			read(t, rel)
		}
	}
}

func checkSize(t *testing.T, rel string, want int) {
	t.Helper()
	f, err := os.Open(filepath.Join(root(), rel))
	if err != nil {
		t.Errorf("%s: %v", rel, err)
		return
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Errorf("%s: %v", rel, err)
		return
	}
	if cfg.Width != want || cfg.Height != want {
		t.Errorf("%s is %dx%d, want %dx%d", rel, cfg.Width, cfg.Height, want, want)
	}
}
