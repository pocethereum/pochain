package poc

//
import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"time"

	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/common/math"
	"github.com/pocethereum/pochain/consensus"
	"github.com/pocethereum/pochain/consensus/poc/data"
	"github.com/pocethereum/pochain/consensus/poc/mortgage"
	"github.com/pocethereum/pochain/core/state"
	"github.com/pocethereum/pochain/core/types"
	"github.com/pocethereum/pochain/params"
	plotparams "github.com/pocethereum/pochain/params/plot"
	"github.com/pocethereum/pochain/rpc"
	"github.com/pocethereum/pochain/log"
)

var (
	errZeroBlockTime      = errors.New("timestamp equals parent's")
	errInvalidDeadline    = errors.New("timestamp mismatch with deadline")
	errInvalidGenSig      = errors.New("invalid generation signature")
	errUnclesNotAllowed   = errors.New("uncles not allowed")
	errInvalidDifficulty  = errors.New("non-positive difficulty")
	errPlotdataNotFound   = errors.New("plotdata not found")
	errPlotdataReadFailed = errors.New("plotdata read failed")
	errPocSearchAborted   = errors.New("poc search aborted")
)

var (
	allowedFutureBlockTime = 15 * time.Second
)

type Poc struct {
	config *params.PocConfig
	plots  *data.Plots
}

func New(config *params.PocConfig) *Poc {
	conf := *config
	go func(){
		for{
			log.Info("PlotpathsUpdater waiting...")
			if plotpath, ok := <- conf.PlotpathsUpdater; ok &&  plotpath != ""{
				log.Info("PlotpathsUpdater success", "current", plotpath, "pre", conf.PlotPaths)
				conf.PlotPaths = plotpath
			} else {
				log.Info("PlotpathsUpdater error", "current plotpasth", conf.PlotPaths, "plotpath", plotpath)
				time.Sleep(5*time.Second)
			}
		}
	}()

	return &Poc{
		config: &conf,
	}
}

func (poc *Poc) Config() *params.PocConfig {
	return poc.config
}

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-capacity verified author of the block.
func (poc *Poc) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the poc engine.
func (poc *Poc) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}
	AncestorHeaders, err := poc.getAncestorHeaders(chain, header, plotparams.CalcDiffBlockLimit)
	if err != nil {
		return err
	}
	return poc.verifyHeader(chain, header, AncestorHeaders, false, seal)
}

func (poc *Poc) getAncestorHeaders(chain consensus.ChainReader, header *types.Header, count uint64) ([]*types.Header, error) {
	ancestorHeaders := []*types.Header{}
	number := header.Number.Uint64()
	ancestorHeader := header
	for i := uint64(0); i < count && number > 0; i++ {
		ancestorHeader = chain.GetHeader(ancestorHeader.ParentHash, number-1)
		if ancestorHeader == nil {
			return ancestorHeaders, consensus.ErrUnknownAncestor
		}
		ancestorHeaders = append(ancestorHeaders, ancestorHeader)
		number = ancestorHeader.Number.Uint64()
	}
	return ancestorHeaders, nil
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (poc *Poc) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errs   = make([]error, len(headers))
		abort  = make(chan struct{})
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errs[index] = poc.verifyHeaderWorker(chain, headers, seals, index)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errs[out]
					if out == len(headers)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}

func (poc *Poc) verifyHeaderWorker(chain consensus.ChainReader, headers []*types.Header, seals []bool, index int) error {
	currentIndex := index
	ancestorHeaders := []*types.Header{}

	header := headers[index]
	for i := uint64(0); i < plotparams.CalcDiffBlockLimit && index > 0; i++ {
		if headers[index-1].Hash() != header.ParentHash {
			return consensus.ErrUnknownAncestor
		}
		header = headers[index-1]
		ancestorHeaders = append(ancestorHeaders, header)
		index--
	}

	count := plotparams.CalcDiffBlockLimit - uint64(len(ancestorHeaders))
	newHeaders, err := poc.getAncestorHeaders(chain, header, count)
	if err != nil {
		return err
	}
	ancestorHeaders = append(ancestorHeaders, newHeaders...)

	if chain.GetHeader(headers[currentIndex].Hash(), headers[currentIndex].Number.Uint64()) != nil {
		return nil
	}
	return poc.verifyHeader(chain, headers[currentIndex], ancestorHeaders, false, seals[index])
}

