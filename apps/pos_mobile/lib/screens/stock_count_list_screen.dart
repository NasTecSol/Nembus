import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import 'count_initiation_screen.dart';
import 'stock_count_execution_screen.dart';

class StockCountListScreen extends StatefulWidget {
  final int userId;
  final String username;
  final int? initialStoreId;
  final String? initialStoreName;

  const StockCountListScreen({
    super.key,
    required this.userId,
    this.username = 'Cashier',
    this.initialStoreId,
    this.initialStoreName,
  });

  @override
  State<StockCountListScreen> createState() => _StockCountListScreenState();
}

class _StockCountListScreenState extends State<StockCountListScreen> {
  bool _isLoading = true;
  String? _errorMessage;

  // Stock count sessions data from local SQLite
  List<Map<String, dynamic>> _stockCounts = [];
  List<Map<String, dynamic>> _filteredCounts = [];

  // Summary Metrics
  int _activeInProgressCount = 0;
  int _plannedScheduledCount = 0;
  int _completedReconciledCount = 0;
  double _netVarianceValue = 0.0;

  // Filter state
  String _searchQuery = '';
  int? _selectedStoreId;
  String _selectedStatus = 'ALL';
  String _selectedMethodology = 'ALL';

  // Available filters data
  List<Map<String, dynamic>> _stores = [];
  List<Map<String, dynamic>> _locations = [];

  final TextEditingController _searchController = TextEditingController();

  // App Theme Constants (Matching POS Mobile design system)
  static const Color bgColor = Color(0xFF090D16);
  static const Color cardBgColor = Color(0xFF131B2E);
  static const Color surfaceColor = Color(0xFF1E293B);
  static const Color primarySky = Color(0xFF38BDF8);
  static const Color accentEmerald = Color(0xFF10B981);
  static const Color accentAmber = Color(0xFFF59E0B);
  static const Color accentTeal = Color(0xFF14B8A6);

