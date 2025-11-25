// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Russian (`ru`).
class AppLocalizationsRu extends AppLocalizations {
  AppLocalizationsRu([String locale = 'ru']) : super(locale);

  @override
  String get appTitle => 'Реконтекст';

  @override
  String get login => 'Вход';

  @override
  String get loginSubtitle =>
      'Используйте учетные данные рабочего пространства';

  @override
  String get email => 'Email';

  @override
  String get password => 'Пароль';

  @override
  String get username => 'Имя пользователя';

  @override
  String get logout => 'Выход';

  @override
  String get loginButton => 'Войти';

  @override
  String get loggingIn => 'Вход...';

  @override
  String get loginError => 'Ошибка входа. Проверьте учетные данные.';

  @override
  String get meetings => 'Встречи';

  @override
  String get documents => 'Документы';

  @override
  String get search => 'Поиск';

  @override
  String get settings => 'Настройки';

  @override
  String get newMeeting => 'Новая встреча';

  @override
  String get createMeeting => 'Создать встречу';

  @override
  String get meetingTitle => 'Название встречи';

  @override
  String get meetingSubject => 'Тема';

  @override
  String get meetingType => 'Тип';

  @override
  String get meetingDate => 'Дата';

  @override
  String get meetingTime => 'Время';

  @override
  String get meetingDuration => 'Длительность';

  @override
  String get meetingRecurrence => 'Повторение';

  @override
  String get meetingSpeaker => 'Спикер';

  @override
  String get meetingParticipants => 'Участники';

  @override
  String get meetingDepartments => 'Отделы';

  @override
  String get meetingNotes => 'Дополнительные заметки';

  @override
  String get meetingRecord => 'Запись встречи (Аудио и Видео)';

  @override
  String get meetingTranscription => 'Транскрипция (запись отдельных дорожек)';

  @override
  String get meetingForceEndAtDuration =>
      'Принудительное завершение по истечении времени';

  @override
  String get meetingAllowAnonymous => 'Разрешить анонимных участников';

  @override
  String get recordingOptions => 'Настройки записи';

  @override
  String get typeConference => 'Конференция';

  @override
  String get typePresentation => 'Презентация';

  @override
  String get typeTraining => 'Обучение';

  @override
  String get typeDiscussion => 'Обсуждение';

  @override
  String get recurrenceNone => 'Без повтора';

  @override
  String get recurrenceDaily => 'Ежедневно';

  @override
  String get recurrenceWeekly => 'Еженедельно';

  @override
  String get recurrenceMonthly => 'Ежемесячно';

  @override
  String get recurrencePermanent => 'Постоянная';

  @override
  String get permanent => 'Постоянная';

  @override
  String get duration15 => '15 минут';

  @override
  String get duration30 => '30 минут';

  @override
  String get duration45 => '45 минут';

  @override
  String get duration60 => '1 час';

  @override
  String get duration90 => '1.5 часа';

  @override
  String get duration120 => '2 часа';

  @override
  String get duration180 => '3 часа';

  @override
  String get statusScheduled => 'Запланирована';

  @override
  String get statusInProgress => 'Идет';

  @override
  String get statusCompleted => 'Завершена';

  @override
  String get statusCancelled => 'Отменена';

  @override
  String get statusFinished => 'Завершена';

  @override
  String get statusRecording => 'Запись';

  @override
  String get filterAll => 'Все';

  @override
  String get filterScheduled => 'План';

  @override
  String get filterInProgress => 'Идут';

  @override
  String get filterPermanent => 'Постоянные';

  @override
  String get filterCompleted => 'Завершенные';

  @override
  String get filterCancelled => 'Отмененные';

  @override
  String get searchMeetings => 'Поиск встреч...';

  @override
  String get selectParticipants => 'Выбрать участников';

  @override
  String get selectDepartments => 'Выбрать отделы';

  @override
  String get selectSpeaker => 'Выбрать спикера';

  @override
  String get noSpeaker => 'Без спикера';

  @override
  String get searchUsers => 'Поиск пользователей...';

