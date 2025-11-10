class User {
  final String id;
  final String username;
  final String email;
  final String role;
  final String? departmentId;
  final UserPermissions permissions;
  final String language;

  User({
    required this.id,
    required this.username,
    required this.email,
    required this.role,
    this.departmentId,
    required this.permissions,
    required this.language,
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
    };
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
