# Windows x64 DigitalPersona runtime assets

Place the following files from **DigitalPersona Biometric SDK for Windows 3.4.0** redistributable merge modules (`Redist/*.msm`) in this directory for embedded plug-and-play extraction:

- `dpfj.dll`, `dpfr6.dll`, `dpfr7.dll` (FingerJet matching)
- `dpfpdd.dll`, `dpfpdd_4k.dll`, `dpfpdd5000.dll`, `dpfpdd7k.dll`, `dpfpdd_lumi.dll` (capture)
- `dpDevCtlx64.dll`, `dpDevDatx64.dll`, `dpd00701x64.dll`, `dpi00701x64.dll` (device control)

USB driver files (`dPersona_x64.inf`, `UsbdpFP.sys`, `dpk00701.sys`, `dpersona_x64.cat`) are already vendored for silent `pnputil` staging.

If matching/capture DLLs are absent, `InitRuntime()` falls back to `C:\Windows\System32\dpfj.dll` and `dpfpdd.dll` when the RTE is installed.
