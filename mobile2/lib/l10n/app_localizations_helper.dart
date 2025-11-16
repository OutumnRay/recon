import 'package:flutter/material.dart';

class AppLocalizations {
  final Locale locale;

  AppLocalizations(this.locale);

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  bool get isRussian => locale.languageCode == 'ru';

  // App
  String get appTitle => isRussian ? 'Реконтекст' : 'Recontext';

  // Auth
  String get login => isRussian ? 'Вход' : 'Login';
  String get email => 'Email';
  String get password => isRussian ? 'Пароль' : 'Password';
  String get username => isRussian ? 'Имя пользователя' : 'Username';
  String get logout => isRussian ? 'Выход' : 'Logout';
  String get loginButton => isRussian ? 'Войти' : 'Login';
  String get loggingIn => isRussian ? 'Вход...' : 'Logging in...';
  String get loginError =>
      isRussian ? 'Ошибка входа. Проверьте учетные данные.' : 'Login failed. Please check your credentials.';

  // Navigation
  String get meetings => isRussian ? 'Встречи' : 'Meetings';
  String get documents => isRussian ? 'Документы' : 'Documents';
  String get search => isRussian ? 'Поиск' : 'Search';
  String get settings => isRussian ? 'Настройки' : 'Settings';

  // Meeting Form
  String get newMeeting => isRussian ? 'Новая встреча' : 'New Meeting';
  String get createMeeting => isRussian ? 'Создать встречу' : 'Create Meeting';
  String get meetingTitle => isRussian ? 'Название встречи' : 'Meeting Title';
  String get meetingSubject => isRussian ? 'Тема' : 'Subject';
  String get meetingType => isRussian ? 'Тип' : 'Type';
  String get meetingDate => isRussian ? 'Дата' : 'Date';
  String get meetingTime => isRussian ? 'Время' : 'Time';
  String get meetingDuration => isRussian ? 'Длительность' : 'Duration';
  String get meetingRecurrence => isRussian ? 'Повторение' : 'Recurrence';
  String get meetingSpeaker => isRussian ? 'Спикер' : 'Speaker';
  String get meetingParticipants => isRussian ? 'Участники' : 'Participants';
  String get meetingDepartments => isRussian ? 'Отделы' : 'Departments';
  String get meetingNotes =>
      isRussian ? 'Дополнительные заметки' : 'Additional Notes';
  String get meetingVideoRecord => isRussian ? 'Видеозапись' : 'Video Recording';
  String get meetingAudioRecord => isRussian ? 'Аудиозапись' : 'Audio Recording';
  String get recordingOptions => isRussian ? 'Настройки записи' : 'Recording Options';

  // Meeting Types
  String get typeConference => isRussian ? 'Конференция' : 'Conference';
  String get typePresentation => isRussian ? 'Презентация' : 'Presentation';
  String get typeTraining => isRussian ? 'Обучение' : 'Training';
  String get typeDiscussion => isRussian ? 'Обсуждение' : 'Discussion';

  // Recurrence
  String get recurrenceNone => isRussian ? 'Без повтора' : 'No repeat';
  String get recurrenceDaily => isRussian ? 'Ежедневно' : 'Daily';
  String get recurrenceWeekly => isRussian ? 'Еженедельно' : 'Weekly';
  String get recurrenceMonthly => isRussian ? 'Ежемесячно' : 'Monthly';

  // Duration
  String get duration15 => isRussian ? '15 минут' : '15 minutes';
  String get duration30 => isRussian ? '30 минут' : '30 minutes';
  String get duration45 => isRussian ? '45 минут' : '45 minutes';
  String get duration60 => isRussian ? '1 час' : '1 hour';
  String get duration90 => isRussian ? '1.5 часа' : '1.5 hours';
  String get duration120 => isRussian ? '2 часа' : '2 hours';
  String get duration180 => isRussian ? '3 часа' : '3 hours';

  // Status
  String get statusScheduled => isRussian ? 'Запланирована' : 'Scheduled';
  String get statusInProgress => isRussian ? 'Идет' : 'In Progress';
  String get statusCompleted => isRussian ? 'Завершена' : 'Completed';
  String get statusCancelled => isRussian ? 'Отменена' : 'Cancelled';

  // Filters
  String get filterAll => isRussian ? 'Все' : 'All';
  String get filterScheduled => isRussian ? 'План' : 'Scheduled';
  String get filterInProgress => isRussian ? 'Идут' : 'In Progress';
  String get filterCompleted => isRussian ? 'Завершенные' : 'Completed';

