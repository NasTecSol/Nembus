import 'dart:convert';
import 'dart:developer' as developer;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../database/db_service.dart';
import '../ffi/nembus_bridge.dart';
import '../singleton/singleton_class.dart';
import '../widgets/thermal_receipt_widget.dart';

class PosSaleScreen extends StatefulWidget {
  final String cartId;
  final String cartNumber;
  final Map<String, dynamic> customer;
  final List<Map<String, dynamic>> cartItems;
  final double subtotal;
  final double taxAmount;
  final double totalAmount;
  final int storeId;
  final String storeName;
  final int? posTerminalId;
  final String? posTerminalName;
  final int userId;
  final String username;
  final VoidCallback onOrderCompleted;

  const PosSaleScreen({
    super.key,
    required this.cartId,
    required this.cartNumber,
    required this.customer,
    required this.cartItems,
    required this.subtotal,
    required this.taxAmount,
    required this.totalAmount,
    required this.storeId,
    required this.storeName,
    this.posTerminalId,
    this.posTerminalName,
    required this.userId,
    required this.username,
    required this.onOrderCompleted,
  });

  @override
  State<PosSaleScreen> createState() => _PosSaleScreenState();
}

class _PosSaleScreenState extends State<PosSaleScreen> {
  String _selectedPaymentMethod = 'cash';
  late TextEditingController _amountPaidController;
  bool _isProcessing = false;
  int _currentWorkflowStep = 0;
  String _stepStatusText = '';
  String? _generatedOrderNumber;
  String? _errorMessage;

  final List<Map<String, dynamic>> _paymentMethods = [
    {
      'id': 'cash',
      'name': 'Cash Payment',
      'icon': Icons.payments_rounded,
      'gradient': [Color(0xFF059669), Color(0xFF10B981)],
      'accent': Color(0xFF10B981),
    },
    {
      'id': 'card',
      'name': 'Card / POS Terminal',
      'icon': Icons.credit_card_rounded,
      'gradient': [Color(0xFF0284C7), Color(0xFF38BDF8)],
      'accent': Color(0xFF38BDF8),
    },
    {
      'id': 'transfer',
      'name': 'Bank Transfer',
      'icon': Icons.account_balance_rounded,
      'gradient': [Color(0xFF7C3AED), Color(0xFFA855F7)],
      'accent': Color(0xFFA855F7),
    },
    {
      'id': 'credit',
      'name': 'Store Credit / Account',
      'icon': Icons.wallet_rounded,
      'gradient': [Color(0xFFD97706), Color(0xFFF59E0B)],
      'accent': Color(0xFFF59E0B),
    },
  ];

  @override
  void initState() {
    super.initState();
    _amountPaidController = TextEditingController(
      text: widget.totalAmount.toStringAsFixed(2),
    );
  }

  @override
  void dispose() {
    _amountPaidController.dispose();
    super.dispose();
  }

  double get _normalizedTotalDue => double.parse(widget.totalAmount.toStringAsFixed(2));

  double get _amountPaid {
    final parsed = double.tryParse(_amountPaidController.text.trim());
    final val = parsed ?? widget.totalAmount;
    return double.parse(val.toStringAsFixed(2));
  }

  double get _changeDue {
    final diff = _amountPaid - _normalizedTotalDue;
    return diff > 0.005 ? double.parse(diff.toStringAsFixed(2)) : 0.0;
  }

  void _setExactAmount() {
    setState(() {
      _amountPaidController.text = _normalizedTotalDue.toStringAsFixed(2);
    });
  }

  void _addQuickCash(double additional) {
    setState(() {
      final current = double.tryParse(_amountPaidController.text) ?? _normalizedTotalDue;
      _amountPaidController.text = (current + additional).toStringAsFixed(2);
    });
  }

  String _generateUuidV4() {
    final now = DateTime.now().millisecondsSinceEpoch;
    final hexTime = now.toRadixString(16).padLeft(12, '0');
    return '00000000-0000-4000-8000-${hexTime.substring(hexTime.length - 12)}';
  }

