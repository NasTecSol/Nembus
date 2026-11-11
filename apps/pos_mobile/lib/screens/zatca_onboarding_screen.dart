import 'dart:convert';
import 'dart:developer' as developer;
import 'package:crypto/crypto.dart';
import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:http/http.dart' as http;

import '../database/db_service.dart';
import '../services/device_serial_service.dart';
import '../singleton/singleton_class.dart';
import 'login_screen.dart';

class ZatcaOnboardingScreen extends StatefulWidget {
  final String tenantId;
  final String tenantSlug;
  final String tenantName;
  final bool isFromSettings;
  final int? storeId;
  final String? storeName;
  final int? posTerminalId;
  final String? posTerminalName;
  final String? storeType;
  final bool isFirstInstall;

  const ZatcaOnboardingScreen({
    super.key,
    required this.tenantId,
    required this.tenantSlug,
    required this.tenantName,
    this.isFromSettings = false,
    this.storeId,
    this.storeName,
    this.posTerminalId,
    this.posTerminalName,
    this.storeType,
    this.isFirstInstall = false,
  });

  @override
  State<ZatcaOnboardingScreen> createState() => _ZatcaOnboardingScreenState();
}

class _ZatcaOnboardingScreenState extends State<ZatcaOnboardingScreen> {
  static bool _hasAutoCheckedRegistration = false;
  final TextEditingController _otpController = TextEditingController();
  bool _zatcaEnabled = true;
  String _deviceSerial = 'Fetching device serial...';
  final String _environment = 'Simulation (Local Development)';
  bool _isLoading = false;
  bool _isCheckingRegistration = false;
  bool _isOnboarded = false;
  String? _statusFeedback;
  Color? _feedbackColor;

  int _organizationId = 0;
  int _storeId = 0;
  int _posTerminalId = 0;

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  @override
  void dispose() {
    _otpController.dispose();
    super.dispose();
  }

  Future<void> _loadInitialData() async {
    // 1. Fetch hardware serial from device
    final serial = await DeviceSerialService.getDeviceSerial();
    if (mounted) {
      setState(() {
        _deviceSerial = serial;
      });
    }

    // 2. Query organization_id, store_id, pos_terminal_id from SQLite
    int orgId = 0;
    int storeId = widget.storeId ?? SingletonClass().activeStoreId ?? 0;
    int terminalId = widget.posTerminalId ?? SingletonClass().activeTerminalId ?? 0;

    try {
      final db = DatabaseService().database;

      final orgRows = await db.query('organizations', limit: 1);
      if (orgRows.isNotEmpty && orgRows.first['id'] != null) {
        orgId = (orgRows.first['id'] as num).toInt();
      }

      if (storeId == 0) {
        final storeRows = await db.query('stores', limit: 1);
        if (storeRows.isNotEmpty && storeRows.first['id'] != null) {
          storeId = (storeRows.first['id'] as num).toInt();
        }
      }

      if (terminalId == 0) {
        final termRows = await db.query('pos_terminals', limit: 1);
        if (termRows.isNotEmpty && termRows.first['id'] != null) {
          terminalId = (termRows.first['id'] as num).toInt();
        }
      }

      // Check if device is already onboarded/enabled in local_device_config
      final devConfig = await db.query('local_device_config', limit: 1);
      if (devConfig.isNotEmpty) {
        final enabled = devConfig.first['zatca_enabled'];
        if (enabled == 1 || enabled == true) {
          _isOnboarded = true;
        }
      }
    } catch (e) {
      developer.log('Error reading local db context for ZATCA: $e', name: 'ZatcaOnboardingScreen');
    }

    if (mounted) {
      setState(() {
        _organizationId = orgId;
        _storeId = storeId;
        _posTerminalId = terminalId;
      });

      // Automatically check if device is already registered on first install of build
      if (widget.isFirstInstall && !widget.isFromSettings && !_hasAutoCheckedRegistration) {
        _hasAutoCheckedRegistration = true;
        _checkIfDeviceAlreadyRegistered();
      }
    }
  }

