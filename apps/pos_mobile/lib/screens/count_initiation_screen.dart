import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import 'stock_count_execution_screen.dart';

class CountInitiationScreen extends StatefulWidget {
  final Map<String, dynamic> session;

  const CountInitiationScreen({
    super.key,
    required this.session,
  });

  @override
  State<CountInitiationScreen> createState() => _CountInitiationScreenState();
}

class _CountInitiationScreenState extends State<CountInitiationScreen> {
  // Dark HUD palette matching POS Mobile design system
  static const Color bgColor = Color(0xFF090D16);
  static const Color cardBgColor = Color(0xFF131B2E);
  static const Color surfaceColor = Color(0xFF1E293B);
  static const Color primarySky = Color(0xFF38BDF8);
  static const Color accentIndigo = Color(0xFF4F46E5);
  static const Color accentEmerald = Color(0xFF10B981);
  static const Color accentAmber = Color(0xFFF59E0B);
  static const Color accentPurple = Color(0xFF8B5CF6);

  // Pre-flight safety protocols checklist
  bool _protocolScannerChecked = true;
  bool _protocolSafetyChecked = true;
  bool _protocolBinFreezeChecked = true;
  bool _protocolBlindAuditChecked = false;

  // Real data state from SQLite
  bool _isLoading = true;
  List<Map<String, dynamic>> _storagePartitions = [];
  int _totalSkuCount = 0;

  int get _readyProtocolsCount {
    int count = 0;
    if (_protocolScannerChecked) count++;
    if (_protocolSafetyChecked) count++;
    if (_protocolBinFreezeChecked) count++;
    if (_protocolBlindAuditChecked) count++;
    return count;
  }

  @override
  void initState() {
    super.initState();
    _loadRealData();
  }

