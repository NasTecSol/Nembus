//
//  NembusClipApp.swift
//  NembusClip
//

import SwiftUI

@main
struct NembusClipApp: App {
    @State private var scannedURL: URL?

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