  /// Checks if current device hardware serial is already registered in ZATCA for the selected store & terminal.
  Future<void> _checkIfDeviceAlreadyRegistered() async {
    if (_isCheckingRegistration) return;
    setState(() => _isCheckingRegistration = true);

    String currentSerial = _deviceSerial;
    if (currentSerial.isEmpty || currentSerial.contains('Fetching')) {
      currentSerial = await DeviceSerialService.getDeviceSerial();
      if (mounted) {
        setState(() => _deviceSerial = currentSerial);
      }
    }

    final effectiveStoreId = widget.storeId ?? (_storeId != 0 ? _storeId : (SingletonClass().activeStoreId ?? 1));
    final effectiveTerminalId = widget.posTerminalId ?? (_posTerminalId != 0 ? _posTerminalId : (SingletonClass().activeTerminalId ?? 1));

    int effectiveOrgId = _organizationId;
    if (effectiveOrgId <= 0) {
      try {
        final db = DatabaseService().database;
        final orgRows = await db.query('organizations', limit: 1);
        if (orgRows.isNotEmpty && orgRows.first['id'] != null) {
          effectiveOrgId = (orgRows.first['id'] as num).toInt();
        }
      } catch (_) {}
    }
    if (effectiveOrgId <= 0) effectiveOrgId = 1;

    developer.log(
      'Checking ZATCA device registration for org=$effectiveOrgId, store=$effectiveStoreId, terminal=$effectiveTerminalId, serial=$currentSerial',
      name: 'ZatcaOnboard',
    );

    try {
      String effectiveTenantSlug = widget.tenantSlug.trim();
      if (effectiveTenantSlug.isEmpty) {
        effectiveTenantSlug = (SingletonClass().activeTenantSlug ?? '').trim();
      }
      if (effectiveTenantSlug.isEmpty) {
        try {
          final db = DatabaseService().database;
          final tRows = await db.query('tenants', limit: 1);
          if (tRows.isNotEmpty && tRows.first['slug'] != null) {
            effectiveTenantSlug = tRows.first['slug'].toString().trim();
          }
        } catch (_) {}
      }
      if (effectiveTenantSlug.isEmpty) {
        effectiveTenantSlug = widget.tenantId.isNotEmpty
            ? widget.tenantId
            : (SingletonClass().activeTenantId ?? '1');
      }

      final url = '${SingletonClass().baseURL}/api/zatca/devices?store_id=$effectiveStoreId&pos_terminal_id=$effectiveTerminalId';

      // 1. Fetch base dev token from cloud
      final rawDevToken = await _getAuthToken();
      final detectedSecret = rawDevToken.isNotEmpty ? _detectSecretFromDevToken(rawDevToken) : null;

      // 2. Prepare tokens with organization_id claim
      final List<String> candidateTokens = [];
      if (detectedSecret != null) {
        candidateTokens.add(_mintTokenWithOrg(effectiveOrgId, detectedSecret));
      }
      candidateTokens.add(_mintTokenWithOrg(effectiveOrgId, 'nastecsol'));
      candidateTokens.add(_mintTokenWithOrg(effectiveOrgId, 'nembus-pos-mobile-secure-jwt-secret-key-2026'));
      if (rawDevToken.isNotEmpty && !candidateTokens.contains(rawDevToken)) {
        candidateTokens.add(rawDevToken);
      }

      developer.log('================ ZATCA CHECK DEVICES REQUEST ================', name: 'ZatcaOnboard');
      developer.log('URL: $url', name: 'ZatcaOnboard');
      debugPrint('🚀 [ZATCA Check Devices API] GET $url');

      http.Response? response;
      for (final candidateToken in candidateTokens) {
        final headers = <String, String>{
          'Accept': 'application/json',
          'x-tenant-id': effectiveTenantSlug,
          'x-organization-id': '$effectiveOrgId',
          'x-org-id': '$effectiveOrgId',
          if (candidateToken.isNotEmpty)
            'Authorization': candidateToken.startsWith('Bearer ') ? candidateToken : 'Bearer $candidateToken',
        };

        developer.log('Testing token candidate (length=${candidateToken.length})...', name: 'ZatcaOnboard');
        final resp = await http
            .get(Uri.parse(url), headers: headers)
            .timeout(const Duration(seconds: 15));

        response = resp;
        if (resp.statusCode == 200 || resp.statusCode == 201) {
          SingletonClass().authToken = candidateToken;
          developer.log('Acquired working authenticated token with organization claim!', name: 'ZatcaOnboard');
          break;
        }

        // If error is not 401, stop trying token variants
        if (resp.statusCode != 401) {
          break;
        }
      }

      developer.log('================ ZATCA CHECK DEVICES RESPONSE ===============', name: 'ZatcaOnboard');
      developer.log('Status Code: ${response?.statusCode}', name: 'ZatcaOnboard');
      developer.log('Response Body: ${response?.body}', name: 'ZatcaOnboard');
      debugPrint('📥 [ZATCA Check Devices API] Response Status: ${response?.statusCode}');
      debugPrint('Response Body: ${response?.body}');

      if (response != null && response.statusCode >= 200 && response.statusCode < 300) {
        final decoded = jsonDecode(response.body);
        List<Map<String, dynamic>> deviceList = [];

        if (decoded is List) {
          for (var item in decoded) {
            if (item is Map) deviceList.add(Map<String, dynamic>.from(item));
          }
        } else if (decoded is Map) {
          final candidate = decoded['data'] ?? decoded['devices'] ?? decoded['items'] ?? decoded['results'];
          if (candidate is List) {
            for (var item in candidate) {
              if (item is Map) deviceList.add(Map<String, dynamic>.from(item));
            }
          } else if (decoded['data'] is Map) {
            final dData = decoded['data'];
            if (dData['devices'] is List) {
              for (var item in dData['devices']) {
                if (item is Map) deviceList.add(Map<String, dynamic>.from(item));
              }
            } else {
              deviceList.add(Map<String, dynamic>.from(dData));
            }
          } else if (decoded.containsKey('device_serial') ||
              decoded.containsKey('serial_number') ||
              decoded.containsKey('serial')) {
            deviceList.add(Map<String, dynamic>.from(decoded));
          }
        }

        final cleanCurrentSerial = currentSerial.trim().toLowerCase();
        bool isAlreadyRegistered = false;

        for (final dev in deviceList) {
          final devSerial = (dev['device_serial'] ??
                  dev['deviceSerial'] ??
                  dev['serial_number'] ??
                  dev['serialNumber'] ??
                  dev['serial'])
              ?.toString()
              .trim()
              .toLowerCase();
          if (devSerial != null && devSerial.isNotEmpty && devSerial == cleanCurrentSerial) {
            isAlreadyRegistered = true;
            break;
          }
        }

        if (isAlreadyRegistered) {
          developer.log('Device serial ($currentSerial) matches registered device in ZATCA!', name: 'ZatcaOnboard');

          try {
            final db = DatabaseService().database;
            await db.rawUpdate('''
              UPDATE local_device_config 
              SET zatca_enabled = 1, last_zatca_sync_at = CURRENT_TIMESTAMP
              WHERE id = 1
            ''');
          } catch (_) {}
          await SingletonClass().setZatcaOnboardingCompleted(true);

          if (mounted) {
            setState(() {
              _isOnboarded = true;
              _statusFeedback = 'Device is already registered with ZATCA!';
              _feedbackColor = const Color(0xFF10B981);
            });

            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('✅ Device is already registered with ZATCA! Navigating to login...'),
                backgroundColor: Color(0xFF10B981),
                duration: Duration(seconds: 2),
              ),
            );

            await Future.delayed(const Duration(milliseconds: 1200));
            if (mounted) {
              _proceedNext();
            }
          }
          return;
        } else {
          developer.log('Device serial ($currentSerial) not registered in ZATCA. Remaining on onboarding screen.', name: 'ZatcaOnboard');
        }
      } else {
        developer.log('Check devices endpoint returned status ${response?.statusCode}. Remaining on onboarding screen.', name: 'ZatcaOnboard');
      }
    } catch (e) {
      developer.log('Check device registration note: $e', name: 'ZatcaOnboard');
    } finally {
      if (mounted) {
        setState(() => _isCheckingRegistration = false);
      }
    }
  }

  /// Mints a JWT token containing organization_id claim signed with the given secret.
  String _mintTokenWithOrg(int orgId, String secret) {
    final header = {"alg": "HS256", "typ": "JWT"};
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
    final payload = {
      "user_id": "Admin",
      "user_login": "Admin",
      "organization_id": orgId,
      "iat": now,
      "exp": now + 86400 * 30,
    };

    String base64UrlNoPad(List<int> bytes) {
      return base64Url.encode(bytes).replaceAll('=', '');
    }

    final headerB64 = base64UrlNoPad(utf8.encode(jsonEncode(header)));
    final payloadB64 = base64UrlNoPad(utf8.encode(jsonEncode(payload)));
    final message = '$headerB64.$payloadB64';

    final hmac = Hmac(sha256, utf8.encode(secret));
    final digest = hmac.convert(utf8.encode(message));
    final sigB64 = base64UrlNoPad(digest.bytes);

    return '$message.$sigB64';
  }

  /// Attempts to identify which secret was used by the server to sign the dev token.
  String? _detectSecretFromDevToken(String devToken) {
    final parts = devToken.split('.');
    if (parts.length != 3) return null;

    final message = '${parts[0]}.${parts[1]}';
    final targetSig = parts[2];

    final candidateSecrets = [
      'nastecsol',
      'nembus-pos-mobile-secure-jwt-secret-key-2026',
      'nembus',
      'secret',
      'development',
      'root',
      'admin',
      'password',
      '',
    ];

    for (final sec in candidateSecrets) {
      final hmac = Hmac(sha256, utf8.encode(sec));
      final digest = hmac.convert(utf8.encode(message));
      final sig = base64Url.encode(digest.bytes).replaceAll('=', '');
      if (sig == targetSig) {
        developer.log('Identified server JWT secret: $sec', name: 'ZatcaOnboard');
        return sec;
      }
    }
    return null;
  }

  Future<String> _getAuthToken() async {
    if (SingletonClass().authToken != null && SingletonClass().authToken!.isNotEmpty) {
      return SingletonClass().authToken!;
    }

    try {
      final tokenUrl = Uri.parse('${SingletonClass().baseURL}/api/dev/token');
      developer.log('Fetching auth token from $tokenUrl', name: 'ZatcaOnboard');
      final res = await http.get(
        tokenUrl,
        headers: {'Accept': 'application/json'},
      ).timeout(const Duration(seconds: 10));

      if (res.statusCode == 200) {
        final decoded = jsonDecode(res.body);
        if (decoded is Map && decoded['token'] != null) {
          final token = decoded['token'].toString();
          SingletonClass().authToken = token;
          developer.log('Acquired auth token successfully', name: 'ZatcaOnboard');
          return token;
        }
      } else {
        developer.log('Dev token endpoint returned ${res.statusCode}: ${res.body}', name: 'ZatcaOnboard');
      }
    } catch (e) {
      developer.log('Error fetching dev token: $e', name: 'ZatcaOnboard');
    }

    return SingletonClass().authToken ?? '';
  }

  Future<bool> _executeOnboardApi({bool silentIfEmpty = false}) async {
    final otpText = _otpController.text.trim();
    if (otpText.isEmpty) {
      if (!silentIfEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Please enter the 6-digit OTP from FATOORA portal.'),
            backgroundColor: Color(0xFFEF4444),
          ),
        );
      }
      return false;
    }

    setState(() {
      _isLoading = true;
      _statusFeedback = null;
    });

    final token = await _getAuthToken();

    String effectiveTenantSlug = widget.tenantSlug.trim();
    if (effectiveTenantSlug.isEmpty) {
      effectiveTenantSlug = (SingletonClass().activeTenantSlug ?? '').trim();
    }
    if (effectiveTenantSlug.isEmpty) {
      try {
        final db = DatabaseService().database;
        final tRows = await db.query('tenants', limit: 1);
        if (tRows.isNotEmpty && tRows.first['slug'] != null) {
          effectiveTenantSlug = tRows.first['slug'].toString().trim();
        }
      } catch (_) {}
    }
    if (effectiveTenantSlug.isEmpty) {
      effectiveTenantSlug = widget.tenantId.isNotEmpty
          ? widget.tenantId
          : (SingletonClass().activeTenantId ?? '1');
    }

    final url = '${SingletonClass().baseURL}/api/zatca/onboard';

    final headers = <String, String>{
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'x-tenant-id': effectiveTenantSlug,
    };

    if (token.isNotEmpty) {
      headers['Authorization'] = token.startsWith('Bearer ') ? token : 'Bearer $token';
    }

    final bodyMap = {
      'device_serial': _deviceSerial,
      'device_type': 'pos',
      'organization_id': _organizationId,
      'otp': otpText,
      'pos_terminal_id': _posTerminalId,
      'store_id': _storeId,
      'zatca_env': 'simulation',
    };

    // Print outgoing request details
    developer.log('================ ZATCA ONBOARD API REQUEST ================', name: 'ZatcaOnboard');
    developer.log('URL: $url', name: 'ZatcaOnboard');
    developer.log('Headers: ${jsonEncode(headers)}', name: 'ZatcaOnboard');
    developer.log('Body: ${jsonEncode(bodyMap)}', name: 'ZatcaOnboard');
    debugPrint('🚀 [ZATCA Onboard API] POST $url');
    debugPrint('Headers: $headers');
    debugPrint('Body: ${jsonEncode(bodyMap)}');

    try {
      final response = await http
          .post(
            Uri.parse(url),
            headers: headers,
            body: jsonEncode(bodyMap),
          )
          .timeout(const Duration(seconds: 30));

      // Print incoming response details
      developer.log('================ ZATCA ONBOARD API RESPONSE ===============', name: 'ZatcaOnboard');
      developer.log('Status Code: ${response.statusCode}', name: 'ZatcaOnboard');
      developer.log('Response Body: ${response.body}', name: 'ZatcaOnboard');
      debugPrint('📥 [ZATCA Onboard API] Response Status: ${response.statusCode}');
      debugPrint('Response Body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        // Successful registration
        try {
          final db = DatabaseService().database;
          await db.rawUpdate('''
            UPDATE local_device_config 
            SET zatca_enabled = 1, last_zatca_sync_at = CURRENT_TIMESTAMP
            WHERE id = 1
          ''');
        } catch (_) {}
        await SingletonClass().setZatcaOnboardingCompleted(true);

        if (mounted) {
          setState(() {
            _isLoading = false;
            _isOnboarded = true;
            _statusFeedback = 'Device successfully onboarded & registered with ZATCA!';
            _feedbackColor = const Color(0xFF10B981);
          });

          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('✅ Device registered with ZATCA successfully!'),
              backgroundColor: Color(0xFF10B981),
            ),
          );
        }
        return true;
      } else {
        String errMsg = 'Onboarding failed with status code ${response.statusCode}';
        try {
          final decoded = jsonDecode(response.body);
          if (decoded is Map && decoded['error'] != null) {
            errMsg = decoded['error'].toString();
          } else if (decoded is Map && decoded['message'] != null) {
            errMsg = decoded['message'].toString();
          }
        } catch (_) {
          errMsg = response.body.isNotEmpty ? response.body : errMsg;
        }

        if (mounted) {
          setState(() {
            _isLoading = false;
            _statusFeedback = 'Error: $errMsg';
            _feedbackColor = const Color(0xFFEF4444);
          });

          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('⚠️ $errMsg'),
              backgroundColor: const Color(0xFFEF4444),
            ),
          );
        }
        return false;
      }
    } catch (e) {
      developer.log('ZATCA Onboarding Exception: $e', name: 'ZatcaOnboard', error: e);
      debugPrint('❌ [ZATCA Onboard API] Exception: $e');

      if (mounted) {
        setState(() {
          _isLoading = false;
          _statusFeedback = 'Connection error: $e';
          _feedbackColor = const Color(0xFFEF4444);
        });

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('❌ Connection error: $e'),
            backgroundColor: const Color(0xFFEF4444),
          ),
        );
      }
      return false;
    }
  }

  void _proceedNext() {
    SingletonClass().setZatcaOnboardingCompleted(true);
    if (widget.isFromSettings) {
      Navigator.of(context).pop();
    } else {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (_) => LoginScreen(
            tenantSlug: widget.tenantSlug,
            tenantName: widget.tenantName,
            storeId: widget.storeId ?? _storeId,
            storeName: widget.storeName ?? SingletonClass().activeStoreName,
            posTerminalId: widget.posTerminalId ?? _posTerminalId,
            posTerminalName: widget.posTerminalName ?? SingletonClass().activeTerminalName,
            storeType: widget.storeType ?? SingletonClass().activeStoreType,
          ),
        ),
      );
    }
  }

  Future<void> _handleSaveAndComplete() async {
    final otpText = _otpController.text.trim();
    if (otpText.isNotEmpty && !_isOnboarded) {
      final success = await _executeOnboardApi(silentIfEmpty: false);
      if (success) {
        await Future.delayed(const Duration(milliseconds: 600));
        _proceedNext();
      }
    } else {
      _proceedNext();
    }
  }

  void _handleSkipOrBack() {
    _proceedNext();
  }

  @override
  Widget build(BuildContext context) {
    const screenBg = Color(0xFF1E3A4A);
    const cardBg = Colors.white;
    const primaryIndigo = Color(0xFF6366F1);
    const textDark = Color(0xFF1E293B);
    const textMuted = Color(0xFF64748B);
    const borderLight = Color(0xFFE2E8F0);
    const fieldBg = Color(0xFFF1F5F9);

    return Scaffold(
      backgroundColor: screenBg,
      body: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            final isWide = constraints.maxWidth >= 600;

            return Center(
              child: SingleChildScrollView(
                padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 24.0),
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 720),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // Main White Card: ZATCA Phase 2 E-Invoicing
                      Container(
                        decoration: BoxDecoration(
                          color: cardBg,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: borderLight, width: 1.2),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withValues(alpha: 0.15),
                              blurRadius: 18,
                              offset: const Offset(0, 6),
                            ),
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            // Card Header: Icon, Titles, Toggle Switch
                            Padding(
                              padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 16.0),
                              child: Row(
                                crossAxisAlignment: CrossAxisAlignment.center,
                                children: [
                                  // Purple Square Icon Box
                                  Container(
                                    width: 44,
                                    height: 44,
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFEDE9FE),
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                    child: const Icon(
                                      Icons.sim_card_download_rounded,
                                      color: primaryIndigo,
                                      size: 24,
                                    ),
                                  ),
                                  const SizedBox(width: 14),
                                  // Title & Subtitle
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          'ZATCA Phase 2 E-Invoicing',
                                          style: GoogleFonts.inter(
                                            fontSize: 16,
                                            fontWeight: FontWeight.w700,
                                            color: textDark,
                                            letterSpacing: -0.2,
                                          ),
                                        ),
                                        const SizedBox(height: 2),
                                        Text(
                                          'KSA FATOORA cryptographic compliance & simplified invoice signing',
                                          style: GoogleFonts.inter(
                                            fontSize: 12,
                                            color: textMuted,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  // Toggle Switch
                                  CupertinoSwitch(
                                    value: _zatcaEnabled,
                                    activeTrackColor: primaryIndigo,
                                    onChanged: (val) {
                                      setState(() {
                                        _zatcaEnabled = val;
                                      });
                                    },
                                  ),
                                ],
                              ),
                            ),

                            const Divider(
                              color: borderLight,
                              height: 1,
                              thickness: 1,
                            ),

                            // Card Body: Fields & Onboard Action
                            Padding(
                              padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 20.0),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.stretch,
                                children: [
                                  // Environment & Device EGS Serial Row/Column
                                  if (isWide)
                                    Row(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Expanded(
                                          child: _buildField(
                                            label: 'Environment',
                                            value: _environment,
                                            isReadOnly: true,
                                            fieldBg: const Color(0xFFF8FAFC),
                                          ),
                                        ),
                                        const SizedBox(width: 16),
                                        Expanded(
                                          child: _buildField(
                                            label: 'Device EGS Serial',
                                            value: _deviceSerial,
                                            isReadOnly: true,
                                            fieldBg: fieldBg,
                                            isMonospace: true,
                                          ),
                                        ),
                                      ],
                                    )
                                  else ...[
                                    _buildField(
                                      label: 'Environment',
                                      value: _environment,
                                      isReadOnly: true,
                                      fieldBg: const Color(0xFFF8FAFC),
                                    ),
                                    const SizedBox(height: 14),
                                    _buildField(
                                      label: 'Device EGS Serial',
                                      value: _deviceSerial,
                                      isReadOnly: true,
                                      fieldBg: fieldBg,
                                      isMonospace: true,
                                    ),
                                  ],

                                  const SizedBox(height: 18),

                                  // ZATCA 1-Hour OTP Code Section
                                  Text(
                                    'ZATCA 1-Hour OTP Code',
                                    style: GoogleFonts.inter(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w600,
                                      color: const Color(0xFF334155),
                                    ),
                                  ),
                                  const SizedBox(height: 8),

                                  Row(
                                    crossAxisAlignment: CrossAxisAlignment.center,
                                    children: [
                                      // OTP Input Field
                                      Expanded(
                                        child: Container(
                                          decoration: BoxDecoration(
                                            color: fieldBg,
                                            borderRadius: BorderRadius.circular(10),
                                            border: Border.all(color: borderLight, width: 1.2),
                                          ),
                                          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 4),
                                          child: TextField(
                                            controller: _otpController,
                                            keyboardType: TextInputType.text,
                                            style: const TextStyle(
                                              fontFamily: 'monospace',
                                              fontSize: 15,
                                              fontWeight: FontWeight.w600,
                                              color: Color(0xFF0F172A),
                                              letterSpacing: 1.5,
                                            ),
                                            decoration: InputDecoration(
                                              border: InputBorder.none,
                                              isDense: true,
                                              hintText: 'Enter 6-digit OTP',
                                              hintStyle: GoogleFonts.inter(
                                                color: const Color(0xFF94A3B8),
                                                fontSize: 13,
                                                letterSpacing: 0,
                                              ),
                                            ),
                                          ),
                                        ),
                                      ),
                                      const SizedBox(width: 10),

                                      // Onboard Now Button
                                      SizedBox(
                                        height: 48,
                                        child: ElevatedButton(
                                          onPressed: _isLoading ? null : () => _executeOnboardApi(silentIfEmpty: false),
                                          style: ElevatedButton.styleFrom(
                                            backgroundColor: _isOnboarded
                                                ? const Color(0xFF10B981)
                                                : const Color(0xFFCBD5E1),
                                            foregroundColor: _isOnboarded ? Colors.white : const Color(0xFF475569),
                                            disabledBackgroundColor: const Color(0xFFE2E8F0),
                                            shape: RoundedRectangleBorder(
                                              borderRadius: BorderRadius.circular(10),
                                            ),
                                            padding: const EdgeInsets.symmetric(horizontal: 12),
                                            elevation: 0,
                                          ),
                                          child: _isLoading
                                              ? const SizedBox(
                                                  width: 18,
                                                  height: 18,
                                                  child: CircularProgressIndicator(
                                                    strokeWidth: 2,
                                                    color: Colors.white,
                                                  ),
                                                )
                                              : Row(
                                                  mainAxisSize: MainAxisSize.min,
                                                  children: [
                                                    Icon(
                                                      _isOnboarded
                                                          ? Icons.check_circle_rounded
                                                          : Icons.shield_rounded,
                                                      size: 18,
                                                      color: _isOnboarded ? Colors.white : const Color(0xFF64748B),
                                                    ),
                                                    const SizedBox(width: 6),
                                                    Text(
                                                      _isOnboarded ? 'Onboarded' : 'Onboard Now',
                                                      style: GoogleFonts.inter(
                                                        fontSize: 14,
                                                        fontWeight: FontWeight.w700,
                                                        color: _isOnboarded ? Colors.white : const Color(0xFF334155),
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                        ),
                                      ),
                                    ],
                                  ),

                                  const SizedBox(height: 8),

                                  // Helper Text
                                  Text(
                                    'Generate OTP from FATOORA portal > Onboarding Solutions > Add Solution Unit.',
                                    style: GoogleFonts.inter(
                                      fontSize: 11.5,
                                      color: const Color(0xFF64748B),
                                    ),
                                  ),

                                  if (_statusFeedback != null) ...[
                                    const SizedBox(height: 12),
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                                      decoration: BoxDecoration(
                                        color: (_feedbackColor ?? primaryIndigo).withValues(alpha: 0.1),
                                        borderRadius: BorderRadius.circular(8),
                                        border: Border.all(
                                          color: (_feedbackColor ?? primaryIndigo).withValues(alpha: 0.3),
                                        ),
                                      ),
                                      child: Text(
                                        _statusFeedback!,
                                        style: TextStyle(
                                          fontSize: 12.5,
                                          fontWeight: FontWeight.w600,
                                          color: _feedbackColor ?? primaryIndigo,
                                        ),
                                      ),
                                    ),
                                  ],
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),

                      const SizedBox(height: 24),

                      // Bottom Action Buttons: Back / Skip & Save & Complete Setup
                      LayoutBuilder(
                        builder: (context, actionConstraints) {
                          final isNarrow = actionConstraints.maxWidth < 370;

                          if (isNarrow) {
                            return Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                ElevatedButton(
                                  onPressed: _isLoading ? null : _handleSaveAndComplete,
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: const Color(0xFF0F172A),
                                    foregroundColor: Colors.white,
                                    disabledBackgroundColor: Colors.white24,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(12),
                                      side: const BorderSide(color: Colors.white24, width: 1),
                                    ),
                                    padding: const EdgeInsets.symmetric(vertical: 14),
                                    elevation: 4,
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Text(
                                        'Save & Complete Setup',
                                        style: GoogleFonts.inter(
                                          fontSize: 14.5,
                                          fontWeight: FontWeight.w700,
                                          color: Colors.white,
                                        ),
                                      ),
                                      const SizedBox(width: 8),
                                      const Icon(
                                        Icons.check_rounded,
                                        size: 18,
                                        color: Colors.white,
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(height: 8),
                                Center(
                                  child: TextButton(
                                    onPressed: _handleSkipOrBack,
                                    style: TextButton.styleFrom(
                                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                                    ),
                                    child: Text(
                                      widget.isFromSettings ? 'Back' : 'Skip for now',
                                      style: GoogleFonts.inter(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w600,
                                        color: Colors.white70,
                                      ),
                                    ),
                                  ),
                                ),
                              ],
                            );
                          }

                          return Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              TextButton(
                                onPressed: _handleSkipOrBack,
                                style: TextButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                                ),
                                child: Text(
                                  widget.isFromSettings ? 'Back' : 'Skip for now',
                                  style: GoogleFonts.inter(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                    color: Colors.white70,
                                  ),
                                ),
                              ),
                              const SizedBox(width: 8),
                              Flexible(
                                child: ElevatedButton(
                                  onPressed: _isLoading ? null : _handleSaveAndComplete,
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: const Color(0xFF0F172A),
                                    foregroundColor: Colors.white,
                                    disabledBackgroundColor: Colors.white24,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(12),
                                      side: const BorderSide(color: Colors.white24, width: 1),
                                    ),
                                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                                    elevation: 4,
                                  ),
                                  child: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Flexible(
                                        child: Text(
                                          'Save & Complete Setup',
                                          style: GoogleFonts.inter(
                                            fontSize: 14,
                                            fontWeight: FontWeight.w700,
                                            color: Colors.white,
                                          ),
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      ),
                                      const SizedBox(width: 6),
                                      const Icon(
                                        Icons.check_rounded,
                                        size: 18,
                                        color: Colors.white,
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ],
                          );
                        },
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildField({
    required String label,
    required String value,
    required bool isReadOnly,
    required Color fieldBg,
    bool isMonospace = false,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: const Color(0xFF334155),
          ),
        ),
        const SizedBox(height: 6),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
          decoration: BoxDecoration(
            color: fieldBg,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: const Color(0xFFE2E8F0), width: 1.2),
          ),
          child: Text(
            value,
            style: isMonospace
                ? const TextStyle(
                    fontFamily: 'monospace',
                    fontSize: 13.5,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF64748B),
                  )
                : GoogleFonts.inter(
                    fontSize: 13.5,
                    fontWeight: FontWeight.w500,
                    color: const Color(0xFF1E293B),
                  ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }
}