  Future<void> _loadRealData() async {
    try {
      final db = DatabaseService().database;
      final storeId = widget.session['store_id'];

      // Query real storage locations for this store / session
      List<Map<String, dynamic>> locRows;
      if (storeId != null) {
        locRows = await db.query(
          'storage_locations',
          where: 'store_id = ?',
          whereArgs: [storeId],
          orderBy: 'name ASC',
        );
      } else {
        locRows = await db.query('storage_locations', orderBy: 'name ASC');
      }

      // Query SKUs count
      final sessionLines = (widget.session['total_lines'] as num?)?.toInt() ?? 0;
      int skus = sessionLines;
      if (skus == 0) {
        final prodRes = await db.rawQuery('SELECT COUNT(*) as count FROM products');
        if (prodRes.isNotEmpty) {
          skus = (prodRes.first['count'] as num?)?.toInt() ?? 0;
        }
      }

      if (mounted) {
        setState(() {
          _storagePartitions = List<Map<String, dynamic>>.from(locRows);
          _totalSkuCount = skus;
          _isLoading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final countNumber = widget.session['count_number']?.toString();
    final countType = widget.session['count_type']?.toString();
    final storeName = widget.session['store_name']?.toString();
    final locationName = widget.session['storage_location_name']?.toString();
    final status = widget.session['status']?.toString();

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
              'Inspection & Safety Scope',
              style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 16, color: Colors.white),
            ),
            Text(
              'Pre-Flight Security Protocols',
              style: GoogleFonts.inter(fontSize: 11, color: Colors.white54),
            ),
          ],
        ),
        actions: [
          if (countNumber != null && countNumber.isNotEmpty)
            Container(
              margin: const EdgeInsets.only(right: 14),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: accentIndigo.withValues(alpha: 0.25),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: const Color(0xFF6366F1).withValues(alpha: 0.4)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(width: 6, height: 6, decoration: const BoxDecoration(color: Color(0xFF818CF8), shape: BoxShape.circle)),
                  const SizedBox(width: 6),
                  Text(
                    countNumber,
                    style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 11, color: const Color(0xFFA5B4FC)),
                  ),
                ],
              ),
            ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: primarySky))
          : SafeArea(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 1. Hero Session Card
                    _buildHeroSessionCard(
                      countNumber: countNumber,
                      countType: countType,
                      storeName: storeName,
                      locationName: locationName,
                      status: status,
                      targetSKUs: _totalSkuCount,
                    ),

                    const SizedBox(height: 20),

                    // 2. Pre-Flight Protocols (Safety Measures)
                    _buildPreFlightProtocolsHeader(),
                    const SizedBox(height: 10),
                    _buildPreFlightProtocolsCard(),

                    const SizedBox(height: 20),

                    // 3. Allocated Partitions (Sub-Zones)
                    _buildAllocatedPartitionsHeader(),
                    const SizedBox(height: 10),
                    _buildAllocatedPartitionsList(),

                    const SizedBox(height: 20),

                    // 4. Deployed Inspector Card
                    _buildDeployedSquadCard(),

                    const SizedBox(height: 80), // spacing for bottom bar
                  ],
                ),
              ),
            ),
      bottomSheet: _buildBottomAction(context),
    );
  }

  /// 1. Hero Session Card
  Widget _buildHeroSessionCard({
    String? countNumber,
    String? countType,
    String? storeName,
    String? locationName,
    String? status,
    required int targetSKUs,
  }) {
    final displayTitle = (countType != null && countType.isNotEmpty)
        ? countType.replaceAll('_', ' ').toUpperCase()
        : 'PHYSICAL STOCK COUNT';

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              if (countType != null && countType.isNotEmpty) ...[
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: accentPurple.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text(
                    countType.replaceAll('_', ' ').toUpperCase(),
                    style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.w800, color: const Color(0xFFC084FC)),
                  ),
                ),
                const SizedBox(width: 6),
              ],
              if (status != null && status.isNotEmpty)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: surfaceColor,
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text(
                    status.replaceAll('_', ' ').toUpperCase(),
                    style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.w700, color: Colors.white70),
                  ),
                ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: accentIndigo.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(Icons.inventory_2_rounded, size: 16, color: Color(0xFF818CF8)),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            displayTitle,
            style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w800, color: Colors.white),
          ),
          if (storeName != null || locationName != null) ...[
            const SizedBox(height: 4),
            Row(
              children: [
                const Icon(Icons.location_on_outlined, color: primarySky, size: 14),
                const SizedBox(width: 4),
                Expanded(
                  child: Text(
                    [
                      if (storeName != null && storeName.isNotEmpty) storeName,
                      if (locationName != null && locationName.isNotEmpty) locationName,
                    ].join(' • '),
                    style: GoogleFonts.inter(fontSize: 12, color: Colors.white70),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ],
          const SizedBox(height: 16),

          // Two Metric Tiles
          Row(
            children: [
              Expanded(
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: surfaceColor,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: Colors.white.withValues(alpha: 0.05)),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Scope Metric', style: GoogleFonts.inter(fontSize: 10, color: Colors.white54, fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.baseline,
                        textBaseline: TextBaseline.alphabetic,
                        children: [
                          Text('$targetSKUs ', style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w800, color: Colors.white)),
                          Text('SKUs', style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.w700, color: const Color(0xFF818CF8))),
                        ],
                      ),
                      const SizedBox(height: 4),
                      Row(
                        children: [
                          const Icon(Icons.qr_code_scanner_rounded, size: 11, color: accentEmerald),
                          const SizedBox(width: 4),
                          Text('Barcoded Ready', style: GoogleFonts.inter(fontSize: 10, color: accentEmerald, fontWeight: FontWeight.w600)),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: surfaceColor,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: Colors.white.withValues(alpha: 0.05)),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Inspection State', style: GoogleFonts.inter(fontSize: 10, color: Colors.white54, fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.baseline,
                        textBaseline: TextBaseline.alphabetic,
                        children: [
                          Text('Active ', style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w800, color: Colors.white)),
                          Text('Session', style: GoogleFonts.inter(fontSize: 11, color: Colors.white54)),
                        ],
                      ),
                      const SizedBox(height: 4),
                      Row(
                        children: [
                          Container(width: 6, height: 6, decoration: const BoxDecoration(color: accentEmerald, shape: BoxShape.circle)),
                          const SizedBox(width: 4),
                          Text('Corridor Locked', style: GoogleFonts.inter(fontSize: 10, color: accentEmerald, fontWeight: FontWeight.w600)),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 14),

          // Corridor Lock Restriction Notice
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            decoration: BoxDecoration(
              color: const Color(0xFF1E1B4B),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: const Color(0xFF4338CA)),
            ),
            child: Row(
              children: [
                const Icon(Icons.lock_outline_rounded, color: Color(0xFFA5B4FC), size: 16),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Inbound dock receiving and ERP stock movement restricted during physical count.',
                    style: GoogleFonts.inter(fontSize: 11, color: const Color(0xFFC7D2FE), height: 1.3),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// 2. Pre-Flight Protocols Header
  Widget _buildPreFlightProtocolsHeader() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Row(
          children: [
            const Icon(Icons.verified_user_outlined, color: primarySky, size: 16),
            const SizedBox(width: 6),
            Text('Pre-Flight Protocols', style: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w800, color: Colors.white)),
          ],
        ),
        Text('$_readyProtocolsCount of 4 Ready', style: GoogleFonts.inter(fontSize: 12, color: accentEmerald, fontWeight: FontWeight.w700)),
      ],
    );
  }

  /// Pre-Flight Protocols Checklist Card
  Widget _buildPreFlightProtocolsCard() {
    return Container(
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        children: [
          _buildProtocolItem(
            icon: Icons.qr_code_scanner_rounded,
            title: 'Barcode & Camera Scanner',
            subtitle: 'Device camera & hardware scanner calibrated',
            badgeText: 'Ready',
            badgeColor: accentEmerald,
            isChecked: _protocolScannerChecked,
            onChanged: (val) => setState(() => _protocolScannerChecked = val ?? false),
          ),
          const Divider(height: 1, color: Colors.white10),
          _buildProtocolItem(
            icon: Icons.shield_outlined,
            title: 'Safety & PPE Verification',
            subtitle: 'High-visibility vest, safety boots & corridor clearance confirmed',
            isChecked: _protocolSafetyChecked,
            onChanged: (val) => setState(() => _protocolSafetyChecked = val ?? false),
          ),
          const Divider(height: 1, color: Colors.white10),
          _buildProtocolItem(
            icon: Icons.lock_clock_rounded,
            title: 'Bin Freeze & Quarantine',
            subtitle: 'Storage location inbound and picking lockout active',
            isChecked: _protocolBinFreezeChecked,
            onChanged: (val) => setState(() => _protocolBinFreezeChecked = val ?? false),
          ),
          const Divider(height: 1, color: Colors.white10),
          _buildProtocolItem(
            icon: Icons.visibility_off_outlined,
            title: 'Blind Audit Mode Mandate',
            subtitle: 'Expected stock quantities masked to prevent counting bias',
            isChecked: _protocolBlindAuditChecked,
            onChanged: (val) => setState(() => _protocolBlindAuditChecked = val ?? false),
          ),
        ],
      ),
    );
  }

  Widget _buildProtocolItem({
    required IconData icon,
    required String title,
    required String subtitle,
    String? badgeText,
    Color? badgeColor,
    required bool isChecked,
    required ValueChanged<bool?> onChanged,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: surfaceColor,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(icon, size: 18, color: const Color(0xFFA5B4FC)),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Flexible(
                      child: Text(
                        title,
                        style: GoogleFonts.inter(fontSize: 13, fontWeight: FontWeight.w700, color: Colors.white),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (badgeText != null) ...[
                      const SizedBox(width: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                        decoration: BoxDecoration(
                          color: (badgeColor ?? accentEmerald).withValues(alpha: 0.2),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(badgeText, style: GoogleFonts.inter(fontSize: 9, fontWeight: FontWeight.w800, color: badgeColor ?? accentEmerald)),
                      ),
                    ],
                  ],
                ),
                const SizedBox(height: 2),
                Text(
                  subtitle,
                  style: GoogleFonts.inter(fontSize: 11, color: Colors.white54),
                  overflow: TextOverflow.ellipsis,
                  maxLines: 2,
                ),
              ],
            ),
          ),
          Checkbox(
            value: isChecked,
            activeColor: const Color(0xFF4F46E5),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
            onChanged: onChanged,
          ),
        ],
      ),
    );
  }

  /// 3. Allocated Partitions Header
  Widget _buildAllocatedPartitionsHeader() {
    final count = _storagePartitions.isNotEmpty ? _storagePartitions.length : 1;
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Row(
          children: [
            const Icon(Icons.grid_view_rounded, color: primarySky, size: 16),
            const SizedBox(width: 6),
            Text('Allocated Partitions', style: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w800, color: Colors.white)),
          ],
        ),
        Text('$count Zone${count > 1 ? 's' : ''}', style: GoogleFonts.inter(fontSize: 12, color: Colors.white54)),
      ],
    );
  }

  /// Allocated Partitions List
  Widget _buildAllocatedPartitionsList() {
    if (_storagePartitions.isEmpty) {
      final locName = widget.session['storage_location_name']?.toString() ?? 'General Storage';
      return _buildPartitionCard(
        title: locName,
        subtitle: 'Primary Inspection Area',
        tag: 'Standard Zone',
      );
    }

    return Column(
      children: _storagePartitions.map((loc) {
        final name = loc['name']?.toString() ?? 'Storage Location';
        final code = loc['code']?.toString();
        final type = loc['type']?.toString();

        return Padding(
          padding: const EdgeInsets.only(bottom: 8),
          child: _buildPartitionCard(
            title: code != null && code.isNotEmpty ? '$name ($code)' : name,
            subtitle: type != null && type.isNotEmpty ? 'Type: ${type.toUpperCase()}' : 'Inspection Area',
            tag: 'Active Zone',
          ),
        );
      }).toList(),
    );
  }

  Widget _buildPartitionCard({
    required String title,
    required String subtitle,
    required String tag,
  }) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Row(
                  children: [
                    Container(width: 7, height: 7, decoration: const BoxDecoration(color: Color(0xFF6366F1), shape: BoxShape.circle)),
                    const SizedBox(width: 6),
                    Expanded(
                      child: Text(
                        title,
                        style: GoogleFonts.inter(fontSize: 13, fontWeight: FontWeight.w800, color: Colors.white),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: surfaceColor,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(tag, style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.w700, color: const Color(0xFFA5B4FC))),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Text(subtitle, style: GoogleFonts.inter(fontSize: 11, color: Colors.white54)),
        ],
      ),
    );
  }

  /// 4. Deployed Inspector Card
  Widget _buildDeployedSquadCard() {
    final countedByName = widget.session['counted_by_name']?.toString() ??
        widget.session['username']?.toString() ??
        'Active Inspector';
    final initial = countedByName.isNotEmpty ? countedByName[0].toUpperCase() : 'I';

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
                  const Icon(Icons.people_alt_outlined, color: primarySky, size: 16),
                  const SizedBox(width: 6),
                  Text('Assigned Personnel', style: GoogleFonts.inter(fontSize: 13, fontWeight: FontWeight.w800, color: Colors.white)),
                ],
              ),
              Text('Lead Inspector', style: GoogleFonts.inter(fontSize: 11, color: Colors.white54)),
            ],
          ),
          const SizedBox(height: 12),

          // Lead Inspector Row
          Row(
            children: [
              CircleAvatar(
                radius: 18,
                backgroundColor: const Color(0xFF4338CA),
                child: Text(initial, style: GoogleFonts.inter(fontWeight: FontWeight.w800, color: Colors.white, fontSize: 13)),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Flexible(
                          child: Text(
                            countedByName,
                            style: GoogleFonts.inter(fontSize: 13, fontWeight: FontWeight.w800, color: Colors.white),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        const SizedBox(width: 6),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                          decoration: BoxDecoration(
                            color: const Color(0xFF6366F1).withValues(alpha: 0.2),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text('LEAD', style: GoogleFonts.inter(fontSize: 9, fontWeight: FontWeight.w800, color: const Color(0xFFA5B4FC))),
                        ),
                      ],
                    ),
                    const SizedBox(height: 2),
                    Text('On-Station • Camera / Scanner Active', style: GoogleFonts.inter(fontSize: 11, color: Colors.white54)),
                  ],
                ),
              ),
              const Icon(Icons.check_circle_outline_rounded, color: accentEmerald, size: 18),
            ],
          ),
        ],
      ),
    );
  }

  /// Bottom Fixed Action Button
  Widget _buildBottomAction(BuildContext context) {
    return Container(
      color: bgColor,
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 14),
      child: SizedBox(
        width: double.infinity,
        child: ElevatedButton(
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF4338CA),
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(vertical: 16),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          ),
          onPressed: () {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => StockCountExecutionScreen(session: widget.session),
              ),
            );
          },
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.qr_code_scanner_rounded, size: 20),
              const SizedBox(width: 8),
              Flexible(
                child: Text(
                  'Start Physical Inspection & Scanning',
                  style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 14),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const SizedBox(width: 8),
              const Icon(Icons.arrow_forward_rounded, size: 18),
            ],
          ),
        ),
      ),
    );
  }
}
