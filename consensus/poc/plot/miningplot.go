package plot

import (
	"crypto/sha256"

	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/common/bitutil"
	plotparams "github.com/pocethereum/pochain/params/plot"
)

type MiningPlot struct {
	address string
	nonce   uint64
	data    [plotparams.PlotSize]byte
}

func NewMiningPlot(address string, nonce uint64) *MiningPlot {
	mp := new(MiningPlot)
	mp.address = address
	mp.nonce = nonce

	addressBytes := common.FromHex(address)
	nonceBytes := common.Uint64ToBytes(nonce)

	h := sha256.New()
	for i, hstart := uint64(0), plotparams.PlotSize; i < plotparams.HashsPerPlot; i, hstart = i+1, hstart-plotparams.HashSize {
		h.Reset()
		if i < plotparams.HashPreImageHashLimit {
			if i != 0 {
				h.Write(mp.data[hstart:])
			}
			h.Write(addressBytes)
			h.Write(nonceBytes)
		} else {
			hend := hstart + plotparams.HashPreImageByteLimit
			h.Write(mp.data[hstart:hend])
		}
		h.Sum(mp.data[hstart-plotparams.HashSize : hstart])
	}

	h.Reset()
	h.Write(mp.data[0:])
	h.Write(addressBytes)
	h.Write(nonceBytes)
	finalHash := h.Sum(nil)

	for i := uint64(0); i < plotparams.PlotSize; i += plotparams.HashSize {
		dest := mp.data[i : i+plotparams.HashSize]
		bitutil.XORBytes(dest, dest, finalHash)
	}

	for i, j := uint64(1), plotparams.HashsPerPlot-1; i < j; i, j = i+2, j-2 {
		istart, jstart := i*plotparams.HashSize, j*plotparams.HashSize
		for k := uint64(0); k < plotparams.HashSize; k++ {
			mp.data[istart+k], mp.data[jstart+k] = mp.data[jstart+k], mp.data[istart+k]
		}
	}
	return mp
}

func (mp *MiningPlot) GetScoop(pos uint64) []byte {
	pos = pos % plotparams.ScoopsPerPlot
	start := pos * plotparams.ScoopSize
	return common.CopyBytes(mp.data[start : start+plotparams.ScoopSize])
}

func (mp *MiningPlot) GetData() []byte {
	return common.CopyBytes(mp.data[0:])
}

func (mp *MiningPlot) Data() []byte {
	return mp.data[0:]
}
