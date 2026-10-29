import 'dart:developer' as developer;
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:sqflite/sqflite.dart';

class DatabaseService {
  static DatabaseService? _instance;
  Database? _db;

  DatabaseService._();

  factory DatabaseService() {
    _instance ??= DatabaseService._();
    return _instance!;
  }

  Database get database {
    if (_db == null) {
      throw StateError('DatabaseService has not been initialized. Call initDatabase() first.');
    }
    return _db!;
  }

  bool get isOpen => _db != null && _db!.isOpen;

  /// Returns the absolute path where the local database resides.
  Future<String> getDatabasePath() async {
    final docsDir = await getApplicationDocumentsDirectory();
    return p.join(docsDir.path, 'nembus_pos_local.db');
  }

  /// Initializes the SQLite engine with WAL mode and NORMAL synchronous writes.
  /// All logs are directed exclusively to the developer console.
  Future<void> initDatabase() async {
    if (_db != null && _db!.isOpen) return;

    final dbPath = await getDatabasePath();
    developer.log('Initializing local SQLite at: $dbPath', name: 'DatabaseService');

    _db = await openDatabase(
      dbPath,
      version: 1,
      onConfigure: (db) async {
        // High concurrency and crash durability settings
        await db.rawQuery('PRAGMA journal_mode = WAL;');
        await db.execute('PRAGMA synchronous = NORMAL;');
        await db.execute('PRAGMA foreign_keys = ON;');
      },
      onCreate: (db, version) async {
        developer.log('Local SQLite file instantiated.', name: 'DatabaseService');
      },
    );

    developer.log('Local SQLite ready (WAL mode active).', name: 'DatabaseService');
  }

  /// Cleans / wipes all data and tables by resetting the database file.
  Future<void> purgeDatabase() async {
    developer.log('Purging local SQLite database file...', name: 'DatabaseService');
    final dbPath = await getDatabasePath();
    if (_db != null && _db!.isOpen) {
      await _db!.close();
      _db = null;
    }
    await deleteDatabase(dbPath);
    await initDatabase();
    developer.log('Local SQLite database cleanly reset.', name: 'DatabaseService');
  }

  Future<void> close() async {
    if (_db != null && _db!.isOpen) {
      await _db!.close();
      _db = null;
    }
  }
}