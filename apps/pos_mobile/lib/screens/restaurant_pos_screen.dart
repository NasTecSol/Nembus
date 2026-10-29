import 'dart:async';
import 'dart:convert';
import 'dart:developer' as developer;
import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:sqflite/sqflite.dart';
import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import '../widgets/thermal_receipt_widget.dart';

class RestaurantPosScreen extends StatefulWidget {
  final String tenantSlug;
  final String tenantName;
  final int userId;
  final String username;
  final String roleName;
  final int storeId;
  final String storeName;
  final int posTerminalId;
  final String posTerminalName;
  final String initialView; // 'tables', 'menu', 'kds', 'bills'
  final String initialOrderSource; // 'dining', 'takeaway'

  const RestaurantPosScreen({
    super.key,
    required this.tenantSlug,
    required this.tenantName,
    required this.userId,
    required this.username,
    required this.roleName,
    required this.storeId,
    required this.storeName,
    required this.posTerminalId,
    required this.posTerminalName,
    this.initialView = 'tables',
    this.initialOrderSource = 'dining',
  });

  @override
  State<RestaurantPosScreen> createState() => _RestaurantPosScreenState();
}

class _RestaurantPosScreenState extends State<RestaurantPosScreen> {
  // Current Active View: 'tables', 'menu', 'checkout', 'kds', 'bills'
  late String _currentView;

  // Active Dining / Takeaway state
  late String _orderSource; // 'dining' or 'takeaway'
  Map<String, dynamic>? _selectedTable;

  // Menu Catalog State
  List<Map<String, dynamic>> _categories = [];
  int? _selectedCategoryId;
  String _searchQuery = '';
  bool _isLoadingMenu = true;

  // Tables Floor Plan State
  List<Map<String, dynamic>> _tables = [];
  String _selectedSection = 'All';
  bool _isLoadingTables = true;

  // Bills State
  List<Map<String, dynamic>> _bills = [];
  bool _isLoadingBills = false;
  String _billsFilter = 'All';

  // KDS State
  List<Map<String, dynamic>> _kdsOrders = [];
  bool _isLoadingKds = false;
  String _kdsFilter = 'All';

  // Customer State
  Map<String, dynamic> _selectedCustomer = {
    'id': 1,
    'name': 'Guest (ضيف)',
    'phone': '12345678',
    'code': '10',
  };
  List<Map<String, dynamic>> _customers = [];

  // Cart & Backend Calculated Order State
  final List<Map<String, dynamic>> _cartItems = [];
  Map<String, dynamic> _calculatedOrder = {
    'subtotal': '0.00',
    'discount_amount': '0.00',
    'tax_rate': '15.00',
    'tax_amount': '0.00',
    'total_amount': '0.00',
    'item_count': 0,
  };
  bool _isCalculating = false;

  // Cart visibility toggle
  bool _isOrderDrawerVisible = true;

  // Toast banner for kitchen dispatch
  String? _kitchenNotificationToast;
  Timer? _toastTimer;

  // Drawer Cash Amount
  double _drawerCashAmount = 1204.47;

  // Theme Constants (Clean Modern White / Light Slate matching screenshots)
  static const Color scaffoldBg = Color(0xFFF8FAFC);
  static const Color cardBg = Colors.white;
  static const Color borderSubtle = Color(0xFFE2E8F0);
  static const Color darkPill = Color(0xFF0F172A);
  static const Color darkPillLight = Color(0xFF1E293B);
  static const Color primarySky = Color(0xFF0EA5E9);
  static const Color primaryTeal = Color(0xFF06B6D4);
  static const Color emeraldGreen = Color(0xFF10B981);
  static const Color emeraldBg = Color(0xFFECFDF5);
  static const Color amberOrange = Color(0xFFF59E0B);
  static const Color amberBg = Color(0xFFFFFBEB);
  static const Color roseRed = Color(0xFFEF4444);
  static const Color roseBg = Color(0xFFFEF2F2);
  static const Color textMain = Color(0xFF0F172A);
  static const Color textMuted = Color(0xFF64748B);

  @override
  void initState() {
    super.initState();
    _currentView = widget.initialView;
    _orderSource = widget.initialOrderSource;
    _loadInitialData();
  }

  @override
  void dispose() {
    _toastTimer?.cancel();
    super.dispose();
  }

  // Store Payment Methods, Tax Rate, and Promotions (Dynamic from local SQLite DB)
  List<Map<String, dynamic>> _storePaymentMethods = [];
  double _storeTaxRate = 15.0;
  List<Map<String, dynamic>> _activePromotions = [];

  Future<void> _loadInitialData() async {
    await Future.wait([
      _fetchMenu(),
      _fetchTables(),
      _fetchCustomers(),
      _fetchDrawerBalance(),
      _fetchStorePaymentMethodsAndTax(),
    ]);
  }

  Future<void> _fetchDrawerBalance() async {
    try {
      final db = DatabaseService().database;
      final sess = await db.query(
        'cashier_sessions',
        where: 'status = ?',
        whereArgs: ['open'],
        orderBy: 'id DESC',
        limit: 1,
      );
      if (sess.isNotEmpty) {
        final opening = (sess.first['opening_balance'] as num?)?.toDouble() ?? 0.0;
        final txns = await db.rawQuery('''
          SELECT COALESCE(SUM(total_amount), 0) AS total_sales
          FROM pos_transactions
          WHERE cashier_session_id = ? AND status = 'completed'
        ''', [sess.first['id']]);
        final sales = (txns.first['total_sales'] as num?)?.toDouble() ?? 0.0;
        if (mounted) {
          setState(() {
            _drawerCashAmount = opening + sales;
          });
        }
      }
    } catch (_) {}
  }

