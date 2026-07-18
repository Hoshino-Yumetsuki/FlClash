// Standalone path-sandbox check without mihomo dependency.
// Run: dart run tool/path_sandbox_check.dart
// (mirrors core/path_sandbox.go logic)

import 'dart:io';

String clean(String p) => p.replaceAll(r'\', '/');

bool underHome(String home, String path) {
  final absHome = clean(File(home).absolute.path);
  final absPath = clean(File(path).absolute.path);
  if (absPath == absHome) return true;
  final prefix = absHome.endsWith('/') ? absHome : '$absHome/';
  return absPath.startsWith(prefix);
}

void main() {
  final home = Directory.systemTemp.createTempSync('flclash_home_').path;
  final inside = '$home/profiles/a.yaml';
  Directory('$home/profiles').createSync(recursive: true);
  File(inside).writeAsStringSync('x');

  assert(underHome(home, inside), 'inside should pass');
  assert(!underHome(home, '$home/../escape.txt'), 'outside should fail');
  assert(
    !underHome(home, '$home/profiles/../../etc/passwd'),
    'traversal should fail',
  );
  print('path_sandbox_check: ok');
  Directory(home).deleteSync(recursive: true);
}
