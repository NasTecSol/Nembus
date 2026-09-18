import 'dart:convert';
import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import 'login_screen.dart';
import 'pos_product_list_screen.dart';
import 'stock_count_list_screen.dart';

class PosDashboardScreen extends StatefulWidget {
  final String tenantSlug;
  final String tenantName;
  final int userId;
  final String username;
  final String roleName;
  final int? storeId;
  final String? storeName;
  final int? posTerminalId;
  final String? posTerminalName;

  const PosDashboardScreen({
    super.key,
    required this.tenantSlug,
    required this.tenantName,
    this.userId = 1,
    this.username = 'Cashier',
    this.roleName = 'Cashier',
    this.storeId,
    this.storeName,
    this.posTerminalId,
    this.posTerminalName,
  });

  @override
  State<PosDashboardScreen> createState() => _PosDashboardScreenState();
}

class _PosDashboardScreenState extends State<PosDashboardScreen> {
  bool _isLoading = true;
  String? _errorMessage;
  List<Map<String, dynamic>> _assignedMenus = [];
  Map<String, dynamic>? _rbacDiagnostics;
  String _searchQuery = '';

  // Outbox Sync & Queue State
  int _pendingOutboxCount = 0;
  int _syncedOutboxCount = 0;
  int _failedOutboxCount = 0;
  bool _isSyncingOutbox = false;
  bool _isPullingCloudData = false;
  String? _lastSyncedAt;

  // Cashier Session State
  bool _isSessionLoading = false;
  int? _activeSessionId;
  String? _activeSessionNumber;
  String _activeSessionStatus = 'closed';
  DateTime? _sessionOpenedAt;
  int? _cashierId;
  String? _cashierCode;

  @override
  void initState() {
    super.initState();
    _restoreActiveCashierSession();
    _fetchUserAssignedMenus();
    _fetchSyncQueueStatus();
  }

  /// Restores any ongoing active open cashier session if one was started in the current runtime session or stored in SQLite
  Future<void> _restoreActiveCashierSession() async {
    try {
      final singleton = SingletonClass();
      if (singleton.isSessionActive && singleton.activeCashierSessionId != null) {
        if (mounted) {
          setState(() {
            _activeSessionId = singleton.activeCashierSessionId;
            _activeSessionNumber = singleton.activeCashierSessionNumber;
            _activeSessionStatus = 'open';
            _sessionOpenedAt = singleton.activeCashierSessionOpenedAt ?? DateTime.now();
            _cashierId = singleton.activeCashierId;
            _cashierCode = singleton.activeCashierCode;
          });
        }
        return;
      }

      // If in-memory state is not active, session remains closed until user explicitly starts it
      if (mounted) {
        setState(() {
          _activeSessionId = null;
          _activeSessionNumber = null;
          _activeSessionStatus = 'closed';
          _sessionOpenedAt = null;
        });
      }
    } catch (e) {
      developer.log('Error restoring cashier session: $e', name: 'PosDashboardScreen');
    }
  }

  /// Ensures a cashier record exists in local SQLite for the current user
  Future<int> _ensureCashierRecord() async {
    if (_cashierId != null && _cashierId! > 0) return _cashierId!;

    final db = DatabaseService().database;
    final storeId = widget.storeId ?? SingletonClass().activeStoreId ?? 1;
    final cashierByUser = await db.query('cashiers', where: 'user_id = ?', whereArgs: [widget.userId], limit: 1);
    if (cashierByUser.isNotEmpty) {
      final cid = (cashierByUser.first['id'] as num).toInt();
      _cashierId = cid;
      _cashierCode = cashierByUser.first['cashier_code']?.toString() ?? 'CASH-${widget.userId}';
      return cid;
    }

    try {
      final resp = NembusBridge().callHandler(
        handler: 'cashier',
        action: 'getCashierByUserID',
        payload: {
          'user_id': widget.userId,
          'store_id': storeId,
        },
      );
      if (resp['success'] == true && resp['data'] is Map) {
        final data = Map<String, dynamic>.from(resp['data'] as Map);
        final cid = (data['id'] as num).toInt();
        _cashierId = cid;
        _cashierCode = data['cashier_code']?.toString() ?? 'CASH-${widget.userId}';
        return cid;
      }
    } catch (_) {}

    final cashierCode = 'CASH-${widget.userId}';
    final cid = await db.insert('cashiers', {
      'user_id': widget.userId,
      'store_id': storeId,
      'cashier_code': cashierCode,
      'is_active': 1,
      'created_at': DateTime.now().toIso8601String(),
    });
    _cashierId = cid;
    _cashierCode = cashierCode;
    return cid;
  }

