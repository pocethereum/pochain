package mortgage

import (
	"bytes"
	"github.com/pocethereum/pochain/common"
	"github.com/pocethereum/pochain/common/hexutil"
	"github.com/pocethereum/pochain/core/state"
	"github.com/pocethereum/pochain/crypto"
	"github.com/pocethereum/pochain/log"
	math2 "math"
	"math/big"
)

var (
	bigInt2e20  *big.Int = new(big.Int).Exp(big.NewInt(2), big.NewInt(20), nil)
	bigInt10e18 *big.Int = big.NewInt(1000000000000000000)
	bigInt10e20 *big.Int = new(big.Int).Mul(big.NewInt(100), bigInt10e18)

	MortgageSystemK                   = big.NewInt(10000)
	MortgageOneBlockFullReward        = new(big.Int).Mul(big.NewInt(168), bigInt10e18)
	MortgageSystemMaxReward           = new(big.Int).Mul(big.NewInt(210000000), bigInt10e18)
	MortgageContractAddr              = common.BytesToAddress([]byte{129})
	MortgageSystemCode         []byte = hexutil.MustDecode("0x" + "60606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063396ffa1b1461005f57806343794dda146100a6578063d8a830c6146100f5578063db006a7514610118575b610000565b3461000057610090600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190505061014d565b6040518082815260200191505060405180910390f35b6100db600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091908035906020019091905050610197565b604051808215151515815260200191505060405180910390f35b34610000576101026102d3565b6040518082815260200191505060405180910390f35b346100005761013360048080359060200190919050506102de565b604051808215151515815260200191505060405180910390f35b6000600060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490505b919050565b6000600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614156101d757600090506102cd565b3373ffffffffffffffffffffffffffffffffffffffff166108fc839081150290604051809050600060405180830381858888f193505050501561021d57600090506102cd565b81600060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540192505081905550816001600082825401925050819055508273ffffffffffffffffffffffffffffffffffffffff167fbddecaad150f4a9f75fb6864ff351a7f06b19c5b9cf533c22b5bc05ecebc0790836040518082815260200191505060405180910390a2600190505b92915050565b600060015490505b90565b600081600060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410156103305760009050610426565b3073ffffffffffffffffffffffffffffffffffffffff166108fc839081150290604051809050600060405180830381858888f19350505050156103765760009050610426565b81600060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540392505081905550816001600082825403925050819055503373ffffffffffffffffffffffffffffffffffffffff167f222838db2794d11532d940e8dec38ae307ed0b63cd97c233322e221f998767a6836040518082815260200191505060405180910390a2600190505b9190505600a165627a7a723058207e421fa2dcbcd7c959b6bef29e877962d006a7c9203251cbeec9598aa467a7490029")
	MortgageMappingPos                = hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000000")
	MortgageTotalMortgagePos          = hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000001")
	MortgageTotalRewardedPos          = hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000002")
)

// Total inner system reward
func GetTotalRewarded(state *state.StateDB) *big.Int {
	//positions of solidity storage var _totalRewarded
	pos_var_rewards := MortgageTotalRewardedPos

	//get _totalRewarded
	var key = common.BytesToHash(pos_var_rewards)
	x := state.GetState(MortgageContractAddr, key).Big()

	return x
}

// Total Mortgage POC
func GetTotalMortgage(state *state.StateDB) *big.Int {
	//positions of solidity storage var _totalRewarded
	pos_var_mortgage := MortgageTotalMortgagePos

	//get _totalRewarded
	var key = common.BytesToHash(pos_var_mortgage)
	x := state.GetState(MortgageContractAddr, key).Big()

	return x
}

func AddTotalRewarded(state *state.StateDB, r *big.Int) {
	//positions of solidity storage var _totalRewarded
	pos_var_rewards := MortgageTotalRewardedPos

	//get _totalRewarded
	var key = common.BytesToHash(pos_var_rewards)
	x := state.GetState(MortgageContractAddr, key).Big()
	y := common.BigToHash(new(big.Int).Add(x, r))

	state.SetState(MortgageContractAddr, key, y)
}

