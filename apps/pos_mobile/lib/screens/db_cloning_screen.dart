import 'dart:convert';
import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:http/http.dart' as http;
import 'package:pos_mobile/screens/store_selection_screen.dart';
import 'package:sqflite/sqflite.dart';
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
        _errorMessage = null;
      });
      await DatabaseService().purgeDatabase();

      setState(() {
        _currentPhase = SyncPhase.migrating;
        _progress = 0.40;
        _statusMessage = 'Applying latest mobile database schema...';
      });
      await MigrationRunner.applyMigrations(
        onProgress: (step, prog) {
          if (mounted) {
            setState(() {
              _progress = 0.40 + (prog * 0.25);
            });
          }
        },
      );

      final dbPath = await DatabaseService().getDatabasePath();
      final initRes = NembusBridge().initMobile(dbPath);
      developer.log('Go Engine Init Result: $initRes', name: 'DbCloningScreen');
      if (initRes['success'] == false) {
        throw Exception('Native database initialization failed: ${initRes['error']}');
      }

      setState(() {
        _currentPhase = SyncPhase.syncingData;
        _progress = 0.70;
        _statusMessage = 'Cloning tenant master data snapshot...';
      });

      // Allow UI thread frame to render progress before heavy FFI call
      await Future.delayed(const Duration(milliseconds: 100));

      final targetUrl = SingletonClass().baseURL.isNotEmpty
          ? SingletonClass().baseURL
          : SingletonClass().grpcAddress;
      final targetSlug = widget.tenantSlug.isNotEmpty ? widget.tenantSlug : widget.tenantId;

      final syncResult = await Future(() => NembusBridge().fetchCompleteTenantMasterData(
        targetSlug,
        targetUrl,
        dbPath: dbPath,
      ));

      developer.log('FFI Sync Response: $syncResult', name: 'DbCloningScreen');

      if (syncResult['success'] != true) {
        throw Exception(syncResult['error'] ?? 'Database synchronization failed');
      }

      var recordsSynced = (syncResult['data']?['records_synced'] as num?)?.toInt() ?? 0;
      if (recordsSynced == 0) {
        throw Exception('Synchronization finished but 0 records were cloned for tenant "${widget.tenantName}". Please verify tenant configuration and network connection.');
      }

      // Step 3.5: Verify that the users table contains user records for cashier logins.
      // If users table is empty (e.g. older native FFI build or constraint issue),
      // perform a direct Dart HTTP sync of users and cashier permissions.
      final localUserCount = await _ensureUsersCloned(targetSlug, targetUrl);
      developer.log('Verified local users count: $localUserCount', name: 'DbCloningScreen');
      if (localUserCount > 0) {
        recordsSynced += localUserCount;
      }

      // Step 3.6: Verify that the products table contains records.
      // If products table is empty (e.g. constraint issue, missing organization_id, or older native binary),
      // perform a direct Dart HTTP sync of products, variants, barcodes, prices, and stock.
      final localProductCount = await _ensureProductsCloned(targetSlug, targetUrl);
      developer.log('Verified local products count: $localProductCount', name: 'DbCloningScreen');
      if (localProductCount > 0) {
        recordsSynced += localProductCount;
      }

      // Step 4: Finalize & Transition
      setState(() {
        _currentPhase = SyncPhase.completed;
        _progress = 1.0;
        _statusMessage = 'Terminal initialized with $recordsSynced records.';
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
              isFirstInstall: true,
            ),
          ),
        );
      }
    } catch (e, stack) {
      developer.log('DB Cloning Error: $e', name: 'DbCloningScreen', error: e, stackTrace: stack);
      if (mounted) {
        setState(() {
          _currentPhase = SyncPhase.failed;
          _errorMessage = e.toString().replaceAll('Exception: ', '');
          _statusMessage = 'Synchronization failed. Please retry.';
        });
      }
    }
  }

  /// Checks local SQLite for cloned users. If empty or missing password hashes, directly fetches users from cloud REST API and inserts them.
  Future<int> _ensureUsersCloned(String tenantSlug, String rawUrl) async {
    try {
      final db = DatabaseService().database;
      final usersWithHash = await db.rawQuery(
        "SELECT COUNT(*) AS cnt FROM users WHERE password_hash IS NOT NULL AND password_hash != ''",
      );
      final hashCount = Sqflite.firstIntValue(usersWithHash) ?? 0;
      final totalUsers = await db.rawQuery('SELECT COUNT(*) AS cnt FROM users');
      final totalCount = Sqflite.firstIntValue(totalUsers) ?? 0;

      if (totalCount > 0 && hashCount > 0) {
        developer.log('SQLite already has $totalCount users with valid password hashes cloned.', name: 'DbCloningScreen');
        return totalCount;
      }

      developer.log('SQLite has $totalCount users ($hashCount with hash). Performing direct user sync to fetch hashes...', name: 'DbCloningScreen');
      if (mounted) {
        setState(() {
          _statusMessage = 'Fetching and caching user accounts with credentials...';
        });
      }

      // Resolve base URL
      String baseUrl = rawUrl.trim();
      if (baseUrl.startsWith('grpc://')) {
        baseUrl = baseUrl.replaceFirst('grpc://', 'http://');
      } else if (baseUrl.startsWith('grpcs://')) {
        baseUrl = baseUrl.replaceFirst('grpcs://', 'https://');
      }
      if (!baseUrl.startsWith('http://') && !baseUrl.startsWith('https://')) {
        baseUrl = 'http://$baseUrl';
      }
      while (baseUrl.endsWith('/')) {
        baseUrl = baseUrl.substring(0, baseUrl.length - 1);
      }

      // 1. Fetch Cloud Auth Token
      String authToken = SingletonClass().authToken ?? '';
      if (authToken.isEmpty) {
        try {
          final tokenRes = await http.get(
            Uri.parse('$baseUrl/api/dev/token'),
            headers: {'Accept': 'application/json'},
          ).timeout(const Duration(seconds: 10));
          if (tokenRes.statusCode == 200) {
            final body = json.decode(tokenRes.body);
            if (body is Map && body['token'] != null) {
              authToken = body['token'].toString();
            }
          }
        } catch (e) {
          developer.log('Token fetch fallback note: $e', name: 'DbCloningScreen');
        }
      }

      final cleanSlug = tenantSlug.trim();
      final headers = <String, String>{
        'Accept': 'application/json',
        'Content-Type': 'application/json',
        if (cleanSlug.isNotEmpty && cleanSlug.toLowerCase() != 'masterdb') ...{
          'x-tenant-id': cleanSlug,
          'x-tenant-slug': cleanSlug,
          'X-Tenant-ID': cleanSlug,
          'X-Tenant-Slug': cleanSlug,
        },
        if (authToken.isNotEmpty) 'Authorization': 'Bearer $authToken',
      };

      // 2. Fetch User Details & Standard Users
      final List<Map<String, dynamic>> usersToInsert = [];
      final List<Map<String, dynamic>> userRolesToInsert = [];
      final List<Map<String, dynamic>> userStoresToInsert = [];

      try {
        final detailsUrl = Uri.parse('$baseUrl/api/users/details?limit=500');
        final dResp = await http.get(detailsUrl, headers: headers).timeout(const Duration(seconds: 15));
        if (dResp.statusCode == 200) {
          final dJson = json.decode(dResp.body);
          if (dJson is Map && dJson['data'] is List) {
            for (final u in dJson['data']) {
              if (u is Map) {
                usersToInsert.add(Map<String, dynamic>.from(u));
                final uId = u['id'];
                if (uId != null) {
                  if (u['roles'] is List) {
                    for (final r in u['roles']) {
                      if (r is Map && r['id'] != null) {
                        userRolesToInsert.add({
                          'user_id': uId,
                          'role_id': r['id'],
                          'metadata': '{}',
                        });
                      }
                    }
                  }
                  if (u['stores'] is List) {
                    for (final s in u['stores']) {
                      if (s is Map && s['id'] != null) {
                        userStoresToInsert.add({
                          'user_id': uId,
                          'store_id': s['id'],
                          'is_primary': (s['is_primary'] == true || s['is_primary'] == 1) ? 1 : 0,
                          'metadata': '{}',
                        });
                      }
                    }
                  }
                }
              }
            }
          }
        }
      } catch (e) {
        developer.log('User details HTTP fetch note: $e', name: 'DbCloningScreen');
      }

      // If user details didn't return users, fetch from standard /api/users
      if (usersToInsert.isEmpty) {
        try {
          final listUrl = Uri.parse('$baseUrl/api/users?limit=500');
          final lResp = await http.get(listUrl, headers: headers).timeout(const Duration(seconds: 15));
          if (lResp.statusCode == 200) {
            final lJson = json.decode(lResp.body);
            if (lJson is Map && lJson['data'] is List) {
              for (final item in lJson['data']) {
                if (item is Map) {
                  usersToInsert.add(Map<String, dynamic>.from(item));
                }
              }
            }
          }
        } catch (e) {
          developer.log('Users list HTTP fetch note: $e', name: 'DbCloningScreen');
        }
      }

      if (usersToInsert.isEmpty) {
        developer.log('Warning: No user records returned from cloud APIs.', name: 'DbCloningScreen');
        return 0;
      }

      // 3. Insert users into SQLite
      await db.execute('PRAGMA foreign_keys = OFF;');
      int insertedCount = 0;
      for (final u in usersToInsert) {
        final id = u['id'];
        if (id == null) continue;

        String unwrapText(dynamic val) {
          if (val == null) return '';
          if (val is String) return val;
          if (val is Map) {
            if (val['Valid'] == true && val['String'] != null) {
              return val['String'].toString();
            }
            if (val['String'] != null && val['Valid'] == null) {
              return val['String'].toString();
            }
            return '';
          }
          return val.toString();
        }

        final username = (u['username'] ?? 'user_$id').toString();
        final email = (u['email'] ?? '$username@local.pos').toString();
        dynamic rawHash = u['password_hash'] ?? u['passwordHash'] ?? u['PasswordHash'] ?? u['password'] ?? u['hash'];
        String passwordHash = (rawHash ?? '').toString();
        if (passwordHash.isEmpty) {
          // Preserve any existing password_hash already in local SQLite for this user ID
          try {
            final existingRows = await db.query('users', columns: ['password_hash'], where: 'id = ?', whereArgs: [id], limit: 1);
            if (existingRows.isNotEmpty && existingRows.first['password_hash'] != null && existingRows.first['password_hash'].toString().isNotEmpty) {
              passwordHash = existingRows.first['password_hash'].toString();
            }
          } catch (_) {}
        }
        final firstName = unwrapText(u['first_name']);
        final lastName = unwrapText(u['last_name']);
        final employeeCode = unwrapText(u['employee_code']);
        final isActive = (u['is_active'] == true || u['is_active'] == 1 || u['is_active'] == 'true') ? 1 : 0;
        final orgId = u['organization_id'] ?? 1;

        try {
          await db.rawInsert('''
            INSERT OR REPLACE INTO users (
              id, organization_id, username, email, password_hash,
              first_name, last_name, employee_code, is_active,
              metadata, created_at, updated_at, must_reset_password, sap_imported
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0, 0)
          ''', [
            id,
            orgId,
            username,
            email,
            passwordHash,
            firstName,
            lastName,
            employeeCode,
            isActive,
            '{}',
          ]);
          insertedCount++;
        } catch (insertErr) {
          developer.log('Error inserting user $username ($id): $insertErr', name: 'DbCloningScreen');
        }
      }

      // 4. Insert user_roles and user_store_access
      for (final r in userRolesToInsert) {
        try {
          await db.rawInsert('''
            INSERT OR REPLACE INTO user_roles (user_id, role_id, metadata, assigned_at)
            VALUES (?, ?, ?, CURRENT_TIMESTAMP)
          ''', [r['user_id'], r['role_id'], r['metadata'] ?? '{}']);
        } catch (_) {}
      }

      for (final s in userStoresToInsert) {
        try {
          await db.rawInsert('''
            INSERT OR REPLACE INTO user_store_access (user_id, store_id, is_primary, metadata, granted_at)
            VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
          ''', [s['user_id'], s['store_id'], s['is_primary'] ?? 0, s['metadata'] ?? '{}']);
        } catch (_) {}
      }

      await db.execute('PRAGMA foreign_keys = ON;');
      developer.log('Direct HTTP sync inserted $insertedCount users into SQLite.', name: 'DbCloningScreen');
      return insertedCount;
    } catch (e, st) {
      developer.log('Error in _ensureUsersCloned: $e', name: 'DbCloningScreen', error: e, stackTrace: st);
      return 0;
    }
  }

  /// Checks local SQLite for cloned products. If empty, directly fetches products,
  /// variants, barcodes, prices, uom conversions, and inventory stock from cloud REST APIs and inserts them.
  Future<int> _ensureProductsCloned(String tenantSlug, String rawUrl) async {
    try {
      final db = DatabaseService().database;
      final totalProducts = await db.rawQuery('SELECT COUNT(*) AS cnt FROM products');
      final prodCount = Sqflite.firstIntValue(totalProducts) ?? 0;

      if (prodCount > 0) {
        developer.log('SQLite already has $prodCount products cloned.', name: 'DbCloningScreen');
        return prodCount;
      }

      developer.log('SQLite has 0 products! Performing direct product catalog sync from cloud...', name: 'DbCloningScreen');
      if (mounted) {
        setState(() {
          _statusMessage = 'Cloning product catalog, prices, and barcodes...';
        });
      }

      // Resolve base URL
      String baseUrl = rawUrl.trim();
      if (baseUrl.startsWith('grpc://')) {
        baseUrl = baseUrl.replaceFirst('grpc://', 'http://');
      } else if (baseUrl.startsWith('grpcs://')) {
        baseUrl = baseUrl.replaceFirst('grpcs://', 'https://');
      }
      if (!baseUrl.startsWith('http://') && !baseUrl.startsWith('https://')) {
        baseUrl = 'http://$baseUrl';
      }
      while (baseUrl.endsWith('/')) {
        baseUrl = baseUrl.substring(0, baseUrl.length - 1);
      }

      // 1. Fetch Cloud Auth Token
      String authToken = SingletonClass().authToken ?? '';
      if (authToken.isEmpty) {
        try {
          final tokenRes = await http.get(
            Uri.parse('$baseUrl/api/dev/token'),
            headers: {'Accept': 'application/json'},
          ).timeout(const Duration(seconds: 10));
          if (tokenRes.statusCode == 200) {
            final body = json.decode(tokenRes.body);
            if (body is Map && body['token'] != null) {
              authToken = body['token'].toString();
            }
          }
        } catch (e) {
          developer.log('Token fetch fallback note: $e', name: 'DbCloningScreen');
        }
      }

      final cleanSlug = tenantSlug.trim();
      final headers = <String, String>{
        'Accept': 'application/json',
        'Content-Type': 'application/json',
        if (cleanSlug.isNotEmpty && cleanSlug.toLowerCase() != 'masterdb') ...{
          'x-tenant-id': cleanSlug,
          'x-tenant-slug': cleanSlug,
          'X-Tenant-ID': cleanSlug,
          'X-Tenant-Slug': cleanSlug,
        },
        if (authToken.isNotEmpty) 'Authorization': 'Bearer $authToken',
      };

      int orgId = 1;
      try {
        final orgRows = await db.query('organizations', columns: ['id'], limit: 1);
        if (orgRows.isNotEmpty && orgRows.first['id'] != null) {
          orgId = (orgRows.first['id'] as num).toInt();
        }
      } catch (_) {}

      // 2. Fetch Master Product Catalog
      List<dynamic> productItems = [];
      try {
        final masterUrl = Uri.parse('$baseUrl/api/products/master-catalog?limit=2000&organization_id=$orgId');
        final mResp = await http.get(masterUrl, headers: headers).timeout(const Duration(seconds: 20));
        if (mResp.statusCode == 200) {
          final mJson = json.decode(mResp.body);
          if (mJson is Map) {
            final dataField = mJson['data'];
            if (dataField is Map && dataField['data'] is List) {
              productItems = dataField['data'] as List;
            } else if (dataField is List) {
              productItems = dataField;
            }
          }
        }
      } catch (e) {
        developer.log('Master catalog fetch note: $e', name: 'DbCloningScreen');
      }

      // Fallback 1: /api/products/catalog
      if (productItems.isEmpty) {
        try {
          final catUrl = Uri.parse('$baseUrl/api/products/catalog?limit=2000&organization_id=$orgId');
          final cResp = await http.get(catUrl, headers: headers).timeout(const Duration(seconds: 15));
          if (cResp.statusCode == 200) {
            final cJson = json.decode(cResp.body);
            if (cJson is Map && cJson['data'] is List) {
              productItems = cJson['data'] as List;
            }
          }
        } catch (e) {
          developer.log('Standard catalog fallback fetch note: $e', name: 'DbCloningScreen');
        }
      }

      // Fallback 2: /api/pos/stores/:store_id/products
      if (productItems.isEmpty) {
        try {
          final storeRows = await db.query('stores', columns: ['id'], limit: 5);
          for (final sRow in storeRows) {
            final sId = sRow['id'];
            if (sId == null) continue;
            final storeProdUrl = Uri.parse('$baseUrl/api/pos/stores/$sId/products?limit=2000');
            final spResp = await http.get(storeProdUrl, headers: headers).timeout(const Duration(seconds: 15));
            if (spResp.statusCode == 200) {
              final spJson = json.decode(spResp.body);
              if (spJson is Map && spJson['data'] is List) {
                final list = spJson['data'] as List;
                if (list.isNotEmpty) {
                  productItems.addAll(list);
                  break;
                }
              }
            }
          }
        } catch (e) {
          developer.log('Store products fallback fetch note: $e', name: 'DbCloningScreen');
        }
      }

      if (productItems.isEmpty) {
        developer.log('Warning: No product records returned from cloud endpoints.', name: 'DbCloningScreen');
        return 0;
      }

      // 3. Insert products and associated relations into SQLite
      await db.execute('PRAGMA foreign_keys = OFF;');
      // Ensure default price list
      await db.execute("INSERT OR IGNORE INTO price_lists (id, name, code, is_default, is_active) VALUES (1, 'Standard Price List', 'STANDARD', 1, 1);");

      int insertedCount = 0;
      for (final item in productItems) {
        if (item is! Map) continue;
        final p = Map<String, dynamic>.from(item);

        final rawId = p['product_id'] ?? p['id'];
        final pId = (rawId as num?)?.toInt();
        if (pId == null || pId <= 0) continue;

        final sku = (p['sku'] ?? p['product_sku'] ?? 'SKU-$pId').toString();
        final name = (p['name'] ?? p['product_name'] ?? 'Product #$pId').toString();
        final description = p['description']?.toString() ?? '';
        final categoryId = (p['category_id'] as num?)?.toInt();
        final brandId = (p['brand_id'] as num?)?.toInt();
        final baseUomId = (p['base_uom_id'] as num?)?.toInt();
        final taxCategoryId = (p['tax_category_id'] as num?)?.toInt();
        final productType = (p['product_type'] ?? 'standard').toString();
        final isActive = (p['is_active'] == false || p['is_active'] == 0 || p['is_active'] == 'false') ? 0 : 1;
        final isSellable = (p['is_sellable'] == false || p['is_sellable'] == 0 || p['is_sellable'] == 'false') ? 0 : 1;
        final isPurchasable = (p['is_purchasable'] == false || p['is_purchasable'] == 0 || p['is_purchasable'] == 'false') ? 0 : 1;
        final trackInventory = (p['track_inventory'] == false || p['track_inventory'] == 0 || p['track_inventory'] == 'false') ? 0 : 1;

        try {
          await db.rawInsert('''
            INSERT OR REPLACE INTO products (
              id, organization_id, sku, name, description,
              category_id, brand_id, base_uom_id, tax_category_id,
              product_type, is_active, is_sellable, is_purchasable,
              track_inventory, metadata, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
          ''', [
            pId,
            orgId,
            sku,
            name,
            description,
            categoryId,
            brandId,
            baseUomId,
            taxCategoryId,
            productType,
            isActive,
            isSellable,
            isPurchasable,
            trackInventory,
          ]);
          insertedCount++;
        } catch (insertErr) {
          developer.log('Error inserting product $sku ($pId): $insertErr', name: 'DbCloningScreen');
        }

        // Insert direct retail price if present at root
        final rootPrice = (p['retail_price'] ?? p['effective_price'] ?? p['price'] as num?)?.toDouble();
        if (rootPrice != null && rootPrice > 0) {
          try {
            await db.rawInsert('''
              INSERT OR REPLACE INTO product_prices (
                product_id, price_list_id, price, min_quantity, is_active, metadata, created_at, updated_at
              ) VALUES (?, 1, ?, 1, 1, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
            ''', [pId, rootPrice]);
          } catch (_) {}
        }

        // Insert direct barcode if present at root
        final rootBarcode = p['barcode']?.toString();
        if (rootBarcode != null && rootBarcode.isNotEmpty) {
          try {
            await db.rawInsert('''
              INSERT OR REPLACE INTO product_barcodes (
                product_id, barcode, barcode_type, is_primary, metadata, created_at
              ) VALUES (?, ?, 'CODE128', 1, '{}', CURRENT_TIMESTAMP)
            ''', [pId, rootBarcode]);
          } catch (_) {}
        }

        // Insert nested Prices
        if (p['prices'] is List) {
          for (final pr in p['prices'] as List) {
            if (pr is Map) {
              final priceVal = (pr['price'] as num?)?.toDouble() ?? 0.0;
              final prListId = (pr['price_list_id'] as num?)?.toInt() ?? 1;
              final prId = (pr['id'] as num?)?.toInt();
              final vId = (pr['product_variant_id'] as num?)?.toInt();
              try {
                if (prId != null && prId > 0) {
                  await db.rawInsert('''
                    INSERT OR REPLACE INTO product_prices (
                      id, product_id, product_variant_id, price_list_id, price, is_active, metadata, created_at, updated_at
                    ) VALUES (?, ?, ?, ?, ?, 1, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                  ''', [prId, pId, vId, prListId, priceVal]);
                } else {
                  await db.rawInsert('''
                    INSERT OR REPLACE INTO product_prices (
                      product_id, product_variant_id, price_list_id, price, is_active, metadata, created_at, updated_at
                    ) VALUES (?, ?, ?, ?, 1, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                  ''', [pId, vId, prListId, priceVal]);
                }
              } catch (_) {}
            }
          }
        }

        // Insert nested Variants
        if (p['variants'] is List) {
          for (final vr in p['variants'] as List) {
            if (vr is Map) {
              final vId = (vr['id'] as num?)?.toInt();
              final vSku = (vr['variant_sku'] ?? vr['sku'] ?? '$sku-V').toString();
              final vName = (vr['variant_name'] ?? vr['name'] ?? name).toString();
              final vAttrs = vr['variant_attributes'] is Map
                  ? json.encode(vr['variant_attributes'])
                  : (vr['variant_attributes']?.toString() ?? '{}');
              try {
                if (vId != null && vId > 0) {
                  await db.rawInsert('''
                    INSERT OR REPLACE INTO product_variants (
                      id, product_id, variant_sku, variant_name, variant_attributes, is_active, metadata, created_at, updated_at
                    ) VALUES (?, ?, ?, ?, ?, 1, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                  ''', [vId, pId, vSku, vName, vAttrs]);
                } else {
                  await db.rawInsert('''
                    INSERT OR REPLACE INTO product_variants (
                      product_id, variant_sku, variant_name, variant_attributes, is_active, metadata, created_at, updated_at
                    ) VALUES (?, ?, ?, ?, 1, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                  ''', [pId, vSku, vName, vAttrs]);
                }
              } catch (_) {}
            }
          }
        }

        // Insert nested Barcodes
        if (p['barcodes'] is List) {
          for (final bc in p['barcodes'] as List) {
            if (bc is Map) {
              final bCode = bc['barcode']?.toString();
              if (bCode == null || bCode.isEmpty) continue;
              final bId = (bc['id'] as num?)?.toInt();
              final bType = (bc['barcode_type'] ?? 'CODE128').toString();
              final isPrim = (bc['is_primary'] == true || bc['is_primary'] == 1) ? 1 : 0;
              final vId = (bc['product_variant_id'] as num?)?.toInt();
              try {
                if (bId != null && bId > 0) {
                  await db.rawInsert('''
                    INSERT OR REPLACE INTO product_barcodes (
                      id, product_id, product_variant_id, barcode, barcode_type, is_primary, metadata, created_at
                    ) VALUES (?, ?, ?, ?, ?, ?, '{}', CURRENT_TIMESTAMP)
                  ''', [bId, pId, vId, bCode, bType, isPrim]);
                } else {
                  await db.rawInsert('''
                    INSERT OR REPLACE INTO product_barcodes (
                      product_id, product_variant_id, barcode, barcode_type, is_primary, metadata, created_at
                    ) VALUES (?, ?, ?, ?, ?, '{}', CURRENT_TIMESTAMP)
                  ''', [pId, vId, bCode, bType, isPrim]);
                }
              } catch (_) {}
            }
          }
        }

        // Insert nested Inventory Stock
        if (p['inventory'] is List) {
          for (final inv in p['inventory'] as List) {
            if (inv is Map) {
              final sId = (inv['store_id'] as num?)?.toInt() ?? 1;
              final qtyOnHand = (inv['quantity_on_hand'] as num?)?.toDouble() ?? 0.0;
              final qtyAvail = (inv['quantity_available'] as num?)?.toDouble() ?? qtyOnHand;
              final vId = (inv['product_variant_id'] as num?)?.toInt();
              try {
                await db.rawInsert('''
                  INSERT OR REPLACE INTO inventory_stock (
                    product_id, product_variant_id, store_id, quantity_on_hand, quantity_available, metadata, created_at, updated_at
                  ) VALUES (?, ?, ?, ?, ?, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ''', [pId, vId, sId, qtyOnHand, qtyAvail]);
              } catch (_) {}
            }
          }
        }
      }

      await db.execute('PRAGMA foreign_keys = ON;');
      developer.log('Direct HTTP sync inserted $insertedCount products into SQLite.', name: 'DbCloningScreen');
      return insertedCount;
    } catch (e, st) {
      developer.log('Error in _ensureProductsCloned: $e', name: 'DbCloningScreen', error: e, stackTrace: st);
      return 0;
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
                _currentPhase == SyncPhase.failed && _errorMessage != null
                    ? _errorMessage!
                    : _statusMessage,
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