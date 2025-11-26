import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../l10n/app_localizations.dart';
import '../services/api_client.dart';
import '../theme/app_colors.dart';
import '../widgets/app_card.dart';
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

  int _countDocumentsByType(String type) {
    if (_documents == null) return 0;
    return _documents!.where((doc) => doc.type == type).length;
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
        return AppColors.secondary500;
      case 'audio':
        return AppColors.warning;
      case 'transcript':
        return AppColors.primary500;
      default:
        return AppColors.textTertiary;
    }
  }

  String _getTypeLabel(AppLocalizations l10n, String type) {
    switch (type) {
      case 'video':
        return l10n.filterVideo;
      case 'audio':
        return l10n.filterAudio;
      case 'transcript':
        return l10n.filterTranscripts;
      default:
        return l10n.filterOther;
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

  String _formatDateTime(BuildContext context, DateTime date) {
    final locale = Localizations.localeOf(context).toString();
    final formatter = DateFormat.yMMMd(locale).add_Hm();
    return formatter.format(date);
  }

  void _showDocumentOptions(Document document) {
    final l10n = AppLocalizations.of(context)!;
    showModalBottomSheet(
      context: context,
      backgroundColor: AppColors.surfaceCard,
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
              leading: const Icon(Icons.download_rounded,
                  color: AppColors.primary600),
              title: Text(l10n.download),
              onTap: () {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(l10n.downloading(document.name)),
                    backgroundColor: AppColors.primary600,
                    behavior: SnackBarBehavior.floating,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.share_rounded,
                  color: AppColors.primary600),
              title: Text(l10n.share),
              onTap: () {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(l10n.shareDocument(document.name)),
                    backgroundColor: AppColors.primary600,
                    behavior: SnackBarBehavior.floating,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.info_outline_rounded,
                  color: AppColors.primary600),
              title: Text(l10n.details),
              onTap: () {
                Navigator.pop(context);
                _showDocumentDetails(document);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete_outline_rounded,
                  color: AppColors.danger),
              title: Text(
                l10n.delete,
                style: const TextStyle(color: AppColors.danger),
              ),
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
        backgroundColor: AppColors.surfaceCard,
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
            _buildDetailRow(
              l10n.uploaded,
              _formatDateTime(context, document.uploadedAt),
            ),
            const SizedBox(height: 8),
            _buildDetailRow(l10n.uploadedBy, document.uploadedBy),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            style: TextButton.styleFrom(
              foregroundColor: AppColors.primary600,
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
        backgroundColor: AppColors.surfaceCard,
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            style: TextButton.styleFrom(
              foregroundColor: AppColors.textSecondary,
            ),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () {
              Navigator.pop(context);
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text(l10n.deleted(document.name)),
                  backgroundColor: AppColors.danger,
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
              backgroundColor: AppColors.danger,
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
    super.build(context);
    final l10n = AppLocalizations.of(context)!;
    final filteredDocuments = _getFilteredDocuments();

    Widget bodyContent;
    if (_isLoading) {
      bodyContent = const Center(
        child: CircularProgressIndicator(color: AppColors.primary500),
      );
    } else if (_error != null) {
      bodyContent = FullScreenError(
        error: _error!,
        onRetry: _loadDocuments,
        title: l10n.failedToLoadDocuments,
      );
    } else {
      bodyContent = RefreshIndicator(
        color: AppColors.primary500,
        onRefresh: _loadDocuments,
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          slivers: [
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(20, 24, 20, 0),
              sliver: SliverList(
                delegate: SliverChildListDelegate(
                  [
                    _buildHeroSection(context, l10n),
                    const SizedBox(height: 20),
                    _buildFiltersCard(context, l10n),
                    const SizedBox(height: 20),
                  ],
                ),
              ),
            ),
            if (filteredDocuments.isEmpty)
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 0, 20, 120),
                sliver: SliverToBoxAdapter(
                  child: _buildEmptyState(context, l10n),
                ),
              )
            else
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 0, 20, 120),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate(
                    (context, index) =>
                        _buildDocumentCard(context, filteredDocuments[index]),
                    childCount: filteredDocuments.length,
                  ),
                ),
              ),
          ],
        ),
      );
    }

    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: Text(l10n.documentsTitle),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded),
            onPressed: _isLoading ? null : _loadDocuments,
            tooltip: l10n.refresh,
          ),
        ],
      ),
      body: SafeArea(child: bodyContent),
      floatingActionButton: FloatingActionButton.extended(
        heroTag: 'documents_fab',
        onPressed: () {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(l10n.uploadDocumentComingSoon),
              backgroundColor: AppColors.primary600,
            ),
          );
        },
        backgroundColor: AppColors.primary600,
        foregroundColor: Colors.white,
        icon: const Icon(Icons.upload_file_rounded),
        label: Text(l10n.upload),
      ),
    );
  }

  Widget _buildHeroSection(BuildContext context, AppLocalizations l10n) {
    final docs = _documents ?? [];
    final totalSize = docs.fold<int>(0, (sum, doc) => sum + doc.size);

    final stats = [
      _buildHeroStat(
        context,
        icon: Icons.folder_open_rounded,
        label: l10n.documentsTitle,
        value: docs.length.toString(),
      ),
      _buildHeroStat(
        context,
        icon: Icons.play_circle_outline,
        label: l10n.filterVideo,
        value: _countDocumentsByType('video').toString(),
      ),
      _buildHeroStat(
        context,
        icon: Icons.library_music_outlined,
        label: l10n.filterAudio,
        value: _countDocumentsByType('audio').toString(),
      ),
      _buildHeroStat(
        context,
        icon: Icons.language_rounded,
        label: l10n.filterTranscripts,
        value: _countDocumentsByType('transcript').toString(),
      ),
      _buildHeroStat(
        context,
        icon: Icons.storage_rounded,
        label: l10n.size,
        value: _formatFileSize(totalSize),
      ),
    ];

    return GradientHeroCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.documentsTitle,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.w700,
                ),
          ),
          const SizedBox(height: 6),
          Text(
            l10n.documentsSubtitle,
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: 20),
          Wrap(
            spacing: 12,
            runSpacing: 12,
            children: stats,
          ),
        ],
      ),
    );
  }

  Widget _buildFiltersCard(BuildContext context, AppLocalizations l10n) {
    return SurfaceCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSectionHeader(
            context,
            icon: Icons.tune_rounded,
            title: l10n.filtersTitle,
            subtitle: l10n.tryChangingFilter,
          ),
          const SizedBox(height: 16),
          Wrap(
            spacing: 12,
            runSpacing: 12,
            children: [
              _buildFilterChip(l10n.filterAll, 'all'),
              _buildFilterChip(l10n.filterVideo, 'video'),
              _buildFilterChip(l10n.filterAudio, 'audio'),
              _buildFilterChip(l10n.filterTranscripts, 'transcript'),
              _buildFilterChip(l10n.filterOther, 'other'),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDocumentCard(BuildContext context, Document document) {
    final l10n = AppLocalizations.of(context)!;
    final typeColor = _getDocumentColor(document.type);
    final typeLabel = _getTypeLabel(l10n, document.type);
    final dateLabel = _formatDateTime(context, document.uploadedAt);
    final sizeLabel = _formatFileSize(document.size);

    return SurfaceCard(
      margin: const EdgeInsets.only(bottom: 16),
      padding: EdgeInsets.zero,
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(28),
        child: InkWell(
          borderRadius: BorderRadius.circular(28),
          onTap: () {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(l10n.openDocument(document.name)),
                backgroundColor: AppColors.primary600,
              ),
            );
          },
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 52,
                      height: 52,
                      decoration: BoxDecoration(
                        color: typeColor.withValues(alpha: 0.12),
                        borderRadius: BorderRadius.circular(18),
                      ),
                      child: Icon(
                        _getDocumentIcon(document.type),
                        color: typeColor,
                        size: 26,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            document.name,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: Theme.of(context)
                                .textTheme
                                .titleMedium
                                ?.copyWith(
                                  color: AppColors.textPrimary,
                                  fontWeight: FontWeight.w600,
                                ),
                          ),
                          const SizedBox(height: 6),
                          Text(
                            '${l10n.uploadedBy}: ${document.uploadedBy}',
                            style: Theme.of(context)
                                .textTheme
                                .bodyMedium
                                ?.copyWith(color: AppColors.textSecondary),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      onPressed: () => _showDocumentOptions(document),
                      icon: const Icon(Icons.more_horiz_rounded),
                      color: AppColors.textSecondary,
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    _buildMetaChip(
                      icon: Icons.category_outlined,
                      label: typeLabel,
                    ),
                    _buildMetaChip(
                      icon: Icons.storage_rounded,
                      label: sizeLabel,
                    ),
                    _buildMetaChip(
                      icon: Icons.schedule_rounded,
                      label: dateLabel,
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context, AppLocalizations l10n) {
    final subtitle =
        _filter == 'all' ? l10n.uploadFirstDocument : l10n.tryChangingFilter;
    return SurfaceCard(
      padding: const EdgeInsets.symmetric(vertical: 32, horizontal: 20),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.folder_open_rounded,
              size: 56, color: AppColors.textTertiary),
          const SizedBox(height: 16),
          Text(
            l10n.noDocumentsFound,
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          Text(
            subtitle,
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: AppColors.textSecondary),
            textAlign: TextAlign.center,
          ),
        ],
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
      selectedColor: AppColors.primary50,
      side: BorderSide(
        color: isSelected ? AppColors.primary300 : AppColors.border,
        width: 1.5,
      ),
      labelStyle: TextStyle(
        color: isSelected ? AppColors.primary700 : AppColors.textSecondary,
        fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
      ),
    );
  }

  Widget _buildHeroStat(
    BuildContext context, {
    required IconData icon,
    required String label,
    required String value,
  }) {
    final textTheme = Theme.of(context).textTheme;
    return ConstrainedBox(
      constraints: const BoxConstraints(minWidth: 140),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.border),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, size: 20, color: AppColors.primary600),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    label,
                    style: textTheme.labelMedium?.copyWith(
                      color: AppColors.textSecondary,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    value,
                    style: textTheme.titleMedium?.copyWith(
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionHeader(
    BuildContext context, {
    required IconData icon,
    required String title,
    String? subtitle,
  }) {
    final textTheme = Theme.of(context).textTheme;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
            color: AppColors.primary50,
            borderRadius: BorderRadius.circular(16),
          ),
          child: Icon(icon, color: AppColors.primary600, size: 22),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
              ),
              if (subtitle != null) ...[
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: textTheme.bodySmall?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildMetaChip({required IconData icon, required String label}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.border),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: AppColors.textSecondary),
          const SizedBox(width: 6),
          Text(
            label,
            style: const TextStyle(
              color: AppColors.textSecondary,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
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
