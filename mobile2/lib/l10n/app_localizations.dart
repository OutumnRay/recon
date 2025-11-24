import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_ru.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('ru'),
  ];

  /// Application title
  ///
  /// In en, this message translates to:
  /// **'Recontext'**
  String get appTitle;

  /// No description provided for @login.
  ///
  /// In en, this message translates to:
  /// **'Login'**
  String get login;

  /// No description provided for @loginSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Use your workspace credentials to continue'**
  String get loginSubtitle;

  /// No description provided for @email.
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get email;

  /// No description provided for @password.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get password;

  /// No description provided for @username.
  ///
  /// In en, this message translates to:
  /// **'Username'**
  String get username;

  /// No description provided for @logout.
  ///
  /// In en, this message translates to:
  /// **'Logout'**
  String get logout;

  /// No description provided for @loginButton.
  ///
  /// In en, this message translates to:
  /// **'Login'**
  String get loginButton;

  /// No description provided for @loggingIn.
  ///
  /// In en, this message translates to:
  /// **'Logging in...'**
  String get loggingIn;

  /// No description provided for @loginError.
  ///
  /// In en, this message translates to:
  /// **'Login failed. Please check your credentials.'**
  String get loginError;

  /// No description provided for @meetings.
  ///
  /// In en, this message translates to:
  /// **'Meetings'**
  String get meetings;

  /// No description provided for @documents.
  ///
  /// In en, this message translates to:
  /// **'Documents'**
  String get documents;

  /// No description provided for @search.
  ///
  /// In en, this message translates to:
  /// **'Search'**
  String get search;

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @newMeeting.
  ///
  /// In en, this message translates to:
  /// **'New Meeting'**
  String get newMeeting;

  /// No description provided for @createMeeting.
  ///
  /// In en, this message translates to:
  /// **'Create Meeting'**
  String get createMeeting;

  /// No description provided for @meetingTitle.
  ///
  /// In en, this message translates to:
  /// **'Meeting Title'**
  String get meetingTitle;

  /// No description provided for @meetingSubject.
  ///
  /// In en, this message translates to:
  /// **'Subject'**
  String get meetingSubject;

  /// No description provided for @meetingType.
  ///
  /// In en, this message translates to:
  /// **'Type'**
  String get meetingType;

  /// No description provided for @meetingDate.
  ///
  /// In en, this message translates to:
  /// **'Date'**
  String get meetingDate;

  /// No description provided for @meetingTime.
  ///
  /// In en, this message translates to:
  /// **'Time'**
  String get meetingTime;

  /// No description provided for @meetingDuration.
  ///
  /// In en, this message translates to:
  /// **'Duration'**
  String get meetingDuration;

  /// No description provided for @meetingRecurrence.
  ///
  /// In en, this message translates to:
  /// **'Recurrence'**
  String get meetingRecurrence;

  /// No description provided for @meetingSpeaker.
  ///
  /// In en, this message translates to:
  /// **'Speaker'**
  String get meetingSpeaker;

  /// No description provided for @meetingParticipants.
  ///
  /// In en, this message translates to:
  /// **'Participants'**
  String get meetingParticipants;

  /// No description provided for @meetingDepartments.
  ///
  /// In en, this message translates to:
  /// **'Departments'**
  String get meetingDepartments;

  /// No description provided for @meetingNotes.
  ///
  /// In en, this message translates to:
  /// **'Additional Notes'**
  String get meetingNotes;

  /// No description provided for @meetingRecord.
  ///
  /// In en, this message translates to:
  /// **'Record Meeting (Audio & Video)'**
  String get meetingRecord;

  /// No description provided for @meetingTranscription.
  ///
  /// In en, this message translates to:
  /// **'Transcription (recording individual tracks)'**
  String get meetingTranscription;

  /// No description provided for @meetingForceEndAtDuration.
  ///
  /// In en, this message translates to:
  /// **'Force end after time elapses'**
  String get meetingForceEndAtDuration;

  /// No description provided for @meetingAllowAnonymous.
  ///
  /// In en, this message translates to:
  /// **'Allow anonymous members'**
  String get meetingAllowAnonymous;

  /// No description provided for @recordingOptions.
  ///
  /// In en, this message translates to:
  /// **'Recording Options'**
  String get recordingOptions;

  /// No description provided for @typeConference.
  ///
  /// In en, this message translates to:
  /// **'Conference'**
  String get typeConference;

  /// No description provided for @typePresentation.
  ///
  /// In en, this message translates to:
  /// **'Presentation'**
  String get typePresentation;

  /// No description provided for @typeTraining.
  ///
  /// In en, this message translates to:
  /// **'Training'**
  String get typeTraining;

  /// No description provided for @typeDiscussion.
  ///
  /// In en, this message translates to:
  /// **'Discussion'**
  String get typeDiscussion;

  /// No description provided for @recurrenceNone.
  ///
  /// In en, this message translates to:
  /// **'No repeat'**
  String get recurrenceNone;

  /// No description provided for @recurrenceDaily.
  ///
  /// In en, this message translates to:
  /// **'Daily'**
  String get recurrenceDaily;

  /// No description provided for @recurrenceWeekly.
  ///
  /// In en, this message translates to:
  /// **'Weekly'**
  String get recurrenceWeekly;

  /// No description provided for @recurrenceMonthly.
  ///
  /// In en, this message translates to:
  /// **'Monthly'**
  String get recurrenceMonthly;

  /// No description provided for @recurrencePermanent.
  ///
  /// In en, this message translates to:
  /// **'Permanent'**
  String get recurrencePermanent;

  /// No description provided for @permanent.
  ///
  /// In en, this message translates to:
  /// **'Permanent'**
  String get permanent;

  /// No description provided for @duration15.
  ///
  /// In en, this message translates to:
  /// **'15 minutes'**
  String get duration15;

  /// No description provided for @duration30.
  ///
  /// In en, this message translates to:
  /// **'30 minutes'**
  String get duration30;

  /// No description provided for @duration45.
  ///
  /// In en, this message translates to:
  /// **'45 minutes'**
  String get duration45;

  /// No description provided for @duration60.
  ///
  /// In en, this message translates to:
  /// **'1 hour'**
  String get duration60;

  /// No description provided for @duration90.
  ///
  /// In en, this message translates to:
  /// **'1.5 hours'**
  String get duration90;

  /// No description provided for @duration120.
  ///
  /// In en, this message translates to:
  /// **'2 hours'**
  String get duration120;

  /// No description provided for @duration180.
  ///
  /// In en, this message translates to:
  /// **'3 hours'**
  String get duration180;

  /// No description provided for @statusScheduled.
  ///
  /// In en, this message translates to:
  /// **'Scheduled'**
  String get statusScheduled;

  /// No description provided for @statusInProgress.
  ///
  /// In en, this message translates to:
  /// **'In Progress'**
  String get statusInProgress;

  /// No description provided for @statusCompleted.
  ///
  /// In en, this message translates to:
  /// **'Completed'**
  String get statusCompleted;

  /// No description provided for @statusCancelled.
  ///
  /// In en, this message translates to:
  /// **'Cancelled'**
  String get statusCancelled;

  /// No description provided for @filterAll.
  ///
  /// In en, this message translates to:
  /// **'All'**
  String get filterAll;

  /// No description provided for @filterScheduled.
  ///
  /// In en, this message translates to:
  /// **'Scheduled'**
  String get filterScheduled;

  /// No description provided for @filterInProgress.
  ///
  /// In en, this message translates to:
  /// **'In Progress'**
  String get filterInProgress;

  /// No description provided for @filterCompleted.
  ///
  /// In en, this message translates to:
  /// **'Completed'**
  String get filterCompleted;

  /// No description provided for @selectParticipants.
  ///
  /// In en, this message translates to:
  /// **'Select Participants'**
  String get selectParticipants;

  /// No description provided for @selectDepartments.
  ///
  /// In en, this message translates to:
  /// **'Select Departments'**
  String get selectDepartments;

  /// No description provided for @selectSpeaker.
  ///
  /// In en, this message translates to:
  /// **'Select Speaker'**
  String get selectSpeaker;

  /// No description provided for @noSpeaker.
  ///
  /// In en, this message translates to:
  /// **'No speaker'**
  String get noSpeaker;

  /// No description provided for @searchUsers.
  ///
  /// In en, this message translates to:
  /// **'Search users...'**
  String get searchUsers;

  /// Number of participants selected
  ///
  /// In en, this message translates to:
  /// **'{count} selected'**
  String participantsSelected(int count);

  /// Number of departments selected
  ///
  /// In en, this message translates to:
  /// **'{count} selected'**
  String departmentsSelected(int count);

  /// No description provided for @today.
  ///
  /// In en, this message translates to:
  /// **'Today'**
  String get today;

  /// No description provided for @tomorrow.
  ///
  /// In en, this message translates to:
  /// **'Tomorrow'**
  String get tomorrow;

  /// No description provided for @minutes.
  ///
  /// In en, this message translates to:
  /// **'min'**
  String get minutes;

  /// No description provided for @required.
  ///
  /// In en, this message translates to:
  /// **'*'**
  String get required;

  /// No description provided for @optional.
  ///
  /// In en, this message translates to:
  /// **'(optional)'**
  String get optional;

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @save.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get save;

  /// No description provided for @delete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get delete;

  /// No description provided for @edit.
  ///
  /// In en, this message translates to:
  /// **'Edit'**
  String get edit;

  /// No description provided for @refresh.
  ///
  /// In en, this message translates to:
  /// **'Refresh'**
  String get refresh;

  /// No description provided for @retry.
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get retry;

  /// No description provided for @close.
  ///
  /// In en, this message translates to:
  /// **'Close'**
  String get close;

  /// No description provided for @select.
  ///
  /// In en, this message translates to:
  /// **'Select'**
  String get select;

  /// No description provided for @loading.
  ///
  /// In en, this message translates to:
  /// **'Loading...'**
  String get loading;

  /// No description provided for @error.
  ///
  /// In en, this message translates to:
  /// **'Error'**
  String get error;

  /// No description provided for @success.
  ///
  /// In en, this message translates to:
  /// **'Success'**
  String get success;

  /// No description provided for @noMeetingsFound.
  ///
  /// In en, this message translates to:
  /// **'No meetings found'**
  String get noMeetingsFound;

  /// No description provided for @createFirstMeeting.
  ///
  /// In en, this message translates to:
  /// **'Create your first meeting'**
  String get createFirstMeeting;

  /// No description provided for @tryChangingFilter.
  ///
  /// In en, this message translates to:
  /// **'Try changing the filter'**
  String get tryChangingFilter;

  /// No description provided for @meetingCreatedSuccess.
  ///
  /// In en, this message translates to:
  /// **'Meeting created successfully'**
  String get meetingCreatedSuccess;

  /// No description provided for @meetingCreatedError.
  ///
  /// In en, this message translates to:
  /// **'Failed to create meeting'**
  String get meetingCreatedError;

  /// No description provided for @meetingUpdatedSuccess.
  ///
  /// In en, this message translates to:
  /// **'Meeting updated successfully'**
  String get meetingUpdatedSuccess;

  /// No description provided for @meetingDeletedSuccess.
  ///
  /// In en, this message translates to:
  /// **'Meeting deleted successfully'**
  String get meetingDeletedSuccess;

  /// No description provided for @failedToLoadMeetings.
  ///
  /// In en, this message translates to:
  /// **'Failed to load meetings'**
  String get failedToLoadMeetings;

  /// No description provided for @failedToLoadFormData.
  ///
  /// In en, this message translates to:
  /// **'Failed to load form data'**
  String get failedToLoadFormData;

  /// No description provided for @pleaseEnterTitle.
  ///
  /// In en, this message translates to:
  /// **'Please enter a title'**
  String get pleaseEnterTitle;

  /// No description provided for @pleaseSelectSubject.
  ///
  /// In en, this message translates to:
  /// **'Please select a subject'**
  String get pleaseSelectSubject;

  /// No description provided for @apiError.
  ///
  /// In en, this message translates to:
  /// **'API Error'**
  String get apiError;

  /// No description provided for @connectionError.
  ///
  /// In en, this message translates to:
  /// **'Connection Error'**
  String get connectionError;

  /// No description provided for @unexpectedError.
  ///
  /// In en, this message translates to:
  /// **'Unexpected Error'**
  String get unexpectedError;

  /// No description provided for @checkNetwork.
  ///
  /// In en, this message translates to:
  /// **'Please check:\n• Network connection\n• API server is running\n• You are logged in'**
  String get checkNetwork;

  /// No description provided for @language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;

  /// No description provided for @changeLanguage.
  ///
  /// In en, this message translates to:
  /// **'Change Language'**
  String get changeLanguage;

  /// No description provided for @english.
  ///
  /// In en, this message translates to:
  /// **'English'**
  String get english;

  /// No description provided for @russian.
  ///
  /// In en, this message translates to:
  /// **'Russian'**
  String get russian;

  /// No description provided for @serverUrl.
  ///
  /// In en, this message translates to:
  /// **'Server URL'**
  String get serverUrl;

  /// No description provided for @videoServerUrl.
  ///
  /// In en, this message translates to:
  /// **'Video Server URL'**
  String get videoServerUrl;

  /// No description provided for @about.
  ///
  /// In en, this message translates to:
  /// **'About'**
  String get about;

  /// No description provided for @version.
  ///
  /// In en, this message translates to:
  /// **'Version'**
  String get version;

  /// No description provided for @joinMeeting.
  ///
  /// In en, this message translates to:
  /// **'Join Meeting'**
  String get joinMeeting;

  /// No description provided for @leaveMeeting.
  ///
  /// In en, this message translates to:
  /// **'Leave Meeting'**
  String get leaveMeeting;

  /// No description provided for @meetingDetails.
  ///
  /// In en, this message translates to:
  /// **'Meeting Details'**
  String get meetingDetails;

  /// No description provided for @enterTitle.
  ///
  /// In en, this message translates to:
  /// **'Enter meeting title'**
  String get enterTitle;

  /// No description provided for @enterNotes.
  ///
  /// In en, this message translates to:
  /// **'Enter any additional information'**
  String get enterNotes;

  /// No description provided for @documentsTitle.
  ///
  /// In en, this message translates to:
  /// **'Documents'**
  String get documentsTitle;

  /// No description provided for @documentsSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Browse, filter, and download your meeting files'**
  String get documentsSubtitle;

  /// No description provided for @noDocumentsFound.
  ///
  /// In en, this message translates to:
  /// **'No documents found'**
  String get noDocumentsFound;

  /// No description provided for @uploadFirstDocument.
  ///
  /// In en, this message translates to:
  /// **'Upload your first document'**
  String get uploadFirstDocument;

  /// No description provided for @upload.
  ///
  /// In en, this message translates to:
  /// **'Upload'**
  String get upload;

  /// No description provided for @download.
  ///
  /// In en, this message translates to:
  /// **'Download'**
  String get download;

  /// No description provided for @share.
  ///
  /// In en, this message translates to:
  /// **'Share'**
  String get share;

  /// No description provided for @details.
  ///
  /// In en, this message translates to:
  /// **'Details'**
  String get details;

  /// No description provided for @documentDetails.
  ///
  /// In en, this message translates to:
  /// **'Document Details'**
  String get documentDetails;

  /// No description provided for @name.
  ///
  /// In en, this message translates to:
  /// **'Name'**
  String get name;

  /// No description provided for @type.
  ///
  /// In en, this message translates to:
  /// **'Type'**
  String get type;

  /// No description provided for @size.
  ///
  /// In en, this message translates to:
  /// **'Size'**
  String get size;

  /// No description provided for @uploaded.
  ///
  /// In en, this message translates to:
  /// **'Uploaded'**
  String get uploaded;

  /// No description provided for @uploadedBy.
  ///
  /// In en, this message translates to:
  /// **'Uploaded by'**
  String get uploadedBy;

  /// No description provided for @deleteDocument.
  ///
  /// In en, this message translates to:
  /// **'Delete Document'**
  String get deleteDocument;

  /// Confirm document deletion
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to delete \"{name}\"?'**
  String deleteDocumentConfirm(String name);

  /// Document deleted message
  ///
  /// In en, this message translates to:
  /// **'Deleted {name}'**
  String deleted(String name);

  /// No description provided for @undo.
  ///
  /// In en, this message translates to:
  /// **'Undo'**
  String get undo;

  /// Downloading document
  ///
  /// In en, this message translates to:
  /// **'Downloading...'**
  String downloading(String name);

  /// No description provided for @filterVideo.
  ///
  /// In en, this message translates to:
  /// **'Video'**
  String get filterVideo;

  /// No description provided for @filterAudio.
  ///
  /// In en, this message translates to:
  /// **'Audio'**
  String get filterAudio;

  /// No description provided for @filterTranscripts.
  ///
  /// In en, this message translates to:
  /// **'Transcripts'**
  String get filterTranscripts;

  /// No description provided for @filterOther.
  ///
  /// In en, this message translates to:
  /// **'Other'**
  String get filterOther;

  /// No description provided for @filtersTitle.
  ///
  /// In en, this message translates to:
  /// **'Filters'**
  String get filtersTitle;

  /// No description provided for @uploadDocumentComingSoon.
  ///
  /// In en, this message translates to:
  /// **'Upload document - Coming soon'**
  String get uploadDocumentComingSoon;

  /// No description provided for @failedToLoadDocuments.
  ///
  /// In en, this message translates to:
  /// **'Failed to load documents'**
  String get failedToLoadDocuments;

  /// No description provided for @searchTitle.
  ///
  /// In en, this message translates to:
  /// **'Search'**
  String get searchTitle;

  /// No description provided for @searchHint.
  ///
  /// In en, this message translates to:
  /// **'Search meetings, transcripts, documents...'**
  String get searchHint;

  /// No description provided for @semanticSearch.
  ///
  /// In en, this message translates to:
  /// **'Semantic Search'**
  String get semanticSearch;

  /// No description provided for @searchDescription.
  ///
  /// In en, this message translates to:
  /// **'Search across meetings, transcripts, and documents using natural language'**
  String get searchDescription;

  /// No description provided for @searchResultsLabel.
  ///
  /// In en, this message translates to:
  /// **'Results'**
  String get searchResultsLabel;

  /// No description provided for @trySearching.
  ///
  /// In en, this message translates to:
  /// **'Try searching for:'**
  String get trySearching;

  /// No description provided for @budgetDiscussion.
  ///
  /// In en, this message translates to:
  /// **'budget discussion'**
  String get budgetDiscussion;

  /// No description provided for @projectDeadlines.
  ///
  /// In en, this message translates to:
  /// **'project deadlines'**
  String get projectDeadlines;

  /// No description provided for @whoTalkedAbout.
  ///
  /// In en, this message translates to:
  /// **'who talked about...'**
  String get whoTalkedAbout;

  /// No description provided for @noResultsFound.
  ///
  /// In en, this message translates to:
  /// **'No results found'**
  String get noResultsFound;

  /// No description provided for @tryDifferentKeywords.
  ///
  /// In en, this message translates to:
  /// **'Try different keywords'**
  String get tryDifferentKeywords;

  /// No description provided for @yesterday.
  ///
  /// In en, this message translates to:
  /// **'Yesterday'**
  String get yesterday;

  /// Relative day label
  ///
  /// In en, this message translates to:
  /// **'{count} days ago'**
  String daysAgo(int count);

  /// No description provided for @searchButton.
  ///
  /// In en, this message translates to:
  /// **'Search'**
  String get searchButton;

  /// No description provided for @match.
  ///
  /// In en, this message translates to:
  /// **'match'**
  String get match;

  /// No description provided for @meeting.
  ///
  /// In en, this message translates to:
  /// **'Meeting'**
  String get meeting;

  /// No description provided for @transcript.
  ///
  /// In en, this message translates to:
  /// **'Transcript'**
  String get transcript;

  /// No description provided for @document.
  ///
  /// In en, this message translates to:
  /// **'Document'**
  String get document;

  /// No description provided for @searchFailed.
  ///
  /// In en, this message translates to:
  /// **'Search failed'**
  String get searchFailed;

  /// No description provided for @meetingDetailsTitle.
  ///
  /// In en, this message translates to:
  /// **'Meeting Details'**
  String get meetingDetailsTitle;

  /// No description provided for @dateAndTime.
  ///
  /// In en, this message translates to:
  /// **'Date & Time'**
  String get dateAndTime;

  /// No description provided for @participants.
  ///
  /// In en, this message translates to:
  /// **'Participants'**
  String get participants;

  /// No description provided for @noParticipants.
  ///
  /// In en, this message translates to:
  /// **'No participants'**
  String get noParticipants;

  /// No description provided for @departments.
  ///
  /// In en, this message translates to:
  /// **'Departments'**
  String get departments;

  /// No description provided for @settingsTitle.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settingsTitle;

  /// No description provided for @videoRecording.
  ///
  /// In en, this message translates to:
  /// **'Video Recording'**
  String get videoRecording;

  /// No description provided for @audioRecording.
  ///
  /// In en, this message translates to:
  /// **'Audio Recording'**
  String get audioRecording;

  /// No description provided for @transcription.
  ///
  /// In en, this message translates to:
  /// **'Transcription'**
  String get transcription;

  /// No description provided for @roomId.
  ///
  /// In en, this message translates to:
  /// **'Room ID'**
  String get roomId;

  /// No description provided for @notes.
  ///
  /// In en, this message translates to:
  /// **'Notes'**
  String get notes;

  /// No description provided for @metadata.
  ///
  /// In en, this message translates to:
  /// **'Metadata'**
  String get metadata;

  /// No description provided for @created.
  ///
  /// In en, this message translates to:
  /// **'Created'**
  String get created;

  /// No description provided for @updated.
  ///
  /// In en, this message translates to:
  /// **'Updated'**
  String get updated;

  /// No description provided for @createdBy.
  ///
  /// In en, this message translates to:
  /// **'Created By'**
  String get createdBy;

  /// No description provided for @active.
  ///
  /// In en, this message translates to:
  /// **'Active'**
  String get active;

  /// No description provided for @scheduled.
  ///
  /// In en, this message translates to:
  /// **'Scheduled'**
  String get scheduled;

  /// No description provided for @completed.
  ///
  /// In en, this message translates to:
  /// **'Completed'**
  String get completed;

  /// No description provided for @cancelled.
  ///
  /// In en, this message translates to:
  /// **'Cancelled'**
  String get cancelled;

  /// No description provided for @enabled.
  ///
  /// In en, this message translates to:
  /// **'Enabled'**
  String get enabled;

  /// No description provided for @disabled.
  ///
  /// In en, this message translates to:
  /// **'Disabled'**
  String get disabled;

  /// No description provided for @connectingToMeeting.
  ///
  /// In en, this message translates to:
  /// **'Connecting to meeting...'**
  String get connectingToMeeting;

  /// No description provided for @connectionFailed.
  ///
  /// In en, this message translates to:
  /// **'Connection Failed'**
  String get connectionFailed;

  /// No description provided for @failedToJoin.
  ///
  /// In en, this message translates to:
  /// **'Failed to join'**
  String get failedToJoin;

  /// No description provided for @waitingForParticipants.
  ///
  /// In en, this message translates to:
  /// **'Waiting for participants...'**
  String get waitingForParticipants;

  /// Number of participants in room
  ///
  /// In en, this message translates to:
  /// **'{count} participant(s) in room'**
  String participantsInRoom(int count);

  /// No description provided for @you.
  ///
  /// In en, this message translates to:
  /// **'You'**
  String get you;

  /// No description provided for @join.
  ///
  /// In en, this message translates to:
  /// **'Join'**
  String get join;

  /// No description provided for @mic.
  ///
  /// In en, this message translates to:
  /// **'Mic'**
  String get mic;

  /// No description provided for @camera.
  ///
  /// In en, this message translates to:
  /// **'Camera'**
  String get camera;

  /// No description provided for @leave.
  ///
  /// In en, this message translates to:
  /// **'Leave'**
  String get leave;

  /// No description provided for @switchCamera.
  ///
  /// In en, this message translates to:
  /// **'Switch Camera'**
  String get switchCamera;

  /// No description provided for @flipCamera.
  ///
  /// In en, this message translates to:
  /// **'Flip'**
  String get flipCamera;

  /// No description provided for @userRoleUser.
  ///
  /// In en, this message translates to:
  /// **'User'**
  String get userRoleUser;

  /// No description provided for @userRoleAdmin.
  ///
  /// In en, this message translates to:
  /// **'Administrator'**
  String get userRoleAdmin;

  /// No description provided for @userRoleManager.
  ///
  /// In en, this message translates to:
  /// **'Manager'**
  String get userRoleManager;

  /// No description provided for @unknownUser.
  ///
  /// In en, this message translates to:
  /// **'Unknown User'**
  String get unknownUser;

  /// No description provided for @serverConfiguration.
  ///
  /// In en, this message translates to:
  /// **'Server Configuration'**
  String get serverConfiguration;

  /// No description provided for @apiUrl.
  ///
  /// In en, this message translates to:
  /// **'API URL'**
  String get apiUrl;

  /// No description provided for @defaultButton.
  ///
  /// In en, this message translates to:
  /// **'Default'**
  String get defaultButton;

  /// No description provided for @restartNote.
  ///
  /// In en, this message translates to:
  /// **'Note: You need to restart the app after changing the API URL.'**
  String get restartNote;

  /// No description provided for @apiUrlEmpty.
  ///
  /// In en, this message translates to:
  /// **'API URL cannot be empty'**
  String get apiUrlEmpty;

  /// No description provided for @apiUrlSaved.
  ///
  /// In en, this message translates to:
  /// **'API URL saved'**
  String get apiUrlSaved;

  /// No description provided for @notifications.
  ///
  /// In en, this message translates to:
  /// **'Notifications'**
  String get notifications;

  /// No description provided for @privacySecurity.
  ///
  /// In en, this message translates to:
  /// **'Privacy & Security'**
  String get privacySecurity;

  /// No description provided for @helpSupport.
  ///
  /// In en, this message translates to:
  /// **'Help & Support'**
  String get helpSupport;

  /// No description provided for @termsConditions.
  ///
  /// In en, this message translates to:
  /// **'Terms & Conditions'**
  String get termsConditions;

  /// No description provided for @comingSoon.
  ///
  /// In en, this message translates to:
  /// **'Coming soon'**
  String get comingSoon;

  /// No description provided for @notificationsComingSoon.
  ///
  /// In en, this message translates to:
  /// **'Notification settings - Coming soon'**
  String get notificationsComingSoon;

  /// No description provided for @privacyComingSoon.
  ///
  /// In en, this message translates to:
  /// **'Privacy settings - Coming soon'**
  String get privacyComingSoon;

  /// No description provided for @helpComingSoon.
  ///
  /// In en, this message translates to:
  /// **'Help & Support - Coming soon'**
  String get helpComingSoon;

  /// No description provided for @termsComingSoon.
  ///
  /// In en, this message translates to:
  /// **'Terms & Conditions - Coming soon'**
  String get termsComingSoon;

  /// No description provided for @logoutConfirm.
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to logout?'**
  String get logoutConfirm;

  /// No description provided for @appDescription.
  ///
  /// In en, this message translates to:
  /// **'Video conferences & meeting management platform'**
  String get appDescription;

  /// No description provided for @allRightsReserved.
  ///
  /// In en, this message translates to:
  /// **'© 2025 Recontext. All rights reserved.'**
  String get allRightsReserved;

  /// No description provided for @tabInfo.
  ///
  /// In en, this message translates to:
  /// **'Info'**
  String get tabInfo;

  /// No description provided for @tabRecordings.
  ///
  /// In en, this message translates to:
  /// **'Recordings'**
  String get tabRecordings;

  /// No description provided for @recordings.
  ///
  /// In en, this message translates to:
  /// **'Recordings'**
  String get recordings;

  /// No description provided for @noRecordingsFound.
  ///
  /// In en, this message translates to:
  /// **'No recordings found'**
  String get noRecordingsFound;

  /// No description provided for @recordingsHint.
  ///
  /// In en, this message translates to:
  /// **'Recordings will appear here after the meeting ends'**
  String get recordingsHint;

  /// No description provided for @session.
  ///
  /// In en, this message translates to:
  /// **'Session'**
  String get session;

  /// Session number label
  ///
  /// In en, this message translates to:
  /// **'Session {number}'**
  String sessionNumber(int number);

  /// Recording duration in minutes
  ///
  /// In en, this message translates to:
  /// **'Duration: {minutes} min'**
  String recordingDuration(int minutes);

  /// No description provided for @playRecording.
  ///
  /// In en, this message translates to:
  /// **'Play Recording'**
  String get playRecording;

  /// No description provided for @roomRecording.
  ///
  /// In en, this message translates to:
  /// **'Room Recording'**
  String get roomRecording;

  /// No description provided for @recordingStatus.
  ///
  /// In en, this message translates to:
  /// **'Status'**
  String get recordingStatus;

  /// No description provided for @loadingRecordings.
  ///
  /// In en, this message translates to:
  /// **'Loading recordings...'**
  String get loadingRecordings;

  /// No description provided for @failedToLoadRecordings.
  ///
  /// In en, this message translates to:
  /// **'Failed to load recordings'**
  String get failedToLoadRecordings;

  /// No description provided for @retryLoadRecordings.
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get retryLoadRecordings;

  /// No description provided for @roomDisconnected.
  ///
  /// In en, this message translates to:
  /// **'Room Disconnected'**
  String get roomDisconnected;

  /// No description provided for @disconnectedMessage.
  ///
  /// In en, this message translates to:
  /// **'You have been disconnected from the meeting. You may have joined from another device.'**
  String get disconnectedMessage;

  /// No description provided for @goToHome.
  ///
  /// In en, this message translates to:
  /// **'Go to Home'**
  String get goToHome;

  /// No description provided for @confirmLeaveTitle.
  ///
  /// In en, this message translates to:
  /// **'Leave Meeting?'**
  String get confirmLeaveTitle;

  /// No description provided for @confirmLeaveMessage.
  ///
  /// In en, this message translates to:
  /// **'Are you sure you want to leave this meeting?'**
  String get confirmLeaveMessage;

  /// No description provided for @confirmLeave.
  ///
  /// In en, this message translates to:
  /// **'Leave'**
  String get confirmLeave;

  /// No description provided for @speaking.
  ///
  /// In en, this message translates to:
  /// **'Speaking'**
  String get speaking;

  /// No description provided for @downloadRecording.
  ///
  /// In en, this message translates to:
  /// **'Download Recording'**
  String get downloadRecording;

  /// No description provided for @downloadComplete.
  ///
  /// In en, this message translates to:
  /// **'Download Complete'**
  String get downloadComplete;

  /// No description provided for @downloadFailed.
  ///
  /// In en, this message translates to:
  /// **'Download Failed'**
  String get downloadFailed;

  /// No description provided for @recordingSaved.
  ///
  /// In en, this message translates to:
  /// **'Recording saved to Downloads'**
  String get recordingSaved;

  /// No description provided for @onlineCount.
  ///
  /// In en, this message translates to:
  /// **'{count} online'**
  String onlineCount(int count);

  /// No description provided for @anonymousGuestsCount.
  ///
  /// In en, this message translates to:
  /// **'{count} guests'**
  String anonymousGuestsCount(int count);

  /// No description provided for @participantsSummary.
  ///
  /// In en, this message translates to:
  /// **'{total} participants ({online} online)'**
  String participantsSummary(int total, int online);

  /// No description provided for @anonymousLinkLabel.
  ///
  /// In en, this message translates to:
  /// **'Anonymous Join Link'**
  String get anonymousLinkLabel;

  /// No description provided for @copyLink.
  ///
  /// In en, this message translates to:
  /// **'Copy Link'**
  String get copyLink;

  /// No description provided for @linkCopied.
  ///
  /// In en, this message translates to:
  /// **'Copied!'**
  String get linkCopied;

  /// No description provided for @anonymousLinkCopied.
  ///
  /// In en, this message translates to:
  /// **'Anonymous link copied to clipboard'**
  String get anonymousLinkCopied;

  /// No description provided for @allowsAnonymousJoin.
  ///
  /// In en, this message translates to:
  /// **'Allows anonymous join'**
  String get allowsAnonymousJoin;

  /// No description provided for @audioOnlyRecording.
  ///
  /// In en, this message translates to:
  /// **'Audio Only'**
  String get audioOnlyRecording;

  /// No description provided for @failedToLoadRecording.
  ///
  /// In en, this message translates to:
  /// **'Failed to load recording'**
  String get failedToLoadRecording;

  /// No description provided for @unknownError.
  ///
  /// In en, this message translates to:
  /// **'Unknown error'**
  String get unknownError;

  /// No description provided for @welcomeBackTitle.
  ///
  /// In en, this message translates to:
  /// **'Welcome back'**
  String get welcomeBackTitle;

  /// No description provided for @welcomeBackSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Sign in to manage your meetings'**
  String get welcomeBackSubtitle;

  /// No description provided for @validationUsernameRequired.
  ///
  /// In en, this message translates to:
  /// **'Please enter your username'**
  String get validationUsernameRequired;

  /// No description provided for @validationPasswordRequired.
  ///
  /// In en, this message translates to:
  /// **'Please enter your password'**
  String get validationPasswordRequired;

  /// No description provided for @apiUrlHint.
  ///
  /// In en, this message translates to:
  /// **'https://portal.recontext.online'**
  String get apiUrlHint;

  /// No description provided for @pleaseCheck.
  ///
  /// In en, this message translates to:
  /// **'Please check:'**
  String get pleaseCheck;

  /// No description provided for @networkConnection.
  ///
  /// In en, this message translates to:
  /// **'Network connection'**
  String get networkConnection;

  /// No description provided for @apiUrlCorrect.
  ///
  /// In en, this message translates to:
  /// **'API URL is correct'**
  String get apiUrlCorrect;

  /// No description provided for @serverRunning.
  ///
  /// In en, this message translates to:
  /// **'Server is running'**
  String get serverRunning;

  /// No description provided for @apiNotConfigured.
  ///
  /// In en, this message translates to:
  /// **'Server URL not configured'**
  String get apiNotConfigured;

  /// No description provided for @currentServer.
  ///
  /// In en, this message translates to:
  /// **'Connected to'**
  String get currentServer;

  /// No description provided for @tabPlayer.
  ///
  /// In en, this message translates to:
  /// **'Player'**
  String get tabPlayer;

  /// No description provided for @tabTranscript.
  ///
  /// In en, this message translates to:
  /// **'Transcript'**
  String get tabTranscript;

  /// No description provided for @tabMemo.
  ///
  /// In en, this message translates to:
  /// **'Memo'**
  String get tabMemo;

  /// No description provided for @noTranscriptAvailable.
  ///
  /// In en, this message translates to:
  /// **'No transcript available'**
  String get noTranscriptAvailable;

  /// No description provided for @transcriptionNotGenerated.
  ///
  /// In en, this message translates to:
  /// **'Transcription has not been generated for this recording yet.'**
  String get transcriptionNotGenerated;

  /// No description provided for @noMemoAvailable.
  ///
  /// In en, this message translates to:
  /// **'No memo available'**
  String get noMemoAvailable;

  /// No description provided for @memoNotGenerated.
  ///
  /// In en, this message translates to:
  /// **'AI-generated summary has not been created for this recording yet.'**
  String get memoNotGenerated;

  /// No description provided for @aiSummary.
  ///
  /// In en, this message translates to:
  /// **'AI Summary'**
  String get aiSummary;

  /// No description provided for @copyMemo.
  ///
  /// In en, this message translates to:
  /// **'Copy memo'**
  String get copyMemo;

  /// No description provided for @memoCopied.
  ///
  /// In en, this message translates to:
  /// **'Memo copied to clipboard'**
  String get memoCopied;

  /// No description provided for @failedToLoadTranscript.
  ///
  /// In en, this message translates to:
  /// **'Failed to load transcript'**
  String get failedToLoadTranscript;

  /// No description provided for @failedToLoadMemo.
  ///
  /// In en, this message translates to:
  /// **'Failed to load memo'**
  String get failedToLoadMemo;

  /// No description provided for @myProfile.
  ///
  /// In en, this message translates to:
  /// **'My Profile'**
  String get myProfile;

  /// No description provided for @editProfileInfo.
  ///
  /// In en, this message translates to:
  /// **'Edit your profile information'**
  String get editProfileInfo;

  /// No description provided for @profile.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profile;

  /// No description provided for @accountInformation.
  ///
  /// In en, this message translates to:
  /// **'Account Information'**
  String get accountInformation;

  /// No description provided for @personalInformation.
  ///
  /// In en, this message translates to:
  /// **'Personal Information'**
  String get personalInformation;

  /// No description provided for @preferences.
  ///
  /// In en, this message translates to:
  /// **'Preferences'**
  String get preferences;

  /// No description provided for @firstName.
  ///
  /// In en, this message translates to:
  /// **'First Name'**
  String get firstName;

  /// No description provided for @lastName.
  ///
  /// In en, this message translates to:
  /// **'Last Name'**
  String get lastName;

  /// No description provided for @phone.
  ///
  /// In en, this message translates to:
  /// **'Phone'**
  String get phone;

  /// No description provided for @bio.
  ///
  /// In en, this message translates to:
  /// **'Bio'**
  String get bio;

  /// No description provided for @notificationPreferences.
  ///
  /// In en, this message translates to:
  /// **'Notification Preferences'**
  String get notificationPreferences;

  /// No description provided for @notifTracksOnly.
  ///
  /// In en, this message translates to:
  /// **'Individual Tracks Only'**
  String get notifTracksOnly;

  /// No description provided for @notifRoomsOnly.
  ///
  /// In en, this message translates to:
  /// **'Rooms Only'**
  String get notifRoomsOnly;

  /// No description provided for @notifBoth.
  ///
  /// In en, this message translates to:
  /// **'Both'**
  String get notifBoth;

  /// No description provided for @failedToLoadProfile.
  ///
  /// In en, this message translates to:
  /// **'Failed to load profile'**
  String get failedToLoadProfile;

  /// No description provided for @failedToSelectImage.
  ///
  /// In en, this message translates to:
  /// **'Failed to select image'**
  String get failedToSelectImage;

  /// No description provided for @profileUpdatedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Profile updated successfully'**
  String get profileUpdatedSuccessfully;

  /// No description provided for @failedToSaveProfile.
  ///
  /// In en, this message translates to:
  /// **'Failed to save profile'**
  String get failedToSaveProfile;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'ru'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'ru':
      return AppLocalizationsRu();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
