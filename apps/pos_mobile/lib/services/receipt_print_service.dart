import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart' show rootBundle;
import 'package:path_provider/path_provider.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';
import 'package:pos_mobile/services/pos_receipt_engine.dart';
import 'package:pos_mobile/singleton/singleton_class.dart';

/// Comprehensive receipt printing service:
/// 1. Uses POS-client `ReceiptTemplate` & `PosReceiptData` parsing & business logic.
/// 2. Dispatches print business logic to the pos_client Go backend handler (POST /api/print/receipt).
/// 3. Launches the OS System Print Dialog directly with dynamic page sizing (US Letter, A4, Roll80).
class ReceiptPrintService {
  static final ReceiptPrintService _instance = ReceiptPrintService._internal();
  factory ReceiptPrintService() => _instance;
  ReceiptPrintService._internal();

  /// Main entry point to trigger printing:
  /// - Dispatches pos_client handler (localhost:8080/api/print/receipt) if running.
  /// - Directly opens the System Print Dialog on tap with dynamic layout negotiation.
  Future<Map<String, dynamic>> printReceipt({
    PosReceiptData? receiptData,
    ReceiptTemplate? template,
    String? storeName,
    String? orderNumber,
    String? cashierName,
    String? terminalName,
    String? customerName,
    String? paymentMethod,
    List<Map<String, dynamic>>? cartItems,
    double? subtotal,
    double discount = 0.0,
    double taxAmount = 0.0,
    double totalAmount = 0.0,
    double changeDue = 0.0,
    int? orgId,
  }) async {
    // 1. Parse into PosReceiptData & ReceiptTemplate using pos-client business logic
    final data = receiptData ??
        PosReceiptData.parse(
          receiptNumber: orderNumber ?? 'TXN-${DateTime.now().millisecondsSinceEpoch}',
          cashier: cashierName ?? (SingletonClass().activeCashierCode ?? 'Cashier'),
          terminal: terminalName ?? (SingletonClass().activeTerminalName ?? 'T-01'),
          customer: customerName ?? 'Walk-in Customer',
          rawItems: cartItems ?? [],
          rawSubtotal: subtotal,
          discount: discount,
          rawTaxAmount: taxAmount,
          rawTotalAmount: totalAmount,
          changeDue: changeDue,
          paymentMethod: paymentMethod ?? 'CASH',
        );

    final tmpl = template ??
        ReceiptTemplate.fromSession(
          storeName: storeName,
        );

    // Open the System Print Dialog directly on UI
    try {
      final printed = await Printing.layoutPdf(
        name: 'Receipt_${data.receiptNumber}',
        format: PdfPageFormat.roll80,
        onLayout: (PdfPageFormat format) async {
          return generateReceiptPdf(
            pageFormat: format,
            data: data,
            template: tmpl,
          );
        },
        dynamicLayout: true,
      );

      return {
        'success': true,
        'printed': printed,
        'message': 'System print dialog opened successfully',
      };
    } catch (e) {
      debugPrint('Printing.layoutPdf error (Falling back to OS Print): $e');
      final fallbackRes = await _printViaOsHtmlFallback(
        data: data,
        template: tmpl,
      );

      return fallbackRes;
    }
  }

  static bool _hasArabic(String text) {
    return RegExp(r'[\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\uFB50-\uFDFF\uFE70-\uFEFF]').hasMatch(text);
  }

