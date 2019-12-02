package data

type PlotDrive struct {
	directory string
	plotFiles []*PlotFile
}

func NewPlotDrive(directory string, plotFilePaths []string) *PlotDrive {
	pd := new(PlotDrive)
	pd.directory = directory
	pd.plotFiles = []*PlotFile{}
	for _, path := range plotFilePaths {
		if pf := NewPlotFile(path); pf != nil {
			pd.plotFiles = append(pd.plotFiles, pf)
		}
	}
	return pd
}

func (pd *PlotDrive) GetSize() uint64 {
	size := uint64(0)
	for _, plotFile := range pd.plotFiles {
		size += plotFile.GetSize()
	}
	return size
}

func (pd *PlotDrive) GetDirectory() string {
	return pd.directory
}

func (pd *PlotDrive) GetPlotFiles() []*PlotFile {
	return pd.plotFiles
}

func (pd *PlotDrive) CollectStartNonceMap() map[uint64]uint64 {
	startNonceMap := make(map[uint64]uint64)
	for _, pf := range pd.plotFiles {
		start := pf.GetStartNonce()
		size := pf.GetSize()
		startNonceMap[start] = size
	}
	return startNonceMap
}
