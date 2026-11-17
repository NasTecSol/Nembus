import java.io.File
import java.io.FileInputStream
import java.util.Properties

plugins {
    id("com.android.application")
    id("dev.flutter.flutter-gradle-plugin")
}

val keystorePropertiesFile = rootProject.file("key.properties")
val keystoreProperties = Properties()
if (keystorePropertiesFile.exists()) {
    keystoreProperties.load(FileInputStream(keystorePropertiesFile))
}

val buildGoLibrary = tasks.register<Exec>("buildGoLibrary") {
    description = "Compiles Go Core Engine into native .so libraries for target ABIs"
    workingDir = file("${project.projectDir}/..")
    commandLine("bash", "${project.projectDir}/../scripts/build_go_android.sh")
}

tasks.named("preBuild") {
    dependsOn(buildGoLibrary)
}

android {
    namespace = "com.nastecsol.nembus"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    sourceSets {
        getByName("main") {
            jniLibs.srcDirs("src/main/jniLibs")
        }
    }

    defaultConfig {
        applicationId = "com.nastecsol.nembus"
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
        ndk {
            abiFilters.addAll(listOf("arm64-v8a", "armeabi-v7a", "x86_64"))
        }
    }

    signingConfigs {
        create("release") {
            keyAlias = keystoreProperties.getProperty("keyAlias") ?: "nembus"
            keyPassword = keystoreProperties.getProperty("keyPassword") ?: "nembus123"
            val storeFilePath = keystoreProperties.getProperty("storeFile") ?: "nembus.keystore"
            storeFile = when {
                file(storeFilePath).exists() -> file(storeFilePath)
                rootProject.file(storeFilePath).exists() -> rootProject.file(storeFilePath)
                rootProject.file("app/$storeFilePath").exists() -> rootProject.file("app/$storeFilePath")
                else -> file("nembus.keystore")
            }
            storePassword = keystoreProperties.getProperty("storePassword") ?: "nembus123"
        }
    }

    buildTypes {
        release {
            signingConfig = signingConfigs.getByName("release")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

dependencies {
    implementation("androidx.core:core-ktx:1.13.1")
    implementation("androidx.appcompat:appcompat:1.7.0")
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile>().configureEach {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}
