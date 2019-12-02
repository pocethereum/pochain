package data

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/log"
	plotparams "github.com/pocethereum/pochain/params/plot"
)

type PlotFile struct {
	filePath   string
	fileName   string
	address    common.Address
	startNonce uint64
	plots      uint64
	size       uint64
}

func NewPlotFile(path string) *PlotFile {
	pf := new(PlotFile)
	pf.filePath = path
	pf.fileName = filepath.Base(path)

	parts := strings.Split(pf.fileName, "_")
	if len(parts) != 3 {
		log.Warn("Invalid fileName format", "plotfile", pf.filePath)
		return nil
	}

	pf.address = common.HexToAddress(parts[0])

	var err error
	pf.startNonce, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		log.Warn("Parse startNonce failed", "plotfile", pf.filePath)
		return nil
	}

	pf.plots, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		log.Warn("Parse plots failed", "plotfile", pf.filePath)
		return nil
	}
	pf.size = pf.plots * plotparams.PlotSize

	stat, err := os.Stat(pf.filePath)
	if err != nil {
		log.Warn("Stat failed", "plotfile", pf.filePath, "error", err)
		return nil
	}

	if int64(pf.size) != stat.Size() {
		log.Warn("File size mismatch", "expected", pf.size, "actual", stat.Size())
	}

	return pf
}

func (pf *PlotFile) GetFilePath() string {
	return pf.filePath
}

func (pf *PlotFile) GetFileName() string {
	return pf.fileName
}

func (pf *PlotFile) GetStartNonce() uint64 {
	return pf.startNonce
}

func (pf *PlotFile) GetSize() uint64 {
	return pf.size
}
