package converters

import (
	"math"
	"math/big"

	"github.com/ME-MotherEarth/me-core/core"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
)

const (
	numDecimalsInFloatBalance     = 10
	numDecimalsInFloatBalanceMECT = 18
)

var zero = big.NewInt(0)

type balanceConverter struct {
	dividerForDenomination float64
	balancePrecision       float64
	balancePrecisionMECT   float64
}

// NewBalanceConverter will create a new instance of balance converter
func NewBalanceConverter(denomination int) (*balanceConverter, error) {
	if denomination < 0 {
		return nil, indexer.ErrNegativeDenominationValue
	}

	return &balanceConverter{
		balancePrecision:       math.Pow(10, float64(numDecimalsInFloatBalance)),
		balancePrecisionMECT:   math.Pow(10, float64(numDecimalsInFloatBalanceMECT)),
		dividerForDenomination: math.Pow(10, float64(denomination)),
	}, nil
}

// ComputeBalanceAsFloat will compute balance as float
func (bc *balanceConverter) ComputeBalanceAsFloat(balance *big.Int) float64 {
	return bc.computeBalanceAsFloat(balance, bc.balancePrecision)
}

// ComputeMECTBalanceAsFloat will compute MECT balance as float
func (bc *balanceConverter) ComputeMECTBalanceAsFloat(balance *big.Int) float64 {
	return bc.computeBalanceAsFloat(balance, bc.balancePrecisionMECT)
}

func (bc *balanceConverter) computeBalanceAsFloat(balance *big.Int, balancePrecision float64) float64 {
	if balance == nil || balance.Cmp(zero) == 0 {
		return 0
	}

	balanceBigFloat := big.NewFloat(0).SetInt(balance)
	balanceFloat64, _ := balanceBigFloat.Float64()

	bal := balanceFloat64 / bc.dividerForDenomination

	balanceFloatWithDecimals := math.Round(bal*balancePrecision) / balancePrecision

	return core.MaxFloat64(balanceFloatWithDecimals, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bc *balanceConverter) IsInterfaceNil() bool {
	return bc == nil
}

// BigIntToString will convert a big.Int to string
func BigIntToString(value *big.Int) string {
	if value == nil {
		return "0"
	}

	return value.String()
}
