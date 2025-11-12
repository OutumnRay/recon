package models

import "time"

// ServiceStatus представляет статус сервиса
type ServiceStatus string

const (
	ServiceStatusRunning ServiceStatus = "running" // Работает
	ServiceStatusStopped ServiceStatus = "stopped" // Остановлен
	ServiceStatusError   ServiceStatus = "error"   // Ошибка
	ServiceStatusUnknown ServiceStatus = "unknown" // Неизвестен
)

// ServiceType представляет тип сервиса
type ServiceType string

const (
	ServiceTypeManagingPortal      ServiceType = "managing-portal"      // Портал управления
	ServiceTypeUserPortal          ServiceType = "user-portal"          // Пользовательский портал
	ServiceTypeTranscriptionWorker ServiceType = "transcription-worker" // Воркер транскрипции
	ServiceTypeSummarizationWorker ServiceType = "summarization-worker" // Воркер суммаризации
)

// ServiceInfo представляет информацию о сервисе
type ServiceInfo struct {
	// Уникальный идентификатор сервиса
	ID          string        `json:"id"`
	// Название сервиса
	Name        string        `json:"name"`
	// Тип сервиса
	Type        ServiceType   `json:"type"`
	// Текущий статус сервиса
	Status      ServiceStatus `json:"status"`
	// Версия сервиса
	Version     string        `json:"version"`
	// Время запуска сервиса
	StartedAt   time.Time     `json:"started_at"`
	// Время последнего сигнала жизни
	LastHeartbeat time.Time   `json:"last_heartbeat"`
	// Дополнительные метаданные сервиса
	Metadata    map[string]interface{} `json:"metadata"`
}

// SystemStatus представляет общий статус системы
type SystemStatus struct {
	// Общий статус системы
	Status       string                  `json:"status"`
	// Информация о сервисах
	Services     map[string]ServiceInfo  `json:"services"`
	// Информация об инфраструктуре
	Infrastructure map[string]interface{} `json:"infrastructure"`
	// Время получения статуса
	Timestamp    time.Time               `json:"timestamp"`
}

// HealthResponse представляет ответ на проверку здоровья
type HealthResponse struct {
	// Статус здоровья
	Status    string    `json:"status"`
	// Время проверки
	Timestamp time.Time `json:"timestamp"`
	// Версия сервиса
	Version   string    `json:"version"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	// Код ошибки
	Error   string    `json:"error"`
	// Сообщение об ошибке
	Message string    `json:"message"`
	// HTTP код ошибки
	Code    int       `json:"code"`
	// Время возникновения ошибки
	Timestamp time.Time `json:"timestamp"`
}
