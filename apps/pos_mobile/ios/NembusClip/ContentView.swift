//
//  ContentView.swift
//  NembusClip
//  Robust NAC Backend Client for App Clip & Live Activities
//

import SwiftUI
import ActivityKit

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

    private let apiBaseURL = "https://nac-production.up.railway.app"

    var body: some View {
        ZStack {
            Color(.systemBackground).ignoresSafeArea()

            VStack(spacing: 24) {
                Spacer()

                if isSuccess {
                    // MARK: - Success: Added to Lock Screen
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 80))
                        .foregroundColor(.green)
                        .transition(.scale.combined(with: .opacity))

                    VStack(spacing: 8) {
                        Text("Order \(orderNumber)")
                            .font(.system(size: 32, weight: .heavy))

                        Text("Added to Lock Screen!")
                            .font(.title3)
                            .fontWeight(.semibold)

                        if !itemsSummary.isEmpty {
                            Text(itemsSummary)
                                .font(.headline)
                                .foregroundColor(.secondary)
                        }

                        if !shopName.isEmpty {
                            Text(shopName)
                                .font(.subheadline)
                                .foregroundColor(.secondary)
                        }

                        Text("Lock your iPhone to view live preparation status and updates.")
                            .font(.footnote)
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 24)
                            .padding(.top, 8)
                    }

                    HStack(spacing: 8) {
                        Image(systemName: "lock.shield.fill")
                            .foregroundColor(.blue)
                        Text("Live Activity Active")
                            .font(.subheadline)
                            .fontWeight(.bold)
                            .foregroundColor(.blue)
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 8)
                    .background(Color.blue.opacity(0.12))
                    .cornerRadius(20)
                    .padding(.top, 8)

                } else if isLoading {
                    // MARK: - Loading State
                    ProgressView()
                        .scaleEffect(1.6)
                        .padding(.bottom, 12)

                    Text("Fetching Order Details...")
                        .font(.headline)
                    Text("Connecting to server...")
                        .font(.caption)
                        .foregroundColor(.secondary)

                } else if let error = errorMessage {
                    // MARK: - Error State
                    Image(systemName: "exclamationmark.triangle.fill")
                        .font(.system(size: 64))
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

                } else {
                    // MARK: - Waiting for QR Scan
                    Image(systemName: "qrcode.viewfinder")
                        .font(.system(size: 80))
                        .foregroundColor(.blue)

                    VStack(spacing: 8) {
                        Text("Scan an Order QR Code")
                            .font(.title3)
                            .fontWeight(.bold)

                        Text("Point your iPhone camera at the QR code to track your order on the Lock Screen.")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 32)
                    }
                }

                Spacer()
            }
            .padding()
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
    }

    // MARK: - Step 1: Extract orderId from QR URL
    private func handleIncomingURL(_ url: URL) {
        print("Handling QR URL: \(url.absoluteString)")

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

    // MARK: - Step 2: GET /v1/orders/:orderId (Fail-safe, resilient parser)
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

                // Log raw JSON response for debugging
                let rawJSON = String(data: data, encoding: .utf8) ?? ""
                print("RAW API RESPONSE: \(rawJSON)")

                // Flexible dictionary parsing so missing keys never crash or prevent widget creation
                guard let rawDict = (try? JSONSerialization.jsonObject(with: data)) as? [String: Any] else {
                    self.errorMessage = "Could not parse JSON response from server."
                    return
                }

                // Handle both flat JSON and wrapped { "order": { ... } } or { "data": { ... } }
                let dict: [String: Any] = (rawDict["order"] as? [String: Any]) ??
                                          (rawDict["data"] as? [String: Any]) ??
                                          rawDict

                // 1. Order Number (handle both camelCase, snake_case, string and int)
                let parsedOrderNumber: String = {
                    if let s = dict["orderNumber"] as? String, !s.isEmpty { return s }
                    if let s = dict["order_number"] as? String, !s.isEmpty { return s }
                    if let i = dict["orderNumber"] as? Int { return "A-\(i)" }
                    if let i = dict["order_number"] as? Int { return "A-\(i)" }
                    return orderId
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
                    return "received"
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

                // 6. Items Summary (parse string or parse items array)
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

                // Start Live Activity with extracted values!
                self.startLiveActivity(
                    orderId: orderId,
                    orderNumber: parsedOrderNumber,
                    shopName: parsedShopName,
                    itemsSummary: parsedItemsSummary,
                    status: parsedStatus,
                    etaMinutes: parsedETA,
                    updatedAt: parsedUpdatedAt
                )
            }
        }.resume()
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
