import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../models/task.dart';
import '../l10n/app_localizations.dart';

class TaskCard extends Widget {
  final Task task;
  final VoidCallback? onTap;
  final Function(TaskStatus)? onStatusChange;

  const TaskCard({
    Key? key,
    required this.task,
    this.onTap,
    this.onStatusChange,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Title and priority badge
              Row(
                children: [
                  Expanded(
                    child: Text(
                      task.title,
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                            decoration: task.status == TaskStatus.completed
                                ? TextDecoration.lineThrough
                                : null,
                          ),
                    ),
                  ),
                  _buildPriorityBadge(context),
                ],
              ),

              const SizedBox(height: 8),

              // Description
              if (task.description != null && task.description!.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Text(
                    task.description!,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Colors.grey[600],
                        ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),

              // Hint (if present)
              if (task.hint != null && task.hint!.isNotEmpty)
                Container(
                  padding: const EdgeInsets.all(8),
                  margin: const EdgeInsets.only(bottom: 8),
                  decoration: BoxDecoration(
                    color: Colors.blue.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.blue.withOpacity(0.3)),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.lightbulb_outline,
                          size: 16,
                          color: Colors.blue[700]),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          task.hint!,
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.blue[700],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),

              // Status and assignment info
              Row(
                children: [
                  _buildStatusChip(context),
                  const SizedBox(width: 8),

                  // Assigned to
                  if (task.assignedToUser != null)
                    Expanded(
                      child: Row(
                        children: [
                          Icon(Icons.person_outline, size: 14, color: Colors.grey[600]),
                          const SizedBox(width: 4),
                          Expanded(
                            child: Text(
                              _getUserDisplayName(task.assignedToUser!),
                              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),
                    ),
                ],
              ),

              // Due date
              if (task.dueDate != null)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Row(
                    children: [
                      Icon(
                        Icons.calendar_today_outlined,
                        size: 14,
                        color: _isDueSoon(task.dueDate!)
                            ? Colors.red
                            : Colors.grey[600],
                      ),
                      const SizedBox(width: 4),
                      Text(
                        _formatDueDate(task.dueDate!, localizations),
                        style: TextStyle(
                          fontSize: 12,
                          color: _isDueSoon(task.dueDate!)
                              ? Colors.red
                              : Colors.grey[600],
                          fontWeight: _isDueSoon(task.dueDate!)
                              ? FontWeight.bold
                              : FontWeight.normal,
                        ),
                      ),
                    ],
                  ),
                ),

              // AI badge
              if (task.extractedByAi)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Row(
                    children: [
                      Icon(Icons.auto_awesome, size: 14, color: Colors.purple[400]),
                      const SizedBox(width: 4),
                      Text(
                        'AI',
                        style: TextStyle(fontSize: 11, color: Colors.purple[400]),
                      ),
                      if (task.aiConfidence != null)
                        Text(
                          ' (${(task.aiConfidence! * 100).toStringAsFixed(0)}%)',
                          style: TextStyle(fontSize: 11, color: Colors.grey[500]),
                        ),
                    ],
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildPriorityBadge(BuildContext context) {
    Color badgeColor;
    IconData icon;

    switch (task.priority) {
      case TaskPriority.urgent:
        badgeColor = Colors.red;
        icon = Icons.priority_high;
        break;
      case TaskPriority.high:
        badgeColor = Colors.orange;
        icon = Icons.arrow_upward;
        break;
      case TaskPriority.medium:
        badgeColor = Colors.blue;
        icon = Icons.remove;
        break;
      case TaskPriority.low:
        badgeColor = Colors.grey;
        icon = Icons.arrow_downward;
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: badgeColor.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: badgeColor.withOpacity(0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: badgeColor),
        ],
      ),
    );
  }

  Widget _buildStatusChip(BuildContext context) {
    Color chipColor;
    IconData icon;
    String label;

    switch (task.status) {
      case TaskStatus.pending:
        chipColor = Colors.grey;
        icon = Icons.schedule;
        label = 'Pending';
        break;
      case TaskStatus.inProgress:
        chipColor = Colors.blue;
        icon = Icons.play_circle_outline;
        label = 'In Progress';
        break;
      case TaskStatus.completed:
        chipColor = Colors.green;
        icon = Icons.check_circle_outline;
        label = 'Completed';
        break;
      case TaskStatus.cancelled:
        chipColor = Colors.red;
        icon = Icons.cancel_outlined;
        label = 'Cancelled';
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: chipColor.withOpacity(0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: chipColor.withOpacity(0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: chipColor),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 12,
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