// Get mortgage amount
func mortgageOf(addr common.Address, state *state.StateDB) *big.Int {
	//positions of solidity storage var _mortgages
	pos_var_mortgages := MortgageMappingPos
	//key of mapping
	key_var_mortgages := addr.Hash().Bytes()

	var data = bytes.Join([][]byte{key_var_mortgages, pos_var_mortgages}, []byte{})
	var hash = common.BytesToHash(crypto.Keccak256(data))
	x := state.GetState(MortgageContractAddr, hash).Big()

	return x
}

func M(x *big.Int, nonce uint64) *big.Int {
	n := nonce + 1
	div := new(big.Int).Mul(big.NewInt(8), big.NewInt(int64(n)))
	x.Mul(x, bigInt2e20)
	return x.Div(x, div)
}

// Calculate mortgage amount function
func F(x *big.Int) *big.Int {
	//// CASE 1: mortgage is zero or more than 100
	if x.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}
	if x.Cmp(bigInt10e20) >= 0 {
		return MortgageOneBlockFullReward
	}

	//// CASE 2: mortgage is not zero
	oneFloat := big.NewFloat(1)
	oneInt := big.NewInt(1)
	k := new(big.Int).Mul(MortgageSystemK, MortgageOneBlockFullReward)
	e := float64(math2.E)
	a := 3
	b := 0.128

	floatX := new(big.Float).SetInt(new(big.Int).Div(x, bigInt10e18)) //x
	floatX.Mul(floatX, big.NewFloat(float64(b)))                      //bx
	floatX.Add(big.NewFloat(float64(a)), floatX.Neg(floatX))          //a-bx
	float64X, _ := floatX.Float64()                                   //
	ex := math2.Pow(e, float64X)                                      //e^(a-bx)
	tempx := big.NewFloat(ex)                                         //
	tempx.Add(tempx, oneFloat)                                        //1 + e^(a-bx)
	tempx.Mul(tempx, big.NewFloat(10000))                             //
	hx := new(big.Int)                                                //
	tempx.Int(hx)                                                     //
	y := oneInt.Div(k, hx)                                            //10000k/(10000*(1+e^(a-bx)))
	return y
}

func R(reserve *big.Int) *big.Int {
	pow := big.NewInt(1).Div(MortgageSystemMaxReward, reserve)
	log2 := float64(int(math2.Log2(float64(pow.Uint64()))))
	ratio := big.NewInt(int64(math2.Exp2(log2)))

	return ratio
}

var (
	PrivateTestAddress    = common.BytesToAddress(hexutil.MustDecode("0x50c5650E9c1f1D2AC839c587c687D447cD5143fa"))
	PrivateTestBalance, _ = big.NewInt(0).SetString("2000000000000000000000000", 10)
)

// Calculate miner reward
func CalcReward(coinbase common.Address, nonce uint64, state *state.StateDB) *big.Int {
	// Step 1. get fullBlockReward
	x := mortgageOf(coinbase, state)
	fullBlockReward := F(M(x, nonce))
	if fullBlockReward.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}

	// Step 2. get reserve reward
	rewarded := GetTotalRewarded(state)
	reserve := big.NewInt(1).Sub(MortgageSystemMaxReward, rewarded)
	if reserve.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0)
	}

	// Step 3. get ratio
	ratio := R(reserve)

	// Step 4. get real reward
	r := new(big.Int).Div(fullBlockReward, ratio)
	if r.Cmp(big.NewInt(1)) <= 0 {
		r = big.NewInt(1)
		log.Info("Poc CalcReward [less than 1, set to 1]")
	}
	log.Info("Poc CalcReward", "x", x, "reward", r, "full", fullBlockReward, "ratio", ratio, "rewarded", rewarded)
	return r
}
