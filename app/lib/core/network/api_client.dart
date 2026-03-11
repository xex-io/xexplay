import 'package:dio/dio.dart';
import '../config/environment.dart';
import 'interceptors/auth_interceptor.dart';
import 'interceptors/error_interceptor.dart';

class ApiClient {
  late final Dio dio;

  ApiClient({required AuthInterceptor authInterceptor}) {
    dio = Dio(
      BaseOptions(
        baseUrl: Environment.apiBaseUrl,
        connectTimeout: const Duration(seconds: 10),
        receiveTimeout: const Duration(seconds: 10),
        headers: {'Content-Type': 'application/json'},
      ),
    );

    dio.interceptors.addAll([
      authInterceptor,
      ErrorInterceptor(),
    ]);
  }
}
