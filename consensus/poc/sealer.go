package poc

import (
	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/consensus"
	data "github.com/pocethereum/pochain/consensus/poc/data"
	"github.com/pocethereum/pochain/core/types"
	"github.com/pocethereum/pochain/log"
	plotparams "github.com/pocethereum/pochain/params/plot"
	"math/big"
	"os"
	"strings"
	"time"
)

type MineResult struct {
	err      error
	nonce    uint64
	deadline *big.Int
}

func (poc *Poc) Seal(chain consensus.ChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error) {
	parentHeader := chain.GetHeader(block.ParentHash(), block.NumberU64()-1)
	if parentHeader == nil {
		return block, consensus.ErrUnknownAncestor
	}

	abort := make(chan struct{})
	found := make(chan *MineResult)
	ready := make(chan struct{})

	var result *MineResult
	go poc.mine(block, abort, found)

L:
	for {
		select {
		case <-stop:
			close(abort)
			return nil, errPocSearchAborted
		case result = <-found:
			if result.err != nil {
				return nil, result.err
			}

			//JUST for debug
			if common.ISBINGDEBUG {
				log.Warn("=== JUST for debug ==", "deadline", result.deadline)
			} else {
				baseTarget := plotparams.DifficultyToBaseTarget(block.Difficulty())
				result.deadline.Div(result.deadline, baseTarget)
			}

			waitSeconds := big.NewInt(time.Now().Unix())
			waitSeconds.Sub(waitSeconds, parentHeader.Time)
			waitSeconds.Sub(result.deadline, waitSeconds)

			if waitSeconds.Sign() < 0 {
				close(ready)
			} else {
				log.Info("Waiting time to elapse", "seconds", waitSeconds, "deadline", result.deadline)
				go func() {
					time.Sleep(time.Duration(waitSeconds.Int64()) * time.Second)
					close(ready)
				}()
			}
		case <-ready:
			copyHeader := block.Header()
			copyHeader.Nonce = types.EncodeNonce(result.nonce)
			newTime := new(big.Int).Add(parentHeader.Time, result.deadline)
			if copyHeader.Time.Cmp(newTime) < 0 {
				copyHeader.Time.Set(newTime)
			}
			block = block.WithSeal(copyHeader)
			break L
		}
	}
	return block, nil
}

func (poc *Poc) mine(block *types.Block, abort chan struct{}, found chan *MineResult) {
	genSigBytes := block.GetGenerationSignature().Bytes()
	scoopNumber := CalcScoop(genSigBytes, block.NumberU64())

	conf := poc.Config()
	plotPaths := strings.Split(conf.PlotPaths, ",")

	seed := strings.ToLower(block.Coinbase().Hex()[2:])
	if poc.plots == nil || poc.plots.Seed != seed ||
		strings.Join(poc.plots.PlotPaths, ",") != strings.Join(plotPaths, ",") {
		poc.plots = data.NewPlots(plotPaths, seed)
	}
	startNonceMap := poc.plots.GetStartNonceMap()

	result := &MineResult{
		err:      errPlotdataReadFailed,
		deadline: plotparams.MaximumDeadline(),
	}
	if len(startNonceMap) == 0 {
		log.Warn("Plotdata not found", "PlotPaths", conf.PlotPaths, "Seed", seed)
		result.err = errPlotdataNotFound
		found <- result
		return
	}

	log.Info("Start poc search for new nonces", "scoop", scoopNumber)

search:
	for startNonce, size := range startNonceMap {
		pf := poc.plots.GetPlotFileByStartNonce(startNonce)
		if pf == nil {
			log.Warn("Plotfile not found")
			continue
		}

		fd, err := os.Open(pf.GetFilePath())
		if err != nil {
			log.Warn("Plotfile open failed", "error", err)
			continue
		}

		partSize := size / plotparams.ScoopsPerPlot
		_, err2 := fd.Seek(int64(partSize*scoopNumber), os.SEEK_SET)
		if err2 != nil {
			log.Warn("Plotfile seek failed", "error", err2)
			fd.Close()
			continue
		}

		scoopCount := partSize / plotparams.ScoopSize
		scoopDataBytes := make([]byte, plotparams.ScoopSize)
		nonce := startNonce

	readLoop:
		for i := uint64(0); i < scoopCount; i, nonce = i+1, nonce+1 {
			select {
			case <-abort:
				log.Info("Poc search aborted")
				fd.Close()
				break search
			default:
				_, err3 := fd.Read(scoopDataBytes)
				if err3 != nil {
					log.Warn("Plotfile read failed", "error", err3)
					fd.Close()
					break readLoop
				}

				deadline := CalcHit(scoopDataBytes, genSigBytes)
				if deadline.Cmp(result.deadline) < 0 {
					result.err = nil
					result.nonce = nonce
					result.deadline.Set(deadline)
				}
			}
		}
		fd.Close()
	}

	found <- result
}
