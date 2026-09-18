import 'dart:developer' as developer;
import 'package:flutter/services.dart' show rootBundle;
import 'package:sqflite/sqflite.dart';
import 'db_service.dart';

class MigrationRunner {
  /// Loads and executes the SQL migrations under `migrations/` sequentially.
  static Future<void> applyMigrations({
    Function(String step, double progress)? onProgress,
  }) async {
    final db = DatabaseService().database;
    final migrationFiles = [
      'migrations/000001_mobile_schema.sql',
      'migrations/pos_extention.sql',
    ];

    for (int i = 0; i < migrationFiles.length; i++) {
      final path = migrationFiles[i];
      final label = 'Executing $path';
      developer.log(label, name: 'MigrationRunner');
      onProgress?.call(label, (i + 1) / migrationFiles.length);

      final rawSql = await rootBundle.loadString(path);
      await _executeSqlScript(db, rawSql);
    }
  }

  static Future<void> _executeSqlScript(Database db, String script) async {
    // 1. Remove comments line-by-line to prevent semicolons inside comments from breaking statements
    final cleanedLines = <String>[];
    for (final line in script.split('\n')) {
      final noComment = line.replaceAll(RegExp(r'--.*'), '').trim();
      if (noComment.isNotEmpty) {
        cleanedLines.add(noComment);
      }
    }
    var text = cleanedLines.join('\n');

    // 2. Remove PostgreSQL-specific routines and constructs unsupported in SQLite
    text = text
        .replaceAll(RegExp(r'CREATE\s+(OR\s+REPLACE\s+)?FUNCTION[\s\S]*?\$\$;?', caseSensitive: false), '')
        .replaceAll(RegExp(r'CREATE\s+TRIGGER[\s\S]*?;', caseSensitive: false), '')
        .replaceAll(RegExp(r'CREATE\s+EXTENSION[\s\S]*?;', caseSensitive: false), '')
        .replaceAll(RegExp(r'CREATE\s+TYPE[\s\S]*?;', caseSensitive: false), '')
        .replaceAll(RegExp(r'DROP\s+TYPE[\s\S]*?;', caseSensitive: false), '')
        .replaceAll(RegExp(r'ALTER\s+TABLE\s+\w+\s+ADD\s+CONSTRAINT[\s\S]*?;', caseSensitive: false), '')
        .replaceAll(RegExp(r'DEFAULT\s+uuid_generate_v4\(\)', caseSensitive: false), '');

    // 3. Replace PostgreSQL enum type columns
    const enumTypes = [
      'order_type',
      'order_status_v2',
      'order_status',
      'payment_status',
      'fulfillment_status',
      'cart_status',
      'cart_type',
      'invoice_type',
      'invoice_status',
      'quote_status',
      'zatca_doc_status',
    ];
    for (final et in enumTypes) {
      text = text.replaceAll(RegExp(r'(\b\w+\s+)' + et + r'\b', caseSensitive: false), r'$1VARCHAR(50)');
    }

    // 4. Map PostgreSQL column types to SQLite equivalents
    text = text
        .replaceAll(RegExp(r'\bUUID\b', caseSensitive: false), 'TEXT')
        .replaceAll(RegExp(r'\bJSONB\b', caseSensitive: false), 'TEXT')
        .replaceAll(RegExp(r'\bJSON\b', caseSensitive: false), 'TEXT')
        .replaceAll(RegExp(r'\bSERIAL\s+PRIMARY\s+KEY\b', caseSensitive: false), 'INTEGER PRIMARY KEY AUTOINCREMENT')
        .replaceAll(RegExp(r'\bBIGSERIAL\s+PRIMARY\s+KEY\b', caseSensitive: false), 'INTEGER PRIMARY KEY AUTOINCREMENT');

    // 5. Split statements and execute
    final statements = text
        .split(';')
        .map((s) => s.trim())
        .where((s) => s.isNotEmpty)
        .toList();

    await db.execute('PRAGMA foreign_keys = OFF;');
    final batch = db.batch();
    for (final stmt in statements) {
      batch.execute(stmt);
    }
    await batch.commit(noResult: true);
    await db.execute('PRAGMA foreign_keys = ON;');
  }
}