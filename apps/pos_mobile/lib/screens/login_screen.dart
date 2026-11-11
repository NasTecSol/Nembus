import 'dart:convert';
import 'dart:developer' as developer;
import 'package:crypto/crypto.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:http/http.dart' as http;

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
  final String? storeType;

  const LoginScreen({
    super.key,
    this.tenantSlug,
    this.tenantName,
    this.storeId,
    this.storeName,
    this.posTerminalId,
    this.posTerminalName,
    this.storeType,
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
          final email = (c['email'] ?? '').toString().toLowerCase();
          final name = '${c['first_name'] ?? ''} ${c['last_name'] ?? ''}'.toLowerCase();
          return code.contains(query) || username.contains(query) || email.contains(query) || name.contains(query);
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
      if (!DatabaseService().isOpen) {
        await DatabaseService().initDatabase();
      }

      // Ensure SQLite users table has required columns for Nembus Core queries
      try {
        final db = DatabaseService().database;
        await db.execute('ALTER TABLE users ADD COLUMN must_reset_password BOOLEAN DEFAULT 0;');
      } catch (_) {}
      try {
        final db = DatabaseService().database;
        await db.execute('ALTER TABLE users ADD COLUMN sap_imported BOOLEAN DEFAULT 0;');
      } catch (_) {}

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

      // Enrich cashier records with real username and user details from users table
      try {
        final db = DatabaseService().database;
        for (int i = 0; i < list.length; i++) {
          final uId = list[i]['user_id'];
          if (uId != null) {
            final userRows = await db.rawQuery(
              'SELECT username, first_name, last_name, email, employee_code FROM users WHERE id = ?',
              [uId],
            );
            if (userRows.isNotEmpty) {
              final u = userRows.first;
              list[i]['username'] = u['username'];
              list[i]['first_name'] = u['first_name'];
              list[i]['last_name'] = u['last_name'];
              list[i]['email'] = u['email'];
              list[i]['employee_code'] = u['employee_code'];
            }
          }
        }
      } catch (e) {
        developer.log('SQLite cashier user enrichment note: $e', name: 'LoginScreen');
      }

      // 2. Fallback: Query local SQLite for enriched user info (joining cashiers and users)
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

      // 3. Fallback: If still empty, load active users as available cashiers
      if (list.isEmpty) {
        try {
          final db = DatabaseService().database;
          final userRows = await db.rawQuery('''
            SELECT 
              id AS cashier_id,
              id AS user_id,
              ? AS store_id,
              'CSH-' || id AS cashier_code,
              username,
              first_name,
              last_name,
              email,
              employee_code,
              1 AS is_active
            FROM users
            WHERE is_active = 1
            ORDER BY id ASC
          ''', [_effectiveStoreId]);
          if (userRows.isNotEmpty) {
            list = userRows.map((r) => Map<String, dynamic>.from(r)).toList();
          }
        } catch (e) {
          developer.log('SQLite user fallback note: $e', name: 'LoginScreen');
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

  bool _isBridgeSuccess(Map<String, dynamic> resp) {
    if (resp['success'] == true) return true;
    if (resp['statusCode'] == 200) return true;
    final msg = resp['message']?.toString().toLowerCase() ?? '';
    if (msg == 'login successful' || msg == 'failed to generate token') return true;
    return false;
  }

  /// Opens the Password Authentication modal for the selected cashier
  Future<void> _onCashierSelected(Map<String, dynamic> cashier) async {
    if (!DatabaseService().isOpen) {
      await DatabaseService().initDatabase();
    }
    // Enrich cashier email & username from SQLite users table if missing
    if (cashier['email'] == null ||
        cashier['email'].toString().trim().isEmpty ||
        cashier['username'] == null ||
        cashier['username'].toString().trim().isEmpty) {
      try {
        final db = DatabaseService().database;
        final targetUserId = cashier['user_id'];
        if (targetUserId != null) {
          final rows = await db.query(
            'users',
            columns: ['username', 'first_name', 'last_name', 'email', 'employee_code'],
            where: 'id = ?',
            whereArgs: [targetUserId],
            limit: 1,
          );
          if (rows.isNotEmpty) {
            final u = rows.first;
            if (u['email'] != null && u['email'].toString().trim().isNotEmpty) {
              cashier['email'] = u['email'];
            }
            if (cashier['username'] == null || cashier['username'].toString().trim().isEmpty) {
              cashier['username'] = u['username'];
            }
            if (cashier['first_name'] == null) cashier['first_name'] = u['first_name'];
            if (cashier['last_name'] == null) cashier['last_name'] = u['last_name'];
          }
        }
      } catch (e) {
        developer.log('Cashier user resolution note: $e', name: 'LoginScreen');
      }
    }
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
    final initialUsername = (cashier['username'] != null && cashier['username'].toString().trim().isNotEmpty)
        ? cashier['username'].toString().trim()
        : cashierCode;

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
            final inputUsername = usernameCtrl.text.trim();
            final password = passwordCtrl.text;

            if (inputUsername.isEmpty) {
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
              if (!DatabaseService().isOpen) {
                await DatabaseService().initDatabase();
              }
              final db = DatabaseService().database;

              // Ensure columns exist on SQLite users table so Go Core queries never throw missing-column errors
              try {
                await db.execute('ALTER TABLE users ADD COLUMN must_reset_password BOOLEAN DEFAULT 0;');
              } catch (_) {}
              try {
                await db.execute('ALTER TABLE users ADD COLUMN sap_imported BOOLEAN DEFAULT 0;');
              } catch (_) {}

              final cashierId = (cashier['cashier_id'] ?? cashier['id'] ?? 1) as int;
              final targetUserId = cashier['user_id'];

              // 1. Direct offline lookup in SQLite local db to find user and stored password hash
              Map<String, dynamic>? userRow;
              if (inputUsername.isNotEmpty) {
                final rows = await db.query(
                  'users',
                  where: 'username = ?',
                  whereArgs: [inputUsername],
                  limit: 1,
                );
                if (rows.isNotEmpty) userRow = rows.first;
              }
              if (userRow == null && targetUserId != null) {
                final rows = await db.query(
                  'users',
                  where: 'id = ?',
                  whereArgs: [targetUserId],
                  limit: 1,
                );
                if (rows.isNotEmpty) userRow = rows.first;
              }
              if (userRow == null && cashier['email'] != null && cashier['email'].toString().isNotEmpty) {
                final rows = await db.query(
                  'users',
                  where: 'email = ?',
                  whereArgs: [cashier['email'].toString().trim()],
                  limit: 1,
                );
                if (rows.isNotEmpty) userRow = rows.first;
              }

              final resolvedUsername = (userRow?['username'] ?? inputUsername).toString().trim();
              final effectiveUserId = (userRow?['id'] ?? targetUserId ?? 1) as int;
              final userEmail = userRow?['email']?.toString() ?? cashier['email']?.toString();
              final storedHash = (userRow?['password_hash'] ?? '').toString();

              bool isAuthSuccess = false;
              String? lastErrorMsg;

              // 2. Call Go Core auth handler via NembusBridge (matching against local SQLite)
              try {
                var response = NembusBridge().callHandler(
                  handler: 'auth',
                  action: 'login',
                  payload: {
                    'user_login': resolvedUsername,
                    'password': password,
                  },
                );
                developer.log('Auth handler response for $resolvedUsername: $response', name: 'LoginScreen');

                if (_isBridgeSuccess(response)) {
                  isAuthSuccess = true;
                } else if (userEmail != null && userEmail.isNotEmpty) {
                  response = NembusBridge().callHandler(
                    handler: 'auth',
                    action: 'login',
                    payload: {
                      'user_login': userEmail,
                      'password': password,
                    },
                  );
                  developer.log('Auth handler response for $userEmail: $response', name: 'LoginScreen');
                  if (_isBridgeSuccess(response)) {
                    isAuthSuccess = true;
                  } else {
                    lastErrorMsg = response['message']?.toString() ?? response['error']?.toString();
                  }
                } else {
                  lastErrorMsg = response['message']?.toString() ?? response['error']?.toString();
                }
              } catch (bridgeErr) {
                developer.log('Native bridge auth check note: $bridgeErr', name: 'LoginScreen');
              }

              // 3. Fallback: Direct offline SQLite validation (plaintext, MD5, SHA-256)
              if (!isAuthSuccess && storedHash.isNotEmpty) {
                if (storedHash == password) {
                  isAuthSuccess = true;
                } else {
                  final md5Hash = md5.convert(utf8.encode(password)).toString();
                  final sha256Hash = sha256.convert(utf8.encode(password)).toString();
                  if (storedHash.toLowerCase() == md5Hash.toLowerCase() ||
                      storedHash.toLowerCase() == sha256Hash.toLowerCase()) {
                    isAuthSuccess = true;
                  }
                }
              }

              // 4. Online Cloud Authentication Fallback:
              // When the password hash has not yet been seeded into local SQLite (e.g. cloud REST APIs
              // strip password_hash during DB clone for security), authenticate directly against the
              // cloud server /api/auth/login and cache the verified hash in SQLite for future offline use.
              if (!isAuthSuccess) {
                try {
                  developer.log('Attempting online cloud authentication for $resolvedUsername...', name: 'LoginScreen');
                  final cloudBaseUrl = SingletonClass().baseURL.isNotEmpty
                      ? SingletonClass().baseURL
                      : 'https://nembus.nashrms.com';
                  final loginUri = Uri.parse('$cloudBaseUrl/api/auth/login');

                  final cloudHeaders = <String, String>{
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                    if (_effectiveTenantSlug.isNotEmpty && _effectiveTenantSlug != 'masterdb') ...{
                      'x-tenant-id': _effectiveTenantSlug,
                      'x-tenant-slug': _effectiveTenantSlug,
                    },
                  };

                  final credentialsToTry = [
                    resolvedUsername,
                    if (userEmail != null && userEmail.isNotEmpty && userEmail != resolvedUsername) userEmail,
                  ];

                  for (final cred in credentialsToTry) {
                    final cloudResp = await http.post(
                      loginUri,
                      headers: cloudHeaders,
                      body: jsonEncode({
                        'user_login': cred,
                        'password': password,
                      }),
                    ).timeout(const Duration(seconds: 10));

                    developer.log('Cloud auth ($cred): status=${cloudResp.statusCode}, body=${cloudResp.body}', name: 'LoginScreen');

                    if (cloudResp.statusCode >= 200 && cloudResp.statusCode < 300) {
                      isAuthSuccess = true;
                      try {
                        final resJson = jsonDecode(cloudResp.body);
                        if (resJson is Map && resJson['data'] != null) {
                          SingletonClass().authToken = resJson['data'].toString();
                        }
                      } catch (_) {}

                      // CRITICAL: Cache the verified password hash in local SQLite users table!
                      // This ensures that all future cashier logins can run 100% offline.
                      try {
                        final sha256Hash = sha256.convert(utf8.encode(password)).toString();
                        await db.update(
                          'users',
                          {
                            'password_hash': sha256Hash,
                            'updated_at': DateTime.now().toIso8601String(),
                          },
                          where: 'id = ?',
                          whereArgs: [effectiveUserId],
                        );
                        developer.log('Cached verified password hash into SQLite for user $effectiveUserId ($resolvedUsername)', name: 'LoginScreen');
                      } catch (cacheErr) {
                        developer.log('Failed caching password hash to SQLite: $cacheErr', name: 'LoginScreen');
                      }
                      break;
                    } else {
                      try {
                        final errJson = jsonDecode(cloudResp.body);
                        if (errJson is Map && errJson['message'] != null) {
                          lastErrorMsg = errJson['message'].toString();
                        }
                      } catch (_) {}
                    }
                  }
                } catch (cloudErr) {
                  developer.log('Cloud auth fallback network note: $cloudErr', name: 'LoginScreen');
                }
              }

              if (isAuthSuccess) {
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
                        userId: effectiveUserId,
                        username: resolvedUsername,
                        roleName: 'Cashier',
                        storeId: _effectiveStoreId,
                        storeName: _effectiveStoreName,
                        posTerminalId: _effectiveTerminalId,
                        posTerminalName: _effectiveTerminalName,
                        storeType: widget.storeType ?? SingletonClass().activeStoreType,
                      ),
                    ),
                    (route) => false,
                  );
                }
              } else {
                modalSetState(() {
                  isAuthenticating = false;
                  authErrorMessage = lastErrorMsg ?? 'Invalid credentials. Please try again.';
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

                      // Username Field (Autofilled with cashier's username)
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
                            prefixIcon: Icon(Icons.person_outline_rounded, color: primarySky, size: 20),
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
    final email = cashier['email']?.toString() ?? '';
    final username = cashier['username']?.toString() ?? 'operator';
    final displaySubtitle = username.isNotEmpty ? '@$username' : (email.isNotEmpty ? email : code);
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
              '$code • $displaySubtitle',
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
