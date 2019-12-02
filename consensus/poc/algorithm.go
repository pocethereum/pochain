package poc

import (
	"crypto/sha256"
	"strings"

	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/common/hexutil"
	plotpoc "github.com/pocethereum/pochain/consensus/poc/plot"
	"github.com/pocethereum/pochain/core/types"
	plotparams "github.com/pocethereum/pochain/params/plot"
	"math/big"
)

func CalcGenerationSignature(lastGenSigBytes []byte, lastGeneratorBytes []byte) []byte {
	h := sha256.New()
	h.Write(lastGenSigBytes)
	h.Write(lastGeneratorBytes)
	return h.Sum(nil)
}

func CalcScoop(genSigBytes []byte, height uint64) uint64 {
	h := sha256.New()
	h.Write(genSigBytes)
	h.Write(common.Uint64ToBytes(height))
	digest := h.Sum(nil)
	hashnum := new(big.Int).SetBytes(digest)
	return hashnum.Mod(hashnum, new(big.Int).SetUint64(plotparams.ScoopsPerPlot)).Uint64()
}

func CalcHit(scoopDataBytes []byte, genSigBytes []byte) *big.Int {
	h := sha256.New()
	h.Write(scoopDataBytes)
	h.Write(genSigBytes)

	if common.ISBINGDEBUG {
		return big.NewInt(0).SetBytes(hexutil.MustDecode("0x84"))
	} else {
		digest := h.Sum(nil)
		return big.NewInt(0).SetBytes([]byte{digest[7], digest[6],
			digest[5], digest[4], digest[3], digest[2], digest[1], digest[0]})
	}

}

func CalcDeadline(scoopDataBytes []byte, genSigBytes []byte, baseTarget *big.Int) *big.Int {
	hit := CalcHit(scoopDataBytes, genSigBytes)
	return hit.Div(hit, baseTarget)
}

func CalcBlockPoc(header *types.Header) *types.BlockPoc {
	nonce := header.Nonce.Uint64()
	seed := strings.ToLower(header.Coinbase.Hex()[2:])
	mp := plotpoc.NewMiningPlot(seed, nonce)

	genSigBytes := header.GetGenerationSignature().Bytes()
	scoopNumber := CalcScoop(genSigBytes, header.Number.Uint64())
	scoopDataBytes := mp.GetScoop(scoopNumber)
	deadline := CalcHit(scoopDataBytes, genSigBytes)
	baseTarget := plotparams.DifficultyToBaseTarget(header.Difficulty)
	deadline.Div(deadline, baseTarget)

	return &types.BlockPoc{
		Nonce:       header.Nonce,
		ScoopNumber: scoopNumber,
		Deadline:    deadline,
		BaseTarget:  baseTarget,
	}
}
