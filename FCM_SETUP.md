# Firebase Cloud Messaging (FCM) Setup Guide

Это руководство по настройке Firebase Cloud Messaging для отправки push-уведомлений о завершении транскрибации.

## 1. Создание проекта в Firebase Console

1. Перейдите на https://console.firebase.google.com/
2. Нажмите **"Add project"** (Добавить проект)
3. Введите название проекта: `Recontext`
4. Следуйте инструкциям мастера создания проекта
5. Google Analytics можно отключить если не нужен

## 2. Получение Service Account Credentials

Эти credentials нужны для серверной части (Go backend):

1. Откройте ваш проект в Firebase Console
2. Нажмите на шестеренку ⚙️ (Settings) → **Project settings**
3. Перейдите на вкладку **Service accounts**
4. Нажмите кнопку **"Generate new private key"**
5. Подтвердите действие - скачается JSON файл
6. Переименуйте скачанный файл в `recontext-firebase.json`
7. Поместите файл в корень проекта `/Volumes/ExternalData/source/Team21/Recontext.online/`

**⚠️ ВАЖНО:** Этот файл содержит приватные ключи! Он уже добавлен в `.gitignore` и НЕ ДОЛЖЕН коммититься в Git!

## 3. Настройка Android приложения

1. В Firebase Console: Project settings → General → Your apps
2. Нажмите на иконку **Android**
3. Введите **Package name**: Найдите в `mobile2/android/app/build.gradle`
   ```gradle
   applicationId "com.example.recontext" // Ваш package name
   ```
4. Введите **App nickname** (опционально): `Recontext Android`
5. **SHA-1** можно пропустить для начала
6. Нажмите **"Register app"**
7. Скачайте `google-services.json`
8. Поместите файл в `mobile2/android/app/google-services.json`

### Добавьте зависимости в Android

В `mobile2/android/build.gradle`:
```gradle
buildscript {
    dependencies {
        // Add this line
        classpath 'com.google.gms:google-services:4.4.0'
    }
}
```

В `mobile2/android/app/build.gradle`:
```gradle
apply plugin: 'com.android.application'
apply plugin: 'com.google.gms.google-services'  // Add this line
```

## 4. Настройка iOS приложения

1. В Firebase Console: Project settings → General → Your apps
2. Нажмите на иконку **iOS**
3. Введите **Bundle ID**: Найдите в `mobile2/ios/Runner.xcodeproj/project.pbxproj`
   ```
   PRODUCT_BUNDLE_IDENTIFIER = com.example.recontext; // Ваш Bundle ID
   ```
4. Введите **App nickname** (опционально): `Recontext iOS`
5. **App Store ID** можно пропустить
6. Нажмите **"Register app"**
7. Скачайте `GoogleService-Info.plist`
8. Откройте проект в Xcode
9. Перетащите `GoogleService-Info.plist` в папку `Runner` в Xcode
10. Убедитесь что выбран **"Copy items if needed"**

## 5. Настройка Flutter приложения

Добавьте зависимости в `mobile2/pubspec.yaml`:
```yaml
dependencies:
  firebase_core: ^2.24.2
  firebase_messaging: ^14.7.10
```

Установите пакеты:
```bash
cd mobile2
flutter pub get
```

### Инициализация Firebase

Создайте файл `mobile2/lib/services/firebase_service.dart`:
```dart
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';

class FirebaseService {
  static Future<void> initialize() async {
    await Firebase.initializeApp();

    // Request permission (iOS)
    FirebaseMessaging messaging = FirebaseMessaging.instance;
    await messaging.requestPermission(
      alert: true,
      badge: true,
      sound: true,
    );

    // Get FCM token
    String? token = await messaging.getToken();
    print('FCM Token: $token');

    // TODO: Send token to your backend to save in database

    // Handle foreground messages
    FirebaseMessaging.onMessage.listen((RemoteMessage message) {
      print('Got a message whilst in the foreground!');
      print('Message data: ${message.data}');

      if (message.notification != null) {
        print('Message also contained a notification: ${message.notification}');
        // Show local notification
      }
    });
  }
}
```

Обновите `mobile2/lib/main.dart`:
```dart
import 'package:firebase_core/firebase_core.dart';
import 'services/firebase_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Firebase
  await FirebaseService.initialize();

  runApp(MyApp());
}
```

## 6. Регистрация FCM токенов в backend

Добавьте API endpoint для регистрации FCM токенов (уже реализовано):

```
POST /api/fcm/register
Authorization: Bearer <token>
Content-Type: application/json

{
  "device_token": "FCM_TOKEN_HERE",
  "device_type": "android" // или "ios"
}
```

## 7. Перезапуск сервисов

После добавления `recontext-firebase.json` в корень проекта:

```bash
docker-compose up -d user-portal
```

## 8. Проверка работы

1. Запустите Flutter приложение
2. Получите FCM token из логов
3. Зарегистрируйте token через API
4. Завершите встречу с транскрибацией
5. Должно прийти push-уведомление: "Расшифровка готова"

## Troubleshooting

### Backend не находит credentials
```
ERROR: Failed to initialize FCM service: failed to initialize Firebase app
```

**Решение:** Проверьте что файл `recontext-firebase.json` находится в корне проекта

### Flutter не может инициализировать Firebase (Android)
```
MissingPluginException(No implementation found for method Firebase#initializeCore)
```

**Решение:**
1. Проверьте что добавили `apply plugin: 'com.google.gms.google-services'` в `app/build.gradle`
2. Выполните `flutter clean && flutter pub get`
3. Пересоберите приложение

### iOS не получает уведомления

**Решение:**
1. Проверьте что включили **Push Notifications** в Xcode capabilities
2. Добавьте **Background Modes → Remote notifications**
3. Загрузите APNs сертификат в Firebase Console

## Полезные ссылки

- [Firebase Console](https://console.firebase.google.com/)
- [Firebase Cloud Messaging Documentation](https://firebase.google.com/docs/cloud-messaging)
- [Flutter Firebase Messaging](https://firebase.flutter.dev/docs/messaging/overview)
