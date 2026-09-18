import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:http/http.dart' as http;

import '../controller/tenant_model.dart';
import '../singleton/singleton_class.dart';
import 'db_cloning_screen.dart';

typedef TenantSelectionScreen = CompanySelectionScreen;

class CompanySelectionScreen extends StatefulWidget {
  const CompanySelectionScreen({super.key});

  @override
  State<CompanySelectionScreen> createState() => _CompanySelectionScreenState();
}

class _CompanySelectionScreenState extends State<CompanySelectionScreen> {
  final TextEditingController _searchController = TextEditingController();
  final GlobalKey<FormState> _formKey = GlobalKey();
  final SingletonClass singletonClass = SingletonClass();

  List<Data> _suggestions = [];
  Data? _selectedTenant;
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_onSearchInputChanged);
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchInputChanged);
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchInputChanged() {
    final text = _searchController.text.trim();
    if (_selectedTenant != null && _selectedTenant!.slug != text && _selectedTenant!.tenantName != text) {
      setState(() {
        _selectedTenant = null;
      });
    }

    if (text.isEmpty) {
      setState(() {
        _suggestions.clear();
        _errorMessage = null;
      });
    }
  }

  /// Fetches tenant suggestions from the cloud (ONLINE phase only).
  Future<void> getTenantSuggestions(String query) async {
    final trimmed = query.trim().toLowerCase();
    if (trimmed.isEmpty) {
      setState(() {
        _suggestions.clear();
        _errorMessage = null;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final uri = Uri.parse('${singletonClass.baseURL}/api/tenants/active');
      final response = await http.get(
        uri,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ).timeout(const Duration(seconds: 10));

      if (response.statusCode == 200) {
        final decoded = json.decode(response.body);

        List<Data> parsedData = [];
        if (decoded is List) {
          parsedData = decoded
              .whereType<Map<String, dynamic>>()
              .map((e) => Data.fromJson(e))
              .toList();
        } else if (decoded is Map<String, dynamic>) {
          final model = TenantModel.fromJson(decoded);
          parsedData = model.data ?? [];
        }

        if (parsedData.isNotEmpty) {
          setState(() {
            _suggestions = parsedData;
            _isLoading = false;
          });
        } else {
          setState(() {
            _suggestions.clear();
            _isLoading = false;
            _errorMessage = 'No active tenants found matching "$trimmed"';
          });
        }
      } else {
        setState(() {
          _suggestions.clear();
          _isLoading = false;
          _errorMessage = 'Server returned status ${response.statusCode}';
        });
      }
    } catch (e) {
      setState(() {
        _suggestions.clear();
        _isLoading = false;
        _errorMessage = 'Error looking up tenant: $e';
      });
    }
  }

  void _selectTenant(Data tenant) {
    setState(() {
      _selectedTenant = tenant;
      _searchController.text = tenant.slug ?? tenant.tenantName ?? '';
      _suggestions.clear();
    });
  }

  void _navigateToNext() {
    if (_selectedTenant == null) return;
    Navigator.of(context).pushReplacement(
      MaterialPageRoute(
        builder: (_) => DbCloningScreen(
          tenantId: _selectedTenant!.id ?? '',
          tenantSlug: _selectedTenant!.slug ?? '',
          tenantName: _selectedTenant!.tenantName ?? 'Store Terminal',
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    const bgColor = Color(0xFF0F172A); // Dark Slate Background
    const cardColor = Color(0xFF1E293B); // Surface Card Color
    const accentColor = Color(0xFF38BDF8); // Cyan Accent
    const primaryGlow = Color(0xFF6366F1); // Indigo Glow

    return Scaffold(
      backgroundColor: bgColor,
      appBar: AppBar(
        backgroundColor: bgColor,
        elevation: 0,
        centerTitle: true,
        title: Text(
          'Tenant Verification',
          style: GoogleFonts.inter(
            color: Colors.white,
            fontWeight: FontWeight.bold,
            fontSize: 18,
          ),
        ),
      ),
      body: SafeArea(
        child: Form(
          key: _formKey,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Header Logo Icon & Title
                Center(
                  child: Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: cardColor,
                      shape: BoxShape.circle,
                      boxShadow: [
                        BoxShadow(
                          color: primaryGlow.withValues(alpha: 0.3),
                          blurRadius: 20,
                          spreadRadius: 3,
                        ),
                      ],
                      border: Border.all(color: accentColor.withValues(alpha: 0.4), width: 1.5),
                    ),
                    child: const Icon(
                      Icons.store_rounded,
                      size: 40,
                      color: accentColor,
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                Text(
                  'Select POS Company',
                  textAlign: TextAlign.center,
                  style: GoogleFonts.inter(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                    letterSpacing: 0.5,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Search your tenant identifier to connect your POS terminal',
                  textAlign: TextAlign.center,
                  style: GoogleFonts.inter(
                    fontSize: 13,
                    color: Colors.grey[400],
                  ),
                ),
                const SizedBox(height: 24),

                // Search Input Field Row
                Row(
                  children: [
                    Expanded(
                      child: Container(
                        decoration: BoxDecoration(
                          color: cardColor,
                          borderRadius: BorderRadius.circular(16),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withValues(alpha: 0.2),
                              blurRadius: 8,
                              offset: const Offset(0, 2),
                            ),
                          ],
                        ),
                        child: TextFormField(
                          controller: _searchController,
                          validator: (value) {
                            if (value == null || value.trim().isEmpty) {
                              return 'Please enter a company name or identifier';
                            }
                            return null;
                          },
                          cursorColor: accentColor,
                          style: GoogleFonts.inter(
                            fontSize: 15,
                            fontWeight: FontWeight.w500,
                            color: Colors.white,
                          ),
                          decoration: InputDecoration(
                            hintText: 'Search company identifier...',
                            hintStyle: GoogleFonts.inter(
                              color: Colors.grey.shade500,
                              fontSize: 14,
                            ),
                            prefixIcon: Icon(
                              Icons.business_rounded,
                              color: accentColor.withValues(alpha: 0.8),
                              size: 22,
                            ),
                            suffixIcon: _isLoading
                                ? const Padding(
                                    padding: EdgeInsets.all(12.0),
                                    child: SizedBox(
                                      width: 20,
                                      height: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2.5,
                                        color: accentColor,
                                      ),
                                    ),
                                  )
                                : _searchController.text.isNotEmpty
                                    ? IconButton(
                                        onPressed: () {
                                          _searchController.clear();
                                          setState(() {
                                            _selectedTenant = null;
                                            _suggestions.clear();
                                            _errorMessage = null;
                                          });
                                        },
                                        icon: Icon(
                                          Icons.cancel_rounded,
                                          color: Colors.grey.shade500,
                                          size: 20,
                                        ),
                                      )
                                    : null,
                            contentPadding: const EdgeInsets.symmetric(
                              vertical: 16,
                              horizontal: 16,
                            ),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(16.0),
                              borderSide: BorderSide.none,
                            ),
                            focusedBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(16.0),
                              borderSide: const BorderSide(
                                color: accentColor,
                                width: 2,
                              ),
                            ),
                          ),
                          onFieldSubmitted: (val) {
                            if (_formKey.currentState!.validate()) {
                              getTenantSuggestions(val);
                            }
                          },
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    GestureDetector(
                      onTap: () {
                        if (_formKey.currentState!.validate()) {
                          getTenantSuggestions(_searchController.text);
                        }
                      },
                      child: Container(
                        height: 54,
                        width: 54,
                        decoration: BoxDecoration(
                          color: accentColor,
                          borderRadius: BorderRadius.circular(16),
                          boxShadow: [
                            BoxShadow(
                              color: accentColor.withValues(alpha: 0.35),
                              blurRadius: 10,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: const Icon(
                          Icons.search_rounded,
                          color: Colors.black,
                          size: 24,
                        ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 16),

                // Selected Tenant Banner
                if (_selectedTenant != null) ...[
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: const Color(0xFF10B981).withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(14),
                      border: Border.all(color: const Color(0xFF10B981)),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.check_circle_rounded, color: Color(0xFF10B981), size: 28),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                _selectedTenant!.tenantName ?? 'Selected Tenant',
                                style: GoogleFonts.inter(
                                  color: Colors.white,
                                  fontWeight: FontWeight.bold,
                                  fontSize: 16,
                                ),
                              ),
                              const SizedBox(height: 2),
                              Text(
                                'Slug: ${_selectedTenant!.slug ?? ""} • ID: ${_selectedTenant!.id ?? ""}',
                                style: const TextStyle(
                                  color: Colors.white70,
                                  fontSize: 12,
                                  fontFamily: 'monospace',
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                ],

                // Error Notice
                if (_errorMessage != null && _suggestions.isEmpty) ...[
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 8.0),
                    child: Text(
                      _errorMessage!,
                      style: GoogleFonts.inter(color: Colors.amberAccent, fontSize: 13),
                    ),
                  ),
                ],

                // Tenant Suggestions Result List
                Expanded(
                  child: _suggestions.isNotEmpty
                      ? ListView.separated(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                          itemCount: _suggestions.length,
                          separatorBuilder: (context, index) => const SizedBox(height: 10),
                          itemBuilder: (context, index) {
                            final suggestion = _suggestions[index];
                            final isSelected = _selectedTenant?.id == suggestion.id;

                            return InkWell(
                              onTap: () => _selectTenant(suggestion),
                              borderRadius: BorderRadius.circular(16),
                              child: AnimatedContainer(
                                duration: const Duration(milliseconds: 200),
                                padding: const EdgeInsets.all(16),
                                decoration: BoxDecoration(
                                  color: isSelected ? cardColor : const Color(0xFF182234),
                                  borderRadius: BorderRadius.circular(16),
                                  border: Border.all(
                                    color: isSelected ? accentColor : Colors.white12,
                                    width: isSelected ? 2 : 1,
                                  ),
                                  boxShadow: [
                                    BoxShadow(
                                      color: isSelected
                                          ? accentColor.withValues(alpha: 0.2)
                                          : Colors.black.withValues(alpha: 0.1),
                                      blurRadius: isSelected ? 12 : 4,
                                      offset: const Offset(0, 4),
                                    ),
                                  ],
                                ),
                                child: Row(
                                  children: [
                                    Container(
                                      height: 44,
                                      width: 44,
                                      decoration: BoxDecoration(
                                        shape: BoxShape.circle,
                                        color: accentColor.withValues(alpha: 0.15),
                                        border: Border.all(
                                          color: accentColor.withValues(alpha: 0.3),
                                        ),
                                      ),
                                      child: const Icon(
                                        Icons.storefront_rounded,
                                        color: accentColor,
                                        size: 22,
                                      ),
                                    ),
                                    const SizedBox(width: 14),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            suggestion.tenantName ?? 'Unnamed Store',
                                            style: GoogleFonts.inter(
                                              fontSize: 16,
                                              fontWeight: FontWeight.w600,
                                              color: Colors.white,
                                            ),
                                          ),
                                          const SizedBox(height: 4),
                                          Text(
                                            'Identifier: ${suggestion.slug ?? suggestion.id ?? ""}',
                                            style: const TextStyle(
                                              fontSize: 12,
                                              color: Colors.white70,
                                              fontFamily: 'monospace',
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    if (isSelected)
                                      const Icon(Icons.radio_button_checked_rounded, color: accentColor, size: 24)
                                    else
                                      Icon(Icons.radio_button_unchecked_rounded, color: Colors.grey[600], size: 24),
                                  ],
                                ),
                              ),
                            );
                          },
                        )
                      : const SizedBox.shrink(),
                ),

                const SizedBox(height: 16),

                // Primary Next Action Button
                ElevatedButton(
                  onPressed: _selectedTenant != null ? _navigateToNext : null,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: accentColor,
                    disabledBackgroundColor: Colors.white12,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: _selectedTenant != null ? 6 : 0,
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'Next',
                        style: GoogleFonts.inter(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: _selectedTenant != null ? Colors.black : Colors.grey[600],
                        ),
                      ),
                      const SizedBox(width: 8),
                      Icon(
                        Icons.arrow_forward_rounded,
                        color: _selectedTenant != null ? Colors.black : Colors.grey[600],
                        size: 20,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}