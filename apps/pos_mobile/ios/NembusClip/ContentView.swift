//
//  ContentView.swift
//  NembusClip
//  Clean, Regular & Uncongested UI/UX with Live Cooking Countdown & Lock Screen Audio Alarm
//

import SwiftUI
import ActivityKit
import AudioToolbox
import AVFoundation
import Combine
import UserNotifications

// MARK: - Push Token Payload
struct PushTokenPayload: Codable {
    let token: String
    let platform: String
    let environment: String
}

struct ContentView: View {
    @Binding var currentURL: URL?

    @State private var isLoading: Bool = false
    @State private var isSuccess: Bool = false
    @State private var errorMessage: String? = nil
    @State private var orderNumber: String = ""
    @State private var shopName: String = ""
    @State private var itemsSummary: String = ""
    @State private var status: String = ""
    @State private var etaMinutes: Int = 0

    // MARK: - Enhanced Live Timer & Alarm State
    @State private var remainingSeconds: Int = 0
    @State private var totalTimerSeconds: Int = 0
    @State private var isTimerActive: Bool = false
    @State private var isOrderCooked: Bool = false
    @State private var alarmHasRung: Bool = false
    @State private var activeActivity: Activity<NembusOrderAttributes>? = nil
    @State private var audioPlayer: AVAudioPlayer? = nil

    // Animation drivers
    @State private var panRocking: Bool = false
    @State private var flamePulsing: Bool = false
    @State private var steamOffset: CGFloat = 0

    private let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()
    private let apiBaseURL = "https://nac-production.up.railway.app"

    var body: some View {
        ZStack {
            Color(.systemBackground).ignoresSafeArea()

            ScrollView(showsIndicators: false) {
                VStack(spacing: 20) {
                    if isSuccess {
                        // MARK: - Live Order Dashboard (Cooking or Ready)
                        liveOrderTrackingView
                            .transition(.scale.combined(with: .opacity))

                    } else if isLoading {
                        // MARK: - Loading State
                        VStack(spacing: 16) {
                            Spacer().frame(height: 100)
                            ProgressView()
                                .scaleEffect(1.4)
                                .padding(.bottom, 8)
                            Text("Fetching Order Details...")
                                .font(.headline)
                            Text("Connecting to server...")
                                .font(.caption)
                                .foregroundColor(.secondary)
                        }

                    } else if let error = errorMessage {
                        // MARK: - Error State
                        VStack(spacing: 16) {
                            Spacer().frame(height: 80)
                            Image(systemName: "exclamationmark.triangle.fill")
                                .font(.system(size: 56))
                                .foregroundColor(.orange)

                            Text("Could Not Start Widget")
                                .font(.headline)

                            Text(error)
                                .font(.subheadline)
                                .foregroundColor(.secondary)
                                .multilineTextAlignment(.center)
                                .padding(.horizontal, 24)

                            Button(action: {
                                if let url = currentURL {
                                    handleIncomingURL(url)
                                }
                            }) {
                                Text("Try Again")
                                    .font(.headline)
                                    .foregroundColor(.white)
                                    .padding(.horizontal, 24)
                                    .padding(.vertical, 10)
                                    .background(Color.blue)
                                    .cornerRadius(10)
                            }
                            .padding(.top, 8)
                        }

                    } else {
                        // MARK: - Waiting for QR Scan
                        waitingForQRScanView
                    }
                }
                .padding()
            }
        }
        .onAppear {
            if let url = currentURL {
                handleIncomingURL(url)
            } else {
                cleanupStaleActivities()
            }
        }
        .onChange(of: currentURL) { newURL in
            if let newURL = newURL {
                handleIncomingURL(newURL)
            }
        }
        .onReceive(timer) { _ in
            guard isTimerActive && !isOrderCooked else { return }
            if remainingSeconds > 1 {
                remainingSeconds -= 1
            } else {
                remainingSeconds = 0
                isTimerActive = false
                triggerOrderCookedState()
            }
        }
    }