  @override
  String participantsSelected(int count) {
    return 'Выбрано: $count';
  }

  @override
  String departmentsSelected(int count) {
    return 'Выбрано: $count';
  }

  @override
  String get today => 'Сегодня';

  @override
  String get tomorrow => 'Завтра';

  @override
  String get minutes => 'мин';

  @override
  String get required => '*';

  @override
  String get optional => '(необязательно)';

  @override
  String get cancel => 'Отмена';

  @override
  String get save => 'Сохранить';

  @override
  String get delete => 'Удалить';

  @override
  String get edit => 'Редактировать';

  @override
  String get refresh => 'Обновить';

  @override
  String get retry => 'Повторить';

  @override
  String get close => 'Закрыть';

  @override
  String get select => 'Выбрать';

  @override
  String get confirm => 'Подтвердить';

  @override
  String get loading => 'Загрузка...';

  @override
  String get error => 'Ошибка';

  @override
  String get success => 'Успешно';

  @override
  String get noMeetingsFound => 'Встречи не найдены';

  @override
  String get createFirstMeeting => 'Создайте первую встречу';

  @override
  String get tryChangingFilter => 'Попробуйте изменить фильтр';

  @override
  String get meetingCreatedSuccess => 'Встреча успешно создана';

  @override
  String get meetingCreatedError => 'Не удалось создать встречу';

  @override
  String get meetingUpdatedSuccess => 'Встреча успешно обновлена';

  @override
  String get meetingDeletedSuccess => 'Встреча успешно удалена';

  @override
  String get meetingCancelledSuccess => 'Встреча успешно отменена';

  @override
  String get failedToCancelMeeting => 'Не удалось отменить встречу';

  @override
  String get cancelMeetingTitle => 'Отменить встречу';

  @override
  String get cancelMeetingConfirm =>
      'Вы уверены, что хотите отменить эту встречу? Это действие нельзя отменить.';

  @override
  String get failedToLoadMeetings => 'Не удалось загрузить встречи';

  @override
  String get failedToLoadFormData => 'Не удалось загрузить данные формы';

  @override
  String get pleaseEnterTitle => 'Введите название';

  @override
  String get pleaseSelectSubject => 'Выберите тему';

  @override
  String get apiError => 'Ошибка API';

  @override
  String get connectionError => 'Ошибка подключения';

  @override
  String get unexpectedError => 'Неожиданная ошибка';

  @override
  String get checkNetwork =>
      'Пожалуйста, проверьте:\n• Подключение к сети\n• Сервер API запущен\n• Вы вошли в систему';

  @override
  String get language => 'Язык';

  @override
  String get changeLanguage => 'Изменить язык';

  @override
  String get english => 'Английский';

  @override
  String get russian => 'Русский';

  @override
  String get serverUrl => 'URL сервера';

  @override
  String get videoServerUrl => 'URL видеосервера';

  @override
  String get about => 'О приложении';

  @override
  String get version => 'Версия';

  @override
  String get joinMeeting => 'Присоединиться к встрече';

  @override
  String get leaveMeeting => 'Покинуть встречу';

  @override
  String get meetingDetails => 'Детали встречи';

  @override
  String get enterTitle => 'Введите название встречи';

  @override
  String get enterNotes => 'Введите дополнительную информацию';

  @override
  String get documentsTitle => 'Документы';

  @override
  String get documentsSubtitle =>
      'Просматривайте, фильтруйте и скачивайте файлы встреч';

  @override
  String get noDocumentsFound => 'Документы не найдены';

  @override
  String get uploadFirstDocument => 'Загрузите первый документ';

  @override
  String get upload => 'Загрузить';

  @override
  String get download => 'Скачать';

  @override
  String get share => 'Поделиться';

  @override
  String get details => 'Подробности';

  @override
  String get documentDetails => 'Информация о документе';

  @override
  String get name => 'Название';

  @override
  String get type => 'Тип';

  @override
  String get size => 'Размер';

  @override
  String get uploaded => 'Загружен';

  @override
  String get uploadedBy => 'Загрузил';

