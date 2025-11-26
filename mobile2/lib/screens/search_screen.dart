import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../l10n/app_localizations.dart';
import '../services/api_client.dart';
import '../theme/app_colors.dart';
import '../widgets/app_card.dart';
import '../widgets/error_display.dart';

class SearchScreen extends StatefulWidget {
  final ApiClient apiClient;

  const SearchScreen({super.key, required this.apiClient});

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen>
    with AutomaticKeepAliveClientMixin {
  final _searchController = TextEditingController();
  bool _isSearching = false;
  List<SearchResult>? _results;
  String? _error;

  @override
  bool get wantKeepAlive => true;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _performSearch() async {
    final query = _searchController.text.trim();
    if (query.isEmpty) {
      return;
    }

    setState(() {
      _isSearching = true;
      _error = null;
    });

    try {
      // TODO: Implement actual API call when backend is ready
      await Future.delayed(const Duration(seconds: 1));

      // Mock results for now
      final mockResults = <SearchResult>[
        SearchResult(
          id: '1',
          type: 'meeting',
          title: 'Team standup',
          snippet: 'Discussion about project timeline and milestones...',
          date: DateTime.now().subtract(const Duration(days: 2)),
          relevance: 0.95,
        ),
        SearchResult(
          id: '2',
          type: 'transcript',
          title: 'Q3 Planning Meeting',
          snippet: 'Budget allocation for the next quarter was discussed...',
          date: DateTime.now().subtract(const Duration(days: 5)),
          relevance: 0.87,
        ),
      ];

      if (mounted) {
        setState(() {
          _results = mockResults;
          _isSearching = false;
        });
      }
    } on Exception catch (e) {
      if (mounted) {
        setState(() {
          _error = 'Search Error:\n${e.toString()}\n\nPlease check:\n• Network connection\n• Search query format\n• API server is running';
          _isSearching = false;
        });
      }
    } catch (e, stackTrace) {
      if (mounted) {
        setState(() {
          _error = 'Unexpected Error:\n${e.toString()}\n\nStack Trace:\n${stackTrace.toString().split('\n').take(5).join('\n')}';
          _isSearching = false;
        });
      }
    }
  }

  IconData _getResultIcon(String type) {
    switch (type) {
      case 'meeting':
        return Icons.event;
      case 'transcript':
        return Icons.description;
      case 'document':
        return Icons.insert_drive_file;
      default:
        return Icons.search;
    }
  }

  Color _getResultColor(String type) {
    switch (type) {
      case 'meeting':
        return AppColors.primary600;
      case 'transcript':
        return AppColors.success;
      case 'document':
        return AppColors.warning;
      default:
        return AppColors.textTertiary;
    }
  }

  String _getResultTypeLabel(String type) {
    final l10n = AppLocalizations.of(context)!;
    switch (type) {
      case 'meeting':
        return l10n.meeting;
      case 'transcript':
        return l10n.transcript;
      case 'document':
        return l10n.document;
      default:
        return type;
    }
  }

  @override
  Widget build(BuildContext context) {
    super.build(context);
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: Text(l10n.searchTitle),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
      ),
      body: SafeArea(
        child: Column(
          children: [
            Flexible(
              flex: 0,
              child: SingleChildScrollView(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(20, 24, 20, 0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildHeroSection(context, l10n),
                      const SizedBox(height: 20),
                      _buildSearchPanel(context, l10n),
                    ],
                  ),
                ),
              ),
            ),
            const SizedBox(height: 20),
            Expanded(child: _buildResultsArea(context, l10n)),
          ],
        ),
      ),
    );
  }

  Widget _buildHeroSection(BuildContext context, AppLocalizations l10n) {
    final total = _results?.length ?? 0;
    final meetingsCount =
        _results?.where((r) => r.type == 'meeting').length ?? 0;
    final transcriptsCount =
        _results?.where((r) => r.type == 'transcript').length ?? 0;
    final documentsCount =
        _results?.where((r) => r.type == 'document').length ?? 0;

    final stats = [
      _buildHeroStat(
        context,
        icon: Icons.search_rounded,
        label: l10n.searchResultsLabel,
        value: total.toString(),
      ),
      _buildHeroStat(
        context,
        icon: Icons.event_note_rounded,
        label: l10n.meetings,
        value: meetingsCount.toString(),
      ),
      _buildHeroStat(
        context,
        icon: Icons.translate_rounded,
        label: l10n.transcript,
        value: transcriptsCount.toString(),
      ),
      _buildHeroStat(
        context,
        icon: Icons.folder_copy_rounded,
        label: l10n.documentsTitle,
        value: documentsCount.toString(),
      ),
    ];

    return GradientHeroCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.semanticSearch,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.w700,
                ),
          ),
          const SizedBox(height: 6),
          Text(
            l10n.searchDescription,
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

  Widget _buildSearchPanel(BuildContext context, AppLocalizations l10n) {
    final suggestions = [
      l10n.budgetDiscussion,
      l10n.projectDeadlines,
      l10n.whoTalkedAbout,
    ];

    return SurfaceCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _searchController,
                  decoration: InputDecoration(
                    hintText: l10n.searchHint,
                    prefixIcon: const Icon(
                      Icons.search_rounded,
                      color: AppColors.primary600,
                    ),
                    suffixIcon: _searchController.text.isNotEmpty
                        ? IconButton(
                            icon: const Icon(Icons.clear_rounded),
                            onPressed: () {
                              _searchController.clear();
                              setState(() {
                                _results = null;
                                _error = null;
                              });
                            },
                          )
                        : null,
                    filled: true,
                    fillColor: AppColors.surfaceMuted,
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(18),
                      borderSide: BorderSide(color: AppColors.border),
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(18),
                      borderSide: BorderSide(color: AppColors.border),
                    ),
                    focusedBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(18),
                      borderSide: const BorderSide(
                        color: AppColors.primary500,
                        width: 2,
                      ),
                    ),
                  ),
                  onSubmitted: (_) => _performSearch(),
                  onChanged: (_) => setState(() {}),
                ),
              ),
              const SizedBox(width: 12),
              FilledButton(
                onPressed: _isSearching ? null : _performSearch,
                style: FilledButton.styleFrom(
                  backgroundColor: AppColors.primary600,
                  foregroundColor: Colors.white,
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(18),
                  ),
                ),
                child: _isSearching
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : Text(l10n.searchButton),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(
            l10n.trySearching,
            style: Theme.of(context).textTheme.labelLarge?.copyWith(
                  color: AppColors.textSecondary,
                ),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: suggestions.map(_buildExampleChip).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildResultsArea(BuildContext context, AppLocalizations l10n) {
    if (_error != null) {
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20),
        child: FullScreenError(
          error: _error!,
          onRetry: _performSearch,
          title: l10n.searchFailed,
        ),
      );
    }

    if (_results == null) {
      return ListView(
        physics: const BouncingScrollPhysics(),
        padding: const EdgeInsets.fromLTRB(20, 0, 20, 40),
        children: [
          _buildPlaceholderCard(
            context,
            icon: Icons.search_rounded,
            title: l10n.semanticSearch,
            description: l10n.searchDescription,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 16),
                Text(
                  l10n.trySearching,
                  style: Theme.of(context).textTheme.labelLarge?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                ),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    _buildExampleChip(l10n.budgetDiscussion),
                    _buildExampleChip(l10n.projectDeadlines),
                    _buildExampleChip(l10n.whoTalkedAbout),
                  ],
                ),
              ],
            ),
          ),
        ],
      );
    }

    if (_results!.isEmpty) {
      return ListView(
        physics: const BouncingScrollPhysics(),
        padding: const EdgeInsets.fromLTRB(20, 0, 20, 40),
        children: [
          _buildPlaceholderCard(
            context,
            icon: Icons.search_off_rounded,
            title: l10n.noResultsFound,
            description: l10n.tryDifferentKeywords,
          ),
        ],
      );
    }

    return ListView.builder(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(20, 0, 20, 40),
      itemCount: _results!.length,
      itemBuilder: (context, index) =>
          _buildResultCard(context, _results![index]),
    );
  }

  Widget _buildResultCard(BuildContext context, SearchResult result) {
    final l10n = AppLocalizations.of(context)!;
    final color = _getResultColor(result.type);
    final matchLabel =
        '${(result.relevance * 100).toInt()}% ${l10n.match.toLowerCase()}';

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
                backgroundColor: AppColors.primary600,
                content: Text(l10n.openSearchResult(result.type, result.title)),
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
                        color: color.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(18),
                      ),
                      child: Icon(
                        _getResultIcon(result.type),
                        color: color,
                        size: 26,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            result.title,
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
                            result.snippet,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: Theme.of(context)
                                .textTheme
                                .bodyMedium
                                ?.copyWith(color: AppColors.textSecondary),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 12),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          matchLabel,
                          style: Theme.of(context)
                              .textTheme
                              .labelLarge
                              ?.copyWith(color: AppColors.textSecondary),
                        ),
                      ],
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
                      label: _getResultTypeLabel(result.type),
                    ),
                    _buildMetaChip(
                      icon: Icons.schedule_rounded,
                      label: _formatResultDate(context, result.date),
                    ),
                    _buildMetaChip(
                      icon: Icons.percent_rounded,
                      label: matchLabel,
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

  Widget _buildPlaceholderCard(
    BuildContext context, {
    required IconData icon,
    required String title,
    required String description,
    Widget? child,
  }) {
    return SurfaceCard(
      padding: const EdgeInsets.symmetric(vertical: 32, horizontal: 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          Icon(icon, size: 56, color: AppColors.textTertiary),
          const SizedBox(height: 16),
          Text(
            title,
            style: Theme.of(context).textTheme.titleMedium,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 8),
          Text(
            description,
            textAlign: TextAlign.center,
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: AppColors.textSecondary),
          ),
          if (child != null) child,
        ],
      ),
    );
  }

  Widget _buildExampleChip(String text) {
    return InkWell(
      onTap: _isSearching
          ? null
          : () {
              _searchController.text = text;
              _performSearch();
            },
      borderRadius: BorderRadius.circular(18),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: AppColors.primary50,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: AppColors.primary200),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(
              Icons.lightbulb_outline,
              size: 16,
              color: AppColors.primary600,
            ),
            const SizedBox(width: 6),
            Text(
              text,
              style: const TextStyle(
                fontSize: 13,
                color: AppColors.primary700,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
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
      constraints: const BoxConstraints(minWidth: 120),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.border),
        ),
        child: Row(
          children: [
            Icon(icon, size: 18, color: AppColors.primary600),
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
          Icon(icon, size: 14, color: AppColors.textSecondary),
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

  String _formatResultDate(BuildContext context, DateTime date) {
    final l10n = AppLocalizations.of(context)!;
    final now = DateTime.now();
    final difference = now.difference(date).inDays;

    if (difference == 0) return l10n.today;
    if (difference == 1) return l10n.yesterday;
    if (difference > 1 && difference < 7) return l10n.daysAgo(difference);

    final locale = Localizations.localeOf(context).toString();
    return DateFormat.yMMMd(locale).format(date);
  }
}

// Mock data model for search results
class SearchResult {
  final String id;
  final String type;
  final String title;
  final String snippet;
  final DateTime date;
  final double relevance;

  SearchResult({
    required this.id,
    required this.type,
    required this.title,
    required this.snippet,
    required this.date,
    required this.relevance,
  });
}
