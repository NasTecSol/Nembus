package geidea

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// CalculateLRC calculates the Longitudinal Redundancy Check (XOR) over a byte slice.
func CalculateLRC(data []byte) byte {
	var lrc byte
	for _, b := range data {
		lrc ^= b
	}
	return lrc
}

// formatAmountTo12Hex formats a float SAR amount to 12 hex digits (6 bytes) in halalas.
// Example: 50.25 -> 5025 halalas -> "000000005025" -> []byte{0x00, 0x00, 0x00, 0x00, 0x50, 0x25}
func formatAmountTo12Hex(amount float64) ([]byte, error) {
	if amount < 0 {
		return nil, fmt.Errorf("amount cannot be negative: %.2f", amount)
	}
	halalas := int64(math.Round(amount * 100))
	halalaStr := fmt.Sprintf("%012d", halalas)
	if len(halalaStr) != 12 {
		return nil, fmt.Errorf("amount exceeds maximum supported digits: %.2f", amount)
	}

	buf := make([]byte, 6)
	for i := 0; i < 6; i++ {
		b, err := strconv.ParseUint(halalaStr[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("failed to encode amount nibbles: %w", err)
		}
		buf[i] = byte(b)
	}
	return buf, nil
}

// formatECRNumber formats an invoice number into 16 ASCII bytes (e.g. "0000000000000001")
func formatECRNumber(ecr string) []byte {
	// Keep only alphanumeric characters, max 16 chars
	var clean strings.Builder
	for _, ch := range ecr {
		if (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			clean.WriteRune(ch)
		}
	}
	s := clean.String()
	if len(s) > 16 {
		s = s[len(s)-16:]
	}
	// Pad with leading zeros up to 16
	for len(s) < 16 {
		s = "0" + s
	}
	return []byte(s)
}

// formatRRN formats a Retrieval Reference Number to 12 ASCII bytes
func formatRRN(rrn string) []byte {
	clean := strings.TrimSpace(rrn)
	if len(clean) > 12 {
		clean = clean[len(clean)-12:]
	}
	for len(clean) < 12 {
		clean = "0" + clean
	}
	return []byte(clean)
}

// BuildPurchaseBuffer creates the hex byte buffer for a purchase or cashback transaction
func BuildPurchaseBuffer(amount float64, ecrRef string, cashback float64) ([]byte, error) {
	amtBytes, err := formatAmountTo12Hex(amount)
	if err != nil {
		return nil, err
	}
	ecrBytes := formatECRNumber(ecrRef)

	var payload []byte
	if cashback > 0 {
		cbBytes, err := formatAmountTo12Hex(cashback)
		if err != nil {
			return nil, err
		}
		// Length: 0x29 (41 bytes)
		payload = append(payload, 0x29)
		payload = append(payload, 0xFF, 0x01, 0x00)
		payload = append(payload, 0xDF, 0x04, 0x06)
		payload = append(payload, amtBytes...)
		payload = append(payload, 0xDF, 0x36, 0x06)
		payload = append(payload, cbBytes...)
		payload = append(payload, 0xDF, 0x81, 0x10)
		payload = append(payload, ecrBytes...)
		payload = append(payload, 0x11) // App ID
	} else {
		// Length: 0x20 (32 bytes)
		payload = append(payload, 0x20)
		payload = append(payload, 0xFF, 0x01, 0x00)
		payload = append(payload, 0xDF, 0x04, 0x06)
		payload = append(payload, amtBytes...)
		payload = append(payload, 0xDF, 0x81, 0x10)
		payload = append(payload, ecrBytes...)
		payload = append(payload, 0x11) // App ID
	}

	lrc := CalculateLRC(payload)

	var frame []byte
	frame = append(frame, ByteSOH, 0x00)
	frame = append(frame, payload...)
	frame = append(frame, lrc)
	return frame, nil
}

// BuildRefundBuffer creates the hex byte buffer for a refund transaction
func BuildRefundBuffer(amount float64, ecrRef string, rrn string) ([]byte, error) {
	amtBytes, err := formatAmountTo12Hex(amount)
	if err != nil {
		return nil, err
	}
	ecrBytes := formatECRNumber(ecrRef)
	rrnBytes := formatRRN(rrn)

	// Length: 0x2E (46 bytes)
	var payload []byte
	payload = append(payload, 0x2E)
	payload = append(payload, 0xFF, 0x01, 0x02)
	payload = append(payload, 0xDF, 0x04, 0x06)
	payload = append(payload, amtBytes...)
	payload = append(payload, 0xDF, 0x25, 0x0C)
	payload = append(payload, rrnBytes...)
	payload = append(payload, 0xDF, 0x81, 0x10)
	payload = append(payload, ecrBytes...)

	lrc := CalculateLRC(payload)

	var frame []byte
	frame = append(frame, ByteSOH, 0x00)
	frame = append(frame, payload...)
	frame = append(frame, lrc)
	return frame, nil
}

// BuildReconciliationBuffer creates the buffer to trigger End of Day reconciliation
func BuildReconciliationBuffer() []byte {
	// Frame: 01 00 03 FF 01 09 F4
	payload := []byte{0x03, 0xFF, 0x01, 0x09}
	lrc := CalculateLRC(payload)
	return append([]byte{ByteSOH, 0x00}, append(payload, lrc)...)
}

// BuildRepeatLastBuffer creates the buffer to retrieve the last transaction response
func BuildRepeatLastBuffer() []byte {
	// Frame: 01 00 03 FF 01 0B F6
	payload := []byte{0x03, 0xFF, 0x01, 0x0B}
	lrc := CalculateLRC(payload)
	return append([]byte{ByteSOH, 0x00}, append(payload, lrc)...)
}

// BuildCancelBuffer creates the abort buffer to cancel an in-progress transaction
func BuildCancelBuffer() []byte {
	// Standard ECR cancel: STX '7' '2' ETX ACK (02 37 32 03 06)
	return []byte{ByteSTX, 0x37, 0x32, 0x03, ByteACK}
}

// BuildCheckStatusBuffer creates the buffer to query terminal availability
func BuildCheckStatusBuffer() []byte {
	// Frame: 01 00 03 FF 01 04 F9
	payload := []byte{0x03, 0xFF, 0x01, 0x04}
	lrc := CalculateLRC(payload)
	return append([]byte{ByteSOH, 0x00}, append(payload, lrc)...)
}

// BuildGetTerminalIDBuffer creates the buffer to query the terminal's TID
func BuildGetTerminalIDBuffer() []byte {
	// Frame: 01 00 03 FF 01 07 FA
	payload := []byte{0x03, 0xFF, 0x01, 0x07}
	lrc := CalculateLRC(payload)
	return append([]byte{ByteSOH, 0x00}, append(payload, lrc)...)
}

// BuildSetSettingBuffer creates the buffer to set menu, printer, and baud rate
func BuildSetSettingBuffer(menu bool, printer bool, speedOption int) []byte {
	menuByte := byte(0x30)
	if menu {
		menuByte = 0x31
	}
	printerByte := byte(0x30)
	if printer {
		printerByte = 0x31
	}
	if speedOption < 1 || speedOption > 6 {
		speedOption = 5 // default 38400
	}
	speedByte := byte(0x30 + speedOption)

	// Payload: 09 FF 01 06 DF 0D 03 menu printer speed
	payload := []byte{0x09, 0xFF, 0x01, 0x06, 0xDF, 0x0D, 0x03, menuByte, printerByte, speedByte}
	lrc := CalculateLRC(payload)
	return append([]byte{ByteSOH, 0x00}, append(payload, lrc)...)
}

// ParseIntermediateStatus checks if the buffer contains a 6-byte terminal status update
// Format: 01 00 02 FF <StatusCode> <LRC>
func ParseIntermediateStatus(buf []byte) (byte, bool) {
	for i := 0; i <= len(buf)-6; i++ {
		if buf[i] == 0x01 && buf[i+1] == 0x00 && buf[i+2] == 0x02 && buf[i+3] == 0xFF {
			statusCode := buf[i+4]
			expectedLRC := 0x02 ^ 0xFF ^ statusCode
			if buf[i+5] == expectedLRC {
				return statusCode, true
			}
		}
	}
	return 0, false
}

// TLVTag holds a tag, length, and raw bytes value
type TLVTag struct {
	Tag   string
	Raw   []byte
	Value string
}

// ExtractTLVs extracts all DFxx tags from a response buffer
func ExtractTLVs(buf []byte) map[string]TLVTag {
	tags := make(map[string]TLVTag)
	i := 0
	for i < len(buf) {
		// Look for 'DF' prefix (0xDF)
		if buf[i] == 0xDF {
			if i+2 > len(buf) {
				break
			}
			tagByte2 := buf[i+1]
			tag := fmt.Sprintf("DF%02X", tagByte2)
			i += 2

			if i >= len(buf) {
				break
			}

			// Parse length: can be single byte or two-byte (0x81 len)
			length := 0
			if buf[i] == 0x81 {
				i++
				if i >= len(buf) {
					break
				}
				length = int(buf[i])
				i++
			} else {
				length = int(buf[i])
				i++
			}

			if length < 0 || i+length > len(buf) {
				// Malformed length, take whatever remains
				length = len(buf) - i
			}

			valBytes := buf[i : i+length]
			i += length

			tags[tag] = TLVTag{
				Tag:   tag,
				Raw:   valBytes,
				Value: string(valBytes),
			}
		} else {
			i++
		}
	}
	return tags
}

// parseHalalaBytes converts 6 numeric/BCD bytes into float SAR amount
// Example: []byte{0x00, 0x00, 0x00, 0x00, 0x50, 0x25} -> 50.25
func parseHalalaBytes(raw []byte) float64 {
	hexStr := hex.EncodeToString(raw)
	val, err := strconv.ParseInt(hexStr, 10, 64)
	if err != nil {
		return 0.0
	}
	return float64(val) / 100.0
}

// ParseTransactionResponse parses the terminal's final response buffer
func ParseTransactionResponse(buf []byte) (*TransactionResult, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("empty response buffer")
	}

	tlvs := ExtractTLVs(buf)
	if len(tlvs) == 0 {
		return nil, fmt.Errorf("no TLV tags found in response (hex: %s)", hex.EncodeToString(buf))
	}

	res := &TransactionResult{
		RawHexResponse: hex.EncodeToString(buf),
	}

	// 1. Response Code (DF27)
	if tag, ok := tlvs["DF27"]; ok {
		res.ResponseCode = strings.TrimSpace(tag.Value)
	}
	if info, ok := ResponseCodes[res.ResponseCode]; ok {
		res.Success = info.Approved
		res.ResponseMessageEn = info.En
		res.ResponseMessageAr = info.Ar
	} else if res.ResponseCode != "" {
		res.Success = false
		res.ResponseMessageEn = fmt.Sprintf("Response Code %s", res.ResponseCode)
		res.ResponseMessageAr = fmt.Sprintf("رمز الرد %s", res.ResponseCode)
	}

	// 2. Amount (DF04)
	if tag, ok := tlvs["DF04"]; ok {
		res.Amount = parseHalalaBytes(tag.Raw)
	}

	// 3. Card Number (DF02) - Masked PAN
	if tag, ok := tlvs["DF02"]; ok {
		res.CardNumber = strings.TrimSpace(tag.Value)
	}

	// 4. STAN (DF0B)
	if tag, ok := tlvs["DF0B"]; ok {
		res.STAN = strings.TrimSpace(tag.Value)
	}

	// 5. Date and Time (DF0C) - Format: YYMMDDhhmmss
	if tag, ok := tlvs["DF0C"]; ok {
		dtStr := strings.TrimSpace(tag.Value)
		if len(dtStr) >= 12 {
			if t, err := time.Parse("060102150405", dtStr[:12]); err == nil {
				res.DateTime = t
			}
		}
	}
	if res.DateTime.IsZero() {
		res.DateTime = time.Now()
	}

	// 6. Card Expiry Date (DF0E)
	if tag, ok := tlvs["DF0E"]; ok {
		exp := strings.TrimSpace(tag.Value)
		if len(exp) == 4 {
			res.ExpiryDate = fmt.Sprintf("%s/%s", exp[2:4], exp[0:2])
		} else {
			res.ExpiryDate = exp
		}
	}

	// 7. RRN (DF25)
	if tag, ok := tlvs["DF25"]; ok {
		res.RRN = strings.TrimSpace(tag.Value)
	}

	// 8. Auth / Approval Code (DF26)
	if tag, ok := tlvs["DF26"]; ok {
		res.AuthCode = strings.TrimSpace(tag.Value)
	}

	// 9. Terminal ID (DF29)
	if tag, ok := tlvs["DF29"]; ok {
		res.TerminalID = strings.TrimSpace(tag.Value)
	}

	// 10. Merchant ID (DF2A)
	if tag, ok := tlvs["DF2A"]; ok {
		res.MerchantID = strings.TrimSpace(tag.Value)
	}

	// 11. Additional / NAQD Amount (DF36)
	if tag, ok := tlvs["DF36"]; ok {
		res.CashbackAmount = parseHalalaBytes(tag.Raw)
	}

	// 12. ECR Reference Number (DF81)
	if tag, ok := tlvs["DF81"]; ok {
		res.ECRNumber = strings.TrimSpace(tag.Value)
	}

	// 13. Card Scheme (DF82)
	if tag, ok := tlvs["DF82"]; ok {
		schemeCode := strings.TrimSpace(tag.Value)
		res.CardSchemeCode = schemeCode
		if name, exists := CardSchemeNames[schemeCode]; exists {
			res.CardScheme = name
		} else {
			res.CardScheme = schemeCode
		}
	}

	// 14. Bank ID (DF87)
	if tag, ok := tlvs["DF87"]; ok {
		res.BankID = strings.TrimSpace(tag.Value)
	}

	// 15. MCC (DF88)
	if tag, ok := tlvs["DF88"]; ok {
		res.MCC = strings.TrimSpace(tag.Value)
	}

	// 16. EMV Tags (DF90)
	// Example format: SPAN|DIPPED|A0000002281010|0000048000|E800|00|420000|7C012B29...
	if tag, ok := tlvs["DF90"]; ok {
		parts := strings.Split(tag.Value, "|")
		if len(parts) >= 1 && res.CardScheme == "" {
			res.CardScheme = parts[0]
		}
		if len(parts) >= 2 {
			res.CardEntryMode = parts[1]
		}
		if len(parts) >= 3 {
			res.AID = parts[2]
		}
		if len(parts) >= 4 {
			res.TVR = parts[3]
		}
		if len(parts) >= 5 {
			res.TSI = parts[4]
		}
		if len(parts) >= 6 {
			res.CVR = parts[5]
		}
		if len(parts) >= 7 {
			res.Cryptogram = parts[6]
		}
	}

	return res, nil
}

