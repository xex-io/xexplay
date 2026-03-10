import 'package:dio/dio.dart';
import 'package:firebase_crashlytics/firebase_crashlytics.dart';

class ErrorInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    if (err.response?.statusCode == 401) {
      // Token expired or invalid — will be handled by auth state
    }

    // Record non-fatal network errors in Crashlytics so they surface
    // in the dashboard without crashing the app.
    if (err.type != DioExceptionType.cancel) {
      FirebaseCrashlytics.instance.recordError(
        err,
        err.stackTrace,
        reason: 'Dio ${err.type.name}: '
            '${err.requestOptions.method} ${err.requestOptions.uri}',
        fatal: false,
      );
    }

    handler.next(err);
  }
}