  /// Starts a new Cashier Session using Go handler and stores in local SQLite DB
  Future<void> _startCashierSession({double openingBalance = 0.0}) async {
    if (_isSessionLoading) return;
    setState(() => _isSessionLoading = true);

    try {
      final cashierId = await _ensureCashierRecord();
      final posTerminalId = widget.posTerminalId ?? SingletonClass().activeTerminalId ?? 1;
      final sessionNumber = 'SES-${DateTime.now().millisecondsSinceEpoch}';
      final now = DateTime.now();

      developer.log(
        '🚀 [Cashier Session] Starting session: cashier=$cashierId, terminal=$posTerminalId, number=$sessionNumber, float=$openingBalance',
        name: 'PosDashboardScreen',
      );

      // 1. Call Go handler
      final resp = NembusBridge().callHandler(
        handler: 'cashier_session',
        action: 'openCashierSession',
        payload: {
          'cashier_id': cashierId,
          'pos_terminal_id': posTerminalId,
          'session_number': sessionNumber,
          'opening_balance': openingBalance.toStringAsFixed(2),
        },
      );
      developer.log('Go openCashierSession resp: $resp', name: 'PosDashboardScreen');

      int sessionId = 0;
      if (resp['success'] == true && resp['data'] is Map) {
        final data = Map<String, dynamic>.from(resp['data'] as Map);
        sessionId = (data['id'] as num?)?.toInt() ?? 0;
      }

      // 2. Insert or update in local SQLite DB
      final db = DatabaseService().database;
      if (sessionId > 0) {
        final existing = await db.query('cashier_sessions', where: 'id = ?', whereArgs: [sessionId]);
        if (existing.isEmpty) {
          await db.insert('cashier_sessions', {
            'id': sessionId,
            'cashier_id': cashierId,
            'pos_terminal_id': posTerminalId,
            'session_number': sessionNumber,
            'opening_time': now.toIso8601String(),
            'opening_balance': openingBalance,
            'expected_balance': openingBalance,
            'status': 'open',
            'created_at': now.toIso8601String(),
            'updated_at': now.toIso8601String(),
          });
        }
      } else {
        sessionId = await db.insert('cashier_sessions', {
          'cashier_id': cashierId,
          'pos_terminal_id': posTerminalId,
          'session_number': sessionNumber,
          'opening_time': now.toIso8601String(),
          'opening_balance': openingBalance,
          'expected_balance': openingBalance,
          'status': 'open',
          'created_at': now.toIso8601String(),
          'updated_at': now.toIso8601String(),
        });
      }

      // 3. Enqueue to sync_queue for cloud synchronization
      final sessRows = await db.query('cashier_sessions', where: 'id = ?', whereArgs: [sessionId]);
      if (sessRows.isNotEmpty) {
        await db.insert('sync_queue', {
          'entity_type': 'cashier_sessions',
          'entity_id': sessionId.toString(),
          'action': 'INSERT',
          'payload': jsonEncode(sessRows.first),
          'status': 'pending',
          'priority': 25,
          'correlation_id': sessionNumber,
          'created_at': now.toIso8601String(),
        });
        await _fetchSyncQueueStatus();
      }

      // 4. Update Singleton state
      await SingletonClass().startCashierSession(
        sessionId: sessionId,
        sessionNumber: sessionNumber,
        cashierId: cashierId,
        cashierCode: _cashierCode,
        openedAt: now,
      );

      if (mounted) {
        setState(() {
          _activeSessionId = sessionId;
          _activeSessionNumber = sessionNumber;
          _activeSessionStatus = 'open';
          _sessionOpenedAt = now;
          _isSessionLoading = false;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFF10B981),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            content: Row(
              children: [
                const Icon(Icons.play_circle_filled_rounded, color: Colors.white),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Cashier Session Started (#$sessionNumber)',
                        style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 13, color: Colors.white),
                      ),
                      Text(
                        'Cart and sales are now fully unlocked!',
                        style: GoogleFonts.inter(fontSize: 11, color: Colors.white70),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      }
    } catch (e) {
      developer.log('Error starting cashier session: $e', name: 'PosDashboardScreen');
      if (mounted) {
        setState(() => _isSessionLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFEF4444),
            content: Text('Failed to start cashier session: $e'),
          ),
        );
      }
    }
  }

  /// Closes the active Cashier Session using Go handler and updates local SQLite DB
  Future<void> _stopCashierSession({
    required double closingBalance,
    required double expectedBalance,
    required double variance,
    String closingNote = 'Closed from Dashboard',
    VoidCallback? onSessionClosed,
  }) async {
    if (_activeSessionId == null || _isSessionLoading) return;
    setState(() => _isSessionLoading = true);

    try {
      final sessionId = _activeSessionId!;
      final sessionNum = _activeSessionNumber ?? 'SES-$sessionId';
      final now = DateTime.now();

      developer.log('🛑 [Cashier Session] Closing session ID=$sessionId: Expected=$expectedBalance, Closing=$closingBalance, Variance=$variance', name: 'PosDashboardScreen');

      // 1. Call Go handler
      final resp = NembusBridge().callHandler(
        handler: 'cashier_session',
        action: 'closeCashierSession',
        payload: {
          'id': sessionId,
          'closing_balance': closingBalance.toStringAsFixed(2),
          'expected_balance': expectedBalance.toStringAsFixed(2),
          'variance': variance.toStringAsFixed(2),
          'closed_by': widget.userId,
          'closing_note': closingNote,
        },
      );
      developer.log('Go closeCashierSession resp: $resp', name: 'PosDashboardScreen');

      // 2. Update local SQLite DB
      final db = DatabaseService().database;
      await db.update(
        'cashier_sessions',
        {
          'closing_time': now.toIso8601String(),
          'closing_balance': closingBalance,
          'expected_balance': expectedBalance,
          'variance': variance,
          'status': 'closed',
          'updated_at': now.toIso8601String(),
        },
        where: 'id = ?',
        whereArgs: [sessionId],
      );

      // 3. Enqueue session update to sync_queue
      final sessRows = await db.query('cashier_sessions', where: 'id = ?', whereArgs: [sessionId]);
      if (sessRows.isNotEmpty) {
        await db.insert('sync_queue', {
          'entity_type': 'cashier_sessions',
          'entity_id': sessionId.toString(),
          'action': 'UPDATE',
          'payload': jsonEncode(sessRows.first),
          'status': 'pending',
          'priority': 25,
          'correlation_id': sessionNum,
          'created_at': now.toIso8601String(),
        });
        await _fetchSyncQueueStatus();
      }

      // 4. Update Singleton state
      await SingletonClass().closeCashierSession();

      if (mounted) {
        setState(() {
          _activeSessionId = null;
          _activeSessionNumber = null;
          _activeSessionStatus = 'closed';
          _sessionOpenedAt = null;
          _isSessionLoading = false;
        });

        final varianceStr = variance >= 0 
            ? '+SAR ${variance.toStringAsFixed(2)}' 
            : '-SAR ${variance.abs().toStringAsFixed(2)}';

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFF10B981),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            content: Row(
              children: [
                const Icon(Icons.check_circle_rounded, color: Colors.white),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Cashier Session Closed ($sessionNum)',
                        style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 13, color: Colors.white),
                      ),
                      Text(
                        'Expected: SAR ${expectedBalance.toStringAsFixed(2)} | Actual: SAR ${closingBalance.toStringAsFixed(2)} | Variance: $varianceStr',
                        style: GoogleFonts.inter(fontSize: 11, color: Colors.white70),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );

        if (onSessionClosed != null) {
          onSessionClosed();
        }
      }
    } catch (e) {
      developer.log('Error closing cashier session: $e', name: 'PosDashboardScreen');
      if (mounted) {
        setState(() => _isSessionLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFEF4444),
            content: Text('Failed to close cashier session: $e'),
          ),
        );
      }
    }
  }

  /// Dialog to confirm and optionally specify opening float
  void _showStartSessionDialog() {
    final floatController = TextEditingController(text: '0.00');
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: const Color(0xFF131B2E),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
          side: BorderSide(color: const Color(0xFF10B981).withValues(alpha: 0.4)),
        ),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFF10B981).withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(10),
              ),
              child: const Icon(Icons.play_arrow_rounded, color: Color(0xFF10B981), size: 24),
            ),
            const SizedBox(width: 12),
            Text(
              'Start Cashier Session',
              style: GoogleFonts.inter(
                color: Colors.white,
                fontWeight: FontWeight.w700,
                fontSize: 16,
              ),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Starting a session registers your shift and unlocks the POS Cart for taking orders.',
              style: GoogleFonts.inter(color: Colors.white70, fontSize: 13),
            ),
            const SizedBox(height: 16),
            Text(
              'Opening Cash Float / Balance (SAR)',
              style: GoogleFonts.inter(color: const Color(0xFF38BDF8), fontSize: 12, fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 6),
            Container(
              decoration: BoxDecoration(
                color: const Color(0xFF0F172A),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: Colors.white12),
              ),
              child: TextField(
                controller: floatController,
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
                style: GoogleFonts.inter(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w600),
                decoration: const InputDecoration(
                  prefixIcon: Icon(Icons.account_balance_wallet_outlined, color: Colors.white54, size: 20),
                  border: InputBorder.none,
                  contentPadding: EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                ),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: Text('Cancel', style: GoogleFonts.inter(color: Colors.white54)),
          ),
          ElevatedButton.icon(
            onPressed: () {
              Navigator.of(ctx).pop();
              final floatVal = double.tryParse(floatController.text.trim()) ?? 0.0;
              _startCashierSession(openingBalance: floatVal);
            },
            icon: const Icon(Icons.play_arrow_rounded, size: 18),
            label: const Text('Start Shift'),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF10B981),
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
            ),
          ),
        ],
      ),
    );
  }

  /// Dialog to confirm closing the active session with live expected balance and variance calculation
  Future<void> _showStopSessionDialog({VoidCallback? onSessionClosed}) async {
    if (_activeSessionId == null) return;
    final sessionId = _activeSessionId!;
    final db = DatabaseService().database;

    // 1. Fetch opening balance for the session
    final sessRows = await db.query('cashier_sessions', where: 'id = ?', whereArgs: [sessionId]);
    double openingBalance = 0.0;
    if (sessRows.isNotEmpty && sessRows.first['opening_balance'] != null) {
      openingBalance = (sessRows.first['opening_balance'] as num).toDouble();
    }

    // 2. Fetch sales summary for this session (Cash only vs Non-cash)
    final salesSummary = await db.rawQuery('''
      SELECT 
        COALESCE(SUM(
          CASE 
            WHEN LOWER(COALESCE(p.payment_method, '')) = 'cash' THEN (t.total_amount)
            ELSE 0 
          END
        ), 0) AS cash_sales,
        COALESCE(SUM(
          CASE 
            WHEN LOWER(COALESCE(p.payment_method, '')) != 'cash' AND p.payment_method IS NOT NULL THEN (t.total_amount)
            ELSE 0 
          END
        ), 0) AS non_cash_sales,
        COUNT(t.id) AS transaction_count
      FROM pos_transactions t
      LEFT JOIN pos_payments p ON p.transaction_id = t.id
      WHERE t.cashier_session_id = ? AND t.status = 'completed'
    ''', [sessionId]);

    double cashSales = 0.0;
    double nonCashSales = 0.0;
    int transactionCount = 0;
    if (salesSummary.isNotEmpty) {
      cashSales = (salesSummary.first['cash_sales'] as num?)?.toDouble() ?? 0.0;
      nonCashSales = (salesSummary.first['non_cash_sales'] as num?)?.toDouble() ?? 0.0;
      transactionCount = (salesSummary.first['transaction_count'] as num?)?.toInt() ?? 0;
    }

    // Expected Balance = Opening Balance + Net Cash Sales
    final double expectedBalance = openingBalance + cashSales;

    if (!mounted) return;

    final closeBalanceController = TextEditingController(text: expectedBalance.toStringAsFixed(2));
    final noteController = TextEditingController(text: 'Shift closed successfully');

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (context, setDialogState) {
          final enteredCloseVal = double.tryParse(closeBalanceController.text.trim()) ?? 0.0;
          final currentVariance = enteredCloseVal - expectedBalance;

          Color varianceBg;
          Color varianceBorder;
          Color varianceText;
          String varianceLabel;
          IconData varianceIcon;

          if (currentVariance.abs() < 0.005) {
            varianceBg = const Color(0xFF10B981).withValues(alpha: 0.15);
            varianceBorder = const Color(0xFF10B981);
            varianceText = const Color(0xFF34D399);
            varianceLabel = 'PERFECT MATCH (0.00)';
            varianceIcon = Icons.check_circle_outline_rounded;
          } else if (currentVariance > 0) {
            varianceBg = const Color(0xFFF59E0B).withValues(alpha: 0.15);
            varianceBorder = const Color(0xFFF59E0B);
            varianceText = const Color(0xFFFBBF24);
            varianceLabel = 'OVERAGE (+SAR ${currentVariance.toStringAsFixed(2)})';
            varianceIcon = Icons.arrow_circle_up_rounded;
          } else {
            varianceBg = const Color(0xFFEF4444).withValues(alpha: 0.15);
            varianceBorder = const Color(0xFFEF4444);
            varianceText = const Color(0xFFF87171);
            varianceLabel = 'SHORTAGE (-SAR ${currentVariance.abs().toStringAsFixed(2)})';
            varianceIcon = Icons.arrow_circle_down_rounded;
          }

          return AlertDialog(
            backgroundColor: const Color(0xFF131B2E),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(20),
              side: BorderSide(color: const Color(0xFFEF4444).withValues(alpha: 0.4)),
            ),
            title: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: const Color(0xFFEF4444).withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(Icons.stop_rounded, color: Color(0xFFEF4444), size: 24),
                ),
                const SizedBox(width: 12),
                Text(
                  'Close Cashier Session',
                  style: GoogleFonts.inter(
                    color: Colors.white,
                    fontWeight: FontWeight.w700,
                    fontSize: 16,
                  ),
                ),
              ],
            ),
            content: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Session: ${_activeSessionNumber ?? 'SES-$sessionId'} ($transactionCount transactions)',
                    style: GoogleFonts.inter(color: Colors.white70, fontSize: 13),
                  ),
                  const SizedBox(height: 14),

                  // Session Financial Breakdown Card
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFF0F172A),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.white12),
                    ),
                    child: Column(
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text('Opening Balance:', style: GoogleFonts.inter(color: Colors.white60, fontSize: 12)),
                            Text('SAR ${openingBalance.toStringAsFixed(2)}', style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 12)),
                          ],
                        ),
                        const SizedBox(height: 6),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Row(
                              children: [
                                Text('Cash Sales:', style: GoogleFonts.inter(color: Colors.white60, fontSize: 12)),
                                const SizedBox(width: 4),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                                  decoration: BoxDecoration(color: const Color(0xFF10B981).withValues(alpha: 0.2), borderRadius: BorderRadius.circular(4)),
                                  child: Text('Cash Only', style: GoogleFonts.inter(color: const Color(0xFF34D399), fontSize: 9, fontWeight: FontWeight.bold)),
                                ),
                              ],
                            ),
                            Text('+SAR ${cashSales.toStringAsFixed(2)}', style: GoogleFonts.inter(color: const Color(0xFF34D399), fontWeight: FontWeight.w700, fontSize: 12)),
                          ],
                        ),
                        if (nonCashSales > 0) ...[
                          const SizedBox(height: 6),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text('Bank / Card Sales:', style: GoogleFonts.inter(color: Colors.white38, fontSize: 11)),
                              Text('SAR ${nonCashSales.toStringAsFixed(2)} (not in drawer)', style: GoogleFonts.inter(color: Colors.white38, fontSize: 11)),
                            ],
                          ),
                        ],
                        const Divider(color: Colors.white12, height: 16),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text('Expected Balance:', style: GoogleFonts.inter(color: const Color(0xFF38BDF8), fontWeight: FontWeight.w700, fontSize: 13)),
                            Text('SAR ${expectedBalance.toStringAsFixed(2)}', style: GoogleFonts.inter(color: const Color(0xFF38BDF8), fontWeight: FontWeight.w800, fontSize: 14)),
                          ],
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 14),
                  Text(
                    'Closing Cash Count / Physical Drawer Balance (SAR)',
                    style: GoogleFonts.inter(color: const Color(0xFF38BDF8), fontSize: 12, fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 6),
                  Container(
                    decoration: BoxDecoration(
                      color: const Color(0xFF0F172A),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: Colors.white24),
                    ),
                    child: TextField(
                      controller: closeBalanceController,
                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                      style: GoogleFonts.inter(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w600),
                      onChanged: (_) => setDialogState(() {}),
                      decoration: const InputDecoration(
                        prefixIcon: Icon(Icons.point_of_sale_rounded, color: Colors.white54, size: 20),
                        border: InputBorder.none,
                        contentPadding: EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                      ),
                    ),
                  ),

                  const SizedBox(height: 10),
                  // Live Variance Banner
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    decoration: BoxDecoration(
                      color: varianceBg,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: varianceBorder.withValues(alpha: 0.5)),
                    ),
                    child: Row(
                      children: [
                        Icon(varianceIcon, color: varianceText, size: 18),
                        const SizedBox(width: 8),
                        Text(
                          'Variance: ',
                          style: GoogleFonts.inter(color: Colors.white70, fontSize: 12, fontWeight: FontWeight.w500),
                        ),
                        Expanded(
                          child: Text(
                            varianceLabel,
                            style: GoogleFonts.inter(color: varianceText, fontSize: 12, fontWeight: FontWeight.w800),
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 12),
                  Text(
                    'Closing Notes / Handover',
                    style: GoogleFonts.inter(color: Colors.white60, fontSize: 12),
                  ),
                  const SizedBox(height: 6),
                  Container(
                    decoration: BoxDecoration(
                      color: const Color(0xFF0F172A),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: Colors.white12),
                    ),
                    child: TextField(
                      controller: noteController,
                      style: GoogleFonts.inter(color: Colors.white, fontSize: 13),
                      decoration: const InputDecoration(
                        border: InputBorder.none,
                        contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: Text('Cancel', style: GoogleFonts.inter(color: Colors.white54)),
              ),
              ElevatedButton.icon(
                onPressed: () {
                  Navigator.of(ctx).pop();
                  final closeVal = double.tryParse(closeBalanceController.text.trim()) ?? 0.0;
                  final varianceVal = closeVal - expectedBalance;
                  _stopCashierSession(
                    closingBalance: closeVal,
                    expectedBalance: expectedBalance,
                    variance: varianceVal,
                    closingNote: noteController.text.trim(),
                    onSessionClosed: onSessionClosed,
                  );
                },
                icon: const Icon(Icons.stop_rounded, size: 18),
                label: const Text('Close Shift'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFEF4444),
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  /// Intercepts sign out to ensure any active cashier session is closed before leaving.
  Future<void> _handleSignOutOrLock() async {
    if (_activeSessionId != null && _activeSessionStatus == 'open') {
      final shouldProceed = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          backgroundColor: const Color(0xFF131B2E),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
            side: BorderSide(color: const Color(0xFFF59E0B).withValues(alpha: 0.4)),
          ),
          title: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFF59E0B).withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: const Icon(Icons.warning_amber_rounded, color: Color(0xFFF59E0B), size: 24),
              ),
              const SizedBox(width: 12),
              Text(
                'Active Cashier Shift',
                style: GoogleFonts.inter(
                  color: Colors.white,
                  fontWeight: FontWeight.w700,
                  fontSize: 16,
                ),
              ),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'You have an active shift open (${_activeSessionNumber ?? 'SES-$_activeSessionId'}).',
                style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 13),
              ),
              const SizedBox(height: 8),
              Text(
                'Please perform your cash count and close your shift before logging out.',
                style: GoogleFonts.inter(color: Colors.white70, fontSize: 12),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(false),
              child: Text('Cancel', style: GoogleFonts.inter(color: Colors.white54)),
            ),
            ElevatedButton.icon(
              onPressed: () => Navigator.of(ctx).pop(true),
              icon: const Icon(Icons.stop_circle_outlined, size: 18),
              label: const Text('Close Shift & Sign Out'),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFFEF4444),
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
              ),
            ),
          ],
        ),
      );

      if (shouldProceed == true) {
        await _showStopSessionDialog(onSessionClosed: () {
          _navigateToLogin();
        });
      }
      return;
    }

    _navigateToLogin();
  }

  void _navigateToLogin() {
    if (!mounted) return;
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(
        builder: (_) => LoginScreen(
          tenantSlug: widget.tenantSlug,
          tenantName: widget.tenantName,
          storeId: widget.storeId,
          storeName: widget.storeName,
          posTerminalId: widget.posTerminalId,
          posTerminalName: widget.posTerminalName,
        ),
      ),
      (route) => false,
    );
  }

  /// Fetches local sync_queue statistics via Go SyncService
  Future<void> _fetchSyncQueueStatus() async {
    try {
      final resp = NembusBridge().callHandler(
        handler: 'sync',
        action: 'getSyncStatus',
        payload: {},
      );
      if (resp['success'] == true && resp['data'] is Map) {
        final data = Map<String, dynamic>.from(resp['data'] as Map);
        if (mounted) {
          setState(() {
            _pendingOutboxCount = (data['pending_count'] as num?)?.toInt() ?? 0;
            _syncedOutboxCount = (data['synced_count'] as num?)?.toInt() ?? 0;
            _failedOutboxCount = (data['failed_count'] as num?)?.toInt() ?? 0;
            _lastSyncedAt = data['last_synced_at']?.toString();
          });
        }
      }
    } catch (e) {
      developer.log('Error fetching sync queue status: $e', name: 'PosDashboardScreen');
    }
  }

  /// Displays a modal loading overlay during sync operations to prevent race conditions and inform the user
  void _showSyncLoadingDialog({
    required String title,
    required String message,
    required Color accentColor,
    required IconData icon,
    required bool isUpload,
  }) {
    showDialog(
      context: context,
      barrierDismissible: false,
      barrierColor: Colors.black.withValues(alpha: 0.75),
      builder: (ctx) => PopScope(
        canPop: false,
        child: Dialog(
          backgroundColor: const Color(0xFF131B2E),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
            side: BorderSide(color: accentColor.withValues(alpha: 0.4)),
          ),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 28),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Stack(
                  alignment: Alignment.center,
                  children: [
                    Container(
                      width: 80,
                      height: 80,
                      decoration: BoxDecoration(
                        color: accentColor.withValues(alpha: 0.12),
                        shape: BoxShape.circle,
                      ),
                    ),
                    SizedBox(
                      width: 76,
                      height: 76,
                      child: CircularProgressIndicator(
                        strokeWidth: 3.0,
                        valueColor: AlwaysStoppedAnimation<Color>(
                          accentColor.withValues(alpha: 0.7),
                        ),
                      ),
                    ),
                    _AnimatedSyncIcon(
                      icon: icon,
                      color: accentColor,
                      isUpload: isUpload,
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                Text(
                  title,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.inter(
                    color: Colors.white,
                    fontWeight: FontWeight.w700,
                    fontSize: 16,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  message,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.inter(
                    color: Colors.white70,
                    fontSize: 12,
                    height: 1.4,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  /// Uploads offline local records from sync_queue to Cloud Server via gRPC (Push only)
  Future<void> _uploadOfflineOutboxToCloud() async {
    if (_isSyncingOutbox) return;

    setState(() {
      _isSyncingOutbox = true;
      _errorMessage = null;
    });

    _showSyncLoadingDialog(
      title: 'Uploading Offline Outbox',
      message: 'Sending pending transactions and offline changes to cloud server. Please wait...',
      accentColor: const Color(0xFF10B981),
      icon: Icons.cloud_upload_rounded,
      isUpload: true,
    );

    // Allow dialog to mount and render on UI thread
    await Future.delayed(const Duration(milliseconds: 150));

    final tenantSlug = widget.tenantSlug.isNotEmpty
        ? widget.tenantSlug
        : (SingletonClass().activeTenantSlug ?? 'qitaf');
    final grpcAddr = SingletonClass().grpcAddress;

    developer.log(
      '🚀 [Outbox Queue Sync] Uploading offline outbox to Cloud Server for tenant "$tenantSlug" via gRPC ($grpcAddr)...',
      name: 'PosDashboardScreen',
    );

    try {
      // Drain offline outbox queue to Cloud Server via gRPC StreamPush (Push only)
      final syncResp = NembusBridge().callHandler(
        handler: 'sync',
        action: 'drainOutbox',
        payload: {
          'tenant_slug': tenantSlug,
          'cloud_url': grpcAddr,
          'store_id': widget.storeId ?? SingletonClass().activeStoreId ?? 1,
        },
      );

      developer.log('Outbox Sync Result: $syncResp', name: 'PosDashboardScreen');

      final pushed = (syncResp['pushed_count'] as num?)?.toInt() ?? 0;
      final bool isSuccess = syncResp['success'] == true;

      await _fetchSyncQueueStatus();

      if (mounted) {
        // Dismiss loading modal dialog
        Navigator.of(context, rootNavigator: true).pop();

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: isSuccess ? const Color(0xFF10B981) : const Color(0xFFEF4444),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            content: Row(
              children: [
                Icon(
                  isSuccess ? Icons.cloud_done_rounded : Icons.cloud_off_rounded,
                  color: Colors.white,
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    isSuccess
                        ? (pushed > 0
                            ? 'Successfully uploaded $pushed offline records to Cloud Server!'
                            : 'All local offline records are up to date.')
                        : 'Upload Notice: ${syncResp['error'] ?? 'Check server connection'}',
                    style: GoogleFonts.inter(fontWeight: FontWeight.w600, fontSize: 13),
                  ),
                ),
              ],
            ),
          ),
        );
      }
    } catch (e) {
      developer.log('Outbox upload error: $e', name: 'PosDashboardScreen');
      if (mounted) {
        // Dismiss loading modal dialog
        Navigator.of(context, rootNavigator: true).pop();

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFEF4444),
            content: Text('Upload Error: $e'),
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isSyncingOutbox = false;
        });
      }
    }
  }

  /// Pulls latest database changes and master data from Cloud Server to Local SQLite (Clone DB Pull)
  Future<void> _pullLatestFromCloudServer() async {
    if (_isPullingCloudData) return;

    setState(() {
      _isPullingCloudData = true;
      _errorMessage = null;
    });

    _showSyncLoadingDialog(
      title: 'Fetching Cloud Data',
      message: 'Pulling latest catalog and database changes from cloud server. Please wait...',
      accentColor: const Color(0xFF38BDF8),
      icon: Icons.cloud_download_rounded,
      isUpload: false,
    );

    // Allow dialog to mount and render on UI thread
    await Future.delayed(const Duration(milliseconds: 150));

    final tenantSlug = widget.tenantSlug.isNotEmpty
        ? widget.tenantSlug
        : (SingletonClass().activeTenantSlug ?? 'qitaf');
    final grpcAddr = SingletonClass().grpcAddress;

    developer.log(
      '📥 [Cloud Pull Sync] Pulling latest database changes for tenant "$tenantSlug" from Cloud Server via gRPC ($grpcAddr)...',
      name: 'PosDashboardScreen',
    );

    try {
      final syncResp = NembusBridge().fetchCompleteTenantMasterData(
        tenantSlug,
        grpcAddr,
      );

      developer.log('Cloud Pull Sync Result: $syncResp', name: 'PosDashboardScreen');

      final bool isSuccess = syncResp['success'] == true;

      // Refresh menus and local sync status
      await _fetchUserAssignedMenus();
      await _fetchSyncQueueStatus();

      if (mounted) {
        // Dismiss loading modal dialog
        Navigator.of(context, rootNavigator: true).pop();

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: isSuccess ? const Color(0xFF10B981) : const Color(0xFFEF4444),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            content: Row(
              children: [
                Icon(
                  isSuccess ? Icons.cloud_done_rounded : Icons.cloud_off_rounded,
                  color: Colors.white,
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    isSuccess
                        ? 'Successfully fetched and synchronized latest data from Cloud Server!'
                        : 'Pull Sync Notice: ${syncResp['error'] ?? 'Check server connection'}',
                    style: GoogleFonts.inter(fontWeight: FontWeight.w600, fontSize: 13),
                  ),
                ),
              ],
            ),
          ),
        );
      }
    } catch (e) {
      developer.log('Cloud pull sync error: $e', name: 'PosDashboardScreen');
      if (mounted) {
        // Dismiss loading modal dialog
        Navigator.of(context, rootNavigator: true).pop();

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFEF4444),
            content: Text('Pull Sync Error: $e'),
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isPullingCloudData = false;
        });
      }
    }
  }

  /// Fetches assigned menus for the authenticated user using Core Go Handlers.
  /// Flow: UI -> Bridge (FFI) -> handler.PermissionHandler.GetUserAccessibleMenus / NavigationHandler -> SQLite
  Future<void> _fetchUserAssignedMenus() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    developer.log(
      '🚀 [RBAC Menu Request] Initiating menu discovery for Target User:'
      '\n • User ID: ${widget.userId}'
      '\n • Username: ${widget.username}'
      '\n • Role: ${widget.roleName}'
      '\n • Tenant: ${widget.tenantName} (${widget.tenantSlug})',
      name: 'PosDashboardScreen',
    );

    // Direct SQLite table audit from Flutter side
    try {
      final db = DatabaseService().database;
      final allUsers = await db.rawQuery('SELECT id, username, email FROM users');
      final allRoles = await db.rawQuery('SELECT id, name, code FROM roles');
      final allUserRoles = await db.rawQuery('SELECT * FROM user_roles');
      final userRolesForThisUser = await db.rawQuery('SELECT * FROM user_roles WHERE user_id = ?', [widget.userId]);
      final allMenus = await db.rawQuery('SELECT id, name, code, is_active, display_order FROM menus');
      final allSubmenus = await db.rawQuery('SELECT id, menu_id, name, code, is_active FROM submenus');
      final allRolePerms = await db.rawQuery('SELECT * FROM role_permissions');
      final allMenuPerms = await db.rawQuery('SELECT * FROM menu_permissions');

      developer.log(
        '\n====================== DIRECT SQLITE AUDIT ======================'
        '\n1. USERS IN DB (${allUsers.length}):\n   $allUsers'
        '\n------------------------------------------------------------'
        '\n2. ROLES IN DB (${allRoles.length}):\n   $allRoles'
        '\n------------------------------------------------------------'
        '\n3. ALL USER_ROLES (${allUserRoles.length}):\n   $allUserRoles'
        '\n   -> User ${widget.userId} assigned roles: $userRolesForThisUser'
        '\n------------------------------------------------------------'
        '\n4. MENUS IN DB (${allMenus.length}):\n   $allMenus'
        '\n------------------------------------------------------------'
        '\n5. SUBMENUS IN DB (${allSubmenus.length}):\n   $allSubmenus'
        '\n------------------------------------------------------------'
        '\n6. ROLE_PERMISSIONS COUNT: ${allRolePerms.length}'
        '\n------------------------------------------------------------'
        '\n7. MENU_PERMISSIONS COUNT: ${allMenuPerms.length}'
        '\n=================================================================',
        name: 'PosDashboardScreen',
      );

      final diag = {
        'total_users': allUsers.length,
        'total_roles': allRoles.length,
        'total_menus': allMenus.length,
        'total_submenus': allSubmenus.length,
        'total_user_roles': allUserRoles.length,
        'total_role_permissions': allRolePerms.length,
        'total_menu_permissions': allMenuPerms.length,
        'user_assigned_roles': userRolesForThisUser,
      };

      if (mounted) {
        setState(() {
          _rbacDiagnostics = diag;
        });
      }

      // If user_roles or menus are empty in local SQLite, automatically sync from cloud
      if (allUserRoles.isEmpty || allMenus.isEmpty) {
        developer.log('⚠️ RBAC tables are empty in local SQLite! Initiating cloud sync...', name: 'PosDashboardScreen');
        final tenantSlug = widget.tenantSlug.isNotEmpty
            ? widget.tenantSlug
            : (SingletonClass().activeTenantSlug ?? 'qitaf');
        final grpcAddr = SingletonClass().grpcAddress;
        final syncRes = NembusBridge().fetchCompleteTenantMasterData(
          tenantSlug,
          grpcAddr,
        );
        developer.log('Auto-sync response: $syncRes', name: 'PosDashboardScreen');
      }
    } catch (dbErr) {
      developer.log('Direct SQLite query error: $dbErr', name: 'PosDashboardScreen');
    }

    try {
      // Step 1: Call Core Permission Handler to get menus permitted for this user
      final response = NembusBridge().callHandler(
        handler: 'permission',
        action: 'getUserAccessibleMenus',
        payload: {
          'user_id': widget.userId,
        },
      );

      developer.log(
        '📋 [PermissionHandler.GetUserAccessibleMenus Response]'
        '\n • Success: ${response['success']}'
        '\n • Data Type: ${response['data']?.runtimeType}'
        '\n • Item Count: ${(response['data'] as List?)?.length ?? 0}'
        '\n • Raw Data: ${response['data']}'
        '\n • Error: ${response['error']}',
        name: 'PosDashboardScreen',
      );

      if (response['success'] == true && response['data'] is List && (response['data'] as List).isNotEmpty) {
        final rawList = response['data'] as List;
        final menus = rawList.map((e) => Map<String, dynamic>.from(e as Map)).toList();

        // Sort menus by display_order
        menus.sort((a, b) {
          final orderA = (a['display_order'] as num?)?.toInt() ?? 0;
          final orderB = (b['display_order'] as num?)?.toInt() ?? 0;
          return orderA.compareTo(orderB);
        });

        if (mounted) {
          setState(() {
            _assignedMenus = menus;
            _isLoading = false;
          });
        }
        return;
      }

      // Step 2: Alternative Navigation Handler fallback if permission query was empty
      final navResponse = NembusBridge().callHandler(
        handler: 'navigation',
        action: 'getUserNavigation',
        payload: {
          'user_id': widget.userId,
        },
      );

      developer.log(
        '🧭 [NavigationHandler.GetUserNavigation Response]'
        '\n • Success: ${navResponse['success']}'
        '\n • Raw Data: ${navResponse['data']}'
        '\n • Error: ${navResponse['error']}',
        name: 'PosDashboardScreen',
      );

      if (navResponse['success'] == true && navResponse['data'] is List) {
        final navList = navResponse['data'] as List;
        final List<Map<String, dynamic>> extractedMenus = [];
        for (final item in navList) {
          if (item is Map) {
            final submenus = item['menus'] ?? item['submenus'];
            if (submenus is List) {
              for (final m in submenus) {
                if (m is Map) extractedMenus.add(Map<String, dynamic>.from(m));
              }
            } else {
              extractedMenus.add(Map<String, dynamic>.from(item));
            }
          }
        }

        if (mounted) {
          setState(() {
            _assignedMenus = extractedMenus;
            _isLoading = false;
          });
        }
        return;
      }

      // Fallback or empty result
      if (mounted) {
        setState(() {
          _assignedMenus = [];
          _isLoading = false;
          if (response['error'] != null) {
            _errorMessage = response['error'].toString();
          }
        });
      }
    } catch (e, st) {
      developer.log(
        '❌ [PosDashboardScreen Error] Failed to fetch user menus: $e\n$st',
        name: 'PosDashboardScreen',
      );
      if (mounted) {
        setState(() {
          _errorMessage = 'Failed to load assigned menus: $e';
          _isLoading = false;
        });
      }
    }
  }

  /// Opens the Submenu Bottom Sheet by fetching submenus for menu_id from Go SubmenuHandler
  void _openSubmenus(Map<String, dynamic> menu, Color accentColor) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) => _SubmenuBottomSheet(
        menu: menu,
        accentColor: accentColor,
        userId: widget.userId,
        username: widget.username,
        storeId: widget.storeId,
        storeName: widget.storeName,
        posTerminalId: widget.posTerminalId,
        posTerminalName: widget.posTerminalName,
      ),
    );
  }

  IconData _resolveMenuIcon(String? iconKey, String? code, String name) {
    final key = (iconKey ?? code ?? name).toLowerCase();
    if (key.contains('cart') || key.contains('checkout') || key.contains('sale') || key.contains('pos')) {
      return Icons.point_of_sale_rounded;
    }
    if (key.contains('receipt') || key.contains('order') || key.contains('invoice')) {
      return Icons.receipt_long_rounded;
    }
    if (key.contains('customer') || key.contains('people') || key.contains('client') || key.contains('loyalty')) {
      return Icons.people_alt_rounded;
    }
    if (key.contains('inventory') || key.contains('stock') || key.contains('product') || key.contains('item')) {
      return Icons.inventory_2_rounded;
    }
    if (key.contains('shift') || key.contains('drawer') || key.contains('cash') || key.contains('wallet')) {
      return Icons.account_balance_wallet_rounded;
    }
    if (key.contains('report') || key.contains('analytic') || key.contains('chart')) {
      return Icons.bar_chart_rounded;
    }
    if (key.contains('setting') || key.contains('config') || key.contains('terminal')) {
      return Icons.settings_suggest_rounded;
    }
    return Icons.widgets_rounded;
  }

  List<Color> _resolveCardGradient(int index) {
    final palettes = [
      [const Color(0xFF0284C7), const Color(0xFF0369A1)], // Sky Blue
      [const Color(0xFF6366F1), const Color(0xFF4F46E5)], // Indigo
      [const Color(0xFF10B981), const Color(0xFF059669)], // Emerald
      [const Color(0xFFF59E0B), const Color(0xFFD97706)], // Amber
      [const Color(0xFF8B5CF6), const Color(0xFF7C3AED)], // Purple
      [const Color(0xFFEC4899), const Color(0xFFDB2777)], // Pink
      [const Color(0xFF14B8A6), const Color(0xFF0D9488)], // Teal
    ];
    return palettes[index % palettes.length];
  }

  @override
  Widget build(BuildContext context) {
    const bgColor = Color(0xFF090D16);
    const cardBgColor = Color(0xFF131B2E);
    const surfaceColor = Color(0xFF1E293B);
    const primarySky = Color(0xFF38BDF8);

    final filteredMenus = _assignedMenus.where((menu) {
      final name = (menu['name'] ?? '').toString().toLowerCase();
      final code = (menu['code'] ?? '').toString().toLowerCase();
      final query = _searchQuery.toLowerCase();
      return name.contains(query) || code.contains(query);
    }).toList();

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        _handleSignOutOrLock();
      },
      child: Scaffold(
        backgroundColor: bgColor,
        appBar: AppBar(
          backgroundColor: cardBgColor,
          elevation: 0,
          toolbarHeight: 70,
          automaticallyImplyLeading: false,
          leading: null,
          title: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Flexible(
                    child: Text(
                      widget.tenantName,
                      style: GoogleFonts.inter(
                        fontWeight: FontWeight.w700,
                        fontSize: 16,
                        color: Colors.white,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  const SizedBox(width: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: primarySky.withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(color: primarySky.withValues(alpha: 0.4)),
                    ),
                    child: Text(
                      widget.tenantSlug.toUpperCase(),
                      style: GoogleFonts.inter(
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                        color: primarySky,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 2),
              Row(
                children: [
                  Container(
                    width: 7,
                    height: 7,
                    decoration: const BoxDecoration(
                      color: Color(0xFF10B981),
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      '${widget.username} • ${widget.roleName}',
                      style: GoogleFonts.inter(
                        fontSize: 11,
                        color: Colors.white60,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ],
          ),
          actions: [
            // Push Sync Queue Button
            Stack(
              alignment: Alignment.center,
              children: [
                IconButton(
                  tooltip: 'Upload Offline Outbox to Cloud (Push Sync)',
                  icon: _isSyncingOutbox
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Color(0xFF10B981),
                          ),
                        )
                      : const Icon(Icons.cloud_upload_rounded, color: Color(0xFF10B981)),
                  onPressed: _uploadOfflineOutboxToCloud,
                ),
                if (_pendingOutboxCount > 0)
                  Positioned(
                    top: 8,
                    right: 8,
                    child: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: const BoxDecoration(
                        color: Color(0xFFF59E0B),
                        shape: BoxShape.circle,
                      ),
                      constraints: const BoxConstraints(
                        minWidth: 16,
                        minHeight: 16,
                      ),
                      child: Text(
                        '$_pendingOutboxCount',
                        style: GoogleFonts.inter(
                          color: Colors.black,
                          fontSize: 9,
                          fontWeight: FontWeight.w900,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ),
                  ),
              ],
            ),
            // Pull Sync / Clone DB Button
            IconButton(
              tooltip: 'Pull Latest Changes from Cloud (Clone DB)',
              icon: _isPullingCloudData
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Color(0xFF38BDF8),
                      ),
                    )
                  : const Icon(Icons.cloud_download_rounded, color: Color(0xFF38BDF8)),
              onPressed: _pullLatestFromCloudServer,
            ),
            // Refresh Menus Button
            IconButton(
              tooltip: 'Refresh Assigned Menus',
              icon: const Icon(Icons.refresh_rounded, color: primarySky),
              onPressed: () async {
                await _fetchUserAssignedMenus();
                await _fetchSyncQueueStatus();
              },
            ),
            // Lock Terminal Button
            IconButton(
              tooltip: 'Lock Terminal / Sign Out',
              icon: const Icon(Icons.lock_outline_rounded, color: Colors.white70),
              onPressed: _handleSignOutOrLock,
            ),
            const SizedBox(width: 8),
          ],
        ),
      body: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            final screenWidth = constraints.maxWidth;
            final isTablet = screenWidth >= 600;
            final hPadding = isTablet ? 24.0 : 16.0;

            int crossAxisCount = 2;
            double childAspectRatio = 1.32;
            if (screenWidth >= 1100) {
              crossAxisCount = 4;
              childAspectRatio = 1.5;
            } else if (screenWidth >= 750) {
              crossAxisCount = 3;
              childAspectRatio = 1.42;
            } else if (screenWidth >= 500) {
              crossAxisCount = 2;
              childAspectRatio = 1.35;
            } else {
              crossAxisCount = 2;
              childAspectRatio = 1.25;
            }

            return Padding(
              padding: EdgeInsets.symmetric(horizontal: hPadding, vertical: 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 3D Cashier Session Hero Card (Play / Stop Controls)
                  Container(
                    margin: const EdgeInsets.only(bottom: 12),
                    padding: EdgeInsets.all(isTablet ? 16 : 14),
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: _activeSessionStatus == 'open'
                            ? [
                                const Color(0xFF064E3B).withValues(alpha: 0.7),
                                const Color(0xFF0F172A),
                              ]
                            : [
                                const Color(0xFF1E293B),
                                const Color(0xFF0F172A),
                              ],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: _activeSessionStatus == 'open'
                            ? const Color(0xFF10B981).withValues(alpha: 0.5)
                            : Colors.white.withValues(alpha: 0.1),
                        width: 1.2,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: _activeSessionStatus == 'open'
                              ? const Color(0xFF10B981).withValues(alpha: 0.15)
                              : Colors.black.withValues(alpha: 0.3),
                          blurRadius: 12,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: Row(
                      children: [
                        // Status Icon Avatar
                        Container(
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            color: _activeSessionStatus == 'open'
                                ? const Color(0xFF10B981).withValues(alpha: 0.2)
                                : Colors.white.withValues(alpha: 0.06),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: _activeSessionStatus == 'open'
                                  ? const Color(0xFF10B981)
                                  : Colors.white24,
                            ),
                          ),
                          child: Icon(
                            _activeSessionStatus == 'open'
                                ? Icons.point_of_sale_rounded
                                : Icons.lock_clock_rounded,
                            color: _activeSessionStatus == 'open'
                                ? const Color(0xFF10B981)
                                : Colors.white54,
                            size: isTablet ? 24 : 22,
                          ),
                        ),
                        const SizedBox(width: 12),
                        // Status Details
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Container(
                                    width: 8,
                                    height: 8,
                                    decoration: BoxDecoration(
                                      color: _activeSessionStatus == 'open'
                                          ? const Color(0xFF10B981)
                                          : const Color(0xFFF59E0B),
                                      shape: BoxShape.circle,
                                    ),
                                  ),
                                  const SizedBox(width: 6),
                                  Text(
                                    _activeSessionStatus == 'open'
                                        ? 'CASHIER SESSION ACTIVE'
                                        : 'SESSION CLOSED',
                                    style: GoogleFonts.inter(
                                      fontSize: isTablet ? 12 : 11,
                                      fontWeight: FontWeight.w800,
                                      letterSpacing: 0.8,
                                      color: _activeSessionStatus == 'open'
                                          ? const Color(0xFF34D399)
                                          : const Color(0xFFF59E0B),
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 3),
                              Text(
                                _activeSessionStatus == 'open'
                                    ? (_activeSessionNumber ?? 'Session In Progress')
                                    : 'Start session to unlock Cart & make sales',
                                style: GoogleFonts.inter(
                                  fontSize: isTablet ? 14 : 13,
                                  fontWeight: FontWeight.w600,
                                  color: Colors.white,
                                ),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                              if (_activeSessionStatus == 'open' && _sessionOpenedAt != null)
                                Padding(
                                  padding: const EdgeInsets.only(top: 2),
                                  child: Text(
                                    'Terminal: ${widget.posTerminalName ?? SingletonClass().activeTerminalName ?? 'POS-1'} • Started: ${_sessionOpenedAt!.hour.toString().padLeft(2, '0')}:${_sessionOpenedAt!.minute.toString().padLeft(2, '0')}',
                                    style: GoogleFonts.inter(
                                      fontSize: isTablet ? 11.5 : 10.5,
                                      color: Colors.white60,
                                    ),
                                  ),
                                ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 10),
                        // Action Button (Play / Stop)
                        _isSessionLoading
                            ? const SizedBox(
                                width: 32,
                                height: 32,
                                child: CircularProgressIndicator(strokeWidth: 2, color: primarySky),
                              )
                            : _activeSessionStatus == 'open'
                                ? ElevatedButton.icon(
                                    onPressed: _showStopSessionDialog,
                                    icon: const Icon(Icons.stop_rounded, size: 16),
                                    label: const Text('Stop'),
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: const Color(0xFFEF4444),
                                      foregroundColor: Colors.white,
                                      padding: EdgeInsets.symmetric(horizontal: isTablet ? 18 : 14, vertical: isTablet ? 12 : 10),
                                      shape: RoundedRectangleBorder(
                                        borderRadius: BorderRadius.circular(10),
                                      ),
                                      elevation: 2,
                                    ),
                                  )
                                : ElevatedButton.icon(
                                    onPressed: _showStartSessionDialog,
                                    icon: const Icon(Icons.play_arrow_rounded, size: 16),
                                    label: const Text('Play'),
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: const Color(0xFF10B981),
                                      foregroundColor: Colors.white,
                                      padding: EdgeInsets.symmetric(horizontal: isTablet ? 18 : 14, vertical: isTablet ? 12 : 10),
                                      shape: RoundedRectangleBorder(
                                        borderRadius: BorderRadius.circular(10),
                                      ),
                                      elevation: 3,
                                    ),
                                  ),
                      ],
                    ),
                  ),

                  // Search & Filter Header
                  Row(
                    children: [
                      Expanded(
                        child: Container(
                          height: 46,
                          decoration: BoxDecoration(
                            color: surfaceColor,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.08),
                            ),
                          ),
                          child: TextField(
                            onChanged: (val) => setState(() => _searchQuery = val),
                            style: GoogleFonts.inter(color: Colors.white, fontSize: 14),
                            decoration: InputDecoration(
                              hintText: 'Search authorized menus & tools...',
                              hintStyle: GoogleFonts.inter(
                                color: Colors.white38,
                                fontSize: 13,
                              ),
                              prefixIcon: const Icon(
                                Icons.search_rounded,
                                color: Colors.white54,
                                size: 20,
                              ),
                              border: InputBorder.none,
                              contentPadding: const EdgeInsets.symmetric(
                                horizontal: 14,
                                vertical: 12,
                              ),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      Container(
                        height: 46,
                        padding: const EdgeInsets.symmetric(horizontal: 14),
                        decoration: BoxDecoration(
                          color: primarySky.withValues(alpha: 0.12),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: primarySky.withValues(alpha: 0.3)),
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.verified_user_rounded,
                                size: 16, color: primarySky),
                            const SizedBox(width: 6),
                            Text(
                              '${_assignedMenus.length} Authorized',
                              style: GoogleFonts.inter(
                                fontSize: 12,
                                fontWeight: FontWeight.w700,
                                color: primarySky,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Title Section
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          'OPERATIONAL MENU & MODULES',
                          overflow: TextOverflow.ellipsis,
                          style: GoogleFonts.inter(
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                            letterSpacing: 1.1,
                            color: Colors.white54,
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        'Tap to Open Submenu',
                        style: GoogleFonts.inter(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          color: const Color(0xFF10B981),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),

                  // Content Area
                  Expanded(
                    child: _isLoading
                        ? const Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                CircularProgressIndicator(
                                  valueColor: AlwaysStoppedAnimation<Color>(
                                    Color(0xFF38BDF8),
                                  ),
                                ),
                                SizedBox(height: 16),
                                Text(
                                  'Resolving permissions from SQLite...',
                                  style: TextStyle(
                                    color: Colors.white60,
                                    fontSize: 13,
                                  ),
                                ),
                              ],
                            ),
                          )
                        : _errorMessage != null && _assignedMenus.isEmpty
                            ? Center(
                                child: Padding(
                                  padding: const EdgeInsets.all(24.0),
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      const Icon(
                                        Icons.error_outline_rounded,
                                        color: Color(0xFFEF4444),
                                        size: 48,
                                      ),
                                      const SizedBox(height: 16),
                                      Text(
                                        _errorMessage!,
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.inter(
                                          color: Colors.white70,
                                          fontSize: 14,
                                        ),
                                      ),
                                      const SizedBox(height: 16),
                                      ElevatedButton.icon(
                                        onPressed: _fetchUserAssignedMenus,
                                        icon: const Icon(Icons.refresh),
                                        label: const Text('Retry Menu Query'),
                                        style: ElevatedButton.styleFrom(
                                          backgroundColor: primarySky,
                                          foregroundColor: Colors.black87,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              )
                            : filteredMenus.isEmpty
                                ? Center(
                                    child: Column(
                                      mainAxisAlignment: MainAxisAlignment.center,
                                      children: [
                                        const Icon(
                                          Icons.lock_clock_rounded,
                                          color: Colors.white30,
                                          size: 54,
                                        ),
                                        const SizedBox(height: 14),
                                        Text(
                                          'No Menus Authorized for "${widget.roleName}"',
                                          style: GoogleFonts.inter(
                                            color: Colors.white,
                                            fontWeight: FontWeight.w600,
                                            fontSize: 16,
                                          ),
                                        ),
                                        const SizedBox(height: 6),
                                        Text(
                                          'Tables or role permissions may be missing in local SQLite.',
                                          textAlign: TextAlign.center,
                                          style: GoogleFonts.inter(
                                            color: Colors.white54,
                                            fontSize: 13,
                                          ),
                                        ),
                                        const SizedBox(height: 16),
                                        ElevatedButton.icon(
                                          onPressed: _fetchUserAssignedMenus,
                                          icon: const Icon(Icons.refresh_rounded),
                                          label: const Text('Refresh Menus'),
                                          style: ElevatedButton.styleFrom(
                                            backgroundColor: primarySky,
                                            foregroundColor: Colors.white,
                                            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                                            shape: RoundedRectangleBorder(
                                              borderRadius: BorderRadius.circular(10),
                                            ),
                                          ),
                                        ),
                                      ],
                                    ),
                                  )
                                : GridView.builder(
                                    physics: const BouncingScrollPhysics(),
                                    gridDelegate:
                                        SliverGridDelegateWithFixedCrossAxisCount(
                                      crossAxisCount: crossAxisCount,
                                      crossAxisSpacing: isTablet ? 16 : 12,
                                      mainAxisSpacing: isTablet ? 16 : 12,
                                      childAspectRatio: childAspectRatio,
                                    ),
                                    itemCount: filteredMenus.length,
                                    itemBuilder: (context, index) {
                                      final menu = filteredMenus[index];
                                      final name =
                                          menu['name']?.toString() ?? 'Menu Item';
                                      final code =
                                          menu['code']?.toString() ?? 'menu_code';
                                      final iconKey = menu['icon']?.toString();
                                      final routePath =
                                          menu['route_path']?.toString() ?? '';
                                      final iconData =
                                          _resolveMenuIcon(iconKey, code, name);
                                      final gradientColors =
                                          _resolveCardGradient(index);

                                      return Material(
                                        color: Colors.transparent,
                                        child: InkWell(
                                          borderRadius: BorderRadius.circular(16),
                                          onTap: () => _openSubmenus(menu, gradientColors[0]),
                                          child: Container(
                                            padding: EdgeInsets.all(isTablet ? 16 : 14),
                                            decoration: BoxDecoration(
                                              color: cardBgColor,
                                              borderRadius:
                                                  BorderRadius.circular(16),
                                              border: Border.all(
                                                color: Colors.white
                                                    .withValues(alpha: 0.08),
                                              ),
                                              boxShadow: [
                                                BoxShadow(
                                                  color: Colors.black
                                                      .withValues(alpha: 0.3),
                                                  blurRadius: 10,
                                                  offset: const Offset(0, 4),
                                                ),
                                              ],
                                            ),
                                            child: Column(
                                              crossAxisAlignment:
                                                  CrossAxisAlignment.start,
                                              mainAxisAlignment:
                                                  MainAxisAlignment.spaceBetween,
                                              children: [
                                                Row(
                                                  mainAxisAlignment:
                                                      MainAxisAlignment.spaceBetween,
                                                  children: [
                                                    Container(
                                                      padding:
                                                          const EdgeInsets.all(9),
                                                      decoration: BoxDecoration(
                                                        gradient: LinearGradient(
                                                          colors: gradientColors,
                                                          begin: Alignment.topLeft,
                                                          end: Alignment
                                                              .bottomRight,
                                                        ),
                                                        borderRadius:
                                                            BorderRadius.circular(
                                                                12),
                                                        boxShadow: [
                                                          BoxShadow(
                                                            color: gradientColors[0]
                                                                .withValues(
                                                                    alpha: 0.3),
                                                            blurRadius: 8,
                                                            offset:
                                                                const Offset(0, 2),
                                                          ),
                                                        ],
                                                      ),
                                                      child: Icon(
                                                        iconData,
                                                        color: Colors.white,
                                                        size: 22,
                                                      ),
                                                    ),
                                                    Icon(
                                                      Icons.arrow_forward_ios_rounded,
                                                      size: 13,
                                                      color: primarySky.withValues(alpha: 0.7),
                                                    ),
                                                  ],
                                                ),
                                                Column(
                                                  crossAxisAlignment:
                                                      CrossAxisAlignment.start,
                                                  children: [
                                                    Text(
                                                      name,
                                                      maxLines: 2,
                                                      overflow:
                                                          TextOverflow.ellipsis,
                                                      style: GoogleFonts.inter(
                                                        fontWeight: FontWeight.w700,
                                                        fontSize: isTablet ? 14.5 : 13.5,
                                                        color: Colors.white,
                                                        height: 1.2,
                                                      ),
                                                    ),
                                                    if (routePath.isNotEmpty) ...[
                                                      const SizedBox(height: 2),
                                                      Text(
                                                        routePath,
                                                        style: GoogleFonts.inter(
                                                          fontSize: isTablet ? 11 : 10.5,
                                                          color: primarySky
                                                              .withValues(
                                                                  alpha: 0.8),
                                                          fontWeight:
                                                              FontWeight.w500,
                                                        ),
                                                      ),
                                                    ],
                                                  ],
                                                ),
                                              ],
                                            ),
                                          ),
                                        ),
                                      );
                                    },
                                  ),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    ));
  }
}

/// Submenu Modal Bottom Sheet fetching active submenus by menu_id via Go SubmenuHandler
class _SubmenuBottomSheet extends StatefulWidget {
  final Map<String, dynamic> menu;
  final Color accentColor;
  final int userId;
  final String username;
  final int? storeId;
  final String? storeName;
  final int? posTerminalId;
  final String? posTerminalName;

  const _SubmenuBottomSheet({
    required this.menu,
    required this.accentColor,
    required this.userId,
    this.username = 'Cashier',
    this.storeId,
    this.storeName,
    this.posTerminalId,
    this.posTerminalName,
  });

  @override
  State<_SubmenuBottomSheet> createState() => _SubmenuBottomSheetState();
}

class _SubmenuBottomSheetState extends State<_SubmenuBottomSheet> {
  bool _isLoading = true;
  String? _errorMessage;
  List<Map<String, dynamic>> _submenus = [];

  @override
  void initState() {
    super.initState();
    _fetchSubmenus();
  }

  Future<void> _fetchSubmenus() async {
    final menuId = widget.menu['id'];
    developer.log(
      'Fetching submenus for menu_id=$menuId via Go SubmenuHandler',
      name: 'SubmenuBottomSheet',
    );

    try {
      // Step 1: Call Go SubmenuHandler -> listActiveSubmenusByMenu
      final response = NembusBridge().callHandler(
        handler: 'submenu',
        action: 'listActiveSubmenusByMenu',
        payload: {
          'menu_id': menuId,
        },
      );

      developer.log('SubmenuHandler Response: $response',
          name: 'SubmenuBottomSheet');

      if (response['success'] == true && response['data'] is List) {
        final list = (response['data'] as List)
            .map((e) => Map<String, dynamic>.from(e as Map))
            .toList();

        // Sort by display_order
        list.sort((a, b) {
          final orderA = (a['display_order'] as num?)?.toInt() ?? 0;
          final orderB = (b['display_order'] as num?)?.toInt() ?? 0;
          return orderA.compareTo(orderB);
        });

        if (mounted) {
          setState(() {
            _submenus = list;
            _isLoading = false;
          });
        }
        return;
      }

      // Step 2: Fallback query if active submenus list was empty
      final listResp = NembusBridge().callHandler(
        handler: 'submenu',
        action: 'listSubmenusByMenu',
        payload: {
          'menu_id': menuId,
        },
      );

      if (listResp['success'] == true && listResp['data'] is List) {
        final list = (listResp['data'] as List)
            .map((e) => Map<String, dynamic>.from(e as Map))
            .toList();
        if (mounted) {
          setState(() {
            _submenus = list;
            _isLoading = false;
          });
        }
        return;
      }

      if (mounted) {
        setState(() {
          _submenus = [];
          _isLoading = false;
        });
      }
    } catch (e) {
      developer.log('Error loading submenus: $e', name: 'SubmenuBottomSheet');
      if (mounted) {
        setState(() {
          _errorMessage = 'Failed to load submenus: $e';
          _isLoading = false;
        });
      }
    }
  }

  IconData _resolveSubmenuIcon(String? iconKey, String? code, String name) {
    final key = (iconKey ?? code ?? name).toLowerCase();
    if (key.contains('add') || key.contains('new') || key.contains('create')) {
      return Icons.add_circle_outline_rounded;
    }
    if (key.contains('list') || key.contains('history') || key.contains('records')) {
      return Icons.format_list_bulleted_rounded;
    }
    if (key.contains('return') || key.contains('refund')) {
      return Icons.assignment_return_rounded;
    }
    if (key.contains('cart') || key.contains('hold') || key.contains('park')) {
      return Icons.shopping_basket_rounded;
    }
    if (key.contains('scan') || key.contains('barcode')) {
      return Icons.qr_code_scanner_rounded;
    }
    if (key.contains('close') || key.contains('reconcile') || key.contains('end')) {
      return Icons.lock_clock_rounded;
    }
    return Icons.subdirectory_arrow_right_rounded;
  }

  @override
  Widget build(BuildContext context) {
    const sheetBg = Color(0xFF0F172A);
    const cardBg = Color(0xFF1E293B);
    final menuName = widget.menu['name']?.toString() ?? 'Menu';
    final menuRoute = widget.menu['route_path']?.toString() ?? '';

    return Align(
      alignment: Alignment.bottomCenter,
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: 640,
          maxHeight: MediaQuery.of(context).size.height * 0.75,
        ),
        child: Container(
          decoration: BoxDecoration(
            color: sheetBg,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
            border: Border.all(
              color: widget.accentColor.withValues(alpha: 0.3),
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.6),
                blurRadius: 20,
                offset: const Offset(0, -5),
              ),
            ],
          ),
          child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Drag handle
          Container(
            margin: const EdgeInsets.only(top: 12, bottom: 8),
            width: 44,
            height: 4,
            decoration: BoxDecoration(
              color: Colors.white24,
              borderRadius: BorderRadius.circular(2),
            ),
          ),

          // Header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: widget.accentColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: widget.accentColor.withValues(alpha: 0.4),
                    ),
                  ),
                  child: Icon(
                    Icons.dashboard_customize_rounded,
                    color: widget.accentColor,
                    size: 22,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        menuName,
                        style: GoogleFonts.inter(
                          fontSize: 16,
                          fontWeight: FontWeight.w800,
                          color: Colors.white,
                        ),
                      ),
                      if (menuRoute.isNotEmpty)
                        Text(
                          'Route: $menuRoute',
                          style: GoogleFonts.inter(
                            fontSize: 11,
                            color: widget.accentColor,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                    ],
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close_rounded, color: Colors.white60),
                  onPressed: () => Navigator.of(context).pop(),
                ),
              ],
            ),
          ),
          const Divider(color: Colors.white10, height: 1),

          // Submenu List
          Flexible(
            child: _isLoading
                ? const Padding(
                    padding: EdgeInsets.all(32.0),
                    child: Center(
                      child: CircularProgressIndicator(
                        valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF38BDF8)),
                      ),
                    ),
                  )
                : _errorMessage != null
                    ? Padding(
                        padding: const EdgeInsets.all(24.0),
                        child: Text(
                          _errorMessage!,
                          style: const TextStyle(color: Colors.redAccent),
                        ),
                      )
                    : _submenus.isEmpty
                        ? Padding(
                            padding: const EdgeInsets.all(32.0),
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(
                                  Icons.check_circle_outline_rounded,
                                  size: 48,
                                  color: widget.accentColor,
                                ),
                                const SizedBox(height: 12),
                                Text(
                                  'Direct Action Ready',
                                  style: GoogleFonts.inter(
                                    fontSize: 15,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.white,
                                  ),
                                ),
                                const SizedBox(height: 6),
                                Text(
                                  'This menu does not have sub-items and executes directly.',
                                  textAlign: TextAlign.center,
                                  style: GoogleFonts.inter(
                                    fontSize: 12,
                                    color: Colors.white60,
                                  ),
                                ),
                                const SizedBox(height: 16),
                                ElevatedButton(
                                  onPressed: () {
                                    Navigator.of(context).pop();
                                    final combined = '${widget.menu['name']} ${widget.menu['code']} ${widget.menu['route_path']}'.toLowerCase();
                                    final isStockCount = combined.contains('stock_count') || combined.contains('stock-count') || combined.contains('stock count') || combined.contains('counts') || combined.contains('inventory/counts');
                                    final isWholesale = combined.contains('wholesale') || combined.contains('b2b');
                                    final isRetail = combined.contains('retail') || combined.contains('pos') || combined.contains('sales') || combined.contains('product') || combined.contains('terminal');
                                    final isProductAction = isRetail || isWholesale;
                                    if (isStockCount) {
                                      Navigator.of(context).push(
                                        MaterialPageRoute(
                                          builder: (_) => StockCountListScreen(
                                            userId: widget.userId,
                                            username: widget.username,
                                            initialStoreId: widget.storeId ?? SingletonClass().activeStoreId,
                                            initialStoreName: widget.storeName ?? SingletonClass().activeStoreName,
                                          ),
                                        ),
                                      );
                                    } else if (isProductAction) {
                                      Navigator.of(context).push(
                                        MaterialPageRoute(
                                          builder: (_) => PosProductListScreen(
                                            userId: widget.userId,
                                            username: widget.username,
                                            initialStoreId: widget.storeId ?? SingletonClass().activeStoreId,
                                            initialStoreName: widget.storeName ?? SingletonClass().activeStoreName,
                                            posTerminalId: widget.posTerminalId ?? SingletonClass().activeTerminalId,
                                            posTerminalName: widget.posTerminalName ?? SingletonClass().activeTerminalName,
                                            initialSaleType: isWholesale ? 'wholesale' : 'retail',
                                          ),
                                        ),
                                      );
                                    } else {
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        SnackBar(
                                          backgroundColor: cardBg,
                                          content: Text('Launching $menuName...'),
                                        ),
                                      );
                                    }
                                  },
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: widget.accentColor,
                                    foregroundColor: Colors.black87,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                  ),
                                  child: const Text('Open Terminal View'),
                                ),
                              ],
                            ),
                          )
                        : ListView.separated(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 14,
                            ),
                            shrinkWrap: true,
                            itemCount: _submenus.length,
                            separatorBuilder: (_, __) => const SizedBox(height: 10),
                            itemBuilder: (context, idx) {
                              final sm = _submenus[idx];
                              final smName = sm['name']?.toString() ?? 'Submenu';
                              final smCode = sm['code']?.toString() ?? '';
                              final smRoute = sm['route_path']?.toString() ?? '';
                              final smIcon = _resolveSubmenuIcon(
                                sm['icon']?.toString(),
                                smCode,
                                smName,
                              );

                              final combinedSub = '$smName $smCode $smRoute'.toLowerCase();
                              final isCreateAction = combinedSub.contains('create') || combinedSub.contains('new') || combinedSub.contains('add');
                              final isStockCountSub = !isCreateAction && (combinedSub.contains('stock_count_list') || combinedSub.contains('stock-count-list') || combinedSub.contains('stock count list') || (combinedSub.contains('list') && combinedSub.contains('count')) || combinedSub.contains('/inventory/counts/list'));
                              final isWholesale = combinedSub.contains('wholesale') || combinedSub.contains('b2b');
                              final isRetail = combinedSub.contains('retail') || combinedSub.contains('pos') || combinedSub.contains('sale') || combinedSub.contains('product') || combinedSub.contains('order') || combinedSub.contains('terminal');
                              final isProductSubmenu = isRetail || isWholesale;

                              return Material(
                                color: Colors.transparent,
                                child: InkWell(
                                  borderRadius: BorderRadius.circular(14),
                                  onTap: () {
                                    Navigator.of(context).pop();
                                    if (isStockCountSub) {
                                      Navigator.of(context).push(
                                        MaterialPageRoute(
                                          builder: (_) => StockCountListScreen(
                                            userId: widget.userId,
                                            username: widget.username,
                                            initialStoreId: widget.storeId ?? SingletonClass().activeStoreId,
                                            initialStoreName: widget.storeName ?? SingletonClass().activeStoreName,
                                          ),
                                        ),
                                      );
                                    } else if (isProductSubmenu) {
                                      Navigator.of(context).push(
                                        MaterialPageRoute(
                                          builder: (_) => PosProductListScreen(
                                            userId: widget.userId,
                                            username: widget.username,
                                            initialStoreId: widget.storeId ?? SingletonClass().activeStoreId,
                                            initialStoreName: widget.storeName ?? SingletonClass().activeStoreName,
                                            posTerminalId: widget.posTerminalId ?? SingletonClass().activeTerminalId,
                                            posTerminalName: widget.posTerminalName ?? SingletonClass().activeTerminalName,
                                            initialSaleType: isWholesale ? 'wholesale' : 'retail',
                                          ),
                                        ),
                                      );
                                    } else {
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        SnackBar(
                                          backgroundColor: cardBg,
                                          behavior: SnackBarBehavior.floating,
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(12),
                                            side: BorderSide(
                                              color: widget.accentColor
                                                  .withValues(alpha: 0.5),
                                            ),
                                          ),
                                          content: Row(
                                            children: [
                                              Icon(smIcon, color: widget.accentColor),
                                              const SizedBox(width: 10),
                                              Expanded(
                                                child: Text(
                                                  'Navigating to $smName ($smRoute)',
                                                  style: GoogleFonts.inter(
                                                    fontWeight: FontWeight.w600,
                                                    color: Colors.white,
                                                  ),
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      );
                                    }
                                  },
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 16,
                                      vertical: 14,
                                    ),
                                    decoration: BoxDecoration(
                                      color: cardBg,
                                      borderRadius: BorderRadius.circular(14),
                                      border: Border.all(
                                        color: Colors.white.withValues(alpha: 0.06),
                                      ),
                                    ),
                                    child: Row(
                                      children: [
                                        Container(
                                          padding: const EdgeInsets.all(8),
                                          decoration: BoxDecoration(
                                            color: widget.accentColor
                                                .withValues(alpha: 0.12),
                                            borderRadius: BorderRadius.circular(10),
                                          ),
                                          child: Icon(
                                            smIcon,
                                            size: 20,
                                            color: widget.accentColor,
                                          ),
                                        ),
                                        const SizedBox(width: 14),
                                        Expanded(
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              Text(
                                                smName,
                                                style: GoogleFonts.inter(
                                                  fontSize: 14,
                                                  fontWeight: FontWeight.w600,
                                                  color: Colors.white,
                                                ),
                                              ),
                                              if (smRoute.isNotEmpty) ...[
                                                const SizedBox(height: 2),
                                                Text(
                                                  smRoute,
                                                  style: GoogleFonts.inter(
                                                    fontSize: 11,
                                                    color: Colors.white38,
                                                  ),
                                                ),
                                              ],
                                            ],
                                          ),
                                        ),
                                        Icon(
                                          Icons.chevron_right_rounded,
                                          color: Colors.white38,
                                          size: 20,
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              );
                            },
                          ),
          ),
          const SizedBox(height: 12),
        ],
      ),
    ),
  ),
);
  }
}

class _AnimatedSyncIcon extends StatefulWidget {
  final IconData icon;
  final Color color;
  final bool isUpload;

  const _AnimatedSyncIcon({
    required this.icon,
    required this.color,
    required this.isUpload,
  });

  @override
  State<_AnimatedSyncIcon> createState() => _AnimatedSyncIconState();
}

class _AnimatedSyncIconState extends State<_AnimatedSyncIcon>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _translateAnimation;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat(reverse: true);

    // Negative Y moves upward, positive Y moves downward
    final double offset = widget.isUpload ? -7.0 : 7.0;
    _translateAnimation = Tween<double>(begin: -offset, end: offset).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
    _scaleAnimation = Tween<double>(begin: 0.90, end: 1.10).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return Transform.translate(
          offset: Offset(0, _translateAnimation.value),
          child: Transform.scale(
            scale: _scaleAnimation.value,
            child: Icon(
              widget.icon,
              color: widget.color,
              size: 32,
            ),
          ),
        );
      },
    );
  }
}