import 'package:flutter/material.dart';
import '../l10n/app_localizations.dart';
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
        return const Color(0xFF9C27B0); // Purple
      case 'audio':
        return const Color(0xFFFF9800); // Orange
      case 'transcript':
        return const Color(0xFF26C6DA); // Cyan
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
    final l10n = AppLocalizations.of(context)!;
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 40,
              height: 4,
              margin: const EdgeInsets.symmetric(vertical: 12),
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.download, color: Color(0xFF26C6DA)),
              title: Text(l10n.download),
              onTap: () {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(l10n.downloading(document.name)),
                    behavior: SnackBarBehavior.floating,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.share, color: Color(0xFF26C6DA)),
              title: Text(l10n.share),
              onTap: () {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Share ${document.name}'),
                    behavior: SnackBarBehavior.floating,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.info_outline, color: Color(0xFF26C6DA)),
              title: Text(l10n.details),
              onTap: () {
                Navigator.pop(context);
                _showDocumentDetails(document);
              },
            ),
            ListTile(
              leading: Icon(Icons.delete, color: Colors.red.shade700),
              title:
                  Text(l10n.delete, style: TextStyle(color: Colors.red.shade700)),
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
    final l10n = AppLocalizations.of(context)!;
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
        ),
        title: Text(l10n.documentDetails),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildDetailRow(l10n.name, document.name),
            const SizedBox(height: 8),
            _buildDetailRow(l10n.type, document.type.toUpperCase()),
            const SizedBox(height: 8),
            _buildDetailRow(l10n.size, _formatFileSize(document.size)),
            const SizedBox(height: 8),
            _buildDetailRow(l10n.uploaded, _formatDate(document.uploadedAt)),
            const SizedBox(height: 8),
            _buildDetailRow(l10n.uploadedBy, document.uploadedBy),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            style: TextButton.styleFrom(
              foregroundColor: const Color(0xFF26C6DA),
            ),
            child: Text(l10n.close),
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
    final l10n = AppLocalizations.of(context)!;
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
        ),
        title: Text(l10n.deleteDocument),
        content: Text(l10n.deleteDocumentConfirm(document.name)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            style: TextButton.styleFrom(
              foregroundColor: Colors.grey.shade700,
            ),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () {
              Navigator.pop(context);
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text(l10n.deleted(document.name)),
                  behavior: SnackBarBehavior.floating,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  action: SnackBarAction(
                    label: l10n.undo,
                    textColor: Colors.white,
                    onPressed: () {},
                  ),
                ),
              );
            },
            style: FilledButton.styleFrom(
              backgroundColor: Colors.red,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: Text(l10n.delete),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    super.build(context); // Required for AutomaticKeepAliveClientMixin
    final l10n = AppLocalizations.of(context)!;

    final filteredDocuments = _getFilteredDocuments();

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: Text(l10n.documentsTitle),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF26C6DA),
        elevation: 1,
        shadowColor: Colors.black.withValues(alpha: 0.1),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadDocuments,
            tooltip: l10n.refresh,
          ),
        ],
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(48),
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
            child: Row(
              children: [
                _buildFilterChip(l10n.filterAll, 'all'),
                const SizedBox(width: 8),
                _buildFilterChip(l10n.filterVideo, 'video'),
                const SizedBox(width: 8),
                _buildFilterChip(l10n.filterAudio, 'audio'),
                const SizedBox(width: 8),
                _buildFilterChip(l10n.filterTranscripts, 'transcript'),
                const SizedBox(width: 8),
                _buildFilterChip(l10n.filterOther, 'other'),
              ],
            ),
          ),
        ),
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(
                color: Color(0xFF26C6DA),
              ),
            )
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
                            l10n.noDocumentsFound,
                            style: Theme.of(context).textTheme.titleLarge,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _filter == 'all'
                                ? l10n.uploadFirstDocument
                                : l10n.tryChangingFilter,
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
                              vertical: 8,
                              horizontal: 12,
                            ),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(20),
                            ),
                            elevation: 2,
                            child: ListTile(
                              contentPadding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 8,
                              ),
                              leading: CircleAvatar(
                                radius: 28,
                                backgroundColor:
                                    _getDocumentColor(document.type)
                                        .withValues(alpha: 0.15),
                                child: Icon(
                                  _getDocumentIcon(document.type),
                                  color: _getDocumentColor(document.type),
                                  size: 26,
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
            SnackBar(content: Text(l10n.uploadDocumentComingSoon)),
          );
        },
        backgroundColor: const Color(0xFF26C6DA),
        foregroundColor: Colors.white,
        elevation: 4,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
        ),
        icon: const Icon(Icons.upload_file),
        label: Text(l10n.upload),
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
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      backgroundColor: Colors.white,
      selectedColor: const Color(0xFF26C6DA).withValues(alpha: 0.15),
      side: BorderSide(
        color: isSelected ? const Color(0xFF26C6DA) : Colors.grey.shade300,
        width: 1.5,
      ),
      labelStyle: TextStyle(
        color: isSelected ? const Color(0xFF00ACC1) : Colors.grey.shade700,
        fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
      ),
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
