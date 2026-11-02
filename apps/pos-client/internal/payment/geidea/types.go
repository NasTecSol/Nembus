package geidea

import (
	"fmt"
	"time"
)

// Terminal wire control bytes
const (
	ByteSOH byte = 0x01
	ByteSTX byte = 0x02
	ByteETX byte = 0x03
	ByteACK byte = 0x06
	ByteNAK byte = 0x15

	// AppIDSAMA is the static SAMA Application ID defined by the standard
	AppIDSAMA string = "11"
)

// Transaction Types
type TransactionType string

const (
	TxTypePurchase       TransactionType = "1"
	TxTypePurchaseNaqd   TransactionType = "2"
	TxTypeRefund         TransactionType = "3"
	TxTypeReconciliation TransactionType = "4"
)

// Intermediate progress status codes sent by terminal during transaction processing
const (
	StatusPaperOut                byte = 0x4F
	StatusSwipeOrInsert           byte = 0x41
	StatusUserCancelledOrTimeout  byte = 0x47
	StatusCardInserted            byte = 0x58
	StatusCardReadError           byte = 0x70
	StatusCardRemoved             byte = 0x53
	StatusDataError               byte = 0x66
	StatusBlockedCard             byte = 0x65
	StatusNoActiveApplication     byte = 0x42
	StatusInsertCardSmartReader   byte = 0x6D
	StatusMaxAmountLimitExceeded  byte = 0x77
	StatusPinEntry                byte = 0x43
	StatusPinQuit                 byte = 0x45
	StatusNetworkDownCable        byte = 0x49
	StatusInsertSIM               byte = 0x4A
	StatusConnecting              byte = 0x46
	StatusReconnecting            byte = 0x4B
	StatusNetworkDown             byte = 0x71
	StatusConnectionMade          byte = 0x4D
	StatusSendingRequest          byte = 0x7A
	StatusTransactionCancelled    byte = 0x52
	StatusWaitingResponse         byte = 0x79
	StatusWrongPin                byte = 0x44
	StatusRetryLimitExceeded      byte = 0x57
	StatusConnectionClosed        byte = 0x51
	StatusRemoveCard              byte = 0x50
	StatusSAFProcessing           byte = 0x38
	StatusMaxSAFReached           byte = 0x56
	StatusMagStripeNotSupported   byte = 0x4C
	StatusCardSchemeNotSupported  byte = 0x55
	StatusDeSAFFailed             byte = 0x61
	StatusReconciliationNoBusiness byte = 0x62
	StatusReconciliationFailed   byte = 0x63
	StatusReconciliationSuccess  byte = 0x64
	StatusIdleScreen              byte = 0x39
	StatusTerminalTMSNotLoaded    byte = 0x75
	StatusTerminalBusy            byte = 0x07
)

// TerminalStatusEvent represents an intermediate event during processing
type TerminalStatusEvent struct {
	Code      byte   `json:"code"`
	HexCode   string `json:"hex_code"`
	MessageEn string `json:"message_en"`
	MessageAr string `json:"message_ar"`
}

