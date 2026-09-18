import java.io.File

plugins {
    id("com.android.application")
    id("dev.flutter.flutter-gradle-plugin")
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
    namespace = "com.example.pos_mobile"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    sourceSets {
        getByName("main") {
            jniLibs.srcDirs("src/main/jniLibs")
        }
    }

    defaultConfig {
        applicationId = "com.example.pos_mobile"
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
        ndk {
            abiFilters.addAll(listOf("arm64-v8a", "armeabi-v7a", "x86_64"))
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}