  @override
  void initState() {
    super.initState();
    _selectedStoreId = widget.initialStoreId ?? SingletonClass().activeStoreId;
    _loadStoresAndLocations();
    _fetchStockCounts();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  /// Loads available stores and storage locations for filters
  Future<void> _loadStoresAndLocations() async {
    try {
      final db = DatabaseService().database;
      final storeRows = await db.query('stores', orderBy: 'name ASC');
      final locRows = await db.query('storage_locations', orderBy: 'name ASC');

      if (mounted) {
        setState(() {
          _stores = List<Map<String, dynamic>>.from(storeRows);
          _locations = List<Map<String, dynamic>>.from(locRows);
        });
      }
    } catch (_) {}
  }

  /// Fetches real stock counts from Go Handler route `api/stock-counts` with local SQLite fallback
  Future<void> _fetchStockCounts() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      List<Map<String, dynamic>> loadedSessions = [];

      // 1. Try Go Core Bridge Handler call (/api/stock-counts)
      try {
        final bridgeResp = NembusBridge().callHandler(
          handler: 'stock_counts',
          action: 'listStockCounts',
          payload: {
            if (_selectedStoreId != null && _selectedStoreId! > 0) 'store_id': _selectedStoreId,
            if (_selectedStatus != 'ALL') 'status': _selectedStatus.toLowerCase(),
            if (_selectedMethodology != 'ALL') 'count_type': _selectedMethodology.toLowerCase(),
            'page': 1,
            'limit': 100,
          },
        );

        if (bridgeResp['success'] == true && bridgeResp['data'] is Map) {
          final dataMap = bridgeResp['data'] as Map;
          final list = dataMap['data'];
          if (list is List && list.isNotEmpty) {
            loadedSessions = list.map((e) => Map<String, dynamic>.from(e as Map)).toList();
          }
        } else if (bridgeResp['success'] == true && bridgeResp['data'] is List && (bridgeResp['data'] as List).isNotEmpty) {
          loadedSessions = (bridgeResp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
      } catch (_) {}

      // 2. Direct local SQLite query to ensure rich local data joined with stores & lines
      if (loadedSessions.isEmpty) {
        final db = DatabaseService().database;

        // Ensure table exists
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

        final rawRows = await db.rawQuery('''
          SELECT 
            sc.id,
            sc.count_number,
            sc.store_id,
            sc.storage_location_id,
            sc.count_type,
            sc.status,
            sc.scheduled_date,
            sc.started_at,
            sc.completed_at,
            sc.counted_by,
            sc.approved_by,
            sc.metadata,
            sc.created_at,
            st.name AS store_name,
            sl.name AS storage_location_name,
            u_cnt.username AS counted_by_name,
            u_app.username AS approved_by_name,
            COUNT(scl.id) AS total_lines,
            COALESCE(SUM(CASE WHEN scl.variance != 0 THEN 1 ELSE 0 END), 0) AS lines_with_variance,
            COALESCE(SUM(scl.variance_value), 0.0) AS total_variance_value
          FROM stock_counts sc
          LEFT JOIN stores st ON sc.store_id = st.id
          LEFT JOIN storage_locations sl ON sc.storage_location_id = sl.id
          LEFT JOIN users u_cnt ON sc.counted_by = u_cnt.id
          LEFT JOIN users u_app ON sc.approved_by = u_app.id
          LEFT JOIN stock_count_lines scl ON sc.id = scl.stock_count_id
          GROUP BY sc.id
          ORDER BY sc.created_at DESC, sc.id DESC
        ''');

        loadedSessions = rawRows.map((r) => Map<String, dynamic>.from(r)).toList();
      }

      _stockCounts = loadedSessions;
      _computeMetrics();
      _applyFilters();

      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    } catch (e) {
      developer.log('Error fetching stock counts: $e', name: 'StockCountListScreen');
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  void _computeMetrics() {
    int active = 0;
    int scheduled = 0;
    int completed = 0;
    double netVariance = 0.0;

    for (final s in _stockCounts) {
      final status = (s['status']?.toString() ?? '').toLowerCase();
      if (status == 'in_progress' || status == 'in-progress' || status == 'active') {
        active++;
      } else if (status == 'scheduled' || status == 'draft' || status == 'planned') {
        scheduled++;
      } else if (status == 'completed' || status == 'approved' || status == 'reconciled') {
        completed++;
      }

      final varianceVal = (s['total_variance_value'] as num?)?.toDouble() ?? 0.0;
      netVariance += varianceVal;
    }

    _activeInProgressCount = active;
    _plannedScheduledCount = scheduled;
    _completedReconciledCount = completed;
    _netVarianceValue = netVariance;
  }

  void _applyFilters() {
    setState(() {
      _filteredCounts = _stockCounts.where((item) {
        // Store filter
        if (_selectedStoreId != null && _selectedStoreId! > 0) {
          final storeId = (item['store_id'] as num?)?.toInt();
          if (storeId != _selectedStoreId) return false;
        }

        // Status filter
        if (_selectedStatus != 'ALL') {
          final status = (item['status']?.toString() ?? '').toUpperCase();
          if (_selectedStatus == 'IN_PROGRESS' &&
              status != 'IN_PROGRESS' &&
              status != 'IN-PROGRESS' &&
              status != 'ACTIVE') {
            return false;
          }
          if (_selectedStatus == 'SCHEDULED' &&
              status != 'SCHEDULED' &&
              status != 'DRAFT' &&
              status != 'PLANNED') {
            return false;
          }
          if (_selectedStatus == 'COMPLETED' &&
              status != 'COMPLETED' &&
              status != 'APPROVED' &&
              status != 'RECONCILED') {
            return false;
          }
          if (_selectedStatus == 'CANCELLED' && status != 'CANCELLED' && status != 'REJECTED') {
            return false;
          }
        }

        // Methodology filter
        if (_selectedMethodology != 'ALL') {
          final countType = (item['count_type']?.toString() ?? '').toUpperCase();
          if (countType != _selectedMethodology) return false;
        }

        // Search query
        if (_searchQuery.isNotEmpty) {
          final q = _searchQuery.toLowerCase();
          final countNumber = (item['count_number'] ?? '').toString().toLowerCase();
          final storeName = (item['store_name'] ?? '').toString().toLowerCase();
          final locationName = (item['storage_location_name'] ?? '').toString().toLowerCase();
          final countedByName = (item['counted_by_name'] ?? '').toString().toLowerCase();

          final match = countNumber.contains(q) ||
              storeName.contains(q) ||
              locationName.contains(q) ||
              countedByName.contains(q);
          if (!match) return false;
        }

        return true;
      }).toList();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: bgColor,
      appBar: AppBar(
        backgroundColor: cardBgColor,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded, color: Colors.white),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Stock Counts',
              style: GoogleFonts.inter(
                fontSize: 16,
                fontWeight: FontWeight.w700,
                color: Colors.white,
              ),
            ),
            Text(
              'Local Inventory Sessions',
              style: GoogleFonts.inter(
                fontSize: 11,
                color: Colors.white54,
              ),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded, color: Colors.white70),
            onPressed: _fetchStockCounts,
            tooltip: 'Refresh',
          ),
        ],
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: RefreshIndicator(
                onRefresh: _fetchStockCounts,
                color: primarySky,
                backgroundColor: cardBgColor,
                child: ListView(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                  children: [
                    // 1. Mobile 2x2 Summary Metric Cards
                    _buildMobileSummaryMetrics(),

                    const SizedBox(height: 12),

                    // 2. Search Input Field
                    _buildMobileSearchBar(),

                    const SizedBox(height: 10),

                    // 3. Status Filter Chips (Horizontal Scrollable)
                    _buildMobileStatusChips(),

                    const SizedBox(height: 10),

                    // 4. Store & Methodology Dropdown Filters
                    _buildMobileDropdownFilters(),

                    const SizedBox(height: 14),

                    // 5. Stock Count List Items or Empty State
                    if (_isLoading)
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 40),
                        child: Center(
                          child: CircularProgressIndicator(color: primarySky),
                        ),
                      )
                    else if (_filteredCounts.isEmpty)
                      _buildMobileEmptyState()
                    else
                      ..._filteredCounts.map((session) => _buildStockCountCard(session)),

                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// 1. Mobile 2x2 Metric Cards
  Widget _buildMobileSummaryMetrics() {
    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: _buildMetricTile(
                title: 'In Progress',
                value: '$_activeInProgressCount',
                icon: Icons.timelapse_rounded,
                color: accentAmber,
                bgColor: const Color(0xFF1C1917),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _buildMetricTile(
                title: 'Scheduled',
                value: '$_plannedScheduledCount',
                icon: Icons.calendar_today_rounded,
                color: primarySky,
                bgColor: const Color(0xFF0C2A4D),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(
              child: _buildMetricTile(
                title: 'Completed',
                value: '$_completedReconciledCount',
                icon: Icons.check_circle_outline_rounded,
                color: accentTeal,
                bgColor: const Color(0xFF062828),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _buildMetricTile(
                title: 'Net Variance',
                value: _netVarianceValue.abs() > 0.001
                    ? '\$${_netVarianceValue.toStringAsFixed(2)}'
                    : '\$0.00',
                icon: Icons.show_chart_rounded,
                color: _netVarianceValue.abs() > 0.001 ? accentAmber : accentEmerald,
                bgColor: const Color(0xFF142E1F),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildMetricTile({
    required String title,
    required String value,
    required IconData icon,
    required Color color,
    required Color bgColor,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withValues(alpha: 0.06)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(icon, size: 18, color: color),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  value,
                  style: GoogleFonts.inter(
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                    color: Colors.white,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                Text(
                  title,
                  style: GoogleFonts.inter(
                    fontSize: 11,
                    fontWeight: FontWeight.w500,
                    color: Colors.white60,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// 2. Mobile Search Bar
  Widget _buildMobileSearchBar() {
    return Container(
      height: 42,
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
      ),
      child: TextField(
        controller: _searchController,
        style: GoogleFonts.inter(color: Colors.white, fontSize: 13),
        decoration: InputDecoration(
          hintText: 'Search session number, store...',
          hintStyle: GoogleFonts.inter(color: Colors.white38, fontSize: 13),
          prefixIcon: const Icon(Icons.search_rounded, color: Colors.white38, size: 20),
          suffixIcon: _searchQuery.isNotEmpty
              ? IconButton(
                  icon: const Icon(Icons.clear_rounded, color: Colors.white38, size: 18),
                  onPressed: () {
                    _searchController.clear();
                    setState(() {
                      _searchQuery = '';
                      _applyFilters();
                    });
                  },
                )
              : null,
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 11),
        ),
        onChanged: (val) {
          _searchQuery = val.trim();
          _applyFilters();
        },
      ),
    );
  }

  /// 3. Status Filter Chips
  Widget _buildMobileStatusChips() {
    final statuses = ['ALL', 'IN_PROGRESS', 'SCHEDULED', 'COMPLETED', 'CANCELLED'];

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: statuses.map((st) {
          final isSelected = _selectedStatus == st;
          String label;
          switch (st) {
            case 'ALL':
              label = 'All';
              break;
            case 'IN_PROGRESS':
              label = 'In Progress';
              break;
            case 'SCHEDULED':
              label = 'Scheduled';
              break;
            case 'COMPLETED':
              label = 'Completed';
              break;
            case 'CANCELLED':
              label = 'Cancelled';
              break;
            default:
              label = st;
          }

          return Padding(
            padding: const EdgeInsets.only(right: 6),
            child: ChoiceChip(
              label: Text(label),
              selected: isSelected,
              labelStyle: GoogleFonts.inter(
                fontSize: 11,
                fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                color: isSelected ? Colors.black : Colors.white70,
              ),
              selectedColor: primarySky,
              backgroundColor: cardBgColor,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
                side: BorderSide(
                  color: isSelected ? primarySky : Colors.white.withValues(alpha: 0.08),
                ),
              ),
              onSelected: (selected) {
                if (selected) {
                  setState(() {
                    _selectedStatus = st;
                    _applyFilters();
                  });
                }
              },
              showCheckmark: false,
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            ),
          );
        }).toList(),
      ),
    );
  }

  /// 4. Store & Methodology Dropdown Filters
  Widget _buildMobileDropdownFilters() {
    final storeItems = <DropdownMenuItem<int?>>[
      DropdownMenuItem<int?>(
        value: null,
        child: Text('All Stores', style: GoogleFonts.inter(fontSize: 11, color: Colors.white70)),
      ),
    ];

    final seenStoreIds = <int>{};
    for (final st in _stores) {
      final id = (st['id'] as num?)?.toInt();
      if (id != null && seenStoreIds.add(id)) {
        storeItems.add(
          DropdownMenuItem<int?>(
            value: id,
            child: Text(
              st['name']?.toString() ?? 'Store $id',
              style: GoogleFonts.inter(fontSize: 11, color: Colors.white),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        );
      }
    }

    final int? effectiveStoreId =
        (_selectedStoreId != null && seenStoreIds.contains(_selectedStoreId))
            ? _selectedStoreId
            : null;

    const validMethodologies = {'ALL', 'FULL', 'CYCLE', 'SPOT', 'ANNUAL'};
    final effectiveMethodology = validMethodologies.contains(_selectedMethodology)
        ? _selectedMethodology
        : 'ALL';

    return Row(
      children: [
        // Store Selector
        Expanded(
          child: Container(
            height: 38,
            padding: const EdgeInsets.symmetric(horizontal: 10),
            decoration: BoxDecoration(
              color: cardBgColor,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
            ),
            child: DropdownButtonHideUnderline(
              child: DropdownButton<int?>(
                isExpanded: true,
                dropdownColor: cardBgColor,
                value: effectiveStoreId,
                icon: const Icon(Icons.keyboard_arrow_down_rounded, size: 16, color: Colors.white54),
                items: storeItems,
                onChanged: (val) {
                  setState(() {
                    _selectedStoreId = val;
                    _applyFilters();
                  });
                },
              ),
            ),
          ),
        ),
        const SizedBox(width: 8),

        // Methodology Selector
        Expanded(
          child: Container(
            height: 38,
            padding: const EdgeInsets.symmetric(horizontal: 10),
            decoration: BoxDecoration(
              color: cardBgColor,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
            ),
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                isExpanded: true,
                dropdownColor: cardBgColor,
                value: effectiveMethodology,
                icon: const Icon(Icons.keyboard_arrow_down_rounded, size: 16, color: Colors.white54),
                items: const [
                  DropdownMenuItem(value: 'ALL', child: Text('All Types', style: TextStyle(fontSize: 11, color: Colors.white70))),
                  DropdownMenuItem(value: 'FULL', child: Text('Full Count', style: TextStyle(fontSize: 11, color: Colors.white))),
                  DropdownMenuItem(value: 'CYCLE', child: Text('Cycle Count', style: TextStyle(fontSize: 11, color: Colors.white))),
                  DropdownMenuItem(value: 'SPOT', child: Text('Spot Check', style: TextStyle(fontSize: 11, color: Colors.white))),
                  DropdownMenuItem(value: 'ANNUAL', child: Text('Annual Audit', style: TextStyle(fontSize: 11, color: Colors.white))),
                ],
                onChanged: (val) {
                  if (val != null) {
                    setState(() {
                      _selectedMethodology = val;
                      _applyFilters();
                    });
                  }
                },
              ),
            ),
          ),
        ),
      ],
    );
  }

  /// 5. Stock Count Card Item
  Widget _buildStockCountCard(Map<String, dynamic> session) {
    final countNumber = session['count_number']?.toString() ?? 'SC-${session['id']}';
    final status = session['status']?.toString().toUpperCase() ?? 'SCHEDULED';
    final storeName = session['store_name']?.toString();
    final locationName = session['storage_location_name']?.toString();
    final countType = session['count_type']?.toString().toUpperCase() ?? 'FULL';
    final totalLines = (session['total_lines'] as num?)?.toInt() ?? 0;
    final varianceLines = (session['lines_with_variance'] as num?)?.toInt() ?? 0;
    final totalVarianceVal = (session['total_variance_value'] as num?)?.toDouble() ?? 0.0;
    final countedBy = session['counted_by_name']?.toString();
    final scheduledDate = session['scheduled_date']?.toString();

    Color statusColor;
    Color statusBg;
    switch (status) {
      case 'IN_PROGRESS':
      case 'IN-PROGRESS':
      case 'ACTIVE':
        statusColor = accentAmber;
        statusBg = accentAmber.withValues(alpha: 0.15);
        break;
      case 'COMPLETED':
      case 'APPROVED':
      case 'RECONCILED':
        statusColor = accentEmerald;
        statusBg = accentEmerald.withValues(alpha: 0.15);
        break;
      case 'CANCELLED':
      case 'REJECTED':
        statusColor = const Color(0xFFEF4444);
        statusBg = const Color(0xFFEF4444).withValues(alpha: 0.15);
        break;
      default:
        statusColor = primarySky;
        statusBg = primarySky.withValues(alpha: 0.15);
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Colors.white.withValues(alpha: 0.06)),
      ),
      child: InkWell(
        onTap: () => _openStockCountExecution(session),
        borderRadius: BorderRadius.circular(14),
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Row 1: Count Number + Status Chip
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: surfaceColor,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      countNumber,
                      style: GoogleFonts.inter(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: Colors.white,
                      ),
                    ),
                  ),
                  const Spacer(),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: statusBg,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      status,
                      style: GoogleFonts.inter(
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                        color: statusColor,
                      ),
                    ),
                  ),
                ],
              ),

              // Row 2: Store & Location (Only if present in SQLite)
              if (storeName != null && storeName.isNotEmpty) ...[
                const SizedBox(height: 8),
                Row(
                  children: [
                    const Icon(Icons.storefront_rounded, size: 14, color: primarySky),
                    const SizedBox(width: 5),
                    Expanded(
                      child: Text(
                        locationName != null && locationName.isNotEmpty
                            ? '$storeName • $locationName'
                            : storeName,
                        style: GoogleFonts.inter(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: Colors.white,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
              ],

              const SizedBox(height: 10),

              // Row 3: Methodology & Metrics Chips
              Wrap(
                spacing: 6,
                runSpacing: 4,
                children: [
                  _buildTagChip(countType, const Color(0xFF3B82F6)),
                  if (totalLines > 0)
                    _buildTagChip('$totalLines SKUs', Colors.white60),
                  if (varianceLines > 0)
                    _buildTagChip('$varianceLines Variances', accentAmber),
                  if (totalVarianceVal.abs() > 0.001)
                    _buildTagChip('\$${totalVarianceVal.toStringAsFixed(2)}', accentAmber),
                ],
              ),

              // Row 4: Operator & Scheduled Date (Only if present in SQLite)
              if ((countedBy != null && countedBy.isNotEmpty) || (scheduledDate != null && scheduledDate.isNotEmpty)) ...[
                const SizedBox(height: 10),
                const Divider(height: 1, color: Colors.white10),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    if (countedBy != null && countedBy.isNotEmpty)
                      Row(
                        children: [
                          const Icon(Icons.person_outline_rounded, size: 12, color: Colors.white38),
                          const SizedBox(width: 4),
                          Text(
                            countedBy,
                            style: GoogleFonts.inter(fontSize: 11, color: Colors.white54),
                          ),
                        ],
                      )
                    else
                      const SizedBox.shrink(),
                    if (scheduledDate != null && scheduledDate.isNotEmpty)
                      Text(
                        scheduledDate.length > 10 ? scheduledDate.substring(0, 10) : scheduledDate,
                        style: GoogleFonts.inter(fontSize: 11, color: Colors.white38),
                      ),
                  ],
                ),
              ],

              const SizedBox(height: 12),

              // Action CTA Button
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF4338CA),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                  ),
                  onPressed: () => _openStockCountExecution(session),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.qr_code_scanner_rounded, size: 16),
                      const SizedBox(width: 6),
                      Text(
                        'Start Physical Count / Inspection',
                        style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 12),
                      ),
                      const SizedBox(width: 4),
                      const Icon(Icons.arrow_forward_rounded, size: 14),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _openStockCountExecution(Map<String, dynamic> session) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => CountInitiationScreen(session: session),
      ),
    ).then((_) => _fetchStockCounts());
  }

  Widget _buildTagChip(String text, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: surfaceColor,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        text,
        style: GoogleFonts.inter(
          fontSize: 10,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }

  Widget _buildMobileEmptyState() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(vertical: 40, horizontal: 20),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: surfaceColor,
              shape: BoxShape.circle,
            ),
            child: const Icon(Icons.inventory_2_outlined, size: 36, color: Colors.white38),
          ),
          const SizedBox(height: 14),
          Text(
            'No Stock Counts Found',
            style: GoogleFonts.inter(
              fontSize: 15,
              fontWeight: FontWeight.w700,
              color: Colors.white,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            'No local database sessions match the selected filters.',
            textAlign: TextAlign.center,
            style: GoogleFonts.inter(
              fontSize: 12,
              color: Colors.white54,
            ),
          ),
        ],
      ),
    );
  }
}
