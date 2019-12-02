package data

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pocethereum/pochain/log"
)

type Plots struct {
	plotDrives    []*PlotDrive
	startNonceMap map[uint64]uint64
	PlotPaths     []string
	Seed          string
}

func NewPlots(plotPaths []string, address string) *Plots {
	ps := new(Plots)
	ps.plotDrives = []*PlotDrive{}
	ps.startNonceMap = map[uint64]uint64{}
	ps.Seed = address
	ps.PlotPaths = plotPaths

	plotFilesLookup := collectPlotFiles(plotPaths, address)
	for plotDirectory, plotFilePaths := range plotFilesLookup {
		plotDrive := NewPlotDrive(plotDirectory, plotFilePaths)
		ps.plotDrives = append(ps.plotDrives, plotDrive)

		startNonceMap := plotDrive.CollectStartNonceMap()
		expectedSize := len(ps.startNonceMap) + len(startNonceMap)
		for startNonce, size := range startNonceMap {
			ps.startNonceMap[startNonce] = size
		}

		if len(ps.startNonceMap) != expectedSize {
			log.Warn("Possible duplicate/overlapping plotfile", "directory", plotDrive.GetDirectory())
		}
	}
	return ps
}

func (ps *Plots) GetPlotDrives() []*PlotDrive {
	return ps.plotDrives
}

func (ps *Plots) GetSize() uint64 {
	size := uint64(0)
	for _, plotDrive := range ps.plotDrives {
		size += plotDrive.GetSize()
	}
	return size
}

func (ps *Plots) GetStartNonceMap() map[uint64]uint64 {
	return ps.startNonceMap
}

func (ps *Plots) GetPlotFileByStartNonce(startNonce uint64) *PlotFile {
	for _, plotDrive := range ps.plotDrives {
		plotFiles := plotDrive.GetPlotFiles()
		for _, plotFile := range plotFiles {
			if plotFile.GetStartNonce() == startNonce {
				return plotFile
			}
		}
	}
	return nil
}

func (ps *Plots) PrintPlotFiles() {
	for _, plotDrive := range ps.plotDrives {
		plotDataDir := plotDrive.GetDirectory()
		fmt.Printf("PlotDataDir=%s\n", plotDataDir)
		for _, plotFile := range plotDrive.GetPlotFiles() {
			fmt.Printf(" #PlotFile=%s\n", plotFile.GetFileName())
		}
	}
}

func collectPlotFiles(plotDirectories []string, address string) map[string][]string {
	plotFilesLookup := make(map[string][]string)
	for _, plotDirectory := range plotDirectories {
		files, err := ioutil.ReadDir(plotDirectory)
		if err == nil {
			plotFilePaths := []string{}
			for _, file := range files {
				fileName := file.Name()
				if strings.HasPrefix(fileName, address) && filepath.Ext(fileName) == "" {
					plotFilePaths = append(plotFilePaths, filepath.Join(plotDirectory, fileName))
				}
			}
			plotFilesLookup[plotDirectory] = plotFilePaths
		}
	}
	return plotFilesLookup
}
