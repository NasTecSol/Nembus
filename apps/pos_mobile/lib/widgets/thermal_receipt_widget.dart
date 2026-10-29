import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:pos_mobile/services/pos_receipt_engine.dart';
import 'package:pos_mobile/services/receipt_print_service.dart';

/// Custom Painter to draw a rounded dashed/dotted gold border for thermal receipts
class DashedRoundedBorderPainter extends CustomPainter {
  final Color color;
  final double strokeWidth;
  final double dashWidth;
  final double dashSpace;
  final double borderRadius;

  const DashedRoundedBorderPainter({
    this.color = const Color(0xFFFACC15),
    this.strokeWidth = 1.6,
    this.dashWidth = 5.0,
    this.dashSpace = 4.0,
    this.borderRadius = 16.0,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color
      ..strokeWidth = strokeWidth
      ..style = PaintingStyle.stroke;

    final rrect = RRect.fromRectAndRadius(
      Rect.fromLTWH(0, 0, size.width, size.height),
      Radius.circular(borderRadius),
    );

    final path = Path()..addRRect(rrect);
    final metrics = path.computeMetrics();

    for (final metric in metrics) {
      double distance = 0.0;
      while (distance < metric.length) {
        final length = math.min(dashWidth, metric.length - distance);
        final extractPath = metric.extractPath(distance, distance + length);
        canvas.drawPath(extractPath, paint);
        distance += dashWidth + dashSpace;
      }
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

/// 3D Animated Thermal Receipt Preview Widget matching pos-client template.go & reference UI
class ThermalReceiptWidget extends StatefulWidget {
  final PosReceiptData? receiptData;
  final ReceiptTemplate? template;

  final String? storeName;
  final String? orderNumber;
  final String cashierName;
  final String terminalName;
  final String customerName;
  final String paymentMethod;
  final List<dynamic> cartItems;
  final double subtotal;
  final double discount;
  final double taxAmount;
  final double totalAmount;
  final double changeDue;
  final VoidCallback onProceedWithoutPrint;
  final VoidCallback? onBack;

  const ThermalReceiptWidget({
    super.key,
    this.receiptData,
    this.template,
    this.storeName,
    this.orderNumber,
    this.cashierName = 'Cashier',
    this.terminalName = 'T-01',
    this.customerName = 'Walk-in Customer',
    this.paymentMethod = 'CASH',
    this.cartItems = const [],
    this.subtotal = 0.0,
    this.discount = 0.0,
    this.taxAmount = 0.0,
    this.totalAmount = 0.0,
    this.changeDue = 0.0,
    required this.onProceedWithoutPrint,
    this.onBack,
  });

  @override
  State<ThermalReceiptWidget> createState() => _ThermalReceiptWidgetState();
}

class _ThermalReceiptWidgetState extends State<ThermalReceiptWidget> with SingleTickerProviderStateMixin {
  late AnimationController _animController;
  late Animation<double> _slideAnimation;
  late Animation<double> _fadeAnimation;
  late Animation<double> _perspectiveAnimation;

  late PosReceiptData _receiptData;
  late ReceiptTemplate _template;

  bool _isPrinting = false;

  @override
  void initState() {
    super.initState();

    // 1. Parse into PosReceiptData & ReceiptTemplate using pos-client business logic
    _receiptData = widget.receiptData ??
        PosReceiptData.parse(
          receiptNumber: widget.orderNumber ?? 'TXN-${DateTime.now().millisecondsSinceEpoch}',
          cashier: widget.cashierName,
          terminal: widget.terminalName,
          customer: widget.customerName,
          rawItems: widget.cartItems,
          rawSubtotal: widget.subtotal,
          discount: widget.discount,
          rawTaxAmount: widget.taxAmount,
          rawTotalAmount: widget.totalAmount,
          changeDue: widget.changeDue,
          paymentMethod: widget.paymentMethod,
        );

    _template = widget.template ??
        ReceiptTemplate.fromSession(
          storeName: widget.storeName,
        );

    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 950),
    );

    // Smooth top-to-bottom roll out animation
    _slideAnimation = CurvedAnimation(
      parent: _animController,
      curve: Curves.easeOutCubic,
    );

    _fadeAnimation = CurvedAnimation(
      parent: _animController,
      curve: const Interval(0.0, 0.65, curve: Curves.easeIn),
    );

    _perspectiveAnimation = Tween<double>(begin: -0.22, end: 0.0).animate(
      CurvedAnimation(
        parent: _animController,
        curve: Curves.easeOutCubic,
      ),
    );

    _animController.forward();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  /// Direct print trigger: opens system print dialog immediately with zero network delay
  Future<void> _handlePrintReceipt() async {
    setState(() => _isPrinting = true);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        backgroundColor: const Color(0xFF1E293B),
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 2),
        content: Row(
          children: const [
            SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(strokeWidth: 2, color: Color(0xFF38BDF8)),
            ),
            SizedBox(width: 12),
            Text(
              'Opening system print dialog...',
              style: TextStyle(color: Colors.white, fontSize: 12.5),
            ),
          ],
        ),
      ),
    );

