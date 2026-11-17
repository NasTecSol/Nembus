package com.nastecsol.nembus

import android.content.Intent
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {
    private val CHANNEL = "com.nastecsol.nembus/tracker"

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL).setMethodCallHandler { call, result ->
            when (call.method) {
                "openOrderTracker" -> {
                    val orderJson = call.argument<String>("order_json")
                    val intent = Intent(this, OrderTrackingActivity::class.java).apply {
                        if (!orderJson.isNullOrEmpty()) {
                            putExtra("order_json", orderJson)
                        }
                    }
                    startActivity(intent)
                    result.success(true)
                }
                "startLockScreenTimer" -> {
                    val orderNumber = call.argument<String>("orderNumber") ?: "A-1"
                    val shopName = call.argument<String>("shopName") ?: "Demo Coffee"
                    val etaMinutes = call.argument<Int>("etaMinutes") ?: 25
                    val totalSecs = etaMinutes * 60
                    OrderNotificationManager.showCookingNotification(this, orderNumber, shopName, totalSecs, totalSecs)
                    result.success(true)
                }
                else -> result.notImplemented()
            }
        }
    }
}
