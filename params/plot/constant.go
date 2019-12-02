package plot

import (
	"math"
	"math/big"

	"github.com/pocethereum/pochain/common"
)

const (
	MaxNonce      uint64 = math.MaxUint64
	HashSize      uint64 = 32
	HashsPerScoop uint64 = 2
	ScoopSize     uint64 = HashSize * HashsPerScoop
	ScoopsPerPlot uint64 = 4096
	PlotSize      uint64 = ScoopSize * ScoopsPerPlot
	HashsPerPlot  uint64 = HashsPerScoop * ScoopsPerPlot

	HashPreImageHashLimit uint64 = 128
	HashPreImageByteLimit uint64 = HashSize * HashPreImageHashLimit
	CalcDiffBlockLimit    uint64 = 25
	DurationLimit         uint64 = 180 //240
)

var (
	mathUint64        = new(big.Int).Exp(big.NewInt(2), big.NewInt(64), big.NewInt(0))
	GenesisBaseTarget = big.NewInt(18325193796000)
	//GenesisBaseTarget = big.NewInt(112589990684262400)
	MaximumBaseTarget = new(big.Int).Set(GenesisBaseTarget)

	GenesisDifficulty = BaseTargetToDifficulty(GenesisBaseTarget)
	MinimumDifficulty = BaseTargetToDifficulty(MaximumBaseTarget)
	MochizukiDifficulty =BaseTargetToDifficulty(big.NewInt(183251937960))
)

func DifficultyToBaseTarget(difficulty *big.Int) *big.Int {
	return new(big.Int).Div(mathUint64, difficulty)
}

func BaseTargetToDifficulty(baseTarget *big.Int) *big.Int {
	return new(big.Int).Div(mathUint64, baseTarget)
}

func MaximumDeadline() *big.Int {
	return new(big.Int).Set(mathUint64)
}

var (
	MainnetGenesisHash = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
	TestnetGenesisHash = common.HexToHash("0x41941023680923e0fe4d74a34bdac8141f2540e3ae90623718e47d66d1ca4a2d")
)
