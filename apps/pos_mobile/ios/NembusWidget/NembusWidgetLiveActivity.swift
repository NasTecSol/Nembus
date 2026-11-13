//
//  NembusWidgetLiveActivity.swift
//  NembusWidget
//  Clean, Regular & Uncongested UI/UX with Live Cooking Countdown & Ready State
//

import ActivityKit
import WidgetKit
import SwiftUI

struct NembusWidgetLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: NembusOrderAttributes.self) { context in
            // MARK: - Lock Screen / Banner Widget View
            LockScreenOrderView(context: context)
                .activityBackgroundTint(Color(red: 0.08, green: 0.10, blue: 0.16))
                .activitySystemActionForegroundColor(Color.white)

        } dynamicIsland: { context in
            // MARK: - Dynamic Island View (iPhone 14 Pro, 15, 16+)
            DynamicIsland {
                // Expanded Leading
                DynamicIslandExpandedRegion(.leading) {
                    VStack(alignment: .leading, spacing: 2) {
                        HStack(spacing: 5) {
                            Image(systemName: "cup.and.saucer.fill")
                                .font(.caption.bold())
                                .foregroundColor(.cyan)
                            Text(context.attributes.shopName.isEmpty ? "Demo Coffee" : context.attributes.shopName)
                                .font(.caption.bold())
                                .foregroundColor(.white.opacity(0.85))
                                .lineLimit(1)
                        }
                        Text("Order \(context.state.orderNumber)")
                            .font(.system(size: 18, weight: .heavy, design: .rounded))
                            .foregroundColor(.white)
                    }
                    .padding(.leading, 4)
                }

                // Expanded Trailing (Timer or Ready Status)
                DynamicIslandExpandedRegion(.trailing) {
                    VStack(alignment: .trailing, spacing: 2) {
                        let isReady = context.state.status.lowercased() == "ready" || context.state.status.lowercased() == "picked_up"

                        if isReady {
                            HStack(spacing: 4) {
                                Image(systemName: "bell.and.waves.left.and.right.fill")
                                    .symbolEffect(.variableColor.iterative)
                                    .foregroundColor(.green)
                                Text("READY")
                                    .font(.system(size: 13, weight: .heavy))
                                    .foregroundColor(.green)
                            }
                        } else if context.state.etaMinutes > 0 {
                            let totalSec = max(context.state.etaMinutes, 1) * 60
                            let rawStart = Date(timeIntervalSince1970: TimeInterval(context.state.updatedAt))
                            let startDate = (rawStart > Date() || Date().timeIntervalSince(rawStart) > 86400) ? Date() : rawStart
                            let calcEnd = startDate.addingTimeInterval(TimeInterval(totalSec))
                            let endDate = calcEnd > Date() ? calcEnd : Date().addingTimeInterval(TimeInterval(totalSec))

                            HStack(spacing: 4) {
                                Image(systemName: "flame.fill")
                                    .symbolEffect(.pulse)
                                    .foregroundColor(.orange)
                                Text(timerInterval: Date()...endDate, countsDown: true)
                                    .font(.system(size: 14, weight: .bold, design: .rounded))
                                    .monospacedDigit()
                                    .foregroundColor(.orange)
                            }
                        } else {
                            Text(context.state.status.capitalized)
                                .font(.caption.bold())
                                .foregroundColor(.cyan)
                        }
                    }
                    .padding(.trailing, 4)
                }

                // Expanded Bottom (Progress Bar & Status Message)
                DynamicIslandExpandedRegion(.bottom) {
                    VStack(spacing: 6) {
                        let isReady = context.state.status.lowercased() == "ready" || context.state.status.lowercased() == "picked_up"

                        if isReady {
                            HStack(spacing: 6) {
                                Image(systemName: "bell.and.waves.left.and.right.fill")
                                    .font(.system(size: 13, weight: .bold))
                                    .symbolEffect(.variableColor.iterative)
                                    .foregroundColor(.green)

                                Text("Order is cooked! You can collect from counter")
                                    .font(.system(size: 12, weight: .bold))
                                    .foregroundColor(.green)
                                    .lineLimit(1)
                            }
                            .padding(.vertical, 4)
                            .frame(maxWidth: .infinity)
                            .background(Color.green.opacity(0.15))
                            .cornerRadius(8)
                        } else {
                            let totalSec = max(context.state.etaMinutes, 1) * 60
                            let rawStart = Date(timeIntervalSince1970: TimeInterval(context.state.updatedAt))
                            let startDate = (rawStart > Date() || Date().timeIntervalSince(rawStart) > 86400) ? Date() : rawStart
                            let calcEnd = startDate.addingTimeInterval(TimeInterval(totalSec))
                            let endDate = calcEnd > Date() ? calcEnd : Date().addingTimeInterval(TimeInterval(totalSec))

                            VStack(spacing: 4) {
                                ProgressView(timerInterval: startDate...endDate, countsDown: false)
                                    .tint(LinearGradient(colors: [.orange, .yellow], startPoint: .leading, endPoint: .trailing))

                                HStack {
                                    HStack(spacing: 4) {
                                        Image(systemName: "flame.fill")
                                            .font(.system(size: 9))
                                            .symbolEffect(.pulse)
                                        Text("Cooking in progress...")
                                            .font(.caption2.bold())
                                            .foregroundColor(.orange)
                                    }
                                    Spacer()
                                    Text("~\(context.state.etaMinutes)m")
                                        .font(.caption2.bold())
                                        .foregroundColor(.white.opacity(0.7))
                                }
                            }
                        }
                    }
                    .padding(.horizontal, 4)
                    .padding(.top, 2)
                }
            } compactLeading: {
                let isReady = context.state.status.lowercased() == "ready" || context.state.status.lowercased() == "picked_up"
                if isReady {
                    Image(systemName: "bell.badge.fill")
                        .symbolEffect(.bounce)
                        .foregroundColor(.green)
                } else {
                    Image(systemName: "flame.fill")
                        .symbolEffect(.pulse)
                        .foregroundColor(.orange)
                }
            } compactTrailing: {
                let isReady = context.state.status.lowercased() == "ready" || context.state.status.lowercased() == "picked_up"
                if isReady {
                    Text("READY")
                        .font(.system(size: 11, weight: .heavy))
                        .foregroundColor(.green)
                } else if context.state.etaMinutes > 0 {
                    let totalSec = max(context.state.etaMinutes, 1) * 60
                    let rawStart = Date(timeIntervalSince1970: TimeInterval(context.state.updatedAt))
                    let startDate = (rawStart > Date() || Date().timeIntervalSince(rawStart) > 86400) ? Date() : rawStart
                    let calcEnd = startDate.addingTimeInterval(TimeInterval(totalSec))
                    let endDate = calcEnd > Date() ? calcEnd : Date().addingTimeInterval(TimeInterval(totalSec))

                    Text(endDate, style: .timer)
                        .font(.system(size: 12, weight: .bold, design: .rounded))
                        .monospacedDigit()
                        .foregroundColor(.orange)
                        .frame(width: 42)
                } else {
                    Text(context.state.orderNumber)
                        .font(.system(size: 11, weight: .bold))
                        .foregroundColor(.cyan)
                }
            } minimal: {
                let isReady = context.state.status.lowercased() == "ready" || context.state.status.lowercased() == "picked_up"
                if isReady {
                    Image(systemName: "bell.badge.fill")
                        .foregroundColor(.green)
                } else {
                    Image(systemName: "flame.fill")
                        .foregroundColor(.orange)
                }
            }
            .keylineTint(Color.orange)
        }
    }
}