  @override
  String get deleteDocument => 'Удалить документ';

  @override
  String deleteDocumentConfirm(String name) {
    return 'Вы уверены, что хотите удалить \"$name\"?';
  }

  @override
  String deleted(String name) {
    return 'Удалено $name';
  }

  @override
  String get undo => 'Отменить';

  @override
  String downloading(String name) {
    return 'Загрузка...';
  }

  @override
  String get filterVideo => 'Видео';

  @override
  String get filterAudio => 'Аудио';

  @override
  String get filterTranscripts => 'Транскрипты';

  @override
  String get filterOther => 'Другое';

  @override
  String get filtersTitle => 'Фильтры';

  @override
  String get uploadDocumentComingSoon => 'Загрузка документов - скоро';

  @override
  String get failedToLoadDocuments => 'Не удалось загрузить документы';

  @override
  String get searchTitle => 'Поиск';

  @override
  String get searchHint => 'Поиск встреч, транскриптов, документов...';

  @override
  String get semanticSearch => 'Семантический поиск';

  @override
  String get searchDescription =>
      'Ищите по встречам, транскриптам и документам используя естественный язык';

  @override
  String get searchResultsLabel => 'Результаты';

  @override
  String get trySearching => 'Попробуйте найти:';

  @override
  String get budgetDiscussion => 'обсуждение бюджета';

  @override
  String get projectDeadlines => 'сроки проекта';

  @override
  String get whoTalkedAbout => 'кто говорил о...';

  @override
  String get noResultsFound => 'Результаты не найдены';

  @override
  String get tryDifferentKeywords => 'Попробуйте другие ключевые слова';

  @override
  String get yesterday => 'Вчера';

  @override
  String daysAgo(int count) {
    return '$count дн. назад';
  }

  @override
  String get searchButton => 'Искать';

  @override
  String get match => 'совпадение';

  @override
  String get meeting => 'Встреча';

  @override
  String get transcript => 'Транскрипт';

  @override
  String get document => 'Документ';

  @override
  String get searchFailed => 'Ошибка поиска';

  @override
  String get meetingDetailsTitle => 'Детали встречи';

  @override
  String get dateAndTime => 'Дата и время';

  @override
  String get participants => 'Участники';

  @override
  String get noParticipants => 'Нет участников';

  @override
  String get departments => 'Отделы';

  @override
  String get settingsTitle => 'Настройки';

  @override
  String get videoRecording => 'Видеозапись';

  @override
  String get audioRecording => 'Аудиозапись';

  @override
  String get transcription => 'Транскрипция';

  @override
  String get roomId => 'ID комнаты';

  @override
  String get notes => 'Заметки';

  @override
  String get metadata => 'Метаданные';

  @override
  String get created => 'Создано';

  @override
  String get updated => 'Обновлено';

  @override
  String get createdBy => 'Создал';

  @override
  String get active => 'Активен';

  @override
  String get scheduled => 'Запланирована';

  @override
  String get completed => 'Завершена';

  @override
  String get cancelled => 'Отменена';

  @override
  String get enabled => 'Включено';

  @override
  String get disabled => 'Выключено';

  @override
  String get connectingToMeeting => 'Подключение к встрече...';

  @override
  String get connectionFailed => 'Ошибка подключения';

  @override
  String get failedToJoin => 'Не удалось присоединиться';

  @override
  String get waitingForParticipants => 'Ожидание участников...';

  @override
  String participantsInRoom(int count) {
    return 'Участников в комнате: $count';
  }

  @override
  String get you => 'Вы';

  @override
  String get join => 'Присоединиться';

  @override
  String get mic => 'Микрофон';

  @override
  String get camera => 'Камера';

  @override
  String get leave => 'Выйти';

  @override
  String get switchCamera => 'Сменить камеру';

  @override
  String get flipCamera => 'Сменить';

  @override
  String get userRoleUser => 'Пользователь';

  @override
  String get userRoleAdmin => 'Администратор';

  @override
  String get userRoleManager => 'Менеджер';

