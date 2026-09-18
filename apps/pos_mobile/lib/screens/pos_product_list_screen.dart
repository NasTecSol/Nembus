import 'dart:convert';
import 'dart:developer' as developer;
import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import 'pos_sale_screen.dart';

class PosProductListScreen extends StatefulWidget {
  final int userId;
  final String username;
  final int? initialStoreId;
  final String? initialStoreName;
  final int? posTerminalId;
  final String? posTerminalName;
  final String? initialSaleType;

  const PosProductListScreen({
    super.key,
    required this.userId,
    this.username = 'Cashier',
    this.initialStoreId,
    this.initialStoreName,
    this.posTerminalId,
    this.posTerminalName,
    this.initialSaleType,
  });

  @override
  State<PosProductListScreen> createState() => _PosProductListScreenState();
}

class _PosProductListScreenState extends State<PosProductListScreen>
    with SingleTickerProviderStateMixin {
  bool _isLoading = true;
  String? _errorMessage;
  int? _activeStoreId;
  String _activeStoreName = 'Store';
  int _activePriceListId = 1;
  String _activePriceListName = 'Default Retail';
  List<Map<String, dynamic>> _products = [];
  List<Map<String, dynamic>> _categories = [];
  List<Map<String, dynamic>> _activePromotions = [];
  int? _selectedCategoryId;
  String _searchQuery = '';

  // Customer Management
  List<Map<String, dynamic>> _storeCustomers = [];
  Map<String, dynamic>? _selectedCustomer;
  bool _isLoadingCustomers = false;

  // Cart Management
  String _activeCartId = '';
  String _activeCartNumber = '';
  final Map<int, Map<String, dynamic>> _cartItemsMap = {}; // Keyed strictly by unique product_id

  static const List<String> _assetImages = [
    'images/image.png',
    'images/image (1).png',
    'images/image (2).png',
  ];

  @override
  void initState() {
    super.initState();
    _initializeCartSession();
    _resolveStoreAndFetchProducts();
    _fetchStoreCustomers();
  }

  String _generateUuidV4() {
    final now = DateTime.now().millisecondsSinceEpoch;
    final hexTime = now.toRadixString(16).padLeft(12, '0');
    return '00000000-0000-4000-8000-${hexTime.substring(hexTime.length - 12)}';
  }

  void _initializeCartSession() {
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    _activeCartId = _generateUuidV4();
    _activeCartNumber = 'CT-${timestamp.toString().substring(5)}';
    _ensureCartInLocalDb();
  }

  Future<void> _ensureCartInLocalDb() async {
    try {
      final db = DatabaseService().database;
      int orgId = 1;
      try {
        final orgs = await db.query('organizations', limit: 1);
        if (orgs.isNotEmpty) orgId = (orgs.first['id'] as num).toInt();
      } catch (_) {}

      final existing = await db.query('carts', where: 'id = ?', whereArgs: [_activeCartId]);
      if (existing.isEmpty) {
        await db.insert('carts', {
          'id': _activeCartId,
          'cart_number': _activeCartNumber,
          'organization_id': orgId,
          'store_id': _activeStoreId ?? 1,
          'customer_id': _selectedCustomer?['id'],
          'cart_status': 'active',
          'subtotal': _cartSubtotal,
          'total_amount': _cartTotalAmount,
          'created_at': DateTime.now().toIso8601String(),
          'updated_at': DateTime.now().toIso8601String(),
        });
      } else {
        await db.update('carts', {
          'customer_id': _selectedCustomer?['id'],
          'subtotal': _cartSubtotal,
          'total_amount': _cartTotalAmount,
          'updated_at': DateTime.now().toIso8601String(),
        }, where: 'id = ?', whereArgs: [_activeCartId]);
      }
    } catch (e) {
      developer.log('Error saving cart to local DB: $e', name: 'PosProductListScreen');
    }
  }

  /// Fetches customers for store via Go Core CustomerHandler / SQLite
  Future<void> _fetchStoreCustomers() async {
    setState(() => _isLoadingCustomers = true);
    try {
      final resp = NembusBridge().callHandler(
        handler: 'customer',
        action: 'listActiveCustomers',
        payload: {},
      );

      List<Map<String, dynamic>> custs = [];
      if (resp['success'] == true && resp['data'] is List && (resp['data'] as List).isNotEmpty) {
        custs = (resp['data'] as List)
            .map((e) => Map<String, dynamic>.from(e as Map))
            .toList();
      } else {
        try {
          final db = DatabaseService().database;
          final rows = await db.query('customers', orderBy: 'name ASC');
          custs = rows.map((r) => Map<String, dynamic>.from(r)).toList();
        } catch (dbErr) {
          developer.log('SQLite customers query: $dbErr', name: 'PosProductListScreen');
        }
      }

      if (mounted) {
        setState(() {
          _storeCustomers = custs;
          if (_selectedCustomer == null && custs.isNotEmpty) {
            _selectedCustomer = custs.first;
          }
          _isLoadingCustomers = false;
        });
      }
    } catch (e) {
      developer.log('Error fetching customers: $e', name: 'PosProductListScreen');
      if (mounted) setState(() => _isLoadingCustomers = false);
    }
  }

  /// Creates a new customer on-the-fly in store DB via Go CustomerHandler / SQLite
  Future<void> _createCustomerInStore({
    required String name,
    required String phone,
    required String email,
    required String customerCode,
    String customerType = 'retail',
    String address = '',
    double creditLimit = 0.0,
  }) async {
    try {
      final payload = {
        'name': name,
        'phone': phone,
        'email': email,
        'customer_code': customerCode,
        'customer_type': customerType,
        'address': address,
        'credit_limit': creditLimit,
        'organization_id': 1,
      };

      final resp = NembusBridge().callHandler(
        handler: 'customer',
        action: 'createCustomer',
        payload: payload,
      );

      Map<String, dynamic>? newCust;

      if (resp['success'] == true && resp['data'] is Map) {
        newCust = Map<String, dynamic>.from(resp['data'] as Map);
      } else {
        final db = DatabaseService().database;
        int orgId = 1;
        try {
          final orgRows = await db.query('organizations', limit: 1);
          if (orgRows.isNotEmpty) orgId = (orgRows.first['id'] as num).toInt();
        } catch (_) {}

        final newId = await db.insert('customers', {
          'organization_id': orgId,
          'customer_code': customerCode,
          'name': name,
          'phone': phone,
          'email': email,
          'customer_type': customerType,
          'address': address,
          'credit_limit': creditLimit,
          'loyalty_points': 50.0, // Welcome points
          'is_active': 1,
          'created_at': DateTime.now().toIso8601String(),
          'updated_at': DateTime.now().toIso8601String(),
        });

        newCust = {
          'id': newId,
          'name': name,
          'customer_code': customerCode,
          'phone': phone,
          'email': email,
          'customer_type': customerType,
          'address': address,
          'loyalty_points': 50.0,
        };
      }

      // Enqueue created customer into sync_queue for cloud synchronization
      if (newCust != null) {
        try {
          final db = DatabaseService().database;
          await db.insert('sync_queue', {
            'entity_type': 'customers',
            'entity_id': (newCust['id'] ?? customerCode).toString(),
            'action': 'INSERT',
            'payload': jsonEncode(newCust),
            'status': 'pending',
            'priority': 5,
            'created_at': DateTime.now().toIso8601String(),
          });
        } catch (syncErr) {
          developer.log('Customer outbox enqueue note: $syncErr', name: 'PosProductListScreen');
        }
      }

      if (newCust != null && mounted) {
        setState(() {
          _storeCustomers.insert(0, newCust!);
          _selectedCustomer = newCust;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFF10B981),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            content: Row(
              children: [
                const Icon(Icons.check_circle_rounded, color: Colors.white),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Customer "$name" registered & selected!',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ),
        );
      }
    } catch (e) {
      developer.log('Error creating customer: $e', name: 'PosProductListScreen');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFEF4444),
            behavior: SnackBarBehavior.floating,
            content: Text('Failed to create customer: $e'),
          ),
        );
      }
    }
  }

  /// Opens the 3D Animated Customer Picker / Creator Modal
  void _openCustomerPickerModal() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => _Customer3DPickerSheet(
        customers: _storeCustomers,
        selectedCustomer: _selectedCustomer,
        onSelectCustomer: (cust) {
          setState(() => _selectedCustomer = cust);
          Navigator.of(ctx).pop();
        },
        onCreateNewCustomer: (name, phone, email, code, type, address) async {
          await _createCustomerInStore(
            name: name,
            phone: phone,
            email: email,
            customerCode: code,
            customerType: type,
            address: address,
          );
          if (mounted) Navigator.of(ctx).pop();
        },
      ),
    );
  }

  /// Fetches products with distinct product IDs, SKUs, and Prices
  Future<void> _resolveStoreAndFetchProducts() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      int storeId = widget.initialStoreId ?? 1;
      String storeName = widget.initialStoreName ?? 'POS Store #1';

      if (widget.initialStoreId == null) {
        final storeResp = NembusBridge().callHandler(
          handler: 'user',
          action: 'getUserPrimaryStore',
          payload: {'id': widget.userId, 'user_id': widget.userId},
        );

        if (storeResp['success'] == true && storeResp['data'] is Map) {
          final sData = storeResp['data'] as Map;
          if (sData['id'] != null) {
            storeId = (sData['id'] as num).toInt();
            storeName = sData['name']?.toString() ?? 'Store #$storeId';
          }
        } else {
          final storesResp = NembusBridge().callHandler(
            handler: 'stores',
            action: 'listPOSEnabledStores',
            payload: {},
          );
          if (storesResp['success'] == true && storesResp['data'] is List) {
            final list = storesResp['data'] as List;
            if (list.isNotEmpty && list.first is Map) {
              final firstStore = list.first as Map;
              storeId = (firstStore['id'] as num?)?.toInt() ?? 1;
              storeName = firstStore['name']?.toString() ?? 'Store #$storeId';
            }
          }
        }
      }

      _activeStoreId = storeId;
      _activeStoreName = storeName;

      int priceListId = 1;
      String priceListName = 'Standard Retail';
      final priceListResp = NembusBridge().callHandler(
        handler: 'price_lists',
        action: 'getDefaultPriceList',
        payload: {},
      );

      if (priceListResp['success'] == true && priceListResp['data'] is Map) {
        final plData = priceListResp['data'] as Map;
        priceListId = (plData['id'] as num?)?.toInt() ?? 1;
        priceListName = plData['name']?.toString() ?? 'Default Retail';
      }

      _activePriceListId = priceListId;
      _activePriceListName = priceListName;

      List<Map<String, dynamic>> finalProductsList = [];
      final Set<int> seenProductIds = {};

      // 1. Primary POS Route: pos/stores/:store_id/products (GET /api/pos/stores/{store_id}/products)
      developer.log('Fetching POS products for store_id=$storeId via Go route pos/stores/$storeId/products', name: 'PosProductListScreen');
      final posProductsResp = NembusBridge().callHandler(
        handler: 'pos',
        action: 'listProducts',
        payload: {
          'store_id': storeId,
          'include_out_of_stock': true,
        },
      );

      if (posProductsResp['success'] == true &&
          posProductsResp['data'] is List &&
          (posProductsResp['data'] as List).isNotEmpty) {
        for (final item in posProductsResp['data'] as List) {
          if (item is Map) {
            final pid = (item['product_id'] ?? item['id'] as num?)?.toInt() ?? 0;
            if (pid > 0 && !seenProductIds.contains(pid)) {
              seenProductIds.add(pid);
              final priceVal = (item['effective_price'] ?? item['retail_price'] ?? item['price'] as num?)?.toDouble() ?? 0.0;
              final retailPrice = (item['retail_price'] as num?)?.toDouble() ?? priceVal;
              final effectivePrice = (item['effective_price'] as num?)?.toDouble() ?? priceVal;
              final stockQty = (item['quantity_available'] as num?)?.toInt() ??
                  (item['quantity_on_hand'] as num?)?.toInt() ??
                  0;

              finalProductsList.add({
                'product_id': pid,
                'product_name': item['product_name'] ?? 'Product #$pid',
                'sku': item['sku'] ?? item['product_sku'] ?? 'SKU-$pid',
                'description': item['description'] ?? '',
                'price': priceVal,
                'retail_price': retailPrice,
                'effective_price': effectivePrice,
                'category_id': item['category_id'],
                'category_name': item['category_name'] ?? '',
                'brand_name': item['brand_name'] ?? '',
                'barcode': item['barcode'] ?? '',
                'uom_code': item['uom_code'] ?? 'PCS',
                'decimal_places': item['decimal_places'] ?? 0,
                'quantity_available': stockQty,
                'quantity_on_hand': (item['quantity_on_hand'] as num?)?.toInt() ?? stockQty,
                'has_promotion': item['has_promotion'] == true || item['has_promotion'] == 1,
                'promotion_name': item['promotion_name'],
                'discount_percent': item['discount_percent'],
                'promo_min_quantity': item['promo_min_quantity'],
                'package_n_price': item['package_n_price'],
                'product_uom_conversions': item['product_uom_conversions'],
              });
            }
          }
        }
      }

      // 2. Secondary fallback via product_pricing and inventory_stock handlers if pos route returned empty
      if (finalProductsList.isEmpty) {
        final stockResp = NembusBridge().callHandler(
          handler: 'inventory_stock',
          action: 'listInventoryStockByStore',
          payload: {'store_id': storeId},
        );

        final Map<int, Map<String, dynamic>> stockByProduct = {};
        if (stockResp['success'] == true && stockResp['data'] is List) {
          for (final item in stockResp['data'] as List) {
            if (item is Map) {
              final pid = (item['product_id'] as num?)?.toInt();
              if (pid != null) {
                stockByProduct[pid] = Map<String, dynamic>.from(item);
              }
            }
          }
        }

        final pricingResp = NembusBridge().callHandler(
          handler: 'product_pricing',
          action: 'listPricesByPriceList',
          payload: {'price_list_id': priceListId},
        );

        if (pricingResp['success'] == true &&
            pricingResp['data'] is List &&
            (pricingResp['data'] as List).isNotEmpty) {
          int idx = 1;
          for (final pRow in pricingResp['data'] as List) {
            if (pRow is Map) {
              final rawPid = (pRow['product_id'] ?? pRow['id'] as num?)?.toInt();
              final pid = (rawPid != null && rawPid > 0) ? rawPid : idx;
              if (!seenProductIds.contains(pid)) {
                seenProductIds.add(pid);
                final stockInfo = stockByProduct[pid];
                final priceVal = pRow['price'];

                finalProductsList.add({
                  'product_id': pid,
                  'product_name': pRow['product_name'] ?? 'Product #$pid',
                  'sku': pRow['product_sku'] ?? 'SKU-$pid',
                  'description': pRow['description'] ?? '',
                  'price': priceVal,
                  'retail_price': priceVal,
                  'effective_price': priceVal,
                  'category_id': pRow['category_id'],
                  'category_name': pRow['category_name'] ?? '',
                  'quantity_available': (stockInfo?['quantity_available'] as num?)?.toInt() ?? (stockInfo?['quantity_on_hand'] as num?)?.toInt() ?? 0,
                  'quantity_on_hand': (stockInfo?['quantity_on_hand'] as num?)?.toInt() ?? 0,
                });
              }
              idx++;
            }
          }
        }
      }

      // 3. Tertiary fallback: direct SQLite query filtered by storeId
      if (finalProductsList.isEmpty) {
        try {
          final db = DatabaseService().database;
          final prodRows = await db.rawQuery('''
            SELECT p.id as product_id, p.name as product_name, p.sku, p.description,
                   c.name as category_name, p.category_id,
                   u.code as uom_code, u.name as uom_name, p.base_uom_id,
                   COALESCE(pp.price, 0.00) as price,
                   COALESCE(inv.quantity_available, inv.quantity_on_hand, 0) as quantity_available,
                   COALESCE(inv.quantity_on_hand, 0) as quantity_on_hand
            FROM products p
            LEFT JOIN categories c ON p.category_id = c.id
            LEFT JOIN units_of_measure u ON p.base_uom_id = u.id
            LEFT JOIN product_prices pp ON p.id = pp.product_id AND (pp.is_active = 1 OR pp.is_active IS NULL)
            LEFT JOIN inventory_stock inv ON p.id = inv.product_id AND inv.store_id = ?
            WHERE p.is_active = 1 OR p.is_active IS NULL
            ORDER BY p.id ASC
          ''', [storeId]);

          for (final row in prodRows) {
            final pid = (row['product_id'] as num?)?.toInt() ?? 0;
            if (pid > 0 && !seenProductIds.contains(pid)) {
              seenProductIds.add(pid);
              finalProductsList.add(Map<String, dynamic>.from(row));
            }
          }
        } catch (dbErr) {
          developer.log('SQLite products query: $dbErr', name: 'PosProductListScreen');
        }
      }

      // Fetch promotions from Go Core or local SQLite
      List<Map<String, dynamic>> activePromos = [];
      try {
        final promoResp = NembusBridge().callHandler(
          handler: 'promotion',
          action: 'listActivePromotions',
          payload: {'store_id': storeId},
        );
        if (promoResp['success'] == true && promoResp['data'] is List) {
          activePromos = (promoResp['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
      } catch (_) {}

      if (activePromos.isEmpty) {
        try {
          final db = DatabaseService().database;
          final promoRows = await db.query('promotions', where: 'is_active = 1 OR is_active IS NULL');
          activePromos = promoRows.map((r) => Map<String, dynamic>.from(r)).toList();
        } catch (dbErr) {
          developer.log('SQLite promotions query: $dbErr', name: 'PosProductListScreen');
        }
      }

      final catResp = NembusBridge().callHandler(
        handler: 'pos',
        action: 'getCategories',
        payload: {'store_id': storeId},
      );
      List<Map<String, dynamic>> cats = [];
      if (catResp['success'] == true && catResp['data'] is List) {
        cats = (catResp['data'] as List)
            .map((e) => Map<String, dynamic>.from(e as Map))
            .toList();
      }

      if (mounted) {
        setState(() {
          _categories = cats;
          _products = finalProductsList;
          _activePromotions = activePromos;
          _isLoading = false;
        });
      }
    } catch (e) {
      developer.log('Error loading products: $e', name: 'PosProductListScreen');
      if (mounted) {
        setState(() {
          _errorMessage = 'Failed to load products: $e';
          _isLoading = false;
        });
      }
    }
  }

  /// Displays a dialog informing the cashier that a session must be opened to add products
  void _showSessionClosedDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: const Color(0xFF131B2E),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
          side: BorderSide(color: const Color(0xFFF59E0B).withValues(alpha: 0.4)),
        ),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFFF59E0B).withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(10),
              ),
              child: const Icon(Icons.lock_clock_rounded, color: Color(0xFFF59E0B), size: 24),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                'Cashier Session Closed',
                style: GoogleFonts.inter(
                  color: Colors.white,
                  fontWeight: FontWeight.w700,
                  fontSize: 16,
                ),
              ),
            ),
          ],
        ),
        content: Text(
          'Cannot add products to cart because the cashier session is currently stopped or closed.\n\nPlease start a cashier session using the Play button on the POS Dashboard to enable sales and cart operations.',
          style: GoogleFonts.inter(color: Colors.white70, fontSize: 13, height: 1.4),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: Text('Dismiss', style: GoogleFonts.inter(color: Colors.white54)),
          ),
          ElevatedButton.icon(
            onPressed: () {
              Navigator.of(ctx).pop();
              Navigator.of(context).pop();
            },
            icon: const Icon(Icons.dashboard_rounded, size: 16),
            label: const Text('Go to Dashboard'),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF38BDF8),
              foregroundColor: Colors.black87,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
            ),
          ),
        ],
      ),
    );
  }

  Map<String, dynamic>? _getMatchingPromotionForProduct(Map<String, dynamic> product) {
    final pid = (product['product_id'] as num?)?.toInt() ?? 0;
    final catId = (product['category_id'] as num?)?.toInt();
    for (final promo in _activePromotions) {
      final appliesTo = (promo['applies_to'] ?? 'all').toString().toLowerCase();
      if (appliesTo == 'all') return promo;
      if (appliesTo == 'product') {
        final targetIds = promo['target_product_ids'];
        if (targetIds != null && targetIds.toString().contains('$pid')) {
          return promo;
        }
      }
      if (appliesTo == 'category' && catId != null) {
        final targetCatIds = promo['target_category_ids'];
        if (targetCatIds != null && targetCatIds.toString().contains('$catId')) {
          return promo;
        }
      }
    }
    return null;
  }

  void _openProductDetailModal(Map<String, dynamic> product) {
    final matchingPromo = _getMatchingPromotionForProduct(product);
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => _ProductUomAndPromotionBottomSheet(
        product: product,
        activePriceListId: _activePriceListId,
        matchedPromotion: matchingPromo,
        allPromotions: _activePromotions,
        initialSaleType: widget.initialSaleType,
        getProductAssetImage: _getProductAssetImage,
        onAddToCart: (qty, packageOption, effectivePrice, promoDesc) {
          _addProductToCart(
            product,
            quantity: qty,
            selectedPackage: packageOption,
            customUnitPrice: effectivePrice,
            appliedPromo: promoDesc,
          );
        },
      ),
    );
  }

  void _addProductToCart(
    Map<String, dynamic> product, {
    int quantity = 1,
    Map<String, dynamic>? selectedPackage,
    Map<String, dynamic>? selectedUom,
    double? customUnitPrice,
    String? appliedPromo,
  }) {
    if (!SingletonClass().isSessionActive) {
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          backgroundColor: const Color(0xFFEF4444),
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          content: const Row(
            children: [
              Icon(Icons.lock_clock_rounded, color: Colors.white, size: 18),
              SizedBox(width: 8),
              Expanded(
                child: Text(
                  'Cart locked: Cashier session is closed. Start session from Dashboard.',
                  style: TextStyle(color: Colors.white, fontSize: 12),
                ),
              ),
            ],
          ),
        ),
      );
      return;
    }

    if (_selectedCustomer == null) {
      _openCustomerPickerModal();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          backgroundColor: const Color(0xFFF59E0B),
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          content: const Text('Please select or create a customer before adding products'),
        ),
      );
      return;
    }

    final pid = (product['product_id'] as num?)?.toInt() ?? 0;
    if (pid == 0) return;

    final isWholesale = widget.initialSaleType?.toLowerCase() == 'wholesale';
    final stockQty = (product['quantity_available'] as num?)?.toInt() ?? 0;
    final currentQty = _cartItemsMap[pid]?['quantity'] ?? 0;

    // In wholesale scenario manage stock count: block if out of stock or stock exhausted
    if (isWholesale) {
      if (stockQty <= 0) {
        ScaffoldMessenger.of(context).hideCurrentSnackBar();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFEF4444),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            content: Row(
              children: [
                const Icon(Icons.block_rounded, color: Colors.white, size: 18),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Out of Stock: "${product['product_name']}" is not available in stock.',
                    style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
                  ),
                ),
              ],
            ),
          ),
        );
        return;
      }

      if (currentQty + quantity > stockQty) {
        ScaffoldMessenger.of(context).hideCurrentSnackBar();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFF59E0B),
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            content: Row(
              children: [
                const Icon(Icons.warning_amber_rounded, color: Colors.white, size: 18),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Cannot add more: Max available stock ($stockQty) already in cart.',
                    style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
                  ),
                ),
              ],
            ),
          ),
        );
        return;
      }
    }

    setState(() {
      final basePrice = double.tryParse(product['price']?.toString() ?? '0') ?? 0.0;
      final effectiveUnitPrice = customUnitPrice ?? basePrice;
      final pkgOption = selectedPackage ?? selectedUom;
      final pkgName = pkgOption?['package_name']?.toString() ?? pkgOption?['name']?.toString() ?? 'Single Unit';
      final uomCode = pkgOption?['uom_code']?.toString() ?? pkgOption?['code']?.toString() ?? (product['uom_code']?.toString() ?? 'PCS');
      final uomName = pkgOption?['uom_name']?.toString() ?? pkgOption?['name']?.toString() ?? (product['uom_name']?.toString() ?? 'Pieces');
      final convFactor = (pkgOption?['conversion_factor'] as num?)?.toDouble() ?? 1.0;
      final convText = pkgOption?['conversion_text']?.toString() ??
          (convFactor > 1
              ? '1 $uomCode = ${convFactor.toStringAsFixed(convFactor.truncateToDouble() == convFactor ? 0 : 2)} ${product['uom_code'] ?? 'PCS'}'
              : '1 $uomCode = 1 Base Unit');

      if (_cartItemsMap.containsKey(pid)) {
        final current = _cartItemsMap[pid]!;
        final currentQty = (current['quantity'] as num).toInt();
        final unitPrice = customUnitPrice ?? (current['unit_price'] as num).toDouble();
        final newQty = currentQty + quantity;
        _cartItemsMap[pid] = {
          ...current,
          'quantity': newQty,
          'unit_price': unitPrice,
          'package_name': pkgName,
          'uom_code': uomCode,
          'uom_name': uomName,
          'conversion_factor': convFactor,
          'conversion_text': convText,
          if (appliedPromo != null) 'applied_promo': appliedPromo,
          'line_total': newQty * unitPrice,
        };
      } else {
        _cartItemsMap[pid] = {
          'product_id': pid,
          'product_name': product['product_name'] ?? 'Product #$pid',
          'sku': product['sku'] ?? 'SKU-$pid',
          'unit_price': effectiveUnitPrice,
          'base_price': basePrice,
          'package_name': pkgName,
          'uom_code': uomCode,
          'uom_name': uomName,
          'conversion_factor': convFactor,
          'conversion_text': convText,
          if (appliedPromo != null) 'applied_promo': appliedPromo,
          'quantity': quantity,
          'line_total': effectiveUnitPrice * quantity,
        };
      }
    });

    _ensureCartInLocalDb();

    ScaffoldMessenger.of(context).hideCurrentSnackBar();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        backgroundColor: const Color(0xFF064E3B),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(milliseconds: 1000),
        content: Row(
          children: [
            const Icon(Icons.add_shopping_cart_rounded, color: Color(0xFF34D399), size: 18),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'Added $quantity of "${product['product_name']}" (${selectedPackage?['package_name'] ?? (selectedUom?['code'] ?? (product['uom_code'] ?? 'PCS'))})',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _discardCurrentCart({BuildContext? modalContext}) async {
    final oldCartId = _activeCartId;
    setState(() {
      _cartItemsMap.clear();
      _initializeCartSession();
    });

    // Delete cart & cart items from local SQLite
    try {
      final db = DatabaseService().database;
      await db.delete('cart_items', where: 'cart_id = ?', whereArgs: [oldCartId]);
      await db.delete('carts', where: 'id = ?', whereArgs: [oldCartId]);
    } catch (e) {
      developer.log('Error discarding empty cart from SQLite: $e', name: 'PosProductListScreen');
    }

    // Call Go bridge to clear cart
    try {
      NembusBridge().callHandler(
        handler: 'cart',
        action: 'clearCart',
        payload: {'cart_id': oldCartId},
      );
    } catch (_) {}

    // Dismiss the modal sheet if it's currently open
    if (modalContext != null && Navigator.of(modalContext).canPop()) {
      Navigator.of(modalContext).pop();
    }

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          backgroundColor: const Color(0xFF1E293B),
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
          content: Row(
            children: const [
              Icon(Icons.remove_shopping_cart_rounded, color: Color(0xFFF59E0B), size: 18),
              SizedBox(width: 8),
              Text('Cart was empty and has been discarded.', style: TextStyle(color: Colors.white, fontSize: 13)),
            ],
          ),
          duration: const Duration(seconds: 2),
        ),
      );
    }
  }

  void _decrementCartItem(int productId, {BuildContext? modalContext}) {
    setState(() {
      if (_cartItemsMap.containsKey(productId)) {
        final current = _cartItemsMap[productId]!;
        final currentQty = (current['quantity'] as num).toInt();
        final unitPrice = (current['unit_price'] as num).toDouble();
        if (currentQty > 1) {
          final newQty = currentQty - 1;
          _cartItemsMap[productId] = {
            ...current,
            'quantity': newQty,
            'line_total': newQty * unitPrice,
          };
        } else {
          _cartItemsMap.remove(productId);
        }
      }
    });

    if (_cartItemsMap.isEmpty) {
      _discardCurrentCart(modalContext: modalContext);
    } else {
      _ensureCartInLocalDb();
    }
  }

  void _removeCartItem(int productId, {BuildContext? modalContext}) {
    setState(() {
      _cartItemsMap.remove(productId);
    });

    if (_cartItemsMap.isEmpty) {
      _discardCurrentCart(modalContext: modalContext);
    } else {
      _ensureCartInLocalDb();
    }
  }

  void _clearCart({BuildContext? modalContext}) {
    _discardCurrentCart(modalContext: modalContext);
  }

  int get _totalCartItemCount {
    return _cartItemsMap.values.fold<int>(
      0,
      (sum, item) => sum + (item['quantity'] as num).toInt(),
    );
  }

  double get _cartSubtotal {
    final raw = _cartItemsMap.values.fold<double>(
      0.0,
      (sum, item) => sum + (item['line_total'] as num).toDouble(),
    );
    return double.parse(raw.toStringAsFixed(2));
  }

  double get _cartTaxAmount => double.parse((_cartSubtotal * 0.15).toStringAsFixed(2)); // 15% VAT rounded to 2 decimals
  double get _cartTotalAmount => double.parse((_cartSubtotal + _cartTaxAmount).toStringAsFixed(2));

  void _openCartModalSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (context, setModalState) {
          return _Cart3DReviewSheet(
            cartItems: _cartItemsMap.values.toList(),
            subtotal: _cartSubtotal,
            taxAmount: _cartTaxAmount,
            totalAmount: _cartTotalAmount,
            customer: _selectedCustomer,
            onIncrement: (pid) {
              final prod = _products.firstWhere((p) => (p['product_id'] as num?)?.toInt() == pid);
              _addProductToCart(prod);
              setModalState(() {});
            },
            onDecrement: (pid) {
              _decrementCartItem(pid, modalContext: ctx);
              if (_cartItemsMap.isNotEmpty && mounted) {
                setModalState(() {});
              }
            },
            onRemove: (pid) {
              _removeCartItem(pid, modalContext: ctx);
              if (_cartItemsMap.isNotEmpty && mounted) {
                setModalState(() {});
              }
            },
            onCheckout: () {
              Navigator.of(ctx).pop();
              _proceedToCheckout();
            },
          );
        },
      ),
    );
  }

  void _proceedToCheckout() {
    if (_cartItemsMap.isEmpty) return;
    if (_selectedCustomer == null) {
      _openCustomerPickerModal();
      return;
    }

    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => PosSaleScreen(
          cartId: _activeCartId,
          cartNumber: _activeCartNumber,
          customer: _selectedCustomer!,
          cartItems: _cartItemsMap.values.toList(),
          subtotal: _cartSubtotal,
          taxAmount: _cartTaxAmount,
          totalAmount: _cartTotalAmount,
          storeId: _activeStoreId ?? 1,
          storeName: _activeStoreName,
          posTerminalId: widget.posTerminalId ?? SingletonClass().activeTerminalId,
          posTerminalName: widget.posTerminalName ?? SingletonClass().activeTerminalName,
          userId: widget.userId,
          username: widget.username,
          onOrderCompleted: () {
            _clearCart();
          },
        ),
      ),
    );
  }

  String _formatPrice(dynamic price) {
    if (price == null) return '0.00';
    if (price is num) return price.toStringAsFixed(2);
    final parsed = double.tryParse(price.toString());
    if (parsed != null) return parsed.toStringAsFixed(2);
    return price.toString();
  }

  String _getProductAssetImage(int productId, int index) {
    if (_assetImages.isEmpty) return '';
    final hash = (productId != 0 ? productId.abs() : (index + 1)) * 31;
    final selectedIndex = hash % _assetImages.length;
    return _assetImages[selectedIndex];
  }

  @override
  Widget build(BuildContext context) {
    const bgColor = Color(0xFF070B14);
    const cardBgColor = Color(0xFF111827);
    const surfaceColor = Color(0xFF1F2937);
    const primarySky = Color(0xFF38BDF8);
    const primaryIndigo = Color(0xFF6366F1);
    const emeraldGreen = Color(0xFF10B981);

    final filteredProducts = _products.where((p) {
      if (_selectedCategoryId != null) {
        final catId = (p['category_id'] as num?)?.toInt();
        if (catId != _selectedCategoryId) return false;
      }
      if (_searchQuery.trim().isNotEmpty) {
        final query = _searchQuery.toLowerCase();
        final name = (p['product_name'] ?? p['name'] ?? '').toString().toLowerCase();
        final sku = (p['sku'] ?? '').toString().toLowerCase();
        return name.contains(query) || sku.contains(query);
      }
      return true;
    }).toList();

    return Scaffold(
      backgroundColor: bgColor,
      appBar: AppBar(
        backgroundColor: cardBgColor,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded, color: Colors.white70),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              widget.initialSaleType?.toLowerCase() == 'wholesale'
                  ? 'POS Wholesale'
                  : (widget.initialSaleType?.toLowerCase() == 'retail'
                      ? 'POS Retail'
                      : 'POS Product Catalog'),
              style: GoogleFonts.inter(
                fontWeight: FontWeight.w800,
                fontSize: 17,
                color: Colors.white,
                letterSpacing: -0.3,
              ),
            ),
            Text(
              '$_activeStoreName (ID: ${_activeStoreId ?? 1}) • $_activePriceListName',
              overflow: TextOverflow.ellipsis,
              style: GoogleFonts.inter(fontSize: 11, color: Colors.white54),
            ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: 'Sync Products',
            icon: const Icon(Icons.refresh_rounded, color: primarySky),
            onPressed: () {
              _resolveStoreAndFetchProducts();
              _fetchStoreCustomers();
            },
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            final screenWidth = constraints.maxWidth;
            final isTablet = screenWidth >= 600;
            final hPadding = isTablet ? 24.0 : 16.0;

            int crossAxisCount = 2;
            double childAspectRatio = 0.72;
            if (screenWidth >= 1200) {
              crossAxisCount = 5;
              childAspectRatio = 0.80;
            } else if (screenWidth >= 900) {
              crossAxisCount = 4;
              childAspectRatio = 0.78;
            } else if (screenWidth >= 650) {
              crossAxisCount = 3;
              childAspectRatio = 0.76;
            } else {
              crossAxisCount = 2;
              childAspectRatio = 0.72;
            }

            return Column(
              children: [
                // 3D Glass Customer Header Card
                Container(
                  margin: EdgeInsets.fromLTRB(hPadding, 12, hPadding, 8),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(18),
                    gradient: LinearGradient(
                      colors: _selectedCustomer != null
                          ? [
                              const Color(0xFF065F46).withValues(alpha: 0.7),
                              const Color(0xFF1E293B).withValues(alpha: 0.9),
                            ]
                          : [
                              const Color(0xFF854D0E).withValues(alpha: 0.7),
                              const Color(0xFF1E293B).withValues(alpha: 0.9),
                            ],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    border: Border.all(
                      color: _selectedCustomer != null
                          ? emeraldGreen.withValues(alpha: 0.5)
                          : const Color(0xFFF59E0B).withValues(alpha: 0.5),
                      width: 1.5,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: (_selectedCustomer != null ? emeraldGreen : const Color(0xFFF59E0B))
                            .withValues(alpha: 0.2),
                        blurRadius: 16,
                        offset: const Offset(0, 6),
                      ),
                    ],
                  ),
                  child: Material(
                    color: Colors.transparent,
                    child: InkWell(
                      borderRadius: BorderRadius.circular(18),
                      onTap: _openCustomerPickerModal,
                      child: Padding(
                        padding: EdgeInsets.symmetric(horizontal: isTablet ? 20 : 16, vertical: isTablet ? 14 : 12),
                        child: Row(
                          children: [
                            // 3D Avatar Halo
                            Container(
                              width: isTablet ? 48 : 44,
                              height: isTablet ? 48 : 44,
                              decoration: BoxDecoration(
                                shape: BoxShape.circle,
                                gradient: LinearGradient(
                                  colors: _selectedCustomer != null
                                      ? [emeraldGreen, primarySky]
                                      : [const Color(0xFFF59E0B), const Color(0xFFEF4444)],
                                ),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black.withValues(alpha: 0.4),
                                    blurRadius: 8,
                                    offset: const Offset(0, 4),
                                  ),
                                ],
                              ),
                              alignment: Alignment.center,
                              child: Text(
                                (_selectedCustomer?['name'] ?? 'C')[0].toUpperCase(),
                                style: GoogleFonts.inter(
                                  fontSize: isTablet ? 20 : 18,
                                  fontWeight: FontWeight.w900,
                                  color: Colors.white,
                                ),
                              ),
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
                                          _selectedCustomer != null
                                              ? _selectedCustomer!['name']
                                              : 'No Customer Selected',
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                          style: GoogleFonts.inter(
                                            color: Colors.white,
                                            fontSize: isTablet ? 15.5 : 14.5,
                                            fontWeight: FontWeight.w800,
                                          ),
                                        ),
                                      ),
                                      if (_selectedCustomer != null) ...[
                                        const SizedBox(width: 6),
                                        Container(
                                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                          decoration: BoxDecoration(
                                            color: Colors.black.withValues(alpha: 0.4),
                                            borderRadius: BorderRadius.circular(6),
                                            border: Border.all(color: emeraldGreen.withValues(alpha: 0.4)),
                                          ),
                                          child: Text(
                                            (_selectedCustomer!['customer_type'] ?? 'Retail').toString().toUpperCase(),
                                            style: GoogleFonts.inter(
                                              fontSize: 9,
                                              fontWeight: FontWeight.w800,
                                              color: emeraldGreen,
                                            ),
                                          ),
                                        ),
                                      ],
                                    ],
                                  ),
                                  const SizedBox(height: 2),
                                  Text(
                                    _selectedCustomer != null
                                        ? 'Code: ${_selectedCustomer!['customer_code'] ?? 'N/A'} • Points: ${_selectedCustomer!['loyalty_points'] ?? '50'} pts'
                                        : 'Tap to pick or register in-store customer (Required)',
                                    style: GoogleFonts.inter(color: Colors.white60, fontSize: isTablet ? 12 : 11.5),
                                  ),
                                ],
                              ),
                            ),
                            // 3D Glass Action Button
                            Container(
                              padding: EdgeInsets.symmetric(horizontal: isTablet ? 14 : 10, vertical: isTablet ? 8 : 6),
                              decoration: BoxDecoration(
                                gradient: const LinearGradient(
                                  colors: [primaryIndigo, primarySky],
                                ),
                                borderRadius: BorderRadius.circular(10),
                                boxShadow: [
                                  BoxShadow(
                                    color: primarySky.withValues(alpha: 0.3),
                                    blurRadius: 8,
                                    offset: const Offset(0, 3),
                                  ),
                                ],
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(
                                    _selectedCustomer != null ? Icons.swap_horiz_rounded : Icons.person_add_rounded,
                                    color: Colors.white,
                                    size: 14,
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    _selectedCustomer != null ? 'Change' : 'Select',
                                    style: GoogleFonts.inter(
                                      color: Colors.white,
                                      fontWeight: FontWeight.w800,
                                      fontSize: isTablet ? 12.5 : 11.5,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),

                // Cashier Session Closed Warning Banner
                if (!SingletonClass().isSessionActive)
                  Container(
                    margin: EdgeInsets.fromLTRB(hPadding, 8, hPadding, 6),
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                    decoration: BoxDecoration(
                      color: const Color(0xFFEF4444).withValues(alpha: 0.12),
                      borderRadius: BorderRadius.circular(14),
                      border: Border.all(color: const Color(0xFFEF4444).withValues(alpha: 0.4)),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.lock_clock_rounded, color: Color(0xFFEF4444), size: 20),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Text(
                            'Session Closed — Cart additions are locked. Start a session from the Dashboard.',
                            style: GoogleFonts.inter(
                              color: const Color(0xFFFCA5A5),
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                        TextButton(
                          onPressed: () => Navigator.of(context).pop(),
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                            minimumSize: Size.zero,
                            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                          ),
                          child: Text(
                            'Dashboard',
                            style: GoogleFonts.inter(
                              color: const Color(0xFF38BDF8),
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),

                // Search Bar with 3D inset depth
                Padding(
                  padding: EdgeInsets.symmetric(horizontal: hPadding, vertical: 4),
                  child: Container(
                    height: 48,
                    decoration: BoxDecoration(
                      color: surfaceColor,
                      borderRadius: BorderRadius.circular(14),
                      border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withValues(alpha: 0.3),
                          blurRadius: 6,
                          offset: const Offset(0, 3),
                        ),
                      ],
                    ),
                    child: TextField(
                      onChanged: (val) => setState(() => _searchQuery = val),
                      style: GoogleFonts.inter(color: Colors.white, fontSize: 13.5),
                      decoration: InputDecoration(
                        hintText: 'Search SKU, product title...',
                        hintStyle: GoogleFonts.inter(color: Colors.white38, fontSize: 12.5),
                        prefixIcon: const Icon(Icons.search_rounded, color: primarySky, size: 20),
                        border: InputBorder.none,
                        contentPadding: const EdgeInsets.symmetric(vertical: 14),
                      ),
                    ),
                  ),
                ),

                // Category Chips with neon active state
                if (_categories.isNotEmpty) ...[
                  Container(
                    height: 42,
                    margin: const EdgeInsets.symmetric(vertical: 6),
                    child: ListView.separated(
                      padding: EdgeInsets.symmetric(horizontal: hPadding),
                      scrollDirection: Axis.horizontal,
                      itemCount: _categories.length + 1,
                      separatorBuilder: (_, __) => const SizedBox(width: 8),
                      itemBuilder: (context, idx) {
                        if (idx == 0) {
                          final isSelected = _selectedCategoryId == null;
                          return ChoiceChip(
                            label: Text('All (${_products.length})'),
                            selected: isSelected,
                            onSelected: (_) => _filterProductsByCategory(null),
                            selectedColor: primarySky,
                            labelStyle: GoogleFonts.inter(
                              fontSize: 11.5,
                              fontWeight: FontWeight.w700,
                              color: isSelected ? const Color(0xFF090D16) : Colors.white70,
                            ),
                            backgroundColor: cardBgColor,
                          );
                        }
                        final cat = _categories[idx - 1];
                        final catId = (cat['id'] as num?)?.toInt();
                        final isSelected = _selectedCategoryId == catId;
                        return ChoiceChip(
                          label: Text(cat['name']?.toString() ?? 'Category'),
                          selected: isSelected,
                          onSelected: (_) => _filterProductsByCategory(catId),
                          selectedColor: primarySky,
                          labelStyle: GoogleFonts.inter(
                            fontSize: 11.5,
                            fontWeight: FontWeight.w700,
                            color: isSelected ? const Color(0xFF090D16) : Colors.white70,
                          ),
                          backgroundColor: cardBgColor,
                        );
                      },
                    ),
                  ),
                ],

                // 3D Product Catalog Grid
                Expanded(
                  child: _isLoading
                      ? const Center(child: CircularProgressIndicator(color: primarySky))
                      : _errorMessage != null
                          ? Center(child: Text(_errorMessage!, style: const TextStyle(color: Colors.redAccent)))
                          : filteredProducts.isEmpty
                              ? Center(
                                  child: Text(
                                    'No products found.',
                                    style: GoogleFonts.inter(color: Colors.white54),
                                  ),
                                )
                              : GridView.builder(
                                  padding: EdgeInsets.fromLTRB(hPadding, 6, hPadding, 90),
                                  gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                                    crossAxisCount: crossAxisCount,
                                    childAspectRatio: childAspectRatio,
                                    crossAxisSpacing: isTablet ? 16 : 14,
                                    mainAxisSpacing: isTablet ? 16 : 14,
                                  ),
                                  itemCount: filteredProducts.length,
                                  itemBuilder: (context, index) {
                                    final product = filteredProducts[index];
                                    final pid = (product['product_id'] as num?)?.toInt() ?? (index + 1);
                                    final name = product['product_name'] ?? 'Product #$pid';
                                    final sku = product['sku'] ?? 'SKU-$pid';
                                    final priceStr = _formatPrice(product['price']);
                                    final productUomCode = product['uom_code']?.toString() ?? 'PCS';
                                    final promo = _getMatchingPromotionForProduct(product);
                                    final isWholesale = widget.initialSaleType?.toLowerCase() == 'wholesale';
                                    final stockQty = (product['quantity_available'] as num?)?.toInt() ?? 0;
                                    final cartQty = _cartItemsMap[pid]?['quantity'] ?? 0;
                                    // Out of stock management strictly applies to Wholesale
                                    final isOutOfStock = isWholesale && stockQty <= 0;
                                    final isStockExhausted = isWholesale && (cartQty >= stockQty && stockQty > 0);
                                    final canAddToCart = !isWholesale || (!isOutOfStock && !isStockExhausted);
                                    final inStock = stockQty > 0;
                                    final assetImg = _getProductAssetImage(pid, index);

                                    return Material(
                                      color: Colors.transparent,
                                      child: InkWell(
                                        onTap: () => _openProductDetailModal(product),
                                        borderRadius: BorderRadius.circular(20),
                                        child: Container(
                                          decoration: BoxDecoration(
                                            color: cardBgColor,
                                            borderRadius: BorderRadius.circular(20),
                                            border: Border.all(
                                              color: cartQty > 0
                                                  ? primarySky.withValues(alpha: 0.8)
                                                  : (isOutOfStock
                                                      ? const Color(0xFFEF4444).withValues(alpha: 0.4)
                                                      : Colors.white.withValues(alpha: 0.08)),
                                              width: cartQty > 0 ? 1.8 : 1,
                                            ),
                                            boxShadow: [
                                              BoxShadow(
                                                color: cartQty > 0
                                                    ? primarySky.withValues(alpha: 0.25)
                                                    : Colors.black.withValues(alpha: 0.35),
                                                blurRadius: cartQty > 0 ? 14 : 8,
                                                offset: const Offset(0, 5),
                                              ),
                                            ],
                                          ),
                                          child: Column(
                                            crossAxisAlignment: CrossAxisAlignment.start,
                                            children: [
                                              // Image with 3D Overlay Badge
                                              Stack(
                                                children: [
                                                  ClipRRect(
                                                    borderRadius: const BorderRadius.vertical(top: Radius.circular(19)),
                                                    child: Container(
                                                      height: isTablet ? 128 : 116,
                                                      width: double.infinity,
                                                      color: surfaceColor,
                                                      child: Stack(
                                                        fit: StackFit.expand,
                                                        children: [
                                                          assetImg.isNotEmpty
                                                              ? Opacity(
                                                                  opacity: isOutOfStock ? 0.45 : 1.0,
                                                                  child: Image.asset(assetImg, fit: BoxFit.cover),
                                                                )
                                                              : const Icon(Icons.inventory_2_outlined, color: Colors.white38, size: 36),
                                                          if (isOutOfStock)
                                                            Center(
                                                              child: Container(
                                                                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                                                decoration: BoxDecoration(
                                                                  color: const Color(0xFFEF4444).withValues(alpha: 0.92),
                                                                  borderRadius: BorderRadius.circular(8),
                                                                  boxShadow: [
                                                                    BoxShadow(
                                                                      color: Colors.black.withValues(alpha: 0.5),
                                                                      blurRadius: 8,
                                                                    ),
                                                                  ],
                                                                ),
                                                                child: Text(
                                                                  'OUT OF STOCK',
                                                                  style: GoogleFonts.inter(
                                                                    fontSize: 10,
                                                                    fontWeight: FontWeight.w900,
                                                                    letterSpacing: 0.8,
                                                                    color: Colors.white,
                                                                  ),
                                                                ),
                                                              ),
                                                            ),
                                                        ],
                                                      ),
                                                    ),
                                                  ),
                                                  Positioned(
                                                    top: 8,
                                                    left: 8,
                                                    child: Container(
                                                      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
                                                      decoration: BoxDecoration(
                                                        color: Colors.black.withValues(alpha: 0.85),
                                                        borderRadius: BorderRadius.circular(8),
                                                        border: Border.all(
                                                          color: isOutOfStock
                                                              ? const Color(0xFFEF4444)
                                                              : (isStockExhausted
                                                                  ? const Color(0xFFF59E0B)
                                                                  : emeraldGreen),
                                                          width: 0.9,
                                                        ),
                                                      ),
                                                      child: Row(
                                                        mainAxisSize: MainAxisSize.min,
                                                        children: [
                                                          Icon(
                                                            isOutOfStock
                                                                ? Icons.block_rounded
                                                                : (isStockExhausted
                                                                    ? Icons.warning_amber_rounded
                                                                    : (inStock ? Icons.inventory_2_outlined : Icons.check_circle_outline_rounded)),
                                                            size: 11,
                                                            color: isOutOfStock
                                                                ? const Color(0xFFEF4444)
                                                                : (isStockExhausted ? const Color(0xFFF59E0B) : emeraldGreen),
                                                          ),
                                                          const SizedBox(width: 4),
                                                          Text(
                                                            isOutOfStock
                                                                ? 'Out of Stock'
                                                                : (isStockExhausted
                                                                    ? 'Max in Cart ($stockQty)'
                                                                    : (inStock ? 'Stock: $stockQty' : 'In Stock')),
                                                            style: GoogleFonts.inter(
                                                              fontSize: 9.5,
                                                              fontWeight: FontWeight.w800,
                                                              color: isOutOfStock
                                                                  ? const Color(0xFFEF4444)
                                                                  : (isStockExhausted ? const Color(0xFFF59E0B) : emeraldGreen),
                                                            ),
                                                          ),
                                                        ],
                                                      ),
                                                    ),
                                                  ),
                                                  if (promo != null)
                                                    Positioned(
                                                      top: 8,
                                                      right: cartQty > 0 ? 84 : 8,
                                                      child: Container(
                                                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                                                        decoration: BoxDecoration(
                                                          gradient: const LinearGradient(
                                                            colors: [Color(0xFFF59E0B), Color(0xFFEF4444)],
                                                          ),
                                                          borderRadius: BorderRadius.circular(8),
                                                          boxShadow: [
                                                            BoxShadow(
                                                              color: const Color(0xFFEF4444).withValues(alpha: 0.4),
                                                              blurRadius: 6,
                                                              offset: const Offset(0, 2),
                                                            ),
                                                          ],
                                                        ),
                                                        child: Row(
                                                          mainAxisSize: MainAxisSize.min,
                                                          children: [
                                                            const Icon(Icons.local_offer_rounded, size: 9, color: Colors.white),
                                                            const SizedBox(width: 3),
                                                            Text(
                                                              promo['promotion_type'] == 'percentage_discount'
                                                                  ? '${(promo['discount_value'] as num?)?.toInt() ?? 10}% OFF'
                                                                  : 'PROMO',
                                                              style: GoogleFonts.inter(
                                                                fontSize: 9,
                                                                fontWeight: FontWeight.w900,
                                                                color: Colors.white,
                                                              ),
                                                            ),
                                                          ],
                                                        ),
                                                      ),
                                                    ),
                                                  if (cartQty > 0)
                                                    Positioned(
                                                      top: 8,
                                                      right: 8,
                                                      child: Container(
                                                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                                        decoration: BoxDecoration(
                                                          gradient: const LinearGradient(
                                                            colors: [primaryIndigo, primarySky],
                                                          ),
                                                          borderRadius: BorderRadius.circular(8),
                                                          boxShadow: [
                                                            BoxShadow(
                                                              color: primarySky.withValues(alpha: 0.4),
                                                              blurRadius: 6,
                                                              offset: const Offset(0, 2),
                                                            ),
                                                          ],
                                                        ),
                                                        child: Text(
                                                          '${cartQty}x in cart',
                                                          style: GoogleFonts.inter(
                                                            fontSize: 10,
                                                            fontWeight: FontWeight.w900,
                                                            color: Colors.white,
                                                          ),
                                                        ),
                                                      ),
                                                    ),
                                                ],
                                              ),

                                              // Details
                                              Expanded(
                                                child: Padding(
                                                  padding: const EdgeInsets.all(10),
                                                  child: Column(
                                                    crossAxisAlignment: CrossAxisAlignment.start,
                                                    children: [
                                                      Text(
                                                        sku,
                                                        maxLines: 1,
                                                        overflow: TextOverflow.ellipsis,
                                                        style: GoogleFonts.inter(fontSize: 10, color: Colors.white38, fontWeight: FontWeight.w600),
                                                      ),
                                                      const SizedBox(height: 2),
                                                      Text(
                                                        name.toString(),
                                                        maxLines: 2,
                                                        overflow: TextOverflow.ellipsis,
                                                        style: GoogleFonts.inter(
                                                          fontSize: isTablet ? 13.5 : 13,
                                                          fontWeight: FontWeight.w700,
                                                          color: isOutOfStock && isWholesale ? Colors.white60 : Colors.white,
                                                          height: 1.2,
                                                        ),
                                                      ),
                                                      const Spacer(),
                                                      Row(
                                                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                                        children: [
                                                          Flexible(
                                                            child: Row(
                                                              mainAxisSize: MainAxisSize.min,
                                                              children: [
                                                                Flexible(
                                                                  child: Text(
                                                                    'SAR $priceStr',
                                                                    maxLines: 1,
                                                                    overflow: TextOverflow.ellipsis,
                                                                    style: GoogleFonts.inter(
                                                                      fontSize: isTablet ? 14.5 : 13.5,
                                                                      fontWeight: FontWeight.w900,
                                                                      color: isOutOfStock && isWholesale ? Colors.white38 : primarySky,
                                                                    ),
                                                                  ),
                                                                ),
                                                                const SizedBox(width: 4),
                                                                Container(
                                                                  padding: const EdgeInsets.symmetric(horizontal: 4.5, vertical: 1.5),
                                                                  decoration: BoxDecoration(
                                                                    color: primarySky.withValues(alpha: 0.12),
                                                                    borderRadius: BorderRadius.circular(5),
                                                                    border: Border.all(color: primarySky.withValues(alpha: 0.25), width: 0.6),
                                                                  ),
                                                                  child: Text(
                                                                    productUomCode,
                                                                    style: GoogleFonts.inter(
                                                                      fontSize: 8.5,
                                                                      fontWeight: FontWeight.w800,
                                                                      color: primarySky,
                                                                    ),
                                                                  ),
                                                                ),
                                                              ],
                                                            ),
                                                          ),
                                                          const SizedBox(width: 4),
                                                          // 3D Add Button
                                                          Material(
                                                            color: Colors.transparent,
                                                            child: InkWell(
                                                              onTap: () => _openProductDetailModal(product),
                                                              borderRadius: BorderRadius.circular(10),
                                                              child: Container(
                                                                padding: const EdgeInsets.all(7),
                                                                decoration: BoxDecoration(
                                                                  gradient: canAddToCart
                                                                      ? const LinearGradient(
                                                                          colors: [primaryIndigo, primarySky],
                                                                        )
                                                                      : const LinearGradient(
                                                                          colors: [Color(0xFF334155), Color(0xFF1E293B)],
                                                                        ),
                                                                  borderRadius: BorderRadius.circular(10),
                                                                  boxShadow: [
                                                                    if (canAddToCart)
                                                                      BoxShadow(
                                                                        color: primarySky.withValues(alpha: 0.3),
                                                                        blurRadius: 6,
                                                                        offset: const Offset(0, 2),
                                                                      ),
                                                                  ],
                                                                ),
                                                                child: Icon(
                                                                  isOutOfStock && isWholesale
                                                                      ? Icons.block_rounded
                                                                      : (isStockExhausted ? Icons.lock_clock_rounded : Icons.add_rounded),
                                                                  color: canAddToCart ? Colors.white : Colors.white38,
                                                                  size: 18,
                                                                ),
                                                              ),
                                                            ),
                                                          ),
                                                        ],
                                                      ),
                                                    ],
                                                  ),
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ),
                                    );
                                  },
                                ),
                ),
              ],
            );
          },
        ),
      ),

      // 3D Floating Glowing Bottom Cart Bar
      bottomNavigationBar: _cartItemsMap.isNotEmpty
          ? SafeArea(
              top: false,
              child: Center(
                heightFactor: 1.0,
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 720),
                  child: Container(
                    margin: const EdgeInsets.fromLTRB(16, 4, 16, 12),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    decoration: BoxDecoration(
                      borderRadius: BorderRadius.circular(22),
                      gradient: const LinearGradient(
                        colors: [Color(0xFF1E1B4B), Color(0xFF0F172A)],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      border: Border.all(color: primarySky.withValues(alpha: 0.4), width: 1.5),
                      boxShadow: [
                        BoxShadow(
                          color: primarySky.withValues(alpha: 0.25),
                          blurRadius: 20,
                          offset: const Offset(0, -4),
                        ),
                      ],
                    ),
                    child: Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            gradient: const LinearGradient(colors: [primaryIndigo, primarySky]),
                            borderRadius: BorderRadius.circular(14),
                            boxShadow: [
                              BoxShadow(
                                color: primarySky.withValues(alpha: 0.4),
                                blurRadius: 8,
                                offset: const Offset(0, 3),
                              ),
                            ],
                          ),
                          child: Stack(
                            clipBehavior: Clip.none,
                            children: [
                              const Icon(Icons.shopping_cart_rounded, color: Colors.white, size: 22),
                              Positioned(
                                right: -6,
                                top: -6,
                                child: Container(
                                  padding: const EdgeInsets.all(4),
                                  decoration: const BoxDecoration(
                                    color: emeraldGreen,
                                    shape: BoxShape.circle,
                                  ),
                                  child: Text(
                                    '$_totalCartItemCount',
                                    style: const TextStyle(color: Colors.black, fontSize: 9.5, fontWeight: FontWeight.w900),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 14),
                        Expanded(
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'SAR ${_cartTotalAmount.toStringAsFixed(2)}',
                                style: GoogleFonts.inter(
                                  color: Colors.white,
                                  fontSize: 17,
                                  fontWeight: FontWeight.w900,
                                  letterSpacing: -0.3,
                                ),
                              ),
                              Text(
                                'Incl. VAT (15%) • ${_selectedCustomer?['name'] ?? 'No Customer'}',
                                style: GoogleFonts.inter(color: Colors.white60, fontSize: 11.5),
                              ),
                            ],
                          ),
                        ),
                        ElevatedButton(
                          onPressed: _openCartModalSheet,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: emeraldGreen,
                            foregroundColor: Colors.white,
                            padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                            elevation: 4,
                            shadowColor: emeraldGreen.withValues(alpha: 0.5),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text(
                                'View Cart',
                                style: GoogleFonts.inter(fontWeight: FontWeight.w800, fontSize: 13),
                              ),
                              const SizedBox(width: 4),
                              const Icon(Icons.arrow_forward_rounded, size: 16),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            )
          : null,
    );
  }

  void _filterProductsByCategory(int? categoryId) {
    setState(() => _selectedCategoryId = categoryId);
  }
}

/// 3D Animated Customer Selector and In-Store Customer Creator Modal Sheet
class _Customer3DPickerSheet extends StatefulWidget {
  final List<Map<String, dynamic>> customers;
  final Map<String, dynamic>? selectedCustomer;
  final Function(Map<String, dynamic>) onSelectCustomer;
  final Function(String name, String phone, String email, String code, String type, String address)
      onCreateNewCustomer;

  const _Customer3DPickerSheet({
    required this.customers,
    required this.selectedCustomer,
    required this.onSelectCustomer,
    required this.onCreateNewCustomer,
  });

  @override
  State<_Customer3DPickerSheet> createState() => _Customer3DPickerSheetState();
}

class _Customer3DPickerSheetState extends State<_Customer3DPickerSheet> {
  String _search = '';
  bool _isCreatingNew = false;
  String _selectedCustomerType = 'retail';

  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _emailController = TextEditingController();
  final _addressController = TextEditingController();

  final List<Map<String, dynamic>> _customerTypes = [
    {'id': 'retail', 'label': 'Retail', 'icon': Icons.shopping_bag_outlined, 'color': Color(0xFF38BDF8)},
    {'id': 'wholesale', 'label': 'Wholesale', 'icon': Icons.storefront_rounded, 'color': Color(0xFFA855F7)},
    {'id': 'vip', 'label': 'VIP Club', 'icon': Icons.stars_rounded, 'color': Color(0xFFF59E0B)},
    {'id': 'walk_in', 'label': 'Walk-in', 'icon': Icons.directions_walk_rounded, 'color': Color(0xFF10B981)},
  ];

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _emailController.dispose();
    _addressController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    const sheetBg = Color(0xFF0B0F19);
    const cardBg = Color(0xFF161E2E);
    const surfaceColor = Color(0xFF1E293B);
    const primarySky = Color(0xFF38BDF8);
    const primaryIndigo = Color(0xFF6366F1);
    const emeraldGreen = Color(0xFF10B981);

    final filtered = widget.customers.where((c) {
      final q = _search.toLowerCase();
      final name = (c['name'] ?? '').toString().toLowerCase();
      final phone = (c['phone'] ?? '').toString().toLowerCase();
      final code = (c['customer_code'] ?? '').toString().toLowerCase();
      return name.contains(q) || phone.contains(q) || code.contains(q);
    }).toList();

    return Align(
      alignment: Alignment.bottomCenter,
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: 660,
          maxHeight: MediaQuery.of(context).size.height * 0.88,
        ),
        child: Container(
          padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
          decoration: BoxDecoration(
            color: sheetBg,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
            border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.8),
                blurRadius: 30,
                offset: const Offset(0, -8),
              ),
            ],
          ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // 3D Drag handle
          Container(
            margin: const EdgeInsets.only(top: 10, bottom: 4),
            width: 48,
            height: 5,
            decoration: BoxDecoration(
              color: Colors.white24,
              borderRadius: BorderRadius.circular(10),
            ),
          ),

          // Header with Animated Toggle Pill
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 8, 16, 12),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(colors: [primaryIndigo, primarySky]),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Icon(
                        _isCreatingNew ? Icons.person_add_alt_1_rounded : Icons.people_alt_rounded,
                        color: Colors.white,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Text(
                      _isCreatingNew ? 'New Store Customer' : 'Select Customer',
                      style: GoogleFonts.inter(
                        fontSize: 16.5,
                        fontWeight: FontWeight.w800,
                        color: Colors.white,
                      ),
                    ),
                  ],
                ),

                // 3D Switch Pill
                Material(
                  color: Colors.transparent,
                  child: InkWell(
                    borderRadius: BorderRadius.circular(12),
                    onTap: () => setState(() => _isCreatingNew = !_isCreatingNew),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 7),
                      decoration: BoxDecoration(
                        color: surfaceColor,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: primarySky.withValues(alpha: 0.3)),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withValues(alpha: 0.3),
                            blurRadius: 6,
                            offset: const Offset(0, 2),
                          ),
                        ],
                      ),
                      child: Row(
                        children: [
                          Icon(
                            _isCreatingNew ? Icons.format_list_bulleted_rounded : Icons.add_rounded,
                            size: 16,
                            color: primarySky,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            _isCreatingNew ? 'Customer List' : 'Add Customer',
                            style: GoogleFonts.inter(
                              color: primarySky,
                              fontWeight: FontWeight.w800,
                              fontSize: 12,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const Divider(color: Colors.white10, height: 1),

          if (_isCreatingNew)
            // 3D Create Customer Form with Schema Fields
            Flexible(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'CUSTOMER TYPE (SCHEMA)',
                      style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.w800, color: Colors.white54),
                    ),
                    const SizedBox(height: 8),
                    // 3D Customer Type Selector
                    Row(
                      children: _customerTypes.map((t) {
                        final isSel = _selectedCustomerType == t['id'];
                        final tColor = t['color'] as Color;
                        return Expanded(
                          child: Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 3),
                            child: InkWell(
                              borderRadius: BorderRadius.circular(12),
                              onTap: () => setState(() => _selectedCustomerType = t['id']),
                              child: Container(
                                padding: const EdgeInsets.symmetric(vertical: 8),
                                decoration: BoxDecoration(
                                  color: isSel ? tColor.withValues(alpha: 0.2) : cardBg,
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: isSel ? tColor : Colors.white.withValues(alpha: 0.08),
                                    width: isSel ? 1.5 : 1,
                                  ),
                                  boxShadow: [
                                    if (isSel)
                                      BoxShadow(
                                        color: tColor.withValues(alpha: 0.25),
                                        blurRadius: 8,
                                        offset: const Offset(0, 3),
                                      ),
                                  ],
                                ),
                                child: Column(
                                  children: [
                                    Icon(t['icon'] as IconData, size: 18, color: isSel ? tColor : Colors.white54),
                                    const SizedBox(height: 4),
                                    Text(
                                      t['label'] as String,
                                      style: GoogleFonts.inter(
                                        fontSize: 11,
                                        fontWeight: isSel ? FontWeight.w800 : FontWeight.w500,
                                        color: isSel ? Colors.white : Colors.white60,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        );
                      }).toList(),
                    ),
                    const SizedBox(height: 16),

                    // Inputs
                    _build3DTextField(
                      controller: _nameController,
                      label: 'Full Name *',
                      icon: Icons.person_rounded,
                      cardBg: cardBg,
                      primarySky: primarySky,
                    ),
                    const SizedBox(height: 12),
                    _build3DTextField(
                      controller: _phoneController,
                      label: 'Phone Number',
                      icon: Icons.phone_rounded,
                      keyboardType: TextInputType.phone,
                      cardBg: cardBg,
                      primarySky: primarySky,
                    ),
                    const SizedBox(height: 12),
                    _build3DTextField(
                      controller: _emailController,
                      label: 'Email Address',
                      icon: Icons.email_rounded,
                      keyboardType: TextInputType.emailAddress,
                      cardBg: cardBg,
                      primarySky: primarySky,
                    ),
                    const SizedBox(height: 12),
                    _build3DTextField(
                      controller: _addressController,
                      label: 'Store Address / Notes',
                      icon: Icons.location_on_rounded,
                      cardBg: cardBg,
                      primarySky: primarySky,
                    ),
                    const SizedBox(height: 22),

                    // 3D Glowing Save CTA
                    SizedBox(
                      width: double.infinity,
                      height: 50,
                      child: ElevatedButton(
                        onPressed: () {
                          final name = _nameController.text.trim();
                          if (name.isEmpty) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Customer name is required')),
                            );
                            return;
                          }
                          final phone = _phoneController.text.trim();
                          final email = _emailController.text.trim();
                          final address = _addressController.text.trim();
                          final code = 'CUST-${DateTime.now().millisecondsSinceEpoch.toString().substring(7)}';
                          widget.onCreateNewCustomer(
                            name,
                            phone,
                            email,
                            code,
                            _selectedCustomerType,
                            address,
                          );
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: emeraldGreen,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                          elevation: 6,
                          shadowColor: emeraldGreen.withValues(alpha: 0.5),
                        ),
                        child: Text(
                          'Save & Select Customer',
                          style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w800, fontSize: 14),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            )
          else
            // 3D Search & Customer Cards
            Flexible(
              child: Column(
                children: [
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    child: Container(
                      height: 46,
                      decoration: BoxDecoration(
                        color: cardBg,
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                        boxShadow: [
                          BoxShadow(color: Colors.black.withValues(alpha: 0.3), blurRadius: 6, offset: const Offset(0, 3)),
                        ],
                      ),
                      child: TextField(
                        onChanged: (v) => setState(() => _search = v),
                        style: const TextStyle(color: Colors.white, fontSize: 13),
                        decoration: InputDecoration(
                          hintText: 'Search by name, phone or code...',
                          hintStyle: const TextStyle(color: Colors.white38),
                          prefixIcon: const Icon(Icons.search_rounded, color: primarySky, size: 18),
                          border: InputBorder.none,
                          contentPadding: const EdgeInsets.symmetric(vertical: 13),
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    child: filtered.isEmpty
                        ? Center(
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                const Icon(Icons.person_search_rounded, color: Colors.white24, size: 48),
                                const SizedBox(height: 8),
                                const Text('No customers found', style: TextStyle(color: Colors.white54)),
                                const SizedBox(height: 12),
                                ElevatedButton(
                                  onPressed: () => setState(() => _isCreatingNew = true),
                                  style: ElevatedButton.styleFrom(backgroundColor: primarySky),
                                  child: const Text('+ Register In-Store Customer', style: TextStyle(color: Colors.black87)),
                                ),
                              ],
                            ),
                          )
                        : ListView.separated(
                            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                            itemCount: filtered.length,
                            separatorBuilder: (_, __) => const SizedBox(height: 10),
                            itemBuilder: (context, idx) {
                              final cust = filtered[idx];
                              final isSelected = widget.selectedCustomer?['id'] == cust['id'];
                              final cType = (cust['customer_type'] ?? 'retail').toString().toUpperCase();
                              final pts = cust['loyalty_points'] ?? 0;

                              return InkWell(
                                onTap: () => widget.onSelectCustomer(cust),
                                borderRadius: BorderRadius.circular(16),
                                child: Container(
                                  padding: const EdgeInsets.all(12),
                                  decoration: BoxDecoration(
                                    color: isSelected ? emeraldGreen.withValues(alpha: 0.15) : cardBg,
                                    borderRadius: BorderRadius.circular(16),
                                    border: Border.all(
                                      color: isSelected ? emeraldGreen : Colors.white.withValues(alpha: 0.08),
                                      width: isSelected ? 1.5 : 1,
                                    ),
                                    boxShadow: [
                                      BoxShadow(
                                        color: isSelected
                                            ? emeraldGreen.withValues(alpha: 0.2)
                                            : Colors.black.withValues(alpha: 0.25),
                                        blurRadius: 8,
                                        offset: const Offset(0, 3),
                                      ),
                                    ],
                                  ),
                                  child: Row(
                                    children: [
                                      CircleAvatar(
                                        radius: 20,
                                        backgroundColor: primarySky.withValues(alpha: 0.2),
                                        child: Text(
                                          (cust['name'] ?? 'C')[0].toUpperCase(),
                                          style: const TextStyle(color: primarySky, fontWeight: FontWeight.w900),
                                        ),
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
                                                    cust['name'] ?? 'Customer',
                                                    style: GoogleFonts.inter(
                                                      fontWeight: FontWeight.w800,
                                                      color: Colors.white,
                                                      fontSize: 14,
                                                    ),
                                                  ),
                                                ),
                                                const SizedBox(width: 6),
                                                Container(
                                                  padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1.5),
                                                  decoration: BoxDecoration(
                                                    color: surfaceColor,
                                                    borderRadius: BorderRadius.circular(4),
                                                  ),
                                                  child: Text(
                                                    cType,
                                                    style: GoogleFonts.inter(
                                                      fontSize: 8.5,
                                                      fontWeight: FontWeight.w700,
                                                      color: primarySky,
                                                    ),
                                                  ),
                                                ),
                                              ],
                                            ),
                                            const SizedBox(height: 2),
                                            Text(
                                              'Code: ${cust['customer_code'] ?? 'N/A'} • ${cust['phone'] ?? 'No Phone'} • $pts pts',
                                              style: const TextStyle(color: Colors.white54, fontSize: 11.5),
                                            ),
                                          ],
                                        ),
                                      ),
                                      if (isSelected)
                                        Container(
                                          padding: const EdgeInsets.all(4),
                                          decoration: const BoxDecoration(
                                            color: emeraldGreen,
                                            shape: BoxShape.circle,
                                          ),
                                          child: const Icon(Icons.check, color: Colors.black, size: 14),
                                        ),
                                    ],
                                  ),
                                ),
                              );
                            },
                          ),
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

  Widget _build3DTextField({
    required TextEditingController controller,
    required String label,
    required IconData icon,
    required Color cardBg,
    required Color primarySky,
    TextInputType keyboardType = TextInputType.text,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: cardBg,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.25),
            blurRadius: 6,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: TextField(
        controller: controller,
        keyboardType: keyboardType,
        style: const TextStyle(color: Colors.white, fontSize: 13.5),
        decoration: InputDecoration(
          labelText: label,
          labelStyle: const TextStyle(color: Colors.white60, fontSize: 12.5),
          prefixIcon: Icon(icon, color: primarySky, size: 18),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        ),
      ),
    );
  }
}

/// 3D Animated Cart Review & Item Breakdown Sheet
class _Cart3DReviewSheet extends StatelessWidget {
  final List<Map<String, dynamic>> cartItems;
  final double subtotal;
  final double taxAmount;
  final double totalAmount;
  final Map<String, dynamic>? customer;
  final Function(int productId) onIncrement;
  final Function(int productId) onDecrement;
  final Function(int productId) onRemove;
  final VoidCallback onCheckout;

  const _Cart3DReviewSheet({
    required this.cartItems,
    required this.subtotal,
    required this.taxAmount,
    required this.totalAmount,
    required this.customer,
    required this.onIncrement,
    required this.onDecrement,
    required this.onRemove,
    required this.onCheckout,
  });

  @override
  Widget build(BuildContext context) {
    const sheetBg = Color(0xFF0B0F19);
    const cardBg = Color(0xFF161E2E);
    const surfaceColor = Color(0xFF1E293B);
    const primarySky = Color(0xFF38BDF8);
    const primaryIndigo = Color(0xFF6366F1);
    const emeraldGreen = Color(0xFF10B981);

    return Align(
      alignment: Alignment.bottomCenter,
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: 660,
          maxHeight: MediaQuery.of(context).size.height * 0.85,
        ),
        child: Container(
          decoration: BoxDecoration(
            color: sheetBg,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
            border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.8),
                blurRadius: 30,
                offset: const Offset(0, -8),
              ),
            ],
          ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Drag handle
          Container(
            margin: const EdgeInsets.only(top: 10, bottom: 4),
            width: 48,
            height: 5,
            decoration: BoxDecoration(
              color: Colors.white24,
              borderRadius: BorderRadius.circular(10),
            ),
          ),

          // Header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(colors: [primaryIndigo, primarySky]),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: const Icon(Icons.shopping_bag_rounded, color: Colors.white, size: 20),
                    ),
                    const SizedBox(width: 10),
                    Text(
                      'Cart Items (${cartItems.length})',
                      style: GoogleFonts.inter(
                        fontSize: 17,
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                        letterSpacing: -0.3,
                      ),
                    ),
                  ],
                ),
                IconButton(
                  icon: const Icon(Icons.close_rounded, color: Colors.white54),
                  onPressed: () => Navigator.of(context).pop(),
                ),
              ],
            ),
          ),
          const Divider(color: Colors.white10, height: 1),

          // Items List with 3D Stepper
          Flexible(
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: cartItems.length,
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemBuilder: (context, idx) {
                final item = cartItems[idx];
                final pid = (item['product_id'] as num).toInt();
                final pName = item['product_name'] ?? 'Product';
                final qty = (item['quantity'] as num).toInt();
                final unitPrice = (item['unit_price'] as num).toDouble();
                final lineTotal = (item['line_total'] as num).toDouble();

                return Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: cardBg,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.3),
                        blurRadius: 8,
                        offset: const Offset(0, 3),
                      ),
                    ],
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              pName,
                              style: GoogleFonts.inter(
                                fontWeight: FontWeight.w800,
                                color: Colors.white,
                                fontSize: 13.5,
                              ),
                            ),
                            const SizedBox(height: 3),
                            Wrap(
                              spacing: 6,
                              runSpacing: 4,
                              crossAxisAlignment: WrapCrossAlignment.center,
                              children: [
                                if (item['package_name'] != null && item['package_name'].toString().isNotEmpty)
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                    decoration: BoxDecoration(
                                      color: primaryIndigo.withValues(alpha: 0.25),
                                      borderRadius: BorderRadius.circular(6),
                                      border: Border.all(color: primaryIndigo.withValues(alpha: 0.5), width: 0.7),
                                    ),
                                    child: Text(
                                      item['package_name'],
                                      style: GoogleFonts.inter(color: primarySky, fontSize: 10, fontWeight: FontWeight.w700),
                                    ),
                                  ),
                                Text(
                                  'SAR ${unitPrice.toStringAsFixed(2)} / ${item['uom_code'] ?? 'PCS'}',
                                  style: const TextStyle(color: Colors.white70, fontSize: 11.5, fontWeight: FontWeight.w600),
                                ),
                              ],
                            ),
                            if (item['conversion_text'] != null && item['conversion_text'].toString().isNotEmpty) ...[
                              const SizedBox(height: 4),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2.5),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF0369A1).withValues(alpha: 0.18),
                                  borderRadius: BorderRadius.circular(6),
                                  border: Border.all(color: primarySky.withValues(alpha: 0.35), width: 0.7),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    const Icon(Icons.sync_alt_rounded, size: 11.5, color: primarySky),
                                    const SizedBox(width: 4),
                                    Text(
                                      'UOM: ${item['conversion_text']}',
                                      style: GoogleFonts.inter(color: primarySky, fontSize: 10, fontWeight: FontWeight.w700),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                            if (item['applied_promo'] != null) ...[
                              const SizedBox(height: 3),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                decoration: BoxDecoration(
                                  color: const Color(0xFFF59E0B).withValues(alpha: 0.18),
                                  borderRadius: BorderRadius.circular(6),
                                  border: Border.all(color: const Color(0xFFF59E0B).withValues(alpha: 0.35), width: 0.7),
                                ),
                                child: Text(
                                  '🏷️ ${item['applied_promo']}',
                                  style: const TextStyle(color: Color(0xFFF59E0B), fontSize: 10, fontWeight: FontWeight.w700),
                                ),
                              ),
                            ],
                            const SizedBox(height: 4),
                            Text(
                              'Line Total: SAR ${lineTotal.toStringAsFixed(2)}',
                              style: GoogleFonts.inter(
                                color: primarySky,
                                fontWeight: FontWeight.w800,
                                fontSize: 12.5,
                              ),
                            ),
                          ],
                        ),
                      ),

                      // 3D Tactile Stepper
                      Container(
                        decoration: BoxDecoration(
                          color: const Color(0xFF090D16),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                        ),
                        child: Row(
                          children: [
                            IconButton(
                              icon: const Icon(Icons.remove_rounded, color: Colors.white70, size: 16),
                              onPressed: () => onDecrement(pid),
                              constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                              padding: EdgeInsets.zero,
                            ),
                            Text(
                              '$qty',
                              style: GoogleFonts.inter(
                                fontWeight: FontWeight.w900,
                                color: Colors.white,
                                fontSize: 13.5,
                              ),
                            ),
                            IconButton(
                              icon: const Icon(Icons.add_rounded, color: primarySky, size: 16),
                              onPressed: () => onIncrement(pid),
                              constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                              padding: EdgeInsets.zero,
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 8),
                      IconButton(
                        icon: const Icon(Icons.delete_outline_rounded, color: Color(0xFFEF4444), size: 19),
                        onPressed: () => onRemove(pid),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),

          // 3D Receipt Summary Card
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
              color: cardBg,
              border: Border(top: BorderSide(color: Colors.white.withValues(alpha: 0.08))),
              boxShadow: [
                BoxShadow(color: Colors.black.withValues(alpha: 0.4), blurRadius: 10, offset: const Offset(0, -4)),
              ],
            ),
            child: Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Subtotal', style: TextStyle(color: Colors.white60, fontSize: 12.5)),
                    Text('SAR ${subtotal.toStringAsFixed(2)}', style: const TextStyle(color: Colors.white, fontSize: 12.5)),
                  ],
                ),
                const SizedBox(height: 6),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('VAT Tax (15%)', style: TextStyle(color: Colors.white60, fontSize: 12.5)),
                    Text('SAR ${taxAmount.toStringAsFixed(2)}', style: const TextStyle(color: Colors.white, fontSize: 12.5)),
                  ],
                ),
                const Divider(color: Colors.white10, height: 16),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Total Amount', style: TextStyle(color: Colors.white, fontSize: 14.5, fontWeight: FontWeight.bold)),
                    Text(
                      'SAR ${totalAmount.toStringAsFixed(2)}',
                      style: GoogleFonts.inter(
                        color: emeraldGreen,
                        fontSize: 20,
                        fontWeight: FontWeight.w900,
                        letterSpacing: -0.5,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  height: 50,
                  child: ElevatedButton(
                    onPressed: onCheckout,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: emeraldGreen,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                      elevation: 6,
                      shadowColor: emeraldGreen.withValues(alpha: 0.5),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.lock_outline_rounded, size: 18),
                        const SizedBox(width: 8),
                        Text(
                          'Proceed to Checkout (SAR ${totalAmount.toStringAsFixed(2)})',
                          style: GoogleFonts.inter(color: Colors.white, fontWeight: FontWeight.w800, fontSize: 14),
                        ),
                      ],
                    ),
                  ),
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