// StatusMessages maps status code to English and Arabic UI messages
var StatusMessages = map[byte]struct {
	En string
	Ar string
}{
	StatusSwipeOrInsert:          {En: "Please Swipe, Tap, or Insert Card", Ar: "يرجى تمرير أو إدخال البطاقة"},
	StatusCardInserted:           {En: "Card Inserted, Processing...", Ar: "تم إدخال البطاقة، جاري المعالجة..."},
	StatusPinEntry:               {En: "Please Enter PIN on Terminal", Ar: "يرجى إدخال الرقم السري على الجهاز"},
	StatusWrongPin:               {En: "Incorrect PIN, Please Try Again", Ar: "الرقم السري غير صحيح، يرجى المحاولة مرة أخرى"},
	StatusPinQuit:                {En: "PIN Entry Cancelled", Ar: "تم إلغاء إدخال الرقم السري"},
	StatusRemoveCard:             {En: "Please Remove Card", Ar: "يرجى سحب البطاقة"},
	StatusCardRemoved:            {En: "Card Removed", Ar: "تم سحب البطاقة"},
	StatusConnecting:             {En: "Connecting to Payment Gateway...", Ar: "جاري الاتصال بنظام المدفوعات..."},
	StatusConnectionMade:         {En: "Connected to Gateway", Ar: "تم الاتصال بنظام المدفوعات"},
	StatusSendingRequest:         {En: "Authorizing Payment...", Ar: "جاري اعتماد الدفع..."},
	StatusWaitingResponse:        {En: "Waiting for Bank Response...", Ar: "في انتظار رد البنك..."},
	StatusConnectionClosed:       {En: "Connection Closed", Ar: "تم إغلاق الاتصال"},
	StatusTransactionCancelled:   {En: "Transaction Cancelled", Ar: "تم إلغاء العملية"},
	StatusUserCancelledOrTimeout: {En: "User Cancelled or Timed Out", Ar: "تم الإلغاء أو انتهاء الوقت"},
	StatusPaperOut:               {En: "Terminal Paper Out", Ar: "نفد ورق الطابعة بالجهاز"},
	StatusCardReadError:          {En: "Card Read Error", Ar: "خطأ في قراءة البطاقة"},
	StatusDataError:              {En: "Terminal Data Error", Ar: "خطأ في بيانات الجهاز"},
	StatusBlockedCard:            {En: "Card is Blocked", Ar: "البطاقة محظورة"},
	StatusNoActiveApplication:    {En: "No Active Application on Card", Ar: "لا يوجد تطبيق صالح في البطاقة"},
	StatusInsertCardSmartReader:  {En: "Use Chip Reader (Smart Card)", Ar: "يرجى استخدام قارئ الشريحة الذكية"},
	StatusMaxAmountLimitExceeded: {En: "Maximum Amount Limit Exceeded", Ar: "تم تجاوز الحد الأقصى للمبلغ"},
	StatusNetworkDownCable:       {En: "Network Down, Check Cable", Ar: "الشبكة مفصولة، يرجى فحص الكابل"},
	StatusInsertSIM:              {En: "Insert SIM Card", Ar: "يرجى إدخال شريحة البيانات"},
	StatusReconnecting:           {En: "Reconnecting to Gateway...", Ar: "جاري إعادة الاتصال..."},
	StatusNetworkDown:            {En: "SAMA / Bank Network Down", Ar: "شبكة مدى / البنك غير متوفرة"},
	StatusRetryLimitExceeded:     {En: "Retry Limit Exceeded", Ar: "تم تجاوز عدد المحاولات"},
	StatusSAFProcessing:          {En: "SAF Processing...", Ar: "جاري معالجة العمليات المؤجلة..."},
	StatusMaxSAFReached:          {En: "Maximum SAF Limit Reached", Ar: "تم الوصول للحد الأقصى للعمليات المؤجلة"},
	StatusMagStripeNotSupported:  {En: "Magnetic Stripe Not Supported", Ar: "الشريط المغناطيسي غير مدعوم"},
	StatusCardSchemeNotSupported: {En: "Card Scheme Not Supported", Ar: "نوع البطاقة غير مدعوم"},
	StatusDeSAFFailed:            {En: "De-SAF Failed", Ar: "فشلت معالجة العمليات المؤجلة"},
	StatusReconciliationNoBusiness: {En: "Reconciliation: No Transactions Found", Ar: "الموازنة: لا توجد عمليات"},
	StatusReconciliationFailed:   {En: "Reconciliation Failed", Ar: "فشلت الموازنة اليومية"},
	StatusReconciliationSuccess:  {En: "Reconciliation Successful", Ar: "تمت الموازنة اليومية بنجاح"},
	StatusIdleScreen:             {En: "Terminal Ready", Ar: "الجهاز جاهز للعمل"},
	StatusTerminalTMSNotLoaded:   {En: "Terminal TMS Not Registered/Loaded", Ar: "برنامج الجهاز غير مهيأ أو غير مسجل"},
	StatusTerminalBusy:           {En: "Terminal is Busy", Ar: "الجهاز مشغول حالياً"},
}

// GetStatusEvent returns a structured status event for a status code
func GetStatusEvent(code byte) TerminalStatusEvent {
	hexCode := fmt.Sprintf("%02X", code)
	msg, exists := StatusMessages[code]
	if !exists {
		return TerminalStatusEvent{
			Code:      code,
			HexCode:   hexCode,
			MessageEn: fmt.Sprintf("Terminal Event %s", hexCode),
			MessageAr: fmt.Sprintf("حالة الجهاز %s", hexCode),
		}
	}
	return TerminalStatusEvent{
		Code:      code,
		HexCode:   hexCode,
		MessageEn: msg.En,
		MessageAr: msg.Ar,
	}
}