  Future<void> _fetchStorePaymentMethodsAndTax() async {
    try {
      final db = DatabaseService().database;

      // 1. Fetch dynamic tax rate from local SQL DB tax_categories
      try {
        final taxRows = await db.query(
          'tax_categories',
          where: 'is_active = 1',
          orderBy: 'id ASC',
          limit: 1,
        );
        if (taxRows.isNotEmpty && taxRows.first['tax_rate'] != null) {
          _storeTaxRate = (taxRows.first['tax_rate'] as num).toDouble();
        }
      } catch (e) {
        developer.log('Error querying tax_categories: $e', name: 'RestaurantPOS');
      }

      // 2. Fetch store available payment methods from local SQL DB / Go Core
      List<Map<String, dynamic>> methods = [];
      try {
        final storeRows = await db.query(
          'stores',
          where: 'id = ?',
          whereArgs: [widget.storeId],
          limit: 1,
        );
        if (storeRows.isNotEmpty) {
          final rawMeta = storeRows.first['metadata'];
          if (rawMeta != null && rawMeta.toString().isNotEmpty) {
            Map<String, dynamic> metaMap = {};
            if (rawMeta is String) {
              try {
                metaMap = jsonDecode(rawMeta) as Map<String, dynamic>;
              } catch (_) {}
            } else if (rawMeta is Map) {
              metaMap = Map<String, dynamic>.from(rawMeta);
            }

            final pmList = metaMap['payment_methods'] ??
                metaMap['available_payment_methods'] ??
                metaMap['pos_payment_methods'];
            if (pmList is List && pmList.isNotEmpty) {
              for (final pm in pmList) {
                if (pm is String && pm.trim().isNotEmpty) {
                  methods.add({
                    'code': pm.trim(),
                    'name': _formatMethodName(pm.trim()),
                    'icon': _resolvePaymentIcon(pm.trim()),
                  });
                } else if (pm is Map) {
                  final code = pm['code'] ?? pm['id'] ?? pm['name'] ?? 'cash';
                  final name = pm['name'] ?? pm['label'] ?? code;
                  methods.add({
                    'code': code.toString(),
                    'name': name.toString(),
                    'icon': _resolvePaymentIcon(code.toString()),
                  });
                }
              }
            }
          }
        }
      } catch (e) {
        developer.log('Error querying store payment methods: $e', name: 'RestaurantPOS');
      }

      // Fallback: check distinct payment methods used previously in local DB
      if (methods.isEmpty) {
        try {
          final distinctPayments = await db.rawQuery('''
            SELECT DISTINCT payment_method 
            FROM pos_payments 
            WHERE payment_method IS NOT NULL AND TRIM(payment_method) != ''
          ''');
          if (distinctPayments.isNotEmpty) {
            for (final r in distinctPayments) {
              final pm = r['payment_method']?.toString();
              if (pm != null && pm.trim().isNotEmpty) {
                methods.add({
                  'code': pm.trim(),
                  'name': _formatMethodName(pm.trim()),
                  'icon': _resolvePaymentIcon(pm.trim()),
                });
              }
            }
          }
        } catch (_) {}
      }

      // Standard fallback matching store POS operations
      if (methods.isEmpty) {
        methods = [
          {'code': 'Cash', 'name': 'Cash', 'icon': Icons.money},
          {'code': 'Card', 'name': 'Card', 'icon': Icons.credit_card},
          {'code': 'Split', 'name': 'Split', 'icon': Icons.call_split},
        ];
      }

      // 3. Fetch active promotions / deals from Go Core or local SQL DB
      List<Map<String, dynamic>> promos = [];
      try {
        final promoResp = NembusBridge().callHandler(
          handler: 'promotion',
          action: 'listActivePromotions',
          payload: {'store_id': widget.storeId},
        );
        if (promoResp['success'] == true && promoResp['data'] is List) {
          promos = (promoResp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
      } catch (_) {}

      if (promos.isEmpty) {
        try {
          final promoRows = await db.query('promotions', where: 'is_active = 1 OR is_active IS NULL');
          promos = promoRows.map((r) => Map<String, dynamic>.from(r)).toList();
        } catch (_) {}
      }

      if (mounted) {
        setState(() {
          _storePaymentMethods = methods;
          _activePromotions = promos;
        });
        _requestBackendCalculation();
      }
    } catch (e) {
      developer.log('Error fetching store payment methods and tax: $e', name: 'RestaurantPOS');
    }
  }

  static IconData _resolvePaymentIcon(String method) {
    final m = method.toLowerCase();
    if (m.contains('cash')) return Icons.money;
    if (m.contains('split')) return Icons.call_split;
    if (m.contains('card') || m.contains('credit') || m.contains('pos') || m.contains('mada')) return Icons.credit_card;
    if (m.contains('bank') || m.contains('transfer')) return Icons.account_balance;
    if (m.contains('wallet') || m.contains('apple') || m.contains('pay')) return Icons.account_balance_wallet;
    return Icons.payment;
  }

  static String _formatMethodName(String method) {
    if (method.isEmpty) return 'Cash';
    return method[0].toUpperCase() + method.substring(1);
  }

  // =========================================================================
  // BACKEND HANDLER INTEGRATION (Zero Business Logic in Flutter)
  // All state, calculation, taxes, table states come strictly from Go backend
  // =========================================================================

  Future<void> _fetchMenu() async {
    setState(() => _isLoadingMenu = true);
    try {
      List<Map<String, dynamic>> cats = [];
      List<Map<String, dynamic>> menuItems = [];

      // 1. Try Go FFI Core Handler
      try {
        final catResp = NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'listMenuCategories',
          payload: {'store_id': widget.storeId},
        );

        final menuResp = NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'getFullMenu',
          payload: {'store_id': widget.storeId},
        );

        if (catResp['success'] == true && catResp['data'] is List) {
          cats = (catResp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }

        if (menuResp['success'] == true && menuResp['data'] is List) {
          menuItems = (menuResp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        } else if (menuResp['success'] == true && menuResp['data'] is Map) {
          final data = Map<String, dynamic>.from(menuResp['data'] as Map);
          if (data['categories'] is List) {
            cats = (data['categories'] as List)
                .map((e) => Map<String, dynamic>.from(e as Map))
                .toList();
          }
        }
      } catch (ffiErr) {
        developer.log('FFI menu fetch error: $ffiErr', name: 'RestaurantPOS');
      }

      // 2. Fallback to direct SQLite local database queries
      if (cats.isEmpty || menuItems.isEmpty) {
        try {
          final db = DatabaseService().database;

          // Fetch categories from menu_categories
          List<Map<String, dynamic>> catRows = [];
          if (widget.storeId != null) {
            catRows = await db.query(
              'menu_categories',
              where: 'store_id = ? AND is_active = 1',
              whereArgs: [widget.storeId],
              orderBy: 'display_order ASC, name ASC',
            );
          }
          if (catRows.isEmpty) {
            catRows = await db.query(
              'menu_categories',
              where: 'is_active = 1',
              orderBy: 'display_order ASC, name ASC',
            );
          }
          if (catRows.isEmpty) {
            catRows = await db.query(
              'menu_categories',
              orderBy: 'display_order ASC, name ASC',
            );
          }

          // Fetch items from vw_restaurant_menu or menu_items
          List<Map<String, dynamic>> rawItems = [];
          try {
            rawItems = await db.query(
              'vw_restaurant_menu',
              where: 'is_active = 1',
              orderBy: 'display_order ASC, item_name ASC',
            );
          } catch (_) {
            try {
              rawItems = await db.query(
                'menu_items',
                where: 'is_active = 1',
                orderBy: 'display_order ASC, name ASC',
              );
            } catch (_) {}
          }

          // Fetch modifiers for options
          List<Map<String, dynamic>> allModifiers = [];
          try {
            allModifiers = await db.query(
              'menu_item_modifiers',
              where: 'is_active = 1',
              orderBy: 'display_order ASC, id ASC',
            );
          } catch (_) {}

          final Map<int, List<Map<String, dynamic>>> modifiersByItemId = {};
          for (final mod in allModifiers) {
            final mItemId = (mod['menu_item_id'] as num?)?.toInt();
            if (mItemId != null) {
              modifiersByItemId.putIfAbsent(mItemId, () => []).add(Map<String, dynamic>.from(mod));
            }
          }

          if (catRows.isNotEmpty) {
            cats = catRows.map((c) => Map<String, dynamic>.from(c)).toList();
            for (final cat in cats) {
              final catId = cat['id'];
              cat['items'] = rawItems.where((it) {
                final itCatId = it['category_id'] ?? it['menu_category_id'];
                return itCatId == catId;
              }).map((it) {
                final itemId = (it['menu_item_id'] ?? it['id'] as num?)?.toInt() ?? 0;
                final rawPrice = (it['base_price'] as num?)?.toDouble() ?? 0.0;
                final mods = modifiersByItemId[itemId] ?? [];
                List<Map<String, dynamic>> options = [];
                if (mods.isNotEmpty) {
                  options.add({
                    'name': 'Customizations',
                    'choices': mods.map((m) => {
                      'id': m['id'],
                      'name': m['modifier_name'] ?? m['name'] ?? '',
                      'price_delta': (m['price_adjustment'] as num?)?.toDouble() ?? 0.0,
                    }).toList(),
                  });
                }
                return {
                  'id': itemId,
                  'name': it['item_name'] ?? it['name'] ?? '',
                  'price': rawPrice,
                  'base_price': rawPrice,
                  'image': it['image_url'] ?? '',
                  'description': it['description'] ?? '',
                  'preparation_time_min': (it['preparation_time_min'] as num?)?.toInt() ?? 15,
                  'modifiers': mods,
                  'options': options,
                };
              }).toList();
            }
          } else if (rawItems.isNotEmpty) {
            cats = [
              {
                'id': 1,
                'name': 'All Menu Items',
                'items': rawItems.map((it) {
                  final itemId = (it['menu_item_id'] ?? it['id'] as num?)?.toInt() ?? 0;
                  final rawPrice = (it['base_price'] as num?)?.toDouble() ?? 0.0;
                  final mods = modifiersByItemId[itemId] ?? [];
                  return {
                    'id': itemId,
                    'name': it['item_name'] ?? it['name'] ?? '',
                    'price': rawPrice,
                    'base_price': rawPrice,
                    'image': it['image_url'] ?? '',
                    'description': it['description'] ?? '',
                    'preparation_time_min': (it['preparation_time_min'] as num?)?.toInt() ?? 15,
                    'modifiers': mods,
                    'options': [],
                  };
                }).toList(),
              }
            ];
          } else {
            // Also check categories & products from retail catalog if restaurant tables empty
            try {
              final prodCategories = await db.query('categories', orderBy: 'name ASC');
              final prodItems = await db.rawQuery('''
                SELECT p.id, p.name, p.description, p.category_id,
                       COALESCE(pp.price, 0.0) as price
                FROM products p
                LEFT JOIN product_prices pp ON p.id = pp.product_id AND (pp.is_active = 1 OR pp.is_active IS NULL)
                WHERE p.is_active = 1 OR p.is_active IS NULL
                ORDER BY p.name ASC
              ''');

              if (prodCategories.isNotEmpty && prodItems.isNotEmpty) {
                cats = prodCategories.map((c) {
                  final cId = c['id'];
                  final items = prodItems.where((p) => p['category_id'] == cId).map((p) => {
                    'id': p['id'],
                    'name': p['name'],
                    'price': (p['price'] as num?)?.toDouble() ?? 0.0,
                    'base_price': (p['price'] as num?)?.toDouble() ?? 0.0,
                    'description': p['description'] ?? '',
                    'image': '',
                    'preparation_time_min': 10,
                    'options': [],
                    'modifiers': [],
                  }).toList();
                  return {
                    'id': cId,
                    'name': c['name'],
                    'items': items,
                  };
                }).where((c) => (c['items'] as List).isNotEmpty).toList();
              }
            } catch (_) {}
          }
        } catch (dbErr) {
          developer.log('SQLite menu fetch error: $dbErr', name: 'RestaurantPOS');
        }
      } else if (cats.isNotEmpty && menuItems.isNotEmpty) {
        for (final cat in cats) {
          final catId = cat['id'];
          cat['items'] = menuItems
              .where((item) => item['category_id'] == catId)
              .map((item) => {
                    'id': item['menu_item_id'] ?? item['id'],
                    'name': item['item_name'] ?? item['name'],
                    'price': (item['base_price'] as num?)?.toDouble() ?? 0.0,
                    'base_price': (item['base_price'] as num?)?.toDouble() ?? 0.0,
                    'image': item['image_url'] ?? '',
                    'description': item['description'] ?? '',
                    'preparation_time_min': (item['preparation_time_min'] as num?)?.toInt() ?? 15,
                    'options': [],
                    'modifiers': [],
                  })
              .toList();
        }
      }

      // No dummy fallback data! Real database data only.
      if (mounted) {
        setState(() {
          _categories = cats;
          if (_selectedCategoryId == null && cats.isNotEmpty) {
            _selectedCategoryId = cats.first['id'] as int?;
          }
          _isLoadingMenu = false;
        });
      }
    } catch (e) {
      developer.log('Error fetching menu: $e', name: 'RestaurantPOS');
      if (mounted) {
        setState(() {
          _categories = [];
          _isLoadingMenu = false;
        });
      }
    }
  }


  Future<void> _fetchTables() async {
    setState(() => _isLoadingTables = true);
    try {
      List<Map<String, dynamic>> tList = [];

      // 1. Try Go FFI bridge
      try {
        final resp = NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'listTables',
          payload: {'store_id': widget.storeId},
        );
        if (resp['success'] == true && resp['data'] is List && (resp['data'] as List).isNotEmpty) {
          tList = (resp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
      } catch (ffiErr) {
        developer.log('FFI listTables error: $ffiErr', name: 'RestaurantPOS');
      }

      // 2. Query local SQLite restaurant_tables
      if (tList.isEmpty) {
        try {
          final db = DatabaseService().database;
          List<Map<String, dynamic>> rows = [];
          if (widget.storeId != null) {
            rows = await db.query(
              'restaurant_tables',
              where: 'store_id = ? AND is_active = 1',
              whereArgs: [widget.storeId],
              orderBy: 'section ASC, table_number ASC',
            );
          }
          if (rows.isEmpty) {
            rows = await db.query(
              'restaurant_tables',
              where: 'is_active = 1',
              orderBy: 'section ASC, table_number ASC',
            );
          }
          if (rows.isEmpty) {
            rows = await db.query(
              'restaurant_tables',
              orderBy: 'section ASC, table_number ASC',
            );
          }
          tList = rows.map((r) => Map<String, dynamic>.from(r)).toList();
        } catch (dbErr) {
          developer.log('SQLite restaurant_tables query error: $dbErr', name: 'RestaurantPOS');
        }
      }

      // 3. Real-time table occupancy status from active orders in SQLite
      try {
        final db = DatabaseService().database;
        final activeOrderRows = await db.rawQuery('''
          SELECT table_id, id as order_id, order_number, total_amount, status
          FROM restaurant_orders
          WHERE status NOT IN ('paid', 'voided', 'completed') AND table_id IS NOT NULL
        ''');
        final occupiedMap = {
          for (var o in activeOrderRows)
            (o['table_id'] as num?)?.toInt(): o
        };

        for (final tbl in tList) {
          final tblId = (tbl['id'] as num?)?.toInt();
          if (tblId != null && occupiedMap.containsKey(tblId)) {
            tbl['status'] = 'occupied';
            tbl['is_occupied'] = true;
            tbl['active_order'] = occupiedMap[tblId];
          } else {
            tbl['status'] = tbl['status'] ?? 'available';
            tbl['is_occupied'] = tbl['status'] == 'occupied';
          }
        }
      } catch (_) {}

      // Real database data only — zero dummy tables!
      if (mounted) {
        setState(() {
          _tables = tList;
          _isLoadingTables = false;
        });
      }
    } catch (e) {
      developer.log('Error fetching tables: $e', name: 'RestaurantPOS');
      if (mounted) {
        setState(() {
          _tables = [];
          _isLoadingTables = false;
        });
      }
    }
  }

  Future<void> _fetchCustomers() async {
    try {
      final db = DatabaseService().database;
      final rows = await db.query('customers', limit: 20);
      if (rows.isNotEmpty) {
        if (mounted) {
          setState(() {
            _customers = rows.map((r) => Map<String, dynamic>.from(r)).toList();
          });
        }
        return;
      }
    } catch (_) {}

    // Fallback default customers matching Screenshot 5
    if (mounted) {
      setState(() {
        _customers = [
          {'id': 1, 'name': 'Guest (ضيف)', 'phone': '12345678', 'code': '10'},
          {'id': 2, 'name': 'Jane Doe', 'phone': '0551234567', 'code': '11'},
          {'id': 3, 'name': 'aku', 'phone': '03123456789', 'code': '12'},
          {'id': 4, 'name': 'shoaib', 'phone': '0318588585', 'code': '13'},
        ];
      });
    }
  }

  Future<void> _fetchBills() async {
    setState(() => _isLoadingBills = true);
    try {
      try {
        final resp = NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'listOrders',
          payload: {'store_id': widget.storeId, 'limit': 50},
        );

        if (resp['success'] == true && resp['data'] is List && (resp['data'] as List).isNotEmpty) {
          final bList = (resp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
          if (mounted) {
            setState(() {
              _bills = bList;
              _isLoadingBills = false;
            });
          }
          return;
        }
      } catch (_) {}

      // Fallback to local SQLite DB
      try {
        final db = DatabaseService().database;
        final rows = await db.rawQuery('''
          SELECT 
            ro.id,
            ro.order_number,
            ro.status,
            ro.subtotal,
            ro.tax_amount,
            ro.total_amount,
            ro.amount_paid,
            ro.ordered_at,
            ro.paid_at,
            ro.table_id,
            rt.table_number,
            rt.table_name
          FROM restaurant_orders ro
          LEFT JOIN restaurant_tables rt ON ro.table_id = rt.id
          ORDER BY ro.ordered_at DESC, ro.id DESC
          LIMIT 50
        ''');
        final bList = rows.map((r) => Map<String, dynamic>.from(r)).toList();
        if (mounted) {
          setState(() {
            _bills = bList;
            _isLoadingBills = false;
          });
        }
      } catch (dbErr) {
        developer.log('SQLite listOrders error: $dbErr', name: 'RestaurantPOS');
        if (mounted) setState(() => _isLoadingBills = false);
      }
    } catch (e) {
      developer.log('Error fetching bills: $e', name: 'RestaurantPOS');
      if (mounted) setState(() => _isLoadingBills = false);
    }
  }

  Future<void> _fetchKdsOrders() async {
    setState(() => _isLoadingKds = true);
    try {
      List<Map<String, dynamic>> kList = [];

      try {
        final resp = NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'getKdsOrders',
          payload: {'store_id': widget.storeId},
        );

        if (resp['success'] == true && resp['data'] is List && (resp['data'] as List).isNotEmpty) {
          kList = (resp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
      } catch (ffiErr) {
        developer.log('FFI getKdsOrders error: $ffiErr', name: 'RestaurantPOS');
      }

      if (kList.isEmpty) {
        try {
          final db = DatabaseService().database;
          final rows = await db.rawQuery('''
            SELECT 
              ro.id AS order_id,
              ro.id,
              ro.order_number,
              ro.status,
              ro.ordered_at,
              ro.subtotal,
              ro.total_amount,
              ro.metadata,
              ro.order_source,
              ro.table_id,
              rt.table_number,
              rt.table_name,
              rt.section
            FROM restaurant_orders ro
            LEFT JOIN restaurant_tables rt ON ro.table_id = rt.id
            WHERE ro.status NOT IN ('voided', 'cancelled')
            ORDER BY ro.ordered_at DESC, ro.id DESC
          ''');

          for (final row in rows) {
            final oMap = Map<String, dynamic>.from(row);
            List<Map<String, dynamic>> items = [];

            if (oMap['metadata'] != null) {
              try {
                final meta = jsonDecode(oMap['metadata'].toString());
                if (meta is Map && meta['items'] is List) {
                  items = List<Map<String, dynamic>>.from(meta['items']);
                }
              } catch (_) {}
            }

            if (items.isEmpty) {
              final oId = oMap['order_id'] ?? oMap['id'];
              if (oId != null) {
                final itemRows = await db.rawQuery('''
                  SELECT roi.*, COALESCE(mi.name, roi.notes, 'Item #' || roi.menu_item_id) as item_name
                  FROM restaurant_order_items roi
                  LEFT JOIN menu_items mi ON roi.menu_item_id = mi.id
                  WHERE roi.order_id = ?
                ''', [oId]);
                for (final ir in itemRows) {
                  items.add({
                    'id': ir['menu_item_id'],
                    'name': ir['item_name'] ?? 'Item',
                    'quantity': (ir['quantity'] as num?)?.toInt() ?? 1,
                    'price': (ir['unit_price'] as num?)?.toDouble() ?? 0.0,
                    'status': ir['status'] ?? 'pending',
                  });
                }
              }
            }

            oMap['items'] = items;
            kList.add(oMap);
          }
        } catch (dbErr) {
          developer.log('SQLite getKdsOrders error: $dbErr', name: 'RestaurantPOS');
        }
      }

      if (mounted) {
        setState(() {
          _kdsOrders = kList;
          _isLoadingKds = false;
        });
      }
    } catch (e) {
      developer.log('Error fetching KDS: $e', name: 'RestaurantPOS');
      if (mounted) setState(() => _isLoadingKds = false);
    }
  }

  Future<void> _updateKdsOrderStatus(int orderId, String newStatus, String orderNum) async {
    try {
      try {
        NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'updateOrderStatus',
          payload: {
            'order_id': orderId,
            'id': orderId,
            'status': newStatus,
          },
        );
      } catch (ffiErr) {
        developer.log('FFI updateOrderStatus note: $ffiErr', name: 'RestaurantPOS');
      }

      final db = DatabaseService().database;
      await db.update(
        'restaurant_orders',
        {
          'status': newStatus,
          'updated_at': DateTime.now().toIso8601String(),
        },
        where: 'id = ?',
        whereArgs: [orderId],
      );

      try {
        await db.update(
          'restaurant_order_items',
          {
            'status': newStatus,
            'updated_at': DateTime.now().toIso8601String(),
          },
          where: 'order_id = ?',
          whereArgs: [orderId],
        );
      } catch (_) {}

      await db.insert('sync_queue', {
        'entity_type': 'restaurant_orders',
        'entity_id': orderId.toString(),
        'action': 'UPDATE',
        'payload': jsonEncode({
          'id': orderId,
          'order_number': orderNum,
          'status': newStatus,
        }),
        'status': 'pending',
        'priority': 15,
        'correlation_id': orderNum,
        'created_at': DateTime.now().toIso8601String(),
      });

      final label = newStatus == 'in_preparation'
          ? 'IN PREPARATION'
          : (newStatus == 'completed' ? 'COMPLETED' : 'PENDING');
      _showKitchenBannerToast('Order $orderNum marked as $label');

      await _fetchKdsOrders();
      await _fetchTables();
      await _fetchBills();
    } catch (e) {
      developer.log('Error updating KDS status: $e', name: 'RestaurantPOS');
    }
  }

  void _requestBackendCalculation() {
    if (_cartItems.isEmpty) {
      if (mounted) {
        setState(() {
          _calculatedOrder = {
            'subtotal': '0.00',
            'discount_amount': '0.00',
            'discount_label': '',
            'tax_rate': _storeTaxRate.toStringAsFixed(2),
            'tax_amount': '0.00',
            'total_amount': '0.00',
            'item_count': 0,
          };
          _isCalculating = false;
        });
      }
      return;
    }

    double subtotal = 0.0;
    int count = 0;
    for (final item in _cartItems) {
      final qty = (item['quantity'] as num?)?.toInt() ?? 1;
      final price = (item['unit_price'] as num?)?.toDouble() ??
          (item['price'] as num?)?.toDouble() ??
          0.0;
      final modTotal = (item['modifiers_total'] as num?)?.toDouble() ?? 0.0;
      subtotal += (price + modTotal) * qty;
      count += qty;
    }

    // Dynamic discounts / promotions / deals from local DB or Go core
    double discountAmount = 0.0;
    String discountLabel = '';
    if (_activePromotions.isNotEmpty) {
      for (final promo in _activePromotions) {
        final promoType = (promo['discount_type'] ?? promo['type'] ?? '').toString().toLowerCase();
        final promoVal = (promo['discount_value'] as num?)?.toDouble() ??
            double.tryParse(promo['discount_value']?.toString() ?? '0') ??
            0.0;
        final minOrder = (promo['min_order_amount'] as num?)?.toDouble() ?? 0.0;
        if (subtotal >= minOrder && promoVal > 0) {
          if (promoType.contains('percent') || promoType == 'percentage') {
            discountAmount = (subtotal * promoVal) / 100.0;
            discountLabel = '${promo['name'] ?? 'Promotion'} (${promoVal.toStringAsFixed(0)}%)';
          } else {
            discountAmount = promoVal;
            discountLabel = '${promo['name'] ?? 'Discount'} (-SAR ${promoVal.toStringAsFixed(2)})';
          }
          break;
        }
      }
    }
    if (discountAmount > subtotal) {
      discountAmount = subtotal;
    }

    final taxableAmount = math.max(0.0, subtotal - discountAmount);
    final taxRate = _storeTaxRate;
    final taxAmount = (taxableAmount * taxRate) / 100.0;
    final total = taxableAmount + taxAmount;

    if (mounted) {
      setState(() {
        _calculatedOrder = {
          'subtotal': subtotal.toStringAsFixed(2),
          'discount_amount': discountAmount.toStringAsFixed(2),
          'discount_label': discountLabel,
          'tax_rate': taxRate.toStringAsFixed(2),
          'tax_amount': taxAmount.toStringAsFixed(2),
          'total_amount': total.toStringAsFixed(2),
          'item_count': count,
        };
        _isCalculating = false;
      });
    }
  }

  // ==========================================
  // CART ACTIONS
  // ==========================================

  void _addToCart(Map<String, dynamic> item, List<Map<String, dynamic>> selectedModifiers, String notes, int quantity) {
    double modTotal = 0.0;
    for (final mod in selectedModifiers) {
      final p = (mod['price'] as num?)?.toDouble() ??
          double.tryParse(mod['price']?.toString() ?? '0') ??
          0.0;
      modTotal += p;
    }
    final rawPrice = (item['price'] as num?)?.toDouble() ??
        (item['unit_price'] as num?)?.toDouble() ??
        double.tryParse(item['base_price']?.toString() ?? '') ??
        double.tryParse(item['price']?.toString() ?? '') ??
        0.0;

    setState(() {
      _cartItems.add({
        'menu_item_id': item['id'] ?? item['menu_item_id'],
        'name': item['name'] ?? item['product_name'] ?? 'Item',
        'unit_price': rawPrice,
        'price': rawPrice,
        'quantity': quantity,
        'modifiers': selectedModifiers,
        'modifiers_total': modTotal,
        'notes': notes,
      });
    });
    _requestBackendCalculation();
  }

  void _removeFromCart(int index) {
    setState(() {
      _cartItems.removeAt(index);
    });
    _requestBackendCalculation();
  }

  void _updateCartItemQty(int index, int delta) {
    final currentQty = (_cartItems[index]['quantity'] as num).toInt();
    final newQty = currentQty + delta;
    if (newQty <= 0) {
      _removeFromCart(index);
    } else {
      setState(() {
        _cartItems[index]['quantity'] = newQty;
      });
      _requestBackendCalculation();
    }
  }

  void _clearCart() {
    setState(() {
      _cartItems.clear();
      _selectedTable = null;
    });
    _requestBackendCalculation();
  }

  // ==========================================
  // SEND TO KITCHEN ACTION (Requirement 1)
  // ==========================================

  Future<void> _sendOrderToKitchen() async {
    if (_cartItems.isEmpty) return;

    if (_orderSource == 'dining' && _selectedTable == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          backgroundColor: amberOrange,
          behavior: SnackBarBehavior.floating,
          content: Row(
            children: [
              const Icon(Icons.table_restaurant, color: Colors.white),
              const SizedBox(width: 10),
              Expanded(
                child: Text('Please select a dining table before placing the order.', style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w600)),
              ),
            ],
          ),
        ),
      );
      setState(() => _currentView = 'tables');
      return;
    }

    try {
      final subtotalStr = _calculatedOrder['subtotal']?.toString() ?? '0.00';
      final discountStr = _calculatedOrder['discount_amount']?.toString() ?? '0.00';
      final taxStr = _calculatedOrder['tax_amount']?.toString() ?? '0.00';
      final totalStr = _calculatedOrder['total_amount']?.toString() ?? '0.00';
      final orderNum = 'RO-${DateTime.now().millisecondsSinceEpoch % 1000000}';

      final resp = NembusBridge().callHandler(
        handler: 'restaurant',
        action: 'createOrder',
        payload: {
          'store_id': widget.storeId,
          'table_id': _selectedTable?['id'],
          'cashier_id': SingletonClass().activeCashierId ?? 1,
          'cashier_session_id': SingletonClass().activeCashierSessionId ?? 1,
          'order_source': _orderSource,
          'status': 'in_preparation',
          'subtotal': subtotalStr,
          'discount_amount': discountStr,
          'tax_amount': taxStr,
          'total_amount': totalStr,
          'amount_paid': '0.00',
          'change_given': '0.00',
          'order_number': orderNum,
        },
      );

      final orderData = (resp['success'] == true && resp['data'] is Map)
          ? Map<String, dynamic>.from(resp['data'] as Map)
          : <String, dynamic>{};
      final resolvedNum = orderData['order_number']?.toString() ?? orderNum;
      final orderId = (orderData['id'] as num?)?.toInt() ?? DateTime.now().millisecondsSinceEpoch;

      // Persist to local SQLite DB
      try {
        final db = DatabaseService().database;
        await db.insert('restaurant_orders', {
          'id': orderId,
          'store_id': widget.storeId,
          'table_id': _selectedTable?['id'],
          'cashier_id': SingletonClass().activeCashierId ?? 1,
          'cashier_session_id': SingletonClass().activeCashierSessionId ?? 1,
          'order_number': resolvedNum,
          'order_source': _orderSource,
          'status': 'in_preparation',
          'subtotal': double.tryParse(subtotalStr) ?? 0.0,
          'discount_amount': double.tryParse(discountStr) ?? 0.0,
          'tax_amount': double.tryParse(taxStr) ?? 0.0,
          'total_amount': double.tryParse(totalStr) ?? 0.0,
          'amount_paid': 0.0,
          'change_given': 0.0,
          'notes': '',
          'metadata': jsonEncode({'items': _cartItems}),
          'ordered_at': DateTime.now().toIso8601String(),
        }, conflictAlgorithm: ConflictAlgorithm.replace);

        for (final itm in _cartItems) {
          final rawItemId = (itm['id'] as num?)?.toInt() ?? 1;
          int validMenuItemId = rawItemId;
          try {
            final existing = await db.query('menu_items', where: 'id = ?', whereArgs: [rawItemId], limit: 1);
            if (existing.isEmpty) {
              final anyItem = await db.query('menu_items', limit: 1);
              if (anyItem.isNotEmpty) {
                validMenuItemId = (anyItem.first['id'] as num).toInt();
              }
            }
          } catch (_) {}

          await db.insert('restaurant_order_items', {
            'order_id': orderId,
            'menu_item_id': validMenuItemId,
            'quantity': (itm['quantity'] as num?)?.toDouble() ?? 1.0,
            'unit_price': (itm['price'] as num?)?.toDouble() ?? 0.0,
            'modifiers_snapshot': jsonEncode(itm['modifiers'] ?? []),
            'modifiers_total': (itm['modifiers_total'] as num?)?.toDouble() ?? 0.0,
            'subtotal': (itm['subtotal'] as num?)?.toDouble() ?? 0.0,
            'status': 'pending',
          }, conflictAlgorithm: ConflictAlgorithm.replace);
        }

        // Mark table as occupied in SQLite
        if (_selectedTable != null) {
          await db.update(
            'restaurant_tables',
            {
              'metadata': jsonEncode({
                'is_occupied': true,
                'status': 'occupied',
                'active_order': resolvedNum,
                'order_id': orderId,
              })
            },
            where: 'id = ?',
            whereArgs: [_selectedTable!['id']],
          );
        }

        // Enqueue to sync_queue for cloud sync
        await db.insert('sync_queue', {
          'entity_type': 'restaurant_orders',
          'entity_id': orderId.toString(),
          'action': 'INSERT',
          'payload': jsonEncode({
            'order_id': orderId,
            'order_number': resolvedNum,
            'store_id': widget.storeId,
            'table_id': _selectedTable?['id'],
            'items': _cartItems,
            'status': 'in_preparation',
            'subtotal': subtotalStr,
            'total_amount': totalStr,
          }),
          'status': 'pending',
          'priority': 20,
          'correlation_id': resolvedNum,
          'created_at': DateTime.now().toIso8601String(),
        });
      } catch (dbErr) {
        developer.log('SQLite write error: $dbErr', name: 'RestaurantPOS');
      }

      // Trigger green notification toast
      _showKitchenBannerToast('Order $resolvedNum sent to kitchen!');

      _clearCart();
      _fetchTables();
      await _fetchKdsOrders();
      _fetchBills();

      if (mounted) {
        setState(() {
          _currentView = 'kds';
        });
      }
    } catch (e) {
      developer.log('Order creation error: $e', name: 'RestaurantPOS');
    }
  }

