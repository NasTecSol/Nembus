import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import 'pos_dashboard_screen.dart';
import 'store_selection_screen.dart';

class LoginScreen extends StatefulWidget {
  final String? tenantSlug;
  final String? tenantName;
  final int? storeId;
  final String? storeName;
  final int? posTerminalId;
  final String? posTerminalName;

  const LoginScreen({
    super.key,
    this.tenantSlug,
    this.tenantName,
    this.storeId,
    this.storeName,
    this.posTerminalId,
    this.posTerminalName,
  });

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> with SingleTickerProviderStateMixin {
  final TextEditingController _searchController = TextEditingController();

  List<Map<String, dynamic>> _cashiers = [];
  List<Map<String, dynamic>> _filteredCashiers = [];

  bool _isLoadingCashiers = true;
  String? _errorMessage;

  late AnimationController _idleAnimController;

  @override
  void initState() {
    super.initState();
    _idleAnimController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 4),
    )..repeat(reverse: true);

    _searchController.addListener(_filterCashiers);
    _fetchCashiers();
  }

  @override
  void dispose() {
    _idleAnimController.dispose();
    _searchController.removeListener(_filterCashiers);
    _searchController.dispose();
    super.dispose();
  }

  int get _effectiveStoreId =>
      widget.storeId ?? SingletonClass().activeStoreId ?? 1;

  String get _effectiveStoreName =>
      widget.storeName ?? SingletonClass().activeStoreName ?? 'Main Branch';

  int get _effectiveTerminalId =>
      widget.posTerminalId ?? SingletonClass().activeTerminalId ?? 1;

  String get _effectiveTerminalName =>
      widget.posTerminalName ?? SingletonClass().activeTerminalName ?? 'Register 1';

  String get _effectiveTenantSlug =>
      widget.tenantSlug ?? SingletonClass().activeTenantSlug ?? 'default';

  String get _effectiveTenantName =>
      widget.tenantName ?? SingletonClass().activeTenantName ?? 'POS System';

  void _filterCashiers() {
    final query = _searchController.text.trim().toLowerCase();
    if (query.isEmpty) {
      setState(() => _filteredCashiers = List.from(_cashiers));
    } else {
      setState(() {
        _filteredCashiers = _cashiers.where((c) {
          final code = (c['cashier_code'] ?? '').toString().toLowerCase();
          final username = (c['username'] ?? '').toString().toLowerCase();
          final name = '${c['first_name'] ?? ''} ${c['last_name'] ?? ''}'.toLowerCase();
          return code.contains(query) || username.contains(query) || name.contains(query);
        }).toList();
      });
    }
  }

  /// Fetches cashiers assigned to the active store/terminal via Go Core Handlers & local SQLite
  Future<void> _fetchCashiers() async {
    setState(() {
      _isLoadingCashiers = true;
      _errorMessage = null;
    });

    try {
      List<Map<String, dynamic>> list = [];

      try {
        final resp = NembusBridge().callHandler(
          handler: 'cashiers',
          action: 'listActiveCashiersByStore',
          payload: {'store_id': _effectiveStoreId},
        );

        if (resp['success'] == true && resp['data'] is List && (resp['data'] as List).isNotEmpty) {
          final rawList = (resp['data'] as List).map((e) => Map<String, dynamic>.from(e as Map)).toList();
          list = rawList;
        }
      } catch (e) {
        developer.log('Go cashier handler fetch note: $e', name: 'LoginScreen');
      }

      // 2. Query local SQLite for enriched user info (joining cashiers and users)
      if (list.isEmpty) {
        try {
          final db = DatabaseService().database;
          final enrichedRows = await db.rawQuery('''
            SELECT 
              c.id AS cashier_id,
              c.user_id,
              c.store_id,
              c.cashier_code,
              c.drawer_limit,
              c.discount_limit,
              c.is_active,
              u.username,
              u.first_name,
              u.last_name,
              u.email,
              u.employee_code
            FROM cashiers c
            LEFT JOIN users u ON c.user_id = u.id
            WHERE c.store_id = ? AND c.is_active = 1
            ORDER BY c.cashier_code ASC
          ''', [_effectiveStoreId]);

          if (enrichedRows.isNotEmpty) {
            list = enrichedRows.map((r) => Map<String, dynamic>.from(r)).toList();
          }
        } catch (e) {
          developer.log('SQLite cashier query note: $e', name: 'LoginScreen');
        }
      }

      if (mounted) {
        setState(() {
          _cashiers = list;
          _filteredCashiers = list;
          _isLoadingCashiers = false;
        });
      }
    } catch (e) {
      developer.log('Error loading cashiers: $e', name: 'LoginScreen');
      if (mounted) {
        setState(() {
          _isLoadingCashiers = false;
          _errorMessage = 'Failed to load cashiers: $e';
        });
      }
    }
  }

  /// Opens the Password Authentication modal for the selected cashier
  void _onCashierSelected(Map<String, dynamic> cashier) {
    _showPasswordAuthModal(cashier);
  }

  /// Shows the password authentication bottom sheet with autofilled username
  void _showPasswordAuthModal(Map<String, dynamic> cashier) {
    const cardSurface = Color(0xFF161F30);
    const primarySky = Color(0xFF38BDF8);
    const primaryIndigo = Color(0xFF6366F1);
    const emeraldGreen = Color(0xFF10B981);

    final firstName = cashier['first_name']?.toString() ?? 'Cashier';
    final lastName = cashier['last_name']?.toString() ?? '';
    final cashierName = '$firstName $lastName'.trim();
    final cashierCode = cashier['cashier_code']?.toString() ?? 'CSH-01';
    final initialUsername = cashier['username']?.toString() ?? cashierCode;

    final usernameCtrl = TextEditingController(text: initialUsername);
    final passwordCtrl = TextEditingController();
    bool obscurePassword = true;
    bool isAuthenticating = false;
    String? authErrorMessage;

    final initials = (firstName.isNotEmpty ? firstName[0] : 'C') +
        (lastName.isNotEmpty ? lastName[0] : '');

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (modalCtx, modalSetState) {
          Future<void> handleLogin() async {
            final username = usernameCtrl.text.trim();
            final password = passwordCtrl.text;

            if (username.isEmpty) {
              modalSetState(() => authErrorMessage = 'Username cannot be empty');
              return;
            }
            if (password.isEmpty) {
              modalSetState(() => authErrorMessage = 'Please enter password');
              return;
            }

            modalSetState(() {
              isAuthenticating = true;
              authErrorMessage = null;
            });

            try {
              // Call Go Core auth handler via NembusBridge
              var response = NembusBridge().callHandler(
                handler: 'auth',
                action: 'login',
                payload: {
                  'user_login': username,
                  'password': password,
                },
              );

              developer.log('Auth handler response for $username: $response', name: 'LoginScreen');

              // If initial attempt fails and cashier has email, try with email
              if (response['success'] != true && cashier['email'] != null && cashier['email'].toString().isNotEmpty) {
                response = NembusBridge().callHandler(
                  handler: 'auth',
                  action: 'login',
                  payload: {
                    'user_login': cashier['email'].toString(),
                    'password': password,
                  },
                );
              }

              if (response['success'] == true) {
                final cashierId = (cashier['cashier_id'] ?? cashier['id'] ?? 1) as int;
                final userId = (cashier['user_id'] ?? 1) as int;

                // Persist session configuration in SingletonClass
                await SingletonClass().saveStoreAndTerminalState(
                  storeId: _effectiveStoreId,
                  storeName: _effectiveStoreName,
                  storeCode: 'STR-$_effectiveStoreId',
                  terminalId: _effectiveTerminalId,
                  terminalName: _effectiveTerminalName,
                  terminalCode: 'TERM-$_effectiveTerminalId',
                );
                SingletonClass().activeCashierId = cashierId;
                SingletonClass().activeCashierCode = cashierCode;

                // Ensure cashier session is closed on login so the cashier must explicitly start a shift
                try {
                  await SingletonClass().closeCashierSession();
                  final db = DatabaseService().database;
                  await db.update(
                    'cashier_sessions',
                    {
                      'status': 'closed',
                      'closing_time': DateTime.now().toIso8601String(),
                      'updated_at': DateTime.now().toIso8601String(),
                    },
                    where: 'status = ?',
                    whereArgs: ['open'],
                  );
                  developer.log('Ensured all cashier sessions are closed on login', name: 'LoginScreen');
                } catch (e) {
                  developer.log('Reset open session on login note: $e', name: 'LoginScreen');
                }

                if (mounted) {
                  Navigator.of(context).pop(); // Close modal

                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      backgroundColor: emeraldGreen,
                      behavior: SnackBarBehavior.floating,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      content: Row(
                        children: [
                          const Icon(Icons.check_circle_outline, color: Colors.white),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Text(
                              'Welcome, $cashierName! ($cashierCode)',
                              style: GoogleFonts.inter(fontWeight: FontWeight.w600, color: Colors.white),
                            ),
                          ),
                        ],
                      ),
                    ),
                  );

                  Navigator.of(context).pushAndRemoveUntil(
                    MaterialPageRoute(
                      builder: (_) => PosDashboardScreen(
                        tenantSlug: _effectiveTenantSlug,
                        tenantName: _effectiveTenantName,
                        userId: userId,
                        username: username,
                        roleName: 'Cashier',
                        storeId: _effectiveStoreId,
                        storeName: _effectiveStoreName,
                        posTerminalId: _effectiveTerminalId,
                        posTerminalName: _effectiveTerminalName,
                      ),
                    ),
                    (route) => false,
                  );
                }
              } else {
                final err = response['error']?.toString() ?? 'Invalid username or password';
                modalSetState(() {
                  isAuthenticating = false;
                  authErrorMessage = err;
                });
              }
            } catch (e) {
              modalSetState(() {
                isAuthenticating = false;
                authErrorMessage = 'Authentication error: $e';
              });
            }
          }

          return Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 460),
              child: Container(
                margin: EdgeInsets.only(
                  bottom: MediaQuery.of(modalCtx).viewInsets.bottom + 20,
                  left: 16,
                  right: 16,
                  top: 40,
                ),
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  color: cardSurface,
                  borderRadius: BorderRadius.circular(28),
                  border: Border.all(color: primarySky.withValues(alpha: 0.3), width: 1.5),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.7),
                      blurRadius: 36,
                      offset: const Offset(0, 16),
                    ),
                  ],
                ),
                child: SingleChildScrollView(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // Header Drag Handle
                      Center(
                        child: Container(
                          width: 40,
                          height: 4,
                          decoration: BoxDecoration(
                            color: Colors.white24,
                            borderRadius: BorderRadius.circular(2),
                          ),
                        ),
                      ),
                      const SizedBox(height: 18),

                      // Selected Cashier Avatar & Info Card
                      Row(
                        children: [
                          Container(
                            width: 52,
                            height: 52,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              gradient: const LinearGradient(
                                colors: [primarySky, primaryIndigo],
                                begin: Alignment.topLeft,
                                end: Alignment.bottomRight,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: primarySky.withValues(alpha: 0.3),
                                  blurRadius: 10,
                                  offset: const Offset(0, 3),
                                ),
                              ],
                            ),
                            child: Center(
                              child: Text(
                                initials.toUpperCase(),
                                style: GoogleFonts.inter(
                                  fontSize: 18,
                                  fontWeight: FontWeight.w900,
                                  color: Colors.white,
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 14),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  cashierName.isNotEmpty ? cashierName : 'Cashier Operator',
                                  style: GoogleFonts.inter(
                                    fontSize: 17,
                                    fontWeight: FontWeight.w800,
                                    color: Colors.white,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  '$cashierCode • $_effectiveStoreName • $_effectiveTerminalName',
                                  style: GoogleFonts.inter(fontSize: 12, color: Colors.white60),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),

                      const SizedBox(height: 22),

                      // Username Field (Autofilled)
                      Text(
                        'Cashier Username',
                        style: GoogleFonts.inter(
                          fontSize: 12.5,
                          fontWeight: FontWeight.w600,
                          color: Colors.white70,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Container(
                        decoration: BoxDecoration(
                          color: const Color(0xFF0F172A),
                          borderRadius: BorderRadius.circular(14),
                          border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
                        ),
                        child: TextField(
                          controller: usernameCtrl,
                          style: GoogleFonts.inter(color: Colors.white, fontSize: 14.5, fontWeight: FontWeight.w600),
                          decoration: const InputDecoration(
                            prefixIcon: Icon(Icons.alternate_email_rounded, color: primarySky, size: 20),
                            border: InputBorder.none,
                            contentPadding: EdgeInsets.symmetric(horizontal: 14, vertical: 14),
                          ),
                        ),
                      ),

                      const SizedBox(height: 16),

                      // Password Field
                      Text(
                        'Password',
                        style: GoogleFonts.inter(
                          fontSize: 12.5,
                          fontWeight: FontWeight.w600,
                          color: Colors.white70,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Container(
                        decoration: BoxDecoration(
                          color: const Color(0xFF0F172A),
                          borderRadius: BorderRadius.circular(14),
                          border: Border.all(
                            color: authErrorMessage != null
                                ? const Color(0xFFEF4444)
                                : Colors.white.withValues(alpha: 0.1),
                          ),
                        ),
                        child: TextField(
                          controller: passwordCtrl,
                          autofocus: true,
                          obscureText: obscurePassword,
                          style: GoogleFonts.inter(color: Colors.white, fontSize: 14.5),
                          textInputAction: TextInputAction.done,
                          onSubmitted: (_) => handleLogin(),
                          onChanged: (_) {
                            if (authErrorMessage != null) {
                              modalSetState(() => authErrorMessage = null);
                            }
                          },
                          decoration: InputDecoration(
                            prefixIcon: const Icon(Icons.lock_outline_rounded, color: primarySky, size: 20),
                            hintText: 'Enter password',
                            hintStyle: GoogleFonts.inter(color: Colors.white30, fontSize: 13.5),
                            border: InputBorder.none,
                            contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
                            suffixIcon: IconButton(
                              icon: Icon(
                                obscurePassword ? Icons.visibility_outlined : Icons.visibility_off_outlined,
                                color: Colors.white54,
                                size: 20,
                              ),
                              onPressed: () => modalSetState(() => obscurePassword = !obscurePassword),
                            ),
                          ),
                        ),
                      ),

                      // Error message if any
                      if (authErrorMessage != null) ...[
                        const SizedBox(height: 12),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                          decoration: BoxDecoration(
                            color: const Color(0xFFEF4444).withValues(alpha: 0.15),
                            borderRadius: BorderRadius.circular(10),
                            border: Border.all(color: const Color(0xFFEF4444).withValues(alpha: 0.4)),
                          ),
                          child: Row(
                            children: [
                              const Icon(Icons.error_outline_rounded, size: 16, color: Color(0xFFEF4444)),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  authErrorMessage!,
                                  style: GoogleFonts.inter(color: const Color(0xFFEF4444), fontSize: 12, fontWeight: FontWeight.w600),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],

                      const SizedBox(height: 22),

                      // Authenticate / Sign In Button
                      SizedBox(
                        height: 50,
                        child: ElevatedButton(
                          onPressed: isAuthenticating ? null : handleLogin,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: emeraldGreen,
                            foregroundColor: Colors.white,
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                            elevation: 6,
                          ),
                          child: isAuthenticating
                              ? const SizedBox(
                                  width: 22,
                                  height: 22,
                                  child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2.5),
                                )
                              : Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    const Icon(Icons.lock_open_rounded, size: 20),
                                    const SizedBox(width: 8),
                                    Text(
                                      'Sign In & Start Shift',
                                      style: GoogleFonts.inter(fontSize: 15, fontWeight: FontWeight.w800),
                                    ),
                                  ],
                                ),
                        ),
                      ),

                      const SizedBox(height: 10),
                      TextButton(
                        onPressed: () => Navigator.pop(modalCtx),
                        child: Text(
                          'Cancel / Choose Another Cashier',
                          style: GoogleFonts.inter(color: Colors.white60, fontSize: 13),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    const bgDark = Color(0xFF0B0F19);
    const cardSurface = Color(0xFF161F30);
    const primarySky = Color(0xFF38BDF8);
    const primaryIndigo = Color(0xFF6366F1);

    return Scaffold(
      backgroundColor: bgDark,
      body: Stack(
        children: [
          Positioned.fill(child: _build3DAmbientBackground()),
          SafeArea(
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 960),
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 14.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      _buildHeaderBar(primarySky, primaryIndigo),
                      const SizedBox(height: 18),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Select Cashier Operator',
                                  style: GoogleFonts.inter(
                                    fontSize: 22,
                                    fontWeight: FontWeight.w900,
                                    color: Colors.white,
                                    letterSpacing: -0.3,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  'Tap your profile and enter password to start terminal shift.',
                                  style: GoogleFonts.inter(fontSize: 13, color: Colors.white60),
                                ),
                              ],
                            ),
                          ),
                          IconButton(
                            tooltip: 'Refresh Cashiers',
                            icon: const Icon(Icons.refresh_rounded, color: primarySky),
                            onPressed: _fetchCashiers,
                          ),
                        ],
                      ),
                      const SizedBox(height: 14),
                      Container(
                        decoration: BoxDecoration(
                          color: cardSurface,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                        ),
                        child: TextField(
                          controller: _searchController,
                          cursorColor: primarySky,
                          style: GoogleFonts.inter(color: Colors.white, fontSize: 14),
                          decoration: InputDecoration(
                            hintText: 'Search cashier by name, code or username...',
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
                      Expanded(
                        child: _isLoadingCashiers
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
                                          onPressed: _fetchCashiers,
                                          style: ElevatedButton.styleFrom(backgroundColor: primarySky),
                                          child: const Text('Retry', style: TextStyle(color: Colors.black)),
                                        ),
                                      ],
                                    ),
                                  )
                                : _filteredCashiers.isEmpty
                                    ? Center(
                                        child: Padding(
                                          padding: const EdgeInsets.symmetric(vertical: 40),
                                          child: Column(
                                            mainAxisAlignment: MainAxisAlignment.center,
                                            children: [
                                              Icon(
                                                Icons.badge_outlined,
                                                size: 56,
                                                color: Colors.white.withValues(alpha: 0.3),
                                              ),
                                              const SizedBox(height: 14),
                                              Text(
                                                _searchController.text.isNotEmpty
                                                    ? 'No cashiers matching "${_searchController.text}"'
                                                    : 'No cashiers available',
                                                style: GoogleFonts.inter(
                                                  fontSize: 16,
                                                  fontWeight: FontWeight.w700,
                                                  color: Colors.white,
                                                ),
                                              ),
                                              const SizedBox(height: 6),
                                              Text(
                                                _searchController.text.isNotEmpty
                                                    ? 'Try searching with a different cashier code or name.'
                                                    : 'There are no active cashier operators configured for $_effectiveStoreName.',
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
                                                    onPressed: _fetchCashiers,
                                                    icon: const Icon(Icons.refresh_rounded, size: 18, color: Colors.black),
                                                    label: const Text('Refresh', style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold)),
                                                    style: ElevatedButton.styleFrom(backgroundColor: primarySky),
                                                  ),
                                                  const SizedBox(width: 12),
                                                  OutlinedButton.icon(
                                                    onPressed: () {
                                                      Navigator.of(context).push(
                                                        MaterialPageRoute(
                                                          builder: (_) => StoreSelectionScreen(
                                                            tenantSlug: _effectiveTenantSlug,
                                                            tenantName: _effectiveTenantName,
                                                            tenantId: SingletonClass().activeTenantId,
                                                          ),
                                                        ),
                                                      );
                                                    },
                                                    icon: const Icon(Icons.swap_horiz_rounded, size: 18, color: primarySky),
                                                    label: const Text('Change Store / Terminal', style: TextStyle(color: primarySky, fontWeight: FontWeight.bold)),
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
                                    : LayoutBuilder(
                                        builder: (context, constraints) {
                                          final isTablet = constraints.maxWidth >= 600;
                                          final crossAxisCount = isTablet ? (constraints.maxWidth >= 850 ? 4 : 3) : 2;
                                          final childAspectRatio = isTablet ? 1.15 : 0.95;

                                          return GridView.builder(
                                            physics: const BouncingScrollPhysics(),
                                            gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                                              crossAxisCount: crossAxisCount,
                                              mainAxisSpacing: 14,
                                              crossAxisSpacing: 14,
                                              childAspectRatio: childAspectRatio,
                                            ),
                                            itemCount: _filteredCashiers.length,
                                            itemBuilder: (context, index) {
                                              final cashier = _filteredCashiers[index];
                                              return _buildCashierCard(cashier, cardSurface, primarySky, primaryIndigo);
                                            },
                                          );
                                        },
                                      ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeaderBar(Color primarySky, Color primaryIndigo) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF161F30).withValues(alpha: 0.9),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.4),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              gradient: LinearGradient(colors: [primaryIndigo, primarySky]),
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(Icons.point_of_sale_rounded, color: Colors.white, size: 20),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'NEMBUS POS',
                  style: GoogleFonts.inter(
                    fontWeight: FontWeight.w900,
                    fontSize: 14,
                    letterSpacing: 1.5,
                    color: Colors.white,
                  ),
                ),
                Text(
                  '$_effectiveStoreName • $_effectiveTerminalName',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: GoogleFonts.inter(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: primarySky,
                  ),
                ),
              ],
            ),
          ),
          InkWell(
            onTap: () {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => StoreSelectionScreen(
                    tenantSlug: _effectiveTenantSlug,
                    tenantName: _effectiveTenantName,
                    tenantId: SingletonClass().activeTenantId,
                  ),
                ),
              );
            },
            borderRadius: BorderRadius.circular(10),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
              decoration: BoxDecoration(
                color: primarySky.withValues(alpha: 0.18),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: primarySky.withValues(alpha: 0.4)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.swap_horiz_rounded, size: 15, color: Color(0xFF38BDF8)),
                  const SizedBox(width: 4),
                  Text(
                    'CHANGE',
                    style: GoogleFonts.inter(
                      fontSize: 10.5,
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF38BDF8),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCashierCard(
    Map<String, dynamic> cashier,
    Color cardSurface,
    Color primarySky,
    Color primaryIndigo,
  ) {
    final firstName = cashier['first_name']?.toString() ?? 'Cashier';
    final lastName = cashier['last_name']?.toString() ?? '';
    final fullName = '$firstName $lastName'.trim();
    final code = cashier['cashier_code']?.toString() ?? 'CSH-01';
    final username = cashier['username']?.toString() ?? 'operator';
    final initials = (firstName.isNotEmpty ? firstName[0] : 'C') +
        (lastName.isNotEmpty ? lastName[0] : '');

    return InkWell(
      onTap: () => _onCashierSelected(cashier),
      borderRadius: BorderRadius.circular(22),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
        decoration: BoxDecoration(
          color: cardSurface,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.35),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          mainAxisSize: MainAxisSize.min,
          children: [
            Stack(
              alignment: Alignment.center,
              children: [
                Container(
                  width: 52,
                  height: 52,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: const LinearGradient(
                      colors: [Color(0xFF38BDF8), Color(0xFF6366F1)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: primarySky.withValues(alpha: 0.3),
                        blurRadius: 8,
                        offset: const Offset(0, 3),
                      ),
                    ],
                  ),
                  child: Center(
                    child: Text(
                      initials.toUpperCase(),
                      style: GoogleFonts.inter(
                        fontSize: 18,
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
                Positioned(
                  bottom: 0,
                  right: 0,
                  child: Container(
                    width: 13,
                    height: 13,
                    decoration: BoxDecoration(
                      color: const Color(0xFF10B981),
                      shape: BoxShape.circle,
                      border: Border.all(color: cardSurface, width: 2),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              fullName.isNotEmpty ? fullName : 'Operator $code',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: GoogleFonts.inter(
                fontSize: 13.5,
                fontWeight: FontWeight.w800,
                color: Colors.white,
              ),
            ),
            const SizedBox(height: 2),
            Text(
              '$code • @$username',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: GoogleFonts.inter(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: Colors.white54,
              ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: primarySky.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: primarySky.withValues(alpha: 0.3)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.lock_open_rounded, size: 11, color: Color(0xFF38BDF8)),
                  const SizedBox(width: 4),
                  Text(
                    'Sign In',
                    style: GoogleFonts.inter(
                      fontSize: 10.5,
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF38BDF8),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _build3DAmbientBackground() {
    return Stack(
      children: [
        Positioned(
          top: -100,
          left: -80,
          child: AnimatedBuilder(
            animation: _idleAnimController,
            builder: (context, child) {
              return Container(
                width: 320,
                height: 320,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF38BDF8).withValues(alpha: 0.15 + (_idleAnimController.value * 0.08)),
                      Colors.transparent,
                    ],
                  ),
                ),
              );
            },
          ),
        ),
        Positioned(
          bottom: -120,
          right: -80,
          child: AnimatedBuilder(
            animation: _idleAnimController,
            builder: (context, child) {
              return Container(
                width: 340,
                height: 340,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF6366F1).withValues(alpha: 0.15 + ((1 - _idleAnimController.value) * 0.08)),
                      Colors.transparent,
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}
