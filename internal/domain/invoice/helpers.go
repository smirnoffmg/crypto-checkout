package invoice

import "github.com/shopspring/decimal"

func MustNewDecimal(value string) decimal.Decimal {
	d, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return d
}
