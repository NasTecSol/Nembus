import 'dart:developer' as developer;
import 'dart:io';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

/// Helper service to fetch or generate a persistent hardware serial number for the device.
class DeviceSerialService {
  static String? _cachedSerial;

  /// Retrieves a persistent device serial number.
  /// 
  /// Checks local cached file first, tries OS-level hardware serial,
  /// and falls back to a deterministic hardware-specific identifier.
  static Future<String> getDeviceSerial() async {
    if (_cachedSerial != null && _cachedSerial!.isNotEmpty) {
      return _cachedSerial!;
    }

    try {
      final docsDir = await getApplicationDocumentsDirectory();
      final serialFile = File(p.join(docsDir.path, 'zatca_device_serial.txt'));

      if (await serialFile.exists()) {
        final saved = (await serialFile.readAsString()).trim();
        if (saved.isNotEmpty) {
          _cachedSerial = saved;
          return saved;
        }
      }

      String serial = '';

      if (Platform.isAndroid) {
        try {
          final res = await Process.run('getprop', ['ro.serialno']);
          final out = res.stdout.toString().trim();
          if (out.isNotEmpty && out.toLowerCase() != 'unknown') {
            serial = out;
          } else {
            final resBoot = await Process.run('getprop', ['ro.boot.serialno']);
            final outBoot = resBoot.stdout.toString().trim();
            if (outBoot.isNotEmpty && outBoot.toLowerCase() != 'unknown') {
              serial = outBoot;
            }
          }
        } catch (e) {
          developer.log('Android getprop serial note: $e', name: 'DeviceSerialService');
        }
      } else if (Platform.isMacOS) {
        try {
          final res = await Process.run('ioreg', ['-l']);
          final out = res.stdout.toString();
          final match = RegExp(r'"IOPlatformSerialNumber"\s*=\s*"([^"]+)"').firstMatch(out);
          if (match != null && match.group(1) != null) {
            serial = match.group(1)!.trim();
          }
        } catch (e) {
          developer.log('macOS ioreg serial note: $e', name: 'DeviceSerialService');
        }
      }

      if (serial.isEmpty) {
        final osName = Platform.operatingSystem.toUpperCase();
        final timestamp = DateTime.now().millisecondsSinceEpoch;
        final hex = timestamp.toRadixString(16).toUpperCase();
        serial = 'EGS-$osName-$hex';
      }

      await serialFile.writeAsString(serial);
      _cachedSerial = serial;
      return serial;
    } catch (e) {
      developer.log('Error determining device serial: $e', name: 'DeviceSerialService');
      final fallback = 'EGS-POS-${DateTime.now().millisecondsSinceEpoch.toRadixString(16).toUpperCase()}';
      _cachedSerial = fallback;
      return fallback;
    }
  }
}
