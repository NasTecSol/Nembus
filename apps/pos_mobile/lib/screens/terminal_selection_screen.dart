import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import 'login_screen.dart';

class TerminalSelectionScreen extends StatefulWidget {
  final String tenantSlug;
  final String tenantName;
  final String? tenantId;
  final int storeId;
  final String storeName;
  final String storeCode;

  const TerminalSelectionScreen({
    super.key,
    required this.tenantSlug,
    required this.tenantName,
    this.tenantId,
    required this.storeId,
    required this.storeName,
    required this.storeCode,
  });

  @override
  State<TerminalSelectionScreen> createState() => _TerminalSelectionScreenState();
}

class _TerminalSelectionScreenState extends State<TerminalSelectionScreen> {
  List<Map<String, dynamic>> _terminals = [];
  Map<String, dynamic>? _selectedTerminal;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _fetchTerminals();
  }

  Future<void> _fetchTerminals() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      List<Map<String, dynamic>> list = [];

      // 1. Fetch terminals via Go Core pos_terminals handler
      try {
        final resp = NembusBridge().callHandler(
          handler: 'pos_terminals',
          action: 'listPOSTerminalsByStore',
          payload: {'store_id': widget.storeId},
        );
        if (resp['success'] == true && resp['data'] is List && (resp['data'] as List).isNotEmpty) {
          list = (resp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
      } catch (e) {
        developer.log('Go terminal handler note: $e', name: 'TerminalSelectionScreen');
      }

      // 2. Fallback to direct SQLite query
      if (list.isEmpty) {
        final db = DatabaseService().database;
        final rows = await db.query(
          'pos_terminals',
          where: 'store_id = ?',
          whereArgs: [widget.storeId],
          orderBy: 'id ASC',
        );
        list = rows.map((r) => Map<String, dynamic>.from(r)).toList();
      }

      if (mounted) {
        setState(() {
          _terminals = list;
          _isLoading = false;
          _selectedTerminal = list.isNotEmpty ? list.first : null;
        });
      }
    } catch (e) {
      developer.log('Error fetching terminals: $e', name: 'TerminalSelectionScreen');
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Failed to load terminals: $e';
        });
      }
    }
  }

  Future<void> _createCustomTerminal() async {
    final nameCtrl = TextEditingController(text: '${widget.storeName} - Counter ${_terminals.length + 1}');
    final codeCtrl = TextEditingController(text: 'TERM-${widget.storeCode}-0${_terminals.length + 1}');

    final created = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: const Color(0xFF1E293B),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text(
          'Add New POS Terminal',
          style: GoogleFonts.inter(fontWeight: FontWeight.bold, color: Colors.white),
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameCtrl,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                labelText: 'Terminal Name',
                labelStyle: const TextStyle(color: Colors.white70),
                filled: true,
                fillColor: const Color(0xFF0F172A),
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: codeCtrl,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                labelText: 'Terminal Code',
                labelStyle: const TextStyle(color: Colors.white70),
                filled: true,
                fillColor: const Color(0xFF0F172A),
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel', style: TextStyle(color: Colors.white60)),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF38BDF8)),
            child: const Text('Add Terminal', style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );

    if (created == true) {
      try {
        final db = DatabaseService().database;
        final newId = await db.insert('pos_terminals', {
          'store_id': widget.storeId,
          'terminal_code': codeCtrl.text.trim(),
          'terminal_name': nameCtrl.text.trim(),
          'device_id': 'DEV-${DateTime.now().millisecondsSinceEpoch}',
          'is_active': 1,
          'created_at': DateTime.now().toIso8601String(),
          'updated_at': DateTime.now().toIso8601String(),
        });
        await _fetchTerminals();
        setState(() {
          _selectedTerminal = _terminals.firstWhere((t) => t['id'] == newId, orElse: () => _terminals.last);
        });
      } catch (e) {
        developer.log('Error adding terminal: $e', name: 'TerminalSelectionScreen');
      }
    }
  }

  Future<void> _proceedToLogin() async {
    if (_selectedTerminal == null) return;

    final terminalId = (_selectedTerminal!['id'] as num).toInt();
    final terminalName = _selectedTerminal!['terminal_name']?.toString() ?? 'Terminal #$terminalId';
    final terminalCode = _selectedTerminal!['terminal_code']?.toString() ?? 'TERM-$terminalId';

    // 1. Update local_device_config in SQLite
    try {
      final db = DatabaseService().database;
      await db.rawInsert('''
        INSERT OR REPLACE INTO local_device_config (id, device_name, store_id, pos_terminal_id, last_zatca_sync_at, zatca_enabled, created_at)
        VALUES (1, ?, ?, ?, '1970-01-01 00:00:00', 0, CURRENT_TIMESTAMP)
      ''', [terminalName, widget.storeId, terminalId]);
    } catch (e) {
      developer.log('Error saving local_device_config: $e', name: 'TerminalSelectionScreen');
    }

    // 2. Persist in SingletonClass state
    await SingletonClass().saveStoreAndTerminalState(
      storeId: widget.storeId,
      storeName: widget.storeName,
      storeCode: widget.storeCode,
      terminalId: terminalId,
      terminalName: terminalName,
      terminalCode: terminalCode,
    );

    if (mounted) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (_) => LoginScreen(
            tenantSlug: widget.tenantSlug,
            tenantName: widget.tenantName,
            storeId: widget.storeId,
            storeName: widget.storeName,
            posTerminalId: terminalId,
            posTerminalName: terminalName,
          ),
        ),
      );
    }
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
          'Select POS Terminal',
          style: GoogleFonts.inter(
            color: Colors.white,
            fontWeight: FontWeight.bold,
            fontSize: 18,
          ),
        ),
        actions: [
          IconButton(
            tooltip: 'Add Terminal',
            icon: const Icon(Icons.add_circle_outline_rounded, color: primarySky),
            onPressed: _createCustomTerminal,
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 12.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Store Context Header Card
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFF162032),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: primarySky.withValues(alpha: 0.3)),
                  boxShadow: [
                    BoxShadow(
                      color: primarySky.withValues(alpha: 0.1),
                      blurRadius: 12,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(colors: [primaryIndigo, primarySky]),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(Icons.store_rounded, color: Colors.white, size: 24),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            widget.storeName,
                            style: GoogleFonts.inter(
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                              fontSize: 15.5,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            'Store Code: ${widget.storeCode} • Tenant: ${widget.tenantSlug}',
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

              const SizedBox(height: 24),

              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Workstation Registers',
                    style: GoogleFonts.inter(
                      fontSize: 20,
                      fontWeight: FontWeight.w800,
                      color: Colors.white,
                      letterSpacing: -0.3,
                    ),
                  ),
                  TextButton.icon(
                    onPressed: _createCustomTerminal,
                    icon: const Icon(Icons.add_rounded, size: 16, color: primarySky),
                    label: const Text('Add Register', style: TextStyle(color: primarySky, fontSize: 13, fontWeight: FontWeight.bold)),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                'Select the specific checkout terminal to configure this hardware instance.',
                style: GoogleFonts.inter(fontSize: 13, color: Colors.white60),
              ),

              const SizedBox(height: 16),

              // Terminals List
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
                                  onPressed: _fetchTerminals,
                                  style: ElevatedButton.styleFrom(backgroundColor: primarySky),
                                  child: const Text('Retry', style: TextStyle(color: Colors.black)),
                                ),
                              ],
                            ),
                          )
                        : _terminals.isEmpty
                            ? Center(
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(vertical: 40),
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.point_of_sale_outlined,
                                        size: 56,
                                        color: Colors.white.withValues(alpha: 0.3),
                                      ),
                                      const SizedBox(height: 14),
                                      Text(
                                        'No terminals available for this store',
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.inter(
                                          fontSize: 16,
                                          fontWeight: FontWeight.w700,
                                          color: Colors.white,
                                        ),
                                      ),
                                      const SizedBox(height: 6),
                                      Text(
                                        'There are no workstation registers configured for ${widget.storeName}.',
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.inter(
                                          fontSize: 13,
                                          color: Colors.white60,
                                        ),
                                      ),
                                      const SizedBox(height: 18),
                                      Row(
                                        mainAxisAlignment: MainAxisAlignment.center,
                                        children: [
                                          ElevatedButton.icon(
                                            onPressed: _fetchTerminals,
                                            icon: const Icon(Icons.refresh_rounded, size: 18, color: Colors.black),
                                            label: const Text('Refresh', style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold)),
                                            style: ElevatedButton.styleFrom(backgroundColor: primarySky),
                                          ),
                                          const SizedBox(width: 12),
                                          OutlinedButton.icon(
                                            onPressed: _createCustomTerminal,
                                            icon: const Icon(Icons.add_rounded, size: 18, color: primarySky),
                                            label: const Text('Add Register', style: TextStyle(color: primarySky, fontWeight: FontWeight.bold)),
                                            style: OutlinedButton.styleFrom(
                                              side: const BorderSide(color: primarySky),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ],
                                  ),
                                ),
                              )
                            : ListView.separated(
                            physics: const BouncingScrollPhysics(),
                            itemCount: _terminals.length,
                            separatorBuilder: (_, __) => const SizedBox(height: 12),
                            itemBuilder: (context, index) {
                              final term = _terminals[index];
                              final isSelected = _selectedTerminal?['id'] == term['id'];
                              final name = term['terminal_name'] ?? 'Register #${term['id']}';
                              final code = term['terminal_code'] ?? 'TERM-${term['id']}';
                              final deviceId = term['device_id'] ?? 'Online Ready';

                              return InkWell(
                                onTap: () => setState(() => _selectedTerminal = term),
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
                                          Icons.point_of_sale_rounded,
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
                                                Container(
                                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                                  decoration: BoxDecoration(
                                                    color: emeraldGreen.withValues(alpha: 0.15),
                                                    borderRadius: BorderRadius.circular(6),
                                                  ),
                                                  child: const Text(
                                                    'READY',
                                                    style: TextStyle(
                                                      color: emeraldGreen,
                                                      fontSize: 9.5,
                                                      fontWeight: FontWeight.w800,
                                                    ),
                                                  ),
                                                ),
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

              // Confirm Terminal & Go to Login Button
              ElevatedButton(
                onPressed: _selectedTerminal != null ? _proceedToLogin : null,
                style: ElevatedButton.styleFrom(
                  backgroundColor: primarySky,
                  disabledBackgroundColor: Colors.white12,
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  elevation: _selectedTerminal != null ? 8 : 0,
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      'Confirm & Proceed to Login',
                      style: GoogleFonts.inter(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: _selectedTerminal != null ? Colors.black : Colors.white38,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Icon(
                      Icons.login_rounded,
                      color: _selectedTerminal != null ? Colors.black : Colors.white38,
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
