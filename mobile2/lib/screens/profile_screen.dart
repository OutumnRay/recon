import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import '../models/user.dart';
import '../services/users_service.dart';
import '../services/api_client.dart';
import '../services/storage_service.dart';
import '../theme/app_colors.dart';
import '../utils/logger.dart';

class ProfileScreen extends StatefulWidget {
  final ApiClient apiClient;

  const ProfileScreen({super.key, required this.apiClient});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final _formKey = GlobalKey<FormState>();
  final _firstNameController = TextEditingController();
  final _lastNameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _bioController = TextEditingController();

  final StorageService _storageService = StorageService();
  late final UsersService _usersService;

  User? _user;
  bool _isLoading = true;
  bool _isEditMode = false;
  bool _isSaving = false;
  String? _errorMessage;
  String? _successMessage;
  File? _selectedImage;
  String? _avatarUrl;
  String _selectedLanguage = 'en';
  String _selectedNotifications = 'both';

  @override
  void initState() {
    super.initState();
    _usersService = UsersService(widget.apiClient);
    _loadUserProfile();
  }

  Future<void> _loadUserProfile() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final userData = await _storageService.getUserData();
      final userId = userData['userId'];

      if (userId == null) {
        throw Exception('User ID not found');
      }

      final user = await _usersService.getUserProfile(userId);