// VerifyUncles verifies that the given block's uncles conform to the consensus rules of the poc engine.
func (poc *Poc) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errUnclesNotAllowed
	}
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the poc engine.
func (poc *Poc) verifyHeader(chain consensus.ChainReader, header *types.Header,
	ancestorHeaders []*types.Header, uncle bool, seal bool) error {
	if len(ancestorHeaders) == 0 {
		return consensus.ErrUnknownAncestor // parent header not found
	}
	parentHeader := ancestorHeaders[0]

	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}
	if header.Time.Cmp(big.NewInt(time.Now().Add(allowedFutureBlockTime).Unix())) > 0 {
		return consensus.ErrFutureBlock
	}
	if header.Time.Cmp(parentHeader.Time) <= 0 {
		return errZeroBlockTime
	}

	// Verify dificiculty
	if header.Difficulty.Sign() != 1 {
		return errInvalidDifficulty
	}

	expected := CalcDifficulty(header, ancestorHeaders)
	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}

	// Verify that the gas limit is <= 2^63-1
	maxcap := uint64(0x7fffffffffffffff)
	if header.GasLimit > maxcap {
		return fmt.Errorf("invalid gasLimit: have %d, max %d", header.GasLimit, math.MaxBig63)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parentHeader.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parentHeader.GasLimit / params.GasLimitBoundDivisor
	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parentHeader.GasLimit, limit)
	}

	// Verify the block number and the generation signature
	currentHeader := header
	childHeader := currentHeader
	for _, header := range ancestorHeaders {
		if diff := new(big.Int).Sub(childHeader.Number, header.Number); diff.Cmp(big.NewInt(1)) != 0 {
			return consensus.ErrInvalidNumber
		}

		genSigHash := header.GetGenerationSignature()
		calcGenSigBytes := CalcGenerationSignature(genSigHash[:], header.Coinbase[:])
		childGenSigHash := childHeader.GetGenerationSignature()
		if !bytes.Equal(calcGenSigBytes, childGenSigHash[:]) {
			return errInvalidGenSig
		}
		childHeader = header
	}

	if seal {
		if err := poc.verifySeal(chain, currentHeader, parentHeader); err != nil {
			return err
		}
	}
	return nil
}

// VerifySeal implements consensus.Engine, checking whether the given block satisfies
// the PoC difficulty requirements.
func (poc *Poc) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	return poc.verifySeal(chain, header, nil)
}

