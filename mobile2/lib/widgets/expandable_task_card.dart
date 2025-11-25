import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../models/task.dart';
import '../l10n/app_localizations.dart';
import '../theme/app_colors.dart';

class ExpandableTaskCard extends StatefulWidget {
  final Task task;
  final Function(bool)? onStatusToggle;

  const ExpandableTaskCard({
    super.key,
    required this.task,
    this.onStatusToggle,
  });

  @override
  State<ExpandableTaskCard> createState() => _ExpandableTaskCardState();
}

class _ExpandableTaskCardState extends State<ExpandableTaskCard> {
  bool _isExpanded = false;

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;
    final isCompleted = widget.task.status == TaskStatus.completed;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: isCompleted ? AppColors.success.withValues(alpha: 0.3) : AppColors.border,
        ),
      ),
      child: Column(
        children: [
          InkWell(
            onTap: () => setState(() => _isExpanded = !_isExpanded),
            borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  // Checkbox
                  Checkbox(
                    value: isCompleted,
                    onChanged: widget.onStatusToggle != null
                        ? (value) => widget.onStatusToggle!(value ?? false)
                        : null,
                    activeColor: AppColors.success,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(4),
                    ),
                  ),
                  const SizedBox(width: 12),

                  // Title and priority
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          widget.task.title,
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                                decoration: isCompleted
                                    ? TextDecoration.lineThrough
                                    : null,
                                color: isCompleted
                                    ? AppColors.textSecondary
                                    : AppColors.textPrimary,
                              ),
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            _buildPriorityBadge(context),
                            const SizedBox(width: 8),
                            _buildStatusChip(context),
                          ],
                        ),
                      ],
                    ),
                  ),

                  // Expand icon
                  Icon(
                    _isExpanded
                        ? Icons.keyboard_arrow_up_rounded
                        : Icons.keyboard_arrow_down_rounded,
                    color: AppColors.textSecondary,
                  ),
                ],
              ),
            ),
          ),

          // Expanded content
          if (_isExpanded)
            Container(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              decoration: BoxDecoration(
                color: AppColors.surfaceMuted.withValues(alpha: 0.3),
                borderRadius: const BorderRadius.vertical(
                  bottom: Radius.circular(12),
                ),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Divider(),
                  const SizedBox(height: 8),

                  // Description
                  if (widget.task.description != null && widget.task.description!.isNotEmpty) ...[
                    Text(
                      'Description',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      widget.task.description!,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: AppColors.textSecondary,
                          ),
                    ),
                    const SizedBox(height: 12),
                  ],

                  // Hint
                  if (widget.task.hint != null && widget.task.hint!.isNotEmpty) ...[
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.blue.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.blue.withValues(alpha: 0.3)),
                      ),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(Icons.lightbulb_outline, color: Colors.blue[700], size: 20),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Hint',
                                  style: TextStyle(
                                    fontSize: 12,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.blue[700],
                                  ),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  widget.task.hint!,
                                  style: TextStyle(
                                    fontSize: 13,
                                    color: Colors.blue[700],
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),
                  ],

                  // Assigned to
                  if (widget.task.assignedToUser != null) ...[
                    Row(
                      children: [
                        Icon(Icons.person_outline, size: 18, color: AppColors.textSecondary),
                        const SizedBox(width: 8),
                        Text(
                          'Assigned to:',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: AppColors.textSecondary,
                                fontWeight: FontWeight.w600,
                              ),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            _getUserDisplayName(widget.task.assignedToUser!),
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                  ],

                  // Due date
                  if (widget.task.dueDate != null) ...[
                    Row(
                      children: [
                        Icon(
                          Icons.calendar_today_outlined,
                          size: 18,
                          color: _isDueSoon(widget.task.dueDate!)
                              ? AppColors.danger
                              : AppColors.textSecondary,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Due:',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: AppColors.textSecondary,
                                fontWeight: FontWeight.w600,
                              ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          _formatDueDate(widget.task.dueDate!, localizations),
                          style: TextStyle(
                            fontSize: 14,
                            color: _isDueSoon(widget.task.dueDate!)
                                ? AppColors.danger
                                : AppColors.textPrimary,
                            fontWeight: _isDueSoon(widget.task.dueDate!)
                                ? FontWeight.bold
                                : FontWeight.normal,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                  ],

                  // AI badge
                  if (widget.task.extractedByAi) ...[
                    Row(
                      children: [
                        Icon(Icons.auto_awesome, size: 18, color: Colors.purple[400]),
                        const SizedBox(width: 8),
                        Text(
                          'AI Extracted',
                          style: TextStyle(fontSize: 12, color: Colors.purple[400]),
                        ),
                        if (widget.task.aiConfidence != null)
                          Text(
                            ' (${(widget.task.aiConfidence! * 100).toStringAsFixed(0)}% confidence)',
                            style: TextStyle(fontSize: 12, color: Colors.grey[500]),
                          ),
                      ],
                    ),
                    if (widget.task.sourceSegment != null && widget.task.sourceSegment!.isNotEmpty) ...[
                      const SizedBox(height: 8),
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.purple.withValues(alpha: 0.05),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: Colors.purple.withValues(alpha: 0.2)),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Source:',
                              style: TextStyle(
                                fontSize: 11,
                                fontWeight: FontWeight.bold,
                                color: Colors.purple[900],
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              widget.task.sourceSegment!,
                              style: TextStyle(
                                fontSize: 12,
                                color: Colors.purple[900],
                                fontStyle: FontStyle.italic,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ],
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildPriorityBadge(BuildContext context) {
    Color badgeColor;
    IconData icon;
    String label;

    switch (widget.task.priority) {
      case TaskPriority.urgent:
        badgeColor = AppColors.danger;
        icon = Icons.priority_high;
        label = 'Urgent';
        break;
      case TaskPriority.high:
        badgeColor = Colors.orange;
        icon = Icons.arrow_upward;
        label = 'High';
        break;
      case TaskPriority.medium:
        badgeColor = AppColors.primary500;
        icon = Icons.remove;
        label = 'Medium';
        break;
      case TaskPriority.low:
        badgeColor = AppColors.textTertiary;
        icon = Icons.arrow_downward;
        label = 'Low';
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: badgeColor.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: badgeColor.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: badgeColor),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              color: badgeColor,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatusChip(BuildContext context) {
    Color chipColor;
    IconData icon;
    String label;

    switch (widget.task.status) {
      case TaskStatus.pending:
        chipColor = AppColors.textTertiary;
        icon = Icons.schedule;
        label = 'Pending';
        break;
      case TaskStatus.inProgress:
        chipColor = AppColors.warning;
        icon = Icons.play_circle_outline;
        label = 'In Progress';
        break;
      case TaskStatus.completed:
        chipColor = AppColors.success;
        icon = Icons.check_circle;
        label = 'Completed';
        break;
      case TaskStatus.cancelled:
        chipColor = AppColors.danger;
        icon = Icons.cancel;
        label = 'Cancelled';
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: chipColor.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: chipColor.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: chipColor),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              color: chipColor,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  String _getUserDisplayName(dynamic user) {
    if (user.firstName != null && user.firstName!.isNotEmpty) {
      return '${user.firstName} ${user.lastName ?? ''}'.trim();
    }
    return user.username ?? user.email ?? 'Unknown';
  }

  bool _isDueSoon(DateTime dueDate) {
    final now = DateTime.now();
    final diff = dueDate.difference(now);
    return diff.inDays <= 2 && diff.inDays >= 0;
  }

  String _formatDueDate(DateTime dueDate, AppLocalizations localizations) {
    final now = DateTime.now();
    final diff = dueDate.difference(now);

    if (diff.inDays < 0) {
      return 'Overdue: ${DateFormat.yMMMd().format(dueDate)}';
    } else if (diff.inDays == 0) {
      return 'Due today';
    } else if (diff.inDays == 1) {
      return 'Due tomorrow';
    } else if (diff.inDays <= 7) {
      return 'Due in ${diff.inDays} days';
    } else {
      return DateFormat.yMMMd().format(dueDate);
    }
  }
}
