import 'dart:convert';
import 'dart:developer' as developer;
import 'dart:io';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';

/// Describes the current operating mode of the POS application.
///
/// - [online]  — app is connected to the cloud; used from Splash through
///               Tenant Selection until a successful DB clone.
/// - [offline] — app operates entirely from local SQLite; used after a
///               successful clone or on subsequent cold starts with cloned data.
/// - [unknown] — initial state before Splash initialization completes.
enum AppMode { online, offline, unknown }

/// Application-wide singleton that holds shared configuration and runtime state.
class SingletonClass {
  factory SingletonClass() {
    _singleton ??= SingletonClass._internal();
    return _singleton!;
  }

  SingletonClass._internal();

  static SingletonClass? _singleton;

  /// Whether the singleton has been fully initialized.
  bool initialized = false;

  /// Base URL for the Nembus cloud server. Used for tenant lookup (ONLINE phase).
  final String baseURL = 'https://nembus.nashrms.com';

  /// Direct gRPC address for raw binary backup stream (no HTTPS).
  final String grpcAddress = 'nembus.nashrms.com:50051';

  /// Current operating mode. Set by SplashScreen after initialization.
  AppMode appMode = AppMode.unknown;

  /// The ID of the tenant currently loaded in the local SQLite database.
  String? activeTenantId;

  /// The slug of the tenant currently loaded in the local SQLite database.
  String? activeTenantSlug;

  /// The human-readable name of the active tenant.
  String? activeTenantName;

  /// The active store ID selected by the user.
  int? activeStoreId;

  /// The active store Name.
  String? activeStoreName;

  /// The active store Code.
  String? activeStoreCode;

  /// The active POS Terminal ID selected by the user.
  int? activeTerminalId;

  /// The active POS Terminal Name.
  String? activeTerminalName;

  /// The active POS Terminal Code.
  String? activeTerminalCode;

  /// Active Cashier Session Information
  int? activeCashierSessionId;
  String? activeCashierSessionNumber;
  String? activeCashierSessionStatus = 'closed';
  DateTime? activeCashierSessionOpenedAt;
  int? activeCashierId;
  String? activeCashierCode;

  /// Returns true if a cashier session is currently active and open.
  bool get isSessionActive =>
      activeCashierSessionId != null && activeCashierSessionStatus == 'open';

  bool get isCashierSessionActive => isSessionActive;

  /// Updates singleton state when a cashier session is started.
  Future<void> startCashierSession({
    required int sessionId,
    required String sessionNumber,
    int? cashierId,
    String? cashierCode,
    DateTime? openedAt,
  }) async {
    activeCashierSessionId = sessionId;
    activeCashierSessionNumber = sessionNumber;
    activeCashierSessionStatus = 'open';
    if (cashierId != null) activeCashierId = cashierId;
    if (cashierCode != null) activeCashierCode = cashierCode;
    activeCashierSessionOpenedAt = openedAt ?? DateTime.now();
    await _persistCurrentState();
    developer.log(
      'Cashier session started: ID=$sessionId, Number=$sessionNumber, CashierID=$activeCashierId',
      name: 'SingletonClass',
    );
  }

  /// Updates singleton state when a cashier session is stopped/closed.
  Future<void> closeCashierSession() async {
    final prevId = activeCashierSessionId;
    activeCashierSessionId = null;
    activeCashierSessionNumber = null;
    activeCashierSessionStatus = 'closed';
    activeCashierSessionOpenedAt = null;
    await _persistCurrentState();
    developer.log('Cashier session closed (previous ID=$prevId)', name: 'SingletonClass');
  }

