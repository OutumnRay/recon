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
import 'main_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _usernameController = TextEditingController(text: 'admin@recontext.online');
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

    // Check if already logged in
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
        // Update locale from user profile
        final localeService = Provider.of<LocaleService>(context, listen: false);
        await localeService.loadUserLocale();

        // Initialize and register FCM device for push notifications
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
          _error = 'Connection Error:\n${e.toString()}\n\nAPI URL: $_currentApiUrl\n\nPlease check:\n• Network connection\n• API URL is correct\n• Server is running';
        });
      }
    } catch (e, stackTrace) {
      if (mounted) {
        setState(() {
          _error = 'Unexpected Error:\n${e.toString()}\n\nStack Trace:\n${stackTrace.toString().split('\n').take(5).join('\n')}';
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

      // Create FCM service instance
      final fcmService = FCMService(apiClient);

      // Initialize FCM and request permissions
      await fcmService.initialize();

      // Register device with backend
      if (fcmService.isInitialized) {
        await fcmService.registerDevice();
        Logger.logSuccess('FCM device registered successfully');
      }
    } catch (e) {
      // Don't block login if FCM fails
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
      backgroundColor: Colors.white,
      body: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [
              const Color(0xFF4DD0E1).withOpacity(0.05),
              Colors.white,
            ],
          ),
        ),
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24.0),
              child: Form(
                key: _formKey,
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Logo
                    Container(
                      padding: const EdgeInsets.all(20),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(32),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.08),
                            blurRadius: 20,
                            offset: const Offset(0, 10),
                          ),
                        ],
                      ),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(16),
                        child: Image.asset(
                          'assets/icon/app_icon.png',
                          width: 80,
                          height: 80,
                          fit: BoxFit.cover,
                        ),
                      ),
                    ),
                    const SizedBox(height: 32),

                    // Title
                    Text(
                      'Recontext',
                      style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                            fontWeight: FontWeight.bold,
                            color: const Color(0xFF00ACC1),
                          ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 8),

                    // Subtitle
                    Text(
                      'Video conferences & meeting management',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Colors.grey[600],
                          ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 48),

                  // API URL Configuration (Collapsible)
                  Card(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(24),
                    ),
                    elevation: 2,
                    child: Column(
                      children: [
                        ListTile(
                          leading: Icon(
                            Icons.settings,
                            color: Theme.of(context).colorScheme.primary,
                          ),
                          title: const Text('Server Configuration'),
                          subtitle: Text(
                            _currentApiUrl.isEmpty ? 'Not configured' : _currentApiUrl,
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey[600],
                            ),
                          ),
                          trailing: Icon(
                            _showApiConfig
                                ? Icons.keyboard_arrow_up
                                : Icons.keyboard_arrow_down,
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
                                    labelText: 'API URL',
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
                                          padding: const EdgeInsets.symmetric(vertical: 14),
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(16),
                                          ),
                                          side: const BorderSide(color: Color(0xFF26C6DA)),
                                          foregroundColor: const Color(0xFF26C6DA),
                                        ),
                                        child: const Text('Default'),
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    Expanded(
                                      child: FilledButton(
                                        onPressed: _saveApiUrl,
                                        style: FilledButton.styleFrom(
                                          padding: const EdgeInsets.symmetric(vertical: 14),
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(16),
                                          ),
                                          backgroundColor: const Color(0xFF26C6DA),
                                        ),
                                        child: const Text('Save'),
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 24),

                  // Username field
                  TextFormField(
                    controller: _usernameController,
                    decoration: InputDecoration(
                      labelText: l10n.username,
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
                      prefixIcon: const Icon(Icons.person, color: Color(0xFF26C6DA)),
                    ),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'Please enter username'; // TODO: Add to localization
                      }
                      return null;
                    },
                    textInputAction: TextInputAction.next,
                  ),
                  const SizedBox(height: 16),

                  // Password field
                  TextFormField(
                    controller: _passwordController,
                    decoration: InputDecoration(
                      labelText: l10n.password,
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
                      prefixIcon: const Icon(Icons.lock, color: Color(0xFF26C6DA)),
                    ),
                    obscureText: true,
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'Please enter password'; // TODO: Add to localization
                      }
                      return null;
                    },
                    onFieldSubmitted: (_) => _login(),
                  ),
                  const SizedBox(height: 24),

                  // Error message
                  if (_error != null) ...[
                    ErrorDisplay(
                      error: _error!,
                      onRetry: _login,
                    ),
                    const SizedBox(height: 16),
                  ],

                  // Login button
                  FilledButton(
                    onPressed: _isLoading ? null : _login,
                    style: FilledButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 18),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      backgroundColor: const Color(0xFF26C6DA),
                      elevation: 4,
                    ),
                    child: _isLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                          )
                        : Text(l10n.loginButton, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  ),
                  const SizedBox(height: 24),

                  // Version info
                  Center(
                    child: Text(
                      'Version 1.0.0+2 (Enhanced Error Display)',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Colors.grey[600],
                          ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
      ),
    );
  }
}
