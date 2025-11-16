import 'package:flutter/material.dart';
import '../l10n/app_localizations.dart';
import '../services/api_client.dart';
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
        return const Color(0xFF26C6DA); // Cyan
      case 'transcript':
        return const Color(0xFF66BB6A); // Green
      case 'document':
        return const Color(0xFFFF9800); // Orange
      default:
        return Colors.grey;
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
    super.build(context); // Required for AutomaticKeepAliveClientMixin
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        title: Text(l10n.searchTitle),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF26C6DA),
        elevation: 1,
        shadowColor: Colors.black.withValues(alpha: 0.1),
      ),
      body: Column(
        children: [
          // Search bar
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surface,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _searchController,
                    decoration: InputDecoration(
                      hintText: l10n.searchHint,
                      prefixIcon: const Icon(Icons.search, color: Color(0xFF26C6DA)),
                      suffixIcon: _searchController.text.isNotEmpty
                          ? IconButton(
                              icon: const Icon(Icons.clear),
                              onPressed: () {
                                _searchController.clear();
                                setState(() {
                                  _results = null;
                                  _error = null;
                                });
                              },
                            )
                          : null,
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                        borderSide: BorderSide(color: Colors.grey.shade300),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                        borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                      ),
                      filled: true,
                      fillColor: Colors.grey.shade50,
                    ),
                    onSubmitted: (_) => _performSearch(),
                    onChanged: (value) => setState(() {}),
                  ),
                ),
                const SizedBox(width: 8),
                FilledButton(
                  onPressed: _isSearching ? null : _performSearch,
                  style: FilledButton.styleFrom(
                    backgroundColor: const Color(0xFF26C6DA),
                    padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                  child: _isSearching
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                        )
                      : Text(l10n.searchButton),
                ),
              ],
            ),
          ),

          // Results area
          Expanded(
            child: _error != null
                ? FullScreenError(
                    error: _error!,
                    onRetry: _performSearch,
                    title: l10n.searchFailed,
                  )
                : _results == null
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.search,
                                size: 80, color: Colors.grey.shade400),
                            const SizedBox(height: 16),
                            Text(
                              l10n.semanticSearch,
                              style: Theme.of(context).textTheme.titleLarge,
                            ),
                            const SizedBox(height: 8),
                            Padding(
                              padding:
                                  const EdgeInsets.symmetric(horizontal: 32),
                              child: Text(
                                l10n.searchDescription,
                                style: Theme.of(context).textTheme.bodyMedium,
                                textAlign: TextAlign.center,
                              ),
                            ),
                            const SizedBox(height: 24),
                            Padding(
                              padding:
                                  const EdgeInsets.symmetric(horizontal: 32),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    l10n.trySearching,
                                    style: Theme.of(context)
                                        .textTheme
                                        .labelLarge
                                        ?.copyWith(
                                          color: Colors.grey[600],
                                        ),
                                  ),
                                  const SizedBox(height: 8),
                                  _buildExampleChip(l10n.budgetDiscussion),
                                  const SizedBox(height: 4),
                                  _buildExampleChip(l10n.projectDeadlines),
                                  const SizedBox(height: 4),
                                  _buildExampleChip(l10n.whoTalkedAbout),
                                ],
                              ),
                            ),
                          ],
                        ),
                      )
                    : _results!.isEmpty
                        ? Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.search_off,
                                    size: 64, color: Colors.grey.shade400),
                                const SizedBox(height: 16),
                                Text(
                                  l10n.noResultsFound,
                                  style:
                                      Theme.of(context).textTheme.titleLarge,
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  l10n.tryDifferentKeywords,
                                  style:
                                      Theme.of(context).textTheme.bodyMedium,
                                ),
                              ],
                            ),
                          )
                        : ListView.builder(
                            itemCount: _results!.length,
                            padding: const EdgeInsets.all(8),
                            itemBuilder: (context, index) {
                              final result = _results![index];
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
                                    backgroundColor: _getResultColor(result.type)
                                        .withValues(alpha: 0.15),
                                    child: Icon(
                                      _getResultIcon(result.type),
                                      color: _getResultColor(result.type),
                                      size: 26,
                                    ),
                                  ),
                                  title: Text(
                                    result.title,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  subtitle: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      const SizedBox(height: 4),
                                      Text(
                                        result.snippet,
                                        maxLines: 2,
                                        overflow: TextOverflow.ellipsis,
                                        style: TextStyle(
                                          fontSize: 13,
                                          color: Colors.grey[700],
                                        ),
                                      ),
                                      const SizedBox(height: 4),
                                      Row(
                                        children: [
                                          Chip(
                                            label: Text(
                                              _getResultTypeLabel(result.type),
                                              style: const TextStyle(
                                                fontSize: 11,
                                                fontWeight: FontWeight.w500,
                                              ),
                                            ),
                                            backgroundColor:
                                                _getResultColor(result.type)
                                                    .withValues(alpha: 0.15),
                                            side: BorderSide(
                                              color: _getResultColor(result.type)
                                                  .withValues(alpha: 0.3),
                                              width: 1,
                                            ),
                                            shape: RoundedRectangleBorder(
                                              borderRadius: BorderRadius.circular(12),
                                            ),
                                            padding: const EdgeInsets.symmetric(horizontal: 8),
                                            visualDensity:
                                                VisualDensity.compact,
                                          ),
                                          const SizedBox(width: 8),
                                          Icon(Icons.access_time,
                                              size: 12, color: Colors.grey[600]),
                                          const SizedBox(width: 2),
                                          Text(
                                            _formatDate(result.date),
                                            style: TextStyle(
                                              fontSize: 11,
                                              color: Colors.grey[600],
                                            ),
                                          ),
                                        ],
                                      ),
                                    ],
                                  ),
                                  trailing: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Text(
                                        '${(result.relevance * 100).toInt()}%',
                                        style: TextStyle(
                                          fontSize: 12,
                                          fontWeight: FontWeight.bold,
                                          color: Colors.grey[600],
                                        ),
                                      ),
                                      Text(
                                        l10n.match,
                                        style: TextStyle(
                                          fontSize: 10,
                                          color: Colors.grey[500],
                                        ),
                                      ),
                                    ],
                                  ),
                                  isThreeLine: true,
                                  onTap: () {
                                    // TODO: Navigate to result details
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      SnackBar(
                                        content: Text(
                                            'Open ${result.type}: ${result.title}'),
                                      ),
                                    );
                                  },
                                ),
                              );
                            },
                          ),
          ),
        ],
      ),
    );
  }

  Widget _buildExampleChip(String text) {
    return InkWell(
      onTap: () {
        _searchController.text = text;
        _performSearch();
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: const Color(0xFF26C6DA).withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: const Color(0xFF26C6DA).withValues(alpha: 0.3),
            width: 1,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(
              Icons.lightbulb_outline,
              size: 16,
              color: Color(0xFF00ACC1),
            ),
            const SizedBox(width: 6),
            Text(
              text,
              style: const TextStyle(
                fontSize: 13,
                color: Color(0xFF00ACC1),
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatDate(DateTime date) {
    final l10n = AppLocalizations.of(context)!;
    final now = DateTime.now();
    final difference = now.difference(date);

    if (difference.inDays == 0) {
      return l10n.today;
    } else if (difference.inDays == 1) {
      return 'Yesterday'; // TODO: Add to localization
    } else if (difference.inDays < 7) {
      return '${difference.inDays} days ago'; // TODO: Add to localization
    } else {
      return '${date.day}.${date.month}.${date.year}';
    }
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
