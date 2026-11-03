# Nembus Biometric Agent (`apps/biometric-agent`)

The **Biometric Agent** is a Windows desktop companion for **back-office and enrollment workstations** that use a DigitalPersona U.are.U USB fingerprint reader with the Nembus Cloud backoffice. It is **not** installed on POS terminals; the **Nembus POS client** embeds the same local bridge on register hardware.

---

## What it does

1. **Local hardware bridge** — HTTP API on `http://127.0.0.1:5050/api/v1` (`/device`, `/capture`, `/enrollment/*`) for browser-based backoffice flows.
2. **Driver setup** — Stages WHQL DigitalPersona USB drivers via UAC (`pnputil` / DPInst fallback).
3. **Setup dashboard** — `/setup` diagnostics UI (device status, test capture, driver repair).
4. **System tray** — Live reader/driver status with dynamic icon states (ready, capturing, disconnected, error).

---

## Quick commands

From the monorepo root:

```bash
make build-biometric-agent
make dev-biometric-agent
powershell -ExecutionPolicy Bypass -File scripts/install_biometric_agent.ps1 -Launch
powershell -ExecutionPolicy Bypass -File scripts/uninstall_biometric_agent.ps1
```

From `apps/biometric-agent` or the installed copy:

```bash
# Install Desktop + Start Menu shortcuts, icon, and sign-in auto-start
nembus-biometric-agent.exe --install

# Remove agent (shortcuts, auto-start, install folder, runtime cache)
nembus-biometric-agent.exe --uninstall
nembus-biometric-agent.exe --uninstall --uninstall-remove-drivers
```

Flags:

| Flag | Description |
|------|-------------|
| `--install` | Copy to `%LocalAppData%\\Programs\\Nembus`, create Desktop/Start Menu shortcuts with icon, register auto-start |
| `--uninstall` | Remove shortcuts, auto-start, install folder, and runtime cache |
| `--uninstall-remove-drivers` | With `--uninstall`, also remove DigitalPersona driver packages (admin) |
| `--setup` | Open `/setup` in the default browser after start |
| `--headless` | Run without system tray (Ctrl+C stops bridge gracefully) |
| `--status` | Print JSON diagnostics and exit |
| `--install-drivers` | Install/repair DigitalPersona USB drivers |

---

## Deployment model

| Machine type | Software | Bridge owner |
|--------------|----------|--------------|
| POS register | Nembus POS client | Embedded bridge in POS |
| Back-office PC | Nembus Biometric Agent | This agent |

Both speak the same **`127.0.0.1:5050/api/v1`** contract; they are not co-installed on the same device.

### U.are.U **WBF** readers (edge case)

Some scanners appear in Windows as **“U.are.U … (WBF)”** (Windows Biometric Framework). Legacy **`dpfpdd`** often reports **0** readers on that stack.

The agent **falls back to WinBio** (`winbio.dll`): enumeration, capture, and ISO **19794-2** FMD extraction via FingerJet — same local API as legacy readers (`WBF:{unitId}:…` device names).

On the setup dashboard, check **SDK Readers** (includes WinBio units when legacy count is 0):

| Count | Meaning |
|-------|---------|
| **≥ 1** | Capture and enrollment should work (legacy and/or WinBio native). |
| **0** with WBF in PnP | WinBio service did not enumerate the reader — restart **Windows Biometric Service** (`WbioSrvc`), replug USB, or run **Install / Repair Drivers** (admin). Legacy INF alone cannot replace a missing WinBio unit. |

For POS registers, prefer standard **non-WBF** U.are.U readers when possible; back-office PCs can use WBF-bound hardware via this agent.

---

## Security notes

- The bridge listens on **loopback only** (`127.0.0.1`).
- There is **no authentication** on local capture endpoints; any process on the same user session can invoke capture. This matches standard local biometric helper patterns—run only on trusted back-office machines.

---

## Release

Automated builds use `.github/workflows/release-biometric-agent.yml` (tag prefix `bio-v*`, upload app name `biometric-agent`).
