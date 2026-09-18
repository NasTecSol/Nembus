import 'dart:async';
import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import 'login_screen.dart';
import 'store_selection_screen.dart';
import 'tenant_selection_screen.dart';

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _animController;
  late Animation<double> _fadeAnimation;
  late Animation<double> _scaleAnimation;
  bool _isTenantConfigured = false;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    );

    _fadeAnimation =
        CurvedAnimation(parent: _animController, curve: Curves.easeIn);
    _scaleAnimation = Tween<double>(begin: 0.88, end: 1.0).animate(
      CurvedAnimation(parent: _animController, curve: Curves.easeOutCubic),
    );

    _animController.forward();
    _performStartupInit();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  Future<void> _performStartupInit() async {
    final stopwatch = Stopwatch()..start();
    try {
      // 1. Initialize local SQLite Engine (WAL & NORMAL sync)
      await DatabaseService().initDatabase();
      final dbPath = await DatabaseService().getDatabasePath();

      // Auto-close any lingering unclosed sessions left from a killed / terminated app instance
      await SingletonClass().autoCloseActiveSession(reason: 'Auto-closed on app launch from previous unclosed state');

      // 2. Initialize Go Core Engine via FFI
      try {
        final ffiRes = NembusBridge().initMobile(dbPath);
        developer.log('Go Engine Initialized: $ffiRes', name: 'SplashScreen');
      } catch (e) {
        developer.log('Go FFI Init Note: $e (proceeding with local runtime)',
            name: 'SplashScreen');
      }

      // 3. Check if tenant selection and master data have already been cloned
      final hasSavedTenant = await SingletonClass().loadTenantState();
      _isTenantConfigured = hasSavedTenant;

      if (hasSavedTenant) {
        SingletonClass().appMode = AppMode.offline;
        developer.log(
          'Existing tenant detected (${SingletonClass().activeTenantSlug}). Routing to LoginScreen.',
          name: 'SplashScreen',
        );
      } else {
        SingletonClass().appMode = AppMode.online;
        developer.log(
          'First run detected. Routing to TenantSelectionScreen.',
          name: 'SplashScreen',
        );
      }
    } catch (e, stack) {
      developer.log('Startup initialization failure: $e',
          name: 'SplashScreen', error: e, stackTrace: stack);
    } finally {
      stopwatch.stop();
      final elapsed = stopwatch.elapsedMilliseconds;
      final remaining = 1800 - elapsed;
      if (remaining > 0) {
        await Future.delayed(Duration(milliseconds: remaining));
      }
      if (mounted) {
        Widget targetScreen;
        if (_isTenantConfigured) {
          if (SingletonClass().activeStoreId != null && SingletonClass().activeTerminalId != null) {
            targetScreen = LoginScreen(
              tenantSlug: SingletonClass().activeTenantSlug,
              tenantName: SingletonClass().activeTenantName,
              storeId: SingletonClass().activeStoreId,
              storeName: SingletonClass().activeStoreName,
              posTerminalId: SingletonClass().activeTerminalId,
              posTerminalName: SingletonClass().activeTerminalName,
            );
          } else {
            targetScreen = StoreSelectionScreen(
              tenantSlug: SingletonClass().activeTenantSlug ?? 'default',
              tenantName: SingletonClass().activeTenantName ?? 'Store',
              tenantId: SingletonClass().activeTenantId,
            );
          }
        } else {
          targetScreen = const TenantSelectionScreen();
        }

        Navigator.of(context).pushReplacement(
          PageRouteBuilder(
            pageBuilder: (_, __, ___) => targetScreen,
            transitionsBuilder: (_, animation, __, child) =>
                FadeTransition(opacity: animation, child: child),
            transitionDuration: const Duration(milliseconds: 400),
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    const bgColor = Color(0xFF0F172A);
    const accentColor = Color(0xFF38BDF8);

    return Scaffold(
      backgroundColor: bgColor,
      body: Center(
        child: FadeTransition(
          opacity: _fadeAnimation,
          child: ScaleTransition(
            scale: _scaleAnimation,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    color: const Color(0xFF1E293B),
                    shape: BoxShape.circle,
                    boxShadow: [
                      BoxShadow(
                        color: accentColor.withValues(alpha: 0.25),
                        blurRadius: 32,
                        spreadRadius: 6,
                      ),
                    ],
                    border: Border.all(
                      color: accentColor.withValues(alpha: 0.6),
                      width: 2,
                    ),
                  ),
                  child: const Icon(
                    Icons.point_of_sale_rounded,
                    size: 64,
                    color: accentColor,
                  ),
                ),
                const SizedBox(height: 28),
                Text(
                  'N E M B U S',
                  style: GoogleFonts.inter(
                    fontSize: 28,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 6.0,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Enterprise Point of Sale',
                  style: GoogleFonts.inter(
                    fontSize: 14,
                    color: Colors.white60,
                    letterSpacing: 1.2,
                  ),
                ),
                const SizedBox(height: 48),
                const SizedBox(
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(
                    strokeWidth: 2.5,
                    color: accentColor,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}