    // MARK: - Live Order Tracking View
    private var liveOrderTrackingView: some View {
        VStack(spacing: 16) {
            // Header Row: Shop Name & Order Pill
            HStack(alignment: .center) {
                HStack(spacing: 8) {
                    ZStack {
                        Circle()
                            .fill(LinearGradient(colors: [.cyan.opacity(0.3), .blue.opacity(0.3)], startPoint: .topLeading, endPoint: .bottomTrailing))
                            .frame(width: 34, height: 34)
                        Image(systemName: "cup.and.saucer.fill")
                            .font(.system(size: 15, weight: .bold))
                            .foregroundColor(.cyan)
                    }

                    VStack(alignment: .leading, spacing: 1) {
                        Text(shopName.isEmpty ? "Demo Coffee" : shopName)
                            .font(.system(size: 16, weight: .bold))
                        if !itemsSummary.isEmpty {
                            Text(itemsSummary)
                                .font(.caption)
                                .foregroundColor(.secondary)
                                .lineLimit(1)
                        }
                    }
                }

                Spacer()

                Text("Order \(orderNumber)")
                    .font(.system(size: 14, weight: .heavy, design: .rounded))
                    .padding(.horizontal, 10)
                    .padding(.vertical, 5)
                    .background(Color.blue.opacity(0.12))
                    .foregroundColor(.blue)
                    .cornerRadius(8)
            }
            .padding(.horizontal, 4)

            Divider()

            if isOrderCooked {
                // MARK: - Cooked & Ready Alarm Presentation
                readyForPickupCard
            } else {
                // MARK: - Cooking in Progress & Countdown Timer
                cookingInProgressCard
            }

            // Lock Screen Live Activity Info Banner
            HStack(spacing: 8) {
                Image(systemName: "lock.shield.fill")
                    .foregroundColor(.blue)
                Text("Alarm ringtone set for Lock Screen & Dynamic Island")
                    .font(.footnote)
                    .fontWeight(.medium)
                    .foregroundColor(.blue)
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 7)
            .background(Color.blue.opacity(0.08))
            .cornerRadius(16)
            .padding(.top, 4)
        }
    }

    // MARK: - Cooking in Progress Card (Regular Size & Well-Spaced)
    private var cookingInProgressCard: some View {
        VStack(spacing: 16) {
            // Status Badge
            HStack(spacing: 6) {
                Image(systemName: "flame.fill")
                    .symbolEffect(.pulse)
                    .foregroundColor(.orange)
                Text("COOKING IN PROGRESS")
                    .font(.system(size: 12, weight: .bold))
                    .foregroundColor(.orange)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 5)
            .background(Color.orange.opacity(0.12))
            .cornerRadius(14)

            // Regular Sized Animated Cooking Visual
            ZStack {
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [Color.orange.opacity(0.2), Color.orange.opacity(0.03), Color.clear],
                            center: .center,
                            startRadius: 15,
                            endRadius: 60
                        )
                    )
                    .frame(width: 120, height: 120)

                VStack(spacing: -4) {
                    // Rising steam waves
                    HStack(spacing: 5) {
                        Image(systemName: "smoke.fill")
                            .font(.system(size: 11))
                            .foregroundColor(.orange.opacity(0.6))
                            .offset(y: steamOffset)
                        Image(systemName: "smoke.fill")
                            .font(.system(size: 15))
                            .foregroundColor(.orange.opacity(0.8))
                            .offset(y: steamOffset * 1.2)
                        Image(systemName: "smoke.fill")
                            .font(.system(size: 11))
                            .foregroundColor(.orange.opacity(0.6))
                            .offset(y: steamOffset * 0.9)
                    }

                    // Sizzling Frying Pan (Regular 38pt size)
                    Image(systemName: "frying.pan.fill")
                        .font(.system(size: 38))
                        .foregroundColor(.orange)
                        .rotationEffect(.degrees(panRocking ? -4 : 4))
                        .animation(.easeInOut(duration: 0.8).repeatForever(autoreverses: true), value: panRocking)

                    // Flickering heat flames
                    HStack(spacing: 4) {
                        Image(systemName: "flame.fill")
                            .font(.system(size: 13))
                            .foregroundColor(.yellow)
                            .scaleEffect(flamePulsing ? 1.15 : 0.9)
                        Image(systemName: "flame.fill")
                            .font(.system(size: 16))
                            .foregroundColor(.orange)
                            .scaleEffect(flamePulsing ? 0.9 : 1.2)
                        Image(systemName: "flame.fill")
                            .font(.system(size: 13))
                            .foregroundColor(.yellow)
                            .scaleEffect(flamePulsing ? 1.15 : 0.9)
                    }
                    .animation(.easeInOut(duration: 0.6).repeatForever(autoreverses: true), value: flamePulsing)
                }
            }
            .onAppear {
                panRocking = true
                flamePulsing = true
                withAnimation(.easeInOut(duration: 1.2).repeatForever(autoreverses: true)) {
                    steamOffset = -5
                }
            }

