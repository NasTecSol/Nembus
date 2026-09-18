import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import 'terminal_selection_screen.dart';

class StoreSelectionScreen extends StatefulWidget {
  final String tenantSlug;
  final String tenantName;
  final String? tenantId;

  const StoreSelectionScreen({
    super.key,
    required this.tenantSlug,
    required this.tenantName,
    this.tenantId,
  });

  @override
  State<StoreSelectionScreen> createState() => _StoreSelectionScreenState();
}

class _StoreSelectionScreenState extends State<StoreSelectionScreen> {
  final TextEditingController _searchController = TextEditingController();
  List<Map<String, dynamic>> _stores = [];
  List<Map<String, dynamic>> _filteredStores = [];
  Map<String, dynamic>? _selectedStore;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_filterStores);
    _fetchStores();
  }

  @override
  void dispose() {
    _searchController.removeListener(_filterStores);
    _searchController.dispose();
    super.dispose();
  }

  void _filterStores() {
    final query = _searchController.text.trim().toLowerCase();
    if (query.isEmpty) {
      setState(() => _filteredStores = List.from(_stores));
    } else {
      setState(() {
        _filteredStores = _stores.where((s) {
          final name = (s['name'] ?? '').toString().toLowerCase();
          final code = (s['code'] ?? '').toString().toLowerCase();
          final type = (s['store_type'] ?? '').toString().toLowerCase();
          return name.contains(query) || code.contains(query) || type.contains(query);
        }).toList();
      });
    }
  }

  Future<void> _fetchStores() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      List<Map<String, dynamic>> fetched = [];

      // 1. Try Go Core Store Handler via NembusBridge
      try {
        final resp = NembusBridge().callHandler(
          handler: 'stores',
          action: 'listStores',
          payload: {},
        );
        if (resp['success'] == true && resp['data'] is List && (resp['data'] as List).isNotEmpty) {
          fetched = (resp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
      } catch (e) {
        developer.log('Go store handler note: $e', name: 'StoreSelectionScreen');
      }

      // 2. Fallback to direct local SQLite query
      if (fetched.isEmpty) {
        final db = DatabaseService().database;
        final rows = await db.query('stores', orderBy: 'name ASC');
        fetched = rows.map((r) => Map<String, dynamic>.from(r)).toList();
      }

      if (mounted) {
        setState(() {
          _stores = fetched;
          _filteredStores = fetched;
          _isLoading = false;
          // Auto-select if only 1 store exists
          if (fetched.length == 1) {
            _selectedStore = fetched.first;
          }
        });
      }
    } catch (e) {
      developer.log('Error fetching stores: $e', name: 'StoreSelectionScreen');
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Failed to load stores: $e';
        });
      }
    }
  }

  void _proceedToTerminalSelection() {
    if (_selectedStore == null) return;

    final storeId = (_selectedStore!['id'] as num).toInt();
    final storeName = _selectedStore!['name']?.toString() ?? 'Store #$storeId';
    final storeCode = _selectedStore!['code']?.toString() ?? 'STR-$storeId';

    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => TerminalSelectionScreen(
          tenantSlug: widget.tenantSlug,
          tenantName: widget.tenantName,
          tenantId: widget.tenantId,
          storeId: storeId,
          storeName: storeName,
          storeCode: storeCode,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    const bgColor = Color(0xFF0F172A);
    const cardBg = Color(0xFF1E293B);
    const primarySky = Color(0xFF38BDF8);
    const primaryIndigo = Color(0xFF6366F1);
    const emeraldGreen = Color(0xFF10B981);

    return Scaffold(
      backgroundColor: bgColor,
      appBar: AppBar(
        backgroundColor: bgColor,
        elevation: 0,
        centerTitle: true,
        title: Text(
          'Select Store Branch',
          style: GoogleFonts.inter(
            color: Colors.white,
            fontWeight: FontWeight.bold,
            fontSize: 18,
          ),
        ),
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 12.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Tenant Context Header Badge
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                decoration: BoxDecoration(
                  color: primaryIndigo.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: primaryIndigo.withValues(alpha: 0.4)),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: primaryIndigo,
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: const Icon(Icons.business_rounded, color: Colors.white, size: 20),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            widget.tenantName,
                            style: GoogleFonts.inter(
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                              fontSize: 15,
                            ),
                          ),
                          Text(
                            'Tenant: ${widget.tenantSlug}',
                            style: const TextStyle(
                              color: Colors.white70,
                              fontSize: 12,
                              fontFamily: 'monospace',
                            ),
                          ),
                        ],
                      ),
                    ),
                    const Icon(Icons.check_circle_rounded, color: emeraldGreen, size: 20),
                  ],
                ),
              ),

              const SizedBox(height: 20),

              Text(
                'Available Store Locations',
                style: GoogleFonts.inter(
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                  color: Colors.white,
                  letterSpacing: -0.3,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                'Select the physical store location this POS device belongs to.',
                style: GoogleFonts.inter(fontSize: 13, color: Colors.white60),
              ),

              const SizedBox(height: 16),

              // Search Filter Bar
              Container(
                decoration: BoxDecoration(
                  color: cardBg,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                ),
                child: TextField(
                  controller: _searchController,
                  cursorColor: primarySky,
                  style: GoogleFonts.inter(color: Colors.white, fontSize: 14),
                  decoration: InputDecoration(
                    hintText: 'Search store by name or code...',
                    hintStyle: GoogleFonts.inter(color: Colors.white38, fontSize: 13),
                    prefixIcon: const Icon(Icons.search_rounded, color: primarySky, size: 22),
                    suffixIcon: _searchController.text.isNotEmpty
                        ? IconButton(
                            icon: const Icon(Icons.clear_rounded, color: Colors.white38, size: 18),
                            onPressed: () => _searchController.clear(),
                          )
                        : null,
                    border: InputBorder.none,
                    contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                  ),
                ),
              ),

              const SizedBox(height: 16),

              // Store List
              Expanded(
                child: _isLoading
                    ? const Center(
                        child: CircularProgressIndicator(color: primarySky, strokeWidth: 2.5),
                      )
                    : _errorMessage != null
                        ? Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(Icons.error_outline_rounded, color: Colors.redAccent, size: 40),
                                const SizedBox(height: 12),
                                Text(_errorMessage!, style: const TextStyle(color: Colors.white70)),
                                const SizedBox(height: 16),
                                ElevatedButton(
                                  onPressed: _fetchStores,
                                  style: ElevatedButton.styleFrom(backgroundColor: primarySky),
                                  child: const Text('Retry', style: TextStyle(color: Colors.black)),
                                ),
                              ],
                            ),
                          )
                        : _filteredStores.isEmpty
                            ? Center(
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(vertical: 40),
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.storefront_outlined,
                                        size: 56,
                                        color: Colors.white.withValues(alpha: 0.3),
                                      ),
                                      const SizedBox(height: 14),
                                      Text(
                                        _searchController.text.isNotEmpty
                                            ? 'No stores matching "${_searchController.text}"'
                                            : 'No stores available',
                                        style: GoogleFonts.inter(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w700,
                                          color: Colors.white,
                                        ),
                                      ),
                                      const SizedBox(height: 6),
                                      Text(
                                        _searchController.text.isNotEmpty
                                            ? 'Try searching with a different store name or code.'
                                            : 'No stores found. Please sync master data or contact your administrator.',
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.inter(
                                          fontSize: 13,
                                          color: Colors.white60,
                                        ),
                                      ),
                                      const SizedBox(height: 18),
                                      ElevatedButton.icon(
                                        onPressed: _fetchStores,
                                        icon: const Icon(Icons.refresh_rounded, size: 18, color: Colors.black),
                                        label: const Text('Refresh Stores', style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold)),
                                        style: ElevatedButton.styleFrom(backgroundColor: primarySky),
                                      ),
                                    ],
                                  ),
                                ),
                              )
                            : ListView.separated(
                                physics: const BouncingScrollPhysics(),
                                itemCount: _filteredStores.length,
                                separatorBuilder: (_, __) => const SizedBox(height: 12),
                                itemBuilder: (context, index) {
                                  final store = _filteredStores[index];
                                  final isSelected = _selectedStore?['id'] == store['id'];
                                  final name = store['name'] ?? 'Store';
                                  final code = store['code'] ?? 'STR';
                                  final storeType = (store['store_type'] ?? 'Retail').toString().toUpperCase();
                                  final isPosEnabled = store['is_pos_enabled'] == 1 || store['is_pos_enabled'] == true;

                                  return InkWell(
                                    onTap: () => setState(() => _selectedStore = store),
                                    borderRadius: BorderRadius.circular(18),
                                    child: AnimatedContainer(
                                      duration: const Duration(milliseconds: 200),
                                      padding: const EdgeInsets.all(16),
                                      decoration: BoxDecoration(
                                        color: isSelected ? const Color(0xFF162032) : cardBg,
                                        borderRadius: BorderRadius.circular(18),
                                        border: Border.all(
                                          color: isSelected ? primarySky : Colors.white.withValues(alpha: 0.08),
                                          width: isSelected ? 2 : 1,
                                        ),
                                        boxShadow: [
                                          BoxShadow(
                                            color: isSelected
                                                ? primarySky.withValues(alpha: 0.22)
                                                : Colors.black.withValues(alpha: 0.25),
                                            blurRadius: isSelected ? 14 : 6,
                                            offset: const Offset(0, 4),
                                          ),
                                        ],
                                      ),
                                      child: Row(
                                        children: [
                                          Container(
                                            padding: const EdgeInsets.all(12),
                                            decoration: BoxDecoration(
                                              gradient: isSelected
                                                  ? const LinearGradient(colors: [primaryIndigo, primarySky])
                                                  : null,
                                              color: isSelected ? null : const Color(0xFF0F172A),
                                              borderRadius: BorderRadius.circular(14),
                                            ),
                                            child: Icon(
                                              Icons.storefront_rounded,
                                              color: isSelected ? Colors.white : primarySky,
                                              size: 26,
                                            ),
                                          ),
                                          const SizedBox(width: 14),
                                          Expanded(
                                            child: Column(
                                              crossAxisAlignment: CrossAxisAlignment.start,
                                              children: [
                                                Text(
                                                  name,
                                                  style: GoogleFonts.inter(
                                                    fontWeight: FontWeight.w700,
                                                    color: Colors.white,
                                                    fontSize: 15.5,
                                                  ),
                                                ),
                                                const SizedBox(height: 6),
                                                Wrap(
                                                  spacing: 6,
                                                  runSpacing: 4,
                                                  crossAxisAlignment: WrapCrossAlignment.center,
                                                  children: [
                                                    Container(
                                                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                                      decoration: BoxDecoration(
                                                        color: Colors.white.withValues(alpha: 0.08),
                                                        borderRadius: BorderRadius.circular(6),
                                                      ),
                                                      child: Text(
                                                        code,
                                                        style: const TextStyle(
                                                          color: Colors.white70,
                                                          fontSize: 11,
                                                          fontFamily: 'monospace',
                                                          fontWeight: FontWeight.bold,
                                                        ),
                                                      ),
                                                    ),
                                                    Text(
                                                      storeType,
                                                      style: TextStyle(
                                                        color: Colors.grey.shade400,
                                                        fontSize: 11,
                                                        fontWeight: FontWeight.w600,
                                                      ),
                                                    ),
                                                    if (isPosEnabled) ...[
                                                      Container(
                                                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                                        decoration: BoxDecoration(
                                                          color: emeraldGreen.withValues(alpha: 0.15),
                                                          borderRadius: BorderRadius.circular(6),
                                                        ),
                                                        child: const Text(
                                                          'POS ACTIVE',
                                                          style: TextStyle(
                                                            color: emeraldGreen,
                                                            fontSize: 9.5,
                                                            fontWeight: FontWeight.w800,
                                                          ),
                                                        ),
                                                      ),
                                                    ],
                                                  ],
                                                ),
                                              ],
                                            ),
                                          ),
                                          Icon(
                                            isSelected
                                                ? Icons.radio_button_checked_rounded
                                                : Icons.radio_button_unchecked_rounded,
                                            color: isSelected ? primarySky : Colors.white38,
                                            size: 24,
                                          ),
                                        ],
                                      ),
                                    ),
                                  );
                                },
                              ),
              ),

              const SizedBox(height: 16),

              // Proceed Action Button
              ElevatedButton(
                onPressed: _selectedStore != null ? _proceedToTerminalSelection : null,
                style: ElevatedButton.styleFrom(
                  backgroundColor: primarySky,
                  disabledBackgroundColor: Colors.white12,
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  elevation: _selectedStore != null ? 8 : 0,
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      'Next: Select Terminal',
                      style: GoogleFonts.inter(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: _selectedStore != null ? Colors.black : Colors.white38,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Icon(
                      Icons.arrow_forward_rounded,
                      color: _selectedStore != null ? Colors.black : Colors.white38,
                      size: 20,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
