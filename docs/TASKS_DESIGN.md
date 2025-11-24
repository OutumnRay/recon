# Tasks System Design

## Обзор

Система задач (Tasks) для управления действиями после встреч с поддержкой AI-извлечения из транскрипций.

## Модель данных

### Task (Задача)

```go
type Task struct {
    ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

    // Связь с сессией/встречей
    SessionID   uuid.UUID  `gorm:"type:uuid;not null;index"`
    MeetingID   *uuid.UUID `gorm:"type:uuid;index"` // Опционально - привязка к встрече

    // Информация о задаче
    Title       string     `gorm:"type:varchar(500);not null"` // Краткое описание
    Description *string    `gorm:"type:text"` // Подробное описание
    Hint        *string    `gorm:"type:text"` // Подсказка как решить

    // Назначение
    AssignedTo  *uuid.UUID `gorm:"type:uuid;index"` // Кому назначена (NULL = никому)
    AssignedBy  *uuid.UUID `gorm:"type:uuid"` // Кто назначил

    // Статус
    Status      string     `gorm:"type:varchar(50);not null;default:'pending';index"` // pending, in_progress, completed, cancelled
    Priority    string     `gorm:"type:varchar(50);default:'medium'"` // low, medium, high, urgent

    // Дедлайн
    DueDate     *time.Time `gorm:"index"` // Срок выполнения

    // AI extraction metadata
    ExtractedByAI bool      `gorm:"default:false"` // Извлечена AI или создана вручную
    AIConfidence  *float64  `gorm:"type:double precision"` // Уверенность AI (0-1)
    SourceSegment *string   `gorm:"type:text"` // Фрагмент транскрипции откуда извлечена

    // Timestamps
    CompletedAt *time.Time
    CreatedAt   time.Time  `gorm:"autoCreateTime"`
    UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}
```

### TaskStatus (Enum)

- `pending` - Ожидает выполнения
- `in_progress` - В работе
- `completed` - Завершена
- `cancelled` - Отменена

### TaskPriority (Enum)

- `low` - Низкий приоритет
- `medium` - Средний приоритет
- `high` - Высокий приоритет
- `urgent` - Срочно

## API Endpoints

### Managing Portal (Admin)

```
GET    /api/v1/sessions/{session_id}/tasks              - Список задач сессии
POST   /api/v1/sessions/{session_id}/tasks              - Создать задачу
GET    /api/v1/sessions/{session_id}/tasks/{task_id}    - Детали задачи
PUT    /api/v1/sessions/{session_id}/tasks/{task_id}    - Обновить задачу
DELETE /api/v1/sessions/{session_id}/tasks/{task_id}    - Удалить задачу

POST   /api/v1/sessions/{session_id}/tasks/extract      - AI-извлечение задач из транскрипции
```

### User Portal

```
GET    /api/v1/my-tasks                                  - Мои задачи (все)
GET    /api/v1/my-tasks?status=pending                   - Мои задачи по статусу
PUT    /api/v1/tasks/{task_id}/status                    - Изменить статус задачи
GET    /api/v1/meetings/{meeting_id}/tasks               - Задачи встречи
```

## Request/Response Models

### CreateTaskRequest

```json
{
  "title": "Подготовить отчет по продажам",
  "description": "Нужно собрать данные за Q4 и создать презентацию",
  "hint": "Использовать шаблон из прошлого квартала",
  "assigned_to": "user-uuid-here",
  "due_date": "2025-12-31T23:59:59Z",
  "priority": "high"
}
```

### UpdateTaskRequest

```json
{
  "title": "Обновленное название",
  "description": "Обновленное описание",
  "hint": "Новая подсказка",
  "assigned_to": "user-uuid-here",
  "status": "in_progress",
  "priority": "urgent",
  "due_date": "2025-12-15T23:59:59Z"
}
```

### UpdateTaskStatusRequest

```json
{
  "status": "completed"
}
```

### TaskResponse

```json
{
  "id": "task-uuid",
  "session_id": "session-uuid",
  "meeting_id": "meeting-uuid",
  "title": "Подготовить отчет",
  "description": "Подробное описание",
  "hint": "Подсказка как решить",
  "assigned_to": "user-uuid",
  "assigned_by": "user-uuid",
  "assigned_to_user": {
    "id": "user-uuid",
    "username": "ivan.petrov",
    "email": "ivan@example.com",
    "first_name": "Иван",
    "last_name": "Петров"
  },
  "assigned_by_user": {
    "id": "user-uuid",
    "username": "admin",
    "email": "admin@example.com"
  },
  "status": "pending",
  "priority": "high",
  "due_date": "2025-12-31T23:59:59Z",
  "extracted_by_ai": false,
  "ai_confidence": null,
  "source_segment": null,
  "completed_at": null,
  "created_at": "2025-11-24T12:00:00Z",
  "updated_at": "2025-11-24T12:00:00Z"
}
```

