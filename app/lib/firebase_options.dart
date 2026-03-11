// File generated manually from Firebase project xexplay-prod.
// FlutterFire configure equivalent.

import 'package:firebase_core/firebase_core.dart' show FirebaseOptions;
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, kIsWeb, TargetPlatform;

class DefaultFirebaseOptions {
  static FirebaseOptions get currentPlatform {
    if (kIsWeb) {
      throw UnsupportedError('Web platform is not configured for Firebase.');
    }
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        return android;
      case TargetPlatform.iOS:
        return ios;
      default:
        throw UnsupportedError(
          'DefaultFirebaseOptions are not supported for this platform.',
        );
    }
  }

  static const FirebaseOptions android = FirebaseOptions(
    apiKey: 'AIzaSyBLkcAW4KScGiy_-fbCw17nE7LO2AQSmW8',
    appId: '1:292500384931:android:fa05fe6d9e36d4c822e316',
    messagingSenderId: '292500384931',
    projectId: 'xexplay-prod',
    storageBucket: 'xexplay-prod.firebasestorage.app',
  );

  static const FirebaseOptions ios = FirebaseOptions(
    apiKey: 'AIzaSyCmvDxmw-0-ERabAlEdoRdevrHmxz-1vKg',
    appId: '1:292500384931:ios:7c1cb6a6eae230b022e316',
    messagingSenderId: '292500384931',
    projectId: 'xexplay-prod',
    storageBucket: 'xexplay-prod.firebasestorage.app',
    iosBundleId: 'io.xexplay.app',
  );
}