    final res = await ReceiptPrintService().printReceipt(
      receiptData: _receiptData,
      template: _template,
    );

    if (mounted) {
      setState(() => _isPrinting = false);
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      if (res['success'] != true && res['error'] != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFFDC2626),
            behavior: SnackBarBehavior.floating,
            content: Text('Print Error: ${res['error']}'),
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    const sheetBg = Color(0xFF0B0F19);
    const cardBg = Color(0xFF161E2E);
    const goldAccent = Color(0xFFFACC15);

    final itemCount = _receiptData.items.length;

    return Container(
      color: sheetBg,
      child: Column(
        children: [
          // Top Header Bar
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: [
                InkWell(
                  onTap: () {
                    if (widget.onBack != null) {
                      widget.onBack!();
                    } else {
                      widget.onProceedWithoutPrint();
                    }
                  },
                  borderRadius: BorderRadius.circular(12),
                  child: Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: cardBg,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.white12),
                    ),
                    child: const Icon(Icons.arrow_back_rounded, color: Colors.white, size: 18),
                  ),
                ),
                const SizedBox(width: 14),
                const Text('🧾', style: TextStyle(fontSize: 20)),
                const SizedBox(width: 8),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Text(
                          'Receipt Preview',
                          style: GoogleFonts.inter(
                            fontSize: 16,
                            fontWeight: FontWeight.w800,
                            color: Colors.white,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.12),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Text(
                            '$itemCount items',
                            style: GoogleFonts.inter(
                              color: Colors.white70,
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 2),
                    Text(
                      'Continuous thermal roll preview',
                      style: GoogleFonts.inter(
                        fontSize: 11,
                        color: Colors.white54,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const Divider(color: Colors.white10, height: 1),

          // Printer Output Slot Visual Indicator
          Container(
            height: 6,
            width: double.infinity,
            margin: const EdgeInsets.symmetric(horizontal: 40),
            decoration: BoxDecoration(
              color: Colors.black,
              borderRadius: BorderRadius.circular(3),
              boxShadow: const [
                BoxShadow(color: Colors.black54, blurRadius: 4, offset: Offset(0, 2)),
              ],
            ),
          ),

          // Animated Receipt Scrollable Body (Top to Bottom transition)
          Expanded(
            child: SingleChildScrollView(
              physics: const BouncingScrollPhysics(),
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
              child: AnimatedBuilder(
                animation: _animController,
                builder: (context, child) {
                  return SlideTransition(
                    position: Tween<Offset>(
                      begin: const Offset(0, -0.25),
                      end: Offset.zero,
                    ).animate(_slideAnimation),
                    child: FadeTransition(
                      opacity: _fadeAnimation,
                      child: Transform(
                        alignment: Alignment.topCenter,
                        transform: Matrix4.identity()
                          ..setEntry(3, 2, 0.001)
                          ..rotateX(_perspectiveAnimation.value),
                        child: child,
                      ),
                    ),
                  );
                },
                child: Center(
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 480),
                    child: CustomPaint(
                      painter: const DashedRoundedBorderPainter(
                        color: goldAccent,
                        strokeWidth: 1.6,
                        dashWidth: 5.5,
                        dashSpace: 4.5,
                        borderRadius: 18,
                      ),
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 26),
                        decoration: BoxDecoration(
                          color: const Color(0xFFFFFDF9), // Authentic thermal receipt paper off-white
                          borderRadius: BorderRadius.circular(18),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withValues(alpha: 0.45),
                              blurRadius: 24,
                              offset: const Offset(0, 10),
                            ),
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            // ── HEADER (from ReceiptTemplate OrgHeader) ──
                            Text(
                              _template.header.name.toUpperCase(),
                              textAlign: TextAlign.center,
                              style: GoogleFonts.spaceMono(
                                fontSize: 17,
                                fontWeight: FontWeight.w900,
                                color: const Color(0xFF0F172A),
                                letterSpacing: 1.2,
                              ),
                            ),
                            if (_template.header.address.isNotEmpty) ...[
                              const SizedBox(height: 3),
                              Text(
                                _template.header.address,
                                textAlign: TextAlign.center,
                                style: GoogleFonts.spaceMono(
                                  fontSize: 10.5,
                                  color: const Color(0xFF475569),
                                ),
                              ),
                            ],
                            if (_template.header.phone.isNotEmpty) ...[
                              const SizedBox(height: 2),
                              Text(
                                'Tel: ${_template.header.phone}',
                                textAlign: TextAlign.center,
                                style: GoogleFonts.spaceMono(
                                  fontSize: 10.5,
                                  color: const Color(0xFF475569),
                                ),
                              ),
                            ],
                            if (_template.header.taxId.isNotEmpty) ...[
                              const SizedBox(height: 2),
                              Text(
                                'TIN: ${_template.header.taxId}',
                                textAlign: TextAlign.center,
                                style: GoogleFonts.spaceMono(
                                  fontSize: 10.5,
                                  color: const Color(0xFF475569),
                                ),
                              ),
                            ],
                            const SizedBox(height: 3),
                            Text(
                              'Point of Sale System',
                              textAlign: TextAlign.center,
                              style: GoogleFonts.spaceMono(
                                fontSize: 11,
                                color: const Color(0xFF475569),
                              ),
                            ),
                            const SizedBox(height: 8),
                            // DASHED LINE
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 8),
                            // ── RECEIPT METADATA (from PosReceiptData) ──
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Receipt #: ${_receiptData.receiptNumber}',
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w800,
                                    color: const Color(0xFF1E293B),
                                  ),
                                ),
                                Text(
                                  _receiptData.formattedDateTime,
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 10.5,
                                    color: const Color(0xFF64748B),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 3),
                            Text(
                              'Cashier: ${_receiptData.cashier} • Terminal: ${_receiptData.terminal}',
                              textAlign: TextAlign.left,
                              style: GoogleFonts.spaceMono(
                                fontSize: 10.5,
                                color: const Color(0xFF475569),
                              ),
                            ),
                            const SizedBox(height: 3),
                            Text(
                              'Customer: ${_receiptData.customer}',
                              textAlign: TextAlign.left,
                              style: GoogleFonts.spaceMono(
                                fontSize: 11,
                                fontWeight: FontWeight.w700,
                                color: const Color(0xFF1E293B),
                              ),
                            ),
                            const SizedBox(height: 10),

                            // DASHED LINE
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 8),

                            // ── COLUMN HEADERS ──
                            Row(
                              children: [
                                Expanded(
                                  flex: 5,
                                  child: Text(
                                    'Item',
                                    style: GoogleFonts.spaceMono(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                      color: const Color(0xFF0F172A),
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 2,
                                  child: Text(
                                    'Qty',
                                    textAlign: TextAlign.center,
                                    style: GoogleFonts.spaceMono(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                      color: const Color(0xFF0F172A),
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 2,
                                  child: Text(
                                    'Price',
                                    textAlign: TextAlign.right,
                                    style: GoogleFonts.spaceMono(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                      color: const Color(0xFF0F172A),
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 2,
                                  child: Text(
                                    'Total',
                                    textAlign: TextAlign.right,
                                    style: GoogleFonts.spaceMono(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                      color: const Color(0xFF0F172A),
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 6),
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 8),

                            // ── LINE ITEMS ──
                            ListView.separated(
                              shrinkWrap: true,
                              physics: const NeverScrollableScrollPhysics(),
                              itemCount: _receiptData.items.length,
                              separatorBuilder: (_, __) => const SizedBox(height: 8),
                              itemBuilder: (context, idx) {
                                final item = _receiptData.items[idx];
                                return Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Expanded(
                                      flex: 5,
                                      child: Text(
                                        '${idx + 1}. ${item.name}',
                                        style: GoogleFonts.spaceMono(
                                          fontSize: 11.5,
                                          fontWeight: FontWeight.w800,
                                          color: const Color(0xFF0F172A),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: Text(
                                        item.qty.toStringAsFixed(0),
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.spaceMono(
                                          fontSize: 11,
                                          color: const Color(0xFF475569),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: Text(
                                        item.price.toStringAsFixed(2),
                                        textAlign: TextAlign.right,
                                        style: GoogleFonts.spaceMono(
                                          fontSize: 11,
                                          color: const Color(0xFF475569),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: Text(
                                        item.lineTotal.toStringAsFixed(2),
                                        textAlign: TextAlign.right,
                                        style: GoogleFonts.spaceMono(
                                          fontSize: 11.5,
                                          fontWeight: FontWeight.w800,
                                          color: const Color(0xFF0F172A),
                                        ),
                                      ),
                                    ),
                                  ],
                                );
                              },
                            ),
                            const SizedBox(height: 10),

                            // DASHED LINE
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 8),

                            // TOTAL ITEMS COUNT
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Total Items Count:',
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 11.5,
                                    color: const Color(0xFF475569),
                                  ),
                                ),
                                Text(
                                  '${_receiptData.items.length} items',
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 11.5,
                                    fontWeight: FontWeight.w800,
                                    color: const Color(0xFF0F172A),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),

                            // DASHED LINE
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 10),

                            // ── TOTALS (Calculated using pos-client business logic) ──
                            _buildReceiptSummaryRow('Subtotal', '${_receiptData.currency} ${_receiptData.subtotal.toStringAsFixed(2)}'),
                            if (_receiptData.discount > 0)
                              _buildReceiptSummaryRow('Discount (${_receiptData.discountPercentage.toStringAsFixed(0)}%)', '- ${_receiptData.currency} ${_receiptData.discount.toStringAsFixed(2)}'),
                            _buildReceiptSummaryRow('VAT (${(_receiptData.taxRate * 100).toStringAsFixed(0)}%)', '${_receiptData.currency} ${_receiptData.taxAmount.toStringAsFixed(2)}'),
                            const SizedBox(height: 6),

                            // SOLID LINE
                            Container(height: 1.5, color: goldAccent),
                            const SizedBox(height: 8),

                            // GRAND TOTAL
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'TOTAL',
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 15.5,
                                    fontWeight: FontWeight.w900,
                                    color: const Color(0xFF0F172A),
                                  ),
                                ),
                                Text(
                                  '${_receiptData.currency} ${_receiptData.grandTotal.toStringAsFixed(2)}',
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 15.5,
                                    fontWeight: FontWeight.w900,
                                    color: const Color(0xFF0F172A),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),

                            // DASHED LINE
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 8),

                            // Cash Paid & Change
                            _buildReceiptSummaryRow('Cash Paid:', '${_receiptData.currency} ${_receiptData.paid.toStringAsFixed(2)}'),
                            if (_receiptData.changeDue > 0)
                              _buildReceiptSummaryRow('Change Given:', '${_receiptData.currency} ${_receiptData.changeDue.toStringAsFixed(2)}'),
                            const SizedBox(height: 4),

                            // DASHED LINE
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 8),

                            // PAYMENT METHOD
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Payment Method:',
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 11.5,
                                    color: const Color(0xFF475569),
                                  ),
                                ),
                                Text(
                                  _receiptData.paymentMethod,
                                  style: GoogleFonts.spaceMono(
                                    fontSize: 11.5,
                                    fontWeight: FontWeight.w800,
                                    color: const Color(0xFF0F172A),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),

                            // DASHED LINE
                            _buildDashedLine(goldAccent),
                            const SizedBox(height: 12),

                            // ── FOOTER (from ReceiptTemplate OrgFooter) ──
                            if (_template.footer.thankYou.isNotEmpty) ...[
                              Text(
                                '** ${_template.footer.thankYou} **',
                                textAlign: TextAlign.center,
                                style: GoogleFonts.spaceMono(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w700,
                                  color: const Color(0xFF1E293B),
                                ),
                              ),
                              const SizedBox(height: 4),
                            ],
                            if (_template.footer.returnNote.isNotEmpty) ...[
                              Text(
                                _template.footer.returnNote,
                                textAlign: TextAlign.center,
                                style: GoogleFonts.spaceMono(
                                  fontSize: 10,
                                  color: const Color(0xFF475569),
                                ),
                              ),
                              const SizedBox(height: 4),
                            ],
                            if (_template.footer.website.isNotEmpty) ...[
                              Text(
                                _template.footer.website,
                                textAlign: TextAlign.center,
                                style: GoogleFonts.spaceMono(
                                  fontSize: 10,
                                  color: const Color(0xFF475569),
                                ),
                              ),
                              const SizedBox(height: 6),
                            ],

                            // MONOSPACE BARCODE
                            Text(
                              '||| | |||| || ||| |||| ||||',
                              textAlign: TextAlign.center,
                              style: GoogleFonts.spaceMono(
                                fontSize: 16,
                                fontWeight: FontWeight.w900,
                                color: const Color(0xFF1E293B),
                                letterSpacing: 3,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              _receiptData.barcode,
                              textAlign: TextAlign.center,
                              style: GoogleFonts.spaceMono(
                                fontSize: 9.5,
                                color: const Color(0xFF64748B),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),

          // Bottom Action Buttons
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
            decoration: BoxDecoration(
              color: cardBg,
              border: Border(top: BorderSide(color: Colors.white.withValues(alpha: 0.08))),
            ),
            child: Row(
              children: [
                // Print Receipt Button (White rounded container)
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _isPrinting ? null : _handlePrintReceipt,
                    icon: _isPrinting
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2, color: Color(0xFF0F172A)),
                          )
                        : const Icon(Icons.print_rounded, size: 18, color: Color(0xFF0F172A)),
                    label: Text(
                      _isPrinting ? 'Printing...' : 'Print Receipt',
                      style: GoogleFonts.inter(
                        fontSize: 13.5,
                        fontWeight: FontWeight.w800,
                        color: const Color(0xFF0F172A),
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      backgroundColor: Colors.white,
                      foregroundColor: const Color(0xFF0F172A),
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      side: const BorderSide(color: Colors.white),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(14),
                      ),
                      elevation: 2,
                    ),
                  ),
                ),
                const SizedBox(width: 14),

                // Proceed without Print Button (Dark Teal Gradient)
                Expanded(
                  child: Container(
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFF0F394C), Color(0xFF0E6068), Color(0xFF007A78)],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      borderRadius: BorderRadius.circular(14),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF007A78).withValues(alpha: 0.4),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: ElevatedButton.icon(
                      onPressed: widget.onProceedWithoutPrint,
                      icon: const Icon(Icons.check_circle_rounded, size: 18, color: Colors.white),
                      label: Text(
                        'Proceed without Print',
                        style: GoogleFonts.inter(
                          fontSize: 13,
                          fontWeight: FontWeight.w800,
                          color: Colors.white,
                        ),
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.transparent,
                        shadowColor: Colors.transparent,
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(14),
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDashedLine(Color color) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final boxWidth = constraints.constrainWidth();
        const dashWidth = 5.0;
        const dashHeight = 1.2;
        const dashSpace = 4.0;
        final dashCount = (boxWidth / (dashWidth + dashSpace)).floor();
        return Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: List.generate(dashCount, (_) {
            return SizedBox(
              width: dashWidth,
              height: dashHeight,
              child: DecoratedBox(
                decoration: BoxDecoration(color: color),
              ),
            );
          }),
        );
      },
    );
  }

  Widget _buildReceiptSummaryRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2.5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: GoogleFonts.spaceMono(
              fontSize: 11.5,
              color: const Color(0xFF475569),
            ),
          ),
          Text(
            value,
            style: GoogleFonts.spaceMono(
              fontSize: 11.5,
              fontWeight: FontWeight.w700,
              color: const Color(0xFF0F172A),
            ),
          ),
        ],
      ),
    );
  }
}