### ListTasksResponse

```json
{
  "items": [/* TaskResponse[] */],
  "total": 25,
  "page_size": 20,
  "offset": 0
}
```

## AI Task Extraction

### Prompt Template для LLM

```
You are analyzing a meeting transcript to extract actionable tasks.

Transcript:
{transcript_text}

Participants:
{participants_list}

Extract all actionable tasks from this transcript. For each task:
1. Create a clear, concise title (max 100 characters)
2. Provide a detailed description
3. Identify who should do it (if mentioned)
4. Suggest a hint on how to solve it (if context allows)
5. Estimate priority based on urgency in discussion

Return JSON array:
[
  {
    "title": "Task title",
    "description": "Detailed description",
    "hint": "How to solve (optional)",
    "assigned_to_username": "username or null",
    "priority": "low|medium|high|urgent",
    "confidence": 0.95,
    "source_segment": "relevant quote from transcript"
  }
]

Rules:
- Only extract explicit action items, not general discussion
- If no clear assignee, set assigned_to_username to null
- Confidence should reflect how clear the task is (0.0-1.0)
- Source segment should be the exact quote where task was mentioned
```

### AI Extraction Endpoint

```
POST /api/v1/sessions/{session_id}/tasks/extract
```

**Request:**
```json
{
  "llm_provider": "openai",  // openai, anthropic, ollama
  "model": "gpt-4",
  "auto_assign": true,       // Автоматически назначать по username
  "min_confidence": 0.7      // Минимальная уверенность для сохранения
}
```

**Response:**
```json
{
  "extracted_count": 5,
  "saved_count": 4,
  "skipped_count": 1,
  "tasks": [/* TaskResponse[] */]
}
```

## Frontend Components

### Tasks Tab (Meeting Detail Screen - Mobile)

```dart
// mobile2/lib/screens/meeting_detail_screen.dart

// Новая вкладка "Tasks" (Задачи)
Tab(text: l10n.tabTasks),

// Content
TabBarView(
  children: [
    _buildInfoTab(),
    _buildRecordingsTab(),
    _buildTasksTab(),  // НОВАЯ ВКЛАДКА
  ],
)

Widget _buildTasksTab() {
  if (_isLoadingTasks) {
    return Center(child: CircularProgressIndicator());
  }

  if (_tasks.isEmpty) {
    return _buildEmptyTasksState();
  }

  return ListView.builder(
    itemCount: _tasks.length,
    itemBuilder: (context, index) {
      return TaskCard(
        task: _tasks[index],
        onStatusChange: _updateTaskStatus,
      );
    },
  );
}
```

### Task Card Widget

```dart
// mobile2/lib/widgets/task_card.dart

class TaskCard extends StatelessWidget {
  final Task task;
  final Function(String) onStatusChange;

  Widget build(BuildContext context) {
    return Card(
      child: Column(
        children: [
          // Заголовок с приоритетом
          ListTile(
            leading: _getPriorityIcon(),
            title: Text(task.title),
            subtitle: task.description,
            trailing: _getStatusBadge(),
          ),

          // Назначение
          if (task.assignedToUser != null)
            _buildAssigneeRow(),

          // Подсказка
          if (task.hint != null)
            _buildHintSection(),

          // Дедлайн
          if (task.dueDate != null)
            _buildDueDateRow(),

          // Действия
          _buildActionButtons(),
        ],
      ),
    );
  }
}
```

### Managing Portal (React)

```typescript
// front/managing-portal/src/components/SessionTasks.tsx

interface Task {
  id: string;
  title: string;
  description?: string;
  hint?: string;
  assigned_to?: string;
  assigned_to_user?: User;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  due_date?: string;
  extracted_by_ai: boolean;
  ai_confidence?: number;
  source_segment?: string;
}

export const SessionTasks: React.FC<{sessionId: string}> = ({sessionId}) => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isExtracting, setIsExtracting] = useState(false);

  const extractTasksWithAI = async () => {
    setIsExtracting(true);
    const result = await api.post(`/sessions/${sessionId}/tasks/extract`, {
      llm_provider: 'openai',
      model: 'gpt-4',
      auto_assign: true,
      min_confidence: 0.7
    });
    setTasks([...tasks, ...result.tasks]);
    setIsExtracting(false);
  };

  return (
    <div>
      <button onClick={extractTasksWithAI} disabled={isExtracting}>
        {isExtracting ? 'Извлечение задач...' : 'Извлечь задачи из транскрипции'}
      </button>

      <TaskList tasks={tasks} onUpdate={handleTaskUpdate} />
    </div>
  );
};
```

