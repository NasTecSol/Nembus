//go:build windows

package installer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// Reader stack constants reported to the setup dashboard.
const (
	ReaderStackLegacy  = "legacy"
	ReaderStackWBF     = "wbf"
	ReaderStackHello   = "hello"
	ReaderStackUnknown = "unknown"
)

// USBReaderPnP describes a DigitalPersona USB reader visible to Windows Plug-and-Play.
type USBReaderPnP struct {
	Present           bool   `json:"present"`
	FriendlyName      string `json:"friendly_name,omitempty"`
	InstanceID        string `json:"instance_id,omitempty"`
	Class             string `json:"class,omitempty"`
	Status            string `json:"status,omitempty"`
	DriverDesc        string `json:"driver_desc,omitempty"`
	ReaderStack       string `json:"reader_stack,omitempty"`
	UsingWindowsHello bool   `json:"using_windows_hello_driver"`
}

// ListUSBReaders returns VID_05BA fingerprint readers registered in Windows PnP.
func ListUSBReaders() []USBReaderPnP {
	out, err := runHiddenPowerShell(`
$items = Get-PnpDevice | Where-Object { $_.InstanceId -match 'VID_05BA' }
if (-not $items) { '[]'; exit 0 }
$list = foreach ($i in $items) {
  $driver = ''
  try {
    $driver = (Get-PnpDeviceProperty -InstanceId $i.InstanceId -KeyName 'DEVPKEY_Device_DriverDesc' -ErrorAction Stop).Data
  } catch {}
  $stack = 'unknown'
  $name = [string]$i.FriendlyName
  $drv = [string]$driver
  if ($name -match 'WBF' -or $drv -match 'WBF' -or $drv -match 'Windows Biometric Framework') {
    $stack = 'wbf'
  } elseif ($i.Class -eq 'Authentication Devices') {
    $stack = 'hello'
  } elseif ($i.Class -eq 'Biometric' -and ($drv -match 'UsbdpFP|DigitalPersona|U\.are\.U' -and $drv -notmatch 'WBF')) {
    $stack = 'legacy'
  } elseif ($i.Class -eq 'Biometric') {
    $stack = 'wbf'
  }
  [PSCustomObject]@{
    Status = [string]$i.Status
    Class = [string]$i.Class
    FriendlyName = $name
    InstanceId = [string]$i.InstanceId
    DriverDesc = $drv
    ReaderStack = $stack
  }
}
$list | ConvertTo-Json -Compress
`)
	if err != nil || len(bytes.TrimSpace(out)) == 0 {
		return nil
	}
	raw := bytes.TrimSpace(out)
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}
	var rows []map[string]string
	switch v := decoded.(type) {
	case []any:
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			rows = append(rows, mapFromAny(m))
		}
	case map[string]any:
		rows = append(rows, mapFromAny(v))
	}
	result := make([]USBReaderPnP, 0, len(rows))
	for _, row := range rows {
		className := row["Class"]
		stack := row["ReaderStack"]
		if stack == "" {
			stack = classifyReaderStack(row["FriendlyName"], row["DriverDesc"], className)
		}
		r := USBReaderPnP{
			Present:           true,
			FriendlyName:      row["FriendlyName"],
			InstanceID:        row["InstanceId"],
			Class:             className,
			Status:            row["Status"],
			DriverDesc:        row["DriverDesc"],
			ReaderStack:       stack,
			UsingWindowsHello: stack == ReaderStackHello,
		}
		result = append(result, r)
	}
	return result
}

func classifyReaderStack(friendlyName, driverDesc, className string) string {
	combined := strings.ToUpper(friendlyName + " " + driverDesc)
	if strings.Contains(combined, "WBF") || strings.Contains(combined, "WINDOWS BIOMETRIC FRAMEWORK") {
		return ReaderStackWBF
	}
	if strings.EqualFold(className, "Authentication Devices") {
		return ReaderStackHello
	}
	if strings.EqualFold(className, "Biometric") {
		if strings.Contains(strings.ToUpper(driverDesc), "USBDPFP") || strings.Contains(strings.ToUpper(driverDesc), "DIGITALPERSONA") {
			return ReaderStackLegacy
		}
		return ReaderStackWBF
	}
	return ReaderStackUnknown
}

func mapFromAny(m map[string]any) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func hasLegacyBiometricReader(readers []USBReaderPnP) bool {
	for _, r := range readers {
		if r.ReaderStack == ReaderStackLegacy && strings.EqualFold(r.Class, "Biometric") {
			return true
		}
	}
	return false
}

func hasWBFReader(readers []USBReaderPnP) bool {
	for _, r := range readers {
		if r.ReaderStack == ReaderStackWBF {
			return true
		}
	}
	return false
}

func needsDriverRebind(readers []USBReaderPnP) bool {
	for _, r := range readers {
		if r.ReaderStack == ReaderStackHello || r.ReaderStack == ReaderStackWBF {
			return true
		}
		if !strings.EqualFold(r.Class, "Biometric") {
			return true
		}
	}
	return false
}

func primaryReaderStack(readers []USBReaderPnP) string {
	if len(readers) == 0 {
		return ""
	}
	for _, r := range readers {
		if r.ReaderStack == ReaderStackWBF {
			return ReaderStackWBF
		}
	}
	return readers[0].ReaderStack
}

func rebindDigitalPersonaReaders(infPath string) error {
	escaped := strings.ReplaceAll(infPath, `'`, `''`)
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Continue'
$InfPath = '%s'
if ($InfPath -and (Test-Path -LiteralPath $InfPath)) {
  pnputil /add-driver $InfPath /install 2>&1 | Out-String | Write-Host
}
$readers = Get-PnpDevice | Where-Object { $_.InstanceId -match 'VID_05BA' }
foreach ($r in $readers) {
  $driver = ''
  try { $driver = (Get-PnpDeviceProperty -InstanceId $r.InstanceId -KeyName 'DEVPKEY_Device_DriverDesc').Data } catch {}
  $isWbf = ([string]$r.FriendlyName -match 'WBF') -or ([string]$driver -match 'WBF') -or ($r.Class -ne 'Biometric') -or ($r.Class -eq 'Biometric' -and [string]$driver -match 'WBF')
  if ($isWbf -or $r.Class -ne 'Biometric') {
    Write-Host "Removing device node for legacy rebind: $($r.FriendlyName) [$($r.InstanceId)] driver=$driver class=$($r.Class)"
    pnputil /remove-device "$($r.InstanceId)" 2>&1 | Out-String | Write-Host
  }
}
Start-Sleep -Seconds 3
pnputil /scan-devices 2>&1 | Out-String | Write-Host
`, escaped)
	_, err := runHiddenPowerShell(script)
	return err
}

func runHiddenPowerShell(script string) ([]byte, error) {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.CombinedOutput()
}
