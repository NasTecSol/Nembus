import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:pos_mobile/screens/store_selection_screen.dart';
import '../database/db_service.dart';
import '../database/migration_runner.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
enum SyncPhase { idle, wiping, migrating, syncingData, completed, failed }

class DbCloningScreen extends StatefulWidget {
  final String tenantId;
  final String tenantSlug;
  final String tenantName;

  const DbCloningScreen({
    super.key,
    required this.tenantId,
    required this.tenantSlug,
    required this.tenantName,
  });

  @override
  State<DbCloningScreen> createState() => _DbCloningScreenState();
}

class _DbCloningScreenState extends State<DbCloningScreen> {
  SyncPhase _currentPhase = SyncPhase.idle;
  double _progress = 0.0;
  String _statusMessage = 'Initializing database synchronization...';
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _executeFullSyncWorkflow());
  }

  Future<void> _executeFullSyncWorkflow() async {
    try {
      setState(() {
        _currentPhase = SyncPhase.wiping;
        _progress = 0.15;
        _statusMessage = 'Purging existing local database structures...';
      });
      await DatabaseService().purgeDatabase();
      setState(() {
        _currentPhase = SyncPhase.migrating;
        _progress = 0.40;
        _statusMessage = 'Applying latest mobile database schema...';
      });
      await MigrationRunner.applyMigrations(
        onProgress: (step, prog) {
          setState(() {
            _progress = 0.40 + (prog * 0.30);
          });
        },
      );
      setState(() {
        _currentPhase = SyncPhase.syncingData;
        _progress = 0.75;
        _statusMessage = 'Pulling tenant master data snapshot via gRPC...';
      });

      final grpcAddress = SingletonClass().grpcAddress;
      final syncResult = NembusBridge().fetchCompleteTenantMasterData(
        widget.tenantSlug.isNotEmpty ? widget.tenantSlug : widget.tenantId,
        grpcAddress,
      );

      developer.log('FFI Sync Response: $syncResult', name: 'DbCloningScreen');

      if (syncResult['success'] == false) {
        developer.log('Data sync note: ${syncResult['error']}', name: 'DbCloningScreen');
      }

      // Step 4: Finalize & Transition
      setState(() {
        _currentPhase = SyncPhase.completed;
        _progress = 1.0;
        _statusMessage = 'Terminal initialized and ready.';
      });

      // Ensure any cloned/synced cashier sessions start closed
      try {
        final db = DatabaseService().database;
        await db.update('cashier_sessions', {'status': 'closed'}, where: "status = 'open'");
      } catch (_) {}
      await SingletonClass().closeCashierSession();

      // Persist tenant state to local disk so app restart skips setup
      await SingletonClass().saveTenantState(
        tenantId: widget.tenantId,
        tenantSlug: widget.tenantSlug,
        tenantName: widget.tenantName,
      );

      await Future.delayed(const Duration(milliseconds: 900));

      if (mounted) {
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (_) => StoreSelectionScreen(
              tenantId: widget.tenantId,
              tenantSlug: widget.tenantSlug,
              tenantName: widget.tenantName,
            ),
          ),
        );
      }
    } catch (e, stack) {
      developer.log('DB Cloning Error: $e', name: 'DbCloningScreen', error: e, stackTrace: stack);
      setState(() {
        _currentPhase = SyncPhase.failed;
        _errorMessage = e.toString();
        _statusMessage = 'Synchronization failed. Please retry.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    const bgColor = Color(0xFF0F172A);
    const cardColor = Color(0xFF1E293B);
    const accentColor = Color(0xFF38BDF8);

    return Scaffold(
      backgroundColor: bgColor,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 28.0, vertical: 24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Spacer(),
              Center(
                child: Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: cardColor,
                    shape: BoxShape.circle,
                    border: Border.all(color: accentColor.withValues(alpha: 0.4), width: 1.5),
                    boxShadow: [
                      BoxShadow(
                        color: accentColor.withValues(alpha: 0.2),
                        blurRadius: 24,
                        spreadRadius: 2,
                      ),
                    ],
                  ),
                  child: Icon(
                    _currentPhase == SyncPhase.completed
                        ? Icons.check_circle_outline_rounded
                        : _currentPhase == SyncPhase.failed
                        ? Icons.error_outline_rounded
                        : Icons.cloud_sync_rounded,
                    size: 48,
                    color: _currentPhase == SyncPhase.failed ? Colors.redAccent : accentColor,
                  ),
                ),
              ),
              const SizedBox(height: 24),
              Text(
                'Syncing Terminal Database',
                textAlign: TextAlign.center,
                style: GoogleFonts.inter(
                  fontSize: 22,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Configuring tenant: ${widget.tenantName}',
                textAlign: TextAlign.center,
                style: GoogleFonts.inter(
                  fontSize: 14,
                  color: Colors.white60,
                ),
              ),
              const SizedBox(height: 36),
              ClipRRect(
                borderRadius: BorderRadius.circular(10),
                child: LinearProgressIndicator(
                  value: _progress,
                  minHeight: 8,
                  backgroundColor: cardColor,
                  valueColor: AlwaysStoppedAnimation<Color>(
                    _currentPhase == SyncPhase.failed ? Colors.redAccent : accentColor,
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Text(
                _statusMessage,
                textAlign: TextAlign.center,
                style: GoogleFonts.inter(
                  fontSize: 13,
                  color: _currentPhase == SyncPhase.failed ? Colors.redAccent : Colors.white70,
                ),
              ),
              const Spacer(),
              if (_currentPhase == SyncPhase.failed) ...[
                ElevatedButton.icon(
                  onPressed: () {
                    setState(() {
                      _errorMessage = null;
                    });
                    _executeFullSyncWorkflow();
                  },
                  icon: const Icon(Icons.refresh_rounded, color: Colors.black),
                  label: Text(
                    'Retry Synchronization',
                    style: GoogleFonts.inter(fontWeight: FontWeight.bold, color: Colors.black),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: accentColor,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                ),
                const SizedBox(height: 16),
              ],
            ],
          ),
        ),
      ),
    );
  }
}