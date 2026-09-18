import 'dart:convert';
import 'dart:developer' as developer;
import 'dart:ffi';
import 'dart:io';
import 'package:ffi/ffi.dart';

// Native function typedefs
typedef _CInitNembusMobile = Pointer<Utf8> Function(Pointer<Utf8> dbPath);
typedef _InitNembusMobile = Pointer<Utf8> Function(Pointer<Utf8> dbPath);

typedef _CFetchCompleteTenantMasterData = Pointer<Utf8> Function(
  Pointer<Utf8> tenantId,
  Pointer<Utf8> cloudUrl,
);
typedef _FetchCompleteTenantMasterData = Pointer<Utf8> Function(
  Pointer<Utf8> tenantId,
  Pointer<Utf8> cloudUrl,
);

typedef _CCallHandler = Pointer<Utf8> Function(Pointer<Utf8> reqJSON);
typedef _CallHandler = Pointer<Utf8> Function(Pointer<Utf8> reqJSON);

typedef _CFreeCString = Void Function(Pointer<Utf8> ptr);
typedef _FreeCString = void Function(Pointer<Utf8> ptr);

class NembusBridge {
  static NembusBridge? _instance;
  DynamicLibrary? _lib;
  _FreeCString? _freeCString;

  NembusBridge._() {
    try {
      _lib = _openLibrary();
      _freeCString = _resolveFreeCString();
    } catch (e) {
      developer.log('Native library load note: $e', name: 'NembusBridge');
    }
  }

  factory NembusBridge() {
    _instance ??= NembusBridge._();
    return _instance!;
  }

  static DynamicLibrary _openLibrary() {
    if (Platform.isIOS) {
      return DynamicLibrary.process();
    }
    if (Platform.isAndroid || Platform.isLinux) {
      return DynamicLibrary.open('libnembus_core.so');
    }
    if (Platform.isMacOS) {
      final candidates = [
        'libnembus_core.dylib',
        'apps/pos_mobile/libnembus_core.dylib',
        '../libnembus_core.dylib',
      ];
      for (final path in candidates) {
        try {
          return DynamicLibrary.open(path);
        } catch (_) {}
      }
      return DynamicLibrary.process();
    }
    if (Platform.isWindows) {
      return DynamicLibrary.open('libnembus_core.dll');
    }
    throw UnsupportedError(
        'Platform not supported: ${Platform.operatingSystem}');
  }

  _FreeCString _resolveFreeCString() {
    if (_lib == null) {
      return (Pointer<Utf8> ptr) {
        try {
          calloc.free(ptr);
        } catch (_) {}
      };
    }
    for (final name in [
      'FreeCString',
      '_FreeCString',
      'NembusffiFreeCString'
    ]) {
      try {
        return _lib!.lookupFunction<_CFreeCString, _FreeCString>(name);
      } catch (_) {}
    }
    return (Pointer<Utf8> ptr) {
      try {
        calloc.free(ptr);
      } catch (_) {}
    };
  }

  Map<String, dynamic> _parseAndFree(Pointer<Utf8>? resPtr) {
    if (resPtr == null || resPtr.address == 0) {
      return {'success': false, 'error': 'Null pointer received from Go FFI'};
    }
    try {
      final jsonStr = resPtr.toDartString();
      return jsonDecode(jsonStr) as Map<String, dynamic>;
    } catch (e) {
      return {'success': false, 'error': 'JSON decode error: $e'};
    } finally {
      if (_freeCString != null) {
        _freeCString!(resPtr);
      } else {
        try {
          calloc.free(resPtr);
        } catch (_) {}
      }
    }
  }

  /// Initializes the Go Core runtime with the local SQLite database path.
  Map<String, dynamic> initMobile(String dbPath) {
    developer.log('Calling Go FFI: InitNembusMobile("$dbPath")',
        name: 'NembusBridge');
    if (_lib == null) {
      return {'success': false, 'error': 'Native library not loaded'};
    }

    _InitNembusMobile? fn;
    for (final name in ['InitNembusMobile', '_InitNembusMobile']) {
      try {
        fn = _lib!.lookupFunction<_CInitNembusMobile, _InitNembusMobile>(name);
        break;
      } catch (_) {}
    }

    if (fn == null) {
      return {
        'success': false,
        'error': 'Symbol "InitNembusMobile" not found in native library'
      };
    }

    final pathPtr = dbPath.toNativeUtf8();
    try {
      final resPtr = fn(pathPtr);
      return _parseAndFree(resPtr);
    } catch (e) {
      return {'success': false, 'error': 'FFI call exception in initMobile: $e'};
    } finally {
      calloc.free(pathPtr);
    }
  }

  /// Triggers full master data backup stream from Cloud Server via gRPC.
  Map<String, dynamic> fetchCompleteTenantMasterData(
      String tenantId, String cloudUrl) {
    developer.log(
        'Calling Go FFI: FetchCompleteTenantMasterData($tenantId, $cloudUrl)',
        name: 'NembusBridge');
    if (_lib == null) {
      return {'success': false, 'error': 'Native library not loaded'};
    }

    _FetchCompleteTenantMasterData? fn;
    for (final name in [
      'FetchCompleteTenantMasterData',
      '_FetchCompleteTenantMasterData'
    ]) {
      try {
        fn = _lib!.lookupFunction<_CFetchCompleteTenantMasterData,
            _FetchCompleteTenantMasterData>(name);
        break;
      } catch (_) {}
    }

    if (fn == null) {
      return {
        'success': false,
        'error':
            'Symbol "FetchCompleteTenantMasterData" not found in native library'
      };
    }

    final tenantPtr = tenantId.toNativeUtf8();
    final urlPtr = cloudUrl.toNativeUtf8();
    try {
      final resPtr = fn(tenantPtr, urlPtr);
      return _parseAndFree(resPtr);
    } catch (e) {
      return {
        'success': false,
        'error': 'FFI call exception in fetchCompleteTenantMasterData: $e'
      };
    } finally {
      calloc.free(tenantPtr);
      calloc.free(urlPtr);
    }
  }

  /// Dispatches any business logic call directly to Go core handlers.
  /// All business logic executes exclusively inside Go (packages/core/handler).
  Map<String, dynamic> callHandler({
    required String handler,
    required String action,
    Map<String, dynamic>? payload,
  }) {
    if (_lib == null) {
      return {'success': false, 'error': 'Go native library not loaded'};
    }

    _CallHandler? fn;
    for (final name in ['CallHandler', '_CallHandler']) {
      try {
        fn = _lib!.lookupFunction<_CCallHandler, _CallHandler>(name);
        break;
      } catch (_) {}
    }

    if (fn == null) {
      return {
        'success': false,
        'error': 'Symbol "CallHandler" not found in Go native library'
      };
    }

    final reqMap = {
      'handler': handler,
      'action': action,
      'payload': payload ?? {}
    };
    final reqPtr = jsonEncode(reqMap).toNativeUtf8();
    try {
      final resPtr = fn(reqPtr);
      return _parseAndFree(resPtr);
    } catch (e) {
      return {
        'success': false,
        'error': 'FFI exception in CallHandler ($handler/$action): $e'
      };
    } finally {
      calloc.free(reqPtr);
    }
  }
}