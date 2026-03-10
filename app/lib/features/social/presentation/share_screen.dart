import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/app_button.dart';
import 'share_card_widget.dart';

/// Screen that previews a branded share card and lets the user share it.
class ShareScreen extends StatefulWidget {
  const ShareScreen({super.key, required this.data});

  final ShareData data;

  @override
  State<ShareScreen> createState() => _ShareScreenState();
}

class _ShareScreenState extends State<ShareScreen> {
  final _repaintKey = GlobalKey();
  bool _sharing = false;

  Future<void> _shareImage() async {
    setState(() => _sharing = true);
    try {
      final image = await ShareCardController.captureImage(_repaintKey);
      if (image == null) return;

      final byteData = await image.toByteData(format: ui.ImageByteFormat.png);
      if (byteData == null) return;

      final bytes = byteData.buffer.asUint8List();
      await Share.shareXFiles(
        [
          XFile.fromData(
            bytes,
            mimeType: 'image/png',
            name: 'xexplay_share.png',
          ),
        ],
        text: 'Check out my results on XEX Play!',
      );
    } finally {
      if (mounted) setState(() => _sharing = false);
    }
  }

  Future<void> _copyLink() async {
    await Clipboard.setData(
      const ClipboardData(text: 'https://play.xex.com'),
    );
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Link copied to clipboard'),
        behavior: SnackBarBehavior.floating,
        duration: Duration(seconds: 2),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final bgColor =
        isDark ? AppColors.darkBackground : AppColors.lightBackground;

    return Scaffold(
      backgroundColor: bgColor,
      appBar: AppBar(
        title: const Text('Share'),
        centerTitle: true,
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          child: Column(
            children: [
              const Spacer(),

              // Card preview
              Center(
                child: ShareCardWidget(
                  data: widget.data,
                  repaintKey: _repaintKey,
                ),
              ),

              const Spacer(),

              // Share button
              SizedBox(
                width: double.infinity,
                child: PrimaryButton(
                  label: 'Share to...',
                  isLoading: _sharing,
                  onPressed: _sharing ? null : _shareImage,
                ),
              ),
              const SizedBox(height: 12),

              // Copy link button
              SizedBox(
                width: double.infinity,
                child: SecondaryButton(
                  label: 'Copy Link',
                  onPressed: _copyLink,
                ),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ),
    );
  }
}
