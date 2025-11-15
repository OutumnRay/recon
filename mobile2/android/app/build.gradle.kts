import com.github.triplet.gradle.androidpublisher.ReleaseStatus
import java.io.File
import java.util.Locale
import java.util.Properties

plugins {
    id("com.android.application")
    id("kotlin-android")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
    id("com.github.triplet.play")
    id("com.google.gms.google-services")
}

val keystoreProperties = Properties()
val keystorePropertiesFile = rootProject.file("key.properties")
if (keystorePropertiesFile.exists()) {
    keystorePropertiesFile.inputStream().use { keystoreProperties.load(it) }
} else {
    project.logger.warn("key.properties not found; release builds will be signed with the debug key.")
}

val keystorePath = keystoreProperties.getProperty("storeFile")
val resolvedKeystoreFile: File? = keystorePath?.let { path ->
    val direct = File(path)
    val candidates = mutableListOf<File>()
    if (direct.isAbsolute) {
        candidates += direct
    } else {
        candidates += rootProject.file(path)
        candidates += project.file(path)
    }
    candidates.firstOrNull { it.exists() }
}

if (!keystorePath.isNullOrBlank() && resolvedKeystoreFile == null) {
    project.logger.warn("Configured keystore path '$keystorePath' was not found. Place the keystore there or update key.properties.")
}

val hasReleaseSigning =
    resolvedKeystoreFile != null &&
        !keystoreProperties.getProperty("storePassword").isNullOrBlank() &&
        !keystoreProperties.getProperty("keyAlias").isNullOrBlank() &&
        !keystoreProperties.getProperty("keyPassword").isNullOrBlank()

android {
    namespace = "uk.team21.recontext"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }

    kotlinOptions {
        jvmTarget = JavaVersion.VERSION_11.toString()
    }

    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId = "uk.team21.recontext"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    signingConfigs {
        create("release") {
            resolvedKeystoreFile?.let { storeFile = it }
            storePassword = keystoreProperties.getProperty("storePassword")
            keyAlias = keystoreProperties.getProperty("keyAlias")
            keyPassword = keystoreProperties.getProperty("keyPassword")
        }
    }

    buildTypes {
        release {
            if (hasReleaseSigning) {
                signingConfig = signingConfigs.getByName("release")
            } else {
                project.logger.warn("Release signing config incomplete; falling back to debug signing.")
                signingConfig = signingConfigs.getByName("debug")
            }
        }
    }
}

flutter {
    source = "../.."
}

play {
    val credentialsPath =
        System.getenv("PLAY_SERVICE_ACCOUNT_JSON")
            ?: (project.findProperty("PLAY_SERVICE_ACCOUNT_JSON") as? String)
    if (!credentialsPath.isNullOrBlank()) {
        val credentialsFile = project.file(credentialsPath)
        if (credentialsFile.exists()) {
            serviceAccountCredentials.set(credentialsFile)
        } else {
            project.logger.warn("Play credentials file '$credentialsPath' was not found; publishing tasks will fail.")
        }
    }

    val resolvedTrack =
        System.getenv("PLAY_TRACK")
            ?: (project.findProperty("PLAY_TRACK") as? String)
            ?: "internal"
    track.set(resolvedTrack)

    val statusInput =
        System.getenv("PLAY_RELEASE_STATUS")
            ?: (project.findProperty("PLAY_RELEASE_STATUS") as? String)
    val status =
        when (statusInput?.lowercase(Locale.US)) {
            "completed" -> ReleaseStatus.COMPLETED
            "halted" -> ReleaseStatus.HALTED
            "inprogress", "in_progress", "in-progress" -> ReleaseStatus.IN_PROGRESS
            else -> ReleaseStatus.DRAFT
        }
    releaseStatus.set(status)

    val releaseNameInput =
        System.getenv("PLAY_RELEASE_NAME")
            ?: (project.findProperty("PLAY_RELEASE_NAME") as? String)
    if (!releaseNameInput.isNullOrBlank()) {
        releaseName.set(releaseNameInput)
    }

    defaultToAppBundles.set(true)
}