/// 3D Modal Bottom Sheet for Product Packaging, Pricing & UOM Conversion Reference
class _ProductUomAndPromotionBottomSheet extends StatefulWidget {
  final Map<String, dynamic> product;
  final int? activePriceListId;
  final Map<String, dynamic>? matchedPromotion;
  final List<Map<String, dynamic>> allPromotions;
  final String? initialSaleType;
  final String Function(int pid, int index) getProductAssetImage;
  final Function(int quantity, Map<String, dynamic> packageOption, double effectivePrice, String? promoDesc) onAddToCart;

  const _ProductUomAndPromotionBottomSheet({
    required this.product,
    this.activePriceListId,
    this.matchedPromotion,
    this.allPromotions = const [],
    this.initialSaleType,
    required this.getProductAssetImage,
    required this.onAddToCart,
  });

  @override
  State<_ProductUomAndPromotionBottomSheet> createState() => _ProductUomAndPromotionBottomSheetState();
}

class _ProductUomAndPromotionBottomSheetState extends State<_ProductUomAndPromotionBottomSheet> {
  int _quantity = 1;
  bool _isLoading = true;

  // Selectable packaging & pricing options
  List<Map<String, dynamic>> _packagingOptions = [];
  Map<String, dynamic>? _selectedPackage;

