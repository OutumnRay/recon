import 'package:flutter/material.dart';
import 'package:recontext/screens/login_screen.dart';

void main() {
  runApp(const RecontextApp());
}

class RecontextApp extends StatelessWidget {
  const RecontextApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Recontext',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF46afba), // Seafoam color from design
        ),
        useMaterial3: true,
      ),
      home: const LoginScreen(),
    );
  }
}