  @override
  String get unknownUser => 'Неизвестный пользователь';

  @override
  String get serverConfiguration => 'Настройка сервера';

  @override
  String get apiUrl => 'API URL';

  @override
  String get defaultButton => 'По умолчанию';

  @override
  String get restartNote =>
      'Примечание: После изменения URL API необходимо перезапустить приложение.';

  @override
  String get apiUrlEmpty => 'API URL не может быть пустым';

  @override
  String get apiUrlSaved => 'API URL сохранен';

  @override
  String get notifications => 'Уведомления';

  @override
  String get privacySecurity => 'Приватность и безопасность';

  @override
  String get helpSupport => 'Помощь и поддержка';

  @override
  String get termsConditions => 'Условия использования';

  @override
  String get comingSoon => 'Скоро';

  @override
  String get notificationsComingSoon => 'Настройки уведомлений - скоро';

  @override
  String get privacyComingSoon => 'Настройки приватности - скоро';

  @override
  String get helpComingSoon => 'Помощь и поддержка - скоро';

  @override
  String get termsComingSoon => 'Условия использования - скоро';

  @override
  String get logoutConfirm => 'Вы уверены, что хотите выйти?';

  @override
  String get appDescription =>
      'Платформа для видеоконференций и управления встречами';

  @override
  String get allRightsReserved => '© 2025 Recontext. Все права защищены.';

  @override
  String get tabInfo => 'Информация';

  @override
  String get tabRecordings => 'Записи';

  @override
  String get tabTasks => 'Задачи';

  @override
  String get recordings => 'Записи';

  @override
  String get noRecordingsFound => 'Записи не найдены';

  @override
  String get recordingsHint => 'Записи появятся здесь после завершения встречи';

  @override
  String get session => 'Сессия';

  @override
  String sessionNumber(int number) {
    return 'Сессия $number';
  }

  @override
  String recordingDuration(int minutes) {
    return 'Длительность: $minutes мин';
  }

  @override
  String get playRecording => 'Воспроизвести запись';

  @override
  String get roomRecording => 'Запись комнаты';

  @override
  String get recordingStatus => 'Статус';

  @override
  String get loadingRecordings => 'Загрузка записей...';

  @override
  String get failedToLoadRecordings => 'Не удалось загрузить записи';

  @override
  String get retryLoadRecordings => 'Повторить';

  @override
  String get roomDisconnected => 'Вы были отключены';

  @override
  String get disconnectedMessage =>
      'Вы были отключены от встречи. Возможно, вы вошли с другого устройства.';

  @override
  String get goToHome => 'Вернуться на главную';

  @override
  String get confirmLeaveTitle => 'Покинуть встречу?';

  @override
  String get confirmLeaveMessage =>
      'Вы уверены, что хотите покинуть эту встречу?';

  @override
  String get confirmLeave => 'Выйти';

  @override
  String get speaking => 'Говорит';

  @override
  String get downloadRecording => 'Скачать запись';

  @override
  String get downloadComplete => 'Загрузка завершена';

  @override
  String get downloadFailed => 'Ошибка загрузки';

  @override
  String get recordingSaved => 'Запись сохранена в Downloads';

  @override
  String get viewRecordings => 'Просмотр сессий';

  @override
  String get sessionStatusRecording => 'Запись';

  @override
  String get sessionStatusCompleted => 'Завершена';

  @override
  String get sessionStatusProcessing => 'Обработка';

  @override
  String get sessionStatusFailed => 'Ошибка';

  @override
  String get sessionStatusFinished => 'Завершена';

  @override
  String get viewSession => 'Открыть сессию';

  @override
  String get viewSessionDetails =>
      'Просмотр деталей сессии, треков, транскрипции и саммари';

  @override
  String onlineCount(int count) {
    return '$count онлайн';
  }

  @override
  String anonymousGuestsCount(int count) {
    return '$count гостей';
  }

  @override
  String participantsSummary(int total, int online) {
    return '$total участников ($online онлайн)';
  }

  @override
  String get anonymousLinkLabel => 'Ссылка для анонимного входа';