// ParseReconciliationResponse parses the End of Day settlement report
func ParseReconciliationResponse(buf []byte) (*ReconciliationResult, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("empty reconciliation buffer")
	}

	tlvs := ExtractTLVs(buf)
	res := &ReconciliationResult{
		RawHexResponse: hex.EncodeToString(buf),
		DateTime:       time.Now(),
	}

	if tag, ok := tlvs["DF27"]; ok {
		res.ResponseCode = strings.TrimSpace(tag.Value)
		res.Success = (res.ResponseCode == "000" || res.ResponseCode == "501")
	}
	if tag, ok := tlvs["DF0B"]; ok {
		res.STAN = strings.TrimSpace(tag.Value)
	}
	if tag, ok := tlvs["DF29"]; ok {
		res.TerminalID = strings.TrimSpace(tag.Value)
	}
	if tag, ok := tlvs["DF2A"]; ok {
		res.MerchantID = strings.TrimSpace(tag.Value)
	}

	// Parse terminal scheme totals (DF93)
	if tag, ok := tlvs["DF93"]; ok {
		res.TerminalTotals = parseSchemeTotals(tag.Value)
	}

	// Parse host scheme totals (DF94)
	if tag, ok := tlvs["DF94"]; ok {
		res.HostTotals = parseSchemeTotals(tag.Value)
	}

	return res, nil
}