  /// Automatically closes any active cashier session in SQLite, Go bridge, and sync queue.
  /// Used on app lifecycle detach/kill, background removal, or startup cleanup.
  Future<void> autoCloseActiveSession({String reason = 'Auto-closed on app termination'}) async {
    try {
      if (!DatabaseService().isOpen) {
        await DatabaseService().initDatabase();
      }
      final db = DatabaseService().database;

      // Find any open sessions in SQLite
      final openSessions = await db.query(
        'cashier_sessions',
        where: 'status = ?',
        whereArgs: ['open'],
      );

      if (openSessions.isEmpty) {
        await closeCashierSession();
        return;
      }

      final nowIso = DateTime.now().toIso8601String();

      for (final sess in openSessions) {
        final sessId = (sess['id'] as num).toInt();
        final sessNum = sess['session_number']?.toString() ?? 'SES-$sessId';
        final openingBalance = (sess['opening_balance'] as num?)?.toDouble() ?? 0.0;

        // Calculate cash sales for this session
        final salesSummary = await db.rawQuery('''
          SELECT COALESCE(SUM(
            CASE 
              WHEN LOWER(COALESCE(p.payment_method, '')) = 'cash' THEN (t.total_amount)
              ELSE 0 
            END
          ), 0) AS cash_sales
          FROM pos_transactions t
          LEFT JOIN pos_payments p ON p.transaction_id = t.id
          WHERE t.cashier_session_id = ? AND t.status = 'completed'
        ''', [sessId]);

        double cashSales = 0.0;
        if (salesSummary.isNotEmpty) {
          cashSales = (salesSummary.first['cash_sales'] as num?)?.toDouble() ?? 0.0;
        }

        final expectedBalance = openingBalance + cashSales;
        final closingBalance = expectedBalance;
        const variance = 0.0;

        developer.log(
          '🛑 [AutoClose] Closing session ID=$sessId ($sessNum): Expected=$expectedBalance, Reason=$reason',
          name: 'SingletonClass',
        );

        // Update SQLite
        await db.update(
          'cashier_sessions',
          {
            'closing_time': nowIso,
            'closing_balance': closingBalance,
            'expected_balance': expectedBalance,
            'variance': variance,
            'status': 'closed',
            'updated_at': nowIso,
          },
          where: 'id = ?',
          whereArgs: [sessId],
        );

        // Notify Go bridge
        try {
          NembusBridge().callHandler(
            handler: 'cashier_session',
            action: 'closeCashierSession',
            payload: {
              'id': sessId,
              'closing_balance': closingBalance.toStringAsFixed(2),
              'expected_balance': expectedBalance.toStringAsFixed(2),
              'variance': variance.toStringAsFixed(2),
              'closing_note': reason,
            },
          );
        } catch (_) {}

        // Enqueue to sync_queue
        final sessRows = await db.query('cashier_sessions', where: 'id = ?', whereArgs: [sessId]);
        if (sessRows.isNotEmpty) {
          await db.insert('sync_queue', {
            'entity_type': 'cashier_sessions',
            'entity_id': sessId.toString(),
            'action': 'UPDATE',
            'payload': jsonEncode(sessRows.first),
            'status': 'pending',
            'priority': 25,
            'correlation_id': sessNum,
            'created_at': nowIso,
          });
        }
      }

      await closeCashierSession();
    } catch (e) {
      developer.log('Auto-close cashier session note: $e', name: 'SingletonClass');
    }
  }

  /// State persistence file name
  static const String _stateFileName = 'nembus_app_state.json';

  Future<File> _getStateFile() async {
    final docsDir = await getApplicationDocumentsDirectory();
    return File(p.join(docsDir.path, _stateFileName));
  }

  Future<void> _persistCurrentState() async {
    try {
      final file = await _getStateFile();
      final data = {
        'tenant_id': activeTenantId,
        'tenant_slug': activeTenantSlug,
        'tenant_name': activeTenantName,
        'store_id': activeStoreId,
        'store_name': activeStoreName,
        'store_code': activeStoreCode,
        'terminal_id': activeTerminalId,
        'terminal_name': activeTerminalName,
        'terminal_code': activeTerminalCode,
        'cashier_session_id': activeCashierSessionId,
        'cashier_session_number': activeCashierSessionNumber,
        'cashier_session_status': activeCashierSessionStatus,
        'cashier_session_opened_at': activeCashierSessionOpenedAt?.toIso8601String(),
        'cashier_id': activeCashierId,
        'cashier_code': activeCashierCode,
        'is_setup_completed': activeTenantSlug != null,
        'saved_at': DateTime.now().toIso8601String(),
      };
      await file.writeAsString(jsonEncode(data));
    } catch (e) {
      developer.log('Failed to persist app state: $e', name: 'SingletonClass');
    }
  }

  Future<void> init() async {
    initialized = true;
    await loadTenantState();
  }