  @override
  String get copyLink => 'Копировать';

  @override
  String get linkCopied => 'Скопировано!';

  @override
  String get anonymousLinkCopied =>
      'Анонимная ссылка скопирована в буфер обмена';

  @override
  String get allowsAnonymousJoin => 'Разрешен анонимный вход';

  @override
  String get audioOnlyRecording => 'Только аудио';

  @override
  String get failedToLoadRecording => 'Не удалось загрузить запись';

  @override
  String get unknownError => 'Неизвестная ошибка';

  @override
  String get welcomeBackTitle => 'С возвращением';

  @override
  String get welcomeBackSubtitle => 'Войдите, чтобы управлять встречами';

  @override
  String get validationUsernameRequired => 'Введите имя пользователя';

  @override
  String get validationPasswordRequired => 'Введите пароль';

  @override
  String get apiUrlHint => 'https://portal.recontext.online';

  @override
  String get pleaseCheck => 'Пожалуйста, проверьте:';

  @override
  String get networkConnection => 'Подключение к сети';

  @override
  String get apiUrlCorrect => 'Правильность API URL';

  @override
  String get serverRunning => 'Сервер запущен';

  @override
  String get apiNotConfigured => 'URL сервера не настроен';

  @override
  String get currentServer => 'Подключено к';

  @override
  String get tabPlayer => 'Плеер';

  @override
  String get tabTranscript => 'Транскрипция';

  @override
  String get tabMemo => 'Саммари';

  @override
  String get noTranscriptAvailable => 'Транскрипция недоступна';

  @override
  String get transcriptionNotGenerated =>
      'Транскрипция для этой записи еще не создана.';

  @override
  String get noMemoAvailable => 'Саммари недоступно';

  @override
  String get memoNotGenerated => 'AI-саммари для этой записи еще не создано.';

  @override
  String get aiSummary => 'AI Саммари';

  @override
  String get copyMemo => 'Копировать саммари';

  @override
  String get memoCopied => 'Саммари скопировано в буфер обмена';

  @override
  String get failedToLoadTranscript => 'Не удалось загрузить транскрипцию';

  @override
  String get failedToLoadMemo => 'Не удалось загрузить саммари';

  @override
  String get myProfile => 'Мой профиль';

  @override
  String get editProfileInfo => 'Редактировать информацию профиля';

  @override
  String get profile => 'Профиль';

  @override
  String get accountInformation => 'Информация об аккаунте';

  @override
  String get personalInformation => 'Личная информация';

  @override
  String get preferences => 'Настройки';

  @override
  String get firstName => 'Имя';

  @override
  String get lastName => 'Фамилия';

  @override
  String get phone => 'Телефон';

  @override
  String get bio => 'О себе';

  @override
  String get notificationPreferences => 'Настройки уведомлений';

  @override
  String get notifTracksOnly => 'Только отдельные треки';

  @override
  String get notifRoomsOnly => 'Только комнаты';

  @override
  String get notifBoth => 'Оба';

  @override
  String get failedToLoadProfile => 'Не удалось загрузить профиль';

  @override
  String get failedToSelectImage => 'Не удалось выбрать изображение';

  @override
  String get profileUpdatedSuccessfully => 'Профиль успешно обновлен';

  @override
  String get failedToSaveProfile => 'Не удалось сохранить профиль';

  @override
  String get viewCompositeVideo => 'Просмотр составного видео';

  @override
  String get processingVideo => 'Обработка...';

  @override
  String get processingVideoDescription =>
      'Объединение видео дорожек и создание составного видео';

  @override
  String get recordingInProgress => 'Идет запись';

  @override
  String get recordingInProgressDescription =>
      'Составное видео будет доступно после завершения записи';

  @override
  String get compositeVideoNotAvailable => 'Составное видео недоступно';

  @override
  String get viewIndividualTracks => 'Просмотрите отдельные дорожки участников';

  @override
  String get nowSpeaking => 'Сейчас говорит';

  @override
  String get speaker => 'Докладчик';

  @override
  String get participant => 'Участник';
}
