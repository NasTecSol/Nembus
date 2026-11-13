package com.example.pos_mobile

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.media.AudioAttributes
import android.net.Uri
import android.os.Build
import androidx.core.app.NotificationCompat

object OrderNotificationManager {
    const val TRACKING_CHANNEL_ID = "order_tracking_channel"
    const val ALARM_CHANNEL_ID = "order_ready_alarm_channel"
    const val NOTIFICATION_ID = 1001

    fun createChannels(context: Context) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

            // 1. Live Ongoing Cooking Channel (Lock Screen Live Activity)
            val trackingChannel = NotificationChannel(
                TRACKING_CHANNEL_ID,
                "Live Order Tracking",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "Shows live cooking countdown and status on Lock Screen"
                lockscreenVisibility = Notification.VISIBILITY_PUBLIC
                setShowBadge(false)
            }

            // 2. Alarm Ready Channel (plays timer.mp3 aloud when order is cooked)
            val soundUri = Uri.parse("android.resource://${context.packageName}/raw/timer")
            val audioAttributes = AudioAttributes.Builder()
                .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION)
                .setUsage(AudioAttributes.USAGE_ALARM)
                .build()

            val alarmChannel = NotificationChannel(
                ALARM_CHANNEL_ID,
                "Order Cooked Alarm",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Plays alarm ringtone when order is cooked"
                setSound(soundUri, audioAttributes)
                enableVibration(true)
                vibrationPattern = longArrayOf(0, 500, 200, 500)
                lockscreenVisibility = Notification.VISIBILITY_PUBLIC
            }

            manager.createNotificationChannel(trackingChannel)
            manager.createNotificationChannel(alarmChannel)
        }
    }

    fun showCookingNotification(
        context: Context,
        orderNumber: String,
        shopName: String,
        remainingSeconds: Int,
        totalSeconds: Int
    ) {
        createChannels(context)

        val intent = Intent(context, OrderTrackingActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            context, 0, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or (if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) PendingIntent.FLAG_IMMUTABLE else 0)
        )

        val mins = remainingSeconds / 60
        val secs = remainingSeconds % 60
        val timeFormatted = String.format("%02d:%02d", mins, secs)

        val progress = if (totalSeconds > 0) {
            (((totalSeconds - remainingSeconds).toDouble() / totalSeconds) * 100).toInt()
        } else 0

        val builder = NotificationCompat.Builder(context, TRACKING_CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_flame)
            .setContentTitle("$shopName • Order $orderNumber")
            .setContentText("Cooking in Progress • $timeFormatted remaining")
            .setProgress(100, progress, false)
            .setOngoing(true) // Pinned to lock screen like Live Activity
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .setContentIntent(pendingIntent)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setOnlyAlertOnce(true)

        val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        manager.notify(NOTIFICATION_ID, builder.build())
    }

    fun showOrderCookedNotification(context: Context, orderNumber: String, shopName: String) {
        createChannels(context)

        val intent = Intent(context, OrderTrackingActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            context, 0, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or (if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) PendingIntent.FLAG_IMMUTABLE else 0)
        )

        val soundUri = Uri.parse("android.resource://${context.packageName}/raw/timer")

        val builder = NotificationCompat.Builder(context, ALARM_CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_bell_ready)
            .setContentTitle("ORDER IS COOKED!")
            .setContentText("Order is cooked you can collect from counter at $shopName")
            .setSound(soundUri)
            .setVibrate(longArrayOf(0, 500, 200, 500))
            .setOngoing(false)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .setPriority(NotificationCompat.PRIORITY_MAX)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)

        val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        manager.notify(NOTIFICATION_ID, builder.build())
    }
}
