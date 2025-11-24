// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'Recontext';

  @override
  String get login => 'Login';

  @override
  String get loginSubtitle => 'Use your workspace credentials to continue';

  @override
  String get email => 'Email';

  @override
  String get password => 'Password';

  @override
  String get username => 'Username';

  @override
  String get logout => 'Logout';

  @override
  String get loginButton => 'Login';

  @override
  String get loggingIn => 'Logging in...';

  @override
  String get loginError => 'Login failed. Please check your credentials.';

  @override
  String get meetings => 'Meetings';

  @override
  String get documents => 'Documents';

  @override
  String get search => 'Search';

  @override
  String get settings => 'Settings';

  @override
  String get newMeeting => 'New Meeting';

  @override
  String get createMeeting => 'Create Meeting';

  @override
  String get meetingTitle => 'Meeting Title';

  @override
  String get meetingSubject => 'Subject';

  @override
  String get meetingType => 'Type';

  @override
  String get meetingDate => 'Date';

  @override
  String get meetingTime => 'Time';

  @override
  String get meetingDuration => 'Duration';

  @override
  String get meetingRecurrence => 'Recurrence';

  @override
  String get meetingSpeaker => 'Speaker';

  @override
  String get meetingParticipants => 'Participants';

  @override
  String get meetingDepartments => 'Departments';

  @override
  String get meetingNotes => 'Additional Notes';

  @override
  String get meetingRecord => 'Record Meeting (Audio & Video)';

  @override
  String get meetingTranscription =>
      'Transcription (recording individual tracks)';

  @override
  String get meetingForceEndAtDuration => 'Force end after time elapses';

  @override
  String get meetingAllowAnonymous => 'Allow anonymous members';

  @override
  String get recordingOptions => 'Recording Options';

  @override
  String get typeConference => 'Conference';

  @override
  String get typePresentation => 'Presentation';

  @override
  String get typeTraining => 'Training';

  @override
  String get typeDiscussion => 'Discussion';

  @override
  String get recurrenceNone => 'No repeat';

  @override
  String get recurrenceDaily => 'Daily';

  @override
  String get recurrenceWeekly => 'Weekly';

  @override
  String get recurrenceMonthly => 'Monthly';

  @override
  String get recurrencePermanent => 'Permanent';

  @override
  String get permanent => 'Permanent';

  @override
  String get duration15 => '15 minutes';

  @override
  String get duration30 => '30 minutes';

  @override
  String get duration45 => '45 minutes';

  @override
  String get duration60 => '1 hour';

  @override
  String get duration90 => '1.5 hours';

  @override
  String get duration120 => '2 hours';

  @override
  String get duration180 => '3 hours';

  @override
  String get statusScheduled => 'Scheduled';

  @override
  String get statusInProgress => 'In Progress';

  @override
  String get statusCompleted => 'Completed';

  @override
  String get statusCancelled => 'Cancelled';

  @override
  String get statusFinished => 'Finished';

  @override
  String get statusRecording => 'Recording';

  @override
  String get filterAll => 'All';

  @override
  String get filterScheduled => 'Scheduled';

  @override
  String get filterInProgress => 'In Progress';

  @override
  String get filterCompleted => 'Completed';

  @override
  String get selectParticipants => 'Select Participants';

  @override
  String get selectDepartments => 'Select Departments';

  @override
  String get selectSpeaker => 'Select Speaker';

  @override
  String get noSpeaker => 'No speaker';

  @override
  String get searchUsers => 'Search users...';

  @override
  String participantsSelected(int count) {
    return '$count selected';
  }

  @override
  String departmentsSelected(int count) {
    return '$count selected';
  }

  @override
  String get today => 'Today';

  @override
  String get tomorrow => 'Tomorrow';

  @override
  String get minutes => 'min';

  @override
  String get required => '*';

  @override
  String get optional => '(optional)';

  @override
  String get cancel => 'Cancel';

  @override
  String get save => 'Save';

  @override
  String get delete => 'Delete';

  @override
  String get edit => 'Edit';

  @override
  String get refresh => 'Refresh';

  @override
  String get retry => 'Retry';

  @override
  String get close => 'Close';

  @override
  String get select => 'Select';

  @override
  String get loading => 'Loading...';

  @override
  String get error => 'Error';

  @override
  String get success => 'Success';

  @override
  String get noMeetingsFound => 'No meetings found';

  @override
  String get createFirstMeeting => 'Create your first meeting';

  @override
  String get tryChangingFilter => 'Try changing the filter';

  @override
  String get meetingCreatedSuccess => 'Meeting created successfully';

  @override
  String get meetingCreatedError => 'Failed to create meeting';

  @override
  String get meetingUpdatedSuccess => 'Meeting updated successfully';

  @override
  String get meetingDeletedSuccess => 'Meeting deleted successfully';

  @override
  String get failedToLoadMeetings => 'Failed to load meetings';

  @override
  String get failedToLoadFormData => 'Failed to load form data';

  @override
  String get pleaseEnterTitle => 'Please enter a title';

  @override
  String get pleaseSelectSubject => 'Please select a subject';

  @override
  String get apiError => 'API Error';

  @override
  String get connectionError => 'Connection Error';

  @override
  String get unexpectedError => 'Unexpected Error';

  @override
  String get checkNetwork =>
      'Please check:\n• Network connection\n• API server is running\n• You are logged in';

  @override
  String get language => 'Language';

  @override
  String get changeLanguage => 'Change Language';

  @override
  String get english => 'English';

  @override
  String get russian => 'Russian';

  @override
  String get serverUrl => 'Server URL';

  @override
  String get videoServerUrl => 'Video Server URL';

  @override
  String get about => 'About';

  @override
  String get version => 'Version';

  @override
  String get joinMeeting => 'Join Meeting';

  @override
  String get leaveMeeting => 'Leave Meeting';

  @override
  String get meetingDetails => 'Meeting Details';

  @override
  String get enterTitle => 'Enter meeting title';

  @override
  String get enterNotes => 'Enter any additional information';

  @override
  String get documentsTitle => 'Documents';

  @override
  String get documentsSubtitle =>
      'Browse, filter, and download your meeting files';

  @override
  String get noDocumentsFound => 'No documents found';

  @override
  String get uploadFirstDocument => 'Upload your first document';

  @override
  String get upload => 'Upload';

  @override
  String get download => 'Download';

  @override
  String get share => 'Share';

  @override
  String get details => 'Details';

  @override
  String get documentDetails => 'Document Details';

  @override
  String get name => 'Name';

  @override
  String get type => 'Type';

  @override
  String get size => 'Size';

  @override
  String get uploaded => 'Uploaded';

  @override
  String get uploadedBy => 'Uploaded by';

  @override
  String get deleteDocument => 'Delete Document';

  @override
  String deleteDocumentConfirm(String name) {
    return 'Are you sure you want to delete \"$name\"?';
  }

  @override
  String deleted(String name) {
    return 'Deleted $name';
  }

  @override
  String get undo => 'Undo';

  @override
  String downloading(String name) {
    return 'Downloading...';
  }

  @override
  String get filterVideo => 'Video';

  @override
  String get filterAudio => 'Audio';

  @override
  String get filterTranscripts => 'Transcripts';

  @override
  String get filterOther => 'Other';

  @override
  String get filtersTitle => 'Filters';

  @override
  String get uploadDocumentComingSoon => 'Upload document - Coming soon';

  @override
  String get failedToLoadDocuments => 'Failed to load documents';

  @override
  String get searchTitle => 'Search';

  @override
  String get searchHint => 'Search meetings, transcripts, documents...';

  @override
  String get semanticSearch => 'Semantic Search';

  @override
  String get searchDescription =>
      'Search across meetings, transcripts, and documents using natural language';

  @override
  String get searchResultsLabel => 'Results';

  @override
  String get trySearching => 'Try searching for:';

  @override
  String get budgetDiscussion => 'budget discussion';

  @override
  String get projectDeadlines => 'project deadlines';

  @override
  String get whoTalkedAbout => 'who talked about...';

  @override
  String get noResultsFound => 'No results found';

  @override
  String get tryDifferentKeywords => 'Try different keywords';

  @override
  String get yesterday => 'Yesterday';

  @override
  String daysAgo(int count) {
    return '$count days ago';
  }

  @override
  String get searchButton => 'Search';

  @override
  String get match => 'match';

  @override
  String get meeting => 'Meeting';

  @override
  String get transcript => 'Transcript';

  @override
  String get document => 'Document';

  @override
  String get searchFailed => 'Search failed';

  @override
  String get meetingDetailsTitle => 'Meeting Details';

  @override
  String get dateAndTime => 'Date & Time';

  @override
  String get participants => 'Participants';

  @override
  String get noParticipants => 'No participants';

  @override
  String get departments => 'Departments';

  @override
  String get settingsTitle => 'Settings';

  @override
  String get videoRecording => 'Video Recording';

  @override
  String get audioRecording => 'Audio Recording';

  @override
  String get transcription => 'Transcription';

  @override
  String get roomId => 'Room ID';

  @override
  String get notes => 'Notes';

  @override
  String get metadata => 'Metadata';

  @override
  String get created => 'Created';

  @override
  String get updated => 'Updated';

  @override
  String get createdBy => 'Created By';

  @override
  String get active => 'Active';

  @override
  String get scheduled => 'Scheduled';

  @override
  String get completed => 'Completed';

  @override
  String get cancelled => 'Cancelled';

  @override
  String get enabled => 'Enabled';

  @override
  String get disabled => 'Disabled';

  @override
  String get connectingToMeeting => 'Connecting to meeting...';

  @override
  String get connectionFailed => 'Connection Failed';

  @override
  String get failedToJoin => 'Failed to join';

  @override
  String get waitingForParticipants => 'Waiting for participants...';

  @override
  String participantsInRoom(int count) {
    return '$count participant(s) in room';
  }

  @override
  String get you => 'You';

  @override
  String get join => 'Join';

  @override
  String get mic => 'Mic';

  @override
  String get camera => 'Camera';

  @override
  String get leave => 'Leave';

  @override
  String get switchCamera => 'Switch Camera';

  @override
  String get flipCamera => 'Flip';

  @override
  String get userRoleUser => 'User';

  @override
  String get userRoleAdmin => 'Administrator';

  @override
  String get userRoleManager => 'Manager';

  @override
  String get unknownUser => 'Unknown User';

  @override
  String get serverConfiguration => 'Server Configuration';

  @override
  String get apiUrl => 'API URL';

  @override
  String get defaultButton => 'Default';

  @override
  String get restartNote =>
      'Note: You need to restart the app after changing the API URL.';

  @override
  String get apiUrlEmpty => 'API URL cannot be empty';

  @override
  String get apiUrlSaved => 'API URL saved';

  @override
  String get notifications => 'Notifications';

  @override
  String get privacySecurity => 'Privacy & Security';

  @override
  String get helpSupport => 'Help & Support';

  @override
  String get termsConditions => 'Terms & Conditions';

  @override
  String get comingSoon => 'Coming soon';

  @override
  String get notificationsComingSoon => 'Notification settings - Coming soon';

  @override
  String get privacyComingSoon => 'Privacy settings - Coming soon';

  @override
  String get helpComingSoon => 'Help & Support - Coming soon';

  @override
  String get termsComingSoon => 'Terms & Conditions - Coming soon';

  @override
  String get logoutConfirm => 'Are you sure you want to logout?';

  @override
  String get appDescription =>
      'Video conferences & meeting management platform';

  @override
  String get allRightsReserved => '© 2025 Recontext. All rights reserved.';

  @override
  String get tabInfo => 'Info';

  @override
  String get tabRecordings => 'Recordings';

  @override
  String get recordings => 'Recordings';

  @override
  String get noRecordingsFound => 'No recordings found';

  @override
  String get recordingsHint =>
      'Recordings will appear here after the meeting ends';

  @override
  String get session => 'Session';

  @override
  String sessionNumber(int number) {
    return 'Session $number';
  }

  @override
  String recordingDuration(int minutes) {
    return 'Duration: $minutes min';
  }

  @override
  String get playRecording => 'Play Recording';

  @override
  String get roomRecording => 'Room Recording';

  @override
  String get recordingStatus => 'Status';

  @override
  String get loadingRecordings => 'Loading recordings...';

  @override
  String get failedToLoadRecordings => 'Failed to load recordings';

  @override
  String get retryLoadRecordings => 'Retry';

  @override
  String get roomDisconnected => 'Room Disconnected';

  @override
  String get disconnectedMessage =>
      'You have been disconnected from the meeting. You may have joined from another device.';

  @override
  String get goToHome => 'Go to Home';

  @override
  String get confirmLeaveTitle => 'Leave Meeting?';

  @override
  String get confirmLeaveMessage =>
      'Are you sure you want to leave this meeting?';

  @override
  String get confirmLeave => 'Leave';

  @override
  String get speaking => 'Speaking';

  @override
  String get downloadRecording => 'Download Recording';

  @override
  String get downloadComplete => 'Download Complete';

  @override
  String get downloadFailed => 'Download Failed';

  @override
  String get recordingSaved => 'Recording saved to Downloads';

  @override
  String get viewRecordings => 'View Sessions';

  @override
  String get sessionStatusRecording => 'Recording';

  @override
  String get sessionStatusCompleted => 'Completed';

  @override
  String get sessionStatusProcessing => 'Processing';

  @override
  String get sessionStatusFailed => 'Failed';

  @override
  String get sessionStatusFinished => 'Finished';

  @override
  String get viewSession => 'View Session';

  @override
  String get viewSessionDetails =>
      'View session details, tracks, transcript and memo';

  @override
  String onlineCount(int count) {
    return '$count online';
  }

  @override
  String anonymousGuestsCount(int count) {
    return '$count guests';
  }

  @override
  String participantsSummary(int total, int online) {
    return '$total participants ($online online)';
  }

  @override
  String get anonymousLinkLabel => 'Anonymous Join Link';

  @override
  String get copyLink => 'Copy Link';

  @override
  String get linkCopied => 'Copied!';

  @override
  String get anonymousLinkCopied => 'Anonymous link copied to clipboard';

  @override
  String get allowsAnonymousJoin => 'Allows anonymous join';

  @override
  String get audioOnlyRecording => 'Audio Only';

  @override
  String get failedToLoadRecording => 'Failed to load recording';

  @override
  String get unknownError => 'Unknown error';

  @override
  String get welcomeBackTitle => 'Welcome back';

  @override
  String get welcomeBackSubtitle => 'Sign in to manage your meetings';

  @override
  String get validationUsernameRequired => 'Please enter your username';

  @override
  String get validationPasswordRequired => 'Please enter your password';

  @override
  String get apiUrlHint => 'https://portal.recontext.online';

  @override
  String get pleaseCheck => 'Please check:';

  @override
  String get networkConnection => 'Network connection';

  @override
  String get apiUrlCorrect => 'API URL is correct';

  @override
  String get serverRunning => 'Server is running';

  @override
  String get apiNotConfigured => 'Server URL not configured';

  @override
  String get currentServer => 'Connected to';

  @override
  String get tabPlayer => 'Player';

  @override
  String get tabTranscript => 'Transcript';

  @override
  String get tabMemo => 'Memo';

  @override
  String get noTranscriptAvailable => 'No transcript available';

  @override
  String get transcriptionNotGenerated =>
      'Transcription has not been generated for this recording yet.';

  @override
  String get noMemoAvailable => 'No memo available';

  @override
  String get memoNotGenerated =>
      'AI-generated summary has not been created for this recording yet.';

  @override
  String get aiSummary => 'AI Summary';

  @override
  String get copyMemo => 'Copy memo';

  @override
  String get memoCopied => 'Memo copied to clipboard';

  @override
  String get failedToLoadTranscript => 'Failed to load transcript';

  @override
  String get failedToLoadMemo => 'Failed to load memo';

  @override
  String get myProfile => 'My Profile';

  @override
  String get editProfileInfo => 'Edit your profile information';

  @override
  String get profile => 'Profile';

  @override
  String get accountInformation => 'Account Information';

  @override
  String get personalInformation => 'Personal Information';

  @override
  String get preferences => 'Preferences';

  @override
  String get firstName => 'First Name';

  @override
  String get lastName => 'Last Name';

  @override
  String get phone => 'Phone';

  @override
  String get bio => 'Bio';

  @override
  String get notificationPreferences => 'Notification Preferences';

  @override
  String get notifTracksOnly => 'Individual Tracks Only';

  @override
  String get notifRoomsOnly => 'Rooms Only';

  @override
  String get notifBoth => 'Both';

  @override
  String get failedToLoadProfile => 'Failed to load profile';

  @override
  String get failedToSelectImage => 'Failed to select image';

  @override
  String get profileUpdatedSuccessfully => 'Profile updated successfully';

  @override
  String get failedToSaveProfile => 'Failed to save profile';

  @override
  String get viewCompositeVideo => 'View Composite Video';

  @override
  String get processingVideo => 'Processing...';

  @override
  String get processingVideoDescription =>
      'Merging video tracks and creating composite video';

  @override
  String get nowSpeaking => 'Now Speaking';

  @override
  String get speaker => 'Speaker';

  @override
  String get participant => 'Participant';
}
