import 'package:flutter/material.dart';
import '../models/task.dart';
import '../services/task_service.dart';
import '../services/auth_service.dart';
import '../widgets/task_card.dart';
import '../l10n/app_localizations.dart';

class MyTasksScreen extends StatefulWidget {
  const MyTasksScreen({Key? key}) : super(key: key);

  @override
  State<MyTasksScreen> createState() => _MyTasksScreenState();
}

class _MyTasksScreenState extends State<MyTasksScreen> {
  late TaskService _taskService;
  List<Task> _tasks = [];
  bool _isLoading = true;
  String? _error;
  String? _selectedStatus;

  final List<String?> _statusFilters = [
    null, // All
    'pending',
    'in_progress',
    'completed',
  ];

  @override
  void initState() {
    super.initState();
    _initializeService();
  }

  Future<void> _initializeService() async {
    final authService = AuthService();
    final baseUrl = 'http://localhost:8081'; // TODO: Get from config

    _taskService = TaskService(
      baseUrl: baseUrl,
      authService: authService,
    );

    await _loadTasks();
  }

  Future<void> _loadTasks() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final tasks = await _taskService.getMyTasks(status: _selectedStatus);
      setState(() {
        _tasks = tasks;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _updateTaskStatus(Task task, TaskStatus newStatus) async {
    try {
      await _taskService.updateTaskStatus(task.id, newStatus);
      await _loadTasks(); // Reload tasks

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Task status updated'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to update task: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  void _showTaskDetails(Task task) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.7,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        expand: false,
        builder: (context, scrollController) => _buildTaskDetailSheet(
          task,
          scrollController,
        ),
      ),
    );
  }

  Widget _buildTaskDetailSheet(Task task, ScrollController scrollController) {
    return Container(
      padding: const EdgeInsets.all(24),
      child: ListView(
        controller: scrollController,
        children: [
          // Handle bar
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: Colors.grey[300],
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
          const SizedBox(height: 20),

          // Title
          Text(
            task.title,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 16),

          // Description
          if (task.description != null && task.description!.isNotEmpty) ...[
            Text(
              'Description',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Text(task.description!),
            const SizedBox(height: 16),
          ],

          // Hint
          if (task.hint != null && task.hint!.isNotEmpty) ...[
            Text(
              'Hint',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.blue.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.blue.withOpacity(0.3)),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.lightbulb_outline, color: Colors.blue[700]),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      task.hint!,
                      style: TextStyle(color: Colors.blue[700]),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
          ],

          // Status update
          Text(
            'Update Status',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              _buildStatusButton(task, TaskStatus.pending, 'Pending', Icons.schedule),
              _buildStatusButton(task, TaskStatus.inProgress, 'In Progress', Icons.play_circle),
              _buildStatusButton(task, TaskStatus.completed, 'Completed', Icons.check_circle),
            ],
          ),
          const SizedBox(height: 24),

          // AI Source (if present)
          if (task.extractedByAi && task.sourceSegment != null) ...[
            Text(
              'AI Extracted From',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.purple.withOpacity(0.05),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.purple.withOpacity(0.2)),
              ),
              child: Text(
                task.sourceSegment!,
                style: TextStyle(
                  color: Colors.purple[900],
                  fontStyle: FontStyle.italic,
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildStatusButton(Task task, TaskStatus status, String label, IconData icon) {
    final isSelected = task.status == status;
    Color color;

    switch (status) {
      case TaskStatus.pending:
        color = Colors.grey;
        break;
      case TaskStatus.inProgress:
        color = Colors.blue;
        break;
      case TaskStatus.completed:
        color = Colors.green;
        break;
      case TaskStatus.cancelled:
        color = Colors.red;
        break;
    }

    return ElevatedButton.icon(
      onPressed: isSelected
          ? null
          : () {
              Navigator.pop(context);
              _updateTaskStatus(task, status);
            },
      icon: Icon(icon, size: 18),
      label: Text(label),
      style: ElevatedButton.styleFrom(
        backgroundColor: isSelected ? color : color.withOpacity(0.1),
        foregroundColor: isSelected ? Colors.white : color,
        side: BorderSide(color: color),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: const Text('My Tasks'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadTasks,
          ),
        ],
      ),
      body: Column(
        children: [
          // Status filter chips
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: _statusFilters.map((status) {
                final isSelected = _selectedStatus == status;
                final label = status == null
                    ? 'All'
                    : status.replaceAll('_', ' ').split(' ').map((word) =>
                        word[0].toUpperCase() + word.substring(1)).join(' ');

                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(label),
                    selected: isSelected,
                    onSelected: (selected) {
                      setState(() {
                        _selectedStatus = selected ? status : null;
                      });
                      _loadTasks();
                    },
                  ),
                );
              }).toList(),
            ),
          ),

          // Task list
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _error != null
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.error_outline,
                                size: 64, color: Colors.red),
                            const SizedBox(height: 16),
                            Text(_error!),
                            const SizedBox(height: 16),
                            ElevatedButton(
                              onPressed: _loadTasks,
                              child: const Text('Retry'),
                            ),
                          ],
                        ),
                      )
                    : _tasks.isEmpty
                        ? Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.task_alt,
                                    size: 64, color: Colors.grey[400]),
                                const SizedBox(height: 16),
                                Text(
                                  'No tasks found',
                                  style: TextStyle(
                                    fontSize: 18,
                                    color: Colors.grey[600],
                                  ),
                                ),
                              ],
                            ),
                          )
                        : RefreshIndicator(
                            onRefresh: _loadTasks,
                            child: ListView.builder(
                              itemCount: _tasks.length,
                              itemBuilder: (context, index) {
                                final task = _tasks[index];
                                return TaskCard(
                                  task: task,
                                  onTap: () => _showTaskDetails(task),
                                );
                              },
                            ),
                          ),
          ),
        ],
      ),
    );
  }
}
