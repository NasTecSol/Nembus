import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:pos_mobile/singleton/singleton_class.dart';

/// ─────────────────────────────────────────────────────────────────────────────
/// POS-CLIENT PRINTING TEMPLATE & DATA ENGINE
/// Mirrors `apps/pos-client/internal/printing/template.go` & `packages/core`
/// ─────────────────────────────────────────────────────────────────────────────

/// Organisation details printed at the top of every receipt.
class OrgHeader {
  final String name;
  final String address;
  final String phone;
  final String taxId;
  final String website;

  const OrgHeader({
    this.name = 'QITAF AL AYELA',
    this.address = '',
    this.phone = '',
    this.taxId = '',
    this.website = '',
  });

  factory OrgHeader.fromJson(Map<String, dynamic> json) {
    return OrgHeader(
      name: json['name'] as String? ?? 'QITAF AL AYELA',
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

/// Sign-off lines printed at the bottom of every receipt.
class OrgFooter {
  final String thankYou;
  final String returnNote;
  final String website;

  const OrgFooter({
    this.thankYou = 'Thank you for your visit!',
    this.returnNote = 'Please keep this receipt for any exchange or return.',
    this.website = '',
  });

  factory OrgFooter.fromJson(Map<String, dynamic> json) {
    return OrgFooter(
      thankYou: json['thank_you'] as String? ?? 'Thank you for your visit!',
      returnNote: json['return_note'] as String? ?? 'Please keep this receipt for any exchange or return.',
      website: json['website'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
    'thank_you': thankYou,
    'return_note': returnNote,
    'website': website,
  };
}

/// Combines header + footer branding for one organisation.
class ReceiptTemplate {
  final OrgHeader header;
  final OrgFooter footer;

  const ReceiptTemplate({
    required this.header,
    required this.footer,
  });

  factory ReceiptTemplate.fromJson(Map<String, dynamic> json) {
    return ReceiptTemplate(
      header: OrgHeader.fromJson(json['header'] as Map<String, dynamic>? ?? {}),
      footer: OrgFooter.fromJson(json['footer'] as Map<String, dynamic>? ?? {}),
    );
  }

  Map<String, dynamic> toJson() => {
    'header': header.toJson(),
    'footer': footer.toJson(),
  };

  /// Builds a default template from current application session / singleton state
  factory ReceiptTemplate.fromSession({
    String? storeName,
    String? address,
    String? phone,
    String? taxId,
    String? website,
    String? thankYou,
    String? returnNote,
  }) {
    final singleton = SingletonClass();
    final effectiveStoreName = storeName ??
        singleton.activeStoreName ??
        singleton.activeTenantName ??
        'QITAF AL AYELA';

    return ReceiptTemplate(
      header: OrgHeader(
        name: effectiveStoreName,
        address: address ?? '',
        phone: phone ?? '',
        taxId: taxId ?? '300000000000003',
        website: website ?? '',
      ),
      footer: OrgFooter(
        thankYou: thankYou ?? 'Thank you for shopping with us!',
        returnNote: returnNote ?? 'Please keep this receipt for any exchange or return.',
        website: website ?? '',
      ),
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
    'currency': currency,
  };
}

/// Engine to interact with pos-client backend handler & template logic
class PosReceiptEngine {
  /// Resolves the local pos-client server base URL based on running platform
  static String get posClientBaseUrl {
    if (!kIsWeb && defaultTargetPlatform == TargetPlatform.android) {
      return 'http://10.0.2.2:8080';
    }
    return 'http://127.0.0.1:8080';
  }

  /// Dispatches the receipt payload to pos-client `POST /api/print/receipt`
  static Future<bool> sendToPosClient({
    required PosReceiptData receiptData,
    int? orgId,
    String printerMode = 'system',
    String printerName = 'POS-Receipt',
  }) async {
    final effectiveOrgId = orgId ??
        int.tryParse(SingletonClass().activeTenantId ?? '1') ??
        1;

    final payload = {
      'org_id': effectiveOrgId,
      'printer': {
        'mode': printerMode,
        'printer_name': printerName,
      },
      'receipt': receiptData.toJson(),
    };

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
          .timeout(const Duration(milliseconds: 800));

      if (response.statusCode == 200 || response.statusCode == 201) {
        debugPrint('✅ Receipt dispatched successfully to pos-client print handler');
        return true;
      } else {
        debugPrint('⚠️ pos-client print handler returned ${response.statusCode}: ${response.body}');
        return false;
      }
    } catch (_) {
      debugPrint('ℹ️ pos-client backend handler offline at $baseUrl - using Flutter template pipeline.');
      return false;
    }
  }
}
