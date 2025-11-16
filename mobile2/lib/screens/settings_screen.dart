import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api_client.dart';
import '../services/auth_service.dart';
import '../services/config_service.dart';
import '../services/storage_service.dart';
import '../services/locale_service.dart';
import 'login_screen.dart';
import '../l10n/app_localizations.dart';

class SettingsScreen extends StatefulWidget {
  final ApiClient apiClient;
  final String apiUrl;

  const SettingsScreen({
    super.key,
    required this.apiClient,
    required this.apiUrl,
  });

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen>
    with AutomaticKeepAliveClientMixin {
  final _configService = ConfigService();
  final _storageService = StorageService();
  final _apiUrlController = TextEditingController();

  bool _isLoading = true;
  String? _username;
  String? _email;
  String? _role;
  String _currentApiUrl = '';
  bool _showApiConfig = false;

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    _loadUserData();
    _loadApiUrl();
  }

  @override
  void dispose() {
    _apiUrlController.dispose();
    super.dispose();
  }

  Future<void> _loadUserData() async {
    setState(() => _isLoading = true);

    try {
      final username = await _storageService.getUsername();
      final email = await _storageService.getEmail();
      final role = await _storageService.getRole();

      if (mounted) {
        setState(() {
          _username = username;
          _email = email;
          _role = role;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _loadApiUrl() async {
    final url = await _configService.getApiUrl();
    setState(() {
      _currentApiUrl = url;
      _apiUrlController.text = url;
    });
  }

  Future<void> _saveApiUrl() async {
    final l10n = AppLocalizations.of(context)!;
    final url = _apiUrlController.text.trim();
    if (url.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.apiUrlEmpty)),
      );
      return;
    }

    await _configService.saveApiUrl(url);
    setState(() {
      _currentApiUrl = url;
      _showApiConfig = false;
    });

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(l10n.apiUrlSaved),
          duration: const Duration(seconds: 3),
        ),
      );
    }
  }

  Future<void> _logout() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Row(
          children: [
            const Icon(Icons.logout, color: Colors.red),
            const SizedBox(width: 8),
            Text(
              l10n.logout,
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
          ],
        ),
        content: Text(l10n.logoutConfirm),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            style: TextButton.styleFrom(
              foregroundColor: const Color(0xFF26C6DA),
            ),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            style: FilledButton.styleFrom(
              backgroundColor: Colors.red,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            ),
            child: Text(l10n.logout),
          ),
        ],
      ),
    );

    if (confirmed == true && mounted) {
      final authService = AuthService(widget.apiClient);
      await authService.logout();

      if (mounted) {
        Navigator.of(context).pushAndRemoveUntil(
          MaterialPageRoute(builder: (_) => const LoginScreen()),
          (route) => false,
        );
      }
    }
  }

  String _getRoleDisplayName(String? role) {
    final l10n = AppLocalizations.of(context)!;
    if (role == null) return l10n.userRoleUser;
    switch (role.toLowerCase()) {
      case 'admin':
        return l10n.userRoleAdmin;
      case 'manager':
        return l10n.userRoleManager;
      case 'user':
        return l10n.userRoleUser;
      default:
        return role;
    }
  }

  Future<void> _showLanguageDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final localeService = Provider.of<LocaleService>(context, listen: false);

    final selected = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text(
          l10n.changeLanguage,
          style: const TextStyle(color: Color(0xFF26C6DA), fontWeight: FontWeight.bold),
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            RadioListTile<String>(
              title: Text(l10n.english),
              value: 'en',
              groupValue: localeService.locale.languageCode,
              activeColor: const Color(0xFF26C6DA),
              onChanged: (value) => Navigator.pop(context, value),
            ),
            RadioListTile<String>(
              title: Text(l10n.russian),
              value: 'ru',
              groupValue: localeService.locale.languageCode,
              activeColor: const Color(0xFF26C6DA),
              onChanged: (value) => Navigator.pop(context, value),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            style: TextButton.styleFrom(
              foregroundColor: const Color(0xFF26C6DA),
            ),
            child: Text(l10n.cancel),
          ),
        ],
      ),
    );

    if (selected != null && selected != localeService.locale.languageCode) {
      await localeService.setLocale(Locale(selected));
    }
  }

  @override
  Widget build(BuildContext context) {
    super.build(context); // Required for AutomaticKeepAliveClientMixin
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: Text(l10n.settingsTitle),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF26C6DA),
        elevation: 1,
        shadowColor: Colors.black.withValues(alpha: 0.1),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              children: [
                // User Profile Section
                Card(
                  margin: const EdgeInsets.all(16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(24),
                  ),
                  elevation: 2,
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      children: [
                        CircleAvatar(
                          radius: 50,
                          backgroundColor: const Color(0xFF26C6DA).withValues(alpha: 0.15),
                          child: const Icon(
                            Icons.person,
                            size: 50,
                            color: Color(0xFF00ACC1),
                          ),
                        ),
                        const SizedBox(height: 16),
                        Text(
                          _username ?? l10n.unknownUser,
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          _email ?? '',
                          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                                color: Colors.grey[600],
                              ),
                        ),
                        const SizedBox(height: 8),
                        Chip(
                          label: Text(_getRoleDisplayName(_role)),
                          backgroundColor: Theme.of(context)
                              .colorScheme
                              .secondaryContainer,
                        ),
                      ],
                    ),
                  ),
                ),

                // Server Configuration Section
                Card(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                  elevation: 2,
                  child: Column(
                    children: [
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.dns, color: Color(0xFF26C6DA)),
                        ),
                        title: Text(
                          l10n.serverConfiguration,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        subtitle: Text(
                          _currentApiUrl,
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[600],
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        trailing: Icon(
                          _showApiConfig
                              ? Icons.keyboard_arrow_up
                              : Icons.keyboard_arrow_down,
                          color: const Color(0xFF26C6DA),
                        ),
                        onTap: () {
                          setState(() {
                            _showApiConfig = !_showApiConfig;
                          });
                        },
                      ),
                      if (_showApiConfig)
                        Padding(
                          padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              const Divider(),
                              const SizedBox(height: 8),
                              TextFormField(
                                controller: _apiUrlController,
                                decoration: InputDecoration(
                                  labelText: l10n.apiUrl,
                                  hintText: 'https://portal.recontext.online',
                                  border: OutlineInputBorder(
                                    borderRadius: BorderRadius.circular(16),
                                  ),
                                  enabledBorder: OutlineInputBorder(
                                    borderRadius: BorderRadius.circular(16),
                                    borderSide: BorderSide(color: Colors.grey.shade300),
                                  ),
                                  focusedBorder: OutlineInputBorder(
                                    borderRadius: BorderRadius.circular(16),
                                    borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                                  ),
                                  filled: true,
                                  fillColor: Colors.white,
                                  prefixIcon: const Icon(Icons.link, color: Color(0xFF26C6DA)),
                                ),
                                keyboardType: TextInputType.url,
                              ),
                              const SizedBox(height: 12),
                              Row(
                                children: [
                                  Expanded(
                                    child: OutlinedButton(
                                      onPressed: () {
                                        _apiUrlController.text =
                                            _configService.getDefaultApiUrl();
                                      },
                                      style: OutlinedButton.styleFrom(
                                        foregroundColor: const Color(0xFF26C6DA),
                                        side: const BorderSide(color: Color(0xFF26C6DA), width: 1.5),
                                        padding: const EdgeInsets.symmetric(vertical: 14),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(12),
                                        ),
                                      ),
                                      child: Text(l10n.defaultButton),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: FilledButton(
                                      onPressed: _saveApiUrl,
                                      style: FilledButton.styleFrom(
                                        backgroundColor: const Color(0xFF26C6DA),
                                        padding: const EdgeInsets.symmetric(vertical: 14),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(12),
                                        ),
                                      ),
                                      child: Text(l10n.save),
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Text(
                                l10n.restartNote,
                                style: TextStyle(
                                  fontSize: 12,
                                  color: Colors.orange[700],
                                ),
                              ),
                            ],
                          ),
                        ),
                    ],
                  ),
                ),

                // App Settings Section
                Card(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                  elevation: 2,
                  child: Column(
                    children: [
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.language, color: Color(0xFF26C6DA)),
                        ),
                        title: Text(
                          AppLocalizations.of(context)!.language,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        subtitle: Text(
                          Provider.of<LocaleService>(context).isRussian
                              ? AppLocalizations.of(context)!.russian
                              : AppLocalizations.of(context)!.english,
                        ),
                        trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                        onTap: () => _showLanguageDialog(),
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.notifications, color: Color(0xFF26C6DA)),
                        ),
                        title: Text(
                          l10n.notifications,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.notificationsComingSoon),
                              backgroundColor: const Color(0xFF26C6DA),
                            ),
                          );
                        },
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.security, color: Color(0xFF26C6DA)),
                        ),
                        title: Text(
                          l10n.privacySecurity,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.privacyComingSoon),
                              backgroundColor: const Color(0xFF26C6DA),
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                ),

                // About Section
                Card(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                  elevation: 2,
                  child: Column(
                    children: [
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.info_outline, color: Color(0xFF26C6DA)),
                        ),
                        title: Text(
                          l10n.about,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                        onTap: () {
                          showAboutDialog(
                            context: context,
                            applicationName: 'Recontext',
                            applicationVersion: '1.0.0',
                            applicationIcon: Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(16),
                              ),
                              child: const Icon(Icons.video_call, size: 48, color: Color(0xFF26C6DA)),
                            ),
                            children: [
                              Text(l10n.appDescription),
                            ],
                          );
                        },
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.help_outline, color: Color(0xFF26C6DA)),
                        ),
                        title: Text(
                          l10n.helpSupport,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.helpComingSoon),
                              backgroundColor: const Color(0xFF26C6DA),
                            ),
                          );
                        },
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.description_outlined, color: Color(0xFF26C6DA)),
                        ),
                        title: Text(
                          l10n.termsConditions,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.termsComingSoon),
                              backgroundColor: const Color(0xFF26C6DA),
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 16),

                // Logout Button
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: _logout,
                      icon: const Icon(Icons.logout),
                      label: Text(
                        l10n.logout,
                        style: const TextStyle(fontWeight: FontWeight.w500, fontSize: 16),
                      ),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: Colors.red,
                        side: const BorderSide(color: Colors.red, width: 2),
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                      ),
                    ),
                  ),
                ),

                const SizedBox(height: 32),

                // Version Info
                Center(
                  child: Text(
                    '${l10n.version} 1.0.0',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                ),
                const SizedBox(height: 8),
                Center(
                  child: Text(
                    l10n.allRightsReserved,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                ),
                const SizedBox(height: 32),
              ],
            ),
    );
  }
}
