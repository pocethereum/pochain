package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlots(t *testing.T) {
	plotPath := filepath.Join(os.Getenv("HOME"), "plotdata")
	os.Mkdir(plotPath, os.ModePerm)
	address := "77b45e75cf93e428ae2ac6151666bac9fdbb1aa2"
	ps := NewPlots([]string{plotPath}, address)
	ps.PrintPlotFiles()
}