  /// Converts Cart to Order and executes the 3 mandatory offline order steps:
  /// Step 1: Updating payment status for order
  /// Step 2: Updating fulfillment status for order
  /// Step 3: Updating order status for order
  Future<void> _executeOrderWorkflow() async {
    final totalDue = _normalizedTotalDue;
    if ((_amountPaid < totalDue - 0.005) && _selectedPaymentMethod != 'credit') {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          backgroundColor: const Color(0xFFEF4444),
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          content: const Text('Amount paid cannot be less than total amount due'),
        ),
      );
      return;
    }

    setState(() {
      _isProcessing = true;
      _currentWorkflowStep = 1;
      _stepStatusText = 'Converting Cart to Sales Order...';
      _errorMessage = null;
    });

    try {
      final orderId = _generateUuidV4();
      final orderNumber = 'SO-${DateTime.now().millisecondsSinceEpoch.toString().substring(4)}';
      _generatedOrderNumber = orderNumber;

      final customerId = widget.customer['id'];
      final customerName = widget.customer['name'] ?? 'Walk-in Customer';
      final customerPhone = widget.customer['phone'] ?? '';
      final customerEmail = widget.customer['email'] ?? '';

      int posTerminalId = widget.posTerminalId ?? SingletonClass().activeTerminalId ?? 1;
      int cashierId = 1;
      int cashierSessionId = 1;

      developer.log(
        'Converting cart=${widget.cartId} to order=$orderId ($orderNumber) for customer=$customerName',
        name: 'PosSaleScreen',
      );
      final convertResp = NembusBridge().callHandler(
        handler: 'cart',
        action: 'convertToOrder',
        payload: {
          'cart_id': widget.cartId,
          'order_id': orderId,
          'order_number': orderNumber,
          'customer_id': customerId,
          'payment_method': _selectedPaymentMethod,
          'total_amount': widget.totalAmount,
        },
      );
      developer.log('Cart convertToOrder response: $convertResp', name: 'PosSaleScreen');
      try {
        final db = DatabaseService().database;
        int orgId = 1;
        try {
          final orgRows = await db.query('organizations', limit: 1);
          if (orgRows.isNotEmpty) orgId = (orgRows.first['id'] as num).toInt();
        } catch (_) {}
        final existingCart = await db.query('carts', where: 'id = ?', whereArgs: [widget.cartId]);
        if (existingCart.isEmpty) {
          await db.insert('carts', {
            'id': widget.cartId,
            'cart_number': widget.cartNumber,
            'organization_id': orgId,
            'store_id': widget.storeId,
            'customer_id': customerId,
            'cart_status': 'active',
            'subtotal': widget.subtotal,
            'total_amount': widget.totalAmount,
            'created_at': DateTime.now().toIso8601String(),
            'updated_at': DateTime.now().toIso8601String(),
          });
        }

        // Dynamically resolve terminal, cashier, and active open cashier session
        try {
          if (widget.posTerminalId != null) {
            posTerminalId = widget.posTerminalId!;
          } else if (SingletonClass().activeTerminalId != null) {
            posTerminalId = SingletonClass().activeTerminalId!;
          } else {
            final devRows = await db.query('local_device_config', limit: 1);
            if (devRows.isNotEmpty && devRows.first['pos_terminal_id'] != null) {
              posTerminalId = (devRows.first['pos_terminal_id'] as num).toInt();
            } else {
              final termRows = await db.query('pos_terminals', where: 'store_id = ?', whereArgs: [widget.storeId], limit: 1);
              if (termRows.isNotEmpty) {
                posTerminalId = (termRows.first['id'] as num).toInt();
              }
            }
          }

          // Dynamically find cashier linked to current user
          final cashierByUser = await db.query('cashiers', where: 'user_id = ?', whereArgs: [widget.userId], limit: 1);
          if (cashierByUser.isNotEmpty) {
            cashierId = (cashierByUser.first['id'] as num).toInt();
          } else {
            final cashierRows = await db.query('cashiers', limit: 1);
            if (cashierRows.isNotEmpty) {
              cashierId = (cashierRows.first['id'] as num).toInt();
            } else {
              cashierId = await db.insert('cashiers', {
                'cashier_code': 'CASH-${widget.userId}',
                'user_id': widget.userId,
                'status': 'active',
              });
            }
          }

          // Use active open session from SingletonClass if user has started their shift
          if (SingletonClass().activeCashierSessionId != null && SingletonClass().isSessionActive) {
            cashierSessionId = SingletonClass().activeCashierSessionId!;
          } else {
            cashierSessionId = 0;
          }
        } catch (csErr) {
          developer.log('Terminal/cashier/session resolution note: $csErr', name: 'PosSaleScreen');
        }
        int? validUserId;
        try {
          final userCheck = await db.query('users', where: 'id = ?', whereArgs: [widget.userId]);
          if (userCheck.isNotEmpty) {
            validUserId = widget.userId;
          } else {
            final anyUser = await db.query('users', limit: 1);
            if (anyUser.isNotEmpty) {
              validUserId = (anyUser.first['id'] as num).toInt();
            }
          }
        } catch (_) {}
        final nowIso = DateTime.now().toIso8601String();
        await db.insert('sales_orders_v2', {
          'id': orderId,
          'order_number': orderNumber,
          'organization_id': orgId,
          'store_id': widget.storeId,
          'customer_id': customerId,
          'customer_name': customerName,
          'customer_phone': customerPhone,
          'customer_email': customerEmail,
          'order_type': 'standard',
          'order_status': 'draft',
          'payment_status': 'unpaid',
          'fulfillment_status': 'unfulfilled',
          'sales_channel': 'pos',
          'source_cart_id': widget.cartId,
          'created_by_user_id': validUserId,
          'cashier_id': cashierId,
          'pos_terminal_id': posTerminalId,
          'subtotal': widget.subtotal,
          'tax_amount': widget.taxAmount,
          'total_amount': widget.totalAmount,
          'paid_amount': _amountPaid,
          'shipping_address': 'Store POS Checkout - ${widget.storeName}',
          'billing_address': 'Store POS Checkout - ${widget.storeName}',
          'payment_method': _selectedPaymentMethod,
          'order_date': nowIso,
          'created_at': nowIso,
          'updated_at': nowIso,
        });

        int lineIdx = 1;
        for (final item in widget.cartItems) {
          final pid = (item['product_id'] as num?)?.toInt() ?? 1;
          final pName = item['product_name']?.toString() ?? 'Product #$pid';
          final pSku = item['sku']?.toString() ?? 'SKU-$pid';
          final qty = (item['quantity'] as num?)?.toDouble() ?? 1.0;
          final unitPrice = (item['unit_price'] as num?)?.toDouble() ?? 0.0;
          final lineTotal = (item['line_total'] as num?)?.toDouble() ?? (qty * unitPrice);

          await db.insert('sales_order_lines_v2', {
            'id': 'LINE-${DateTime.now().millisecondsSinceEpoch}-$lineIdx',
            'sales_order_id': orderId,
            'organization_id': orgId,
            'line_number': lineIdx,
            'product_id': pid,
            'product_name': pName,
            'product_sku': pSku,
            'quantity_ordered': qty,
            'quantity_fulfilled': 0.0,
            'unit_price': unitPrice,
            'line_total': lineTotal,
            'line_status': 'pending',
            'created_at': nowIso,
            'updated_at': nowIso,
          });
          lineIdx++;
        }
      } catch (dbErr) {
        developer.log('Order local SQLite insert note: $dbErr', name: 'PosSaleScreen');
      }

      await Future.delayed(const Duration(milliseconds: 650));

      // =========================================================================
      // STEP 1: Updating payment status for order
      // =========================================================================
      setState(() {
        _currentWorkflowStep = 1;
        _stepStatusText = 'Step 1: Updating payment status for order...';
      });

      final payResp = NembusBridge().callHandler(
        handler: 'order',
        action: 'updateOrderPaymentStatus',
        payload: {
          'id': orderId,
          'order_id': orderId,
          'payment_status': 'paid',
          'paid_amount': _amountPaid.toStringAsFixed(2),
          'payment_method': _selectedPaymentMethod,
          'payment_gateway': 'pos',
        },
      );
      developer.log('Step 1 Payment status update response: $payResp', name: 'PosSaleScreen');

      try {
        final db = DatabaseService().database;
        final nowIso = DateTime.now().toIso8601String();

        await db.update(
          'sales_orders_v2',
          {
            'payment_status': 'paid',
            'paid_amount': _amountPaid,
            'payment_gateway': 'pos',
            'updated_at': nowIso,
          },
          where: 'id = ?',
          whereArgs: [orderId],
        );

        final txnNumber = 'TXN-$orderNumber';

        final txnId = await db.insert('pos_transactions', {
          'store_id': widget.storeId,
          'cashier_id': cashierId,
          'cashier_session_id': cashierSessionId > 0 ? cashierSessionId : null,
          'customer_id': customerId,
          'pos_terminal_id': posTerminalId,
          'transaction_number': txnNumber,
          'transaction_date': nowIso,
          'transaction_type': 'sale',
          'subtotal': widget.subtotal,
          'tax_amount': widget.taxAmount,
          'total_amount': widget.totalAmount,
          'amount_paid': _amountPaid,
          'change_given': _changeDue,
          'status': 'completed',
          'sales_order_id': orderId,
          'source_cart_id': widget.cartId,
          'created_at': nowIso,
        });

        int tLineIdx = 1;
        for (final item in widget.cartItems) {
          final pid = (item['product_id'] as num?)?.toInt() ?? 1;
          final qty = (item['quantity'] as num?)?.toDouble() ?? 1.0;
          final unitPrice = (item['unit_price'] as num?)?.toDouble() ?? 0.0;
          final lineTotal = (item['line_total'] as num?)?.toDouble() ?? (qty * unitPrice);

          await db.insert('pos_transaction_lines', {
            'transaction_id': txnId,
            'product_id': pid,
            'quantity': qty,
            'unit_price': unitPrice,
            'subtotal': lineTotal,
            'line_total': lineTotal,
            'line_number': tLineIdx,
            'created_at': nowIso,
          });
          tLineIdx++;
        }

        await db.insert('pos_payments', {
          'transaction_id': txnId,
          'payment_method': _selectedPaymentMethod,
          'payment_gateway': 'pos',
          'amount': _amountPaid,
          'payment_reference': 'POS-REF-${DateTime.now().millisecondsSinceEpoch}',
          'reference_number': orderNumber,
          'payment_date': nowIso,
          'created_at': nowIso,
        });

        // Update expected balance in cashier_sessions if payment method is cash
        if (_selectedPaymentMethod.toLowerCase() == 'cash' && cashierSessionId > 0) {
          final netCash = _amountPaid - _changeDue;
          if (netCash > 0) {
            await db.rawUpdate('''
              UPDATE cashier_sessions
              SET expected_balance = COALESCE(expected_balance, opening_balance, 0) + ?,
                  updated_at = ?
              WHERE id = ?
            ''', [netCash, nowIso, cashierSessionId]);
          }
        }
      } catch (dbErr) {
        developer.log('Step 1 SQLite update: $dbErr', name: 'PosSaleScreen');
      }

      await Future.delayed(const Duration(milliseconds: 750));

      // =========================================================================
      // STEP 2: Updating fulfillment status for order
      // =========================================================================
      setState(() {
        _currentWorkflowStep = 2;
        _stepStatusText = 'Step 2: Updating fulfillment status for order...';
      });

      final fulfillResp = NembusBridge().callHandler(
        handler: 'order',
        action: 'updateOrderFulfillmentStatus',
        payload: {
          'id': orderId,
          'order_id': orderId,
          'fulfillment_status': 'fulfilled',
        },
      );
      developer.log('Step 2 Fulfillment status update response: $fulfillResp', name: 'PosSaleScreen');

      try {
        final db = DatabaseService().database;
        await db.update(
          'sales_orders_v2',
          {
            'fulfillment_status': 'fulfilled',
            'updated_at': DateTime.now().toIso8601String(),
          },
          where: 'id = ?',
          whereArgs: [orderId],
        );
        await db.update(
          'sales_order_lines_v2',
          {
            'line_status': 'fulfilled',
            'updated_at': DateTime.now().toIso8601String(),
          },
          where: 'sales_order_id = ?',
          whereArgs: [orderId],
        );
      } catch (dbErr) {
        developer.log('Step 2 SQLite update: $dbErr', name: 'PosSaleScreen');
      }

      await Future.delayed(const Duration(milliseconds: 750));

      // =========================================================================
      // STEP 3: Updating order status for order
      // =========================================================================
      setState(() {
        _currentWorkflowStep = 3;
        _stepStatusText = 'Step 3: Updating order status for order...';
      });

      final statusResp = NembusBridge().callHandler(
        handler: 'order',
        action: 'updateOrderStatus',
        payload: {
          'id': orderId,
          'order_id': orderId,
          'order_status': 'fulfilled',
        },
      );
      developer.log('Step 3 Order status update response: $statusResp', name: 'PosSaleScreen');

      try {
        final db = DatabaseService().database;
        await db.update(
          'sales_orders_v2',
          {
            'order_status': 'fulfilled',
            'confirmed_date': DateTime.now().toIso8601String(),
            'updated_at': DateTime.now().toIso8601String(),
          },
          where: 'id = ?',
          whereArgs: [orderId],
        );

        await db.update(
          'carts',
          {
            'cart_status': 'converted',
            'converted_to_order_id': orderId,
            'converted_at': DateTime.now().toIso8601String(),
            'updated_at': DateTime.now().toIso8601String(),
          },
          where: 'id = ?',
          whereArgs: [widget.cartId],
        );

        // Enqueue into Transactional Outbox (sync_queue) with strict dependency priority
        try {
          // 1. Enqueue Cashier (Priority 30)
          final cRows = await db.query('cashiers', where: 'id = ?', whereArgs: [cashierId]);
          if (cRows.isNotEmpty) {
            await db.insert('sync_queue', {
              'entity_type': 'cashiers',
              'entity_id': cashierId.toString(),
              'action': 'INSERT',
              'payload': jsonEncode(cRows.first),
              'status': 'pending',
              'priority': 30,
              'correlation_id': orderId,
              'created_at': DateTime.now().toIso8601String(),
            });
          }

          // 2. Enqueue Cashier Session (Priority 25)
          if (cashierSessionId > 0) {
            final sRows = await db.query('cashier_sessions', where: 'id = ?', whereArgs: [cashierSessionId]);
            if (sRows.isNotEmpty) {
              await db.insert('sync_queue', {
                'entity_type': 'cashier_sessions',
                'entity_id': cashierSessionId.toString(),
                'action': 'UPDATE',
                'payload': jsonEncode(sRows.first),
                'status': 'pending',
                'priority': 25,
                'correlation_id': orderId,
                'created_at': DateTime.now().toIso8601String(),
              });
            }
          }

          // 3. Enqueue Customer if present (Priority 20)
          if (customerId != null) {
            final custRows = await db.query('customers', where: 'id = ?', whereArgs: [customerId]);
            if (custRows.isNotEmpty) {
              await db.insert('sync_queue', {
                'entity_type': 'customers',
                'entity_id': customerId.toString(),
                'action': 'INSERT',
                'payload': jsonEncode(custRows.first),
                'status': 'pending',
                'priority': 20,
                'correlation_id': orderId,
                'created_at': DateTime.now().toIso8601String(),
              });
            }
          }

          // 4. Enqueue Sales Order (Priority 15)
          final orderRows = await db.query('sales_orders_v2', where: 'id = ?', whereArgs: [orderId]);
          if (orderRows.isNotEmpty) {
            final orderPayload = Map<String, dynamic>.from(orderRows.first);
            orderPayload['source_cart_id'] = null; // Local cart session UUID; nullified for cloud FK schema compatibility
            await db.insert('sync_queue', {
              'entity_type': 'sales_orders_v2',
              'entity_id': orderId,
              'action': 'INSERT',
              'payload': jsonEncode(orderPayload),
              'status': 'pending',
              'priority': 15,
              'correlation_id': orderId,
              'created_at': DateTime.now().toIso8601String(),
            });
          }

          // 5. Enqueue Sales Order Lines (Priority 14)
          final orderLineRows = await db.query('sales_order_lines_v2', where: 'sales_order_id = ?', whereArgs: [orderId]);
          for (final line in orderLineRows) {
            await db.insert('sync_queue', {
              'entity_type': 'sales_order_lines_v2',
              'entity_id': line['id'].toString(),
              'action': 'INSERT',
              'payload': jsonEncode(line),
              'status': 'pending',
              'priority': 14,
              'correlation_id': orderId,
              'created_at': DateTime.now().toIso8601String(),
            });
          }

          // 6. Enqueue POS Transaction (Priority 10)
          final txnRows = await db.query('pos_transactions', where: 'sales_order_id = ?', whereArgs: [orderId]);
          if (txnRows.isNotEmpty) {
            final txnPayload = Map<String, dynamic>.from(txnRows.first);
            txnPayload['source_cart_id'] = null; // Local cart session UUID; nullified for cloud FK schema compatibility
            final txnIdVal = txnPayload['id'];
            await db.insert('sync_queue', {
              'entity_type': 'pos_transactions',
              'entity_id': txnIdVal.toString(),
              'action': 'INSERT',
              'payload': jsonEncode(txnPayload),
              'status': 'pending',
              'priority': 10,
              'correlation_id': orderId,
              'created_at': DateTime.now().toIso8601String(),
            });

            // 7. Enqueue POS Transaction Lines (Priority 8)
            final txnLineRows = await db.query('pos_transaction_lines', where: 'transaction_id = ?', whereArgs: [txnIdVal]);
            for (final tLine in txnLineRows) {
              await db.insert('sync_queue', {
                'entity_type': 'pos_transaction_lines',
                'entity_id': tLine['id'].toString(),
                'action': 'INSERT',
                'payload': jsonEncode(tLine),
                'status': 'pending',
                'priority': 8,
                'correlation_id': orderId,
                'created_at': DateTime.now().toIso8601String(),
              });
            }
          }

          // 8. Enqueue POS Payments (Priority 8)
          final paymentRows = await db.query('pos_payments', where: 'reference_number = ?', whereArgs: [orderNumber]);
          if (paymentRows.isNotEmpty) {
            await db.insert('sync_queue', {
              'entity_type': 'pos_payments',
              'entity_id': paymentRows.first['id'].toString(),
              'action': 'INSERT',
              'payload': jsonEncode(paymentRows.first),
              'status': 'pending',
              'priority': 8,
              'correlation_id': orderId,
              'created_at': DateTime.now().toIso8601String(),
            });
          }
        } catch (syncErr) {
          developer.log('Outbox enqueue note: $syncErr', name: 'PosSaleScreen');
        }
      } catch (dbErr) {
        developer.log('Step 3 SQLite update: $dbErr', name: 'PosSaleScreen');
      }

      await Future.delayed(const Duration(milliseconds: 800));

      setState(() {
        _currentWorkflowStep = 4;
        _isProcessing = false;
        _stepStatusText = 'Order Complete!';
      });
    } catch (e) {
      developer.log('Error executing order workflow: $e', name: 'PosSaleScreen');
      setState(() {
        _isProcessing = false;
        _errorMessage = 'Order processing error: $e';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    const bgColor = Color(0xFF070B14);
    const cardBgColor = Color(0xFF111827);
    const surfaceColor = Color(0xFF1F2937);
    const primarySky = Color(0xFF38BDF8);
    const emeraldGreen = Color(0xFF10B981);

    final customerName = widget.customer['name'] ?? 'Walk-in Customer';
    final customerCode = widget.customer['customer_code'] ?? 'CUST-001';

    return Scaffold(
      backgroundColor: bgColor,
      appBar: _currentWorkflowStep == 4
          ? null
          : AppBar(
              backgroundColor: cardBgColor,
              elevation: 0,
              leading: IconButton(
                icon: const Icon(Icons.arrow_back_ios_new_rounded, color: Colors.white70),
                onPressed: () => Navigator.of(context).pop(),
              ),
              title: Text(
                '3D Smart Checkout',
                style: GoogleFonts.inter(
                  fontWeight: FontWeight.w800,
                  fontSize: 17,
                  color: Colors.white,
                  letterSpacing: -0.3,
                ),
              ),
            ),
      body: SafeArea(
        child: _currentWorkflowStep == 4
            ? ThermalReceiptWidget(
                storeName: widget.storeName.isNotEmpty
                    ? widget.storeName
                    : (SingletonClass().activeStoreName ?? 'QITAF AL AYELA'),
                orderNumber: _generatedOrderNumber != null
                    ? 'TXN-$_generatedOrderNumber'
                    : 'TXN-ORD-${DateTime.now().millisecondsSinceEpoch}',
                cashierName: widget.username.isNotEmpty
                    ? widget.username
                    : (SingletonClass().activeCashierCode ?? 'Cashier'),
                terminalName: widget.posTerminalName ??
                    SingletonClass().activeTerminalName ??
                    'Main Counter Terminal 1',
                customerName: widget.customer['name'] ?? 'Walk-in Customer',
                paymentMethod: _selectedPaymentMethod,
                cartItems: widget.cartItems,
                subtotal: widget.subtotal,
                discount: 0.0,
                taxAmount: widget.taxAmount,
                totalAmount: widget.totalAmount,
                changeDue: _changeDue,
                onProceedWithoutPrint: () {
                  widget.onOrderCompleted();
                  Navigator.of(context).pop();
                },
                onBack: () {
                  widget.onOrderCompleted();
                  Navigator.of(context).pop();
                },
              )
            : _build3DCheckoutForm(cardBgColor, surfaceColor, primarySky, emeraldGreen, customerName, customerCode),
      ),
    );
  }

  Widget _build3DCheckoutForm(
    Color cardBgColor,
    Color surfaceColor,
    Color primarySky,
    Color emeraldGreen,
    String customerName,
    String customerCode,
  ) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isTablet = constraints.maxWidth >= 720;

        if (isTablet) {
          return SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 20.0),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 1080),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Left Column (Customer Overview & Payment Methods)
                    Expanded(
                      flex: 5,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _buildCustomerOverview(customerName, customerCode, primarySky, emeraldGreen),
                          const SizedBox(height: 20),
                          _buildPaymentSection(cardBgColor, surfaceColor, childAspectRatio: 2.8),
                        ],
                      ),
                    ),
                    const SizedBox(width: 24),
                    // Right Column (Amount Due, Cash presets, Change, Timeline, CTA)
                    Expanded(
                      flex: 5,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _buildAmountDueCard(cardBgColor, surfaceColor, primarySky, emeraldGreen),
                          const SizedBox(height: 20),
                          if (_isProcessing) ...[
                            _buildProcessingTimeline(emeraldGreen, primarySky),
                            const SizedBox(height: 20),
                          ],
                          if (_errorMessage != null) ...[
                            _buildErrorBox(),
                            const SizedBox(height: 16),
                          ],
                          _buildExecuteButton(emeraldGreen),
                          const SizedBox(height: 24),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        }

        // Phone / Mobile Layout
        return SingleChildScrollView(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildCustomerOverview(customerName, customerCode, primarySky, emeraldGreen),
              const SizedBox(height: 20),
              _buildPaymentSection(cardBgColor, surfaceColor, childAspectRatio: 2.1),
              const SizedBox(height: 20),
              _buildAmountDueCard(cardBgColor, surfaceColor, primarySky, emeraldGreen),
              const SizedBox(height: 24),
              if (_isProcessing) ...[
                _buildProcessingTimeline(emeraldGreen, primarySky),
                const SizedBox(height: 20),
              ],
              if (_errorMessage != null) ...[
                _buildErrorBox(),
                const SizedBox(height: 16),
              ],
              _buildExecuteButton(emeraldGreen),
              const SizedBox(height: 24),
            ],
          ),
        );
      },
    );
  }

  Widget _buildCustomerOverview(String customerName, String customerCode, Color primarySky, Color emeraldGreen) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF1E293B), Color(0xFF0F172A)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: primarySky.withValues(alpha: 0.2)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.4),
            blurRadius: 12,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              gradient: const LinearGradient(colors: [Color(0xFF6366F1), Color(0xFF38BDF8)]),
              borderRadius: BorderRadius.circular(14),
              boxShadow: [
                BoxShadow(
                  color: primarySky.withValues(alpha: 0.3),
                  blurRadius: 8,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: const Icon(Icons.person_rounded, color: Colors.white, size: 24),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  customerName,
                  style: GoogleFonts.inter(
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                    color: Colors.white,
                  ),
                ),
                Text(
                  'Code: $customerCode • ${widget.storeName}',
                  style: GoogleFonts.inter(fontSize: 12, color: Colors.white60),
                ),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
            decoration: BoxDecoration(
              color: emeraldGreen.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: emeraldGreen.withValues(alpha: 0.4)),
            ),
            child: Text(
              '${widget.cartItems.length} Items',
              style: GoogleFonts.inter(
                fontSize: 11.5,
                fontWeight: FontWeight.w800,
                color: emeraldGreen,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPaymentSection(Color cardBgColor, Color surfaceColor, {required double childAspectRatio}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'STORE PAYMENT METHOD',
          style: GoogleFonts.inter(
            fontSize: 11.5,
            fontWeight: FontWeight.w800,
            letterSpacing: 1.1,
            color: Colors.white54,
          ),
        ),
        const SizedBox(height: 10),
        GridView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 2,
            mainAxisSpacing: 12,
            crossAxisSpacing: 12,
            childAspectRatio: childAspectRatio,
          ),
          itemCount: _paymentMethods.length,
          itemBuilder: (context, idx) {
            final pm = _paymentMethods[idx];
            final isSelected = _selectedPaymentMethod == pm['id'];
            final accentColor = pm['accent'] as Color;
            final gradientList = pm['gradient'] as List<Color>;

            return InkWell(
              borderRadius: BorderRadius.circular(16),
              onTap: _isProcessing
                  ? null
                  : () => setState(() => _selectedPaymentMethod = pm['id']),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                decoration: BoxDecoration(
                  color: isSelected ? accentColor.withValues(alpha: 0.18) : cardBgColor,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: isSelected ? accentColor : Colors.white.withValues(alpha: 0.08),
                    width: isSelected ? 1.8 : 1,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: isSelected
                          ? accentColor.withValues(alpha: 0.3)
                          : Colors.black.withValues(alpha: 0.25),
                      blurRadius: isSelected ? 12 : 6,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        gradient: isSelected ? LinearGradient(colors: gradientList) : null,
                        color: isSelected ? null : surfaceColor,
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Icon(
                        pm['icon'] as IconData,
                        color: isSelected ? Colors.white : Colors.white60,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        pm['name'] as String,
                        style: GoogleFonts.inter(
                          fontSize: 12.5,
                          fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
                          color: isSelected ? Colors.white : Colors.white70,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        ),
      ],
    );
  }

  Widget _buildAmountDueCard(Color cardBgColor, Color surfaceColor, Color primarySky, Color emeraldGreen) {
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: cardBgColor,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.35), blurRadius: 10, offset: const Offset(0, 4)),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Total Order Due',
                style: GoogleFonts.inter(fontSize: 14, color: Colors.white60, fontWeight: FontWeight.w600),
              ),
              Text(
                'SAR ${widget.totalAmount.toStringAsFixed(2)}',
                style: GoogleFonts.inter(
                  fontSize: 22,
                  fontWeight: FontWeight.w900,
                  color: emeraldGreen,
                  letterSpacing: -0.5,
                ),
              ),
            ],
          ),
          const Divider(color: Colors.white10, height: 24),
          Text(
            'Amount Paid (SAR)',
            style: GoogleFonts.inter(fontSize: 12, fontWeight: FontWeight.w700, color: Colors.white70),
          ),
          const SizedBox(height: 8),
          Container(
            decoration: BoxDecoration(
              color: surfaceColor,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: primarySky.withValues(alpha: 0.3)),
            ),
            child: TextField(
              controller: _amountPaidController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              enabled: !_isProcessing,
              onChanged: (_) => setState(() {}),
              style: GoogleFonts.inter(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
              decoration: InputDecoration(
                prefixIcon:  Icon(Icons.attach_money_rounded, color: primarySky),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
                suffixIcon: IconButton(
                  icon:  Icon(Icons.check_circle_outline_rounded, color: primarySky),
                  tooltip: 'Set Exact',
                  onPressed: _setExactAmount,
                ),
              ),
            ),
          ),
          const SizedBox(height: 12),

          // Quick Cash Presets
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                _buildQuickCashPill('Exact', _setExactAmount, primarySky),
                const SizedBox(width: 8),
                _buildQuickCashPill('+10', () => _addQuickCash(10), Colors.white60),
                const SizedBox(width: 8),
                _buildQuickCashPill('+50', () => _addQuickCash(50), Colors.white60),
                const SizedBox(width: 8),
                _buildQuickCashPill('+100', () => _addQuickCash(100), Colors.white60),
                const SizedBox(width: 8),
                _buildQuickCashPill('+500', () => _addQuickCash(500), Colors.white60),
              ],
            ),
          ),

          if (_changeDue > 0) ...[
            const SizedBox(height: 14),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: primarySky.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: primarySky.withValues(alpha: 0.4)),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Change to Return:',
                    style: GoogleFonts.inter(fontSize: 13, color: Colors.white70, fontWeight: FontWeight.w600),
                  ),
                  Text(
                    'SAR ${_changeDue.toStringAsFixed(2)}',
                    style: GoogleFonts.inter(
                      fontSize: 17,
                      fontWeight: FontWeight.w900,
                      color: primarySky,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildProcessingTimeline(Color emeraldGreen, Color primarySky) {
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF1E1B4B), Color(0xFF0F172A)],
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: primarySky.withValues(alpha: 0.5), width: 1.5),
        boxShadow: [
          BoxShadow(
            color: primarySky.withValues(alpha: 0.3),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            children: [
              SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2.8,
                  valueColor: AlwaysStoppedAnimation<Color>(primarySky),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  _stepStatusText,
                  style: GoogleFonts.inter(
                    color: Colors.white,
                    fontSize: 13.5,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 18),
          Row(
            children: [
              _build3DStepBadge(1, 'Step 1\nPayment', _currentWorkflowStep >= 1, emeraldGreen, primarySky),
              const Expanded(child: Divider(color: Colors.white24, indent: 6, endIndent: 6, thickness: 1.2)),
              _build3DStepBadge(2, 'Step 2\nFulfillment', _currentWorkflowStep >= 2, emeraldGreen, primarySky),
              const Expanded(child: Divider(color: Colors.white24, indent: 6, endIndent: 6, thickness: 1.2)),
              _build3DStepBadge(3, 'Step 3\nComplete', _currentWorkflowStep >= 3, emeraldGreen, primarySky),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildErrorBox() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFEF4444).withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFEF4444)),
      ),
      child: Text(
        _errorMessage!,
        style: GoogleFonts.inter(color: const Color(0xFFEF4444), fontSize: 13),
      ),
    );
  }

  Widget _buildExecuteButton(Color emeraldGreen) {
    return SizedBox(
      width: double.infinity,
      height: 54,
      child: ElevatedButton(
        onPressed: _isProcessing ? null : _executeOrderWorkflow,
        style: ElevatedButton.styleFrom(
          backgroundColor: emeraldGreen,
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          elevation: 8,
          shadowColor: emeraldGreen.withValues(alpha: 0.6),
        ),
        child: _isProcessing
            ? const SizedBox(
                width: 24,
                height: 24,
                child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2.5),
              )
            : Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.check_circle_outline_rounded, size: 20),
                  const SizedBox(width: 8),
                  Text(
                    'Execute 3-Step Sale (SAR ${widget.totalAmount.toStringAsFixed(2)})',
                    style: GoogleFonts.inter(fontSize: 15, fontWeight: FontWeight.w800),
                  ),
                ],
              ),
      ),
    );
  }

  Widget _buildQuickCashPill(String label, VoidCallback onTap, Color color) {
    return InkWell(
      borderRadius: BorderRadius.circular(10),
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: const Color(0xFF1E293B),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: color.withValues(alpha: 0.4)),
        ),
        child: Text(
          label,
          style: GoogleFonts.inter(
            color: color,
            fontWeight: FontWeight.w700,
            fontSize: 12,
          ),
        ),
      ),
    );
  }

  Widget _build3DStepBadge(int stepNum, String title, bool isDoneOrActive, Color doneColor, Color activeColor) {
    final isDone = _currentWorkflowStep > stepNum;
    final isActive = _currentWorkflowStep == stepNum;
    final color = isDone ? doneColor : (isActive ? activeColor : Colors.white30);

    return Column(
      children: [
        Container(
          width: 32,
          height: 32,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: color.withValues(alpha: 0.2),
            border: Border.all(color: color, width: 2),
            boxShadow: [
              if (isActive || isDone)
                BoxShadow(
                  color: color.withValues(alpha: 0.4),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
            ],
          ),
          child: isDone
              ? Icon(Icons.check, size: 18, color: doneColor)
              : Text(
                  '$stepNum',
                  style: GoogleFonts.inter(
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                    color: color,
                  ),
                ),
        ),
        const SizedBox(height: 6),
        Text(
          title,
          textAlign: TextAlign.center,
          style: GoogleFonts.inter(
            fontSize: 10,
            color: color,
            fontWeight: isDoneOrActive ? FontWeight.w800 : FontWeight.w500,
            height: 1.1,
          ),
        ),
      ],
    );
  }

  Widget _build3DSuccessReceipt(Color cardBgColor, Color surfaceColor, Color primarySky, Color emeraldGreen) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20.0),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 560),
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(22),
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [emeraldGreen.withValues(alpha: 0.3), Colors.transparent],
                  ),
                  border: Border.all(color: emeraldGreen, width: 2),
                  boxShadow: [
                    BoxShadow(
                      color: emeraldGreen.withValues(alpha: 0.4),
                      blurRadius: 20,
                      offset: const Offset(0, 6),
                    ),
                  ],
                ),
                child:  Icon(Icons.check_circle_rounded, color: emeraldGreen, size: 64),
              ),
              const SizedBox(height: 16),
              Text(
                'Order Completed & Processed!',
                style: GoogleFonts.inter(
                  fontSize: 20,
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  letterSpacing: -0.3,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Order #: ${_generatedOrderNumber ?? 'SO-SUCCESS'}',
                style: GoogleFonts.inter(
                  fontSize: 14,
                  color: primarySky,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 24),

              // 3 Verified Steps Card
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: cardBgColor,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: emeraldGreen.withValues(alpha: 0.5), width: 1.5),
                  boxShadow: [
                    BoxShadow(color: Colors.black.withValues(alpha: 0.4), blurRadius: 12, offset: const Offset(0, 4)),
                  ],
                ),
                child: Column(
                  children: [
                    _buildReceiptStepRow('Step 1: Payment Status', 'PAID', emeraldGreen),
                    const Divider(color: Colors.white10, height: 16),
                    _buildReceiptStepRow('Step 2: Fulfillment Status', 'FULFILLED', emeraldGreen),
                    const Divider(color: Colors.white10, height: 16),
                    _buildReceiptStepRow('Step 3: Order Status', 'COMPLETED', emeraldGreen),
                  ],
                ),
              ),
              const SizedBox(height: 20),

              // Invoice Details Card
              Container(
                padding: const EdgeInsets.all(18),
                decoration: BoxDecoration(
                  color: cardBgColor,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                  boxShadow: [
                    BoxShadow(color: Colors.black.withValues(alpha: 0.4), blurRadius: 10, offset: const Offset(0, 4)),
                  ],
                ),
                child: Column(
                  children: [
                    _buildReceiptRow('Customer', widget.customer['name'] ?? 'Walk-in'),
                    _buildReceiptRow('Payment Method', _selectedPaymentMethod.toUpperCase()),
                    _buildReceiptRow('Items Purchased', '${widget.cartItems.length} Products'),
                    const Divider(color: Colors.white10, height: 16),
                    _buildReceiptRow('Subtotal', 'SAR ${widget.subtotal.toStringAsFixed(2)}'),
                    _buildReceiptRow('VAT Tax (15%)', 'SAR ${widget.taxAmount.toStringAsFixed(2)}'),
                    _buildReceiptRow('Total Paid', 'SAR ${widget.totalAmount.toStringAsFixed(2)}', isBold: true),
                    if (_changeDue > 0)
                      _buildReceiptRow('Change Given', 'SAR ${_changeDue.toStringAsFixed(2)}', highlight: true),
                  ],
                ),
              ),
              const SizedBox(height: 28),

              // Start New Sale Button
              SizedBox(
                width: double.infinity,
                height: 52,
                child: ElevatedButton.icon(
                  onPressed: () {
                    widget.onOrderCompleted();
                    Navigator.of(context).pop();
                  },
                  icon: const Icon(Icons.add_shopping_cart_rounded, size: 20),
                  label: Text(
                    'Start New Sale',
                    style: GoogleFonts.inter(fontSize: 15, fontWeight: FontWeight.w800),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: primarySky,
                    foregroundColor: const Color(0xFF070B14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                    elevation: 6,
                    shadowColor: primarySky.withValues(alpha: 0.5),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildReceiptRow(String label, String value, {bool isBold = false, bool highlight = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: GoogleFonts.inter(fontSize: 13, color: Colors.white60),
          ),
          Text(
            value,
            style: GoogleFonts.inter(
              fontSize: 13.5,
              fontWeight: isBold ? FontWeight.w900 : FontWeight.w600,
              color: highlight ? const Color(0xFF38BDF8) : Colors.white,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildReceiptStepRow(String stepName, String status, Color color) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Row(
          children: [
            Icon(Icons.check_circle_rounded, color: color, size: 19),
            const SizedBox(width: 8),
            Text(
              stepName,
              style: GoogleFonts.inter(fontSize: 13.5, fontWeight: FontWeight.w700, color: Colors.white),
            ),
          ],
        ),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.2),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: color.withValues(alpha: 0.4)),
          ),
          child: Text(
            status,
            style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.w900, color: color),
          ),
        ),
      ],
    );
  }
}
