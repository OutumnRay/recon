// Lightweight user info for task assignments and references
class UserInfo {
  final String id;
  final String username;
  final String email;
  final String firstName;
  final String lastName;
  final String role;

  UserInfo({
    required this.id,
    required this.username,
    required this.email,
    required this.firstName,
    required this.lastName,
    required this.role,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['id'] as String,
      username: json['username'] as String,
      email: json['email'] as String,
      firstName: json['first_name'] as String? ?? '',
      lastName: json['last_name'] as String? ?? '',
      role: json['role'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'email': email,
      'first_name': firstName,
      'last_name': lastName,
      'role': role,
    };
  }

  String get displayName {
    if (firstName.isNotEmpty && lastName.isNotEmpty) {
      return '$firstName $lastName';
    }
    return username;
  }
}

class User {
  final String id;
  final String username;
  final String email;
  final String role;
  final String? departmentId;
  final UserPermissions permissions;
  final String language;
  final String? firstName;
  final String? lastName;
  final String? phone;
  final String? bio;
  final String? avatar;
  final String? notificationPreferences;

  User({
    required this.id,
    required this.username,
    required this.email,
    required this.role,
    this.departmentId,
    required this.permissions,
    required this.language,
    this.firstName,
    this.lastName,
    this.phone,
    this.bio,
    this.avatar,
    this.notificationPreferences,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      username: json['username'] as String,
      email: json['email'] as String,
      role: json['role'] as String,
      departmentId: json['department_id'] as String?,
      permissions: UserPermissions.fromJson(json['permissions'] as Map<String, dynamic>),
      language: json['language'] as String? ?? 'en',
      firstName: json['first_name'] as String?,
      lastName: json['last_name'] as String?,
      phone: json['phone'] as String?,
      bio: json['bio'] as String?,
      avatar: json['avatar'] as String?,
      notificationPreferences: json['notification_preferences'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'email': email,
      'role': role,
      'department_id': departmentId,
      'permissions': permissions.toJson(),
      'language': language,
      'first_name': firstName,
      'last_name': lastName,
      'phone': phone,
      'bio': bio,
      'avatar': avatar,
      'notification_preferences': notificationPreferences,
    };
  }

  String get displayName {
    if (firstName != null && lastName != null && firstName!.isNotEmpty && lastName!.isNotEmpty) {
      return '$firstName $lastName';
    }
    return username;
  }
}

class UserPermissions {
  final bool canScheduleMeetings;
  final bool canManageDepartment;
  final bool canApproveRecordings;

  UserPermissions({
    required this.canScheduleMeetings,
    required this.canManageDepartment,
    required this.canApproveRecordings,
  });

  factory UserPermissions.fromJson(Map<String, dynamic> json) {
    return UserPermissions(
      canScheduleMeetings: json['can_schedule_meetings'] as bool? ?? false,
      canManageDepartment: json['can_manage_department'] as bool? ?? false,
      canApproveRecordings: json['can_approve_recordings'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'can_schedule_meetings': canScheduleMeetings,
      'can_manage_department': canManageDepartment,
      'can_approve_recordings': canApproveRecordings,
    };
  }
}
