import 'package:flutter/material.dart';
import '../models/transcript.dart';
import '../l10n/app_localizations.dart';

class TranscriptViewer extends StatelessWidget {
  final RoomTranscripts transcripts;
  final Function(double)? onSeekToTime;

  const TranscriptViewer({
    super.key,
    required this.transcripts,
    this.onSeekToTime,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    if (!transcripts.hasTranscripts) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.text_snippet_outlined,
                size: 64,
                color: Colors.grey[400],
              ),
              const SizedBox(height: 16),
              Text(
                l10n.noTranscriptAvailable,
                style: TextStyle(
                  fontSize: 16,
                  color: Colors.grey[600],
                  fontWeight: FontWeight.w500,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                l10n.transcriptionNotGenerated,
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey[500],
                ),
              ),
            ],
          ),
        ),
      );
    }

    final mergedPhrases = transcripts.getMergedPhrases();

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: mergedPhrases.length,
      itemBuilder: (context, index) {
        final phrase = mergedPhrases[index];
        return _TranscriptPhraseCard(
          phrase: phrase,
          onTap: onSeekToTime != null
              ? () => onSeekToTime!(phrase.startTime)
              : null,
        );
      },
    );
  }
}

class _TranscriptPhraseCard extends StatelessWidget {
  final MergedTranscriptPhrase phrase;
  final VoidCallback? onTap;

  const _TranscriptPhraseCard({
    required this.phrase,
    this.onTap,
  });

  String _formatTimestamp(double seconds) {
    final duration = Duration(milliseconds: (seconds * 1000).toInt());
    final hours = duration.inHours;
    final minutes = duration.inMinutes.remainder(60);
    final secs = duration.inSeconds.remainder(60);

    if (hours > 0) {
      return '${hours.toString().padLeft(2, '0')}:${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
    }
    return '${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  Color _getColorForSpeaker(String speakerName) {
    // Generate a consistent color based on speaker name
    final hash = speakerName.hashCode;
    final hue = (hash % 360).toDouble();
    return HSLColor.fromAHSL(1.0, hue, 0.6, 0.5).toColor();
  }

  @override
  Widget build(BuildContext context) {
    final speakerColor = _getColorForSpeaker(phrase.speakerName);

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      elevation: 1,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Speaker avatar
              CircleAvatar(
                radius: 20,
                backgroundColor: speakerColor.withValues(alpha: 0.2),
                child: Text(
                  phrase.speakerInitials,
                  style: TextStyle(
                    color: speakerColor,
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              // Transcript content
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Speaker name and timestamp
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          phrase.speakerName,
                          style: TextStyle(
                            fontWeight: FontWeight.w600,
                            fontSize: 14,
                            color: speakerColor,
                          ),
                        ),
                        Text(
                          _formatTimestamp(phrase.startTime),
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    // Phrase text
                    Text(
                      phrase.text,
                      style: const TextStyle(
                        fontSize: 14,
                        height: 1.4,
                      ),
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
}
