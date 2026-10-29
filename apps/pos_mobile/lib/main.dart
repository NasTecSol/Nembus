import 'package:flutter/material.dart';
import 'package:pos_mobile/singleton/singleton_class.dart';
import 'screens/splash_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  SingletonClass().init();
  runApp(const NembusMobileApp());
}

class NembusMobileApp extends StatefulWidget {
  const NembusMobileApp({super.key});

  @override
  State<NembusMobileApp> createState() => _NembusMobileAppState();
}

class _NembusMobileAppState extends State<NembusMobileApp> with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.detached) {
      // App is being removed / killed by user or OS
      SingletonClass().autoCloseActiveSession(reason: 'Auto-closed on app termination/kill');
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Nembus POS Mobile',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        scaffoldBackgroundColor: const Color(0xFF0F172A),
        colorScheme: const ColorScheme.dark(
          primary: Color(0xFF38BDF8),
          surface: Color(0xFF1E293B),
        ),
        useMaterial3: true,
      ),
      home: const SplashScreen(),
    );
  }
}
