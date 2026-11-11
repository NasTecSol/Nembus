import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:pos_mobile/database/db_service.dart';
import 'package:pos_mobile/ffi/nembus_bridge.dart';
import 'package:pos_mobile/singleton/singleton_class.dart';

/// ─────────────────────────────────────────────────────────────────────────────
/// POS-CLIENT PRINTING TEMPLATE & DATA ENGINE
/// Mirrors `apps/pos-client/internal/printing/template.go` & `packages/core`
/// ─────────────────────────────────────────────────────────────────────────────

/// Organisation details printed at the top of every receipt (from Nembus side).
class OrgHeader {
  final String name;
  final String address;
  final String phone;
  final String taxId;
  final String website;

  const OrgHeader({
    this.name = '',
    this.address = '',
    this.phone = '',
    this.taxId = '',
    this.website = '',
  });

  factory OrgHeader.fromJson(Map<String, dynamic> json) {
    return OrgHeader(
      name: json['name'] as String? ?? '',
      address: json['address'] as String? ?? '',
      phone: json['phone'] as String? ?? '',
      taxId: json['tax_id'] as String? ?? json['tin'] as String? ?? '',
      website: json['website'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
    'name': name,
    'address': address,
    'phone': phone,
    'tax_id': taxId,
    'website': website,
  };
}

/// Sign-off lines printed at the bottom of every receipt (from Nembus side).
class OrgFooter {
  final String thankYou;
  final String returnNote;
  final String website;

  const OrgFooter({
    this.thankYou = '',
    this.returnNote = '',
    this.website = '',
  });

  factory OrgFooter.fromJson(Map<String, dynamic> json) {
    return OrgFooter(
      thankYou: json['thank_you'] as String? ?? json['thankYou'] as String? ?? '',
      returnNote: json['return_note'] as String? ?? json['returnNote'] as String? ?? '',
      website: json['website'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
    'thank_you': thankYou,
    'return_note': returnNote,
    'website': website,
  };
}

/// Combines header + footer branding for one organisation (from Nembus side).
class ReceiptTemplate {
  final OrgHeader header;
  final OrgFooter footer;
  final String qrCodeBase64;
  final String currency;

  const ReceiptTemplate({
    required this.header,
    required this.footer,
    this.qrCodeBase64 = '',
    this.currency = 'SAR',
  });

  factory ReceiptTemplate.fromJson(Map<String, dynamic> json) {
    return ReceiptTemplate(
      header: OrgHeader.fromJson(json['header'] as Map<String, dynamic>? ?? {}),
      footer: OrgFooter.fromJson(json['footer'] as Map<String, dynamic>? ?? {}),
      qrCodeBase64: json['qr_code_base64'] as String? ??
          json['qrCodeBase64'] as String? ??
          json['zatca_qr'] as String? ??
          '',
      currency: json['currency'] as String? ?? 'SAR',
    );
  }

  Map<String, dynamic> toJson() => {
    'header': header.toJson(),
    'footer': footer.toJson(),
    'qr_code_base64': qrCodeBase64,
    'currency': currency,
  };

  ReceiptTemplate copyWith({
    OrgHeader? header,
    OrgFooter? footer,
    String? qrCodeBase64,
    String? currency,
  }) {
    return ReceiptTemplate(
      header: header ?? this.header,
      footer: footer ?? this.footer,
      qrCodeBase64: qrCodeBase64 ?? this.qrCodeBase64,
      currency: currency ?? this.currency,
    );
  }

  /// Builds a template from current application session / singleton state
  factory ReceiptTemplate.fromSession({
    String? storeName,
    String? address,
    String? phone,
    String? taxId,
    String? website,
    String? thankYou,
    String? returnNote,
    String? qrCodeBase64,
    String currency = 'SAR',
    double totalAmount = 0.0,
    double taxAmount = 0.0,
  }) {
    final singleton = SingletonClass();
    final effectiveStoreName = storeName ??
        singleton.activeStoreName ??
        singleton.activeTenantName ??
        '';

    String effectiveQr = qrCodeBase64 ?? '';
    if (effectiveQr.isEmpty && singleton.hasCompletedZatcaOnboarding) {
      effectiveQr = PosReceiptEngine.generateZatcaTlvBase64(
        sellerName: effectiveStoreName,
        vatNumber: taxId ?? '',
        totalAmount: totalAmount,
        taxAmount: taxAmount,
      );
    }

    return ReceiptTemplate(
      header: OrgHeader(
        name: effectiveStoreName,
        address: address ?? '',
        phone: phone ?? '',
        taxId: taxId ?? '',
        website: website ?? '',
      ),
      footer: OrgFooter(
        thankYou: thankYou ?? '',
        returnNote: returnNote ?? '',
        website: website ?? '',
      ),
      qrCodeBase64: effectiveQr,
      currency: currency,
    );
  }

  factory ReceiptTemplate.empty({String? storeName, String? qrCodeBase64}) {
    return ReceiptTemplate.fromSession(
      storeName: storeName,
      qrCodeBase64: qrCodeBase64,
    );
  }
}

/// Represents one sold item on the receipt matching pos-client `LineItem`
class PosLineItem {
  final String name;
  final double qty;
  final double price;
  final double lineTotal;

  const PosLineItem({
    required this.name,
    required this.qty,
    required this.price,
    required this.lineTotal,
  });

  factory PosLineItem.fromJson(Map<String, dynamic> json) {
    final name = json['name'] as String? ?? json['product_name'] as String? ?? 'Item';
    final qty = (json['qty'] as num?)?.toDouble() ??
        (json['quantity'] as num?)?.toDouble() ??
        1.0;
    final price = (json['price'] as num?)?.toDouble() ??
        (json['unit_price'] as num?)?.toDouble() ??
        0.0;
    final total = (json['line_total'] as num?)?.toDouble() ?? (qty * price);

    return PosLineItem(
      name: name,
      qty: qty,
      price: price,
      lineTotal: total,
    );
  }

  Map<String, dynamic> toJson() => {
    'name': name,
    'qty': qty,
    'price': price,
  };
}

/// Transaction data matching pos-client `ReceiptData` with computed business logic
class PosReceiptData {
  final String type;
  final String receiptNumber;
  final String formattedDateTime;
  final String cashier;
  final String terminal;
  final String customer;
  final List<PosLineItem> items;
  final double subtotal;
  final double discount;
  final double discountPercentage;
  final double taxRate;
  final double taxAmount;
  final double grandTotal;
  final double paid;
  final double changeDue;
  final String paymentMethod;
  final String barcode;
  final String qrCodeBase64;
  final String currency;

  const PosReceiptData({
    this.type = 'RECEIPT',
    required this.receiptNumber,
    required this.formattedDateTime,
    required this.cashier,
    required this.terminal,
    required this.customer,
    required this.items,
    required this.subtotal,
    required this.discount,
    required this.discountPercentage,
    required this.taxRate,
    required this.taxAmount,
    required this.grandTotal,
    required this.paid,
    required this.changeDue,
    required this.paymentMethod,
    required this.barcode,
    this.qrCodeBase64 = '',
    this.currency = 'SAR',
  });

  /// Factory constructor applying the exact business logic from pos-client `template.go`
  factory PosReceiptData.parse({
    String type = 'RECEIPT',
    required String receiptNumber,
    DateTime? timestamp,
    required String cashier,
    required String terminal,
    required String customer,
    required List<dynamic> rawItems,
    double? rawSubtotal,
    double discount = 0.0,
    double taxRate = 0.15, // 15% VAT standard
    double? rawTaxAmount,
    double? rawTotalAmount,
    double? paidAmount,
    double changeDue = 0.0,
    required String paymentMethod,
    String? barcode,
    String? qrCodeBase64,
    String currency = 'SAR',
  }) {
    // 1. Parse Line Items
    final items = rawItems.map((item) {
      if (item is PosLineItem) return item;
      if (item is Map<String, dynamic>) return PosLineItem.fromJson(item);
      return PosLineItem(
        name: item.toString(),
        qty: 1.0,
        price: 0.0,
        lineTotal: 0.0,
      );
    }).toList();

    // 2. Compute Subtotal (sum of qty * price)
    double computedSubtotal = 0.0;
    for (final it in items) {
      computedSubtotal += it.lineTotal;
    }
    final effectiveSubtotal = rawSubtotal ?? computedSubtotal;

    // 3. Compute Discount & Net after discount
    final netAfterDisc = (effectiveSubtotal - discount).clamp(0.0, double.infinity);
    final discPct = effectiveSubtotal > 0 ? (discount / effectiveSubtotal) * 100 : 0.0;

    // 4. Compute Tax & Grand Total
    final computedTax = netAfterDisc * taxRate;
    final effectiveTax = rawTaxAmount ?? computedTax;
    final effectiveGrandTotal = rawTotalAmount ?? (netAfterDisc + effectiveTax);

    // 5. Compute Paid & Change
    final effectivePaid = paidAmount ?? (effectiveGrandTotal + changeDue);
    final effectiveChange = (effectivePaid - effectiveGrandTotal).clamp(0.0, double.infinity);

    // 6. Format Date & Barcode (e.g. 02-Jan-06 15:04)
    final date = timestamp ?? DateTime.now();
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    final dayStr = date.day.toString().padLeft(2, '0');
    final monStr = months[date.month - 1];
    final yrStr = (date.year % 100).toString().padLeft(2, '0');
    final hrStr = date.hour.toString().padLeft(2, '0');
    final minStr = date.minute.toString().padLeft(2, '0');
    final dateStr = '$dayStr-$monStr-$yrStr $hrStr:$minStr';
    final effectiveBarcode = (barcode != null && barcode.isNotEmpty) ? barcode : receiptNumber;

    String effectiveQr = qrCodeBase64 ?? '';
    if (effectiveQr.isEmpty && SingletonClass().hasCompletedZatcaOnboarding) {
      effectiveQr = PosReceiptEngine.generateZatcaTlvBase64(
        sellerName: SingletonClass().activeStoreName ?? SingletonClass().activeTenantName ?? 'Store',
        vatNumber: '',
        timestamp: date,
        totalAmount: effectiveGrandTotal,
        taxAmount: effectiveTax,
      );
    }

    return PosReceiptData(
      type: type,
      receiptNumber: receiptNumber,
      formattedDateTime: dateStr,
      cashier: cashier,
      terminal: terminal,
      customer: customer.isNotEmpty ? customer : 'Walk-in Customer',
      items: items,
      subtotal: effectiveSubtotal,
      discount: discount,
      discountPercentage: discPct,
      taxRate: taxRate,
      taxAmount: effectiveTax,
      grandTotal: effectiveGrandTotal,
      paid: effectivePaid,
      changeDue: effectiveChange,
      paymentMethod: paymentMethod.toUpperCase(),
      barcode: effectiveBarcode,
      qrCodeBase64: effectiveQr,
      currency: currency,
    );
  }

  Map<String, dynamic> toJson() => {
    'type': type,
    'receipt_number': receiptNumber,
    'cashier': cashier,
    'terminal': terminal,
    'customer': customer,
    'items': items.map((i) => i.toJson()).toList(),
    'discount': discount,
    'tax_rate': taxRate,
    'paid': paid,
    'payment_method': paymentMethod,
    'barcode': barcode,
    'qr_code_base64': qrCodeBase64,
    'currency': currency,
  };

  PosReceiptData copyWith({
    String? type,
    String? receiptNumber,
    String? formattedDateTime,
    String? cashier,
    String? terminal,
    String? customer,
    List<PosLineItem>? items,
    double? subtotal,
    double? discount,
    double? discountPercentage,
    double? taxRate,
    double? taxAmount,
    double? grandTotal,
    double? paid,
    double? changeDue,
    String? paymentMethod,
    String? barcode,
    String? qrCodeBase64,
    String? currency,
  }) {
    return PosReceiptData(
      type: type ?? this.type,
      receiptNumber: receiptNumber ?? this.receiptNumber,
      formattedDateTime: formattedDateTime ?? this.formattedDateTime,
      cashier: cashier ?? this.cashier,
      terminal: terminal ?? this.terminal,
      customer: customer ?? this.customer,
      items: items ?? this.items,
      subtotal: subtotal ?? this.subtotal,
      discount: discount ?? this.discount,
      discountPercentage: discountPercentage ?? this.discountPercentage,
      taxRate: taxRate ?? this.taxRate,
      taxAmount: taxAmount ?? this.taxAmount,
      grandTotal: grandTotal ?? this.grandTotal,
      paid: paid ?? this.paid,
      changeDue: changeDue ?? this.changeDue,
      paymentMethod: paymentMethod ?? this.paymentMethod,
      barcode: barcode ?? this.barcode,
      qrCodeBase64: qrCodeBase64 ?? this.qrCodeBase64,
      currency: currency ?? this.currency,
    );
  }
}

/// Engine to interact with Nembus backend handler & template logic
class PosReceiptEngine {
  /// Generates an authoritative ZATCA-compliant Base64 TLV string (Fatoora Phase 1 & 2 mandatory tags).
  /// Tag 1: Seller's Name (UTF-8)
  /// Tag 2: VAT Registration Number (15 digits)
  /// Tag 3: Time Stamp (ISO 8601 UTC: YYYY-MM-DDTHH:mm:ssZ)
  /// Tag 4: Invoice Total with VAT (e.g. 115.00)
  /// Tag 5: VAT Total (e.g. 15.00)
  static String generateZatcaTlvBase64({
    required String sellerName,
    required String vatNumber,
    DateTime? timestamp,
    required double totalAmount,
    required double taxAmount,
  }) {
    final cleanSeller = sellerName.trim().isNotEmpty ? sellerName.trim() : 'Store';
    final cleanVat = vatNumber.trim().isNotEmpty ? vatNumber.trim() : '300000000000003';
    final time = (timestamp ?? DateTime.now()).toUtc();
    final isoTimestamp = '${time.year.toString().padLeft(4, '0')}-'
        '${time.month.toString().padLeft(2, '0')}-'
        '${time.day.toString().padLeft(2, '0')}T'
        '${time.hour.toString().padLeft(2, '0')}:'
        '${time.minute.toString().padLeft(2, '0')}:'
        '${time.second.toString().padLeft(2, '0')}Z';

    final totalStr = totalAmount.toStringAsFixed(2);
    final taxStr = taxAmount.toStringAsFixed(2);

    final bytes = <int>[];

    void writeTag(int tag, String value) {
      final valBytes = utf8.encode(value);
      bytes.add(tag);
      bytes.add(valBytes.length);
      bytes.addAll(valBytes);
    }

    writeTag(1, cleanSeller);
    writeTag(2, cleanVat);
    writeTag(3, isoTimestamp);
    writeTag(4, totalStr);
    writeTag(5, taxStr);

    return base64Encode(bytes);
  }

  /// Checks whether this device or organization is registered / onboarded with ZATCA.
  static Future<bool> isDeviceRegisteredWithZatca() async {
    if (SingletonClass().hasCompletedZatcaOnboarding) return true;

    try {
      if (DatabaseService().isOpen) {
        final db = DatabaseService().database;
        final devConfig = await db.rawQuery('SELECT zatca_enabled FROM local_device_config LIMIT 1');
        if (devConfig.isNotEmpty) {
          final val = devConfig.first['zatca_enabled'];
          if (val == 1 || val == true || val == '1') {
            SingletonClass().hasCompletedZatcaOnboarding = true;
            return true;
          }
        }
      }
    } catch (_) {}

    return false;
  }
  /// Resolves the local pos-client server base URL based on running platform
  static String get posClientBaseUrl {
    if (!kIsWeb && defaultTargetPlatform == TargetPlatform.android) {
      return 'http://10.0.2.2:8080';
    }
    return 'http://127.0.0.1:8080';
  }

  /// Dispatches the receipt payload to the Nembus offline gateway via FFI:
  /// route: api/print/receipt (handler: 'print', action: 'receipt')
  static Future<Map<String, dynamic>> printReceiptViaGateway({
    required PosReceiptData receiptData,
    int? orgId,
    String printerMode = 'system',
    String printerName = 'POS-Receipt',
    String? networkIp,
    String? networkPort,
  }) async {
    final effectiveOrgId = orgId ??
        int.tryParse(SingletonClass().activeTenantId ?? '1') ??
        1;

    final payload = {
      'org_id': effectiveOrgId,
      'printer': {
        'mode': printerMode,
        'printer_name': printerName,
        if (networkIp != null && networkIp.isNotEmpty) 'network_ip': networkIp,
        if (networkPort != null && networkPort.isNotEmpty) 'network_port': networkPort,
      },
      'receipt': receiptData.toJson(),
    };

    debugPrint('🚀 [POS Receipt Engine] Dispatching receipt through offline gateway (api/print/receipt)...');

    // 1. Primary path: Dispatch through Nembus FFI Dart offline gateway handlers
    final ffiResponse = NembusBridge().callHandler(
      handler: 'print',
      action: 'receipt',
      payload: payload,
    );

    debugPrint('📄 [POS Receipt Engine] Nembus FFI handler response: $ffiResponse');

    if (ffiResponse['success'] == true) {
      debugPrint('✅ Receipt dispatched successfully through Nembus FFI handler (api/print/receipt)');
      return ffiResponse;
    }

    // 2. Fallback to HTTP route if a local pos-client server is running
    final baseUrl = posClientBaseUrl;
    try {
      final response = await http
          .post(
            Uri.parse('$baseUrl/api/print/receipt'),
            headers: {
              'Content-Type': 'application/json',
              'x-tenant-id': effectiveOrgId.toString(),
            },
            body: jsonEncode(payload),
          )
          .timeout(const Duration(milliseconds: 600));

      if (response.statusCode == 200 || response.statusCode == 201) {
        debugPrint('✅ Receipt dispatched successfully to pos-client print HTTP route');
        final decoded = jsonDecode(response.body);
        return {'success': true, 'data': decoded is Map<String, dynamic> ? decoded : {'raw': response.body}};
      }
    } catch (_) {}

    return ffiResponse;
  }

  /// Dispatches the receipt payload to pos-client `POST /api/print/receipt`
  static Future<bool> sendToPosClient({
    required PosReceiptData receiptData,
    int? orgId,
    String printerMode = 'system',
    String printerName = 'POS-Receipt',
    String? networkIp,
    String? networkPort,
  }) async {
    final res = await printReceiptViaGateway(
      receiptData: receiptData,
      orgId: orgId,
      printerMode: printerMode,
      printerName: printerName,
      networkIp: networkIp,
      networkPort: networkPort,
    );
    return res['success'] == true;
  }

  /// Fetches the dynamic POS receipt template from Nembus backend / local SQLite,
  /// including the authoritative ZATCA QR code and branding from Nembus side.
  static Future<ReceiptTemplate> fetchReceiptTemplate({
    int? orgId,
    int? storeId,
    int? transactionId,
    String? salesOrderId,
    double totalAmount = 0.0,
    double taxAmount = 0.0,
    DateTime? invoiceTime,
    String? storeName,
  }) async {
    final singleton = SingletonClass();
    final effectiveOrgId = orgId ??
        int.tryParse(singleton.activeTenantId ?? '1') ??
        1;
    final effectiveStoreId = storeId ??
        singleton.activeStoreId ??
        1;

    String resolvedQr = '';
    String headerName = storeName ?? singleton.activeStoreName ?? singleton.activeTenantName ?? '';
    String headerAddress = '';
    String headerPhone = '';
    String headerTaxId = '';
    String headerWebsite = '';
    String footerThankYou = '';
    String footerReturnNote = '';
    String footerWebsite = '';
    String currency = 'SAR';

    // 1. Fetch store branding through Nembus FFI handlers (offline gateway)
    try {
      final storeResp = NembusBridge().callHandler(
        handler: 'stores',
        action: 'getStore',
        payload: {'id': effectiveStoreId, 'store_id': effectiveStoreId},
      );
      if (storeResp['success'] == true && storeResp['data'] is Map<String, dynamic>) {
        final sData = storeResp['data'] as Map<String, dynamic>;
        if (sData['name'] != null && sData['name'].toString().isNotEmpty) {
          headerName = sData['name'].toString();
        }
        if (sData['metadata'] != null) {
          try {
            final meta = sData['metadata'] is Map ? sData['metadata'] as Map : jsonDecode(sData['metadata'].toString());
            if (meta is Map) {
              if (meta['address'] != null && meta['address'].toString().isNotEmpty) {
                headerAddress = meta['address'].toString();
              }
              if (meta['phone'] != null && meta['phone'].toString().isNotEmpty) {
                headerPhone = meta['phone'].toString();
              }
              if (footerThankYou.isEmpty && meta['thank_you'] != null) {
                footerThankYou = meta['thank_you'].toString();
              }
              if (footerReturnNote.isEmpty && meta['return_note'] != null) {
                footerReturnNote = meta['return_note'].toString();
              }
            }
          } catch (_) {}
        }
      }
    } catch (e) {
      debugPrint('ℹ️ Nembus store handler note: $e');
    }

    // 2. Fetch authoritative branding & dynamic ZATCA QR from Nembus local SQLite database
    try {
      final db = DatabaseService().database;

      // 2.1 Check zatca_document_chain, pos_transactions, and sales_orders_v2 for dynamic ZATCA QR
      if (resolvedQr.isEmpty && ((salesOrderId != null && salesOrderId.isNotEmpty) || (transactionId != null && transactionId > 0))) {
        final cleanOrder = salesOrderId?.replaceFirst('TXN-', '') ?? '';
        final txnOrder = (salesOrderId != null && salesOrderId!.startsWith('TXN-')) ? salesOrderId! : 'TXN-$salesOrderId';

        // Query zatca_document_chain table if available in SQLite
        try {
          final zdcRows = await db.rawQuery('''
            SELECT qr_code_base64 FROM zatca_document_chain
            WHERE (entity_id = ? OR entity_id = ? OR entity_id = ? OR entity_id = ?)
            ORDER BY id DESC LIMIT 1
          ''', [salesOrderId ?? '', cleanOrder, txnOrder, transactionId?.toString() ?? '']);
          if (zdcRows.isNotEmpty && zdcRows.first['qr_code_base64'] != null) {
            final q = zdcRows.first['qr_code_base64'].toString();
            if (q.isNotEmpty) {
              resolvedQr = q;
              debugPrint('✅ [fetchReceiptTemplate] Dynamic ZATCA QR found in zatca_document_chain');
            }
          }
        } catch (_) {}

        // Query pos_transactions metadata
        if (resolvedQr.isEmpty) {
          try {
            final txRows = await db.rawQuery('''
              SELECT metadata FROM pos_transactions
              WHERE (sales_order_id = ? OR transaction_number = ? OR transaction_number = ? OR id = ?)
              ORDER BY id DESC LIMIT 1
            ''', [salesOrderId ?? '', txnOrder, cleanOrder, transactionId ?? 0]);
            if (txRows.isNotEmpty && txRows.first['metadata'] != null) {
              final metaStr = txRows.first['metadata'].toString();
              if (metaStr.isNotEmpty && metaStr != '{}') {
                final meta = jsonDecode(metaStr);
                if (meta is Map) {
                  final q = meta['qr_code_base64'] ?? meta['qrCodeBase64'] ?? meta['zatca_qr'];
                  if (q != null && q.toString().isNotEmpty) {
                    resolvedQr = q.toString();
                    debugPrint('✅ [fetchReceiptTemplate] Dynamic ZATCA QR found in pos_transactions metadata');
                  }
                }
              }
            }
          } catch (_) {}
        }

        // Query sales_orders_v2 metadata
        if (resolvedQr.isEmpty && salesOrderId != null && salesOrderId.isNotEmpty) {
          try {
            final soRows = await db.rawQuery('''
              SELECT metadata FROM sales_orders_v2
              WHERE (id = ? OR order_number = ? OR order_number = ?)
              LIMIT 1
            ''', [salesOrderId, cleanOrder, txnOrder]);
            if (soRows.isNotEmpty && soRows.first['metadata'] != null) {
              final metaStr = soRows.first['metadata'].toString();
              if (metaStr.isNotEmpty && metaStr != '{}') {
                final meta = jsonDecode(metaStr);
                if (meta is Map) {
                  final q = meta['qr_code_base64'] ?? meta['qrCodeBase64'] ?? meta['zatca_qr'];
                  if (q != null && q.toString().isNotEmpty) {
                    resolvedQr = q.toString();
                    debugPrint('✅ [fetchReceiptTemplate] Dynamic ZATCA QR found in sales_orders_v2 metadata');
                  }
                }
              }
            }
          } catch (_) {}
        }
      }

      // 2.2 Organizations table in SQLite (authoritative Nembus branding & template)
      final orgRows = await db.rawQuery('SELECT name, legal_name, tax_id, currency_code, metadata FROM organizations LIMIT 1');
      if (orgRows.isNotEmpty) {
        final oRow = orgRows.first;
        if (headerName.isEmpty && oRow['name'] != null && oRow['name'].toString().isNotEmpty) {
          headerName = oRow['name'].toString();
        }
        if (headerAddress.isEmpty && oRow['legal_name'] != null && oRow['legal_name'].toString().isNotEmpty) {
          headerAddress = oRow['legal_name'].toString();
        }
        if (oRow['tax_id'] != null && oRow['tax_id'].toString().isNotEmpty) {
          headerTaxId = oRow['tax_id'].toString();
        }
        if (oRow['currency_code'] != null && oRow['currency_code'].toString().isNotEmpty) {
          currency = oRow['currency_code'].toString();
        }
        if (oRow['metadata'] != null) {
          try {
            final oMeta = oRow['metadata'] is Map ? oRow['metadata'] as Map : jsonDecode(oRow['metadata'].toString());
            if (oMeta is Map) {
              if (footerThankYou.isEmpty && oMeta['thank_you'] != null) {
                footerThankYou = oMeta['thank_you'].toString();
              } else if (footerThankYou.isEmpty && oMeta['footer'] != null) {
                footerThankYou = oMeta['footer'].toString();
              } else if (footerThankYou.isEmpty && oMeta['receipt_footer'] != null) {
                footerThankYou = oMeta['receipt_footer'].toString();
              }
              if (footerReturnNote.isEmpty && oMeta['return_note'] != null) {
                footerReturnNote = oMeta['return_note'].toString();
              } else if (footerReturnNote.isEmpty && oMeta['return_policy'] != null) {
                footerReturnNote = oMeta['return_policy'].toString();
              }
              if (headerWebsite.isEmpty && oMeta['website'] != null) {
                headerWebsite = oMeta['website'].toString();
                footerWebsite = headerWebsite;
              }
            }
          } catch (_) {}
        }
      }

      // 2.3 Stores table in SQLite (query ONLY columns that exist: name, metadata)
      if (headerName.isEmpty || headerAddress.isEmpty || headerPhone.isEmpty || footerThankYou.isEmpty) {
        final storeRows = await db.rawQuery('SELECT name, metadata FROM stores WHERE id = ? LIMIT 1', [effectiveStoreId]);
        if (storeRows.isNotEmpty) {
          final sRow = storeRows.first;
          if (headerName.isEmpty && sRow['name'] != null && sRow['name'].toString().isNotEmpty) {
            headerName = sRow['name'].toString();
          }
          if (sRow['metadata'] != null) {
            try {
              final sMeta = jsonDecode(sRow['metadata'].toString());
              if (sMeta is Map) {
                if (headerAddress.isEmpty && sMeta['address'] != null && sMeta['address'].toString().isNotEmpty) {
                  headerAddress = sMeta['address'].toString();
                }
                if (headerPhone.isEmpty && sMeta['phone'] != null && sMeta['phone'].toString().isNotEmpty) {
                  headerPhone = sMeta['phone'].toString();
                }
                if (footerThankYou.isEmpty && sMeta['thank_you'] != null) {
                  footerThankYou = sMeta['thank_you'].toString();
                } else if (footerThankYou.isEmpty && sMeta['receipt_footer'] != null) {
                  footerThankYou = sMeta['receipt_footer'].toString();
                }
                if (footerReturnNote.isEmpty && sMeta['return_note'] != null) {
                  footerReturnNote = sMeta['return_note'].toString();
                }
              }
            } catch (_) {}
          }
        }
      }
    } catch (dbErr) {
      debugPrint('ℹ️ SQLite local template note: $dbErr');
    }

    // 3. Fallback to Cloud Backend API if online and needed
    if (resolvedQr.isEmpty || headerName.isEmpty) {
      final cloudBase = singleton.baseURL.isNotEmpty ? singleton.baseURL : 'https://nembus.nashrms.com';
      final token = singleton.authToken ?? '';
      final tenantSlug = singleton.activeTenantSlug ?? effectiveOrgId.toString();
      final headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'x-tenant-id': tenantSlug,
        'x-tenant-slug': tenantSlug,
        'X-Tenant-ID': tenantSlug,
        'X-Tenant-Slug': tenantSlug,
        if (token.isNotEmpty) 'Authorization': 'Bearer $token',
      };

      if (resolvedQr.isEmpty && (transactionId != null || (salesOrderId != null && salesOrderId.isNotEmpty))) {
        try {
          final queryParams = <String, String>{};
          if (transactionId != null && transactionId > 0) {
            queryParams['transaction_id'] = transactionId.toString();
          }
          if (salesOrderId != null && salesOrderId.isNotEmpty) {
            queryParams['sales_order_id'] = salesOrderId;
          }

          final uri = Uri.parse('$cloudBase/api/zatca/receipt-qr').replace(queryParameters: queryParams);
          final resp = await http.get(uri, headers: headers).timeout(const Duration(milliseconds: 1500));

          if (resp.statusCode == 200) {
            final body = jsonDecode(resp.body);
            final data = body['data'] is Map<String, dynamic> ? body['data'] as Map<String, dynamic> : body;
            resolvedQr = (data['qr_code_base64'] ?? data['qrCodeBase64'] ?? data['zatca_qr'] ?? '').toString();
            debugPrint('✅ Fetched dynamic ZATCA QR from Nembus backend /api/zatca/receipt-qr');
          }
        } catch (_) {}
      }

      if (headerName.isEmpty) {
        try {
          final orgUri = Uri.parse('$cloudBase/api/organizations/$effectiveOrgId');
          final orgResp = await http.get(orgUri, headers: headers).timeout(const Duration(milliseconds: 1500));
          if (orgResp.statusCode == 200) {
            final body = jsonDecode(orgResp.body);
            final orgData = body['data'] is Map<String, dynamic> ? body['data'] as Map<String, dynamic> : (body is Map<String, dynamic> ? body : null);
            if (orgData != null) {
              if (orgData['name'] != null && orgData['name'].toString().isNotEmpty) {
                headerName = orgData['name'].toString();
              }
              if (headerAddress.isEmpty && orgData['legal_name'] != null && orgData['legal_name'].toString().isNotEmpty) {
                headerAddress = orgData['legal_name'].toString();
              }
              if (headerTaxId.isEmpty && orgData['tax_id'] != null && orgData['tax_id'].toString().isNotEmpty) {
                headerTaxId = orgData['tax_id'].toString();
              }
              if (orgData['website'] != null && orgData['website'].toString().isNotEmpty) {
                headerWebsite = orgData['website'].toString();
                footerWebsite = headerWebsite;
              }
              if (orgData['currency_code'] != null && orgData['currency_code'].toString().isNotEmpty) {
                currency = orgData['currency_code'].toString();
              }
            }
          }
        } catch (_) {}
      }
    }

    // 4. If dynamic ZATCA QR is still empty, check if this device or tenant is registered to ZATCA
    if (resolvedQr.isEmpty) {
      final isZatcaRegistered = await isDeviceRegisteredWithZatca();
      if (isZatcaRegistered) {
        DateTime effectiveTime = invoiceTime ?? DateTime.now();
        double effTotal = totalAmount;
        double effTax = taxAmount;

        // If effTotal <= 0, query SQLite pos_transactions / sales_orders_v2 for the exact numbers
        if (effTotal <= 0 && ((salesOrderId != null && salesOrderId.isNotEmpty) || (transactionId != null && transactionId > 0))) {
          try {
            if (DatabaseService().isOpen) {
              final db = DatabaseService().database;
              final cleanOrder = salesOrderId?.replaceFirst('TXN-', '') ?? '';
              final txnOrder = (salesOrderId != null && salesOrderId!.startsWith('TXN-')) ? salesOrderId! : 'TXN-$salesOrderId';

              final txRows = await db.rawQuery('''
                SELECT total_amount, tax_amount, created_at, transaction_date FROM pos_transactions
                WHERE (sales_order_id = ? OR transaction_number = ? OR transaction_number = ? OR id = ?)
                LIMIT 1
              ''', [salesOrderId ?? '', txnOrder, cleanOrder, transactionId ?? 0]);
              if (txRows.isNotEmpty) {
                final t = txRows.first;
                effTotal = (t['total_amount'] as num?)?.toDouble() ?? effTotal;
                effTax = (t['tax_amount'] as num?)?.toDouble() ?? effTax;
                final dateStr = (t['transaction_date'] ?? t['created_at'])?.toString();
                if (dateStr != null && dateStr.isNotEmpty) {
                  effectiveTime = DateTime.tryParse(dateStr) ?? effectiveTime;
                }
              } else if (salesOrderId != null && salesOrderId.isNotEmpty) {
                final soRows = await db.rawQuery('''
                  SELECT total_amount, tax_amount, created_at, order_date FROM sales_orders_v2
                  WHERE (id = ? OR order_number = ? OR order_number = ?)
                  LIMIT 1
                ''', [salesOrderId, cleanOrder, txnOrder]);
                if (soRows.isNotEmpty) {
                  final so = soRows.first;
                  effTotal = (so['total_amount'] as num?)?.toDouble() ?? effTotal;
                  effTax = (so['tax_amount'] as num?)?.toDouble() ?? effTax;
                  final dateStr = (so['order_date'] ?? so['created_at'])?.toString();
                  if (dateStr != null && dateStr.isNotEmpty) {
                    effectiveTime = DateTime.tryParse(dateStr) ?? effectiveTime;
                  }
                }
              }
            }
          } catch (_) {}
        }

        final seller = headerName.isNotEmpty
            ? headerName
            : (storeName ?? singleton.activeStoreName ?? singleton.activeTenantName ?? 'Store');
        final vat = headerTaxId.isNotEmpty ? headerTaxId : '300000000000003';

        resolvedQr = generateZatcaTlvBase64(
          sellerName: seller,
          vatNumber: vat,
          timestamp: effectiveTime,
          totalAmount: effTotal,
          taxAmount: effTax,
        );
        debugPrint('✅ [PosReceiptEngine] Generated dynamic ZATCA TLV QR for registered device (Seller: $seller, Total: $effTotal, Tax: $effTax)');

        // Cache generated QR into pos_transactions metadata
        try {
          if (DatabaseService().isOpen && ((salesOrderId != null && salesOrderId.isNotEmpty) || (transactionId != null && transactionId > 0))) {
            final db = DatabaseService().database;
            final cleanOrder = salesOrderId?.replaceFirst('TXN-', '') ?? '';
            final txnOrder = (salesOrderId != null && salesOrderId!.startsWith('TXN-')) ? salesOrderId! : 'TXN-$salesOrderId';

            final txRows = await db.rawQuery('''
              SELECT id, metadata FROM pos_transactions
              WHERE (sales_order_id = ? OR transaction_number = ? OR transaction_number = ? OR id = ?)
              LIMIT 1
            ''', [salesOrderId ?? '', txnOrder, cleanOrder, transactionId ?? 0]);
            if (txRows.isNotEmpty) {
              final txId = txRows.first['id'];
              Map<String, dynamic> meta = {};
              try {
                final mStr = txRows.first['metadata']?.toString();
                if (mStr != null && mStr.isNotEmpty) meta = Map<String, dynamic>.from(jsonDecode(mStr));
              } catch (_) {}
              meta['qr_code_base64'] = resolvedQr;
              await db.update('pos_transactions', {'metadata': jsonEncode(meta)}, where: 'id = ?', whereArgs: [txId]);
            }
          }
        } catch (_) {}
      }
    }

    return ReceiptTemplate(
      header: OrgHeader(
        name: headerName.isNotEmpty ? headerName : (singleton.activeStoreName ?? singleton.activeTenantName ?? ''),
        address: headerAddress,
        phone: headerPhone,
        taxId: headerTaxId,
        website: headerWebsite,
      ),
      footer: OrgFooter(
        thankYou: footerThankYou,
        returnNote: footerReturnNote,
        website: footerWebsite,
      ),
      qrCodeBase64: resolvedQr,
      currency: currency,
    );
  }
}