  /// Generates the PDF document matching the exact pos-client printing template
  Future<Uint8List> generateReceiptPdf({
    PdfPageFormat pageFormat = PdfPageFormat.roll80,
    PosReceiptData? data,
    ReceiptTemplate? template,
    String? storeName,
    String? orderNumber,
    String? cashierName,
    String? terminalName,
    String? customerName,
    String? paymentMethod,
    List<Map<String, dynamic>>? cartItems,
    double? subtotal,
    double discount = 0.0,
    double taxAmount = 0.0,
    double totalAmount = 0.0,
    double changeDue = 0.0,
  }) async {
    final effectiveData = data ??
        PosReceiptData.parse(
          receiptNumber: orderNumber ?? 'TXN-${DateTime.now().millisecondsSinceEpoch}',
          cashier: cashierName ?? 'Cashier',
          terminal: terminalName ?? 'T-01',
          customer: customerName ?? 'Walk-in Customer',
          rawItems: cartItems ?? [],
          rawSubtotal: subtotal,
          discount: discount,
          rawTaxAmount: taxAmount,
          rawTotalAmount: totalAmount,
          changeDue: changeDue,
          paymentMethod: paymentMethod ?? 'CASH',
        );

    final effectiveTmpl = template ??
        ReceiptTemplate.fromSession(
          storeName: storeName,
        );

    pw.Font fontRegular;
    pw.Font fontBold;

    try {
      final fontDataRegular = await rootBundle.load('assets/fonts/Amiri-Regular.ttf');
      final fontDataBold = await rootBundle.load('assets/fonts/Amiri-Bold.ttf');
      fontRegular = pw.Font.ttf(fontDataRegular);
      fontBold = pw.Font.ttf(fontDataBold);
    } catch (e) {
      debugPrint('Failed to load asset font Amiri ($e), trying Cairo or GoogleFonts...');
      try {
        final fontDataRegular = await rootBundle.load('assets/fonts/Cairo-Regular.ttf');
        final fontDataBold = await rootBundle.load('assets/fonts/Cairo-Bold.ttf');
        fontRegular = pw.Font.ttf(fontDataRegular);
        fontBold = pw.Font.ttf(fontDataBold);
      } catch (_) {
        try {
          fontRegular = await PdfGoogleFonts.cairoRegular();
          fontBold = await PdfGoogleFonts.cairoBold();
        } catch (_) {
          fontRegular = pw.Font.courier();
          fontBold = pw.Font.courierBold();
        }
      }
    }

    final doc = pw.Document(
      theme: pw.ThemeData.withFont(
        base: fontRegular,
        bold: fontBold,
      ),
    );

    const blackColor = PdfColors.black;
    final bool isWidePage = pageFormat.width > 300;

    final receiptContent = pw.Container(
      width: isWidePage ? 260 : null,
      color: PdfColors.white,
      padding: const pw.EdgeInsets.symmetric(horizontal: 4, vertical: 6),
      child: pw.Column(
        crossAxisAlignment: pw.CrossAxisAlignment.stretch,
        mainAxisSize: pw.MainAxisSize.min,
        children: [
          // ── HEADER ────────────────────────────────────────────────────────
          pw.Text(
            effectiveTmpl.header.name.toUpperCase(),
            textAlign: pw.TextAlign.center,
            textDirection: _hasArabic(effectiveTmpl.header.name) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
            style: pw.TextStyle(
              font: fontBold,
              fontSize: 14,
              color: blackColor,
            ),
          ),
          pw.SizedBox(height: 2),
          if (effectiveTmpl.header.address.isNotEmpty) ...[
            pw.Text(
              effectiveTmpl.header.address,
              textAlign: pw.TextAlign.center,
              textDirection: _hasArabic(effectiveTmpl.header.address) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
              style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: blackColor),
            ),
            pw.SizedBox(height: 2),
          ],
          if (effectiveTmpl.header.phone.isNotEmpty) ...[
            pw.Text(
              'Tel: ${effectiveTmpl.header.phone}',
              textAlign: pw.TextAlign.center,
              style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: blackColor),
            ),
            pw.SizedBox(height: 2),
          ],
          if (effectiveTmpl.header.taxId.isNotEmpty) ...[
            pw.Text(
              'TIN: ${effectiveTmpl.header.taxId}',
              textAlign: pw.TextAlign.center,
              style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: blackColor),
            ),
            pw.SizedBox(height: 2),
          ],
          pw.Text(
            'Point of Sale System',
            textAlign: pw.TextAlign.center,
            style: pw.TextStyle(
              font: fontRegular,
              fontSize: 8.5,
              color: blackColor,
            ),
          ),
          pw.SizedBox(height: 4),

          // Dashed Line
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 4),

