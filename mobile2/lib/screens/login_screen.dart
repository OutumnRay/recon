import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../l10n/app_localizations.dart';
import '../services/api_client.dart';
import '../services/auth_service.dart';
import '../services/config_service.dart';
import '../services/locale_service.dart';
import '../services/fcm_service.dart';
import '../utils/logger.dart';
import '../widgets/error_display.dart';
import '../widgets/app_card.dart';
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

    final authService = AuthService(ApiClient(baseUrl: url));
    final isLoggedIn = await authService.isLoggedIn();
    if (isLoggedIn && mounted) {
      _navigateToHome(url);
    }
  }

  Future<void> _saveApiUrl() async {
    final url = _apiUrlController.text.trim();
    if (url.isEmpty) {
      setState(() {
        _error = 'API URL cannot be empty';
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
      const SnackBar(content: Text('API URL saved')),
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
      final apiClient = ApiClient(baseUrl: _currentApiUrl);
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
        setState(() {
          _error =
              'Connection Error:\n${e.toString()}\n\nAPI URL: $_currentApiUrl\n\nPlease check:\n• Network connection\n• API URL is correct\n• Server is running';
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
                GradientHeroCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Container(
                        width: 58,
                        height: 58,
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Image.asset('assets/icon/app_icon.png'),
                      ),
                      const SizedBox(height: 18),
                      Text(
                        l10n.welcomeBackTitle,
                        style: Theme.of(context).textTheme.headlineMedium,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        l10n.welcomeBackSubtitle,
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                SurfaceCard(
                  child: Form(
                    key: _formKey,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(
                          l10n.login,
                          style: Theme.of(context)
                              .textTheme
                              .titleLarge
                              ?.copyWith(fontWeight: FontWeight.w700),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          l10n.loginSubtitle,
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                        const SizedBox(height: 24),
                        TextFormField(
                          controller: _usernameController,
                          decoration: InputDecoration(
                            labelText: l10n.username,
                            prefixIcon:
                                const Icon(Icons.person_outline_rounded),
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
                            prefixIcon: const Icon(Icons.lock_outline_rounded),
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
                            icon: const Icon(Icons.settings),
                            label: Text(l10n.serverConfiguration),
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
                                  labelText: 'API URL',
                                  hintText: 'https://portal.recontext.online',
                                  prefixIcon:
                                      const Icon(Icons.link_outlined),
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
                                      child: Text(l10n.cancel),
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: ElevatedButton(
                                      onPressed: _saveApiUrl,
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
                        ElevatedButton(
                          onPressed: _isLoading ? null : _login,
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
                              : Text(l10n.loginButton),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _currentApiUrl.isEmpty
                              ? l10n.apiNotConfigured
                              : '${l10n.currentServer}: $_currentApiUrl',
                          style: Theme.of(context)
                              .textTheme
                              .bodySmall
                              ?.copyWith(color: AppColors.textTertiary),
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