      setState(() {
        _user = user;
        _firstNameController.text = user.firstName ?? '';
        _lastNameController.text = user.lastName ?? '';
        _phoneController.text = user.phone ?? '';
        _bioController.text = user.bio ?? '';
        _avatarUrl = user.avatar;
        _selectedLanguage = user.language;
        _selectedNotifications = user.notificationPreferences ?? 'both';
        _isLoading = false;
      });
    } catch (e) {
      Logger.logError('Failed to load user profile', error: e);
      setState(() {
        _errorMessage = 'Failed to load profile: ${e.toString()}';
        _isLoading = false;
      });
    }
  }

  Future<void> _pickImage() async {
    try {
      final ImagePicker picker = ImagePicker();
      final XFile? image = await picker.pickImage(
        source: ImageSource.gallery,
        maxWidth: 1000,
        maxHeight: 1000,
        imageQuality: 85,
      );

      if (image != null) {
        setState(() {
          _selectedImage = File(image.path);
        });
      }
    } catch (e) {
      Logger.logError('Failed to pick image', error: e);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to select image: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _saveProfile() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isSaving = true;
      _errorMessage = null;
      _successMessage = null;
    });

    try {
      String? uploadedAvatarUrl = _avatarUrl;

      // Upload avatar if selected
      if (_selectedImage != null && _user != null) {
        try {
          uploadedAvatarUrl = await _usersService.uploadAvatar(_user!.id, _selectedImage!);
        } catch (e) {
          Logger.logError('Avatar upload failed', error: e);
          // Continue with profile update even if avatar upload fails
        }
      }

      // Update profile
      if (_user != null) {
        final updatedUser = await _usersService.updateUserProfile(
          _user!.id,
          firstName: _firstNameController.text,
          lastName: _lastNameController.text,
          phone: _phoneController.text,
          bio: _bioController.text,
          language: _selectedLanguage,
          notificationPreferences: _selectedNotifications,
          avatar: uploadedAvatarUrl,
        );

        // Update stored user data
        await _storageService.saveUserData(
          userId: updatedUser.id,
          username: updatedUser.username,
          email: updatedUser.email,
          role: updatedUser.role,
        );

        // Update locale if language changed
        if (_selectedLanguage != _user!.language) {
          await _storageService.saveLocale(_selectedLanguage);
        }

        setState(() {
          _user = updatedUser;
          _avatarUrl = updatedUser.avatar;
          _selectedImage = null;
          _isEditMode = false;
          _isSaving = false;
          _successMessage = 'Profile updated successfully';
        });

        // Clear success message after 3 seconds
        Future.delayed(const Duration(seconds: 3), () {
          if (mounted) {
            setState(() {
              _successMessage = null;
            });
          }
        });
      }
    } catch (e) {
      Logger.logError('Failed to save profile', error: e);
      setState(() {
        _errorMessage = 'Failed to save profile: ${e.toString()}';
        _isSaving = false;
      });
    }
  }

  void _cancelEdit() {
    setState(() {
      _isEditMode = false;
      _selectedImage = null;
      _firstNameController.text = _user?.firstName ?? '';
      _lastNameController.text = _user?.lastName ?? '';
      _phoneController.text = _user?.phone ?? '';
      _bioController.text = _user?.bio ?? '';
      _selectedLanguage = _user?.language ?? 'en';
      _selectedNotifications = _user?.notificationPreferences ?? 'both';
      _errorMessage = null;
    });
  }

  @override
  void dispose() {
    _firstNameController.dispose();
    _lastNameController.dispose();
    _phoneController.dispose();
    _bioController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Profile'),
        actions: [
          if (!_isEditMode && !_isLoading)
            IconButton(
              onPressed: () => setState(() => _isEditMode = true),
              icon: const Icon(Icons.edit_outlined),
            ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Error message
                    if (_errorMessage != null)
                      Container(
                        padding: const EdgeInsets.all(12),
                        margin: const EdgeInsets.only(bottom: 16),
                        decoration: BoxDecoration(
                          color: AppColors.danger.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: AppColors.danger),
                        ),
                        child: Text(
                          _errorMessage!,
                          style: const TextStyle(color: AppColors.danger),
                        ),
                      ),

                    // Success message
                    if (_successMessage != null)
                      Container(
                        padding: const EdgeInsets.all(12),
                        margin: const EdgeInsets.only(bottom: 16),
                        decoration: BoxDecoration(
                          color: AppColors.success.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: AppColors.success),
                        ),
                        child: Text(
                          _successMessage!,
                          style: const TextStyle(color: AppColors.success),
                        ),
                      ),

                    // Avatar section
                    Center(
                      child: GestureDetector(
                        onTap: _isEditMode ? _pickImage : null,
                        child: Stack(
                          children: [
                            CircleAvatar(
                              radius: 60,
                              backgroundColor: AppColors.primary100,
                              backgroundImage: _selectedImage != null
                                  ? FileImage(_selectedImage!)
                                  : (_avatarUrl != null
                                      ? NetworkImage(_avatarUrl!)
                                      : null) as ImageProvider?,
                              child: _selectedImage == null && _avatarUrl == null
                                  ? const Icon(Icons.person, size: 60, color: AppColors.primary500)
                                  : null,
                            ),
                            if (_isEditMode)
                              Positioned(
                                bottom: 0,
                                right: 0,
                                child: Container(
                                  padding: const EdgeInsets.all(8),
                                  decoration: const BoxDecoration(
                                    color: AppColors.primary500,
                                    shape: BoxShape.circle,
                                  ),
                                  child: const Icon(
                                    Icons.camera_alt,
                                    size: 20,
                                    color: Colors.white,
                                  ),
                                ),
                              ),
                          ],
                        ),
                      ),
                    ),

                    const SizedBox(height: 24),

                    // User info card
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Account Information',
                              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                    fontWeight: FontWeight.bold,
                                  ),
                            ),
                            const SizedBox(height: 16),
                            _buildInfoRow(Icons.person_outline, 'Username', _user?.username ?? ''),
                            const SizedBox(height: 12),
                            _buildInfoRow(Icons.email_outlined, 'Email', _user?.email ?? ''),
                            const SizedBox(height: 12),
                            _buildInfoRow(Icons.badge_outlined, 'Role', _user?.role ?? ''),
                          ],
                        ),
                      ),
                    ),

                    const SizedBox(height: 16),

                    // Personal info card
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Personal Information',
                              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                    fontWeight: FontWeight.bold,
                                  ),
                            ),
                            const SizedBox(height: 16),
                            TextFormField(
                              controller: _firstNameController,
                              enabled: _isEditMode,
                              decoration: const InputDecoration(
                                labelText: 'First Name',
                                prefixIcon: Icon(Icons.person_outline),
                              ),
                            ),
                            const SizedBox(height: 16),
                            TextFormField(
                              controller: _lastNameController,
                              enabled: _isEditMode,
                              decoration: const InputDecoration(
                                labelText: 'Last Name',
                                prefixIcon: Icon(Icons.person_outline),
                              ),
                            ),
                            const SizedBox(height: 16),
                            TextFormField(
                              controller: _phoneController,
                              enabled: _isEditMode,
                              decoration: const InputDecoration(
                                labelText: 'Phone',
                                prefixIcon: Icon(Icons.phone_outlined),
                              ),
                            ),
                            const SizedBox(height: 16),
                            TextFormField(
                              controller: _bioController,
                              enabled: _isEditMode,
                              maxLines: 3,
                              decoration: const InputDecoration(
                                labelText: 'Bio',
                                prefixIcon: Icon(Icons.info_outline),
                                alignLabelWithHint: true,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),

                    const SizedBox(height: 16),

                    // Preferences card
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Preferences',
                              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                    fontWeight: FontWeight.bold,
                                  ),
                            ),
                            const SizedBox(height: 16),
                            DropdownButtonFormField<String>(
                              value: _selectedLanguage,
                              decoration: const InputDecoration(
                                labelText: 'Language',
                                prefixIcon: Icon(Icons.language_outlined),
                              ),
                              items: const [
                                DropdownMenuItem(value: 'en', child: Text('English')),
                                DropdownMenuItem(value: 'ru', child: Text('Русский')),
                              ],
                              onChanged: _isEditMode
                                  ? (value) {
                                      if (value != null) {
                                        setState(() => _selectedLanguage = value);
                                      }
                                    }
                                  : null,
                            ),
                            const SizedBox(height: 16),
                            DropdownButtonFormField<String>(
                              value: _selectedNotifications,
                              decoration: const InputDecoration(
                                labelText: 'Notification Preferences',
                                prefixIcon: Icon(Icons.notifications_outlined),
                              ),
                              items: const [
                                DropdownMenuItem(value: 'tracks', child: Text('Individual Tracks Only')),
                                DropdownMenuItem(value: 'rooms', child: Text('Rooms Only')),
                                DropdownMenuItem(value: 'both', child: Text('Both')),
                              ],
                              onChanged: _isEditMode
                                  ? (value) {
                                      if (value != null) {
                                        setState(() => _selectedNotifications = value);
                                      }
                                    }
                                  : null,
                            ),
                          ],
                        ),
                      ),
                    ),

                    const SizedBox(height: 24),

                    // Action buttons
                    if (_isEditMode)
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton(
                              onPressed: _isSaving ? null : _cancelEdit,
                              child: const Text('Cancel'),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: FilledButton(
                              onPressed: _isSaving ? null : _saveProfile,
                              child: _isSaving
                                  ? const SizedBox(
                                      height: 20,
                                      width: 20,
                                      child: CircularProgressIndicator(strokeWidth: 2),
                                    )
                                  : const Text('Save'),
                            ),
                          ),
                        ],
                      ),
                  ],
                ),
              ),
            ),
    );
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      children: [
        Icon(icon, size: 20, color: AppColors.textSecondary),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: const TextStyle(
                  fontSize: 12,
                  color: AppColors.textTertiary,
                ),
              ),
              const SizedBox(height: 2),
              Text(
                value,
                style: const TextStyle(
                  fontSize: 16,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