  // Non-selectable UOM conversion reference rates
  List<Map<String, dynamic>> _uomConversions = [];

  Map<String, dynamic>? _activePromotion;

  @override
  void initState() {
    super.initState();
    _loadPackagingAndUomConversions();
  }

  Future<void> _loadPackagingAndUomConversions() async {
    final pid = (widget.product['product_id'] as num?)?.toInt() ?? 0;
    final basePrice = double.tryParse(widget.product['price']?.toString() ?? '0') ?? 0.0;
    final baseUomCode = widget.product['uom_code']?.toString() ?? 'PCS';
    final baseUomName = widget.product['uom_name']?.toString() ?? 'Pieces';

    List<Map<String, dynamic>> packages = [];
    List<Map<String, dynamic>> conversions = [];

    // 1. Check if product already has package_n_price from Go core handler
    final rawPkgNPrice = widget.product['package_n_price'];
    dynamic parsedPkgNPrice = rawPkgNPrice;
    if (rawPkgNPrice is String && rawPkgNPrice.isNotEmpty) {
      try {
        parsedPkgNPrice = jsonDecode(rawPkgNPrice);
      } catch (_) {}
    }

    if (parsedPkgNPrice is List && parsedPkgNPrice.isNotEmpty) {
      for (final item in parsedPkgNPrice) {
        if (item is Map) {
          final plId = (item['price_list_id'] as num?)?.toInt();
          // Filter strictly by active price list if activePriceListId is provided
          if (widget.activePriceListId != null && plId != null && plId != widget.activePriceListId) {
            continue;
          }
          final pVal = (item['price'] as num?)?.toDouble() ?? 0.0;
          if (pVal <= 0) continue;

          final uCode = item['uom_code']?.toString() ?? item['uom']?.toString() ?? baseUomCode;
          final uName = item['uom_name']?.toString() ?? baseUomName;
          final minQ = (item['min_quantity'] as num?)?.toDouble() ?? 1.0;
          final isBase = minQ <= 1.0 && uCode.toUpperCase() == baseUomCode.toUpperCase();
          final convText = minQ > 1
              ? '1 $uCode = ${minQ.toInt()} $baseUomCode'
              : '1 $uCode = 1 $baseUomName';

          packages.add({
            'package_id': item['price_id'] ?? item['id'] ?? packages.length + 1,
            'package_name': isBase ? 'Single Unit ($baseUomCode)' : '$uName ($uCode - ${minQ.toInt()} $baseUomCode)',
            'uom_code': uCode,
            'uom_name': uName,
            'quantity_in_pack': minQ.toInt(),
            'conversion_factor': minQ,
            'price': pVal,
            'unit_price_equivalent': pVal / (minQ > 0 ? minQ : 1),
            'conversion_text': convText,
            'is_base': isBase,
          });
        }
      }
    }

    // 2. If no packages loaded from Go handler payload, query SQLite product_prices filtered by active price list
    if (packages.isEmpty && pid > 0) {
      try {
        final db = DatabaseService().database;
        String query;
        List<dynamic> args;
        if (widget.activePriceListId != null) {
          query = '''
            SELECT pp.id as price_id, pp.price, pp.min_quantity, pp.max_quantity, pp.price_list_id,
                   u.id as uom_id, u.code as uom_code, u.name as uom_name
            FROM product_prices pp
            LEFT JOIN units_of_measure u ON pp.uom_id = u.id
            WHERE pp.product_id = ? AND pp.price_list_id = ? AND (pp.is_active = 1 OR pp.is_active IS NULL)
            ORDER BY pp.min_quantity ASC
          ''';
          args = [pid, widget.activePriceListId];
        } else {
          query = '''
            SELECT pp.id as price_id, pp.price, pp.min_quantity, pp.max_quantity, pp.price_list_id,
                   u.id as uom_id, u.code as uom_code, u.name as uom_name
            FROM product_prices pp
            LEFT JOIN units_of_measure u ON pp.uom_id = u.id
            WHERE pp.product_id = ? AND (pp.is_active = 1 OR pp.is_active IS NULL)
            ORDER BY pp.min_quantity ASC
          ''';
          args = [pid];
        }

        final priceRows = await db.rawQuery(query, args);
        for (final pr in priceRows) {
          final pVal = (pr['price'] as num?)?.toDouble();
          if (pVal == null || pVal <= 0) continue;
          final uCode = pr['uom_code']?.toString() ?? baseUomCode;
          final uName = pr['uom_name']?.toString() ?? baseUomName;
          final minQ = (pr['min_quantity'] as num?)?.toDouble() ?? 1.0;
          final isBase = minQ <= 1.0 && uCode.toUpperCase() == baseUomCode.toUpperCase();
          final convText = minQ > 1
              ? '1 $uCode = ${minQ.toInt()} $baseUomCode'
              : '1 $uCode = 1 $baseUomName';

          packages.add({
            'package_id': pr['price_id'],
            'package_name': isBase ? 'Single Unit ($baseUomCode)' : '$uName ($uCode - ${minQ.toInt()} $baseUomCode)',
            'uom_code': uCode,
            'uom_name': uName,
            'quantity_in_pack': minQ.toInt(),
            'conversion_factor': minQ,
            'price': pVal,
            'unit_price_equivalent': pVal / (minQ > 0 ? minQ : 1),
            'conversion_text': convText,
            'is_base': isBase,
          });
        }
      } catch (e) {
        developer.log('Error querying packages: $e', name: 'ProductPackaging');
      }
    }

    // 3. Fallback: If no distinct package rows exist, strictly use the product's base price
    if (packages.isEmpty) {
      packages.add({
        'package_id': 1,
        'package_name': 'Single Unit ($baseUomCode)',
        'uom_code': baseUomCode,
        'uom_name': baseUomName,
        'quantity_in_pack': 1,
        'conversion_factor': 1.0,
        'price': basePrice,
        'unit_price_equivalent': basePrice,
        'conversion_text': '1 $baseUomCode = 1 $baseUomName',
        'is_base': true,
      });
    }

    // 4. Load UOM conversions from handler or SQLite table (authentic data only)
    final rawUomConversions = widget.product['product_uom_conversions'];
    dynamic parsedUomConversions = rawUomConversions;
    if (rawUomConversions is String && rawUomConversions.isNotEmpty) {
      try {
        parsedUomConversions = jsonDecode(rawUomConversions);
      } catch (_) {}
    }

    if (parsedUomConversions is List && parsedUomConversions.isNotEmpty) {
      for (final item in parsedUomConversions) {
        if (item is Map) {
          final factor = (item['conversion_factor'] as num?)?.toDouble() ?? 1.0;
          final fCode = item['from_uom_code']?.toString() ?? item['from_code']?.toString() ?? 'UOM';
          final fName = item['from_uom_name']?.toString() ?? item['from_name']?.toString() ?? fCode;
          final tCode = item['to_uom_code']?.toString() ?? item['to_code']?.toString() ?? baseUomCode;
          final tName = item['to_uom_name']?.toString() ?? item['to_name']?.toString() ?? baseUomName;
          final convText = '1 $fCode = ${factor.toStringAsFixed(factor.truncateToDouble() == factor ? 0 : 2)} $tCode';

          conversions.add({
            'from_code': fCode,
            'from_name': fName,
            'to_code': tCode,
            'to_name': tName,
            'factor': factor,
            'conversion_text': convText,
          });
        }
      }
    }

    if (conversions.isEmpty && pid > 0) {
      try {
        final db = DatabaseService().database;
        final convRows = await db.rawQuery('''
          SELECT puc.id, puc.conversion_factor, puc.is_default,
                 fu.code as from_code, fu.name as from_name,
                 tu.code as to_code, tu.name as to_name
          FROM product_uom_conversions puc
          JOIN units_of_measure fu ON puc.from_uom_id = fu.id
          JOIN units_of_measure tu ON puc.to_uom_id = tu.id
          WHERE puc.product_id = ?
        ''', [pid]);

        for (final r in convRows) {
          final factor = (r['conversion_factor'] as num?)?.toDouble() ?? 1.0;
          final fCode = r['from_code']?.toString() ?? 'UOM';
          final fName = r['from_name']?.toString() ?? fCode;
          final tCode = r['to_code']?.toString() ?? baseUomCode;
          final tName = r['to_name']?.toString() ?? baseUomName;
          final convText = '1 $fCode = ${factor.toStringAsFixed(factor.truncateToDouble() == factor ? 0 : 2)} $tCode';

          conversions.add({
            'from_code': fCode,
            'from_name': fName,
            'to_code': tCode,
            'to_name': tName,
            'factor': factor,
            'conversion_text': convText,
          });
        }
      } catch (e) {
        developer.log('Error querying conversions: $e', name: 'ProductPackaging');
      }
    }

    // Match promotion
    Map<String, dynamic>? promo = widget.matchedPromotion;
    if (promo == null && widget.allPromotions.isNotEmpty) {
      final catId = (widget.product['category_id'] as num?)?.toInt();
      for (final p in widget.allPromotions) {
        final appliesTo = (p['applies_to'] ?? 'all').toString().toLowerCase();
        if (appliesTo == 'all') {
          promo = p;
          break;
        }
        if (appliesTo == 'product' && p['target_product_ids'] != null && p['target_product_ids'].toString().contains('$pid')) {
          promo = p;
          break;
        }
        if (appliesTo == 'category' && catId != null && p['target_category_ids'] != null && p['target_category_ids'].toString().contains('$catId')) {
          promo = p;
          break;
        }
      }
    }

    if (mounted) {
      setState(() {
        _packagingOptions = packages;
        _selectedPackage = packages.first;
        _uomConversions = conversions;
        _activePromotion = promo;
        _isLoading = false;
      });
    }
  }