  // Selection
  String get selectParticipants =>
      isRussian ? 'Выбрать участников' : 'Select Participants';
  String get selectDepartments => isRussian ? 'Выбрать отделы' : 'Select Departments';
  String get selectSpeaker => isRussian ? 'Выбрать спикера' : 'Select Speaker';
  String get noSpeaker => isRussian ? 'Без спикера' : 'No speaker';
  String get searchUsers => isRussian ? 'Поиск пользователей...' : 'Search users...';

  String participantsSelected(int count) =>
      isRussian ? 'Выбрано: $count' : '$count selected';
  String departmentsSelected(int count) =>
      isRussian ? 'Выбрано: $count' : '$count selected';

  // Time
  String get today => isRussian ? 'Сегодня' : 'Today';
  String get tomorrow => isRussian ? 'Завтра' : 'Tomorrow';
  String get minutes => isRussian ? 'мин' : 'min';

  // Common
  String get required => '*';
  String get optional => isRussian ? '(необязательно)' : '(optional)';
  String get cancel => isRussian ? 'Отмена' : 'Cancel';
  String get save => isRussian ? 'Сохранить' : 'Save';
  String get delete => isRussian ? 'Удалить' : 'Delete';
  String get edit => isRussian ? 'Редактировать' : 'Edit';
  String get refresh => isRussian ? 'Обновить' : 'Refresh';
  String get retry => isRussian ? 'Повторить' : 'Retry';
  String get close => isRussian ? 'Закрыть' : 'Close';
  String get select => isRussian ? 'Выбрать' : 'Select';

  // Status Messages
  String get loading => isRussian ? 'Загрузка...' : 'Loading...';
  String get error => isRussian ? 'Ошибка' : 'Error';
  String get success => isRussian ? 'Успешно' : 'Success';

  // Meeting Messages
  String get noMeetingsFound => isRussian ? 'Встречи не найдены' : 'No meetings found';
  String get createFirstMeeting =>
      isRussian ? 'Создайте первую встречу' : 'Create your first meeting';
  String get tryChangingFilter =>
      isRussian ? 'Попробуйте изменить фильтр' : 'Try changing the filter';

  String get meetingCreatedSuccess =>
      isRussian ? 'Встреча успешно создана' : 'Meeting created successfully';
  String get meetingCreatedError =>
      isRussian ? 'Не удалось создать встречу' : 'Failed to create meeting';
  String get meetingUpdatedSuccess =>
      isRussian ? 'Встреча успешно обновлена' : 'Meeting updated successfully';
  String get meetingDeletedSuccess =>
      isRussian ? 'Встреча успешно удалена' : 'Meeting deleted successfully';

  String get failedToLoadMeetings =>
      isRussian ? 'Не удалось загрузить встречи' : 'Failed to load meetings';
  String get failedToLoadFormData =>
      isRussian ? 'Не удалось загрузить данные формы' : 'Failed to load form data';
  String get pleaseEnterTitle => isRussian ? 'Введите название' : 'Please enter a title';
  String get pleaseSelectSubject => isRussian ? 'Выберите тему' : 'Please select a subject';

  // Error Messages
  String get apiError => isRussian ? 'Ошибка API' : 'API Error';
  String get connectionError => isRussian ? 'Ошибка подключения' : 'Connection Error';
  String get unexpectedError => isRussian ? 'Неожиданная ошибка' : 'Unexpected Error';
  String get checkNetwork => isRussian
      ? 'Пожалуйста, проверьте:\n• Подключение к сети\n• Сервер API запущен\n• Вы вошли в систему'
      : 'Please check:\n• Network connection\n• API server is running\n• You are logged in';

  // Settings
  String get language => isRussian ? 'Язык' : 'Language';
  String get changeLanguage => isRussian ? 'Изменить язык' : 'Change Language';
  String get english => isRussian ? 'Английский' : 'English';
  String get russian => isRussian ? 'Русский' : 'Russian';

  String get serverUrl => isRussian ? 'URL сервера' : 'Server URL';
  String get videoServerUrl => isRussian ? 'URL видеосервера' : 'Video Server URL';
  String get about => isRussian ? 'О приложении' : 'About';
  String get version => isRussian ? 'Версия' : 'Version';

  String get joinMeeting => isRussian ? 'Присоединиться к встрече' : 'Join Meeting';
  String get leaveMeeting => isRussian ? 'Покинуть встречу' : 'Leave Meeting';
  String get meetingDetails => isRussian ? 'Детали встречи' : 'Meeting Details';

  String get enterTitle => isRussian ? 'Введите название встречи' : 'Enter meeting title';
  String get enterNotes =>
      isRussian ? 'Введите дополнительную информацию' : 'Enter any additional information';
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  bool isSupported(Locale locale) =>
      ['en', 'ru'].contains(locale.languageCode);

  @override
  Future<AppLocalizations> load(Locale locale) async {
    return AppLocalizations(locale);
  }

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}
