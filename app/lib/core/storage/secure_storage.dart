import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorage {
  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';
  static const _localeKey = 'locale';

  final FlutterSecureStorage _storage;

  SecureStorage() : _storage = const FlutterSecureStorage();

  Future<void> saveAccessToken(String token) =>
      _storage.write(key: _accessTokenKey, value: token);

  Future<String?> getAccessToken() =>
      _storage.read(key: _accessTokenKey);

  Future<void> saveRefreshToken(String token) =>
      _storage.write(key: _refreshTokenKey, value: token);

  Future<String?> getRefreshToken() =>
      _storage.read(key: _refreshTokenKey);

  Future<void> saveLocale(String locale) =>
      _storage.write(key: _localeKey, value: locale);

  Future<String?> getLocale() =>
      _storage.read(key: _localeKey);

  Future<void> clearAll() => _storage.deleteAll();
}