## Database Migration

```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES livekit_rooms(id) ON DELETE CASCADE,
    meeting_id UUID REFERENCES meetings(id) ON DELETE SET NULL,

    title VARCHAR(500) NOT NULL,
    description TEXT,
    hint TEXT,

    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority VARCHAR(50) DEFAULT 'medium',

    due_date TIMESTAMPTZ,

    extracted_by_ai BOOLEAN DEFAULT FALSE,
    ai_confidence DOUBLE PRECISION,
    source_segment TEXT,

    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_session_id ON tasks(session_id);
CREATE INDEX idx_tasks_meeting_id ON tasks(meeting_id);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
```

## Implementation Phases

### Phase 1: Basic CRUD (Backend)
1. ✅ Спроектировать модель Task
2. ⏳ Создать миграцию базы данных
3. ⏳ Реализовать CRUD endpoints в Managing Portal
4. ⏳ Реализовать User Portal endpoints (my-tasks)
5. ⏳ Написать тесты

### Phase 2: Frontend (Mobile)
1. ⏳ Создать модели Task в Dart
2. ⏳ Добавить вкладку Tasks в meeting_detail_screen
3. ⏳ Создать TaskCard widget
4. ⏳ Реализовать обновление статуса
5. ⏳ Добавить My Tasks экран

### Phase 3: AI Extraction
1. ⏳ Интегрировать LLM API (OpenAI/Ollama)
2. ⏳ Реализовать prompt для извлечения задач
3. ⏳ Создать endpoint /tasks/extract
4. ⏳ Добавить UI кнопку извлечения в Managing Portal

### Phase 4: Advanced Features
1. ⏳ Push-уведомления о новых задачах
2. ⏳ Напоминания о дедлайнах
3. ⏳ Фильтры и поиск по задачам
4. ⏳ Статистика выполнения задач

## Localization Keys (i18n)

### English
```json
{
  "tabTasks": "Tasks",
  "tasks": "Tasks",
  "myTasks": "My Tasks",
  "taskTitle": "Task Title",
  "taskDescription": "Description",
  "taskHint": "Hint",
  "assignedTo": "Assigned To",
  "assignedBy": "Assigned By",
  "dueDate": "Due Date",
  "priority": "Priority",
  "priorityLow": "Low",
  "priorityMedium": "Medium",
  "priorityHigh": "High",
  "priorityUrgent": "Urgent",
  "statusPending": "Pending",
  "statusInProgress": "In Progress",
  "statusCompleted": "Completed",
  "statusCancelled": "Cancelled",
  "noTasksFound": "No tasks found",
  "createTask": "Create Task",
  "extractTasks": "Extract Tasks from Transcript",
  "extractedByAI": "Extracted by AI",
  "aiConfidence": "AI Confidence",
  "markAsCompleted": "Mark as Completed",
  "markAsInProgress": "Mark as In Progress"
}
```

### Russian
```json
{
  "tabTasks": "Задачи",
  "tasks": "Задачи",
  "myTasks": "Мои задачи",
  "taskTitle": "Название задачи",
  "taskDescription": "Описание",
  "taskHint": "Подсказка",
  "assignedTo": "Назначена",
  "assignedBy": "Назначил",
  "dueDate": "Срок",
  "priority": "Приоритет",
  "priorityLow": "Низкий",
  "priorityMedium": "Средний",
  "priorityHigh": "Высокий",
  "priorityUrgent": "Срочно",
  "statusPending": "Ожидает",
  "statusInProgress": "В работе",
  "statusCompleted": "Завершена",
  "statusCancelled": "Отменена",
  "noTasksFound": "Задачи не найдены",
  "createTask": "Создать задачу",
  "extractTasks": "Извлечь задачи из транскрипции",
  "extractedByAI": "Извлечено AI",
  "aiConfidence": "Уверенность AI",
  "markAsCompleted": "Отметить как выполненную",
  "markAsInProgress": "Отметить как в работе"
}
```

## Next Steps

Выберите этап для реализации:

1. **Phase 1: Basic CRUD** - Базовый функционал без AI
2. **Phase 2: Mobile UI** - Интерфейс в мобильном приложении
3. **Phase 3: AI Extraction** - AI-извлечение задач
4. **All at once** - Реализовать все сразу (займет больше времени)

Какой этап начать первым?