// MARK: - Lock Screen View Implementation
private struct LockScreenOrderView: View {
    let context: ActivityViewContext<NembusOrderAttributes>

    private var isReady: Bool {
        context.state.status.lowercased() == "ready" || context.state.status.lowercased() == "picked_up"
    }

    private var startDate: Date {
        let raw = Date(timeIntervalSince1970: TimeInterval(context.state.updatedAt))
        let now = Date()
        if raw > now || now.timeIntervalSince(raw) > 86400 {
            return now
        }
        return raw
    }

    private var endDate: Date {
        let totalSec = max(context.state.etaMinutes, 1) * 60
        let calcEnd = startDate.addingTimeInterval(TimeInterval(totalSec))
        return calcEnd > Date() ? calcEnd : Date().addingTimeInterval(TimeInterval(totalSec))
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            // MARK: - Row 1: Header (Shop Name & Order Pill)
            HStack(alignment: .center) {
                HStack(spacing: 7) {
                    Image(systemName: "cup.and.saucer.fill")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundColor(.cyan)

                    Text(context.attributes.shopName.isEmpty ? "Demo Coffee" : context.attributes.shopName)
                        .font(.system(size: 15, weight: .bold))
                        .foregroundColor(.white)
                        .lineLimit(1)
                }

                Spacer()

                // Sleek Order Number Pill
                Text("Order \(context.state.orderNumber)")
                    .font(.system(size: 12, weight: .heavy, design: .rounded))
                    .foregroundColor(.white)
                    .padding(.horizontal, 9)
                    .padding(.vertical, 4)
                    .background(Color.white.opacity(0.12))
                    .cornerRadius(6)
            }

            // MARK: - Row 2: Status & Timer (Spacious, Regular & Uncongested)
            if isReady {
                // Cooked & Ready State (Clean layout without buttons)
                HStack(alignment: .center, spacing: 10) {
                    ZStack {
                        Circle()
                            .fill(Color.green.opacity(0.2))
                            .frame(width: 36, height: 36)
                        Image(systemName: "bell.and.waves.left.and.right.fill")
                            .font(.system(size: 17, weight: .bold))
                            .symbolEffect(.variableColor.iterative)
                            .foregroundColor(.green)
                    }

                    VStack(alignment: .leading, spacing: 1) {
                        Text("ORDER IS COOKED!")
                            .font(.system(size: 15, weight: .heavy, design: .rounded))
                            .foregroundColor(.green)
                        Text("You can collect from counter")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundColor(.white.opacity(0.95))
                            .lineLimit(1)
                    }

                    Spacer()

                    Image(systemName: "checkmark.seal.fill")
                        .font(.system(size: 22))
                        .foregroundColor(.green)
                }
                .padding(.vertical, 2)

            } else {
                // Active Cooking Row
                HStack(alignment: .center) {
                    // Left: Cooking Animation & Status Text
                    HStack(spacing: 8) {
                        Image(systemName: "flame.fill")
                            .font(.system(size: 16, weight: .bold))
                            .symbolEffect(.pulse)
                            .foregroundColor(.orange)

                        VStack(alignment: .leading, spacing: 1) {
                            Text("Cooking in Progress")
                                .font(.system(size: 14, weight: .bold))
                                .foregroundColor(.white)
                                .lineLimit(1)
                            Text("Chef is preparing your order")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(.white.opacity(0.65))
                                .lineLimit(1)
                        }
                    }

                    Spacer(minLength: 8)

                    // Right: Regular sized countdown timer
                    if context.state.etaMinutes > 0 {
                        HStack(spacing: 4) {
                            Image(systemName: "timer")
                                .font(.system(size: 12, weight: .semibold))
                                .foregroundColor(.orange)
                            Text(timerInterval: Date()...endDate, countsDown: true)
                                .font(.system(size: 19, weight: .heavy, design: .rounded))
                                .monospacedDigit()
                                .foregroundColor(.orange)
                        }
                    }
                }
                .padding(.vertical, 2)
            }

            // MARK: - Row 3: Progress Bar (Clean, Sleek & Never Overflowing)
            if isReady {
                ProgressView(value: 1.0)
                    .tint(Color.green)
                    .padding(.top, 1)
            } else if context.state.etaMinutes > 0 {
                ProgressView(timerInterval: startDate...endDate, countsDown: false)
                    .tint(LinearGradient(colors: [.orange, .yellow], startPoint: .leading, endPoint: .trailing))
                    .padding(.top, 1)
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
    }
}