// SAMA SPAN2 / mada Response Codes
var ResponseCodes = map[string]struct {
	Approved bool
	En       string
	Ar       string
}{
	"000": {Approved: true, En: "Approved", Ar: "مقبولة"},
	"001": {Approved: true, En: "Honour with Identification", Ar: "مقبولة مع التحقق"},
	"003": {Approved: true, En: "Approved (VIP)", Ar: "مقبولة (VIP)"},
	"007": {Approved: true, En: "Approved, Update ICC", Ar: "مقبولة مع تحديث البطاقة"},
	"087": {Approved: true, En: "Offline Approved (Chip Only)", Ar: "مقبولة بدون اتصال"},
	"089": {Approved: true, En: "Unable to Go Online, Offline Approved", Ar: "مقبولة بدون اتصال (تعذر الاتصال)"},
	"100": {Approved: false, En: "Do Not Honour", Ar: "مرفوضة من البنك"},
	"101": {Approved: false, En: "Expired Card", Ar: "بطاقة منتهية الصلاحية"},
	"102": {Approved: false, En: "Suspected Fraud", Ar: "اشتباه في احتيال"},
	"103": {Approved: false, En: "Contact Card Issuer / Acquirer", Ar: "يرجى الاتصال بالبنك"},
	"104": {Approved: false, En: "Restricted Card", Ar: "بطاقة مقيدة"},
	"105": {Approved: false, En: "Call Acquirer Security", Ar: "يرجى مراجعة أمن البنك"},
	"106": {Approved: false, En: "PIN Tries Exceeded", Ar: "تم تجاوز عدد محاولات الرقم السري"},
	"107": {Approved: false, En: "Refer to Card Issuer", Ar: "يرجى مراجعة البنك المصدر"},
	"108": {Approved: false, En: "Refer to Card Issuer Special Conditions", Ar: "شروط خاصة بالبنك المصدر"},
	"109": {Approved: false, En: "Invalid Merchant", Ar: "التاجر غير صالح"},
	"110": {Approved: false, En: "Invalid Amount", Ar: "المبلغ غير صالح"},
	"111": {Approved: false, En: "Invalid Card Number", Ar: "رقم البطاقة غير صالح"},
	"112": {Approved: false, En: "PIN Data Required", Ar: "مطلوب إدخال الرقم السري"},
	"114": {Approved: false, En: "No Account of Type Requested", Ar: "نوع الحساب غير متوفر"},
	"115": {Approved: false, En: "Requested Function Not Supported", Ar: "الخدمة غير مدعومة"},
	"116": {Approved: false, En: "Insufficient Funds", Ar: "الرصيد غير كافٍ"},
	"117": {Approved: false, En: "Incorrect PIN", Ar: "الرقم السري خاطئ"},
	"118": {Approved: false, En: "No Card Record", Ar: "لا يوجد سجل للبطاقة"},
	"119": {Approved: false, En: "Transaction Not Permitted to Cardholder", Ar: "العملية غير مسموحة لحامل البطاقة"},
	"120": {Approved: false, En: "Transaction Not Permitted to Terminal", Ar: "العملية غير مسموحة للجهاز"},
	"121": {Approved: false, En: "Exceeds Withdrawal Amount Limit", Ar: "تم تجاوز الحد الأقصى للسحب"},
	"122": {Approved: false, En: "Security Violation", Ar: "انتهاك أمني"},
	"123": {Approved: false, En: "Exceeds Withdrawal Frequency Limit", Ar: "تم تجاوز عدد مرات السحب المسموحة"},
	"125": {Approved: false, En: "Card Not Effective", Ar: "البطاقة غير مفعلة"},
	"188": {Approved: false, En: "Offline Declined", Ar: "مرفوضة بدون اتصال"},
	"190": {Approved: false, En: "Unable to Go Online - Offline Declined", Ar: "تعذر الاتصال - مرفوضة"},
	"400": {Approved: true, En: "Accepted", Ar: "مقبولة"},
	"902": {Approved: false, En: "Invalid Transaction", Ar: "عملية غير صالحة"},
	"903": {Approved: false, En: "Re-enter Transaction", Ar: "يرجى إعادة إدخال العملية"},
	"904": {Approved: false, En: "Format Error", Ar: "خطأ في تنسيق البيانات"},
	"907": {Approved: false, En: "Card Issuer or Switch Inoperative", Ar: "البنك المصدر أو المقسم معطل"},
	"909": {Approved: false, En: "System Malfunction", Ar: "عطل في النظام"},
	"911": {Approved: false, En: "Card Issuer Timed Out", Ar: "انتهت مهلة استجابة البنك"},
	"912": {Approved: false, En: "Card Issuer Unavailable", Ar: "البنك المصدر غير متوفر"},
	"913": {Approved: false, En: "Duplicate Transmission", Ar: "عملية مكررة"},
	"940": {Approved: false, En: "Unknown Terminal", Ar: "جهاز غير معرف"},
}