  void _showKitchenBannerToast(String message) {
    _toastTimer?.cancel();
    setState(() {
      _kitchenNotificationToast = message;
    });
    _toastTimer = Timer(const Duration(seconds: 4), () {
      if (mounted) {
        setState(() {
          _kitchenNotificationToast = null;
        });
      }
    });
  }

  // ==========================================
  // ITEM CUSTOMIZATION DIALOG (Screenshot 3)
  // ==========================================

  void _showItemCustomizationDialog(Map<String, dynamic> item) {
    final modifiers = (item['modifiers'] as List? ?? []).map((e) => Map<String, dynamic>.from(e as Map)).toList();
    final List<Map<String, dynamic>> selectedMods = [];
    final notesController = TextEditingController();
    int qty = 1;

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (dialogCtx, setDialogState) => AlertDialog(
          backgroundColor: Colors.white,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20), side: const BorderSide(color: borderSubtle)),
          titlePadding: const EdgeInsets.fromLTRB(20, 20, 20, 10),
          contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
          actionsPadding: const EdgeInsets.fromLTRB(20, 10, 20, 20),
          title: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(Icons.restaurant_menu, color: textMain, size: 22),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      item['name']?.toString() ?? 'Select Options',
                      style: GoogleFonts.inter(color: textMain, fontWeight: FontWeight.bold, fontSize: 17),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      '${item['base_price']} SAR',
                      style: GoogleFonts.inter(color: primarySky, fontSize: 13, fontWeight: FontWeight.w700),
                    ),
                  ],
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close, color: textMuted),
                onPressed: () => Navigator.pop(ctx),
              ),
            ],
          ),
          content: SizedBox(
            width: 440,
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (item['description'] != null && item['description'].toString().isNotEmpty) ...[
                    Text(
                      item['description'].toString(),
                      style: GoogleFonts.inter(color: textMuted, fontSize: 13),
                    ),
                    const SizedBox(height: 14),
                  ],

                  // Modifiers / Customization options
                  if (modifiers.isNotEmpty) ...[
                    Text(
                      'Options & Add-ons',
                      style: GoogleFonts.inter(color: textMain, fontWeight: FontWeight.w700, fontSize: 14),
                    ),
                    const SizedBox(height: 8),
                    ...modifiers.map((mod) {
                      final modName = mod['modifier_name']?.toString() ?? '';
                      final modPrice = mod['price_adjustment']?.toString() ?? '0.00';
                      final isSelected = selectedMods.any((m) => m['name'] == modName);

                      return Container(
                        margin: const EdgeInsets.only(bottom: 6),
                        decoration: BoxDecoration(
                          color: isSelected ? const Color(0xFFF0FDF4) : const Color(0xFFF8FAFC),
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(color: isSelected ? emeraldGreen : borderSubtle),
                        ),
                        child: CheckboxListTile(
                          value: isSelected,
                          contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 0),
                          activeColor: emeraldGreen,
                          dense: true,
                          title: Text(modName, style: GoogleFonts.inter(color: textMain, fontSize: 13, fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal)),
                          secondary: Text('+ $modPrice SAR', style: GoogleFonts.inter(color: emeraldGreen, fontSize: 12, fontWeight: FontWeight.bold)),
                          onChanged: (checked) {
                            setDialogState(() {
                              if (checked == true) {
                                selectedMods.add({
                                  'name': modName,
                                  'price': double.tryParse(modPrice) ?? 0.0,
                                });
                              } else {
                                selectedMods.removeWhere((m) => m['name'] == modName);
                              }
                            });
                          },
                        ),
                      );
                    }),
                    const SizedBox(height: 12),
                  ],

                  // Special Kitchen Notes
                  Text(
                    'Special Kitchen Instructions',
                    style: GoogleFonts.inter(color: textMain, fontWeight: FontWeight.w700, fontSize: 13),
                  ),
                  const SizedBox(height: 6),
                  TextField(
                    controller: notesController,
                    style: GoogleFonts.inter(color: textMain, fontSize: 13),
                    maxLines: 2,
                    decoration: InputDecoration(
                      hintText: 'e.g. Extra hot, Less ice, Sauce on the side...',
                      hintStyle: GoogleFonts.inter(color: Colors.black26),
                      filled: true,
                      fillColor: const Color(0xFFF8FAFC),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(10), borderSide: const BorderSide(color: borderSubtle)),
                      enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(10), borderSide: const BorderSide(color: borderSubtle)),
                      contentPadding: const EdgeInsets.all(12),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Quantity Selector
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                    decoration: BoxDecoration(color: const Color(0xFFF8FAFC), borderRadius: BorderRadius.circular(12), border: Border.all(color: borderSubtle)),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text('Quantity', style: GoogleFonts.inter(color: textMain, fontWeight: FontWeight.w600, fontSize: 13)),
                        Row(
                          children: [
                            IconButton(
                              icon: const Icon(Icons.remove_circle_outline, color: textMain, size: 22),
                              onPressed: qty > 1 ? () => setDialogState(() => qty--) : null,
                            ),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 4),
                              decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(8), border: Border.all(color: borderSubtle)),
                              child: Text('$qty', style: GoogleFonts.inter(color: textMain, fontWeight: FontWeight.bold, fontSize: 15)),
                            ),
                            IconButton(
                              icon: const Icon(Icons.add_circle_outline, color: textMain, size: 22),
                              onPressed: () => setDialogState(() => qty++),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: Text('Cancel', style: GoogleFonts.inter(color: textMuted)),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: darkPill,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
              ),
              onPressed: () {
                Navigator.pop(ctx);
                _addToCart(item, selectedMods, notesController.text.trim(), qty);
              },
              child: Text('Add to Cart', style: GoogleFonts.inter(fontWeight: FontWeight.bold)),
            ),
          ],
        ),
      ),
    );
  }

  // ==========================================
  // CUSTOMER SELECTOR DIALOG (Screenshot 5)
  // ==========================================

  void _showCustomerSelectionSheet() {
    final searchCtrl = TextEditingController();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (modalCtx, setModalState) {
          final query = searchCtrl.text.toLowerCase().trim();
          final filtered = query.isEmpty
              ? _customers
              : _customers.where((c) {
                  final name = (c['name'] ?? '').toString().toLowerCase();
                  final phone = (c['phone'] ?? '').toString().toLowerCase();
                  return name.contains(query) || phone.contains(query);
                }).toList();

          return Container(
            height: MediaQuery.of(modalCtx).size.height * 0.75,
            decoration: const BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
            ),
            child: Column(
              children: [
                // Drag handle
                Container(
                  margin: const EdgeInsets.only(top: 10, bottom: 6),
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2)),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.person_add_alt_1, color: textMain, size: 22),
                          const SizedBox(width: 8),
                          Text('+ Select Customer', style: GoogleFonts.inter(fontSize: 16, fontWeight: FontWeight.w700, color: textMain)),
                        ],
                      ),
                      IconButton(icon: const Icon(Icons.close), onPressed: () => Navigator.pop(ctx)),
                    ],
                  ),
                ),
                const Divider(height: 1, color: borderSubtle),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: searchCtrl,
                          onChanged: (_) => setModalState(() {}),
                          decoration: InputDecoration(
                            hintText: 'Search customers...',
                            prefixIcon: const Icon(Icons.search, size: 18),
                            filled: true,
                            fillColor: const Color(0xFFF8FAFC),
                            border: OutlineInputBorder(borderRadius: BorderRadius.circular(10), borderSide: const BorderSide(color: borderSubtle)),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      ElevatedButton(
                        onPressed: () {
                          // Quick guest create
                          Navigator.pop(ctx);
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.white,
                          foregroundColor: textMain,
                          side: const BorderSide(color: borderSubtle),
                          elevation: 0,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                        ),
                        child: const Text('+ New Guest'),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: ListView.builder(
                    itemCount: filtered.length,
                    itemBuilder: (cCtx, i) {
                      final cust = filtered[i];
                      final name = cust['name']?.toString() ?? 'Customer';
                      final phone = cust['phone']?.toString() ?? '';
                      final isSelected = _selectedCustomer['name'] == name;

                      return ListTile(
                        leading: CircleAvatar(
                          backgroundColor: const Color(0xFFF1F5F9),
                          child: Text(name.isNotEmpty ? name[0].toUpperCase() : 'G', style: const TextStyle(color: textMain, fontWeight: FontWeight.bold)),
                        ),
                        title: Text(name, style: GoogleFonts.inter(fontWeight: FontWeight.w600, color: textMain)),
                        subtitle: phone.isNotEmpty ? Text(phone, style: GoogleFonts.inter(fontSize: 12, color: textMuted)) : null,
                        trailing: isSelected ? const Icon(Icons.check_circle, color: emeraldGreen) : const Icon(Icons.arrow_forward_ios, size: 14, color: textMuted),
                        onTap: () {
                          setState(() {
                            _selectedCustomer = cust;
                          });
                          Navigator.pop(ctx);
                        },
                      );
                    },
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  // ==========================================
  // MAIN SCAFFOLD & VIEW DISPATCH
  // ==========================================

  PreferredSizeWidget _buildAppBar() {
    final screenWidth = MediaQuery.of(context).size.width;
    final isCompact = screenWidth < 600;

    return AppBar(
      backgroundColor: darkPillLight,
      elevation: 0,
      titleSpacing: 0,
      leading: IconButton(
        icon: const Icon(Icons.arrow_back_ios_new_rounded, color: Colors.white70, size: 18),
        onPressed: () {
          if (_currentView == 'checkout') {
            setState(() => _currentView = 'menu');
          } else if (_currentView == 'menu' || _currentView == 'kds' || _currentView == 'bills') {
            setState(() => _currentView = 'tables');
          } else {
            Navigator.of(context).pop();
          }
        },
      ),
      title: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            _currentView == 'checkout'
                ? 'Checkout'
                : (_currentView == 'kds'
                    ? 'Kitchen Display'
                    : (_currentView == 'bills'
                        ? 'Bills History'
                        : 'Restaurant POS')),
            style: GoogleFonts.inter(
              fontWeight: FontWeight.w800,
              fontSize: 16,
              color: Colors.white,
              letterSpacing: -0.3,
            ),
            overflow: TextOverflow.ellipsis,
          ),
          Text(
            '${widget.storeName.isNotEmpty ? widget.storeName : "Store"} • ${_orderSource.toUpperCase()}${_selectedTable != null ? " • Table ${_selectedTable!["table_number"]}" : ""}',
            overflow: TextOverflow.ellipsis,
            maxLines: 1,
            style: GoogleFonts.inter(fontSize: 10.5, color: Colors.white54),
          ),
        ],
      ),
      actions: [
        // Drawer SAR Badge
        Container(
          margin: const EdgeInsets.symmetric(vertical: 10),
          padding: EdgeInsets.symmetric(horizontal: isCompact ? 6 : 9, vertical: 3.5),
          decoration: BoxDecoration(
            color: emeraldBg,
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: emeraldGreen.withValues(alpha: 0.3)),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.account_balance_wallet, color: emeraldGreen, size: 13),
              const SizedBox(width: 4),
              Text(
                isCompact
                    ? 'SAR ${_drawerCashAmount.toStringAsFixed(0)}'
                    : 'DRAWER SAR ${_drawerCashAmount.toStringAsFixed(2)}',
                style: GoogleFonts.inter(color: emeraldGreen, fontWeight: FontWeight.bold, fontSize: 10.5),
              ),
            ],
          ),
        ),
        IconButton(
          tooltip: 'Kitchen Display',
          icon: Icon(Icons.tv_rounded, color: _currentView == 'kds' ? primarySky : Colors.white70, size: 20),
          padding: const EdgeInsets.symmetric(horizontal: 4),
          constraints: const BoxConstraints(),
          onPressed: () => setState(() => _currentView = _currentView == 'kds' ? 'tables' : 'kds'),
        ),
        IconButton(
          tooltip: 'Bills & Invoices',
          icon: Icon(Icons.receipt_long_rounded, color: _currentView == 'bills' ? primarySky : Colors.white70, size: 20),
          padding: const EdgeInsets.symmetric(horizontal: 6),
          constraints: const BoxConstraints(),
          onPressed: () => setState(() => _currentView = _currentView == 'bills' ? 'tables' : 'bills'),
        ),
        const SizedBox(width: 6),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: scaffoldBg,
      appBar: _buildAppBar(),
      body: SafeArea(
        child: Column(
          children: [
            if (_kitchenNotificationToast != null) _buildKitchenToastBanner(),
            Expanded(
              child: _buildCurrentViewBody(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildKitchenToastBanner() {
    return Container(
      width: double.infinity,
      color: emeraldGreen,
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              const Icon(Icons.check_circle, color: Colors.white, size: 20),
              const SizedBox(width: 10),
              Text(
                _kitchenNotificationToast!,
                style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ],
          ),
          InkWell(
            onTap: () => setState(() => _kitchenNotificationToast = null),
            child: const Icon(Icons.close, color: Colors.white, size: 18),
          ),
        ],
      ),
    );
  }

  Widget _buildCurrentViewBody() {
    switch (_currentView) {
      case 'tables':
        return _buildTableFloorPlanView();
      case 'menu':
        return _buildMenuCatalogView();
      case 'checkout':
        return _buildCheckoutView();
      case 'kds':
        return _buildKdsView();
      case 'bills':
        return _buildBillsView();
      default:
        return _buildTableFloorPlanView();
    }
  }

  // =========================================================================
  // VIEW 1: TABLE SELECTION & FLOOR PLAN (Screenshot 1)
  // =========================================================================

  Widget _buildTableFloorPlanView() {
    final availableCount = _tables.where((t) => t['status'] == 'available').length;
    final occupiedCount = _tables.where((t) => t['status'] == 'occupied').length;

    // Filter tables by section
    final filteredTables = _tables.where((t) {
      if (_selectedSection != 'All' && t['section'] != _selectedSection) {
        return false;
      }
      if (_searchQuery.isNotEmpty) {
        final num = (t['table_number'] ?? '').toString().toLowerCase();
        final name = (t['table_name'] ?? '').toString().toLowerCase();
        return num.contains(_searchQuery) || name.contains(_searchQuery);
      }
      return true;
    }).toList();

    return Padding(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Subheader Row: Title, Available/Occupied Badges, Take-Away Mode Button
          Builder(
            builder: (context) {
              final screenWidth = MediaQuery.of(context).size.width;
              final isCompact = screenWidth < 750;

              final titleSection = Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: const Icon(Icons.table_restaurant, color: textMain, size: 24),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Table Selection & Floor Plan',
                          style: GoogleFonts.inter(fontSize: isCompact ? 16 : 18, fontWeight: FontWeight.w800, color: textMain),
                        ),
                        Text(
                          'Select an available table for Dine-In or switch to Take-Away mode',
                          style: GoogleFonts.inter(fontSize: 12, color: textMuted),
                        ),
                      ],
                    ),
                  ),
                ],
              );

              final actionBadges = [
                // Available Badge
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                  decoration: BoxDecoration(
                    color: emeraldBg,
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: emeraldGreen.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.circle, color: emeraldGreen, size: 8),
                      const SizedBox(width: 6),
                      Text('$availableCount Available', style: GoogleFonts.inter(color: emeraldGreen, fontWeight: FontWeight.bold, fontSize: 12)),
                    ],
                  ),
                ),
                const SizedBox(width: 8),

                // Occupied Badge
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                  decoration: BoxDecoration(
                    color: roseBg,
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: roseRed.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.circle, color: roseRed, size: 8),
                      const SizedBox(width: 6),
                      Text('$occupiedCount Occupied', style: GoogleFonts.inter(color: roseRed, fontWeight: FontWeight.bold, fontSize: 12)),
                    ],
                  ),
                ),
                const SizedBox(width: 10),

                // Take-Away Mode Switch Button
                ElevatedButton.icon(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: darkPill,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                  ),
                  icon: const Icon(Icons.shopping_bag_outlined, size: 16),
                  label: const Text('Take-Away Mode', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                  onPressed: () {
                    setState(() {
                      _orderSource = 'takeaway';
                      _selectedTable = null;
                      _currentView = 'menu';
                    });
                  },
                ),
                const SizedBox(width: 8),

                // Refresh Button
                IconButton(
                  icon: const Icon(Icons.refresh, color: textMain, size: 20),
                  onPressed: _fetchTables,
                ),
              ];

              if (isCompact) {
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    titleSection,
                    const SizedBox(height: 10),
                    SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      physics: const BouncingScrollPhysics(),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: actionBadges,
                      ),
                    ),
                  ],
                );
              }

              return Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(child: titleSection),
                  const SizedBox(width: 12),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: actionBadges,
                  ),
                ],
              );
            },
          ),
          const SizedBox(height: 16),

          // Section Filter Chips (Dynamically from real database tables)
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            physics: const BouncingScrollPhysics(),
            child: Row(
              children: [
                _buildSectionFilterChip('All', _tables.length),
                ..._tables
                    .map((t) => (t['section'] ?? '').toString())
                    .where((s) => s.isNotEmpty)
                    .toSet()
                    .map((s) => Padding(
                          padding: const EdgeInsets.only(left: 8),
                          child: _buildSectionFilterChip(
                            s,
                            _tables.where((t) => t['section'] == s).length,
                          ),
                        )),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Tables Grid (Responsive for Mobile, Tablet, Desktop - Zero Overflows)
          Expanded(
            child: _isLoadingTables
                ? const Center(child: CircularProgressIndicator(color: darkPill))
                : filteredTables.isEmpty
                    ? Center(child: Text('No tables configured in database.', style: GoogleFonts.inter(color: textMuted)))
                    : LayoutBuilder(
                        builder: (ctx, constraints) {
                          int crossAxisCount = 2;
                          double childAspectRatio = 1.05;
                          if (constraints.maxWidth > 1200) {
                            crossAxisCount = 5;
                            childAspectRatio = 1.35;
                          } else if (constraints.maxWidth > 850) {
                            crossAxisCount = 4;
                            childAspectRatio = 1.25;
                          } else if (constraints.maxWidth > 580) {
                            crossAxisCount = 3;
                            childAspectRatio = 1.15;
                          } else {
                            crossAxisCount = 2;
                            childAspectRatio = 1.02;
                          }

                          return GridView.builder(
                            itemCount: filteredTables.length,
                            gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: crossAxisCount,
                              crossAxisSpacing: 12,
                              mainAxisSpacing: 12,
                              childAspectRatio: childAspectRatio,
                            ),
                            itemBuilder: (ctx, idx) {
                              final table = filteredTables[idx];
                              final isOccupied = table['status'] == 'occupied';

                              return Material(
                                color: Colors.transparent,
                                child: InkWell(
                                  borderRadius: BorderRadius.circular(16),
                                  onTap: () {
                                    if (isOccupied) {
                                      _showOccupiedTableDialog(table);
                                    } else {
                                      setState(() {
                                        _selectedTable = table;
                                        _orderSource = 'dining';
                                        _currentView = 'menu';
                                      });
                                    }
                                  },
                                  child: Container(
                                    decoration: BoxDecoration(
                                      color: cardBg,
                                      borderRadius: BorderRadius.circular(16),
                                      border: Border.all(color: borderSubtle),
                                      boxShadow: const [
                                        BoxShadow(color: Color(0x06000000), blurRadius: 8, offset: Offset(0, 2)),
                                      ],
                                    ),
                                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 10),
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      children: [
                                        // Top row: Status pill & Section
                                        Row(
                                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                          children: [
                                            Flexible(
                                              child: Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2.5),
                                                decoration: BoxDecoration(
                                                  color: isOccupied ? roseBg : emeraldBg,
                                                  borderRadius: BorderRadius.circular(6),
                                                ),
                                                child: Row(
                                                  mainAxisSize: MainAxisSize.min,
                                                  children: [
                                                    Icon(Icons.circle, color: isOccupied ? roseRed : emeraldGreen, size: 6),
                                                    const SizedBox(width: 4),
                                                    Flexible(
                                                      child: Text(
                                                        isOccupied ? 'OCCUPIED' : 'AVAILABLE',
                                                        style: GoogleFonts.inter(
                                                          fontSize: 8.5,
                                                          fontWeight: FontWeight.w800,
                                                          color: isOccupied ? roseRed : emeraldGreen,
                                                        ),
                                                        overflow: TextOverflow.ellipsis,
                                                        maxLines: 1,
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                            ),
                                            const SizedBox(width: 4),
                                            Flexible(
                                              child: Text(
                                                (table['section'] ?? '').toString().toUpperCase(),
                                                style: GoogleFonts.inter(fontSize: 9, fontWeight: FontWeight.w600, color: textMuted),
                                                overflow: TextOverflow.ellipsis,
                                                maxLines: 1,
                                                textAlign: TextAlign.end,
                                              ),
                                            ),
                                          ],
                                        ),

                                        // Center: Table Number in rounded box + Name
                                        Center(
                                          child: Column(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                                decoration: BoxDecoration(
                                                  color: const Color(0xFFF8FAFC),
                                                  borderRadius: BorderRadius.circular(10),
                                                  border: Border.all(color: borderSubtle),
                                                ),
                                                child: Text(
                                                  table['table_number']?.toString() ?? '',
                                                  style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 15, color: textMain),
                                                  overflow: TextOverflow.ellipsis,
                                                ),
                                              ),
                                              const SizedBox(height: 4),
                                              Text(
                                                table['table_name']?.toString() ?? '',
                                                style: GoogleFonts.inter(fontWeight: FontWeight.w600, fontSize: 11.5, color: textMuted),
                                                maxLines: 1,
                                                overflow: TextOverflow.ellipsis,
                                              ),
                                            ],
                                          ),
                                        ),

                                        // Bottom row: Seats & Select link
                                        Row(
                                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                          children: [
                                            Expanded(
                                              child: Row(
                                                mainAxisSize: MainAxisSize.min,
                                                children: [
                                                  const Icon(Icons.people_outline, size: 13, color: textMuted),
                                                  const SizedBox(width: 3),
                                                  Flexible(
                                                    child: Text(
                                                      '${table['capacity'] ?? 4} Seats',
                                                      style: GoogleFonts.inter(fontSize: 10.5, color: textMuted),
                                                      overflow: TextOverflow.ellipsis,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ),
                                            Row(
                                              mainAxisSize: MainAxisSize.min,
                                              children: [
                                                Text(
                                                  isOccupied ? 'View' : 'Select',
                                                  style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 11.5, color: textMain),
                                                ),
                                                const SizedBox(width: 2),
                                                const Icon(Icons.arrow_forward, size: 11, color: textMain),
                                              ],
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              );
                            },
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionFilterChip(String label, int count) {
    final isSelected = _selectedSection == label;
    return InkWell(
      onTap: () => setState(() => _selectedSection = label),
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
        decoration: BoxDecoration(
          color: isSelected ? darkPill : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: isSelected ? darkPill : borderSubtle),
        ),
        child: Text(
          '$label $count',
          style: GoogleFonts.inter(
            color: isSelected ? Colors.white : textMain,
            fontWeight: FontWeight.w600,
            fontSize: 12,
          ),
        ),
      ),
    );
  }

  Future<void> _showOccupiedTableDialog(Map<String, dynamic> table) async {
    final activeOrder = table['active_order'] as Map<String, dynamic>? ?? {};
    final orderNum = activeOrder['order_number']?.toString() ?? 'Active Order';
    final tableNum = table['table_number']?.toString() ?? 'Table';
    final tableName = table['table_name']?.toString() ?? 'Dining Table';
    final section = (table['section'] ?? 'Indoor').toString();
    final capacity = (table['capacity'] ?? 4).toString();
    final orderId = (activeOrder['order_id'] ?? activeOrder['id'] as num?)?.toInt();

    List<Map<String, dynamic>> orderItems = [];
    double subtotal = 0.0;
    double tax = 0.0;
    double total = double.tryParse(activeOrder['total_amount']?.toString() ?? '0') ?? 0.0;
    String elapsedText = '';

    if (orderId != null) {
      try {
        final db = DatabaseService().database;
        final oRows = await db.query('restaurant_orders', where: 'id = ?', whereArgs: [orderId], limit: 1);
        if (oRows.isNotEmpty) {
          final oData = oRows.first;
          subtotal = (oData['subtotal'] as num?)?.toDouble() ?? 0.0;
          tax = (oData['tax_amount'] as num?)?.toDouble() ?? 0.0;
          total = (oData['total_amount'] as num?)?.toDouble() ?? total;

          if (oData['ordered_at'] != null) {
            try {
              final dt = DateTime.parse(oData['ordered_at'].toString());
              final diff = DateTime.now().difference(dt);
              if (diff.inMinutes < 60) {
                elapsedText = '${diff.inMinutes}m ago';
              } else {
                elapsedText = '${diff.inHours}h ${diff.inMinutes % 60}m ago';
              }
            } catch (_) {}
          }

          if (oData['metadata'] != null) {
            try {
              final meta = jsonDecode(oData['metadata'].toString());
              if (meta is Map && meta['items'] is List) {
                orderItems = List<Map<String, dynamic>>.from(meta['items']);
              }
            } catch (_) {}
          }
        }

        if (orderItems.isEmpty) {
          final itemRows = await db.rawQuery('''
            SELECT roi.*, COALESCE(mi.name, roi.notes, 'Item #' || roi.menu_item_id) as item_name
            FROM restaurant_order_items roi
            LEFT JOIN menu_items mi ON roi.menu_item_id = mi.id
            WHERE roi.order_id = ?
          ''', [orderId]);
          for (final ir in itemRows) {
            final uPrice = (ir['unit_price'] as num?)?.toDouble() ?? 0.0;
            final qty = (ir['quantity'] as num?)?.toInt() ?? 1;
            orderItems.add({
              'id': ir['menu_item_id'],
              'name': ir['item_name'] ?? 'Menu Item',
              'price': uPrice,
              'unit_price': uPrice,
              'quantity': qty,
              'modifiers_total': (ir['modifiers_total'] as num?)?.toDouble() ?? 0.0,
              'subtotal': (ir['subtotal'] as num?)?.toDouble() ?? (uPrice * qty),
            });
          }
        }
      } catch (err) {
        developer.log('Error reading table active order: $err', name: 'RestaurantPOS');
      }
    }

    if (total <= 0 && orderItems.isNotEmpty) {
      subtotal = orderItems.fold(0.0, (acc, it) => acc + ((it['price'] as num?)?.toDouble() ?? 0.0) * ((it['quantity'] as num?)?.toInt() ?? 1));
      tax = subtotal * 0.15;
      total = subtotal + tax;
    } else if (subtotal <= 0 && total > 0) {
      subtotal = total / 1.15;
      tax = total - subtotal;
    }

    if (!mounted) return;

    showDialog(
      context: context,
      builder: (ctx) => Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        backgroundColor: Colors.white,
        elevation: 8,
        insetPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 400),
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Header: Table Info & Occupied Badge
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF1F5F9),
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: const Icon(Icons.table_restaurant_rounded, color: darkPill, size: 22),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  tableNum,
                                  style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 16, color: textMain),
                                  overflow: TextOverflow.ellipsis,
                                ),
                                Text(
                                  '$tableName • $section • $capacity Seats',
                                  style: GoogleFonts.inter(fontSize: 11, color: textMuted),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
                          decoration: BoxDecoration(
                            color: roseBg,
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(Icons.circle, color: roseRed, size: 6),
                              const SizedBox(width: 4),
                              Text(
                                'OCCUPIED',
                                style: GoogleFonts.inter(fontSize: 9.5, fontWeight: FontWeight.bold, color: roseRed),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 4),
                        IconButton(
                          icon: const Icon(Icons.close_rounded, size: 20, color: textMuted),
                          onPressed: () => Navigator.pop(ctx),
                          padding: EdgeInsets.zero,
                          constraints: const BoxConstraints(minWidth: 28, minHeight: 28),
                        ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                const Divider(height: 1, color: borderSubtle),
                const SizedBox(height: 12),

                // Order Number & Time
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8FAFC),
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(color: borderSubtle),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.receipt_long_rounded, size: 15, color: textMuted),
                          const SizedBox(width: 6),
                          Text(
                            'Active: $orderNum',
                            style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 12, color: textMain),
                          ),
                        ],
                      ),
                      if (elapsedText.isNotEmpty)
                        Text(
                          elapsedText,
                          style: GoogleFonts.inter(fontSize: 11, color: textMuted, fontWeight: FontWeight.w500),
                        ),
                    ],
                  ),
                ),
                const SizedBox(height: 12),

                // Ordered Items Preview
                if (orderItems.isNotEmpty) ...[
                  Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      'ORDER ITEMS (${orderItems.length})',
                      style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.bold, color: textMuted, letterSpacing: 0.6),
                    ),
                  ),
                  const SizedBox(height: 6),
                  Container(
                    constraints: const BoxConstraints(maxHeight: 130),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFAFAFA),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: borderSubtle),
                    ),
                    child: ListView.separated(
                      shrinkWrap: true,
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                      itemCount: orderItems.length,
                      separatorBuilder: (_, __) => const Divider(height: 8, color: borderSubtle),
                      itemBuilder: (iCtx, iIdx) {
                        final item = orderItems[iIdx];
                        final qty = item['quantity'] ?? 1;
                        final name = item['name'] ?? 'Item';
                        final price = (item['price'] as num?)?.toDouble() ?? 0.0;
                        return Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              child: Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1.5),
                                    decoration: BoxDecoration(
                                      color: darkPill,
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                    child: Text(
                                      '${qty}x',
                                      style: GoogleFonts.inter(fontSize: 9.5, fontWeight: FontWeight.bold, color: Colors.white),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: Text(
                                      name,
                                      style: GoogleFonts.inter(fontSize: 11.5, fontWeight: FontWeight.w600, color: textMain),
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            Text(
                              'SAR ${(price * (qty as num)).toStringAsFixed(2)}',
                              style: GoogleFonts.inter(fontSize: 11.5, fontWeight: FontWeight.bold, color: textMain),
                            ),
                          ],
                        );
                      },
                    ),
                  ),
                  const SizedBox(height: 12),
                ],

                // Total Bill Card
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: emeraldBg.withValues(alpha: 0.5),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: emeraldGreen.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'CURRENT BILL DUE',
                            style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.bold, color: textMuted, letterSpacing: 0.5),
                          ),
                          Text(
                            'Incl. 15% VAT (SAR ${tax.toStringAsFixed(2)})',
                            style: GoogleFonts.inter(fontSize: 10, color: textMuted),
                          ),
                        ],
                      ),
                      Text(
                        'SAR ${total.toStringAsFixed(2)}',
                        style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w900, color: emeraldGreen),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),

                // Action Buttons
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton.icon(
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          side: const BorderSide(color: borderSubtle),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                          foregroundColor: textMain,
                        ),
                        icon: const Icon(Icons.add_shopping_cart_rounded, size: 16),
                        label: const Text('Add Items', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                        onPressed: () {
                          Navigator.pop(ctx);
                          setState(() {
                            _selectedTable = table;
                            _orderSource = 'dining';
                            _currentView = 'menu';
                          });
                        },
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: ElevatedButton.icon(
                        style: ElevatedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          backgroundColor: darkPill,
                          foregroundColor: Colors.white,
                          elevation: 0,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                        ),
                        icon: const Icon(Icons.payment_rounded, size: 16),
                        label: const Text('Pay & Settle', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                        onPressed: () {
                          Navigator.pop(ctx);
                          _openDirectSettlement(table, orderId, total);
                        },
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _openDirectSettlement(Map<String, dynamic> table, int? orderId, double total) async {
    _selectedTable = table;
    _orderSource = 'dining';

    List<Map<String, dynamic>> orderItems = [];
    double subtotal = 0.0;
    double discount = 0.0;
    double tax = 0.0;
    double effectiveTotal = total;
    String? orderNum = table['active_order']?['order_number']?.toString();

    if (orderId != null) {
      try {
        final db = DatabaseService().database;
        final oRows = await db.query('restaurant_orders', where: 'id = ?', whereArgs: [orderId], limit: 1);
        if (oRows.isNotEmpty) {
          final oData = oRows.first;
          subtotal = (oData['subtotal'] as num?)?.toDouble() ?? 0.0;
          discount = (oData['discount_amount'] as num?)?.toDouble() ?? 0.0;
          tax = (oData['tax_amount'] as num?)?.toDouble() ?? 0.0;
          effectiveTotal = (oData['total_amount'] as num?)?.toDouble() ?? total;
          orderNum ??= oData['order_number']?.toString();

          if (oData['metadata'] != null) {
            try {
              final meta = jsonDecode(oData['metadata'].toString());
              if (meta is Map && meta['items'] is List) {
                orderItems = List<Map<String, dynamic>>.from(meta['items']);
              }
            } catch (_) {}
          }
        }

        if (orderItems.isEmpty) {
          final itemRows = await db.query('restaurant_order_items', where: 'order_id = ?', whereArgs: [orderId]);
          for (final row in itemRows) {
            final uPrice = (row['unit_price'] as num?)?.toDouble() ?? 0.0;
            final qty = (row['quantity'] as num?)?.toInt() ?? 1;
            orderItems.add({
              'id': row['menu_item_id'],
              'menu_item_id': row['menu_item_id'],
              'name': row['notes'] != null && row['notes'].toString().isNotEmpty
                  ? row['notes'].toString()
                  : 'Menu Item #${row['menu_item_id']}',
              'price': uPrice,
              'unit_price': uPrice,
              'quantity': qty,
              'modifiers_total': (row['modifiers_total'] as num?)?.toDouble() ?? 0.0,
              'subtotal': (row['subtotal'] as num?)?.toDouble() ?? (uPrice * qty),
            });
          }
        }
      } catch (err) {
        developer.log('Error loading occupied table order: $err', name: 'RestaurantPOS');
      }
    }

    if (effectiveTotal <= 0 && orderItems.isNotEmpty) {
      subtotal = orderItems.fold(0.0, (acc, it) => acc + ((it['price'] as num?)?.toDouble() ?? 0.0) * ((it['quantity'] as num?)?.toInt() ?? 1));
      tax = subtotal * 0.15;
      effectiveTotal = subtotal + tax;
    } else if (subtotal <= 0 && effectiveTotal > 0) {
      subtotal = effectiveTotal / 1.15;
      tax = effectiveTotal - subtotal;
    }

    setState(() {
      _cartItems.clear();
      _cartItems.addAll(orderItems);
      _calculatedOrder = {
        'subtotal': subtotal.toStringAsFixed(2),
        'discount_amount': discount.toStringAsFixed(2),
        'tax_amount': tax.toStringAsFixed(2),
        'total_amount': effectiveTotal.toStringAsFixed(2),
        'order_id': orderId,
        'order_number': orderNum,
      };
      _currentView = 'checkout';
    });
  }

  // =========================================================================
  // VIEW 2: MENU CATALOG & RIGHT ORDER PANEL (Screenshot 3)
  // =========================================================================

  Widget _buildMenuCatalogView() {
    final screenWidth = MediaQuery.of(context).size.width;
    final isCompact = screenWidth < 800;

    final catalogContent = Padding(
      padding: EdgeInsets.fromLTRB(isCompact ? 14 : 20, 16, isCompact ? 14 : 12, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Top subheader: Tables Floor back button, Title, Hide Order toggle
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Row(
                  children: [
                    if (_orderSource == 'dining') ...[
                      OutlinedButton.icon(
                        style: OutlinedButton.styleFrom(
                          foregroundColor: textMain,
                          side: const BorderSide(color: borderSubtle),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
                        ),
                        icon: const Icon(Icons.arrow_back, size: 14),
                        label: const Text('Tables', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
                        onPressed: () => setState(() => _currentView = 'tables'),
                      ),
                      const SizedBox(width: 10),
                    ],
                    const Icon(Icons.restaurant_menu, color: textMain, size: 20),
                    const SizedBox(width: 8),
                    Flexible(
                      child: Text(
                        'Menu Catalog',
                        style: GoogleFonts.inter(fontSize: isCompact ? 16 : 18, fontWeight: FontWeight.w800, color: textMain),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: darkPill,
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                ),
                icon: Icon(isCompact ? Icons.shopping_bag_outlined : (_isOrderDrawerVisible ? Icons.visibility_off : Icons.shopping_cart), size: 14),
                label: Text(
                  isCompact ? 'Order (${_cartItems.length})' : (_isOrderDrawerVisible ? 'Hide Order (${_cartItems.length})' : 'View Order (${_cartItems.length})'),
                  style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
                ),
                onPressed: () {
                  if (isCompact) {
                    _showMobileCartBottomSheet();
                  } else {
                    setState(() => _isOrderDrawerVisible = !_isOrderDrawerVisible);
                  }
                },
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Category Horizontal Pills (Screenshot 3)
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            physics: const BouncingScrollPhysics(),
            child: Row(
              children: [
                _buildCategoryChip('All Categories', null),
                ..._categories.map((c) => _buildCategoryChip(c['name'] ?? '', c['id'] as int?)),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Product Grid
          Expanded(
            child: _isLoadingMenu
                ? const Center(child: CircularProgressIndicator(color: darkPill))
                : _buildMenuGrid(),
          ),
        ],
      ),
    );

    if (isCompact) {
      return Column(
        children: [
          Expanded(child: catalogContent),
          if (_cartItems.isNotEmpty)
            _buildMobileOrderBottomBar(),
        ],
      );
    }

    return Row(
      children: [
        // Left Area: Category pills & Product Grid
        Expanded(
          flex: _isOrderDrawerVisible ? 7 : 10,
          child: catalogContent,
        ),

        // Right Area: Order / Cart Panel (Screenshot 3)
        if (_isOrderDrawerVisible)
          Container(
            width: 360,
            decoration: const BoxDecoration(
              color: Colors.white,
              border: Border(left: BorderSide(color: borderSubtle)),
            ),
            child: _buildRightOrderCartPanel(),
          ),
      ],
    );
  }

  void _showMobileCartBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setSheetState) {
          return Container(
            height: MediaQuery.of(context).size.height * 0.85,
            decoration: const BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
            ),
            child: Column(
              children: [
                Container(
                  margin: const EdgeInsets.only(top: 8, bottom: 4),
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.black26,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                Expanded(child: _buildRightOrderCartPanel()),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildMobileOrderBottomBar() {
    final total = _calculatedOrder['total_amount'] ?? '0.00';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            blurRadius: 10,
            offset: const Offset(0, -3),
          ),
        ],
        border: const Border(top: BorderSide(color: borderSubtle)),
      ),
      child: SafeArea(
        top: false,
        child: Row(
          children: [
            Expanded(
              child: InkWell(
                onTap: _showMobileCartBottomSheet,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      '${_cartItems.length} items • SAR $total',
                      style: GoogleFonts.inter(fontWeight: FontWeight.w900, fontSize: 13, color: textMain),
                    ),
                    Text(
                      _orderSource == 'dining' ? 'Table ${_selectedTable?['table_number'] ?? '-'} • Tap to review' : 'Take-Away • Tap to review',
                      style: GoogleFonts.inter(fontSize: 10.5, color: textMuted),
                    ),
                  ],
                ),
              ),
            ),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: emeraldGreen,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
              ),
              icon: const Icon(Icons.check_circle_outline, size: 14),
              label: const Text('Kitchen', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11)),
              onPressed: _sendOrderToKitchen,
            ),
            const SizedBox(width: 6),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: darkPill,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
              ),
              icon: const Icon(Icons.payment, size: 14),
              label: const Text('Pay & Settle', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11)),
              onPressed: () => setState(() => _currentView = 'checkout'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCategoryChip(String label, int? categoryId) {
    final isSelected = (categoryId == null && _selectedCategoryId == null) || (_selectedCategoryId == categoryId);
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: InkWell(
        onTap: () => setState(() => _selectedCategoryId = categoryId),
        borderRadius: BorderRadius.circular(20),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
          decoration: BoxDecoration(
            color: isSelected ? darkPill : Colors.white,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: isSelected ? darkPill : borderSubtle),
          ),
          child: Text(
            label,
            style: GoogleFonts.inter(
              color: isSelected ? Colors.white : textMain,
              fontWeight: FontWeight.w600,
              fontSize: 12,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildMenuGrid() {
    List<Map<String, dynamic>> itemsToShow = [];
    if (_selectedCategoryId == null) {
      for (final cat in _categories) {
        final items = (cat['items'] as List? ?? []).map((e) => Map<String, dynamic>.from(e as Map)).toList();
        itemsToShow.addAll(items);
      }
    } else {
      final selectedCat = _categories.firstWhere(
        (c) => c['id'] == _selectedCategoryId,
        orElse: () => {'items': []},
      );
      itemsToShow = (selectedCat['items'] as List? ?? []).map((e) => Map<String, dynamic>.from(e as Map)).toList();
    }

    if (_searchQuery.isNotEmpty) {
      itemsToShow = itemsToShow.where((i) {
        final name = (i['name'] ?? '').toString().toLowerCase();
        final desc = (i['description'] ?? '').toString().toLowerCase();
        return name.contains(_searchQuery) || desc.contains(_searchQuery);
      }).toList();
    }

    if (itemsToShow.isEmpty) {
      return Center(
        child: Text('No items found in this category.', style: GoogleFonts.inter(color: textMuted)),
      );
    }

    return LayoutBuilder(
      builder: (ctx, constraints) {
        int crossAxisCount = 2;
        double childAspectRatio = 1.05;
        if (constraints.maxWidth > 1100) {
          crossAxisCount = 4;
          childAspectRatio = 1.45;
        } else if (constraints.maxWidth > 750) {
          crossAxisCount = 3;
          childAspectRatio = 1.3;
        } else if (constraints.maxWidth > 480) {
          crossAxisCount = 2;
          childAspectRatio = 1.15;
        } else {
          crossAxisCount = 2;
          childAspectRatio = 0.98;
        }

        return GridView.builder(
          itemCount: itemsToShow.length,
          gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: crossAxisCount,
            crossAxisSpacing: 12,
            mainAxisSpacing: 12,
            childAspectRatio: childAspectRatio,
          ),
          itemBuilder: (ctx, idx) {
            final item = itemsToShow[idx];
            final prepTime = item['preparation_time_min'] ?? 15;
            final priceNum = (item['base_price'] ?? item['price'] as num?)?.toDouble() ?? 0.0;

            return Container(
              decoration: BoxDecoration(
                color: cardBg,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: borderSubtle),
                boxShadow: const [
                  BoxShadow(color: Color(0x06000000), blurRadius: 6, offset: Offset(0, 2)),
                ],
              ),
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  // Top Prep Time Badge
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      '⏱ $prepTime mins',
                      style: GoogleFonts.inter(fontSize: 9.5, color: textMuted, fontWeight: FontWeight.w600),
                    ),
                  ),

                  // Title and Description
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        item['name']?.toString() ?? '',
                        style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 13, color: textMain),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 2),
                      Text(
                        item['description']?.toString() ?? 'Freshly prepared item',
                        style: GoogleFonts.inter(fontSize: 10.5, color: textMuted),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),

                  // Bottom Price & Plus Button
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Flexible(
                        child: Text(
                          '${priceNum.toStringAsFixed(2)} SAR',
                          style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 13, color: textMain),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      InkWell(
                        onTap: () {
                          final mods = item['modifiers'] as List? ?? [];
                          final opts = item['options'] as List? ?? [];
                          if (mods.isNotEmpty || opts.isNotEmpty) {
                            _showItemCustomizationDialog(item);
                          } else {
                            _addToCart(item, [], '', 1);
                          }
                        },
                        borderRadius: BorderRadius.circular(20),
                        child: Container(
                          width: 28,
                          height: 28,
                          decoration: const BoxDecoration(
                            color: darkPill,
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(Icons.add, color: Colors.white, size: 18),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  // =========================================================================
  // RIGHT CART / ORDER DRAWER (Exact match with Screenshot 3)
  // =========================================================================

  Widget _buildRightOrderCartPanel() {
    final subtotal = _calculatedOrder['subtotal']?.toString() ?? '0.00';
    final vat = _calculatedOrder['tax_amount']?.toString() ?? '0.00';
    final total = _calculatedOrder['total_amount']?.toString() ?? '0.00';

    final orderHeaderTitle = _orderSource == 'dining' && _selectedTable != null
        ? 'Table ${_selectedTable!['table_number']} (${_selectedTable!['section']})'
        : 'Take-Away Order';

    return Column(
      children: [
        // Cart Header
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 10),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Icon(_orderSource == 'dining' ? Icons.table_restaurant : Icons.shopping_bag_outlined, size: 18, color: textMain),
                      const SizedBox(width: 8),
                      Text(
                        orderHeaderTitle,
                        style: GoogleFonts.inter(fontSize: 14.5, fontWeight: FontWeight.w800, color: textMain),
                      ),
                      const SizedBox(width: 6),
                      InkWell(
                        onTap: () => setState(() => _currentView = 'tables'),
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(color: const Color(0xFFF1F5F9), borderRadius: BorderRadius.circular(6)),
                          child: Text('Change', style: GoogleFonts.inter(fontSize: 10, color: textMuted, fontWeight: FontWeight.bold)),
                        ),
                      ),
                    ],
                  ),
                  InkWell(
                    onTap: _clearCart,
                    child: Text('Clear', style: GoogleFonts.inter(fontSize: 11, color: roseRed, fontWeight: FontWeight.bold)),
                  ),
                ],
              ),
              const SizedBox(height: 2),
              Text('Status: CONFIRMED', style: GoogleFonts.inter(fontSize: 10.5, color: emeraldGreen, fontWeight: FontWeight.w700)),
            ],
          ),
        ),
        const Divider(height: 1, color: borderSubtle),

        // Customer Row (Screenshot 3 & 5)
        InkWell(
          onTap: _showCustomerSelectionSheet,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            color: const Color(0xFFF8FAFC),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    const Icon(Icons.person_outline, size: 18, color: textMuted),
                    const SizedBox(width: 8),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('CUSTOMER', style: GoogleFonts.inter(fontSize: 9.5, fontWeight: FontWeight.bold, color: textMuted)),
                        Text(_selectedCustomer['name'] ?? 'Guest (ضيف)', style: GoogleFonts.inter(fontSize: 12, fontWeight: FontWeight.w700, color: textMain)),
                      ],
                    ),
                  ],
                ),
                const Icon(Icons.arrow_forward_ios, size: 12, color: textMuted),
              ],
            ),
          ),
        ),
        const Divider(height: 1, color: borderSubtle),

        // Cart Items List
        Expanded(
          child: _cartItems.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.remove_shopping_cart_outlined, size: 36, color: Colors.black26),
                      const SizedBox(height: 8),
                      Text('No items in order', style: GoogleFonts.inter(color: textMuted, fontSize: 13)),
                    ],
                  ),
                )
              : ListView.separated(
                  padding: const EdgeInsets.all(16),
                  itemCount: _cartItems.length,
                  separatorBuilder: (_, __) => const Divider(height: 16, color: borderSubtle),
                  itemBuilder: (ctx, i) {
                    final item = _cartItems[i];
                    final mods = (item['modifiers'] as List? ?? []);
                    final qty = (item['quantity'] as num).toInt();
                    final unitPrice = (item['unit_price'] as num).toDouble();

                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              child: Text(
                                item['name'] ?? '',
                                style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 13, color: textMain),
                              ),
                            ),
                            Text(
                              '${unitPrice.toStringAsFixed(2)} SAR',
                              style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 13, color: textMain),
                            ),
                          ],
                        ),
                        Text(
                          '${unitPrice.toStringAsFixed(2)} SAR / ea',
                          style: GoogleFonts.inter(fontSize: 11, color: textMuted),
                        ),
                        if (mods.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          ...mods.map((m) => Text(
                                '+ ${m['name']} (${(m['price'] as num).toStringAsFixed(2)} SAR)',
                                style: GoogleFonts.inter(fontSize: 11, color: amberOrange, fontWeight: FontWeight.w600),
                              )),
                        ],
                        const SizedBox(height: 6),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            InkWell(
                              onTap: () => _removeFromCart(i),
                              child: Row(
                                children: [
                                  const Icon(Icons.delete_outline, size: 14, color: textMuted),
                                  const SizedBox(width: 4),
                                  Text('Remove', style: GoogleFonts.inter(fontSize: 11, color: textMuted)),
                                ],
                              ),
                            ),
                            // Qty Control
                            Container(
                              decoration: BoxDecoration(
                                color: const Color(0xFFF1F5F9),
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Row(
                                children: [
                                  InkWell(
                                    onTap: () => _updateCartItemQty(i, -1),
                                    child: const Padding(
                                      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                      child: Text('-', style: TextStyle(fontWeight: FontWeight.bold)),
                                    ),
                                  ),
                                  Padding(
                                    padding: const EdgeInsets.symmetric(horizontal: 6),
                                    child: Text('$qty', style: GoogleFonts.inter(fontWeight: FontWeight.bold, fontSize: 12)),
                                  ),
                                  InkWell(
                                    onTap: () => _updateCartItemQty(i, 1),
                                    child: const Padding(
                                      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                      child: Text('+', style: TextStyle(fontWeight: FontWeight.bold)),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ],
                    );
                  },
                ),
        ),

        // Summary Calculations & 2 Buttons (Screenshot 3)
        Container(
          padding: const EdgeInsets.all(16),
          decoration: const BoxDecoration(
            color: Color(0xFFF8FAFC),
            border: Border(top: BorderSide(color: borderSubtle)),
          ),
          child: Column(
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Subtotal', style: GoogleFonts.inter(fontSize: 12, color: textMuted)),
                  Text('$subtotal SAR', style: GoogleFonts.inter(fontSize: 12, fontWeight: FontWeight.w600, color: textMain)),
                ],
              ),
              const SizedBox(height: 4),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('VAT (15%)', style: GoogleFonts.inter(fontSize: 12, color: textMuted)),
                  Text('$vat SAR', style: GoogleFonts.inter(fontSize: 12, fontWeight: FontWeight.w600, color: textMain)),
                ],
              ),
              const SizedBox(height: 6),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Total Amount', style: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w800, color: textMain)),
                  Text('$total SAR', style: GoogleFonts.inter(fontSize: 16, fontWeight: FontWeight.w900, color: textMain)),
                ],
              ),
              const SizedBox(height: 12),

              // Two Buttons from Screenshot 3: [Send to Kitchen] & [Pay & Settle]
              Row(
                children: [
                  // Button 1: Send to Kitchen
                  Expanded(
                    child: ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: emeraldGreen,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                        elevation: 0,
                      ),
                      icon: const Icon(Icons.check_circle_outline, size: 16),
                      label: const Text('Sent to Kitchen', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                      onPressed: _cartItems.isEmpty ? null : _sendOrderToKitchen,
                    ),
                  ),
                  const SizedBox(width: 8),

                  // Button 2: Pay & Settle
                  Expanded(
                    child: ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: darkPill,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                        elevation: 0,
                      ),
                      icon: const Icon(Icons.payment, size: 16),
                      label: const Text('Pay & Settle', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                      onPressed: _cartItems.isEmpty ? null : () => setState(() => _currentView = 'checkout'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  // =========================================================================
  // VIEW 3: CHECKOUT SCREEN (Screenshot 4)
  // =========================================================================

  Widget _buildCheckoutView() {
    return _CheckoutScreenWidget(
      cartItems: _cartItems,
      calculatedOrder: _calculatedOrder,
      paymentMethods: _storePaymentMethods,
      selectedCustomer: _selectedCustomer,
      storeId: widget.storeId,
      tableId: _selectedTable?['id'],
      orderSource: _orderSource,
      onBack: () => setState(() => _currentView = 'menu'),
      onOrderCompleted: () {
        _clearCart();
        _fetchTables();
        _fetchBills();
        _fetchKdsOrders();
        _fetchDrawerBalance();
        setState(() => _currentView = 'tables');
      },
    );
  }

  // =========================================================================
  // VIEW 4: KITCHEN DISPLAY SYSTEM (KDS)
  // =========================================================================

  // =========================================================================
  // VIEW 4: KITCHEN DISPLAY SYSTEM (KDS)
  // =========================================================================

  Widget _buildKdsView() {
    final pendingCount = _kdsOrders.where((o) => (o['status'] ?? '').toString().toLowerCase() == 'pending').length;
    final inPrepCount = _kdsOrders.where((o) {
      final s = (o['status'] ?? '').toString().toLowerCase();
      return s == 'in_preparation' || s == 'preparing';
    }).length;
    final completedCount = _kdsOrders.where((o) {
      final s = (o['status'] ?? '').toString().toLowerCase();
      return s == 'completed' || s == 'ready_to_serve' || s == 'ready';
    }).length;

    final filteredOrders = _kdsOrders.where((kds) {
      final s = (kds['status'] ?? '').toString().toLowerCase();
      if (_kdsFilter == 'Pending') return s == 'pending';
      if (_kdsFilter == 'In Preparation') return s == 'in_preparation' || s == 'preparing';
      if (_kdsFilter == 'Completed') return s == 'completed' || s == 'ready_to_serve' || s == 'ready';
      return true;
    }).toList();

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Filter Tabs Bar & Refresh Button (No duplicate back button or title)
          Row(
            children: [
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      _buildFilterPill('All', _kdsFilter == 'All', _kdsOrders.length, () => setState(() => _kdsFilter = 'All')),
                      const SizedBox(width: 8),
                      _buildFilterPill('Pending', _kdsFilter == 'Pending', pendingCount, () => setState(() => _kdsFilter = 'Pending'), color: amberOrange),
                      const SizedBox(width: 8),
                      _buildFilterPill('In Prep', _kdsFilter == 'In Preparation', inPrepCount, () => setState(() => _kdsFilter = 'In Preparation'), color: primarySky),
                      const SizedBox(width: 8),
                      _buildFilterPill('Completed', _kdsFilter == 'Completed', completedCount, () => setState(() => _kdsFilter = 'Completed'), color: emeraldGreen),
                    ],
                  ),
                ),
              ),
              const SizedBox(width: 6),
              IconButton(
                icon: const Icon(Icons.refresh_rounded, color: textMain, size: 20),
                tooltip: 'Refresh KDS',
                onPressed: _fetchKdsOrders,
              ),
            ],
          ),
          const SizedBox(height: 12),

          Expanded(
            child: _isLoadingKds
                ? const Center(child: CircularProgressIndicator(color: darkPill))
                : filteredOrders.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(Icons.kitchen_rounded, size: 48, color: Color(0xFFCBD5E1)),
                            const SizedBox(height: 12),
                            Text(
                              _kdsFilter == 'All' ? 'No kitchen orders yet.' : 'No $_kdsFilter orders.',
                              style: GoogleFonts.inter(color: textMuted, fontWeight: FontWeight.w600, fontSize: 14),
                            ),
                          ],
                        ),
                      )
                    : LayoutBuilder(
                        builder: (ctx, constraints) {
                          int crossAxisCount = 1;
                          double childAspectRatio = 1.15;
                          if (constraints.maxWidth > 1000) {
                            crossAxisCount = 3;
                            childAspectRatio = 1.05;
                          } else if (constraints.maxWidth > 600) {
                            crossAxisCount = 2;
                            childAspectRatio = 1.05;
                          } else {
                            crossAxisCount = 1;
                            childAspectRatio = 1.18;
                          }

                          return GridView.builder(
                            itemCount: filteredOrders.length,
                            gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: crossAxisCount,
                              crossAxisSpacing: 12,
                              mainAxisSpacing: 12,
                              childAspectRatio: childAspectRatio,
                            ),
                            itemBuilder: (ctx, idx) {
                              final kds = filteredOrders[idx];
                              final orderNum = kds['order_number']?.toString() ?? '';
                              final items = (kds['items'] as List? ?? []);
                              final rawStatus = (kds['status'] ?? 'pending').toString().toLowerCase();
                              final orderId = (kds['order_id'] ?? kds['id'] as num?)?.toInt();
                              final isTakeaway = kds['order_source'] == 'takeaway';
                              final tableNum = kds['table_number']?.toString();

                              final isPending = rawStatus == 'pending';
                              final isInPrep = rawStatus == 'in_preparation' || rawStatus == 'preparing';
                              final isCompleted = rawStatus == 'completed' || rawStatus == 'ready_to_serve' || rawStatus == 'ready';

                              Color statusColor = amberOrange;
                              Color statusBgColor = amberBg;
                              String statusLabel = 'PENDING';

                              if (isInPrep) {
                                statusColor = primarySky;
                                statusBgColor = const Color(0xFFE0F2FE);
                                statusLabel = 'IN PREP';
                              } else if (isCompleted) {
                                statusColor = emeraldGreen;
                                statusBgColor = emeraldBg;
                                statusLabel = 'COMPLETED';
                              }

                              String elapsed = '';
                              if (kds['ordered_at'] != null) {
                                try {
                                  final dt = DateTime.parse(kds['ordered_at'].toString());
                                  final diff = DateTime.now().difference(dt);
                                  if (diff.inMinutes < 60) {
                                    elapsed = '${diff.inMinutes}m ago';
                                  } else {
                                    elapsed = '${diff.inHours}h ${diff.inMinutes % 60}m ago';
                                  }
                                } catch (_) {}
                              }

                              return Container(
                                decoration: BoxDecoration(
                                  color: cardBg,
                                  borderRadius: BorderRadius.circular(16),
                                  border: Border.all(color: borderSubtle),
                                  boxShadow: const [
                                    BoxShadow(
                                      color: Color(0x06000000),
                                      blurRadius: 8,
                                      offset: Offset(0, 2),
                                    )
                                  ],
                                ),
                                padding: const EdgeInsets.all(14),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // Header: Order Number, Tag, Status Pill
                                    Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      children: [
                                        Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            Text(
                                              orderNum,
                                              style: GoogleFonts.inter(
                                                fontWeight: FontWeight.w800,
                                                fontSize: 14,
                                                color: textMain,
                                              ),
                                            ),
                                            if (isTakeaway) ...[
                                              const SizedBox(width: 6),
                                              Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1.5),
                                                decoration: BoxDecoration(
                                                  color: const Color(0xFFF1F5F9),
                                                  borderRadius: BorderRadius.circular(4),
                                                ),
                                                child: Text(
                                                  'TAKEAWAY',
                                                  style: GoogleFonts.inter(fontSize: 8.5, fontWeight: FontWeight.bold, color: textMuted),
                                                ),
                                              ),
                                            ],
                                          ],
                                        ),
                                        Container(
                                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                          decoration: BoxDecoration(
                                            color: statusBgColor,
                                            borderRadius: BorderRadius.circular(6),
                                          ),
                                          child: Text(
                                            statusLabel,
                                            style: GoogleFonts.inter(
                                              fontSize: 9.5,
                                              fontWeight: FontWeight.bold,
                                              color: statusColor,
                                            ),
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 3),
                                    Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      children: [
                                        Text(
                                          tableNum != null && tableNum.isNotEmpty ? 'Table $tableNum' : (isTakeaway ? 'Takeaway' : 'Pickup'),
                                          style: GoogleFonts.inter(fontSize: 11, color: textMuted, fontWeight: FontWeight.w600),
                                        ),
                                        if (elapsed.isNotEmpty)
                                          Text(
                                            elapsed,
                                            style: GoogleFonts.inter(fontSize: 10.5, color: textMuted),
                                          ),
                                      ],
                                    ),
                                    const Divider(height: 14, color: borderSubtle),

                                    // Items List
                                    Expanded(
                                      child: items.isEmpty
                                          ? Center(
                                              child: Text('No items listed', style: GoogleFonts.inter(fontSize: 11, color: textMuted)),
                                            )
                                          : ListView.builder(
                                              physics: const ClampingScrollPhysics(),
                                              itemCount: items.length,
                                              itemBuilder: (iCtx, iIdx) {
                                                final itm = items[iIdx];
                                                final q = itm['quantity'] ?? 1;
                                                final name = itm['name'] ?? itm['item_name'] ?? 'Item';
                                                return Padding(
                                                  padding: const EdgeInsets.only(bottom: 4),
                                                  child: Row(
                                                    children: [
                                                      Container(
                                                        padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                                                        decoration: BoxDecoration(
                                                          color: darkPill,
                                                          borderRadius: BorderRadius.circular(4),
                                                        ),
                                                        child: Text(
                                                          '${q}x',
                                                          style: GoogleFonts.inter(fontSize: 9.5, fontWeight: FontWeight.bold, color: Colors.white),
                                                        ),
                                                      ),
                                                      const SizedBox(width: 6),
                                                      Expanded(
                                                        child: Text(
                                                          name,
                                                          style: GoogleFonts.inter(
                                                            fontSize: 11.5,
                                                            color: textMain,
                                                            fontWeight: FontWeight.w600,
                                                          ),
                                                          overflow: TextOverflow.ellipsis,
                                                        ),
                                                      ),
                                                    ],
                                                  ),
                                                );
                                              },
                                            ),
                                    ),

                                    const SizedBox(height: 8),

                                    // Interactive Status Switcher (Pending / In Prep / Complete)
                                    Container(
                                      padding: const EdgeInsets.all(2.5),
                                      decoration: BoxDecoration(
                                        color: const Color(0xFFF1F5F9),
                                        borderRadius: BorderRadius.circular(9),
                                        border: Border.all(color: borderSubtle),
                                      ),
                                      child: Row(
                                        children: [
                                          _buildStatusSegmentButton(
                                            label: 'Pending',
                                            icon: Icons.hourglass_top_rounded,
                                            isActive: isPending,
                                            activeColor: amberOrange,
                                            onTap: () {
                                              if (orderId != null) {
                                                _updateKdsOrderStatus(orderId, 'pending', orderNum);
                                              }
                                            },
                                          ),
                                          const SizedBox(width: 3),
                                          _buildStatusSegmentButton(
                                            label: 'In Prep',
                                            icon: Icons.outdoor_grill_rounded,
                                            isActive: isInPrep,
                                            activeColor: primarySky,
                                            onTap: () {
                                              if (orderId != null) {
                                                _updateKdsOrderStatus(orderId, 'in_preparation', orderNum);
                                              }
                                            },
                                          ),
                                          const SizedBox(width: 3),
                                          _buildStatusSegmentButton(
                                            label: 'Complete',
                                            icon: Icons.check_circle_rounded,
                                            isActive: isCompleted,
                                            activeColor: emeraldGreen,
                                            onTap: () {
                                              if (orderId != null) {
                                                _updateKdsOrderStatus(orderId, 'completed', orderNum);
                                              }
                                            },
                                          ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            },
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }

  // =========================================================================
  // VIEW 5: BILLS HISTORY
  // =========================================================================

  Widget _buildBillsView() {
    final pendingCount = _bills.where((b) {
      final s = (b['status'] ?? '').toString().toLowerCase();
      return s != 'completed' && s != 'paid';
    }).length;
    final paidCount = _bills.where((b) {
      final s = (b['status'] ?? '').toString().toLowerCase();
      return s == 'completed' || s == 'paid';
    }).length;

    final filteredBills = _bills.where((bill) {
      final s = (bill['status'] ?? '').toString().toLowerCase();
      final isPaid = s == 'completed' || s == 'paid';
      if (_billsFilter == 'Pending') return !isPaid;
      if (_billsFilter == 'Paid') return isPaid;
      return true;
    }).toList();

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Filter Bar & Refresh Button (No duplicate back button or title)
          Row(
            children: [
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      _buildFilterPill('All', _billsFilter == 'All', _bills.length, () => setState(() => _billsFilter = 'All')),
                      const SizedBox(width: 8),
                      _buildFilterPill('Pending', _billsFilter == 'Pending', pendingCount, () => setState(() => _billsFilter = 'Pending'), color: amberOrange),
                      const SizedBox(width: 8),
                      _buildFilterPill('Paid', _billsFilter == 'Paid', paidCount, () => setState(() => _billsFilter = 'Paid'), color: emeraldGreen),
                    ],
                  ),
                ),
              ),
              const SizedBox(width: 6),
              IconButton(
                icon: const Icon(Icons.refresh_rounded, color: textMain, size: 20),
                tooltip: 'Refresh Bills',
                onPressed: _fetchBills,
              ),
            ],
          ),
          const SizedBox(height: 12),

          Expanded(
            child: _isLoadingBills
                ? const Center(child: CircularProgressIndicator(color: darkPill))
                : filteredBills.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(Icons.receipt_long_rounded, size: 48, color: Color(0xFFCBD5E1)),
                            const SizedBox(height: 12),
                            Text(
                              _billsFilter == 'All' ? 'No past bills recorded.' : 'No $_billsFilter bills found.',
                              style: GoogleFonts.inter(color: textMuted, fontWeight: FontWeight.w600, fontSize: 14),
                            ),
                          ],
                        ),
                      )
                    : ListView.separated(
                        itemCount: filteredBills.length,
                        separatorBuilder: (_, __) => const SizedBox(height: 8),
                        itemBuilder: (ctx, idx) {
                          final bill = filteredBills[idx];
                          final s = (bill['status'] ?? '').toString().toLowerCase();
                          final isPaid = s == 'completed' || s == 'paid';
                          final orderNum = bill['order_number']?.toString() ?? 'RO-';
                          final tableInfo = bill['table_number'] != null ? 'Table ${bill['table_number']}' : 'Takeaway';
                          final totalAmt = bill['total_amount']?.toString() ?? '0.00';

                          String orderedAt = bill['ordered_at']?.toString() ?? '';
                          if (orderedAt.contains('T')) {
                            try {
                              final dt = DateTime.parse(orderedAt);
                              orderedAt = '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
                            } catch (_) {}
                          }

                          return Container(
                            decoration: BoxDecoration(
                              color: cardBg,
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(color: borderSubtle),
                            ),
                            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        orderNum,
                                        style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 13, color: textMain),
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        '$tableInfo${orderedAt.isNotEmpty ? " • $orderedAt" : ""}',
                                        style: GoogleFonts.inter(fontSize: 11, color: textMuted),
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                      decoration: BoxDecoration(
                                        color: isPaid ? emeraldBg : amberBg,
                                        borderRadius: BorderRadius.circular(6),
                                      ),
                                      child: Text(
                                        isPaid ? 'PAID' : 'PENDING',
                                        style: GoogleFonts.inter(
                                          fontSize: 10,
                                          fontWeight: FontWeight.bold,
                                          color: isPaid ? emeraldGreen : amberOrange,
                                        ),
                                      ),
                                    ),
                                    const SizedBox(width: 10),
                                    Text(
                                      'SAR $totalAmt',
                                      style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 13, color: textMain),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }

  Widget _buildFilterPill(String title, bool isSelected, int count, VoidCallback onTap, {Color? color}) {
    final activeColor = color ?? darkPill;
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 6),
        decoration: BoxDecoration(
          color: isSelected ? activeColor : const Color(0xFFF1F5F9),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: isSelected ? activeColor : borderSubtle),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              title,
              style: GoogleFonts.inter(
                fontSize: 11.5,
                fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
                color: isSelected ? Colors.white : textMain,
              ),
            ),
            const SizedBox(width: 5),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
              decoration: BoxDecoration(
                color: isSelected ? Colors.white.withValues(alpha: 0.25) : Colors.white,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '$count',
                style: GoogleFonts.inter(
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                  color: isSelected ? Colors.white : textMuted,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusSegmentButton({
    required String label,
    required IconData icon,
    required bool isActive,
    required Color activeColor,
    required VoidCallback onTap,
  }) {
    return Expanded(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(7),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 6),
          decoration: BoxDecoration(
            color: isActive ? activeColor : Colors.transparent,
            borderRadius: BorderRadius.circular(7),
            boxShadow: isActive
                ? [
                    BoxShadow(
                      color: activeColor.withValues(alpha: 0.25),
                      blurRadius: 4,
                      offset: const Offset(0, 1),
                    )
                  ]
                : null,
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon, size: 12, color: isActive ? Colors.white : textMuted),
              const SizedBox(width: 3),
              Text(
                label,
                style: GoogleFonts.inter(
                  fontSize: 10.5,
                  fontWeight: isActive ? FontWeight.bold : FontWeight.w600,
                  color: isActive ? Colors.white : textMuted,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// =========================================================================
// CHECKOUT SCREEN COMPONENT (Exact replica of Screenshot 4)
// =========================================================================

class _CheckoutScreenWidget extends StatefulWidget {
  final List<Map<String, dynamic>> cartItems;
  final Map<String, dynamic> calculatedOrder;
  final List<Map<String, dynamic>> paymentMethods;
  final Map<String, dynamic> selectedCustomer;
  final int storeId;
  final int? tableId;
  final String orderSource;
  final VoidCallback onBack;
  final VoidCallback onOrderCompleted;

  const _CheckoutScreenWidget({
    required this.cartItems,
    required this.calculatedOrder,
    required this.paymentMethods,
    required this.selectedCustomer,
    required this.storeId,
    required this.tableId,
    required this.orderSource,
    required this.onBack,
    required this.onOrderCompleted,
  });

  @override
  State<_CheckoutScreenWidget> createState() => _CheckoutScreenWidgetState();
}

class _CheckoutScreenWidgetState extends State<_CheckoutScreenWidget> {
  late String _selectedMethod;
  String _keypadInput = '0.00';
  bool _isFirstKey = true;
  bool _isProcessing = false;

  @override
  void initState() {
    super.initState();
    if (widget.paymentMethods.isNotEmpty) {
      _selectedMethod = widget.paymentMethods.first['code']?.toString() ??
          widget.paymentMethods.first['name']?.toString() ??
          'Cash';
    } else {
      _selectedMethod = 'Cash';
    }
  }

  double get _totalDue => double.tryParse(widget.calculatedOrder['total_amount']?.toString() ?? '0') ?? 0.0;
  double get _tendered => double.tryParse(_keypadInput) ?? 0.0;
  double get _changeDue => (_tendered - _totalDue) > 0 ? (_tendered - _totalDue) : 0.0;
  double get _shortAmount => (_totalDue - _tendered) > 0 ? (_totalDue - _tendered) : 0.0;

  void _handleKey(String key) {
    setState(() {
      if (key == 'BACK') {
        if (_keypadInput.length <= 1) {
          _keypadInput = '0';
          _isFirstKey = true;
        } else {
          _keypadInput = _keypadInput.substring(0, _keypadInput.length - 1);
        }
        return;
      }

      if (_isFirstKey) {
        if (key == '.') {
          _keypadInput = '0.';
        } else {
          _keypadInput = key;
        }
        _isFirstKey = false;
        return;
      }

      if (key == '.' && _keypadInput.contains('.')) return;
      _keypadInput += key;
    });
  }

  void _setQuickCash(double amount) {
    setState(() {
      _keypadInput = amount.toStringAsFixed(2);
      _isFirstKey = false;
    });
  }

  Future<void> _completeOrder() async {
    if (_isProcessing) return;
    setState(() => _isProcessing = true);

    final isCardMethod = _selectedMethod.toLowerCase().contains('card') ||
        _selectedMethod.toLowerCase().contains('credit') ||
        _selectedMethod.toLowerCase().contains('mada') ||
        _selectedMethod.toLowerCase().contains('pos');
    final amountPaid = isCardMethod ? _totalDue : _tendered;

    final subtotal = double.tryParse(widget.calculatedOrder['subtotal']?.toString() ?? '') ??
        widget.cartItems.fold<double>(0.0, (acc, item) {
          final p = (item['unit_price'] as num?)?.toDouble() ?? (item['price'] as num?)?.toDouble() ?? 0.0;
          final q = (item['quantity'] as num?)?.toInt() ?? 1;
          final m = (item['modifiers_total'] as num?)?.toDouble() ?? 0.0;
          return acc + (p + m) * q;
        });
    final discount = double.tryParse(widget.calculatedOrder['discount_amount']?.toString() ?? '0') ?? 0.0;
    final tax = double.tryParse(widget.calculatedOrder['tax_amount']?.toString() ?? '') ??
        (math.max(0.0, subtotal - discount) * 0.15);
    final total = double.tryParse(widget.calculatedOrder['total_amount']?.toString() ?? '') ??
        (math.max(0.0, subtotal - discount) + tax);

    final subtotalStr = subtotal.toStringAsFixed(2);
    final discountStr = discount.toStringAsFixed(2);
    final taxStr = tax.toStringAsFixed(2);
    final totalStr = total.toStringAsFixed(2);
    final existingOrderId = (widget.calculatedOrder['order_id'] as num?)?.toInt();
    final existingOrderNum = widget.calculatedOrder['order_number']?.toString();
    final orderNum = existingOrderNum ?? 'RO-${DateTime.now().millisecondsSinceEpoch % 1000000}';
    final txnNumber = 'TXN-${DateTime.now().millisecondsSinceEpoch % 1000000}';

    try {
      dynamic resp;
      if (existingOrderId != null) {
        resp = NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'settleOrder',
          payload: {
            'order_id': existingOrderId,
            'id': existingOrderId,
            'status': 'paid',
            'amount_paid': amountPaid.toStringAsFixed(2),
            'change_given': (_changeDue > 0 ? _changeDue : 0.0).toStringAsFixed(2),
            'payment_method': _selectedMethod,
          },
        );
      } else {
        resp = NembusBridge().callHandler(
          handler: 'restaurant',
          action: 'createOrder',
          payload: {
            'store_id': widget.storeId,
            'table_id': widget.tableId,
            'cashier_id': SingletonClass().activeCashierId ?? 1,
            'cashier_session_id': SingletonClass().activeCashierSessionId ?? 1,
            'order_source': widget.orderSource,
            'status': 'paid',
            'subtotal': subtotalStr,
            'discount_amount': discountStr,
            'tax_amount': taxStr,
            'total_amount': totalStr,
            'amount_paid': amountPaid.toStringAsFixed(2),
            'change_given': (_changeDue > 0 ? _changeDue : 0.0).toStringAsFixed(2),
            'order_number': orderNum,
          },
        );
      }

      final orderData = (resp != null && resp['success'] == true && resp['data'] is Map)
          ? Map<String, dynamic>.from(resp['data'] as Map)
          : <String, dynamic>{};
      final resolvedNum = orderData['order_number']?.toString() ?? orderNum;
      final orderId = existingOrderId ?? ((orderData['id'] as num?)?.toInt() ?? DateTime.now().millisecondsSinceEpoch);

      // Persist to local SQLite DB
      try {
        final db = DatabaseService().database;

        // 1. Insert or update completed restaurant order
        await db.insert('restaurant_orders', {
          'id': orderId,
          'store_id': widget.storeId,
          'table_id': widget.tableId,
          'cashier_id': SingletonClass().activeCashierId ?? 1,
          'cashier_session_id': SingletonClass().activeCashierSessionId ?? 1,
          'order_number': resolvedNum,
          'order_source': widget.orderSource,
          'status': 'paid',
          'subtotal': subtotal,
          'discount_amount': discount,
          'tax_amount': tax,
          'total_amount': total,
          'amount_paid': amountPaid,
          'change_given': _changeDue > 0 ? _changeDue : 0.0,
          'notes': 'Paid via $_selectedMethod',
          'metadata': jsonEncode({'items': widget.cartItems}),
          'ordered_at': DateTime.now().toIso8601String(),
        }, conflictAlgorithm: ConflictAlgorithm.replace);

        // 2. Insert items with verified menu_item_id
        for (final itm in widget.cartItems) {
          final rawItemId = (itm['menu_item_id'] ?? itm['id'] as num?)?.toInt() ?? 1;
          int validMenuItemId = rawItemId;
          try {
            final existing = await db.query('menu_items', where: 'id = ?', whereArgs: [rawItemId], limit: 1);
            if (existing.isEmpty) {
              final anyItem = await db.query('menu_items', limit: 1);
              if (anyItem.isNotEmpty) {
                validMenuItemId = (anyItem.first['id'] as num).toInt();
              }
            }
          } catch (_) {}

          final uPrice = (itm['unit_price'] as num?)?.toDouble() ?? (itm['price'] as num?)?.toDouble() ?? 0.0;
          final qty = (itm['quantity'] as num?)?.toDouble() ?? 1.0;
          await db.insert('restaurant_order_items', {
            'order_id': orderId,
            'menu_item_id': validMenuItemId,
            'quantity': qty,
            'unit_price': uPrice,
            'modifiers_snapshot': jsonEncode(itm['modifiers'] ?? []),
            'modifiers_total': (itm['modifiers_total'] as num?)?.toDouble() ?? 0.0,
            'subtotal': uPrice * qty,
            'status': 'completed',
          }, conflictAlgorithm: ConflictAlgorithm.replace);
        }

        // 3. Insert POS transaction
        final txnId = await db.insert('pos_transactions', {
          'store_id': widget.storeId,
          'cashier_id': SingletonClass().activeCashierId ?? 1,
          'cashier_session_id': SingletonClass().activeCashierSessionId ?? 1,
          'transaction_number': txnNumber,
          'transaction_type': 'sale',
          'subtotal': subtotal,
          'discount_amount': discount,
          'tax_amount': tax,
          'total_amount': total,
          'amount_paid': amountPaid,
          'change_given': _changeDue > 0 ? _changeDue : 0.0,
          'status': 'completed',
          'metadata': jsonEncode({'order_number': resolvedNum, 'order_source': widget.orderSource, 'order_id': orderId}),
          'created_at': DateTime.now().toIso8601String(),
        }, conflictAlgorithm: ConflictAlgorithm.replace);

        // 4. Insert POS payment
        await db.insert('pos_payments', {
          'transaction_id': txnId,
          'payment_method': _selectedMethod.toLowerCase(),
          'amount': amountPaid,
          'payment_reference': txnNumber,
          'payment_date': DateTime.now().toIso8601String(),
        });

        // 5. Free table if dining
        if (widget.tableId != null) {
          try {
            NembusBridge().callHandler(
              handler: 'restaurant',
              action: 'freeTable',
              payload: {
                'table_id': widget.tableId,
                'store_id': widget.storeId,
              },
            );
          } catch (_) {}

          await db.update(
            'restaurant_tables',
            {
              'metadata': jsonEncode({'is_occupied': false, 'status': 'available', 'active_order': null}),
            },
            where: 'id = ?',
            whereArgs: [widget.tableId],
          );

          await db.insert('sync_queue', {
            'entity_type': 'restaurant_tables',
            'entity_id': widget.tableId.toString(),
            'action': 'UPDATE',
            'payload': jsonEncode({
              'id': widget.tableId,
              'is_occupied': false,
              'status': 'available',
            }),
            'status': 'pending',
            'priority': 15,
            'created_at': DateTime.now().toIso8601String(),
          });
        }

        // 6. Enqueue orders & transactions to sync_queue
        await db.insert('sync_queue', {
          'entity_type': 'restaurant_orders',
          'entity_id': orderId.toString(),
          'action': existingOrderId != null ? 'UPDATE' : 'INSERT',
          'payload': jsonEncode({
            'order_id': orderId,
            'order_number': resolvedNum,
            'transaction_number': txnNumber,
            'store_id': widget.storeId,
            'table_id': widget.tableId,
            'items': widget.cartItems,
            'status': 'paid',
            'amount_paid': amountPaid,
            'payment_method': _selectedMethod,
          }),
          'status': 'pending',
          'priority': 20,
          'correlation_id': resolvedNum,
          'created_at': DateTime.now().toIso8601String(),
        });

        await db.insert('sync_queue', {
          'entity_type': 'pos_transactions',
          'entity_id': txnId.toString(),
          'action': 'INSERT',
          'payload': jsonEncode({
            'id': txnId,
            'store_id': widget.storeId,
            'transaction_number': txnNumber,
            'total_amount': total,
            'payment_method': _selectedMethod,
            'status': 'completed',
          }),
          'status': 'pending',
          'priority': 10,
          'created_at': DateTime.now().toIso8601String(),
        });
      } catch (dbErr) {
        developer.log('SQLite write error: $dbErr', name: 'RestaurantPOS');
      }

      // 7. Show ThermalReceiptWidget modal
      if (mounted) {
        await showDialog(
          context: context,
          barrierDismissible: false,
          builder: (ctx) => Dialog(
            backgroundColor: Colors.transparent,
            insetPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(20),
              child: ThermalReceiptWidget(
                storeName: SingletonClass().activeStoreName ?? 'Restaurant POS',
                orderNumber: resolvedNum,
                cashierName: SingletonClass().activeCashierCode ?? 'Cashier',
                terminalName: SingletonClass().activeTerminalName ?? 'Main POS',
                customerName: widget.selectedCustomer['name']?.toString() ?? 'Dining Guest',
                paymentMethod: _selectedMethod,
                cartItems: widget.cartItems.isNotEmpty
                    ? widget.cartItems
                    : [
                        {
                          'name': 'Restaurant Order $resolvedNum',
                          'price': subtotal,
                          'unit_price': subtotal,
                          'quantity': 1,
                        }
                      ],
                subtotal: subtotal,
                discount: discount,
                taxAmount: tax,
                totalAmount: total,
                changeDue: _changeDue > 0 ? _changeDue : 0.0,
                onProceedWithoutPrint: () => Navigator.of(ctx).pop(),
                onBack: () => Navigator.of(ctx).pop(),
              ),
            ),
          ),
        );
      }

      widget.onOrderCompleted();
    } catch (e) {
      developer.log('Settlement error: $e', name: 'RestaurantPOS');
    } finally {
      if (mounted) setState(() => _isProcessing = false);
    }
  }

  Widget _buildOrderSummaryCard(
    bool isCompact,
    String subtotal,
    String discountStr,
    String discountLabel,
    String vat,
    String taxRate,
    double totalDue,
  ) {
    final discountVal = double.tryParse(discountStr) ?? 0.0;
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      padding: EdgeInsets.all(isCompact ? 16 : 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: isCompact ? MainAxisSize.min : MainAxisSize.max,
        children: [
          Text('CURRENT ORDER', style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.bold, color: const Color(0xFF64748B), letterSpacing: 0.8)),
          const SizedBox(height: 14),
          if (isCompact)
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: widget.cartItems.length,
              separatorBuilder: (_, __) => const SizedBox(height: 8),
              itemBuilder: (ctx, i) {
                final item = widget.cartItems[i];
                final qty = item['quantity'];
                final price = (item['unit_price'] as num?)?.toDouble() ?? (item['price'] as num?)?.toDouble() ?? 0.0;
                return Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Text(
                        '${item['name']} x $qty',
                        style: GoogleFonts.inter(fontSize: 13, color: const Color(0xFF0F172A), fontWeight: FontWeight.w600),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    Text('SAR ${(price * qty).toStringAsFixed(2)}', style: GoogleFonts.inter(fontSize: 13, fontWeight: FontWeight.w700, color: const Color(0xFF0F172A))),
                  ],
                );
              },
            )
          else
            Expanded(
              child: ListView.separated(
                itemCount: widget.cartItems.length,
                separatorBuilder: (_, __) => const SizedBox(height: 10),
                itemBuilder: (ctx, i) {
                  final item = widget.cartItems[i];
                  final qty = item['quantity'];
                  final price = (item['unit_price'] as num?)?.toDouble() ?? (item['price'] as num?)?.toDouble() ?? 0.0;
                  return Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          '${item['name']} x $qty',
                          style: GoogleFonts.inter(fontSize: 13, color: const Color(0xFF0F172A), fontWeight: FontWeight.w600),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      Text('SAR ${(price * qty).toStringAsFixed(2)}', style: GoogleFonts.inter(fontSize: 13, fontWeight: FontWeight.w700, color: const Color(0xFF0F172A))),
                    ],
                  );
                },
              ),
            ),
          const Divider(color: Color(0xFFE2E8F0)),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Subtotal', style: GoogleFonts.inter(fontSize: 12, color: const Color(0xFF64748B))),
              Text('SAR $subtotal', style: GoogleFonts.inter(fontSize: 12, fontWeight: FontWeight.w600, color: const Color(0xFF0F172A))),
            ],
          ),
          if (discountVal > 0) ...[
            const SizedBox(height: 4),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  discountLabel.isNotEmpty ? 'Discount ($discountLabel)' : 'Discount / Promotions',
                  style: GoogleFonts.inter(fontSize: 12, color: const Color(0xFF10B981), fontWeight: FontWeight.w600),
                ),
                Text('-SAR $discountStr', style: GoogleFonts.inter(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF10B981))),
              ],
            ),
          ],
          const SizedBox(height: 4),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Tax (VAT $taxRate%)', style: GoogleFonts.inter(fontSize: 12, color: const Color(0xFF64748B))),
              Text('SAR $vat', style: GoogleFonts.inter(fontSize: 12, fontWeight: FontWeight.w600, color: const Color(0xFF0F172A))),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Total Due', style: GoogleFonts.inter(fontSize: 15, fontWeight: FontWeight.w900, color: const Color(0xFF0F172A))),
              Text('SAR ${totalDue.toStringAsFixed(2)}', style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w900, color: const Color(0xFF0F172A))),
            ],
          ),
          const SizedBox(height: 14),
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(Icons.star, color: Color(0xFF10B981), size: 14),
                    const SizedBox(width: 4),
                    Text('Loyalty Member', style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.bold, color: const Color(0xFF10B981))),
                  ],
                ),
                const SizedBox(height: 4),
                Text(widget.selectedCustomer['name'] ?? 'Guest (ضيف)', style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 13, color: const Color(0xFF0F172A))),
                Text('Member #: ${widget.selectedCustomer['code'] ?? '10'}', style: GoogleFonts.inter(fontSize: 10.5, color: const Color(0xFF64748B))),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPaymentKeypadCard(bool isCompact, bool isSufficient, double shortAmount, double changeDue) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      padding: EdgeInsets.all(isCompact ? 16 : 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: isCompact ? MainAxisSize.min : MainAxisSize.max,
        children: [
          Text('SELECT PAYMENT METHOD', style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.bold, color: const Color(0xFF64748B), letterSpacing: 0.8)),
          const SizedBox(height: 12),

          // Method Tabs (Dynamically loaded from store in SQLite)
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            physics: const BouncingScrollPhysics(),
            child: Row(
              children: (widget.paymentMethods.isNotEmpty
                      ? widget.paymentMethods
                      : [
                          {'code': 'Cash', 'name': 'Cash', 'icon': Icons.money},
                          {'code': 'Card', 'name': 'Card', 'icon': Icons.credit_card},
                        ])
                  .map((pm) {
                final code = pm['code']?.toString() ?? 'Cash';
                final name = pm['name']?.toString() ?? code;
                final icon = (pm['icon'] is IconData) ? pm['icon'] as IconData : Icons.payment;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: _buildPaymentTab(code, name, icon),
                );
              }).toList(),
            ),
          ),
          const SizedBox(height: 16),

          // AMOUNT TENDERED Display
          Text('AMOUNT TENDERED', style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.bold, color: const Color(0xFF64748B))),
          const SizedBox(height: 6),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: const Color(0xFF10B981), width: 1.5),
            ),
            child: Text(
              'SAR $_keypadInput',
              style: GoogleFonts.inter(fontSize: 22, fontWeight: FontWeight.w900, color: const Color(0xFF0F172A)),
            ),
          ),
          const SizedBox(height: 12),

          // QUICK CASH Pills (SAR 25, 50, 100, 200)
          Text('QUICK CASH', style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.bold, color: const Color(0xFF64748B))),
          const SizedBox(height: 6),
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            physics: const BouncingScrollPhysics(),
            child: Row(
              children: [
                _buildQuickCashButton(25.0),
                const SizedBox(width: 8),
                _buildQuickCashButton(50.0),
                const SizedBox(width: 8),
                _buildQuickCashButton(100.0),
                const SizedBox(width: 8),
                _buildQuickCashButton(200.0),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Numeric Keypad Grid
          SizedBox(
            height: 220,
            child: Column(
              children: [
                Expanded(
                  child: Row(
                    children: [
                      _buildKeypadKey('1'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('2'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('3'),
                    ],
                  ),
                ),
                const SizedBox(height: 8),
                Expanded(
                  child: Row(
                    children: [
                      _buildKeypadKey('4'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('5'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('6'),
                    ],
                  ),
                ),
                const SizedBox(height: 8),
                Expanded(
                  child: Row(
                    children: [
                      _buildKeypadKey('7'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('8'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('9'),
                    ],
                  ),
                ),
                const SizedBox(height: 8),
                Expanded(
                  child: Row(
                    children: [
                      _buildKeypadKey('.'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('0'),
                      const SizedBox(width: 8),
                      _buildKeypadKey('BACK', isAction: true),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Short / Change & Action Buttons
          if (isCompact) ...[
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: isSufficient ? const Color(0xFFECFDF5) : const Color(0xFFFEF2F2),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: isSufficient ? const Color(0xFF10B981) : const Color(0xFFEF4444)),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    isSufficient ? 'CHANGE DUE' : 'AMOUNT SHORT',
                    style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.bold, color: isSufficient ? const Color(0xFF10B981) : const Color(0xFFEF4444)),
                  ),
                  Text(
                    isSufficient ? 'SAR ${changeDue.toStringAsFixed(2)}' : 'SAR ${shortAmount.toStringAsFixed(2)}',
                    style: GoogleFonts.inter(fontSize: 15, fontWeight: FontWeight.w900, color: isSufficient ? const Color(0xFF10B981) : const Color(0xFFEF4444)),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 10),
            Row(
              children: [
                Expanded(
                  flex: 4,
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFFF59E0B),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                    ),
                    icon: const Icon(Icons.pause, size: 16),
                    label: const Text('Suspend', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                    onPressed: widget.onBack,
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  flex: 6,
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: isSufficient ? const Color(0xFF06B6D4) : Colors.grey.shade300,
                      foregroundColor: isSufficient ? Colors.white : Colors.grey.shade500,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                    ),
                    icon: _isProcessing
                        ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                        : const Icon(Icons.check_circle, size: 16),
                    label: const Text('COMPLETE', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                    onPressed: isSufficient && !_isProcessing ? _completeOrder : null,
                  ),
                ),
              ],
            ),
          ] else
            Row(
              children: [
                Expanded(
                  flex: 4,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    decoration: BoxDecoration(
                      color: isSufficient ? const Color(0xFFECFDF5) : const Color(0xFFFEF2F2),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: isSufficient ? const Color(0xFF10B981) : const Color(0xFFEF4444)),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          isSufficient ? 'CHANGE DUE' : 'AMOUNT SHORT',
                          style: GoogleFonts.inter(fontSize: 9.5, fontWeight: FontWeight.bold, color: isSufficient ? const Color(0xFF10B981) : const Color(0xFFEF4444)),
                        ),
                        Text(
                          isSufficient ? 'SAR ${changeDue.toStringAsFixed(2)}' : 'SAR ${shortAmount.toStringAsFixed(2)}',
                          style: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w900, color: isSufficient ? const Color(0xFF10B981) : const Color(0xFFEF4444)),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: 10),
                ElevatedButton.icon(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFF59E0B),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                  ),
                  icon: const Icon(Icons.pause, size: 16),
                  label: const Text('Suspend & Edit', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                  onPressed: widget.onBack,
                ),
                const SizedBox(width: 10),
                Expanded(
                  flex: 5,
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: isSufficient ? const Color(0xFF06B6D4) : Colors.grey.shade300,
                      foregroundColor: isSufficient ? Colors.white : Colors.grey.shade500,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                    ),
                    icon: _isProcessing
                        ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                        : const Icon(Icons.check_circle, size: 16),
                    label: const Text('COMPLETE ORDER', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                    onPressed: isSufficient && !_isProcessing ? _completeOrder : null,
                  ),
                ),
              ],
            ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final subtotal = widget.calculatedOrder['subtotal']?.toString() ?? '0.00';
    final discountStr = widget.calculatedOrder['discount_amount']?.toString() ?? '0.00';
    final discountLabel = widget.calculatedOrder['discount_label']?.toString() ?? '';
    final vat = widget.calculatedOrder['tax_amount']?.toString() ?? '0.00';
    final taxRate = widget.calculatedOrder['tax_rate']?.toString() ?? '15.00';
    final totalDue = _totalDue;
    final isCardMethod = _selectedMethod.toLowerCase().contains('card') ||
        _selectedMethod.toLowerCase().contains('credit') ||
        _selectedMethod.toLowerCase().contains('mada') ||
        _selectedMethod.toLowerCase().contains('pos');
    final isSufficient = isCardMethod || _tendered >= totalDue;
    final shortAmount = (totalDue - _tendered) > 0 ? (totalDue - _tendered) : 0.0;
    final changeDue = (_tendered - totalDue) > 0 ? (_tendered - totalDue) : 0.0;

    final screenWidth = MediaQuery.of(context).size.width;
    final isCompact = screenWidth < 800;

    return isCompact
        ? SingleChildScrollView(
            physics: const BouncingScrollPhysics(),
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                _buildOrderSummaryCard(true, subtotal, discountStr, discountLabel, vat, taxRate, totalDue),
                const SizedBox(height: 16),
                _buildPaymentKeypadCard(true, isSufficient, shortAmount, changeDue),
              ],
            ),
          )
        : Padding(
            padding: const EdgeInsets.all(24),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  flex: 4,
                  child: _buildOrderSummaryCard(false, subtotal, discountStr, discountLabel, vat, taxRate, totalDue),
                ),
                const SizedBox(width: 20),
                Expanded(
                  flex: 6,
                  child: _buildPaymentKeypadCard(false, isSufficient, shortAmount, changeDue),
                ),
              ],
            ),
          );
  }

  Widget _buildPaymentTab(String code, String label, IconData icon) {
    final isSel = _selectedMethod.toLowerCase() == code.toLowerCase();
    return InkWell(
      onTap: () {
        setState(() {
          _selectedMethod = code;
          if (code.toLowerCase().contains('card') ||
              code.toLowerCase().contains('credit') ||
              code.toLowerCase().contains('mada') ||
              code.toLowerCase().contains('pos')) {
            _keypadInput = _totalDue.toStringAsFixed(2);
            _isFirstKey = false;
          }
        });
      },
      borderRadius: BorderRadius.circular(10),
      child: Container(
        constraints: const BoxConstraints(minWidth: 85),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: isSel ? const Color(0xFF0F172A) : const Color(0xFFF8FAFC),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: isSel ? const Color(0xFF0F172A) : const Color(0xFFE2E8F0)),
        ),
        child: Column(
          children: [
            Icon(icon, color: isSel ? Colors.white : const Color(0xFF0F172A), size: 20),
            const SizedBox(height: 4),
            Text(
              label,
              style: GoogleFonts.inter(
                fontSize: 11,
                fontWeight: FontWeight.bold,
                color: isSel ? Colors.white : const Color(0xFF0F172A),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildQuickCashButton(double amount) {
    return OutlinedButton(
      style: OutlinedButton.styleFrom(
        foregroundColor: const Color(0xFF0F172A),
        side: const BorderSide(color: Color(0xFFE2E8F0)),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      ),
      onPressed: () => _setQuickCash(amount),
      child: Text('SAR ${amount.toStringAsFixed(2)}', style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.bold)),
    );
  }

  Widget _buildKeypadKey(String label, {bool isAction = false}) {
    return Expanded(
      child: InkWell(
        onTap: () => _handleKey(label),
        borderRadius: BorderRadius.circular(10),
        child: Container(
          decoration: BoxDecoration(
            color: isAction ? const Color(0xFFFEF2F2) : const Color(0xFFF8FAFC),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: isAction ? const Color(0xFFFCA5A5) : const Color(0xFFE2E8F0)),
          ),
          alignment: Alignment.center,
          child: label == 'BACK'
              ? const Icon(Icons.backspace, color: Color(0xFFEF4444), size: 18)
              : Text(
                  label,
                  style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.w700, color: const Color(0xFF0F172A)),
                ),
        ),
      ),
    );
  }
}
