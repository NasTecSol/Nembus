//
//  NembusClipApp.swift
//  NembusClip
//

import SwiftUI
import UserNotifications

class ClipNotificationDelegate: NSObject, UNUserNotificationCenterDelegate {
    static let shared = ClipNotificationDelegate()

    // Ensure audio alert & banner play even when phone screen is open / app in foreground
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        completionHandler([.banner, .sound, .badge, .list])
    }
}

@main
struct NembusClipApp: App {
    @State private var scannedURL: URL?

    init() {
        // Set delegate to ensure alarm sound plays whether screen is locked or open
        UNUserNotificationCenter.current().delegate = ClipNotificationDelegate.shared

        // Request notification permission for alarm alerts
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
            print("Notification permission granted: \(granted)")
        }

        // Copy timer sound files to Library/Sounds for iOS notification daemon
        let soundFiles = ["timer.caf", "timer.wav", "timer.mp3"]
        let fileManager = FileManager.default
        if let libraryDir = fileManager.urls(for: .libraryDirectory, in: .userDomainMask).first {
            let soundsDir = libraryDir.appendingPathComponent("Sounds", isDirectory: true)
            try? fileManager.createDirectory(at: soundsDir, withIntermediateDirectories: true)
            for file in soundFiles {
                let base = (file as NSString).deletingPathExtension
                let ext = (file as NSString).pathExtension
                if let srcURL = Bundle.main.url(forResource: base, withExtension: ext) ??
                                Bundle.main.url(forResource: base, withExtension: ext, subdirectory: "assets") {
                    let destURL = soundsDir.appendingPathComponent(file)
                    try? fileManager.removeItem(at: destURL)
                    try? fileManager.copyItem(at: srcURL, to: destURL)
                    print("Successfully registered \(file) in Library/Sounds for lock screen playback")
                }
            }
        }
    }

    var body: some Scene {
        WindowGroup {
            ContentView(currentURL: $scannedURL)
                // Handle Native Camera & App Clip Code Invocations
                .onContinueUserActivity(NSUserActivityTypeBrowsingWeb) { userActivity in
                    guard let webpageURL = userActivity.webpageURL else { return }
                    print("App Clip launched via NSUserActivity: \(webpageURL)")
                    self.scannedURL = webpageURL
                }
                // Handle Deep Link & Web Invocations
                .onOpenURL { url in
                    print("App Clip launched via onOpenURL: \(url)")
                    self.scannedURL = url
                }
        }
    }
}