func (poc *Poc) verifySeal(chain consensus.ChainReader, header *types.Header, parentHeader *types.Header) error {
	if parentHeader != nil {
		blockPoc := CalcBlockPoc(header)
		intervalTime := new(big.Int).Sub(header.Time, parentHeader.Time)
		if intervalTime.Cmp(blockPoc.Deadline) < 0 {
			return errInvalidDeadline
		}
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the poc protocol. The changes are done inline.
func (poc *Poc) Prepare(chain consensus.ChainReader, header *types.Header) error {
	ancestorHeaders, err := poc.getAncestorHeaders(chain, header, plotparams.CalcDiffBlockLimit)
	if err != nil {
		return err
	}

	parentHeader := ancestorHeaders[0]
	parentGenSig := parentHeader.GetGenerationSignature()
	genSigBytes := CalcGenerationSignature(parentGenSig.Bytes(), parentHeader.Coinbase.Bytes())
	header.SetGenerationSignature(common.BytesToHash(genSigBytes))
	difficulty := CalcDifficulty(header, ancestorHeaders)
	header.Difficulty = new(big.Int).Set(difficulty)
	return nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state and assembling the block.
func (poc *Poc) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	reward := mortgage.CalcReward(header.Coinbase, header.Nonce.Uint64(), state)
	state.AddBalance(header.Coinbase, reward)
	mortgage.AddTotalRewarded(state, reward)
	header.Root = state.IntermediateRoot(true)
	return types.NewBlock(header, txs, uncles, receipts), nil
}

func (poc *Poc) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	// dummy
	return big.NewInt(1)
}
func CalcDifficulty(header *types.Header, ancestorHeaders []*types.Header) *big.Int {
	len := len(ancestorHeaders)
	if len < 5 {
		return plotparams.GenesisDifficulty
	} else if len < int(plotparams.CalcDiffBlockLimit) {
		avgBaseTarget := big.NewInt(0)
		for i := 0; i < 5; i++ {
			baseTarget := plotparams.DifficultyToBaseTarget(ancestorHeaders[i].Difficulty)
			avgBaseTarget.Add(avgBaseTarget, baseTarget)
		}

		avgBaseTarget.Div(avgBaseTarget, big.NewInt(5))
		difTime := new(big.Int).Sub(ancestorHeaders[0].Time, ancestorHeaders[4].Time)
		newBaseTarget := new(big.Int).Mul(avgBaseTarget, difTime)
		newBaseTarget = newBaseTarget.Div(newBaseTarget, big.NewInt(4*int64(plotparams.DurationLimit)))
		if newBaseTarget.Sign() <= 0 || newBaseTarget.Cmp(plotparams.MaximumBaseTarget) > 0 {
			newBaseTarget.Set(plotparams.MaximumBaseTarget)
		}

		delta := new(big.Int).Div(avgBaseTarget, big.NewInt(10))
		floorTarget := new(big.Int).Sub(avgBaseTarget, delta)
		ceilingTarget := new(big.Int).Add(avgBaseTarget, delta)
		if newBaseTarget.Cmp(floorTarget) < 0 {
			newBaseTarget.Set(floorTarget)
		} else if newBaseTarget.Cmp(ceilingTarget) > 0 {
			newBaseTarget.Set(ceilingTarget)
		}
		return plotparams.BaseTargetToDifficulty(newBaseTarget)
	} else {
		avgBaseTarget := big.NewInt(0)
		totalWeight := big.NewInt(0)
		for i := uint64(0); i < plotparams.CalcDiffBlockLimit; i++ {
			baseTarget := plotparams.DifficultyToBaseTarget(ancestorHeaders[i].Difficulty)
			weight := new(big.Int).SetUint64(4*plotparams.CalcDiffBlockLimit - i)
			baseTarget.Mul(baseTarget, weight)
			avgBaseTarget.Add(avgBaseTarget, baseTarget)
			totalWeight.Add(totalWeight, weight)
		}
		avgBaseTarget.Div(avgBaseTarget, totalWeight)
		posIndex := plotparams.CalcDiffBlockLimit - 1
		difTime := new(big.Int).Sub(ancestorHeaders[0].Time, ancestorHeaders[posIndex].Time)
		targetTimeSpan := new(big.Int).SetUint64(posIndex * plotparams.DurationLimit)

		floorDifTime := new(big.Int).Div(targetTimeSpan, big.NewInt(2))
		ceilingDifTime := new(big.Int).Mul(targetTimeSpan, big.NewInt(2))
		if difTime.Cmp(floorDifTime) < 0 {
			difTime.Set(floorDifTime)
		} else if difTime.Cmp(ceilingDifTime) > 0 {
			difTime.Set(ceilingDifTime)
		}

		curBaseTarget := plotparams.DifficultyToBaseTarget(ancestorHeaders[0].Difficulty)
		newBaseTarget := new(big.Int).Mul(avgBaseTarget, difTime)
		newBaseTarget.Div(newBaseTarget, targetTimeSpan)
		if newBaseTarget.Sign() <= 0 || newBaseTarget.Cmp(plotparams.MaximumBaseTarget) > 0 {
			newBaseTarget.Set(plotparams.MaximumBaseTarget)
		}

		delta := new(big.Int).Div(curBaseTarget, big.NewInt(10))
		delta.Mul(delta, big.NewInt(2))
		floorTarget := new(big.Int).Sub(curBaseTarget, delta)
		ceilingTarget := new(big.Int).Add(curBaseTarget, delta)
		if newBaseTarget.Cmp(floorTarget) < 0 {
			newBaseTarget.Set(floorTarget)
		} else if newBaseTarget.Cmp(ceilingTarget) > 0 {
			newBaseTarget.Set(ceilingTarget)
		}
		return plotparams.BaseTargetToDifficulty(newBaseTarget)
	}
}
func (poc *Poc) APIs(chain consensus.ChainReader) []rpc.API {
	return []rpc.API{}
}

func (poc *Poc) GetSize() uint64 {
	return poc.plots.GetSize()
}