  /// Persists the selected and cloned tenant state to local storage.
  Future<void> saveTenantState({
    required String tenantId,
    required String tenantSlug,
    required String tenantName,
    int? storeId,
    String? storeName,
    String? storeCode,
    int? terminalId,
    String? terminalName,
    String? terminalCode,
  }) async {
    activeTenantId = tenantId;
    activeTenantSlug = tenantSlug;
    activeTenantName = tenantName;
    if (storeId != null) activeStoreId = storeId;
    if (storeName != null) activeStoreName = storeName;
    if (storeCode != null) activeStoreCode = storeCode;
    if (terminalId != null) activeTerminalId = terminalId;
    if (terminalName != null) activeTerminalName = terminalName;
    if (terminalCode != null) activeTerminalCode = terminalCode;
    appMode = AppMode.offline;

    try {
      final file = await _getStateFile();
      final data = {
        'tenant_id': tenantId,
        'tenant_slug': tenantSlug,
        'tenant_name': tenantName,
        'store_id': activeStoreId,
        'store_name': activeStoreName,
        'store_code': activeStoreCode,
        'terminal_id': activeTerminalId,
        'terminal_name': activeTerminalName,
        'terminal_code': activeTerminalCode,
        'is_setup_completed': true,
        'saved_at': DateTime.now().toIso8601String(),
      };
      await file.writeAsString(jsonEncode(data));
      developer.log('Tenant selection state persisted: $data', name: 'SingletonClass');
    } catch (e) {
      developer.log('Failed to persist tenant state: $e', name: 'SingletonClass');
    }
  }

  /// Persists the selected Store and Terminal configuration
  Future<void> saveStoreAndTerminalState({
    required int storeId,
    required String storeName,
    required String storeCode,
    required int terminalId,
    required String terminalName,
    required String terminalCode,
  }) async {
    activeStoreId = storeId;
    activeStoreName = storeName;
    activeStoreCode = storeCode;
    activeTerminalId = terminalId;
    activeTerminalName = terminalName;
    activeTerminalCode = terminalCode;

    if (activeTenantSlug != null) {
      await saveTenantState(
        tenantId: activeTenantId ?? '1',
        tenantSlug: activeTenantSlug!,
        tenantName: activeTenantName ?? 'Company Store',
        storeId: storeId,
        storeName: storeName,
        storeCode: storeCode,
        terminalId: terminalId,
        terminalName: terminalName,
        terminalCode: terminalCode,
      );
    }
  }

  /// Loads previously configured tenant state on app restart.
  /// Returns true if a valid cloned tenant configuration exists.
  Future<bool> loadTenantState() async {
    try {
      final file = await _getStateFile();
      if (!await file.exists()) {
        return false;
      }
      final content = await file.readAsString();
      if (content.trim().isEmpty) return false;

      final data = jsonDecode(content) as Map<String, dynamic>;
      final isCompleted = data['is_setup_completed'] == true;
      final slug = data['tenant_slug'] as String?;

      if (isCompleted && slug != null && slug.isNotEmpty) {
        activeTenantId = data['tenant_id'] as String?;
        activeTenantSlug = slug;
        activeTenantName = data['tenant_name'] as String?;
        if (data['store_id'] != null) activeStoreId = (data['store_id'] as num).toInt();
        activeStoreName = data['store_name'] as String?;
        activeStoreCode = data['store_code'] as String?;
        if (data['terminal_id'] != null) activeTerminalId = (data['terminal_id'] as num).toInt();
        activeTerminalName = data['terminal_name'] as String?;
        activeTerminalCode = data['terminal_code'] as String?;
        // Always initialize cashier session to closed on app start/tenant config load
        activeCashierSessionId = null;
        activeCashierSessionNumber = null;
        activeCashierSessionStatus = 'closed';
        activeCashierSessionOpenedAt = null;
        if (data['cashier_id'] != null) {
          activeCashierId = (data['cashier_id'] as num).toInt();
        }
        activeCashierCode = data['cashier_code'] as String?;
        appMode = AppMode.offline;
        developer.log('Tenant configuration restored: $slug ($activeTenantName), Store: $activeStoreName ($activeStoreId), Terminal: $activeTerminalName ($activeTerminalId), ActiveSession=$activeCashierSessionStatus', name: 'SingletonClass');
        return true;
      }
    } catch (e) {
      developer.log('Failed to load persisted tenant state: $e', name: 'SingletonClass');
    }
    return false;
  }

  /// Clears stored tenant state (e.g. on full factory reset).
  Future<void> clearTenantState() async {
    activeTenantId = null;
    activeTenantSlug = null;
    activeTenantName = null;
    activeStoreId = null;
    activeStoreName = null;
    activeStoreCode = null;
    activeTerminalId = null;
    activeTerminalName = null;
    activeTerminalCode = null;
    activeCashierSessionId = null;
    activeCashierSessionNumber = null;
    activeCashierSessionStatus = 'closed';
    activeCashierSessionOpenedAt = null;
    activeCashierId = null;
    activeCashierCode = null;
    appMode = AppMode.online;
    try {
      final file = await _getStateFile();
      if (await file.exists()) {
        await file.delete();
      }
      developer.log('Tenant state cleared.', name: 'SingletonClass');
    } catch (e) {
      developer.log('Failed to delete tenant state file: $e', name: 'SingletonClass');
    }
  }
}