// CardSchemeName maps scheme codes to display names
var CardSchemeNames = map[string]string{
	"P1": "mada",
	"VC": "Visa",
	"MC": "MasterCard",
	"AX": "American Express",
	"DM": "Maestro",
	"UP": "UnionPay",
	"GN": "GCC Net",
}

// TerminalConfig holds connection parameters
type TerminalConfig struct {
	// Mode is "network" (TCP) or "serial" (COM port)
	Mode string `json:"mode"`

	// IP is the terminal's IP address (for network mode, e.g. "192.168.1.50")
	IP string `json:"ip,omitempty"`

	// Port is the TCP port (defaults to "80" if empty)
	Port string `json:"port,omitempty"`

	// SerialPort is the COM port name (for serial mode, e.g. "COM3")
	SerialPort string `json:"serial_port,omitempty"`

	// BaudRate for serial connection (defaults to 38400)
	BaudRate int `json:"baud_rate,omitempty"`

	// TimeoutSeconds for the entire transaction flow (defaults to 120)
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

// PurchaseRequest describes a payment request
type PurchaseRequest struct {
	Amount         float64 `json:"amount"`                    // In SAR (e.g. 50.25)
	CashbackAmount float64 `json:"cashback_amount,omitempty"` // In SAR (0 for standard purchase)
	InvoiceNumber  string  `json:"invoice_number"`            // POS invoice/bill reference (up to 16 digits)
}

// RefundRequest describes a refund request
type RefundRequest struct {
	Amount         float64 `json:"amount"`           // In SAR
	InvoiceNumber  string  `json:"invoice_number"`   // Reference
	RRN            string  `json:"rrn"`              // Retrieval Reference Number of original purchase (12 chars)
	MaskedCardNumber string `json:"masked_card_number,omitempty"`
}

// TransactionResult contains the parsed data from the terminal
type TransactionResult struct {
	Success        bool      `json:"success"`
	ResponseCode   string    `json:"response_code"`
	ResponseMessageEn string `json:"response_message_en"`
	ResponseMessageAr string `json:"response_message_ar"`

	CardNumber     string    `json:"card_number"` // Masked, e.g. 5888********3225
	CardScheme     string    `json:"card_scheme"` // Display name, e.g. "mada", "Visa"
	CardSchemeCode string    `json:"card_scheme_code"` // Wire code, e.g. "P1"
	Amount         float64   `json:"amount"`
	CashbackAmount float64   `json:"cashback_amount"`
	STAN           string    `json:"stan"` // System Trace Audit Number
	DateTime       time.Time `json:"date_time"`
	ExpiryDate     string    `json:"expiry_date"` // MM/YY
	RRN            string    `json:"rrn"`         // Retrieval Reference Number
	AuthCode       string    `json:"auth_code"`   // Approval Code
	TerminalID     string    `json:"terminal_id"`
	MerchantID     string    `json:"merchant_id"`
	ECRNumber      string    `json:"ecr_number"`  // POS Invoice number echoed back
	BankID         string    `json:"bank_id"`     // Acquirer bank
	MCC            string    `json:"mcc"`         // Merchant Category Code

	// EMV compliance tags
	CardEntryMode  string    `json:"card_entry_mode"` // DIPPED, CONTACTLESS, SWIPED
	AID            string    `json:"aid"`
	TVR            string    `json:"tvr"`
	TSI            string    `json:"tsi"`
	CVR            string    `json:"cvr"`
	Cryptogram     string    `json:"cryptogram"`

	RawHexResponse string    `json:"raw_hex_response,omitempty"`
}

// SchemeTotal represents scheme totals in a reconciliation report
type SchemeTotal struct {
	SchemeCode string  `json:"scheme_code"`
	SchemeName string  `json:"scheme_name"`
	BankID     string  `json:"bank_id"`
	Count      int     `json:"count"`
	Amount     float64 `json:"amount"`
}

// ReconciliationResult contains the parsed End-of-Day reconciliation report
type ReconciliationResult struct {
	Success        bool          `json:"success"`
	ResponseCode   string        `json:"response_code"`
	STAN           string        `json:"stan"`
	TerminalID     string        `json:"terminal_id"`
	MerchantID     string        `json:"merchant_id"`
	DateTime       time.Time     `json:"date_time"`
	TerminalTotals []SchemeTotal `json:"terminal_totals"`
	HostTotals     []SchemeTotal `json:"host_totals"`
	RawHexResponse string        `json:"raw_hex_response,omitempty"`
}

// StatusCallback is invoked whenever the terminal sends an intermediate status update
type StatusCallback func(event TerminalStatusEvent)
