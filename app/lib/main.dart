import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';
import 'core/crashlytics/crashlytics_service.dart';
import 'firebase_options.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  // Wire Crashlytics into Flutter's global error handlers.
  final crashlytics = CrashlyticsService();
  await crashlytics.init();

  runApp(
    const ProviderScope(
      child: XexPlayApp(),
    ),
  );
}
