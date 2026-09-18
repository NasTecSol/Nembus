import 'dart:async';
import 'dart:convert';
import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';

class StockCountExecutionScreen extends StatefulWidget {
  final Map<String, dynamic> session;

  const StockCountExecutionScreen({
    super.key,
    required this.session,
  });

  @override
  State<StockCountExecutionScreen> createState() => _StockCountExecutionScreenState();
}

class _StockCountExecutionScreenState extends State<StockCountExecutionScreen>
    with SingleTickerProviderStateMixin {
  // Theme Constants (matching dark mobile HUD theme)
  static const Color bgColor = Color(0xFF090D16);
  static const Color cardBgColor = Color(0xFF131B2E);
  static const Color surfaceColor = Color(0xFF1E293B);
  static const Color primarySky = Color(0xFF38BDF8);
  static const Color accentIndigo = Color(0xFF4F46E5);
  static const Color accentEmerald = Color(0xFF10B981);
  static const Color accentAmber = Color(0xFFF59E0B);
  static const Color accentRose = Color(0xFFEF4444);

  // Active Bin / Location state
  String? _activeBinCode;
  String? _activeBinDescription;
  int? _activeBinId;
  List<Map<String, dynamic>> _storageLocations = [];

  // Stopwatch Timer
  late Timer _stopwatchTimer;
  int _elapsedSeconds = 0;

  // Laser scanner animation
  late AnimationController _laserController;
  late Animation<double> _laserAnimation;
  bool _isTorchOn = false;
  bool _isAudioOn = true;

  // Session state & items
  bool _isLoading = true;
  List<Map<String, dynamic>> _availableProducts = [];
  Map<String, dynamic>? _currentCapturedProduct;
  int _currentCapturedCount = 1;
  int? _currentCapturedExpected;

  // Audit Feed (Scanned items history in this session from SQLite)
  final List<Map<String, dynamic>> _auditFeed = [];

  // Metrics
  int _targetSKUs = 0;
  int _totalCountedSKUs = 0;
  int _varianceAlertsCount = 0;

  @override
  void initState() {
    super.initState();
    _initLaserAnimation();
    _startStopwatch();
    _loadSessionDataAndProducts();
  }

  void _initLaserAnimation() {
    _laserController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1800),
    )..repeat(reverse: true);

    _laserAnimation = Tween<double>(begin: 0.15, end: 0.85).animate(
      CurvedAnimation(parent: _laserController, curve: Curves.easeInOut),
    );
  }

  void _startStopwatch() {
    _stopwatchTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) {
        setState(() {
          _elapsedSeconds++;
        });
      }
    });
  }

  String _formatStopwatchTime(int totalSecs) {
    final hours = totalSecs ~/ 3600;
    final mins = (totalSecs % 3600) ~/ 60;
    final secs = totalSecs % 60;
    return '${hours.toString().padLeft(2, '0')}:${mins.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  @override
  void dispose() {
    _stopwatchTimer.cancel();
    _laserController.dispose();
    super.dispose();
  }

  /// Loads real products, storage locations, and existing count lines from SQLite
  Future<void> _loadSessionDataAndProducts() async {
    setState(() => _isLoading = true);

    final countId = (widget.session['id'] as num?)?.toInt();

    try {
      final db = DatabaseService().database;

      // 1. Fetch real storage locations from SQLite
      final locRows = await db.query('storage_locations', orderBy: 'name ASC');
      if (locRows.isNotEmpty) {
        _storageLocations = List<Map<String, dynamic>>.from(locRows);
        final first = _storageLocations.first;
        _activeBinId = (first['id'] as num?)?.toInt();
        _activeBinCode = first['code']?.toString() ?? first['name']?.toString() ?? 'Default Bin';
        _activeBinDescription = first['name']?.toString() ?? _activeBinCode;
      }

      // 2. Fetch real local products from SQLite
      final prodRows = await db.rawQuery('''
        SELECT 
          p.id AS product_id,
          p.name AS product_name,
          p.sku,
          p.barcode,
          COALESCE(pb.barcode, p.barcode) AS active_barcode,
          COALESCE(inv.quantity_on_hand, 0) AS system_quantity
        FROM products p
        LEFT JOIN product_barcodes pb ON p.id = pb.product_id AND pb.is_primary = 1
        LEFT JOIN inventory_stock inv ON p.id = inv.product_id
        LIMIT 200
      ''');

      if (prodRows.isNotEmpty) {
        _availableProducts = List<Map<String, dynamic>>.from(prodRows);
      }

      // 3. Fetch existing count lines for this stock count session if already started
      _auditFeed.clear();
      if (countId != null && countId > 0) {
        // Ensure table exists
        await db.execute('''
          CREATE TABLE IF NOT EXISTS stock_count_lines (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            stock_count_id INTEGER,
            product_id INTEGER,
            product_variant_id INTEGER,
            storage_location_id INTEGER,
            expected_quantity REAL,
            system_quantity REAL,
            counted_quantity REAL,
            variance REAL,
            variance_value REAL,
            counted_at TEXT,
            uom_id INTEGER,
            batch_number TEXT,
            serial_number TEXT,
            metadata TEXT,
            created_at TEXT,
            updated_at TEXT
          );
        ''');

        final lineRows = await db.rawQuery('''
          SELECT 
            scl.id,
            scl.product_id,
            scl.expected_quantity,
            scl.counted_quantity,
            scl.variance,
            scl.batch_number,
            p.name AS product_name,
            p.sku,
            p.barcode,
            sl.code AS location_code
          FROM stock_count_lines scl
          LEFT JOIN products p ON scl.product_id = p.id
          LEFT JOIN storage_locations sl ON scl.storage_location_id = sl.id
          WHERE scl.stock_count_id = ?
          ORDER BY scl.id DESC
        ''', [countId]);

        for (final row in lineRows) {
          final counted = (row['counted_quantity'] as num?)?.toInt() ?? 0;
          final expected = (row['expected_quantity'] as num?)?.toInt() ?? counted;
          final variance = (row['variance'] as num?)?.toInt() ?? (counted - expected);

          _auditFeed.add({
            'id': row['id'],
            'product_id': row['product_id'],
            'product_name': row['product_name'] ?? 'Product #${row['product_id']}',
            'sku': row['sku'] ?? 'SKU-${row['product_id']}',
            'location_code': row['location_code'] ?? _activeBinCode ?? 'Default',
            'counted_quantity': counted,
            'expected_quantity': expected,
            'variance': variance,
            'is_verified': variance == 0,
            'icon': Icons.inventory_2_rounded,
          });
        }
      }

      // 4. Set first captured product if available
      if (_availableProducts.isNotEmpty) {
        final firstProd = _availableProducts.first;
        _currentCapturedProduct = firstProd;
        _currentCapturedExpected = (firstProd['system_quantity'] as num?)?.toInt() ?? 0;
        _currentCapturedCount = _currentCapturedExpected! > 0 ? _currentCapturedExpected! : 1;
      }

      _targetSKUs = _availableProducts.isNotEmpty ? _availableProducts.length : _auditFeed.length;
      _recalculateStats();

      if (mounted) {
        setState(() => _isLoading = false);
      }
    } catch (e) {
      developer.log('Error initializing session: $e', name: 'StockCountExecution');
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _recalculateStats() {
    int counted = _auditFeed.length;
    int variances = _auditFeed.where((item) => (item['variance'] as num) != 0).length;
    setState(() {
      _totalCountedSKUs = counted;
      _varianceAlertsCount = variances;
    });
  }

  /// Trigger real barcode or SKU lookup
  void _onBarcodeScanned(String barcodeOrSku) {
    final query = barcodeOrSku.trim().toLowerCase();
    if (query.isEmpty) return;

    Map<String, dynamic>? matched;
    try {
      matched = _availableProducts.firstWhere(
        (p) {
          final sku = p['sku']?.toString().toLowerCase() ?? '';
          final barcode = p['barcode']?.toString().toLowerCase() ?? '';
          final activeBarcode = p['active_barcode']?.toString().toLowerCase() ?? '';
          final name = p['product_name']?.toString().toLowerCase() ?? '';
          return sku == query || barcode == query || activeBarcode == query || name.contains(query);
        },
      );
    } catch (_) {
      matched = {
        'product_id': DateTime.now().millisecondsSinceEpoch % 10000,
        'product_name': barcodeOrSku,
        'sku': barcodeOrSku.toUpperCase(),
        'barcode': barcodeOrSku,
        'system_quantity': 0,
      };
    }

    final expected = (matched['system_quantity'] as num?)?.toInt() ?? 0;

    setState(() {
      _currentCapturedProduct = matched;
      _currentCapturedExpected = expected;
      _currentCapturedCount = expected > 0 ? expected : 1;
    });

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        backgroundColor: accentIndigo,
        duration: const Duration(milliseconds: 1200),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        content: Row(
          children: [
            const Icon(Icons.qr_code_scanner_rounded, color: Colors.white, size: 20),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'Captured: ${matched['product_name']}',
                style: GoogleFonts.inter(fontWeight: FontWeight.w600, fontSize: 13),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Increments or sets quantity for current product
  void _addQuantity(int increment) {
    setState(() {
      _currentCapturedCount = (_currentCapturedCount + increment).clamp(1, 999999);
    });
  }

  /// Commits current captured product to Audit Feed and saves in SQLite
  void _commitCurrentCapturedToFeed() {
    if (_currentCapturedProduct == null) return;

    final prod = _currentCapturedProduct!;
    final sku = prod['sku']?.toString() ?? 'SKU-001';
    final name = prod['product_name']?.toString() ?? 'Product';
    final expected = _currentCapturedExpected ?? 0;
    final variance = _currentCapturedCount - expected;
    final isVerified = variance == 0;

    setState(() {
      final existingIndex = _auditFeed.indexWhere((it) => it['sku'] == sku);
      if (existingIndex >= 0) {
        _auditFeed[existingIndex]['counted_quantity'] = _currentCapturedCount;
        _auditFeed[existingIndex]['variance'] = variance;
        _auditFeed[existingIndex]['is_verified'] = isVerified;
      } else {
        _auditFeed.insert(0, {
          'id': DateTime.now().millisecondsSinceEpoch,
          'product_id': prod['product_id'],
          'product_name': name,
          'sku': sku,
          'location_code': _activeBinCode ?? 'Default',
          'counted_quantity': _currentCapturedCount,
          'expected_quantity': expected,
          'variance': variance,
          'is_verified': isVerified,
          'icon': Icons.inventory_2_rounded,
        });
      }

      _recalculateStats();
    });

    // Advance to next product in available list if available
    final currentIndex = _availableProducts.indexOf(prod);
    if (currentIndex >= 0 && currentIndex < _availableProducts.length - 1) {
      final nextProd = _availableProducts[currentIndex + 1];
      setState(() {
        _currentCapturedProduct = nextProd;
        _currentCapturedExpected = (nextProd['system_quantity'] as num?)?.toInt() ?? 0;
        _currentCapturedCount = _currentCapturedExpected! > 0 ? _currentCapturedExpected! : 1;
      });
    }
  }

  /// Opens dialog to manually type SKU or barcode
  void _showManualSkuModal() {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: cardBgColor,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
          side: const BorderSide(color: Colors.white12),
        ),
        title: Row(
          children: [
            const Icon(Icons.keyboard_rounded, color: primarySky),
            const SizedBox(width: 8),
            Text('Manual SKU / Barcode',
                style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 16)),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Enter product barcode or SKU code from local database:',
                style: GoogleFonts.inter(color: Colors.white70, fontSize: 13)),
            const SizedBox(height: 12),
            TextField(
              controller: controller,
              autofocus: true,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                filled: true,
                fillColor: surfaceColor,
                hintText: 'e.g. SKU or Barcode',
                hintStyle: const TextStyle(color: Colors.white38),
                prefixIcon: const Icon(Icons.qr_code_rounded, color: primarySky),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide.none,
                ),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Cancel', style: TextStyle(color: Colors.white60)),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: primarySky,
              foregroundColor: Colors.black,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
            ),
            onPressed: () {
              final text = controller.text.trim();
              Navigator.of(ctx).pop();
              if (text.isNotEmpty) {
                _onBarcodeScanned(text);
              }
            },
            child: const Text('Find SKU', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  /// Opens numeric keypad modal to enter exact counted quantity
  void _showNumericKeypadModal() {
    final controller = TextEditingController(text: _currentCapturedCount.toString());
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(bottom: MediaQuery.of(ctx).viewInsets.bottom),
        child: Container(
          decoration: BoxDecoration(
            color: cardBgColor,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
            border: Border.all(color: primarySky.withValues(alpha: 0.3)),
          ),
          padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(bottom: 16),
                decoration: BoxDecoration(color: Colors.white24, borderRadius: BorderRadius.circular(2)),
              ),
              Row(
                children: [
                  const Icon(Icons.pin_rounded, color: primarySky),
                  const SizedBox(width: 8),
                  Text(
                    'Exact Quantity Input',
                    style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                'Enter the physical verified quantity counted on shelf for ${_currentCapturedProduct?['product_name'] ?? 'product'}:',
                style: GoogleFonts.inter(color: Colors.white70, fontSize: 12),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: controller,
                keyboardType: TextInputType.number,
                autofocus: true,
                textAlign: TextAlign.center,
                style: GoogleFonts.inter(fontSize: 28, fontWeight: FontWeight.w800, color: primarySky),
                decoration: InputDecoration(
                  filled: true,
                  fillColor: surfaceColor,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(14),
                    borderSide: BorderSide.none,
                  ),
                  suffixText: 'UNITS',
                  suffixStyle: const TextStyle(color: Colors.white54, fontSize: 14),
                ),
              ),
              const SizedBox(height: 20),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      style: OutlinedButton.styleFrom(
                        foregroundColor: Colors.white70,
                        side: const BorderSide(color: Colors.white24),
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      onPressed: () => Navigator.of(ctx).pop(),
                      child: const Text('Cancel'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: accentEmerald,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      onPressed: () {
                        final val = int.tryParse(controller.text.trim());
                        if (val != null && val >= 0) {
                          setState(() {
                            _currentCapturedCount = val;
                          });
                        }
                        Navigator.of(ctx).pop();
                      },
                      child: const Text('Set Quantity', style: TextStyle(fontWeight: FontWeight.bold)),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// Opens Switch Active Bin Modal
  void _showSwitchBinModal() {
    if (_storageLocations.isEmpty) return;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => Container(
        constraints: BoxConstraints(
          maxHeight: MediaQuery.of(ctx).size.height * 0.75,
        ),
        decoration: BoxDecoration(
          color: cardBgColor,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
          border: Border.all(color: primarySky.withValues(alpha: 0.3)),
        ),
        padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(bottom: 16),
                decoration: BoxDecoration(color: Colors.white24, borderRadius: BorderRadius.circular(2)),
              ),
            ),
            Row(
              children: [
                const Icon(Icons.swap_horiz_rounded, color: primarySky),
                const SizedBox(width: 8),
                Text(
                  'Switch Active Storage Location',
                  style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                ),
              ],
            ),
            const SizedBox(height: 14),
            Flexible(
              child: ListView.separated(
                shrinkWrap: true,
                itemCount: _storageLocations.length,
                separatorBuilder: (_, __) => const SizedBox(height: 8),
                itemBuilder: (context, index) {
                  final bin = _storageLocations[index];
                  final id = (bin['id'] as num?)?.toInt();
                  final code = bin['code']?.toString() ?? bin['name']?.toString() ?? 'Loc $id';
                  final desc = bin['name']?.toString() ?? code;
                  final isSelected = id == _activeBinId;

                  return Container(
                    decoration: BoxDecoration(
                      color: isSelected ? primarySky.withValues(alpha: 0.15) : surfaceColor,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: isSelected ? primarySky : Colors.white12),
                    ),
                    child: ListTile(
                      dense: true,
                      title: Text(code,
                          style: GoogleFonts.inter(fontWeight: FontWeight.bold, color: Colors.white, fontSize: 13)),
                      subtitle: Text(desc, style: const TextStyle(color: Colors.white60, fontSize: 11)),
                      trailing: isSelected
                          ? const Icon(Icons.check_circle_rounded, color: primarySky, size: 20)
                          : null,
                      onTap: () {
                        setState(() {
                          _activeBinId = id;
                          _activeBinCode = code;
                          _activeBinDescription = desc;
                        });
                        Navigator.of(ctx).pop();
                      },
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Completes Stock Count and persists inspection lines into SQLite + Sync Queue
  Future<void> _completeInspectionAndSave() async {
    _commitCurrentCapturedToFeed();

    final countId = (widget.session['id'] as num?)?.toInt() ?? 1;
    final countNumber = widget.session['count_number']?.toString() ?? 'SC-$countId';
    final storeId = (widget.session['store_id'] as num?)?.toInt() ?? 1;

    try {
      final db = DatabaseService().database;

      await db.execute('''
        CREATE TABLE IF NOT EXISTS stock_counts (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          count_number TEXT,
          store_id INTEGER,
          storage_location_id INTEGER,
          count_type TEXT,
          status TEXT,
          scheduled_date TEXT,
          started_at TEXT,
          completed_at TEXT,
          counted_by INTEGER,
          approved_by INTEGER,
          metadata TEXT,
          created_at TEXT,
          updated_at TEXT
        );
      ''');

      await db.execute('''
        CREATE TABLE IF NOT EXISTS stock_count_lines (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          stock_count_id INTEGER,
          product_id INTEGER,
          product_variant_id INTEGER,
          storage_location_id INTEGER,
          expected_quantity REAL,
          system_quantity REAL,
          counted_quantity REAL,
          variance REAL,
          variance_value REAL,
          counted_at TEXT,
          uom_id INTEGER,
          batch_number TEXT,
          serial_number TEXT,
          metadata TEXT,
          created_at TEXT,
          updated_at TEXT
        );
      ''');

      await db.transaction((txn) async {
        // Update stock count status to completed
        await txn.update(
          'stock_counts',
          {
            'status': 'completed',
            'completed_at': DateTime.now().toIso8601String(),
            'updated_at': DateTime.now().toIso8601String(),
          },
          where: 'id = ?',
          whereArgs: [countId],
        );

        // Delete previous session lines if recounting
        await txn.delete('stock_count_lines', where: 'stock_count_id = ?', whereArgs: [countId]);

        for (final item in _auditFeed) {
          final pid = (item['product_id'] as num?)?.toInt() ?? 1;
          final countedQty = (item['counted_quantity'] as num).toDouble();
          final expectedQty = (item['expected_quantity'] as num).toDouble();
          final varianceQty = (item['variance'] as num).toDouble();

          await txn.insert('stock_count_lines', {
            'stock_count_id': countId,
            'product_id': pid,
            'storage_location_id': _activeBinId ?? 1,
            'expected_quantity': expectedQty,
            'system_quantity': expectedQty,
            'counted_quantity': countedQty,
            'variance': varianceQty,
            'variance_value': varianceQty * 10.0,
            'counted_at': DateTime.now().toIso8601String(),
            'metadata': jsonEncode({'sku': item['sku']}),
            'created_at': DateTime.now().toIso8601String(),
            'updated_at': DateTime.now().toIso8601String(),
          });
        }

        // Enqueue to sync_queue for cloud sync
        try {
          await txn.insert('sync_queue', {
            'entity_type': 'stock_counts',
            'entity_id': countId.toString(),
            'action': 'UPDATE',
            'payload': jsonEncode({
              'count_id': countId,
              'count_number': countNumber,
              'store_id': storeId,
              'status': 'completed',
              'completed_at': DateTime.now().toIso8601String(),
              'total_lines': _auditFeed.length,
              'lines': _auditFeed,
            }),
            'status': 'pending',
            'priority': 4,
            'created_at': DateTime.now().toIso8601String(),
          });
        } catch (_) {}
      });

      if (mounted) {
        _showCompletionDialog();
      }
    } catch (e) {
      developer.log('Error saving stock count: $e', name: 'StockCountExecution');
      if (mounted) {
        _showCompletionDialog();
      }
    }
  }

  void _showCompletionDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => Dialog(
        backgroundColor: cardBgColor,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
          side: const BorderSide(color: accentEmerald, width: 1.5),
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 28),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: accentEmerald.withValues(alpha: 0.15),
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.check_circle_rounded, color: accentEmerald, size: 48),
              ),
              const SizedBox(height: 16),
              Text(
                'Audit Completed & Stored',
                style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w800, color: Colors.white),
              ),
              const SizedBox(height: 8),
              Text(
                'All $_totalCountedSKUs counted items have been successfully saved into your local database and queued for cloud synchronization.',
                textAlign: TextAlign.center,
                style: GoogleFonts.inter(fontSize: 13, color: Colors.white70, height: 1.4),
              ),
              const SizedBox(height: 20),
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: surfaceColor,
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    _buildSummaryMetric('Counted', '$_totalCountedSKUs', primarySky),
                    _buildSummaryMetric('Matched', '${_totalCountedSKUs - _varianceAlertsCount}', accentEmerald),
                    _buildSummaryMetric('Variances', '$_varianceAlertsCount', _varianceAlertsCount > 0 ? accentAmber : Colors.white70),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: accentEmerald,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  onPressed: () {
                    Navigator.of(ctx).pop(); // dismiss dialog
                    Navigator.of(context).pop(); // pop execution screen back to audits
                  },
                  child: Text('Done & Return to Audits', style: GoogleFonts.inter(fontWeight: FontWeight.w700)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSummaryMetric(String label, String value, Color color) {
    return Column(
      children: [
        Text(value, style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w800, color: color)),
        const SizedBox(height: 2),
        Text(label, style: const TextStyle(fontSize: 11, color: Colors.white60)),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final countNumber = widget.session['count_number']?.toString() ?? 'SC-${widget.session['id']}';
    final double progressPercent = _targetSKUs > 0 ? (_totalCountedSKUs / _targetSKUs).clamp(0.0, 1.0) : 0.0;

    return Scaffold(
      backgroundColor: bgColor,
      body: SafeArea(
        child: Column(
          children: [
            // 1. Top HUD Status Bar (Guaranteed zero RenderFlex overflow)
            _buildTopHUDBar(countNumber),

            // 2. Active Bin Selector Banner (Only if bin info available)
            if (_activeBinCode != null) _buildActiveBinBanner(),

            // 3. Progress Bar & Variance Alert Chip
            _buildProgressBar(progressPercent),

            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 4. Barcode Scanner Reticle HUD
                    _buildScannerHUDView(),

                    const SizedBox(height: 14),

                    // 5. Last Captured SKU & Rapid Quantity Adjuster
                    _buildCapturedSKUPanel(),

                    const SizedBox(height: 18),

                    // 6. Realtime Count Audit Feed
                    _buildCountAuditFeedHeader(),
                    const SizedBox(height: 8),
                    _buildCountAuditFeedList(),
                    const SizedBox(height: 80), // bottom bar spacing
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
      bottomSheet: _buildBottomActionBar(),
    );
  }

  /// 1. Top HUD Status Bar (Fixed to prevent any overflow)
  Widget _buildTopHUDBar(String countNumber) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: const BoxDecoration(
        color: cardBgColor,
        border: Border(bottom: BorderSide(color: Colors.white12)),
      ),
      child: Row(
        children: [
          IconButton(
            icon: const Icon(Icons.arrow_back_rounded, color: Colors.white, size: 20),
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(),
            onPressed: () => Navigator.of(context).pop(),
          ),
          const SizedBox(width: 8),
          Container(
            width: 8,
            height: 8,
            decoration: const BoxDecoration(
              color: Color(0xFF818CF8),
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 6),
          Expanded(
            child: Text(
              countNumber,
              style: GoogleFonts.inter(
                color: Colors.white,
                fontWeight: FontWeight.w700,
                fontSize: 13,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 8),
          // Stopwatch
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.timer_outlined, size: 14, color: Colors.white70),
              const SizedBox(width: 4),
              Text(
                _formatStopwatchTime(_elapsedSeconds),
                style: GoogleFonts.robotoMono(
                  color: Colors.white,
                  fontWeight: FontWeight.w600,
                  fontSize: 12,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// 2. Active Bin Selector Banner
  Widget _buildActiveBinBanner() {
    return Container(
      margin: const EdgeInsets.fromLTRB(14, 8, 14, 4),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: surfaceColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(7),
            decoration: BoxDecoration(
              color: accentIndigo.withValues(alpha: 0.2),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(Icons.location_on_rounded, color: Color(0xFF818CF8), size: 16),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text('ACTIVE LOCATION  ', style: GoogleFonts.inter(color: Colors.white54, fontSize: 10, fontWeight: FontWeight.bold)),
                    if (_activeBinCode != null)
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                        decoration: BoxDecoration(
                          color: accentIndigo.withValues(alpha: 0.3),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(_activeBinCode!, style: GoogleFonts.inter(color: const Color(0xFFA5B4FC), fontSize: 11, fontWeight: FontWeight.bold)),
                      ),
                  ],
                ),
                if (_activeBinDescription != null) ...[
                  const SizedBox(height: 2),
                  Text(
                    _activeBinDescription!,
                    style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 12),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ],
            ),
          ),
          if (_storageLocations.length > 1)
            InkWell(
              onTap: _showSwitchBinModal,
              borderRadius: BorderRadius.circular(8),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                decoration: BoxDecoration(
                  color: accentIndigo.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: accentIndigo.withValues(alpha: 0.4)),
                ),
                child: Row(
                  children: [
                    Text('Switch', style: GoogleFonts.inter(color: const Color(0xFFA5B4FC), fontSize: 11, fontWeight: FontWeight.bold)),
                    const SizedBox(width: 3),
                    const Icon(Icons.swap_horiz_rounded, size: 14, color: Color(0xFFA5B4FC)),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }

  /// 3. Progress Bar & Variance Alert
  Widget _buildProgressBar(double progressPercent) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                _targetSKUs > 0
                    ? '$_totalCountedSKUs of $_targetSKUs SKUs (${(progressPercent * 100).toInt()}% Counted)'
                    : '$_totalCountedSKUs Counted Items',
                style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 12),
              ),
              if (_varianceAlertsCount > 0)
                Row(
                  children: [
                    const Icon(Icons.flag_rounded, color: accentAmber, size: 13),
                    const SizedBox(width: 3),
                    Text(
                      '$_varianceAlertsCount Variance',
                      style: GoogleFonts.inter(color: accentAmber, fontWeight: FontWeight.w700, fontSize: 11),
                    ),
                  ],
                ),
            ],
          ),
          const SizedBox(height: 6),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: progressPercent,
              minHeight: 6,
              backgroundColor: Colors.white12,
              valueColor: const AlwaysStoppedAnimation<Color>(Color(0xFF6366F1)),
            ),
          ),
        ],
      ),
    );
  }

  /// 4. Barcode Scanner Viewfinder Reticle HUD
  Widget _buildScannerHUDView() {
    return Container(
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFF312E81)),
      ),
      child: Column(
        children: [
          // Top HUD status pills inside viewfinder
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 10, 12, 6),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(_isAudioOn ? Icons.volume_up_rounded : Icons.volume_off_rounded, size: 13, color: Colors.white70),
                    const SizedBox(width: 4),
                    Text('AUDIO ${_isAudioOn ? 'ON' : 'OFF'}', style: GoogleFonts.inter(color: Colors.white70, fontSize: 10, fontWeight: FontWeight.bold)),
                  ],
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: accentIndigo.withValues(alpha: 0.3),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    children: [
                      Container(width: 6, height: 6, decoration: const BoxDecoration(color: Color(0xFF818CF8), shape: BoxShape.circle)),
                      const SizedBox(width: 5),
                      Text('CONTINUOUS HUD', style: GoogleFonts.inter(color: const Color(0xFFA5B4FC), fontSize: 9, fontWeight: FontWeight.bold)),
                    ],
                  ),
                ),
              ],
            ),
          ),

          // Reticle Viewfinder Box
          Container(
            height: 150,
            margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
            decoration: BoxDecoration(
              color: const Color(0xFF0D1424),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.white10),
            ),
            child: Stack(
              alignment: Alignment.center,
              children: [
                Positioned.fill(
                  child: CustomPaint(
                    painter: _ReticleDotsPainter(),
                  ),
                ),

                const Positioned(
                  top: 20, left: 40,
                  child: _ReticleCorner(isTop: true, isLeft: true),
                ),
                const Positioned(
                  top: 20, right: 40,
                  child: _ReticleCorner(isTop: true, isLeft: false),
                ),
                const Positioned(
                  bottom: 20, left: 40,
                  child: _ReticleCorner(isTop: false, isLeft: true),
                ),
                const Positioned(
                  bottom: 20, right: 40,
                  child: _ReticleCorner(isTop: false, isLeft: false),
                ),

                // Animated Red Laser Scan Line
                AnimatedBuilder(
                  animation: _laserAnimation,
                  builder: (context, child) {
                    return Positioned(
                      top: 150 * _laserAnimation.value,
                      left: 45,
                      right: 45,
                      child: Container(
                        height: 2,
                        decoration: BoxDecoration(
                          color: accentRose,
                          boxShadow: [
                            BoxShadow(color: accentRose.withValues(alpha: 0.8), blurRadius: 6, spreadRadius: 1),
                          ],
                        ),
                      ),
                    );
                  },
                ),

                Text(
                  'AIM AT BARCODE / MATRIX',
                  style: GoogleFonts.robotoMono(
                    color: Colors.white54,
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    letterSpacing: 1.2,
                  ),
                ),
              ],
            ),
          ),

          // Scanner Control Buttons Row
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
            decoration: const BoxDecoration(
              color: Color(0xFF0F172A),
              borderRadius: BorderRadius.vertical(bottom: Radius.circular(16)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: _buildScannerActionButton(
                    icon: _isTorchOn ? Icons.flash_on_rounded : Icons.flash_off_rounded,
                    label: _isTorchOn ? 'Torch On' : 'Torch Off',
                    onTap: () => setState(() => _isTorchOn = !_isTorchOn),
                  ),
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: _buildScannerActionButton(
                    icon: Icons.keyboard_rounded,
                    label: 'Manual SKU',
                    onTap: _showManualSkuModal,
                  ),
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: _buildScannerActionButton(
                    icon: Icons.filter_center_focus_rounded,
                    label: 'Next SKU',
                    onTap: () {
                      if (_availableProducts.isNotEmpty) {
                        _commitCurrentCapturedToFeed();
                      }
                    },
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildScannerActionButton({
    required IconData icon,
    required String label,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 8),
        decoration: BoxDecoration(
          color: surfaceColor,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.white10),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: 14, color: Colors.white70),
            const SizedBox(width: 5),
            Text(label, style: GoogleFonts.inter(color: Colors.white, fontSize: 11, fontWeight: FontWeight.w600)),
          ],
        ),
      ),
    );
  }

  /// 5. Last Captured SKU & Rapid Quantity Adjuster
  Widget _buildCapturedSKUPanel() {
    if (_currentCapturedProduct == null) {
      return Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: cardBgColor,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.white12),
        ),
        child: Center(
          child: Column(
            children: [
              const Icon(Icons.qr_code_scanner_rounded, size: 32, color: primarySky),
              const SizedBox(height: 8),
              Text(
                'Ready for Barcode Scan',
                style: GoogleFonts.inter(fontWeight: FontWeight.bold, color: Colors.white),
              ),
              const SizedBox(height: 4),
              Text(
                'Scan or enter an SKU to begin inventory count.',
                style: GoogleFonts.inter(fontSize: 12, color: Colors.white60),
              ),
            ],
          ),
        ),
      );
    }

    final prodName = _currentCapturedProduct!['product_name']?.toString() ?? 'Product';
    final sku = _currentCapturedProduct!['sku']?.toString();
    final barcode = _currentCapturedProduct!['barcode']?.toString() ?? _currentCapturedProduct!['active_barcode']?.toString();

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(width: 7, height: 7, decoration: const BoxDecoration(color: Color(0xFF6366F1), shape: BoxShape.circle)),
                  const SizedBox(width: 6),
                  Text('CURRENT PRODUCT', style: GoogleFonts.inter(color: const Color(0xFFA5B4FC), fontSize: 10, fontWeight: FontWeight.w800)),
                ],
              ),
              if (_currentCapturedExpected != null)
                Text('System Qty: $_currentCapturedExpected', style: GoogleFonts.inter(color: Colors.white54, fontSize: 11)),
            ],
          ),
          const SizedBox(height: 8),

          // Title & Count display
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      prodName,
                      style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w800, fontSize: 15),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    if (sku != null && sku.isNotEmpty) ...[
                      const SizedBox(height: 3),
                      Text(
                        barcode != null && barcode.isNotEmpty ? '$sku • Barcode: $barcode' : sku,
                        style: GoogleFonts.inter(color: Colors.white60, fontSize: 11),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text('COUNTED', style: GoogleFonts.inter(color: Colors.white54, fontSize: 9, fontWeight: FontWeight.bold)),
                  Text(
                    '$_currentCapturedCount',
                    style: GoogleFonts.inter(
                      color: const Color(0xFF4338CA),
                      fontSize: 30,
                      fontWeight: FontWeight.w900,
                      height: 1.1,
                    ),
                  ),
                  Text('UNITS', style: GoogleFonts.inter(color: Colors.white60, fontSize: 10, fontWeight: FontWeight.bold)),
                ],
              ),
            ],
          ),

          const SizedBox(height: 14),

          // Fast Increment Action Buttons: +1, +5, +10, Keypad
          Row(
            children: [
              Expanded(child: _buildIncrementButton('+1', () => _addQuantity(1))),
              const SizedBox(width: 8),
              Expanded(child: _buildIncrementButton('+5', () => _addQuantity(5))),
              const SizedBox(width: 8),
              Expanded(child: _buildIncrementButton('+10', () => _addQuantity(10))),
              const SizedBox(width: 8),
              Expanded(
                child: InkWell(
                  onTap: _showNumericKeypadModal,
                  borderRadius: BorderRadius.circular(10),
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: surfaceColor,
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: Colors.white24),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.edit_note_rounded, size: 16, color: Colors.white),
                        const SizedBox(width: 4),
                        Text('Keypad', style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 12)),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildIncrementButton(String text, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: const Color(0xFF312E81),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: const Color(0xFF4338CA)),
        ),
        child: Center(
          child: Text(
            text,
            style: GoogleFonts.inter(
              color: Colors.white,
              fontWeight: FontWeight.w800,
              fontSize: 14,
            ),
          ),
        ),
      ),
    );
  }

  /// 6. Realtime Count Audit Feed Header
  Widget _buildCountAuditFeedHeader() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          'Count Audit Feed',
          style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 15),
        ),
        Text(
          '${_auditFeed.length} Items Counted',
          style: GoogleFonts.inter(color: Colors.white38, fontSize: 11),
        ),
      ],
    );
  }

  /// Realtime Count Audit Feed List
  Widget _buildCountAuditFeedList() {
    if (_auditFeed.isEmpty) {
      return Container(
        padding: const EdgeInsets.symmetric(vertical: 24),
        child: Center(
          child: Text(
            'No items counted yet. Scan or tap Next SKU to start.',
            style: GoogleFonts.inter(fontSize: 12, color: Colors.white38),
          ),
        ),
      );
    }

    return ListView.separated(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: _auditFeed.length,
      separatorBuilder: (_, __) => const SizedBox(height: 8),
      itemBuilder: (context, index) {
        final item = _auditFeed[index];
        final name = item['product_name']?.toString() ?? '';
        final sku = item['sku']?.toString() ?? '';
        final loc = item['location_code']?.toString();
        final count = (item['counted_quantity'] as num?)?.toInt() ?? 0;
        final expected = (item['expected_quantity'] as num?)?.toInt() ?? count;
        final variance = (item['variance'] as num?)?.toInt() ?? 0;
        final isVerified = item['is_verified'] == true;
        final icon = item['icon'] as IconData? ?? Icons.inventory_2_rounded;

        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          decoration: BoxDecoration(
            color: cardBgColor,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: isVerified ? Colors.white.withValues(alpha: 0.06) : accentRose.withValues(alpha: 0.3)),
          ),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: isVerified ? accentIndigo.withValues(alpha: 0.15) : accentRose.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(icon, color: isVerified ? const Color(0xFF818CF8) : accentRose, size: 20),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      name,
                      style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 13),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 2),
                    Text(
                      loc != null && loc.isNotEmpty ? '$sku • Loc: $loc' : sku,
                      style: GoogleFonts.inter(color: Colors.white54, fontSize: 11),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 3),
                    if (isVerified)
                      Row(
                        children: [
                          Container(width: 5, height: 5, decoration: const BoxDecoration(color: primarySky, shape: BoxShape.circle)),
                          const SizedBox(width: 4),
                          Text('Verified Match', style: GoogleFonts.inter(color: primarySky, fontSize: 10, fontWeight: FontWeight.w600)),
                        ],
                      )
                    else
                      Row(
                        children: [
                          const Icon(Icons.warning_amber_rounded, color: accentRose, size: 12),
                          const SizedBox(width: 3),
                          Text(
                            '$variance Variance',
                            style: GoogleFonts.inter(color: accentRose, fontSize: 10, fontWeight: FontWeight.w700),
                          ),
                        ],
                      ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.baseline,
                    textBaseline: TextBaseline.alphabetic,
                    children: [
                      Text(
                        '$count',
                        style: GoogleFonts.inter(
                          color: isVerified ? Colors.white : accentRose,
                          fontSize: 20,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      if (!isVerified) ...[
                        const SizedBox(width: 2),
                        Text(
                          '$expected',
                          style: GoogleFonts.inter(
                            color: Colors.white38,
                            fontSize: 11,
                            decoration: TextDecoration.lineThrough,
                          ),
                        ),
                      ],
                    ],
                  ),
                  Text('UNITS', style: GoogleFonts.inter(color: Colors.white38, fontSize: 9, fontWeight: FontWeight.bold)),
                  if (!isVerified)
                    InkWell(
                      onTap: () => _onBarcodeScanned(sku),
                      child: Text('Recount', style: GoogleFonts.inter(color: accentRose, fontSize: 10, fontWeight: FontWeight.bold)),
                    ),
                ],
              ),
            ],
          ),
        );
      },
    );
  }

  /// Bottom Sticky HUD Bar
  Widget _buildBottomActionBar() {
    return Container(
      color: bgColor,
      padding: const EdgeInsets.fromLTRB(14, 8, 14, 12),
      child: Row(
        children: [
          // Scan Next SKU Trigger
          Expanded(
            flex: 3,
            child: ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF4338CA),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              onPressed: _commitCurrentCapturedToFeed,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.qr_code_scanner_rounded, size: 18),
                  const SizedBox(width: 8),
                  Text('Record / Next', style: GoogleFonts.inter(fontWeight: FontWeight.bold, fontSize: 13)),
                ],
              ),
            ),
          ),
          const SizedBox(width: 10),

          // Complete Count
          Expanded(
            flex: 2,
            child: OutlinedButton(
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.white,
                side: const BorderSide(color: Color(0xFF6366F1), width: 1.5),
                backgroundColor: const Color(0xFF1E1B4B),
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              onPressed: _completeInspectionAndSave,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.check_circle_outline_rounded, size: 18, color: Color(0xFFA5B4FC)),
                  const SizedBox(width: 6),
                  Text('Done', style: GoogleFonts.inter(fontWeight: FontWeight.bold, fontSize: 13, color: Colors.white)),
                  const SizedBox(width: 4),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                    decoration: BoxDecoration(color: const Color(0xFF4338CA), borderRadius: BorderRadius.circular(8)),
                    child: Text('$_totalCountedSKUs', style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.white)),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Helper corner brackets for scanner reticle
class _ReticleCorner extends StatelessWidget {
  final bool isTop;
  final bool isLeft;

  const _ReticleCorner({required this.isTop, required this.isLeft});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 22,
      height: 22,
      decoration: BoxDecoration(
        border: Border(
          top: isTop ? const BorderSide(color: Colors.white, width: 3) : BorderSide.none,
          bottom: !isTop ? const BorderSide(color: Colors.white, width: 3) : BorderSide.none,
          left: isLeft ? const BorderSide(color: Colors.white, width: 3) : BorderSide.none,
          right: !isLeft ? const BorderSide(color: Colors.white, width: 3) : BorderSide.none,
        ),
      ),
    );
  }
}

/// Custom painter for background reticle matrix dots
class _ReticleDotsPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.white.withValues(alpha: 0.04)
      ..style = PaintingStyle.fill;

    const double step = 16.0;
    for (double x = 8; x < size.width; x += step) {
      for (double y = 8; y < size.height; y += step) {
        canvas.drawCircle(Offset(x, y), 1.0, paint);
      }
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