// parseSchemeTotals parses the comma-separated reconciliation totals string
func parseSchemeTotals(s string) []SchemeTotal {
	var totals []SchemeTotal
	tokens := strings.Split(s, ",")
	// Token pattern typically repeats: SchemeCode, BankID, Count, Amount, ...
	for i := 0; i+3 < len(tokens); i++ {
		schemeCode := strings.TrimSpace(tokens[i])
		if _, ok := CardSchemeNames[schemeCode]; ok {
			bankID := strings.TrimSpace(tokens[i+1])
			count, _ := strconv.Atoi(strings.TrimSpace(tokens[i+2]))
			amount, _ := strconv.ParseFloat(strings.TrimSpace(tokens[i+3]), 64)

			totals = append(totals, SchemeTotal{
				SchemeCode: schemeCode,
				SchemeName: CardSchemeNames[schemeCode],
				BankID:     bankID,
				Count:      count,
				Amount:     amount,
			})
			i += 3
		}
	}
	return totals
}

// IsTerminalBusyResponse checks if the buffer corresponds to terminal busy
func IsTerminalBusyResponse(buf []byte) bool {
	// Busy response starts with: 06 01 00 01 06 07 or contains 00 01 06 07
	busyPattern := []byte{0x00, 0x01, 0x06, 0x07}
	return bytes.Contains(buf, busyPattern)
}
