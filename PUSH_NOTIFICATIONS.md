# Push Notifications Implementation Guide

## Статус реализации

### ✅ Завершено

1. **Исправлена критическая ошибка** с JSON полем `summaries` в placeholder rooms
2. **Flutter Mobile:**
   - Добавлены зависимости: `firebase_core`, `firebase_messaging`, `flutter_local_notifications`
   - Создан `PushNotificationService` для обработки push-уведомлений
   - Создан `DeeplinkHandler` для навигации из уведомлений
3. **Real-time уведомления через WebSocket:**
   - Backend отправляет уведомления при: composite_video, transcription, summary
   - Frontend и Mobile получают уведомления в реальном времени

### 🚧 Требуется завершить

## 1. Flutter Mobile - Интеграция в main.dart

```dart
// mobile2/lib/main.dart

import 'package:firebase_core/firebase_core.dart';
import 'services/push_notification_service.dart';
import 'services/deeplink_handler.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Firebase
  await Firebase.initializeApp();

  // Initialize Push Notifications
  final pushService = PushNotificationService();
  await pushService.initialize();

  // Set up deeplink handler
  pushService.onNotificationTap = (data) {
    DeeplinkHandler.handleDeeplink(data);
  };

  // Set up foreground notification handler (show in-app)
  pushService.onForegroundMessage = (message) {
    // Show in-app notification when app is open
    // Get current context and show SnackBar
    final context = DeeplinkHandler.navigatorKey.currentContext;
    if (context != null) {
      PushNotificationService.showInAppNotification(
        context,
        title: message.notification?.title ?? 'Notification',
        message: message.notification?.body ?? '',
        onTap: () {
          DeeplinkHandler.handleDeeplink(message.data);
        },
      );
    }
  };

  runApp(MyApp(
    navigatorKey: DeeplinkHandler.navigatorKey,
  ));
}

class MyApp extends StatelessWidget {
  final GlobalKey<NavigatorState> navigatorKey;

  const MyApp({Key? key, required this.navigatorKey}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      navigatorKey: navigatorKey, // ВАЖНО!
      // ... остальной код
    );
  }
}
```

## 2. Backend - Регистрация FCM токенов

### Добавить модель в базу данных:

```sql
CREATE TABLE user_fcm_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL,
    platform VARCHAR(20) NOT NULL, -- 'android', 'ios', 'web'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, fcm_token)
);

CREATE INDEX idx_user_fcm_tokens_user_id ON user_fcm_tokens(user_id);
```

### Добавить endpoint для регистрации токена:

