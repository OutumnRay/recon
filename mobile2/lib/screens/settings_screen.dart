import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../l10n/app_localizations.dart';
import '../services/api_client.dart';
import '../services/auth_service.dart';
import '../services/config_service.dart';
import '../services/locale_service.dart';
import '../services/storage_service.dart';
import '../theme/app_colors.dart';
import 'login_screen.dart';
import 'profile_screen.dart';

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
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        title: Row(
          children: [
            const Icon(Icons.logout, color: AppColors.danger),
            const SizedBox(width: 8),
            Text(
              l10n.logout,
              style: const TextStyle(fontWeight: FontWeight.bold, color: AppColors.textPrimary),
            ),
          ],
        ),
        content: Text(
          l10n.logoutConfirm,
          style: const TextStyle(color: AppColors.textSecondary),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            style: TextButton.styleFrom(
              foregroundColor: AppColors.primary500,
            ),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            style: FilledButton.styleFrom(
              backgroundColor: AppColors.danger,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
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
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        title: Text(
          l10n.changeLanguage,
          style: const TextStyle(color: AppColors.primary500, fontWeight: FontWeight.bold),
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            RadioListTile<String>(
              title: Text(l10n.english, style: const TextStyle(color: AppColors.textPrimary)),
              value: 'en',
              groupValue: localeService.locale.languageCode,
              activeColor: AppColors.primary500,
              onChanged: (value) => Navigator.pop(context, value),
            ),
            RadioListTile<String>(
              title: Text(l10n.russian, style: const TextStyle(color: AppColors.textPrimary)),
              value: 'ru',
              groupValue: localeService.locale.languageCode,
              activeColor: AppColors.primary500,
              onChanged: (value) => Navigator.pop(context, value),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            style: TextButton.styleFrom(
              foregroundColor: AppColors.primary500,
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
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: Text(l10n.settingsTitle),
        backgroundColor: Colors.white,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        scrolledUnderElevation: 0,
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              children: [
                // User Profile Section
                Container(
                  margin: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(24), // radius-xl
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.08),
                        blurRadius: 30,
                        offset: const Offset(0, 12),
                      ),
                    ],
                    border: Border.all(color: AppColors.border),
                  ),
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    children: [
                      CircleAvatar(
                        radius: 50,
                        backgroundColor: AppColors.primary50,
                        child: const Icon(
                          Icons.person,
                          size: 50,
                          color: AppColors.primary500,
                        ),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        _username ?? l10n.unknownUser,
                        style: const TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.w600,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        _email ?? '',
                        style: const TextStyle(
                          fontSize: 14,
                          color: AppColors.textSecondary,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                        decoration: BoxDecoration(
                          color: AppColors.primary50,
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(color: AppColors.primary200),
                        ),
                        child: Text(
                          _getRoleDisplayName(_role),
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: AppColors.primary600,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),

                // Server Configuration Section
                Container(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(24), // radius-xl
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.08),
                        blurRadius: 30,
                        offset: const Offset(0, 12),
                      ),
                    ],
                    border: Border.all(color: AppColors.border),
                  ),
                  child: Column(
                    children: [
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.dns, color: AppColors.primary500),
                        ),
                        title: Text(
                          l10n.serverConfiguration,
                          style: const TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        subtitle: Text(
                          _currentApiUrl,
                          style: const TextStyle(
                            fontSize: 12,
                            color: AppColors.textSecondary,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        trailing: Icon(
                          _showApiConfig
                              ? Icons.keyboard_arrow_up
                              : Icons.keyboard_arrow_down,
                          color: AppColors.primary500,
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
                                    borderRadius: BorderRadius.circular(14), // radius-lg
                                    borderSide: const BorderSide(color: AppColors.border),
                                  ),
                                  enabledBorder: OutlineInputBorder(
                                    borderRadius: BorderRadius.circular(14),
                                    borderSide: const BorderSide(color: AppColors.border),
                                  ),
                                  focusedBorder: OutlineInputBorder(
                                    borderRadius: BorderRadius.circular(14),
                                    borderSide: const BorderSide(color: AppColors.primary400, width: 2),
                                  ),
                                  filled: true,
                                  fillColor: AppColors.surfaceMuted,
                                  prefixIcon: const Icon(Icons.link, color: AppColors.primary500),
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
                                        foregroundColor: AppColors.primary500,
                                        side: const BorderSide(color: AppColors.primary500, width: 1.5),
                                        padding: const EdgeInsets.symmetric(vertical: 14),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(14),
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
                                        backgroundColor: AppColors.primary500,
                                        padding: const EdgeInsets.symmetric(vertical: 14),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(14),
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
                                style: const TextStyle(
                                  fontSize: 12,
                                  color: AppColors.warning,
                                ),
                              ),
                            ],
                          ),
                        ),
                    ],
                  ),
                ),

                // App Settings Section
                Container(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(24), // radius-xl
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.08),
                        blurRadius: 30,
                        offset: const Offset(0, 12),
                      ),
                    ],
                    border: Border.all(color: AppColors.border),
                  ),
                  child: Column(
                    children: [
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.person_outline, color: AppColors.primary500),
                        ),
                        title: const Text(
                          'My Profile',
                          style: TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        subtitle: const Text(
                          'Edit your profile information',
                          style: TextStyle(color: AppColors.textSecondary),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: AppColors.primary500),
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => ProfileScreen(apiClient: widget.apiClient),
                            ),
                          );
                        },
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.language, color: AppColors.primary500),
                        ),
                        title: Text(
                          AppLocalizations.of(context)!.language,
                          style: const TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        subtitle: Text(
                          Provider.of<LocaleService>(context).isRussian
                              ? AppLocalizations.of(context)!.russian
                              : AppLocalizations.of(context)!.english,
                          style: const TextStyle(color: AppColors.textSecondary),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: AppColors.primary500),
                        onTap: () => _showLanguageDialog(),
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.notifications, color: AppColors.primary500),
                        ),
                        title: Text(
                          l10n.notifications,
                          style: const TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: AppColors.primary500),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.notificationsComingSoon),
                              backgroundColor: AppColors.primary500,
                            ),
                          );
                        },
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.security, color: AppColors.primary500),
                        ),
                        title: Text(
                          l10n.privacySecurity,
                          style: const TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: AppColors.primary500),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.privacyComingSoon),
                              backgroundColor: AppColors.primary500,
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                ),

                // About Section
                Container(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(24), // radius-xl
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.08),
                        blurRadius: 30,
                        offset: const Offset(0, 12),
                      ),
                    ],
                    border: Border.all(color: AppColors.border),
                  ),
                  child: Column(
                    children: [
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.info_outline, color: AppColors.primary500),
                        ),
                        title: Text(
                          l10n.about,
                          style: const TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: AppColors.primary500),
                        onTap: () {
                          showAboutDialog(
                            context: context,
                            applicationName: 'Recontext',
                            applicationVersion: '1.0.0',
                            applicationIcon: Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: AppColors.primary50,
                                borderRadius: BorderRadius.circular(16),
                              ),
                              child: const Icon(Icons.video_call, size: 48, color: AppColors.primary500),
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
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.help_outline, color: AppColors.primary500),
                        ),
                        title: Text(
                          l10n.helpSupport,
                          style: const TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: AppColors.primary500),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.helpComingSoon),
                              backgroundColor: AppColors.primary500,
                            ),
                          );
                        },
                      ),
                      const Divider(height: 1, indent: 72, endIndent: 16),
                      ListTile(
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primary50,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(Icons.description_outlined, color: AppColors.primary500),
                        ),
                        title: Text(
                          l10n.termsConditions,
                          style: const TextStyle(fontWeight: FontWeight.w500, color: AppColors.textPrimary),
                        ),
                        trailing: const Icon(Icons.chevron_right, color: AppColors.primary500),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(l10n.termsComingSoon),
                              backgroundColor: AppColors.primary500,
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
                        foregroundColor: AppColors.danger,
                        side: const BorderSide(color: AppColors.danger, width: 2),
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(14), // radius-lg
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
                    style: const TextStyle(
                      fontSize: 12,
                      color: AppColors.textTertiary,
                    ),
                  ),
                ),
                const SizedBox(height: 8),
                Center(
                  child: Text(
                    l10n.allRightsReserved,
                    style: const TextStyle(
                      fontSize: 12,
                      color: AppColors.textTertiary,
                    ),
                  ),
                ),
                const SizedBox(height: 32),
              ],
            ),
    );
  }
}
