import 'package:flutter/material.dart';
import '../services/api_client.dart';
import '../widgets/error_display.dart';

class DocumentsScreen extends StatefulWidget {
  final ApiClient apiClient;

  const DocumentsScreen({super.key, required this.apiClient});

  @override
  State<DocumentsScreen> createState() => _DocumentsScreenState();
}

class _DocumentsScreenState extends State<DocumentsScreen>
    with AutomaticKeepAliveClientMixin {
  List<Document>? _documents;
  bool _isLoading = true;
  String? _error;
  String _filter = 'all'; // all, audio, video, transcript, other

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    _loadDocuments();
  }

  Future<void> _loadDocuments() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      // TODO: Implement actual API call when backend is ready
      await Future.delayed(const Duration(seconds: 1));

      // Mock documents for now
      final mockDocuments = <Document>[
        Document(
          id: '1',
          name: 'Q3_meeting_recording.mp4',
          type: 'video',
          size: 125600000,
          uploadedAt: DateTime.now().subtract(const Duration(days: 2)),
          uploadedBy: 'John Doe',
        ),
        Document(
          id: '2',
          name: 'team_standup_2025-11-08.mp3',
          type: 'audio',
          size: 15400000,
          uploadedAt: DateTime.now().subtract(const Duration(days: 3)),
          uploadedBy: 'Jane Smith',
        ),
        Document(
          id: '3',
          name: 'meeting_transcript_2025-11-05.txt',
          type: 'transcript',
          size: 45000,
          uploadedAt: DateTime.now().subtract(const Duration(days: 5)),
          uploadedBy: 'System',
        ),
        Document(
          id: '4',
          name: 'project_plan.pdf',
          type: 'other',
          size: 2400000,
          uploadedAt: DateTime.now().subtract(const Duration(days: 7)),
          uploadedBy: 'Admin',
        ),
      ];

      if (mounted) {
        setState(() {
          _documents = mockDocuments;
          _isLoading = false;
        });
      }
    } on Exception catch (e) {
      if (mounted) {
        setState(() {
          _error = 'Error loading documents:\n${e.toString()}\n\nPlease check:\n• Network connection\n• API server is running';
          _isLoading = false;
        });
      }
    } catch (e, stackTrace) {
      if (mounted) {
        setState(() {
          _error = 'Unexpected Error:\n${e.toString()}\n\nStack Trace:\n${stackTrace.toString().split('\n').take(5).join('\n')}';
          _isLoading = false;
        });
      }
    }
  }

  List<Document> _getFilteredDocuments() {
    if (_documents == null) return [];
    if (_filter == 'all') return _documents!;
    return _documents!.where((doc) => doc.type == _filter).toList();
  }

  IconData _getDocumentIcon(String type) {
    switch (type) {
      case 'video':
        return Icons.videocam;
      case 'audio':
        return Icons.audiotrack;
      case 'transcript':
        return Icons.description;
      default:
        return Icons.insert_drive_file;
    }
  }

  Color _getDocumentColor(String type) {
    switch (type) {
      case 'video':
        return Colors.purple;
      case 'audio':
        return Colors.orange;
      case 'transcript':
        return Colors.blue;
      default:
        return Colors.grey;
    }
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final difference = now.difference(date);

    if (difference.inDays == 0) {
      return 'Today, ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
    } else if (difference.inDays == 1) {
      return 'Yesterday';
    } else if (difference.inDays < 7) {
      return '${difference.inDays} days ago';
    } else {
      return '${date.day}.${date.month}.${date.year}';
    }
  }

  void _showDocumentOptions(Document document) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.download),
              title: const Text('Download'),
              onTap: () {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Downloading ${document.name}...'),
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.share),
              title: const Text('Share'),
              onTap: () {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Share ${document.name}'),
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.info_outline),
              title: const Text('Details'),
              onTap: () {
                Navigator.pop(context);
                _showDocumentDetails(document);
              },
            ),
            ListTile(
              leading: Icon(Icons.delete, color: Colors.red.shade700),
              title:
                  Text('Delete', style: TextStyle(color: Colors.red.shade700)),
              onTap: () {
                Navigator.pop(context);
                _confirmDelete(document);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showDocumentDetails(Document document) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Document Details'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildDetailRow('Name', document.name),
            const SizedBox(height: 8),
            _buildDetailRow('Type', document.type.toUpperCase()),
            const SizedBox(height: 8),
            _buildDetailRow('Size', _formatFileSize(document.size)),
            const SizedBox(height: 8),
            _buildDetailRow('Uploaded', _formatDate(document.uploadedAt)),
            const SizedBox(height: 8),
            _buildDetailRow('Uploaded by', document.uploadedBy),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Close'),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 90,
          child: Text(
            '$label:',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
        ),
        Expanded(
          child: Text(value),
        ),
      ],
    );
  }

  void _confirmDelete(Document document) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Document'),
        content: Text('Are you sure you want to delete "${document.name}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text('Deleted ${document.name}'),
                  action: SnackBarAction(
                    label: 'Undo',
                    onPressed: () {},
                  ),
                ),
              );
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    super.build(context); // Required for AutomaticKeepAliveClientMixin

    final filteredDocuments = _getFilteredDocuments();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Documents'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadDocuments,
            tooltip: 'Refresh',
          ),
        ],
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(48),
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
            child: Row(
              children: [
                _buildFilterChip('All', 'all'),
                const SizedBox(width: 8),
                _buildFilterChip('Video', 'video'),
                const SizedBox(width: 8),
                _buildFilterChip('Audio', 'audio'),
                const SizedBox(width: 8),
                _buildFilterChip('Transcripts', 'transcript'),
                const SizedBox(width: 8),
                _buildFilterChip('Other', 'other'),
              ],
            ),
          ),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? FullScreenError(
                  error: _error!,
                  onRetry: _loadDocuments,
                  title: 'Failed to load documents',
                )
              : filteredDocuments.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.folder_open,
                              size: 64, color: Colors.grey.shade400),
                          const SizedBox(height: 16),
                          Text(
                            'No documents found',
                            style: Theme.of(context).textTheme.titleLarge,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _filter == 'all'
                                ? 'Upload your first document'
                                : 'Try changing the filter',
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                        ],
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _loadDocuments,
                      child: ListView.builder(
                        itemCount: filteredDocuments.length,
                        padding: const EdgeInsets.all(8),
                        itemBuilder: (context, index) {
                          final document = filteredDocuments[index];
                          return Card(
                            margin: const EdgeInsets.symmetric(
                              vertical: 4,
                              horizontal: 8,
                            ),
                            child: ListTile(
                              leading: CircleAvatar(
                                backgroundColor:
                                    _getDocumentColor(document.type)
                                        .withValues(alpha: 0.1),
                                child: Icon(
                                  _getDocumentIcon(document.type),
                                  color: _getDocumentColor(document.type),
                                ),
                              ),
                              title: Text(
                                document.name,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                              subtitle: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const SizedBox(height: 4),
                                  Row(
                                    children: [
                                      Icon(Icons.person,
                                          size: 14, color: Colors.grey[600]),
                                      const SizedBox(width: 4),
                                      Text(
                                        document.uploadedBy,
                                        style: TextStyle(
                                            fontSize: 13,
                                            color: Colors.grey[600]),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 2),
                                  Row(
                                    children: [
                                      Icon(Icons.access_time,
                                          size: 14, color: Colors.grey[600]),
                                      const SizedBox(width: 4),
                                      Text(
                                        _formatDate(document.uploadedAt),
                                        style: TextStyle(
                                            fontSize: 13,
                                            color: Colors.grey[600]),
                                      ),
                                      const SizedBox(width: 12),
                                      Icon(Icons.storage,
                                          size: 14, color: Colors.grey[600]),
                                      const SizedBox(width: 4),
                                      Text(
                                        _formatFileSize(document.size),
                                        style: TextStyle(
                                            fontSize: 13,
                                            color: Colors.grey[600]),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                              trailing: IconButton(
                                icon: const Icon(Icons.more_vert),
                                onPressed: () =>
                                    _showDocumentOptions(document),
                              ),
                              isThreeLine: true,
                              onTap: () {
                                // TODO: Open document viewer
                                ScaffoldMessenger.of(context).showSnackBar(
                                  SnackBar(
                                    content: Text('Open: ${document.name}'),
                                  ),
                                );
                              },
                            ),
                          );
                        },
                      ),
                    ),
      floatingActionButton: FloatingActionButton.extended(
        heroTag: 'documents_fab',
        onPressed: () {
          // TODO: Navigate to upload document screen
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Upload document - Coming soon')),
          );
        },
        icon: const Icon(Icons.upload_file),
        label: const Text('Upload'),
      ),
    );
  }

  Widget _buildFilterChip(String label, String value) {
    final isSelected = _filter == value;
    return FilterChip(
      label: Text(label),
      selected: isSelected,
      onSelected: (selected) {
        if (selected) {
          setState(() {
            _filter = value;
          });
        }
      },
      showCheckmark: false,
    );
  }
}

// Mock data model for documents
class Document {
  final String id;
  final String name;
  final String type;
  final int size;
  final DateTime uploadedAt;
  final String uploadedBy;

  Document({
    required this.id,
    required this.name,
    required this.type,
    required this.size,
    required this.uploadedAt,
    required this.uploadedBy,
  });
}
