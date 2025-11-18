import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../l10n/app_localizations.dart';
import '../main.dart';
import '../services/api_client.dart';
import '../services/auth_service.dart';
import '../services/config_service.dart';
import '../services/locale_service.dart';
import '../services/fcm_service.dart';
import '../utils/logger.dart';
import '../widgets/error_display.dart';
import '../theme/app_colors.dart';
import 'main_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _usernameController =
      TextEditingController(text: 'admin@recontext.online');
  final _passwordController = TextEditingController(text: 'admin123');
  final _apiUrlController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  final _configService = ConfigService();

  bool _isLoading = false;
  bool _showApiConfig = false;
  String? _error;
  String _currentApiUrl = '';

  @override
  void initState() {
    super.initState();
    _loadApiUrl();
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    _apiUrlController.dispose();
    super.dispose();
  }

  Future<void> _loadApiUrl() async {
    final url = await _configService.getApiUrl();
    setState(() {
      _currentApiUrl = url;
      _apiUrlController.text = url;
    });

    final authService = AuthService(ApiClient(baseUrl: url, navigatorKey: navigatorKey));
    final isLoggedIn = await authService.isLoggedIn();
    if (isLoggedIn && mounted) {
      _navigateToHome(url);
    }
  }

  Future<void> _saveApiUrl() async {
    final l10n = AppLocalizations.of(context)!;
    final url = _apiUrlController.text.trim();
    if (url.isEmpty) {
      setState(() {
        _error = l10n.apiUrlEmpty;
      });
      return;
    }

    await _configService.saveApiUrl(url);
    if (!mounted) return;

    setState(() {
      _currentApiUrl = url;
      _showApiConfig = false;
      _error = null;
    });

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.apiUrlSaved)),
    );
  }

  Future<void> _login() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final apiClient = ApiClient(baseUrl: _currentApiUrl, navigatorKey: navigatorKey);
      final authService = AuthService(apiClient);

      await authService.login(
        _usernameController.text.trim(),
        _passwordController.text,
      );

      if (mounted) {
        final localeService =
            Provider.of<LocaleService>(context, listen: false);
        await localeService.loadUserLocale();

        await _initializeFCM(apiClient);
        _navigateToHome(_currentApiUrl);
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = 'API Error: ${e.message}\nStatus Code: ${e.statusCode}';
        });
      }
    } on Exception catch (e) {
      if (mounted) {
        final l10n = AppLocalizations.of(context)!;
        setState(() {
          _error =
              '${l10n.connectionError}:\n${e.toString()}\n\n${l10n.apiUrl}: $_currentApiUrl\n\n${l10n.pleaseCheck}\n• ${l10n.networkConnection}\n• ${l10n.apiUrlCorrect}\n• ${l10n.serverRunning}';
        });
      }
    } catch (e, stackTrace) {
      if (mounted) {
        setState(() {
          _error =
              'Unexpected Error:\n${e.toString()}\n\nStack Trace:\n${stackTrace.toString().split('\n').take(5).join('\n')}';
        });
      }
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _initializeFCM(ApiClient apiClient) async {
    try {
      Logger.logInfo('Initializing FCM for push notifications');
      final fcmService = FCMService(apiClient);
      await fcmService.initialize();
      if (fcmService.isInitialized) {
        await fcmService.registerDevice();
        Logger.logSuccess('FCM device registered successfully');
      }
    } catch (e) {
      Logger.logWarning('FCM initialization failed (non-critical): $e');
    }
  }

  void _navigateToHome(String apiUrl) {
    Navigator.of(context).pushReplacement(
      MaterialPageRoute(
        builder: (_) => MainScreen(apiUrl: apiUrl),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      body: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [
              AppColors.primary50,
              Colors.white,
              AppColors.surfaceMuted,
            ],
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
          ),
        ),
        child: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Hero Card with App Icon and Welcome Message
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [AppColors.primary50, Color(0xFFFDFDFF)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
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
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Container(
                        width: 58,
                        height: 58,
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: [
                            BoxShadow(
                              color: AppColors.primary500.withValues(alpha: 0.2),
                              blurRadius: 12,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: Image.asset('assets/icon/app_icon.png'),
                      ),
                      const SizedBox(height: 18),
                      Text(
                        l10n.welcomeBackTitle,
                        style: const TextStyle(
                          fontSize: 28,
                          fontWeight: FontWeight.w700,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        l10n.welcomeBackSubtitle,
                        style: const TextStyle(
                          fontSize: 16,
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                // Login Form Card
                Container(
                  padding: const EdgeInsets.all(24),
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
                  child: Form(
                    key: _formKey,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(
                          l10n.login,
                          style: const TextStyle(
                            fontSize: 22,
                            fontWeight: FontWeight.w700,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          l10n.loginSubtitle,
                          style: const TextStyle(
                            fontSize: 14,
                            color: AppColors.textSecondary,
                          ),
                        ),
                        const SizedBox(height: 24),
                        TextFormField(
                          controller: _usernameController,
                          decoration: InputDecoration(
                            labelText: l10n.username,
                            prefixIcon: const Icon(Icons.person_outline_rounded, color: AppColors.primary500),
                            filled: true,
                            fillColor: AppColors.surfaceMuted,
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
                          ),
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return l10n.validationUsernameRequired;
                            }
                            return null;
                          },
                          textInputAction: TextInputAction.next,
                        ),
                        const SizedBox(height: 16),
                        TextFormField(
                          controller: _passwordController,
                          decoration: InputDecoration(
                            labelText: l10n.password,
                            prefixIcon: const Icon(Icons.lock_outline_rounded, color: AppColors.primary500),
                            filled: true,
                            fillColor: AppColors.surfaceMuted,
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
                          ),
                          obscureText: true,
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return l10n.validationPasswordRequired;
                            }
                            return null;
                          },
                          onFieldSubmitted: (_) => _login(),
                        ),
                        const SizedBox(height: 12),
                        Align(
                          alignment: Alignment.centerRight,
                          child: TextButton.icon(
                            onPressed: () {
                              setState(() {
                                _showApiConfig = !_showApiConfig;
                              });
                            },
                            style: TextButton.styleFrom(
                              foregroundColor: AppColors.primary500,
                            ),
                            icon: const Icon(Icons.settings, size: 18),
                            label: Text(
                              l10n.serverConfiguration,
                              style: const TextStyle(fontSize: 14),
                            ),
                          ),
                        ),
                        AnimatedCrossFade(
                          crossFadeState: _showApiConfig
                              ? CrossFadeState.showFirst
                              : CrossFadeState.showSecond,
                          duration: const Duration(milliseconds: 250),
                          firstChild: Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              const Divider(),
                              const SizedBox(height: 12),
                              TextFormField(
                                controller: _apiUrlController,
                                decoration: InputDecoration(
                                  labelText: l10n.apiUrl,
                                  hintText: l10n.apiUrlHint,
                                  prefixIcon: const Icon(Icons.link_outlined, color: AppColors.primary500),
                                  filled: true,
                                  fillColor: AppColors.surfaceMuted,
                                  border: OutlineInputBorder(
                                    borderRadius: BorderRadius.circular(14),
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
                                ),
                                keyboardType: TextInputType.url,
                              ),
                              const SizedBox(height: 12),
                              Row(
                                children: [
                                  Expanded(
                                    child: OutlinedButton(
                                      onPressed: () =>
                                          setState(() => _showApiConfig = false),
                                      style: OutlinedButton.styleFrom(
                                        foregroundColor: AppColors.textSecondary,
                                        side: const BorderSide(color: AppColors.border),
                                        padding: const EdgeInsets.symmetric(vertical: 14),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(14),
                                        ),
                                      ),
                                      child: Text(l10n.cancel),
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: ElevatedButton(
                                      onPressed: _saveApiUrl,
                                      style: ElevatedButton.styleFrom(
                                        backgroundColor: AppColors.primary500,
                                        foregroundColor: Colors.white,
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
                            ],
                          ),
                          secondChild: const SizedBox.shrink(),
                        ),
                        const SizedBox(height: 12),
                        if (_error != null) ...[
                          ErrorDisplay(error: _error!, onRetry: _login),
                          const SizedBox(height: 12),
                        ],
                        // Gradient Login Button
                        Container(
                          height: 56,
                          decoration: BoxDecoration(
                            gradient: _isLoading
                                ? null
                                : const LinearGradient(
                                    colors: [AppColors.primary500, AppColors.primary600],
                                  ),
                            borderRadius: BorderRadius.circular(14),
                            boxShadow: _isLoading
                                ? null
                                : [
                                    BoxShadow(
                                      color: AppColors.primary500.withValues(alpha: 0.3),
                                      blurRadius: 12,
                                      offset: const Offset(0, 4),
                                    ),
                                  ],
                          ),
                          child: ElevatedButton(
                            onPressed: _isLoading ? null : _login,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.transparent,
                              foregroundColor: Colors.white,
                              shadowColor: Colors.transparent,
                              disabledBackgroundColor: AppColors.border,
                              disabledForegroundColor: AppColors.textTertiary,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(14),
                              ),
                            ),
                            child: _isLoading
                                ? const SizedBox(
                                    height: 20,
                                    width: 20,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2.5,
                                      valueColor:
                                          AlwaysStoppedAnimation(Colors.white),
                                    ),
                                  )
                                : Text(
                                    l10n.loginButton,
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _currentApiUrl.isEmpty
                              ? l10n.apiNotConfigured
                              : '${l10n.currentServer}: $_currentApiUrl',
                          style: const TextStyle(
                            fontSize: 12,
                            color: AppColors.textTertiary,
                          ),
                          textAlign: TextAlign.center,
                        ),
                      ],
                    ),
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