  double get _currentPackagePrice => (_selectedPackage?['price'] as num?)?.toDouble() ?? 0.0;

  double get _discountPerUnit {
    if (_activePromotion == null) return 0.0;
    final pType = (_activePromotion!['promotion_type'] ?? 'percentage_discount').toString();
    final dVal = double.tryParse(_activePromotion!['discount_value']?.toString() ?? '0') ?? 0.0;
    if (pType == 'percentage_discount') {
      return _currentPackagePrice * (dVal / 100.0);
    } else if (pType == 'fixed_discount') {
      return math.min(dVal, _currentPackagePrice);
    }
    return 0.0;
  }

  double get _effectiveUnitPrice => math.max(0.0, _currentPackagePrice - _discountPerUnit);
  double get _totalLineAmount => _effectiveUnitPrice * _quantity;

  @override
  Widget build(BuildContext context) {
    const sheetBg = Color(0xFF0B0F19);
    const cardBg = Color(0xFF161E2E);
    const surfaceColor = Color(0xFF1E293B);
    const primarySky = Color(0xFF38BDF8);
    const primaryIndigo = Color(0xFF6366F1);
    const emeraldGreen = Color(0xFF10B981);
    const amberGold = Color(0xFFF59E0B);

    final pid = (widget.product['product_id'] as num?)?.toInt() ?? 1;
    final pName = widget.product['product_name'] ?? 'Product #$pid';
    final pSku = widget.product['sku'] ?? 'SKU-$pid';
    final pCat = widget.product['category_name'] ?? 'General';
    final stockQty = (widget.product['quantity_available'] as num?)?.toInt() ?? 0;
    final isWholesale = widget.initialSaleType?.toLowerCase() == 'wholesale';
    final isOutOfStock = isWholesale && stockQty <= 0;
    final assetImg = widget.getProductAssetImage(pid, 0);

    return Align(
      alignment: Alignment.bottomCenter,
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: 680,
          maxHeight: MediaQuery.of(context).size.height * 0.88,
        ),
        child: Container(
          decoration: BoxDecoration(
            color: sheetBg,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
            border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.85),
                blurRadius: 32,
                offset: const Offset(0, -8),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Drag Handle
              Container(
                margin: const EdgeInsets.only(top: 10, bottom: 6),
                width: 44,
                height: 5,
                decoration: BoxDecoration(
                  color: Colors.white24,
                  borderRadius: BorderRadius.circular(10),
                ),
              ),

              // Header Bar
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            gradient: const LinearGradient(colors: [primaryIndigo, primarySky]),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.inventory_2_rounded, color: Colors.white, size: 20),
                        ),
                        const SizedBox(width: 10),
                        Text(
                          'Packaging, Pricing & UOM',
                          style: GoogleFonts.inter(
                            fontSize: 17,
                            fontWeight: FontWeight.w900,
                            color: Colors.white,
                            letterSpacing: -0.3,
                          ),
                        ),
                      ],
                    ),
                    IconButton(
                      icon: const Icon(Icons.close_rounded, color: Colors.white54),
                      onPressed: () => Navigator.of(context).pop(),
                    ),
                  ],
                ),
              ),
              const Divider(color: Colors.white10, height: 1),

              // Body Content
              Flexible(
                child: _isLoading
                    ? const Center(child: Padding(padding: EdgeInsets.all(32), child: CircularProgressIndicator(color: primarySky)))
                    : SingleChildScrollView(
                        physics: const BouncingScrollPhysics(),
                        padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 14),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Product Hero Overview Card
                            Container(
                              padding: const EdgeInsets.all(14),
                              decoration: BoxDecoration(
                                color: cardBg,
                                borderRadius: BorderRadius.circular(18),
                                border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black.withValues(alpha: 0.35),
                                    blurRadius: 10,
                                    offset: const Offset(0, 4),
                                  ),
                                ],
                              ),
                              child: Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  ClipRRect(
                                    borderRadius: BorderRadius.circular(12),
                                    child: Container(
                                      width: 68,
                                      height: 68,
                                      color: surfaceColor,
                                      child: assetImg.isNotEmpty
                                          ? Image.asset(assetImg, fit: BoxFit.cover)
                                          : const Icon(Icons.inventory_2_outlined, color: Colors.white38, size: 32),
                                    ),
                                  ),
                                  const SizedBox(width: 14),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          pSku,
                                          style: GoogleFonts.inter(fontSize: 10.5, color: Colors.white38, fontWeight: FontWeight.w700),
                                        ),
                                        const SizedBox(height: 2),
                                        Text(
                                          pName,
                                          style: GoogleFonts.inter(fontSize: 15, fontWeight: FontWeight.w800, color: Colors.white),
                                        ),
                                        const SizedBox(height: 6),
                                        Row(
                                          children: [
                                            Container(
                                              padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2.5),
                                              decoration: BoxDecoration(
                                                color: primaryIndigo.withValues(alpha: 0.18),
                                                borderRadius: BorderRadius.circular(6),
                                                border: Border.all(color: primaryIndigo.withValues(alpha: 0.4), width: 0.7),
                                              ),
                                              child: Text(
                                                pCat,
                                                style: GoogleFonts.inter(fontSize: 9.5, color: primarySky, fontWeight: FontWeight.w700),
                                              ),
                                            ),
                                            const SizedBox(width: 8),
                                            Container(
                                              padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2.5),
                                              decoration: BoxDecoration(
                                                color: (isOutOfStock ? const Color(0xFFEF4444) : emeraldGreen).withValues(alpha: 0.18),
                                                borderRadius: BorderRadius.circular(6),
                                                border: Border.all(
                                                  color: (isOutOfStock ? const Color(0xFFEF4444) : emeraldGreen).withValues(alpha: 0.4),
                                                  width: 0.7,
                                                ),
                                              ),
                                              child: Row(
                                                mainAxisSize: MainAxisSize.min,
                                                children: [
                                                  Icon(
                                                    isOutOfStock ? Icons.block_rounded : Icons.check_circle_outline_rounded,
                                                    size: 10,
                                                    color: isOutOfStock ? const Color(0xFFEF4444) : emeraldGreen,
                                                  ),
                                                  const SizedBox(width: 4),
                                                  Text(
                                                    isOutOfStock ? 'Out of Stock' : (stockQty > 0 ? 'Stock: $stockQty' : 'In Stock'),
                                                    style: GoogleFonts.inter(
                                                      fontSize: 9.5,
                                                      fontWeight: FontWeight.w800,
                                                      color: isOutOfStock ? const Color(0xFFEF4444) : emeraldGreen,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 16),

                            // SECTION 1: Select Packaging & Pricing (SELECTABLE)
                            Row(
                              children: [
                                const Icon(Icons.all_inbox_rounded, color: primarySky, size: 19),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    'Select Package & Pricing',
                                    style: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w800, color: Colors.white),
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                                const SizedBox(width: 6),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: primarySky.withValues(alpha: 0.15),
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                  child: const Text(
                                    'Selectable',
                                    style: TextStyle(color: primarySky, fontSize: 9.5, fontWeight: FontWeight.w800),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            ListView.separated(
                              shrinkWrap: true,
                              physics: const NeverScrollableScrollPhysics(),
                              itemCount: _packagingOptions.length,
                              separatorBuilder: (_, __) => const SizedBox(height: 8),
                              itemBuilder: (context, idx) {
                                final pkg = _packagingOptions[idx];
                                final isSelected = _selectedPackage?['package_id'] == pkg['package_id'] &&
                                    _selectedPackage?['package_name'] == pkg['package_name'];
                                final pkgName = pkg['package_name'] ?? 'Package';
                                final uomCode = pkg['uom_code'] ?? 'PCS';
                                final pkgPrice = (pkg['price'] as num?)?.toDouble() ?? 0.0;
                                final unitEq = (pkg['unit_price_equivalent'] as num?)?.toDouble();
                                final savings = pkg['savings_badge']?.toString();

                                return InkWell(
                                  onTap: () {
                                    setState(() => _selectedPackage = pkg);
                                  },
                                  borderRadius: BorderRadius.circular(14),
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                                    decoration: BoxDecoration(
                                      color: isSelected ? primarySky.withValues(alpha: 0.12) : cardBg,
                                      borderRadius: BorderRadius.circular(14),
                                      border: Border.all(
                                        color: isSelected ? primarySky : Colors.white.withValues(alpha: 0.08),
                                        width: isSelected ? 1.8 : 1,
                                      ),
                                      boxShadow: [
                                        BoxShadow(
                                          color: isSelected ? primarySky.withValues(alpha: 0.22) : Colors.black.withValues(alpha: 0.25),
                                          blurRadius: 8,
                                          offset: const Offset(0, 2),
                                        ),
                                      ],
                                    ),
                                    child: Row(
                                      children: [
                                        Container(
                                          padding: const EdgeInsets.all(8),
                                          decoration: BoxDecoration(
                                            gradient: isSelected
                                                ? const LinearGradient(colors: [primaryIndigo, primarySky])
                                                : const LinearGradient(colors: [Color(0xFF334155), Color(0xFF1E293B)]),
                                            borderRadius: BorderRadius.circular(10),
                                          ),
                                          child: Icon(
                                            pkg['is_base'] == true ? Icons.inventory_rounded : Icons.card_giftcard_rounded,
                                            color: Colors.white,
                                            size: 18,
                                          ),
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
                                                      pkgName,
                                                      style: GoogleFonts.inter(
                                                        color: Colors.white,
                                                        fontWeight: FontWeight.w800,
                                                        fontSize: 13.5,
                                                      ),
                                                    ),
                                                  ),
                                                  if (savings != null) ...[
                                                    const SizedBox(width: 6),
                                                    Container(
                                                      padding: const EdgeInsets.symmetric(horizontal: 5.5, vertical: 1.5),
                                                      decoration: BoxDecoration(
                                                        color: emeraldGreen.withValues(alpha: 0.2),
                                                        borderRadius: BorderRadius.circular(5),
                                                        border: Border.all(color: emeraldGreen.withValues(alpha: 0.4), width: 0.6),
                                                      ),
                                                      child: Text(
                                                        savings,
                                                        style: const TextStyle(
                                                          color: emeraldGreen,
                                                          fontSize: 9,
                                                          fontWeight: FontWeight.w900,
                                                        ),
                                                      ),
                                                    ),
                                                  ],
                                                ],
                                              ),
                                              const SizedBox(height: 2),
                                              Text(
                                                pkg['conversion_text'] ?? '1 $uomCode',
                                                style: const TextStyle(color: Colors.white54, fontSize: 11),
                                              ),
                                              if (unitEq != null && pkg['is_base'] != true)
                                                Text(
                                                  'Equivalent: SAR ${unitEq.toStringAsFixed(2)} / unit',
                                                  style: const TextStyle(color: primarySky, fontSize: 10.5, fontWeight: FontWeight.w600),
                                                ),
                                            ],
                                          ),
                                        ),
                                        Column(
                                          crossAxisAlignment: CrossAxisAlignment.end,
                                          children: [
                                            Text(
                                              'SAR ${pkgPrice.toStringAsFixed(2)}',
                                              style: GoogleFonts.inter(
                                                color: isSelected ? primarySky : Colors.white,
                                                fontWeight: FontWeight.w900,
                                                fontSize: 15,
                                              ),
                                            ),
                                            Text(
                                              'per package',
                                              style: const TextStyle(color: Colors.white38, fontSize: 9.5),
                                            ),
                                          ],
                                        ),
                                        const SizedBox(width: 10),
                                        Icon(
                                          isSelected ? Icons.check_circle_rounded : Icons.radio_button_unchecked_rounded,
                                          color: isSelected ? primarySky : Colors.white24,
                                          size: 22,
                                        ),
                                      ],
                                    ),
                                  ),
                                );
                              },
                            ),
                            const SizedBox(height: 18),

                            // SECTION 2: UOM Conversion Rates (NON-SELECTABLE / REFERENCE ONLY)
                            Row(
                              children: [
                                const Icon(Icons.sync_alt_rounded, color: primarySky, size: 18),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    'UOM Conversion Reference',
                                    style: GoogleFonts.inter(fontSize: 13.5, fontWeight: FontWeight.w800, color: Colors.white),
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: const Color(0xFF334155).withValues(alpha: 0.6),
                                    borderRadius: BorderRadius.circular(6),
                                    border: Border.all(color: Colors.white12, width: 0.7),
                                  ),
                                  child: const Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Icon(Icons.lock_outline_rounded, color: Colors.white60, size: 10),
                                      SizedBox(width: 4),
                                      Text(
                                        'Reference',
                                        style: TextStyle(color: Colors.white60, fontSize: 9.5, fontWeight: FontWeight.w700),
                                      ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: const Color(0xFF0F172A),
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  if (_uomConversions.isNotEmpty)
                                    Wrap(
                                      spacing: 8,
                                      runSpacing: 8,
                                      children: _uomConversions.map((conv) {
                                        return Container(
                                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                                          decoration: BoxDecoration(
                                            color: cardBg,
                                            borderRadius: BorderRadius.circular(10),
                                            border: Border.all(color: primarySky.withValues(alpha: 0.25), width: 0.8),
                                          ),
                                          child: Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1.5),
                                                decoration: BoxDecoration(
                                                  color: primaryIndigo.withValues(alpha: 0.3),
                                                  borderRadius: BorderRadius.circular(4),
                                                ),
                                                child: Text(
                                                  conv['from_code'] ?? 'UOM',
                                                  style: const TextStyle(color: primarySky, fontSize: 10.5, fontWeight: FontWeight.w800),
                                                ),
                                              ),
                                              const SizedBox(width: 6),
                                              const Icon(Icons.arrow_forward_rounded, size: 12, color: Colors.white38),
                                              const SizedBox(width: 6),
                                              Text(
                                                '${conv['factor']} ${conv['to_code'] ?? 'PCS'}',
                                                style: GoogleFonts.inter(
                                                  color: Colors.white,
                                                  fontSize: 11.5,
                                                  fontWeight: FontWeight.w700,
                                                ),
                                              ),
                                            ],
                                          ),
                                        );
                                      }).toList(),
                                    )
                                  else
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                                      decoration: BoxDecoration(
                                        color: cardBg,
                                        borderRadius: BorderRadius.circular(10),
                                      ),
                                      child: Row(
                                        children: [
                                          const Icon(Icons.info_outline_rounded, color: primarySky, size: 14),
                                          const SizedBox(width: 8),
                                          Expanded(
                                            child: Text(
                                              'Base Unit: 1 ${(widget.product['uom_code'] ?? 'PCS')} (${(widget.product['uom_name'] ?? 'Pieces')}). No additional conversions configured.',
                                              style: GoogleFonts.inter(color: Colors.white70, fontSize: 11),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  const SizedBox(height: 8),
                                  Row(
                                    children: [
                                      const Icon(Icons.info_outline_rounded, color: Colors.white38, size: 13),
                                      const SizedBox(width: 6),
                                      Expanded(
                                        child: Text(
                                          'Conversion rates are for standard inventory calculation. Select package above to add to cart.',
                                          style: GoogleFonts.inter(color: Colors.white38, fontSize: 10),
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 18),

                            // SECTION 3: Promotions & Offers
                            Row(
                              children: [
                                const Icon(Icons.local_offer_rounded, color: amberGold, size: 18),
                                const SizedBox(width: 8),
                                Text(
                                  'Product Promotions & Discounts',
                                  style: GoogleFonts.inter(fontSize: 13.5, fontWeight: FontWeight.w800, color: Colors.white),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            if (_activePromotion != null)
                              Container(
                                padding: const EdgeInsets.all(14),
                                decoration: BoxDecoration(
                                  gradient: LinearGradient(
                                    colors: [
                                      amberGold.withValues(alpha: 0.18),
                                      primaryIndigo.withValues(alpha: 0.12),
                                    ],
                                    begin: Alignment.topLeft,
                                    end: Alignment.bottomRight,
                                  ),
                                  borderRadius: BorderRadius.circular(16),
                                  border: Border.all(color: amberGold.withValues(alpha: 0.45), width: 1.2),
                                  boxShadow: [
                                    BoxShadow(
                                      color: amberGold.withValues(alpha: 0.15),
                                      blurRadius: 10,
                                      offset: const Offset(0, 3),
                                    ),
                                  ],
                                ),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      children: [
                                        Container(
                                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                          decoration: BoxDecoration(
                                            color: amberGold.withValues(alpha: 0.25),
                                            borderRadius: BorderRadius.circular(6),
                                            border: Border.all(color: amberGold, width: 0.8),
                                          ),
                                          child: Text(
                                            (_activePromotion!['promotion_type'] ?? 'PROMOTION').toString().replaceAll('_', ' ').toUpperCase(),
                                            style: GoogleFonts.inter(
                                              fontSize: 9.5,
                                              fontWeight: FontWeight.w900,
                                              color: amberGold,
                                              letterSpacing: 0.5,
                                            ),
                                          ),
                                        ),
                                        if (_activePromotion!['coupon_code'] != null || _activePromotion!['code'] != null)
                                          Container(
                                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2.5),
                                            decoration: BoxDecoration(
                                              color: Colors.white10,
                                              borderRadius: BorderRadius.circular(6),
                                            ),
                                            child: Text(
                                              'Code: ${_activePromotion!['coupon_code'] ?? _activePromotion!['code']}',
                                              style: GoogleFonts.inter(fontSize: 10, fontWeight: FontWeight.w700, color: Colors.white70),
                                            ),
                                          ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    Text(
                                      _activePromotion!['name']?.toString() ?? 'Special Promotional Discount',
                                      style: GoogleFonts.inter(
                                        color: Colors.white,
                                        fontWeight: FontWeight.w900,
                                        fontSize: 14,
                                      ),
                                    ),
                                    const SizedBox(height: 3),
                                    Text(
                                      _activePromotion!['description']?.toString() ?? 'Automatic discount applied to eligible items in terminal cart.',
                                      style: const TextStyle(color: Colors.white70, fontSize: 11.5),
                                    ),
                                    if (_discountPerUnit > 0) ...[
                                      const SizedBox(height: 10),
                                      Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                                        decoration: BoxDecoration(
                                          color: emeraldGreen.withValues(alpha: 0.2),
                                          borderRadius: BorderRadius.circular(8),
                                        ),
                                        child: Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            const Icon(Icons.savings_rounded, color: emeraldGreen, size: 15),
                                            const SizedBox(width: 6),
                                            Text(
                                              'Discount Applied: -SAR ${_discountPerUnit.toStringAsFixed(2)} per package',
                                              style: GoogleFonts.inter(
                                                color: emeraldGreen,
                                                fontWeight: FontWeight.w800,
                                                fontSize: 11.5,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              )
                            else
                              Container(
                                padding: const EdgeInsets.all(12),
                                decoration: BoxDecoration(
                                  color: cardBg,
                                  borderRadius: BorderRadius.circular(14),
                                  border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                                ),
                                child: Row(
                                  children: [
                                    const Icon(Icons.info_outline_rounded, color: Colors.white38, size: 18),
                                    const SizedBox(width: 10),
                                    Expanded(
                                      child: Text(
                                        'No active promotions currently configured for this product. Standard retail pricing applies.',
                                        style: GoogleFonts.inter(color: Colors.white54, fontSize: 11.5),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                          ],
                        ),
                      ),
              ),

              // Bottom Pricing & Add-to-Cart Action Bar
              Container(
                padding: const EdgeInsets.fromLTRB(18, 14, 18, 18),
                decoration: BoxDecoration(
                  color: cardBg,
                  border: Border(top: BorderSide(color: Colors.white.withValues(alpha: 0.1))),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.4),
                      blurRadius: 10,
                      offset: const Offset(0, -4),
                    ),
                  ],
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        // Quantity Stepper
                        Container(
                          decoration: BoxDecoration(
                            color: const Color(0xFF090D16),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
                          ),
                          child: Row(
                            children: [
                              IconButton(
                                icon: const Icon(Icons.remove_rounded, color: Colors.white70, size: 18),
                                onPressed: _quantity > 1 ? () => setState(() => _quantity--) : null,
                                constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                                padding: EdgeInsets.zero,
                              ),
                              Padding(
                                padding: const EdgeInsets.symmetric(horizontal: 8),
                                child: Text(
                                  '$_quantity',
                                  style: GoogleFonts.inter(
                                    fontWeight: FontWeight.w900,
                                    color: Colors.white,
                                    fontSize: 15,
                                  ),
                                ),
                              ),
                              IconButton(
                                icon: const Icon(Icons.add_rounded, color: primarySky, size: 18),
                                onPressed: () => setState(() => _quantity++),
                                constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                                padding: EdgeInsets.zero,
                              ),
                            ],
                          ),
                        ),

                        // Pricing calculation breakdown
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            if (_discountPerUnit > 0)
                              Text(
                                'SAR ${(_currentPackagePrice * _quantity).toStringAsFixed(2)}',
                                style: const TextStyle(
                                  color: Colors.white38,
                                  fontSize: 12,
                                  decoration: TextDecoration.lineThrough,
                                ),
                              ),
                            Text(
                              'SAR ${_totalLineAmount.toStringAsFixed(2)}',
                              style: GoogleFonts.inter(
                                color: emeraldGreen,
                                fontSize: 20,
                                fontWeight: FontWeight.w900,
                                letterSpacing: -0.5,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),

                    // Add To Cart Button
                    SizedBox(
                      width: double.infinity,
                      height: 50,
                      child: ElevatedButton(
                        onPressed: isOutOfStock
                            ? null
                            : () {
                                final pkg = _selectedPackage ??
                                    (_packagingOptions.isNotEmpty
                                        ? _packagingOptions.first
                                        : {
                                            'package_name': 'Single Unit',
                                            'uom_code': 'PCS',
                                            'uom_name': 'Pieces',
                                            'conversion_factor': 1.0,
                                            'price': _currentPackagePrice,
                                            'conversion_text': '1 PCS = 1 Pieces',
                                          });
                                widget.onAddToCart(
                                  _quantity,
                                  pkg,
                                  _effectiveUnitPrice,
                                  _activePromotion?['name']?.toString(),
                                );
                                Navigator.of(context).pop();
                              },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: emeraldGreen,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                          elevation: 6,
                          shadowColor: emeraldGreen.withValues(alpha: 0.5),
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.add_shopping_cart_rounded, size: 19),
                            const SizedBox(width: 8),
                            Text(
                              'Add to Cart • SAR ${_totalLineAmount.toStringAsFixed(2)}',
                              style: GoogleFonts.inter(
                                color: Colors.white,
                                fontWeight: FontWeight.w800,
                                fontSize: 14.5,
                              ),
                            ),
                          ],
                        ),
                      ),
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