            // Regular Sized Countdown Clock
            VStack(spacing: 3) {
                Text(formatRemainingTime(remainingSeconds))
                    .font(.system(size: 42, weight: .heavy, design: .rounded))
                    .monospacedDigit()
                    .foregroundColor(.orange)

                Text("ESTIMATED REMAINING")
                    .font(.system(size: 10, weight: .bold))
                    .foregroundColor(.secondary)
                    .tracking(0.8)
            }

            // Smooth Progress Bar
            if totalTimerSeconds > 0 {
                ProgressView(value: Double(totalTimerSeconds - remainingSeconds), total: Double(totalTimerSeconds))
                    .tint(LinearGradient(colors: [.orange, .yellow], startPoint: .leading, endPoint: .trailing))
                    .padding(.horizontal, 10)
            }

            // Quick Testing Actions
            HStack(spacing: 12) {
                Button(action: {
                    // Fast forward to 0 and trigger alarm
                    remainingSeconds = 0
                    isTimerActive = false
                    triggerOrderCookedState()
                }) {
                    HStack(spacing: 4) {
                        Image(systemName: "bell.badge.fill")
                        Text("Simulate Ready (0:00)")
                    }
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.white)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 7)
                    .background(Color.orange)
                    .cornerRadius(8)
                }

                Button(action: {
                    // Set 1 minute test
                    remainingSeconds = 60
                    totalTimerSeconds = 60
                    isTimerActive = true
                    isOrderCooked = false
                    alarmHasRung = false
                    scheduleAlarmNotification(orderNumber: orderNumber, shopName: shopName, inSeconds: 60)
                }) {
                    Text("1 Min Test")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.orange)
                        .padding(.horizontal, 12)
                        .padding(.vertical, 7)
                        .background(Color.orange.opacity(0.12))
                        .cornerRadius(8)
                }
            }
            .padding(.top, 2)
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 18)
                .fill(Color(UIColor.secondarySystemBackground))
        )
    }

    // MARK: - Cooked & Ready Alarm Presentation Card
    private var readyForPickupCard: some View {
        VStack(spacing: 16) {
            // Ringing Alarm Bell with Waves
            ZStack {
                Circle()
                    .fill(Color.green.opacity(0.16))
                    .frame(width: 80, height: 80)
                Image(systemName: "bell.and.waves.left.and.right.fill")
                    .font(.system(size: 38, weight: .bold))
                    .symbolEffect(.variableColor.iterative)
                    .foregroundColor(.green)
            }

            // Prepared Food Visuals
            HStack(spacing: 16) {
                Image(systemName: "cup.and.saucer.fill")
                    .font(.system(size: 26))
                    .foregroundColor(.cyan)

                Image(systemName: "sparkles")
                    .font(.system(size: 20))
                    .foregroundColor(.yellow)

                Image(systemName: "takeoutbag.and.cup.and.straw.fill")
                    .font(.system(size: 26))
                    .foregroundColor(.orange)
            }

            // Prominent Announcement
            VStack(spacing: 4) {
                Text("ORDER IS COOKED!")
                    .font(.system(size: 24, weight: .heavy, design: .rounded))
                    .foregroundColor(.green)

                Text("You can collect from counter")
                    .font(.system(size: 16, weight: .bold))
                    .foregroundColor(.primary)
            }

            // Pickup Guidance Card
            VStack(spacing: 6) {
                HStack(spacing: 5) {
                    Image(systemName: "checkmark.seal.fill")
                        .foregroundColor(.green)
                    Text("Ready for Pickup")
                        .font(.subheadline.bold())
                        .foregroundColor(.green)
                }

                Text("Order #\(orderNumber)")
                    .font(.system(size: 22, weight: .heavy, design: .rounded))
                    .foregroundColor(.primary)

                Text("Please show this number at \(shopName.isEmpty ? "the counter" : shopName)")
                    .font(.footnote)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
            }
            .padding(12)
            .frame(maxWidth: .infinity)
            .background(Color.green.opacity(0.08))
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(Color.green.opacity(0.3), lineWidth: 1)
            )

            // Replay Alarm Sound Button
            Button(action: {
                playAlarmRing()
            }) {
                HStack(spacing: 6) {
                    Image(systemName: "bell.badge.fill")
                    Text("Ring Alarm Sound Again")
                }
                .font(.system(size: 13, weight: .bold))
                .foregroundColor(.green)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(Color.green.opacity(0.12))
                .cornerRadius(8)
            }
        }
        .padding(18)
        .background(
            RoundedRectangle(cornerRadius: 18)
                .fill(Color(UIColor.secondarySystemBackground))
                .overlay(
                    RoundedRectangle(cornerRadius: 18)
                        .stroke(Color.green.opacity(0.35), lineWidth: 1.2)
                )
        )
    }

    // MARK: - Waiting for QR Scan Screen
    private var waitingForQRScanView: some View {
        VStack(spacing: 20) {
            Spacer().frame(height: 50)

            Image(systemName: "qrcode.viewfinder")
                .font(.system(size: 72))
                .foregroundColor(.blue)

            VStack(spacing: 6) {
                Text("Scan an Order QR Code")
                    .font(.title3)
                    .fontWeight(.bold)

                Text("Point your iPhone camera at the QR code to track your order on the Lock Screen.")
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)
            }

            // Quick Demo Buttons
            VStack(spacing: 10) {
                Text("Demo Preview")
                    .font(.caption.bold())
                    .foregroundColor(.secondary)

                HStack(spacing: 12) {
                    Button(action: {
                        let demoJSON: [String: Any] = [
                            "orderNumber": "A-1",
                            "shopName": "Demo Coffee",
                            "status": "preparing",
                            "etaMinutes": 25,
                            "updatedAt": Int(Date().timeIntervalSince1970)
                        ]
                        processOrderDictionary(demoJSON, fallbackOrderId: "A-1")
                    }) {
                        Text("Demo Order (25m ETA)")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundColor(.white)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 7)
                            .background(Color.blue)
                            .cornerRadius(8)
                    }

                    Button(action: {
                        let fastJSON: [String: Any] = [
                            "orderNumber": "A-1",
                            "shopName": "Demo Coffee",
                            "status": "preparing",
                            "etaMinutes": 1,
                            "updatedAt": Int(Date().timeIntervalSince1970)
                        ]
                        processOrderDictionary(fastJSON, fallbackOrderId: "A-1")
                    }) {
                        Text("Fast Test (1m ETA)")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundColor(.blue)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 7)
                            .background(Color.blue.opacity(0.12))
                            .cornerRadius(8)
                    }
                }
            }
            .padding(.top, 20)
        }
    }

    // MARK: - Time Formatter Helper
    private func formatRemainingTime(_ seconds: Int) -> String {
        let m = seconds / 60
        let s = seconds % 60
        return String(format: "%02d:%02d", m, s)
    }

    // MARK: - Schedule Local Notification for Lock Screen Alert
    private func scheduleAlarmNotification(orderNumber: String, shopName: String, inSeconds seconds: Int) {
        UNUserNotificationCenter.current().removePendingNotificationRequests(withIdentifiers: ["order_ready_\(orderNumber)"])

        guard seconds > 0 else { return }

        let content = UNMutableNotificationContent()
        content.title = "Order is cooked you can collect from counter"
        content.body = "Order #\(orderNumber) is ready for pickup at \(shopName.isEmpty ? "the counter" : shopName)!"
        // UNNotificationSound requires .caf / .wav container to play custom sound
        content.sound = UNNotificationSound(named: UNNotificationSoundName("timer.caf"))
        if #available(iOS 15.0, *) {
            content.interruptionLevel = .timeSensitive
        }

        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: TimeInterval(seconds), repeats: false)
        let request = UNNotificationRequest(identifier: "order_ready_\(orderNumber)", content: content, trigger: trigger)

        UNUserNotificationCenter.current().add(request) { error in
            if let error = error {
                print("Failed to schedule alarm notification: \(error.localizedDescription)")
            } else {
                print("Successfully scheduled alarm notification in \(seconds)s with timer.caf sound")
            }
        }
    }

    // MARK: - Send Immediate Local Notification of Order Completion
    private func sendOrderCompletionNotification(orderNumber: String, shopName: String) {
        UNUserNotificationCenter.current().removePendingNotificationRequests(withIdentifiers: ["order_ready_\(orderNumber)"])

        let content = UNMutableNotificationContent()
        content.title = "Order is cooked you can collect from counter"
        content.body = "Order #\(orderNumber) is ready for pickup at \(shopName.isEmpty ? "the counter" : shopName)!"
        // Plays custom timer sound on lock screen & open
        content.sound = UNNotificationSound(named: UNNotificationSoundName("timer.caf"))
        if #available(iOS 15.0, *) {
            content.interruptionLevel = .timeSensitive
        }

        // Trigger immediately (0.1 second)
        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: 0.1, repeats: false)
        let request = UNNotificationRequest(
            identifier: "order_completed_\(orderNumber)_\(Int(Date().timeIntervalSince1970))",
            content: content,
            trigger: trigger
        )

        UNUserNotificationCenter.current().add(request) { error in
            if let error = error {
                print("Failed to send order completion notification: \(error.localizedDescription)")
            } else {
                print("Immediate order completion notification dispatched with timer.caf sound!")
            }
        }
    }

    // MARK: - Alarm & Cooked Trigger (When timer == 0:00 or stop)
    private func triggerOrderCookedState() {
        withAnimation(.spring(response: 0.6, dampingFraction: 0.7)) {
            isOrderCooked = true
            status = "ready"
        }

        if !alarmHasRung {
            alarmHasRung = true
            // 1. Play timer tune immediately
            playAlarmRing()
            // 2. Send local notification of order completion immediately
            sendOrderCompletionNotification(orderNumber: orderNumber, shopName: shopName)
        }

        // Update Live Activity on Lock Screen / Dynamic Island
        if let activity = activeActivity {
            Task {
                let readyState = NembusOrderAttributes.ContentState(
                    orderNumber: orderNumber,
                    status: "ready",
                    etaMinutes: 0,
                    updatedAt: Int(Date().timeIntervalSince1970)
                )
                await activity.update(ActivityContent(state: readyState, staleDate: nil))
            }
        }
    }

    // MARK: - Play Alarm Sound & Haptic Alert (using timer.mp3 / timer.caf)
    private func playAlarmRing() {
        TimerSoundService.shared.playTimerSound()
    }

    // MARK: - Step 1: Extract orderId or direct JSON from QR URL
    private func handleIncomingURL(_ url: URL) {
        print("Handling QR URL: \(url.absoluteString)")

        // Support direct JSON payload in QR code (e.g. {"orderNumber":"A-1",...})
        let urlString = url.absoluteString.removingPercentEncoding ?? url.absoluteString
        if let startIdx = urlString.firstIndex(of: "{"),
           let endIdx = urlString.lastIndex(of: "}"),
           startIdx < endIdx {
            let jsonSub = String(urlString[startIdx...endIdx])
            if let data = jsonSub.data(using: .utf8),
               let dict = (try? JSONSerialization.jsonObject(with: data)) as? [String: Any] {
                print("Direct JSON parsed successfully from QR code URL!")
                processOrderDictionary(dict, fallbackOrderId: "A-1")
                return
            }
        }

        // Standard App Clip path: /o/:orderId
        let pathComponents = url.pathComponents
        guard let oIndex = pathComponents.firstIndex(of: "o"), oIndex + 1 < pathComponents.count else {
            self.errorMessage = "Invalid QR code format. Missing order identifier."
            return
        }

        let orderId = pathComponents[oIndex + 1]
        guard !orderId.isEmpty else {
            self.errorMessage = "Order ID is empty."
            return
        }

        fetchOrderAndStartActivity(orderId: orderId)
    }

    // MARK: - Step 2: GET /v1/orders/:orderId
    private func fetchOrderAndStartActivity(orderId: String) {
        guard let requestURL = URL(string: "\(apiBaseURL)/v1/orders/\(orderId)") else {
            self.errorMessage = "Invalid API URL."
            return
        }

        self.isLoading = true
        self.errorMessage = nil

        var request = URLRequest(url: requestURL)
        request.httpMethod = "GET"
        request.timeoutInterval = 10.0

        URLSession.shared.dataTask(with: request) { data, response, error in
            DispatchQueue.main.async {
                self.isLoading = false

                if let error = error {
                    self.errorMessage = "Network error: \(error.localizedDescription)"
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    self.errorMessage = "Server did not respond."
                    return
                }

                if httpResponse.statusCode == 404 {
                    self.errorMessage = "Order not found or has expired."
                    return
                }

                guard (200...299).contains(httpResponse.statusCode), let data = data else {
                    self.errorMessage = "Server error (\(httpResponse.statusCode))."
                    return
                }

                let rawJSON = String(data: data, encoding: .utf8) ?? ""
                print("RAW API RESPONSE: \(rawJSON)")

                guard let rawDict = (try? JSONSerialization.jsonObject(with: data)) as? [String: Any] else {
                    self.errorMessage = "Could not parse JSON response from server."
                    return
                }

                self.processOrderDictionary(rawDict, fallbackOrderId: orderId)
            }
        }.resume()
    }

    // MARK: - Process Parsed Order Dictionary
    private func processOrderDictionary(_ rawDict: [String: Any], fallbackOrderId: String) {
        let dict: [String: Any] = (rawDict["order"] as? [String: Any]) ??
                                  (rawDict["data"] as? [String: Any]) ??
                                  rawDict

        // 1. Order Number
        let parsedOrderNumber: String = {
            if let s = dict["orderNumber"] as? String, !s.isEmpty { return s }
            if let s = dict["order_number"] as? String, !s.isEmpty { return s }
            if let i = dict["orderNumber"] as? Int { return "A-\(i)" }
            if let i = dict["order_number"] as? Int { return "A-\(i)" }
            return fallbackOrderId
        }()

        // 2. Shop Name
        let parsedShopName: String = {
            if let s = dict["shopName"] as? String, !s.isEmpty { return s }
            if let s = dict["shop_name"] as? String, !s.isEmpty { return s }
            if let s = dict["storeName"] as? String, !s.isEmpty { return s }
            return "Demo Coffee"
        }()

        // 3. Status
        let parsedStatus: String = {
            if let s = dict["status"] as? String, !s.isEmpty { return s }
            return "preparing"
        }()

        // 4. ETA Minutes
        let parsedETA: Int = {
            if let i = dict["etaMinutes"] as? Int { return i }
            if let i = dict["eta_minutes"] as? Int { return i }
            if let i = dict["eta"] as? Int { return i }
            if let s = dict["etaMinutes"] as? String, let i = Int(s) { return i }
            return 0
        }()

        // 5. Updated At (convert ms to seconds)
        let parsedUpdatedAt: Int = {
            let raw = dict["updatedAt"] as? Int ?? dict["updated_at"] as? Int ?? Int(Date().timeIntervalSince1970)
            return raw > 10_000_000_000 ? Int(raw / 1000) : raw
        }()

        // 6. Items Summary
        let parsedItemsSummary: String = {
            if let s = dict["itemsSummary"] as? String, !s.isEmpty { return s }
            if let s = dict["items_summary"] as? String, !s.isEmpty { return s }
            if let itemsList = dict["items"] as? [[String: Any]] {
                let parts = itemsList.compactMap { item -> String? in
                    let name = item["name"] as? String ?? item["title"] as? String ?? "Item"
                    let qty = item["quantity"] as? Int ?? item["qty"] as? Int ?? 1
                    return "\(qty)× \(name)"
                }
                return parts.joined(separator: ", ")
            }
            return ""
        }()

        self.startLiveActivity(
            orderId: fallbackOrderId,
            orderNumber: parsedOrderNumber,
            shopName: parsedShopName,
            itemsSummary: parsedItemsSummary,
            status: parsedStatus,
            etaMinutes: parsedETA,
            updatedAt: parsedUpdatedAt
        )
    }

    // MARK: - Step 3: Start ActivityKit Live Activity
    private func startLiveActivity(
        orderId: String,
        orderNumber: String,
        shopName: String,
        itemsSummary: String,
        status: String,
        etaMinutes: Int,
        updatedAt: Int
    ) {
        guard ActivityAuthorizationInfo().areActivitiesEnabled else {
            self.errorMessage = "Live Activities are disabled in iPhone Settings. Enable them in Settings > Face ID & Passcode > Live Activities."
            return
        }

        cleanupStaleActivities()

        let attributes = NembusOrderAttributes(
            orderId: orderId,
            shopName: shopName,
            itemsSummary: itemsSummary
        )

        let contentState = NembusOrderAttributes.ContentState(
            orderNumber: orderNumber,
            status: status,
            etaMinutes: etaMinutes,
            updatedAt: updatedAt
        )

        var activity: Activity<NembusOrderAttributes>?
        var tokenRegistered = false

        do {
            activity = try Activity<NembusOrderAttributes>.request(
                attributes: attributes,
                content: .init(state: contentState, staleDate: nil),
                pushType: .token
            )
            tokenRegistered = true
            print("Successfully started Live Activity with pushType: .token")
        } catch {
            print("pushType: .token failed (\(error.localizedDescription)). Falling back to pushType: nil...")
            do {
                activity = try Activity<NembusOrderAttributes>.request(
                    attributes: attributes,
                    content: .init(state: contentState, staleDate: nil),
                    pushType: nil
                )
                print("Successfully started Live Activity with pushType: nil")
            } catch {
                print("Failed to start Live Activity: \(error.localizedDescription)")
                self.errorMessage = error.localizedDescription
                return
            }
        }

        guard let activeActivity = activity else { return }

        self.orderNumber = orderNumber
        self.shopName = shopName
        self.itemsSummary = itemsSummary
        self.status = status
        self.etaMinutes = etaMinutes
        self.activeActivity = activeActivity

        // Setup live timer state & schedule lock screen audio alarm
        let isAlreadyReady = status.lowercased() == "ready" || status.lowercased() == "picked_up"
        if isAlreadyReady {
            self.remainingSeconds = 0
            self.totalTimerSeconds = 0
            self.isTimerActive = false
            self.isOrderCooked = true
        } else if etaMinutes > 0 {
            let secs = etaMinutes * 60
            self.remainingSeconds = secs
            self.totalTimerSeconds = secs
            self.isTimerActive = true
            self.isOrderCooked = false
            self.alarmHasRung = false

            // Schedule audio alert notification to ring on Lock Screen & Open
            scheduleAlarmNotification(orderNumber: orderNumber, shopName: shopName, inSeconds: secs)
        } else {
            self.remainingSeconds = 0
            self.totalTimerSeconds = 0
            self.isTimerActive = false
            self.isOrderCooked = false
        }

        withAnimation {
            self.isSuccess = true
        }

        UINotificationFeedbackGenerator().notificationOccurred(.success)
        print("Successfully started Live Activity for \(orderNumber) (ID: \(activeActivity.id))")

        // Step 4: Register Push Token with NAC Backend (if push token was requested)
        if tokenRegistered {
            listenAndRegisterPushToken(activity: activeActivity, orderId: orderId)
        }
    }

    // MARK: - Step 4: POST /v1/orders/:orderId/push-token
    private func listenAndRegisterPushToken(activity: Activity<NembusOrderAttributes>, orderId: String) {
        Task {
            for await pushTokenData in activity.pushTokenUpdates {
                let tokenHex = pushTokenData.map { String(format: "%02x", $0) }.joined()
                print("Received Live Activity Push Token: \(tokenHex)")

                #if DEBUG
                let environment = "sandbox"
                #else
                let environment = "production"
                #endif

                await sendPushTokenToBackend(orderId: orderId, tokenHex: tokenHex, environment: environment)
            }
        }
    }

    private func sendPushTokenToBackend(orderId: String, tokenHex: String, environment: String) async {
        guard let url = URL(string: "\(apiBaseURL)/v1/orders/\(orderId)/push-token") else { return }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let payload = PushTokenPayload(token: tokenHex, platform: "live_activity", environment: environment)

        do {
            request.httpBody = try JSONEncoder().encode(payload)
            let (data, response) = try await URLSession.shared.data(for: request)
            if let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 {
                print("Live Activity Push Token successfully registered with NAC Backend!")
            } else {
                print("Push token registration response: \(String(data: data, encoding: .utf8) ?? "")")
            }
        } catch {
            print("Failed to send push token: \(error.localizedDescription)")
        }
    }

    private func cleanupStaleActivities() {
        for activity in Activity<NembusOrderAttributes>.activities {
            Task {
                await activity.end(nil, dismissalPolicy: .immediate)
            }
        }
    }
}