```go
// cmd/user-portal/handlers_push.go

package main

import (
    "encoding/json"
    "net/http"
    "Recontext.online/pkg/auth"
)

type RegisterFCMTokenRequest struct {
    FCMToken string `json:"fcm_token"`
    Platform string `json:"platform"` // android, ios, web
}

// @Summary Register FCM token
// @Description Register device FCM token for push notifications
// @Tags Push Notifications
// @Accept json
// @Produce json
// @Param request body RegisterFCMTokenRequest true "FCM Token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/push/register [post]
func (up *UserPortal) registerFCMTokenHandler(w http.ResponseWriter, r *http.Request) {
    claims, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
        return
    }

    var req RegisterFCMTokenRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
        return
    }

    // Save token to database (upsert)
    _, err := up.db.DB.Exec(`
        INSERT INTO user_fcm_tokens (user_id, fcm_token, platform, updated_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (user_id, fcm_token)
        DO UPDATE SET updated_at = NOW()
    `, claims.UserID, req.FCMToken, req.Platform)

    if err != nil {
        up.respondWithError(w, http.StatusInternalServerError, "Failed to save token", err.Error())
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "FCM token registered successfully",
    })
}
```

### Регистрировать маршрут:

```go
// cmd/user-portal/main.go
mux.HandleFunc("POST /api/v1/push/register", authMiddleware(portal.registerFCMTokenHandler))
```

## 3. Backend - Отправка Push-уведомлений

### Установить Firebase Admin SDK:

```bash
go get firebase.google.com/go/v4
```

### Создать сервис отправки push:

```go
// pkg/push/fcm_service.go

package push

import (
    "context"
    "fmt"
    "log"

    firebase "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/messaging"
    "google.golang.org/api/option"
    "gorm.io/gorm"
)

type FCMService struct {
    client *messaging.Client
    db     *gorm.DB
}

func NewFCMService(credentialsPath string, db *gorm.DB) (*FCMService, error) {
    ctx := context.Background()

    opt := option.WithCredentialsFile(credentialsPath)
    app, err := firebase.NewApp(ctx, nil, opt)
    if err != nil {
        return nil, fmt.Errorf("error initializing firebase app: %w", err)
    }

    client, err := app.Messaging(ctx)
    if err != nil {
        return nil, fmt.Errorf("error getting messaging client: %w", err)
    }

    return &FCMService{
        client: client,
        db:     db,
    }, nil
}

// SendToUser отправляет push всем устройствам пользователя
func (f *FCMService) SendToUser(userID string, notification *messaging.Notification, data map[string]string) error {
    // Получить все токены пользователя
    var tokens []string
    err := f.db.Raw(`
        SELECT fcm_token FROM user_fcm_tokens
        WHERE user_id = $1 AND updated_at > NOW() - INTERVAL '90 days'
    `, userID).Scan(&tokens).Error

    if err != nil {
        return fmt.Errorf("failed to get user tokens: %w", err)
    }

    if len(tokens) == 0 {
        log.Printf("No FCM tokens found for user %s", userID)
        return nil
    }

    // Отправить multicast сообщение
    message := &messaging.MulticastMessage{
        Notification: notification,
        Data:         data,
        Tokens:       tokens,
    }

    ctx := context.Background()
    response, err := f.client.SendMulticast(ctx, message)
    if err != nil {
        return fmt.Errorf("failed to send multicast message: %w", err)
    }

    log.Printf("Successfully sent %d messages to user %s", response.SuccessCount, userID)

    // Удалить невалидные токены
    if response.FailureCount > 0 {
        f.removeInvalidTokens(tokens, response.Responses)
    }

    return nil
}

func (f *FCMService) removeInvalidTokens(tokens []string, responses []*messaging.SendResponse) {
    for i, resp := range responses {
        if resp.Error != nil && messaging.IsRegistrationTokenNotRegistered(resp.Error) {
            // Удалить токен из БД
            f.db.Exec("DELETE FROM user_fcm_tokens WHERE fcm_token = $1", tokens[i])
        }
    }
}
```

### Интегрировать с NotificationService:

```go
// pkg/notifications/notification_service.go

type NotificationService struct {
    subscribers map[string]*Subscriber
    mu          sync.RWMutex
    fcmService  *push.FCMService // Добавить
}

func (ns *NotificationService) Notify(notification *Notification) {
    // Отправить через WebSocket (существующая логика)
    ns.mu.RLock()
    for _, subscriber := range ns.subscribers {
        if ns.shouldReceiveNotification(subscriber, notification) {
            select {
            case subscriber.Channel <- notification:
            default:
            }
        }
    }
    ns.mu.RUnlock()

    // Отправить через FCM (новая логика)
    if ns.fcmService != nil && notification.UserID != nil {
        go ns.sendPushNotification(notification)
    }
}

func (ns *NotificationService) sendPushNotification(n *Notification) {
    // Формируем FCM уведомление
    fcmNotification := &messaging.Notification{
        Title: ns.getNotificationTitle(n.Type),
        Body:  ns.getNotificationBody(n),
    }

    // Формируем data payload для deeplinks
    data := map[string]string{
        "type":       string(n.Type),
        "entity_id":  n.EntityID,
        "meeting_id": n.MeetingID.String(),
    }

    // Отправить всем устройствам пользователя
    if err := ns.fcmService.SendToUser(n.UserID.String(), fcmNotification, data); err != nil {
        log.Printf("Failed to send push notification: %v", err)
    }
}

func (ns *NotificationService) getNotificationTitle(eventType EventType) string {
    switch eventType {
    case EventCompositeVideoCompleted:
        return "Видео готово"
    case EventTranscriptionCompleted:
        return "Транскрипция завершена"
    case EventSummaryCompleted:
        return "Сводка готова"
    default:
        return "Уведомление"
    }
}

func (ns *NotificationService) getNotificationBody(n *Notification) string {
    return n.Message
}
```

## 4. Web Frontend - Firebase Web Push

### Установить Firebase SDK:

```bash
cd front/user-portal
npm install firebase
```

### Создать firebase-messaging-sw.js:

```javascript
// front/user-portal/public/firebase-messaging-sw.js

importScripts('https://www.gstatic.com/firebasejs/10.7.1/firebase-app-compat.js');
importScripts('https://www.gstatic.com/firebasejs/10.7.1/firebase-messaging-compat.js');

firebase.initializeApp({
  apiKey: "YOUR_API_KEY",
  authDomain: "YOUR_AUTH_DOMAIN",
  projectId: "YOUR_PROJECT_ID",
  storageBucket: "YOUR_STORAGE_BUCKET",
  messagingSenderId: "YOUR_MESSAGING_SENDER_ID",
  appId: "YOUR_APP_ID"
});

const messaging = firebase.messaging();

messaging.onBackgroundMessage((payload) => {
  console.log('[firebase-messaging-sw.js] Received background message ', payload);

  const notificationTitle = payload.notification.title;
  const notificationOptions = {
    body: payload.notification.body,
    icon: '/icon-192x192.png',
    data: payload.data
  };

  self.registration.showNotification(notificationTitle, notificationOptions);
});
```

### Добавить web push service:

```typescript
// front/user-portal/src/services/webPushService.ts

import { initializeApp } from 'firebase/app';
import { getMessaging, getToken, onMessage } from 'firebase/messaging';

const firebaseConfig = {
  // ... ваша конфигурация
};

const app = initializeApp(firebaseConfig);
const messaging = getMessaging(app);

export async function registerWebPush() {
  try {
    const token = await getToken(messaging, {
      vapidKey: 'YOUR_VAPID_KEY'
    });

    // Send token to backend
    await fetch('/api/v1/push/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getToken()}`
      },
      body: JSON.stringify({
        fcm_token: token,
        platform: 'web'
      })
    });

    // Handle foreground messages
    onMessage(messaging, (payload) => {
      console.log('Foreground message:', payload);
      // Show in-app notification
    });
  } catch (err) {
    console.error('Failed to register web push:', err);
  }
}
```

## 5. Конфигурация Firebase

### Необходимые файлы:

1. **Android:** `android/app/google-services.json`
2. **iOS:** `ios/Runner/GoogleService-Info.plist`
3. **Backend:** `firebase-adminsdk.json` (Service Account Key)

### Получение файлов:

1. Зайти в Firebase Console
2. Создать проект или использовать существующий
3. Добавить приложения (Android, iOS, Web)
4. Скачать конфигурационные файлы
5. Для backend: Project Settings → Service Accounts → Generate new private key

## Следующие шаги

1. Завершить интеграцию в mobile2/lib/main.dart
2. Добавить backend endpoints для регистрации FCM токенов
3. Настроить Firebase проект и получить конфигурационные файлы
4. Интегрировать FCMService с NotificationService
5. Добавить web push для frontend
6. Тестирование на всех платформах

## Формат Push Notification Data

```json
{
  "notification": {
    "title": "Видео готово",
    "body": "Композитное видео для встречи готово к просмотру"
  },
  "data": {
    "type": "composite_video.completed",
    "entity_id": "room_sid",
    "meeting_id": "meeting_uuid",
    "video_playlist_url": "https://..."
  }
}
```
