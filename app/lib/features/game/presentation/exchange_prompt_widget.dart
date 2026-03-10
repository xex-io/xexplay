import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/constants/app_colors.dart';

/// The context in which the Exchange prompt is displayed.
enum ExchangePromptType {
  postSession,
  reward,
  achievement,
}

/// Contextual Exchange promotion widget.
///
/// Shows a dark card with a subtle blue gradient border, a headline,
/// a contextual message, and a CTA button that opens the Exchange deep link.
class ExchangePromptWidget extends StatefulWidget {
  const ExchangePromptWidget({
    super.key,
    required this.type,
    this.onDismiss,
  });

  final ExchangePromptType type;
  final VoidCallback? onDismiss;

  @override
  State<ExchangePromptWidget> createState() => _ExchangePromptWidgetState();
}

class _ExchangePromptWidgetState extends State<ExchangePromptWidget> {
  bool _dismissed = false;

  static const _exchangeUrl = 'https://xex.exchange';

  String get _message => switch (widget.type) {
        ExchangePromptType.postSession =>
          'Convert your skills to real trading',
        ExchangePromptType.reward => 'Claim tokens on XEX Exchange',
        ExchangePromptType.achievement =>
          'Unlock exclusive trading benefits',
      };

  Future<void> _openExchange() async {
    final uri = Uri.parse(_exchangeUrl);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_dismissed) return const SizedBox.shrink();

    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surface =
        isDark ? AppColors.darkSurface : AppColors.lightSurface;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final primary =
        isDark ? AppColors.darkPrimaryBold : AppColors.lightPrimaryBold;

    return Container(
      margin: const EdgeInsets.symmetric(vertical: 8),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        gradient: LinearGradient(
          colors: [
            primary.withAlpha(60),
            primary.withAlpha(25),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
      ),
      child: Container(
        margin: const EdgeInsets.all(1.5),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(14.5),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                Icon(Icons.show_chart_rounded, color: primary, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Trade on XEX Exchange',
                    style: Theme.of(context).textTheme.titleSmall?.copyWith(
                          color: textPrimary,
                          fontWeight: FontWeight.w700,
                        ),
                  ),
                ),
                GestureDetector(
                  onTap: () {
                    setState(() => _dismissed = true);
                    widget.onDismiss?.call();
                  },
                  child: Icon(Icons.close, size: 18, color: textSecondary),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              _message,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: textSecondary),
            ),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: GestureDetector(
                onTap: _openExchange,
                child: Container(
                  height: 40,
                  decoration: BoxDecoration(
                    color: primary,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  alignment: Alignment.center,
                  child: Text(
                    'Open Exchange',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Colors.white,
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
