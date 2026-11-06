// geidea-terminal-test talks directly to a configured Geidea ECR terminal.
// It does not create a POS sale or payment record.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/NasTecSol/nembus-client/internal/payment/geidea"
)

func main() {
	action := flag.String("action", "status", "status, last, or purchase")
	mode := flag.String("mode", "network", "network or serial")
	ip := flag.String("ip", "", "terminal IP address")
	port := flag.String("port", "80", "terminal TCP port")
	serialPort := flag.String("serial-port", "", "terminal COM port for serial mode")
	baud := flag.Int("baud", 38400, "serial baud rate")
	amount := flag.Float64("amount", 0, "purchase amount in SAR")
	invoice := flag.String("invoice", "", "unique ECR reference for a new purchase")
	timeout := flag.Int("timeout", 120, "terminal timeout in seconds")
	flag.Parse()

	if *timeout <= 0 || (*mode != "network" && *mode != "serial") {
		fail("-timeout must be positive and -mode must be network or serial")
	}
	if (*mode == "network" && *ip == "") || (*mode == "serial" && *serialPort == "") {
		fail("network mode requires -ip; serial mode requires -serial-port")
	}
	if *action == "purchase" && (*amount <= 0 || *invoice == "") {
		fail("a purchase requires a positive -amount and a unique -invoice")
	}

	client := geidea.NewClient(geidea.TerminalConfig{
		Mode: *mode, IP: *ip, Port: *port, SerialPort: *serialPort,
		BaudRate: *baud, TimeoutSeconds: *timeout,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout+10)*time.Second)
	defer cancel()

	switch *action {
	case "status":
		result, err := client.ProbeConnection(ctx)
		if err != nil {
			fail("terminal status: %v", err)
		}
		fmt.Printf("connected=%t message=%s terminal=%s merchant=%s\n", result.OK, result.Message, result.TerminalID, result.MerchantID)
	case "last":
		result, err := client.RepeatLast(ctx)
		if err != nil {
			fail("last transaction: %v", err)
		}
		printResult(result)
	case "purchase":
		result, err := client.Purchase(ctx, geidea.PurchaseRequest{
			Amount: *amount, InvoiceNumber: *invoice,
		}, func(event geidea.TerminalStatusEvent) {
			fmt.Printf("terminal: %s\n", event.MessageEn)
		})
		if err != nil {
			fail("purchase result unknown; check the terminal slip or last transaction before retrying: %v", err)
		}
		printResult(result)
	default:
		fail("-action must be status, last, or purchase")
	}
}

func printResult(result *geidea.TransactionResult) {
	if result == nil {
		fail("terminal returned no transaction result")
	}
	fmt.Printf("success=%t code=%s message=%s amount=%.2f ECR=%s RRN=%s auth=%s card=%s\n",
		result.Success, result.ResponseCode, result.ResponseMessageEn, result.Amount,
		result.ECRNumber, result.RRN, result.AuthCode, result.CardNumber)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
