import 'dart:async';
import 'package:flutter/material.dart';
import '../l10n/app_localizations.dart';
import '../services/api_client.dart';
import '../services/fcm_service.dart';
import 'meetings_screen.dart';
import 'search_screen.dart';
import 'documents_screen.dart';
import 'settings_screen.dart';

class MainScreen extends StatefulWidget {
  final String apiUrl;

  const MainScreen({super.key, required this.apiUrl});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> with WidgetsBindingObserver {
  int _selectedIndex = 0;
  late final ApiClient _apiClient;
  late final List<Widget> _screens;
  late final FCMService _fcmService;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _apiClient = ApiClient(baseUrl: widget.apiUrl);
    _fcmService = FCMService(_apiClient);
    _screens = [
      MeetingsScreen(apiClient: _apiClient),
      SearchScreen(apiClient: _apiClient),
      DocumentsScreen(apiClient: _apiClient),
      SettingsScreen(apiClient: _apiClient, apiUrl: widget.apiUrl),
    ];
    _initializePushNotifications();
  }

  void _initializePushNotifications() {
    Future.microtask(() async {
      try {
        await _fcmService.initialize();
        await _fcmService.registerDevice();
      } catch (e) {
        debugPrint('Failed to initialize push notifications: $e');
      }
    });
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _fcmService.registerDevice();
    }
  }

  void _onItemTapped(int index) {
    setState(() {
      _selectedIndex = index;
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      body: IndexedStack(
        index: _selectedIndex,
        children: _screens,
      ),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.05),
              blurRadius: 10,
              offset: const Offset(0, -5),
            ),
          ],
        ),
        child: NavigationBar(
          selectedIndex: _selectedIndex,
          onDestinationSelected: _onItemTapped,
          backgroundColor: Colors.white,
          indicatorColor: const Color(0xFF26C6DA).withOpacity(0.15),
          height: 70,
          labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
          destinations: [
            NavigationDestination(
              icon: const Icon(Icons.event_outlined),
              selectedIcon: const Icon(Icons.event, color: Color(0xFF26C6DA)),
              label: l10n.meetings,
            ),
            NavigationDestination(
              icon: const Icon(Icons.search_outlined),
              selectedIcon: const Icon(Icons.search, color: Color(0xFF26C6DA)),
              label: l10n.search,
            ),
            NavigationDestination(
              icon: const Icon(Icons.folder_outlined),
              selectedIcon: const Icon(Icons.folder, color: Color(0xFF26C6DA)),
              label: l10n.documents,
            ),
            NavigationDestination(
              icon: const Icon(Icons.settings_outlined),
              selectedIcon: const Icon(Icons.settings, color: Color(0xFF26C6DA)),
              label: l10n.settings,
            ),
          ],
        ),
      ),
    );
  }
}
