package usecase

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var zatcaDecimal = regexp.MustCompile(`^[+-]?[0-9]+(?:\.[0-9]+)?$`)

func decimalValue(value string) (*big.Rat, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "0"
	}
	if !zatcaDecimal.MatchString(value) {
		return nil, fmt.Errorf("invalid decimal %q", value)
	}
	number, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, fmt.Errorf("invalid decimal %q", value)
	}
	return number, nil
}

func addMoney(a, b string) (string, error) {
	x, err := decimalValue(a)
	if err != nil {
		return "", err
	}
	y, err := decimalValue(b)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Add(x, y).FloatString(2), nil
}

func subtractMoney(a, b string) (string, error) {
	x, err := decimalValue(a)
	if err != nil {
		return "", err
	}
	y, err := decimalValue(b)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Sub(x, y).FloatString(2), nil
}

func validateInvoiceAmounts(input InvoiceXMLInput) error {
	if len(input.Lines) == 0 {
		return fmt.Errorf("ZATCA invoice requires at least one line")
	}
	netSum, taxSum := new(big.Rat), new(big.Rat)
	for i, line := range input.Lines {
		values := make([]*big.Rat, 6)
		for j, text := range []string{line.Quantity, line.UnitPrice, line.DiscountAmt, line.LineTotal, line.TaxAmount, line.TaxPercent} {
			value, err := decimalValue(text)
			if err != nil || value.Sign() < 0 {
				return fmt.Errorf("ZATCA line %d has an invalid or negative amount", i+1)
			}
			values[j] = value
		}
		if values[0].Sign() == 0 {
			return fmt.Errorf("ZATCA line %d quantity must be positive", i+1)
		}
		expectedNet := new(big.Rat).Sub(new(big.Rat).Mul(values[0], values[1]), values[2])
		if expectedNet.Sign() < 0 || expectedNet.FloatString(2) != values[3].FloatString(2) {
			return fmt.Errorf("ZATCA line %d net amount must equal quantity × unit_price − discount (%s), got %s", i+1, expectedNet.FloatString(2), values[3].FloatString(2))
		}
		expectedTax := new(big.Rat).Quo(new(big.Rat).Mul(values[3], values[5]), big.NewRat(100, 1))
		if expectedTax.FloatString(2) != values[4].FloatString(2) {
			return fmt.Errorf("ZATCA line %d tax must be %s, got %s", i+1, expectedTax.FloatString(2), values[4].FloatString(2))
		}
		netSum.Add(netSum, values[3])
		taxSum.Add(taxSum, values[4])
	}
	var totals [4]*big.Rat
	for i, text := range []string{input.Subtotal, input.DiscountAmount, input.TaxAmount, input.TotalAmount} {
		value, err := decimalValue(text)
		if err != nil || value.Sign() < 0 {
			return fmt.Errorf("ZATCA invoice totals must be non-negative decimals")
		}
		totals[i] = value
	}
	if netSum.FloatString(2) != totals[0].FloatString(2) {
		return fmt.Errorf("ZATCA subtotal must equal the sum of line net amounts (%s)", netSum.FloatString(2))
	}
	if taxSum.FloatString(2) != totals[2].FloatString(2) {
		return fmt.Errorf("ZATCA tax_amount must equal the sum of line taxes (%s)", taxSum.FloatString(2))
	}
	exclusive := new(big.Rat).Sub(totals[0], totals[1])
	if exclusive.Sign() < 0 {
		return fmt.Errorf("ZATCA discount cannot exceed subtotal")
	}
	rate, err := decimalValue(input.TaxPercent)
	if err != nil {
		return err
	}
	expectedTax := new(big.Rat).Quo(new(big.Rat).Mul(exclusive, rate), big.NewRat(100, 1))
	if expectedTax.FloatString(2) != totals[2].FloatString(2) {
		return fmt.Errorf("ZATCA document tax must equal taxable amount × tax rate (%s); allocate discounts consistently", expectedTax.FloatString(2))
	}
	expectedTotal := new(big.Rat).Add(exclusive, totals[2])
	if expectedTotal.FloatString(2) != totals[3].FloatString(2) {
		return fmt.Errorf("ZATCA total_amount must equal subtotal − discount + tax (%s)", expectedTotal.FloatString(2))
	}
	return nil
}

func validatePOSAmounts(in *PosCreateTransactionInput) error {
	input := InvoiceXMLInput{Subtotal: numericToString(in.Subtotal), DiscountAmount: numericToString(in.DiscountAmount), TaxAmount: numericToString(in.TaxAmount), TotalAmount: numericToString(in.TotalAmount), TaxPercent: "15.00"}
	for _, line := range in.Lines {
		input.Lines = append(input.Lines, InvoiceLineInput{Quantity: numericToString(line.Quantity), UnitPrice: numericToString(line.UnitPrice), DiscountAmt: numericToString(line.DiscountAmount), LineTotal: numericToString(line.Subtotal), TaxAmount: numericToString(line.TaxAmount), TaxPercent: "15.00"})
		expected, err := addMoney(numericToString(line.Subtotal), numericToString(line.TaxAmount))
		if err != nil {
			return err
		}
		actual, err := decimalValue(numericToString(line.LineTotal))
		if err != nil || actual.FloatString(2) != expected {
			return fmt.Errorf("ZATCA line_total must equal line subtotal + tax (%s)", expected)
		}
	}
	return validateInvoiceAmounts(input)
}
