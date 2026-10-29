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
      final candidates = [
        'nembus_core.framework/nembus_core',
        'Frameworks/nembus_core.framework/nembus_core',
      ];
      for (final path in candidates) {
        try {
          return DynamicLibrary.open(path);
        } catch (_) {}
      }
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

  DynamicLibrary? _executableLib;
  DynamicLibrary? get _execLib {
    if (!Platform.isIOS) return null;
    try {
      _executableLib ??= DynamicLibrary.executable();
    } catch (_) {}
    return _executableLib;
  }

  List<DynamicLibrary> get _libraries {
    final libs = <DynamicLibrary>[];
    if (_lib != null) libs.add(_lib!);
    final exec = _execLib;
    if (exec != null) libs.add(exec);
    return libs;
  }

  _FreeCString _resolveFreeCString() {
    for (final lib in _libraries) {
      for (final name in ['FreeCString', '_FreeCString', 'NembusffiFreeCString']) {
        try {
          return lib.lookupFunction<_CFreeCString, _FreeCString>(name);
        } catch (_) {}
      }
    }
    return (Pointer<Utf8> ptr) {
      try {
        calloc.free(ptr);
      } catch (_) {}
    };
  }

  _InitNembusMobile? _resolveInitMobile() {
    for (final lib in _libraries) {
      for (final name in ['InitNembusMobile', '_InitNembusMobile']) {
        try {
          return lib.lookupFunction<_CInitNembusMobile, _InitNembusMobile>(name);
        } catch (_) {}
      }
    }
    return null;
  }

  _FetchCompleteTenantMasterData? _resolveFetchMasterData() {
    for (final lib in _libraries) {
      for (final name in ['FetchCompleteTenantMasterData', '_FetchCompleteTenantMasterData']) {
        try {
          return lib.lookupFunction<_CFetchCompleteTenantMasterData, _FetchCompleteTenantMasterData>(name);
        } catch (_) {}
      }
    }
    return null;
  }

  _CallHandler? _resolveCallHandler() {
    for (final lib in _libraries) {
      for (final name in ['CallHandler', '_CallHandler']) {
        try {
          return lib.lookupFunction<_CCallHandler, _CallHandler>(name);
        } catch (_) {}
      }
    }
    return null;
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
    final fn = _resolveInitMobile();

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

  /// Triggers full master data backup stream from Cloud Server via gRPC or HTTP fallback.
  Map<String, dynamic> fetchCompleteTenantMasterData(
      String tenantId, String cloudUrl, {String? dbPath}) {
    if (dbPath != null && dbPath.isNotEmpty) {
      initMobile(dbPath);
    }
    developer.log(
        'Calling Go FFI: FetchCompleteTenantMasterData($tenantId, $cloudUrl)',
        name: 'NembusBridge');
    final fn = _resolveFetchMasterData();

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
    final fn = _resolveCallHandler();

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