          // ── RECEIPT METADATA ──────────────────────────────────────────────
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              pw.Text(
                'Receipt #: ${effectiveData.receiptNumber}',
                style: pw.TextStyle(font: fontBold, fontSize: 9.0, color: blackColor),
              ),
              pw.Text(
                effectiveData.formattedDateTime,
                style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: blackColor),
              ),
            ],
          ),
          pw.SizedBox(height: 2),
          pw.Text(
            'Cashier: ${effectiveData.cashier} • Terminal: ${effectiveData.terminal}',
            textAlign: pw.TextAlign.left,
            textDirection: (_hasArabic(effectiveData.cashier) || _hasArabic(effectiveData.terminal))
                ? pw.TextDirection.rtl
                : pw.TextDirection.ltr,
            style: pw.TextStyle(
              font: fontRegular,
              fontSize: 8.5,
              color: blackColor,
            ),
          ),
          pw.SizedBox(height: 2),
          pw.Text(
            'Customer: ${effectiveData.customer}',
            textAlign: pw.TextAlign.left,
            textDirection: _hasArabic(effectiveData.customer) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
            style: pw.TextStyle(
              font: fontBold,
              fontSize: 9.0,
              color: blackColor,
            ),
          ),
          pw.SizedBox(height: 5),

          // Dashed Line
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 4),

          // ── COLUMN HEADERS ────────────────────────────────────────────────
          pw.Row(
            children: [
              pw.Expanded(
                flex: 5,
                child: pw.Text('Item', style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: blackColor)),
              ),
              pw.Expanded(
                flex: 2,
                child: pw.Text('Qty', textAlign: pw.TextAlign.center, style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: blackColor)),
              ),
              pw.Expanded(
                flex: 2,
                child: pw.Text('Price', textAlign: pw.TextAlign.right, style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: blackColor)),
              ),
              pw.Expanded(
                flex: 2,
                child: pw.Text('Total', textAlign: pw.TextAlign.right, style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: blackColor)),
              ),
            ],
          ),
          pw.SizedBox(height: 3),
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 4),

          // ── ITEMS LIST ────────────────────────────────────────────────────
          ...List.generate(effectiveData.items.length, (idx) {
            final item = effectiveData.items[idx];
            return pw.Padding(
              padding: const pw.EdgeInsets.only(bottom: 3.5),
              child: pw.Row(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                children: [
                  pw.Expanded(
                    flex: 5,
                    child: pw.Text(
                      '${idx + 1}. ${item.name}',
                      textDirection: _hasArabic(item.name) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
                      style: pw.TextStyle(
                        font: fontBold,
                        fontSize: 9.0,
                        color: blackColor,
                      ),
                    ),
                  ),
                  pw.Expanded(
                    flex: 2,
                    child: pw.Text(
                      item.qty.toStringAsFixed(0),
                      textAlign: pw.TextAlign.center,
                      style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: blackColor),
                    ),
                  ),
                  pw.Expanded(
                    flex: 2,
                    child: pw.Text(
                      item.price.toStringAsFixed(2),
                      textAlign: pw.TextAlign.right,
                      style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: blackColor),
                    ),
                  ),
                  pw.Expanded(
                    flex: 2,
                    child: pw.Text(
                      item.lineTotal.toStringAsFixed(2),
                      textAlign: pw.TextAlign.right,
                      style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: blackColor),
                    ),
                  ),
                ],
              ),
            );
          }),

          pw.SizedBox(height: 4),
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 4),

          // Total items count
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              pw.Text(
                'Total Items Count:',
                style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: blackColor),
              ),
              pw.Text(
                '${effectiveData.items.length} items',
                style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: blackColor),
              ),
            ],
          ),
          pw.SizedBox(height: 4),
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 4),

          // ── TOTALS (Business Logic) ───────────────────────────────────────
          _buildPdfSummaryRow('Subtotal', '${effectiveData.currency} ${effectiveData.subtotal.toStringAsFixed(2)}', fontRegular, fontBold, blackColor, blackColor),
          if (effectiveData.discount > 0)
            _buildPdfSummaryRow('Discount (${effectiveData.discountPercentage.toStringAsFixed(0)}%)', '- ${effectiveData.currency} ${effectiveData.discount.toStringAsFixed(2)}', fontRegular, fontBold, blackColor, blackColor),
          _buildPdfSummaryRow('VAT (${(effectiveData.taxRate * 100).toStringAsFixed(0)}%)', '${effectiveData.currency} ${effectiveData.taxAmount.toStringAsFixed(2)}', fontRegular, fontBold, blackColor, blackColor),
          pw.SizedBox(height: 4),

          // Solid Line
          pw.Container(height: 1.0, color: blackColor),
          pw.SizedBox(height: 4),

          // Grand Total
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              pw.Text(
                'TOTAL',
                style: pw.TextStyle(font: fontBold, fontSize: 11.5, color: blackColor),
              ),
              pw.Text(
                '${effectiveData.currency} ${effectiveData.grandTotal.toStringAsFixed(2)}',
                style: pw.TextStyle(font: fontBold, fontSize: 11.5, color: blackColor),
              ),
            ],
          ),
          pw.SizedBox(height: 4),
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 4),

          // Cash Paid & Change
          _buildPdfSummaryRow('Cash Paid:', '${effectiveData.currency} ${effectiveData.paid.toStringAsFixed(2)}', fontRegular, fontBold, blackColor, blackColor),
          _buildPdfSummaryRow('Change Due:', '${effectiveData.currency} ${effectiveData.changeDue.toStringAsFixed(2)}', fontRegular, fontBold, blackColor, blackColor),
          pw.SizedBox(height: 4),
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 4),

          // Payment Method
          _buildPdfSummaryRow('Payment Method:', effectiveData.paymentMethod, fontRegular, fontBold, blackColor, blackColor),
          pw.SizedBox(height: 4),
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 6),

          // ── FOOTER ────────────────────────────────────────────────────────
          if (effectiveTmpl.footer.thankYou.isNotEmpty) ...[
            pw.Text(
              '** ${effectiveTmpl.footer.thankYou} **',
              textAlign: pw.TextAlign.center,
              textDirection: _hasArabic(effectiveTmpl.footer.thankYou) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
              style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: blackColor),
            ),
            pw.SizedBox(height: 2),
          ],
          if (effectiveTmpl.footer.returnNote.isNotEmpty) ...[
            pw.Text(
              effectiveTmpl.footer.returnNote,
              textAlign: pw.TextAlign.center,
              textDirection: _hasArabic(effectiveTmpl.footer.returnNote) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
              style: pw.TextStyle(font: fontRegular, fontSize: 8.0, color: blackColor),
            ),
            pw.SizedBox(height: 2),
          ],
          if (effectiveTmpl.footer.website.isNotEmpty) ...[
            pw.Text(
              effectiveTmpl.footer.website,
              textAlign: pw.TextAlign.center,
              style: pw.TextStyle(font: fontRegular, fontSize: 8.0, color: blackColor),
            ),
            pw.SizedBox(height: 4),
          ],
          _buildPdfDashedDivider(blackColor),
          pw.SizedBox(height: 5),

          // Barcode representation
          pw.Text(
            '||| | |||| || ||| |||| ||||',
            textAlign: pw.TextAlign.center,
            style: pw.TextStyle(
              font: fontBold,
              fontSize: 13,
              color: blackColor,
              letterSpacing: 2,
            ),
          ),
          pw.SizedBox(height: 2),
          pw.Text(
            effectiveData.barcode,
            textAlign: pw.TextAlign.center,
            style: pw.TextStyle(font: fontRegular, fontSize: 7.5, color: blackColor),
          ),
        ],
      ),
    );

    doc.addPage(
      pw.Page(
        pageFormat: pageFormat,
        margin: isWidePage
            ? const pw.EdgeInsets.symmetric(horizontal: 20, vertical: 24)
            : const pw.EdgeInsets.symmetric(horizontal: 4, vertical: 6),
        build: (pw.Context context) {
          if (isWidePage) {
            return pw.Align(
              alignment: pw.Alignment.topCenter,
              child: receiptContent,
            );
          }
          return receiptContent;
        },
      ),
    );

    return doc.save();
  }

  pw.Widget _buildPdfSummaryRow(
    String label,
    String value,
    pw.Font fontRegular,
    pw.Font fontBold,
    PdfColor labelColor,
    PdfColor valueColor,
  ) {
    return pw.Padding(
      padding: const pw.EdgeInsets.symmetric(vertical: 1.2),
      child: pw.Row(
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        children: [
          pw.Text(
            label,
            textDirection: _hasArabic(label) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
            style: pw.TextStyle(font: fontRegular, fontSize: 8.5, color: labelColor),
          ),
          pw.Text(
            value,
            textDirection: _hasArabic(value) ? pw.TextDirection.rtl : pw.TextDirection.ltr,
            style: pw.TextStyle(font: fontBold, fontSize: 8.5, color: valueColor),
          ),
        ],
      ),
    );
  }

  pw.Widget _buildPdfDashedDivider(PdfColor color) {
    return pw.Container(
      height: 1,
      child: pw.Row(
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        children: List.generate(
          28,
          (_) => pw.Container(
            width: 3.5,
            height: 1,
            color: color,
          ),
        ),
      ),
    );
  }

  /// Native OS HTML Print Fallback
  Future<Map<String, dynamic>> _printViaOsHtmlFallback({
    required PosReceiptData data,
    required ReceiptTemplate template,
  }) async {
    try {
      final html = _generateReceiptHtml(data: data, template: template);
      final tempDir = await getTemporaryDirectory();
      final sanitizedOrder = data.receiptNumber.replaceAll(RegExp(r'[^a-zA-Z0-9]'), '_');
      final tempFile = File('${tempDir.path}/receipt_$sanitizedOrder.html');
      await tempFile.writeAsString(html);

      if (Platform.isMacOS) {
        await Process.run('open', [tempFile.path]);
      } else if (Platform.isLinux) {
        await Process.run('xdg-open', [tempFile.path]);
      } else if (Platform.isWindows) {
        await Process.run('cmd', ['/c', 'start', tempFile.path]);
      }

      return {
        'success': true,
        'message': 'System print dialog opened',
      };
    } catch (e) {
      return {
        'success': false,
        'error': 'Could not launch print dialog: $e',
      };
    }
  }

  String _generateReceiptHtml({
    required PosReceiptData data,
    required ReceiptTemplate template,
  }) {
    final buffer = StringBuffer();
    int idx = 1;
    for (final item in data.items) {
      buffer.write('''
        <div style="margin-bottom: 6px;">
          <div style="display: flex; justify-content: space-between; font-weight: bold; font-size: 12px; color: #0f172a;">
            <span>$idx. ${item.name}</span>
            <span>${data.currency} ${item.lineTotal.toStringAsFixed(2)}</span>
          </div>
          <div style="font-size: 10px; color: #64748b; margin-left: 8px;">
            ${item.qty.toStringAsFixed(0)} x ${data.currency} ${item.price.toStringAsFixed(2)}
          </div>
        </div>
      ''');
      idx++;
    }

    return '''
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Receipt Preview - ${data.receiptNumber}</title>
  <style>
    @page { margin: 0; size: 80mm auto; }
    @media print {
      body { margin: 0; padding: 4mm; background: #fff; }
      .no-print { display: none !important; }
      .receipt-container { box-shadow: none !important; }
    }
    body {
      font-family: 'Courier New', Courier, monospace;
      margin: 0; padding: 16px; background: #f1f5f9;
      display: flex; justify-content: center; align-items: center; min-height: 100vh;
    }
    .receipt-container {
      width: 78mm; background: #fffdf9; border: 2px dashed #facc15;
      border-radius: 16px; padding: 20px 16px; box-shadow: 0 10px 25px rgba(0,0,0,0.15);
      color: #0f172a; box-sizing: border-box;
    }
    .header { text-align: center; margin-bottom: 12px; }
    .store-name { font-size: 16px; font-weight: 900; letter-spacing: 1px; margin-bottom: 4px; text-transform: uppercase; }
    .subtitle { font-size: 11px; color: #475569; margin: 2px 0; }
    .dashed-divider { border-top: 1.5px dashed #facc15; margin: 10px 0; }
    .solid-divider { border-top: 2px solid #facc15; margin: 8px 0; }
    .summary-row { display: flex; justify-content: space-between; font-size: 11.5px; margin: 3px 0; color: #334155; }
    .total-row { display: flex; justify-content: space-between; font-size: 14.5px; font-weight: 900; color: #0f172a; margin: 6px 0; }
    .footer { text-align: center; margin-top: 14px; font-size: 11px; color: #475569; }
    .barcode { text-align: center; font-size: 17px; font-weight: 900; letter-spacing: 3px; margin-top: 8px; }
  </style>
</head>
<body onload="window.print()">
  <div class="receipt-container">
    <div class="header">
      <div class="store-name">${template.header.name.toUpperCase()}</div>
      ${template.header.address.isNotEmpty ? '<div class="subtitle">' + template.header.address + '</div>' : ''}
      ${template.header.phone.isNotEmpty ? '<div class="subtitle">Tel: ' + template.header.phone + '</div>' : ''}
      ${template.header.taxId.isNotEmpty ? '<div class="subtitle">TIN: ' + template.header.taxId + '</div>' : ''}
      <div class="subtitle">Point of Sale System</div>
      <div class="subtitle" style="font-weight: bold; color: #1e293b; margin-top: 4px;">Receipt #: ${data.receiptNumber}</div>
      <div class="subtitle">${data.formattedDateTime}</div>
      <div class="subtitle">Cashier: ${data.cashier} • Terminal: ${data.terminal}</div>
      <div class="subtitle" style="font-weight: bold; color: #1e293b;">Customer: ${data.customer}</div>
    </div>
    <div class="dashed-divider"></div>
    ${buffer.toString()}
    <div class="dashed-divider"></div>
    <div class="summary-row" style="font-weight: bold;">
      <span>Total Items Count:</span>
      <span>${data.items.length} items</span>
    </div>
    <div class="dashed-divider"></div>
    <div class="summary-row">
      <span>Subtotal</span>
      <span>${data.currency} ${data.subtotal.toStringAsFixed(2)}</span>
    </div>
    ${data.discount > 0 ? '<div class="summary-row"><span>Discount (' + data.discountPercentage.toStringAsFixed(0) + '%)</span><span>- ' + data.currency + ' ' + data.discount.toStringAsFixed(2) + '</span></div>' : ''}
    <div class="summary-row">
      <span>VAT (${(data.taxRate * 100).toStringAsFixed(0)}%)</span>
      <span>${data.currency} ${data.taxAmount.toStringAsFixed(2)}</span>
    </div>
    <div class="solid-divider"></div>
    <div class="total-row">
      <span>TOTAL</span>
      <span>${data.currency} ${data.grandTotal.toStringAsFixed(2)}</span>
    </div>
    <div class="dashed-divider"></div>
    <div class="summary-row">
      <span>Cash Paid</span>
      <span>${data.currency} ${data.paid.toStringAsFixed(2)}</span>
    </div>
    <div class="summary-row">
      <span>Change</span>
      <span>${data.currency} ${data.changeDue.toStringAsFixed(2)}</span>
    </div>
    <div class="dashed-divider"></div>
    <div class="summary-row">
      <span>Payment Method:</span>
      <span style="font-weight: bold;">${data.paymentMethod}</span>
    </div>
    <div class="dashed-divider"></div>
    <div class="footer">
      <div>${template.footer.thankYou}</div>
      <div style="font-size: 9.5px; margin-top: 2px;">${template.footer.returnNote}</div>
      <div class="barcode">||| | |||| || ||| |||| ||||</div>
      <div style="font-size: 9px; margin-top: 2px;">${data.barcode}</div>
    </div>
  </div>
</body>
</html>
''';
  }
}
