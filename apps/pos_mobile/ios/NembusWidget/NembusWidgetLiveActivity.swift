//
//  NembusWidgetLiveActivity.swift
//  NembusWidget
//

import ActivityKit
import WidgetKit
import SwiftUI

struct NembusWidgetLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: NembusOrderAttributes.self) { context in
            // MARK: - Lock Screen / Banner Widget View
            VStack(alignment: .leading, spacing: 12) {
                // Top row: Shop Name & ETA Badge
                HStack {
                    HStack(spacing: 6) {
                        Image(systemName: "cup.and.saucer.fill")
                            .font(.system(size: 14, weight: .bold))
                            .foregroundColor(.cyan)
                        Text(context.attributes.shopName.isEmpty ? "Coffee Shop" : context.attributes.shopName)
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundColor(.white)
                    }

                    Spacer()

                    if context.state.etaMinutes > 0 && context.state.status != "ready" && context.state.status != "picked_up" {
                        HStack(spacing: 4) {
                            Image(systemName: "clock.fill")
                                .font(.system(size: 11))
                            Text("~\(context.state.etaMinutes) min")
                                .font(.system(size: 12, weight: .bold))
                        }
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.orange.opacity(0.25))
                        .foregroundColor(.orange)
                        .cornerRadius(6)
                    }
                }

                // Middle: Order Number & Status
                HStack(alignment: .center) {
                    VStack(alignment: .leading, spacing: 3) {
                        Text("Order \(context.state.orderNumber)")
                            .font(.system(size: 22, weight: .heavy))
                            .foregroundColor(.white)

                        if !context.attributes.itemsSummary.isEmpty {
                            Text(context.attributes.itemsSummary)
                                .font(.system(size: 13, weight: .regular))
                                .foregroundColor(.white.opacity(0.8))
                                .lineLimit(1)
                        }
                    }

                    Spacer()

                    // Status Pill
                    statusBadge(for: context.state.status)
                }
            }
            .padding(16)
            .activityBackgroundTint(Color(red: 0.08, green: 0.12, blue: 0.20))
            .activitySystemActionForegroundColor(Color.white)

        } dynamicIsland: { context in
            // MARK: - Dynamic Island View
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Order \(context.state.orderNumber)")
                            .font(.headline)
                            .fontWeight(.bold)
                            .foregroundColor(.white)
                        Text(context.attributes.shopName)
                            .font(.caption2)
                            .foregroundColor(.secondary)
                    }
                }

                DynamicIslandExpandedRegion(.trailing) {
                    VStack(alignment: .trailing, spacing: 2) {
                        statusBadge(for: context.state.status)
                        if context.state.etaMinutes > 0 && context.state.status != "ready" {
                            Text("~\(context.state.etaMinutes) min")
                                .font(.caption2)
                                .foregroundColor(.orange)
                        }
                    }
                }

                DynamicIslandExpandedRegion(.bottom) {
                    if !context.attributes.itemsSummary.isEmpty {
                        Text(context.attributes.itemsSummary)
                            .font(.caption)
                            .foregroundColor(.white.opacity(0.85))
                            .lineLimit(1)
                    }
                }
            } compactLeading: {
                Image(systemName: "cup.and.saucer.fill")
                    .foregroundColor(.cyan)
            } compactTrailing: {
                Text(context.state.orderNumber)
                    .font(.caption2.bold())
                    .foregroundColor(.green)
            } minimal: {
                Image(systemName: "cup.and.saucer.fill")
                    .foregroundColor(.cyan)
            }
            .keylineTint(Color.cyan)
        }
    }

    // Helper: Status badge
    @ViewBuilder
    private func statusBadge(for status: String) -> some View {
        let (title, icon, color) = statusConfig(for: status)
        HStack(spacing: 5) {
            Image(systemName: icon)
                .font(.system(size: 11, weight: .bold))
            Text(title)
                .font(.system(size: 13, weight: .bold))
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 5)
        .background(color.opacity(0.2))
        .foregroundColor(color)
        .cornerRadius(8)
    }

    private func statusConfig(for status: String) -> (String, String, Color) {
        switch status.lowercased() {
        case "received":
            return ("Received", "tray.and.arrow.down.fill", .cyan)
        case "preparing":
            return ("Preparing", "flame.fill", .orange)
        case "ready":
            return ("Ready for Pickup", "checkmark.seal.fill", .green)
        case "picked_up":
            return ("Picked Up", "bag.fill", .gray)
        case "cancelled":
            return ("Cancelled", "xmark.circle.fill", .red)
        default:
            return (status.capitalized, "info.circle.fill", .blue)
        }
    }
}
