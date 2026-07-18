import 'dart:io';
import 'dart:math';

import 'package:fl_clash/common/common.dart';
import 'package:path/path.dart';

/// Install-scoped secret for authenticating to FlClashHelperService.
///
/// Stored under the app support directory (UI) and mirrored to
/// `%ProgramData%\FlClash\helper.auth` (helper) during elevated registration.
class HelperAuth {
  static HelperAuth? _instance;

  HelperAuth._();

  factory HelperAuth() {
    _instance ??= HelperAuth._();
    return _instance!;
  }

  String generateSecret() {
    final random = Random.secure();
    final bytes = List<int>.generate(32, (_) => random.nextInt(256));
    return bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
  }

  Future<String> get localSecretPath async {
    final home = await appPath.homeDirPath;
    return join(home, 'helper.auth');
  }

  String get programDataSecretPath {
    final base = Platform.environment['PROGRAMDATA'] ?? r'C:\ProgramData';
    return join(base, 'FlClash', 'helper.auth');
  }

  Future<String?> loadOrCreate() async {
    if (!system.isWindows) {
      return null;
    }
    final path = await localSecretPath;
    final file = File(path);
    if (await file.exists()) {
      final existing = (await file.readAsString()).trim();
      if (existing.length >= 32) {
        return existing;
      }
    }
    final secret = generateSecret();
    await file.parent.create(recursive: true);
    await file.writeAsString(secret, flush: true);
    return secret;
  }

  /// Elevated install step: write ProgramData secret + lock down ACL.
  Future<String> prepareElevatedInstallScript() async {
    final secret = await loadOrCreate();
    if (secret == null) {
      throw StateError('helper auth secret unavailable');
    }
    final target = programDataSecretPath;
    final dir = dirname(target);
    // Escape for cmd.exe double quotes.
    final escapedSecret = secret.replaceAll('"', '');
    return [
      'if not exist "$dir" mkdir "$dir"',
      'echo $escapedSecret> "$target"',
      'icacls "$target" /inheritance:r >nul',
      'icacls "$target" /grant:r *S-1-5-18:F >nul',
      'icacls "$target" /grant:r *S-1-5-32-544:F >nul',
      'icacls "$target" /grant:r "%USERNAME%":R >nul',
    ].join(' & ');
  }
}

final helperAuth = HelperAuth();
