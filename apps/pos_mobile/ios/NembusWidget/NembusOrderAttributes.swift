//
//  NembusOrderAttributes.swift
//  Shared between NembusWidget and NembusClip
//  Matches backend schemas/live-activity-content-state.json
//

import ActivityKit
import Foundation

public struct NembusOrderAttributes: ActivityAttributes, Codable {
    public struct ContentState: Codable, Hashable {
        // Customer-facing order number (e.g. "A-42")
        public var orderNumber: String
        // Status enum: "received" | "preparing" | "ready" | "picked_up" | "cancelled"
        public var status: String
        // ETA in minutes, 0 if unknown
        public var etaMinutes: Int
        // Unix timestamp in seconds
        public var updatedAt: Int

        public init(orderNumber: String, status: String, etaMinutes: Int = 0, updatedAt: Int = Int(Date().timeIntervalSince1970)) {
            self.orderNumber = orderNumber
            self.status = status
            self.etaMinutes = etaMinutes
            self.updatedAt = updatedAt
        }
    }

    // Static properties set at start - MUST be Codable
    public var orderId: String         // Internal nanoid
    public var shopName: String        // e.g. "Demo Coffee"
    public var itemsSummary: String    // e.g. "1× Latte"

    public init(orderId: String, shopName: String, itemsSummary: String) {
        self.orderId = orderId
        self.shopName = shopName
        self.itemsSummary = itemsSummary
    }
}
