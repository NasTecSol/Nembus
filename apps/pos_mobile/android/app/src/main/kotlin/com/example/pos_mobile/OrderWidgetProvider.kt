package com.example.pos_mobile

import android.app.PendingIntent
import android.appwidget.AppWidgetManager
import android.appwidget.AppWidgetProvider
import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.os.Build
import android.view.View
import android.widget.RemoteViews

class OrderWidgetProvider : AppWidgetProvider() {

    override fun onUpdate(context: Context, appWidgetManager: AppWidgetManager, appWidgetIds: IntArray) {
        for (appWidgetId in appWidgetIds) {
            updateWidget(context, appWidgetManager, appWidgetId, "A-1", "Demo Coffee", "25:00", 25, false)
        }
    }

    companion object {
        fun updateAllWidgets(
            context: Context,
            orderNumber: String,
            shopName: String,
            timeFormatted: String,
            progressPercent: Int,
            isCooked: Boolean
        ) {
            val appWidgetManager = AppWidgetManager.getInstance(context)
            val componentName = ComponentName(context, OrderWidgetProvider::class.java)
            val appWidgetIds = appWidgetManager.getAppWidgetIds(componentName)

            for (appWidgetId in appWidgetIds) {
                updateWidget(
                    context, appWidgetManager, appWidgetId,
                    orderNumber, shopName, timeFormatted, progressPercent, isCooked
                )
            }
        }

        private fun updateWidget(
            context: Context,
            appWidgetManager: AppWidgetManager,
            appWidgetId: Int,
            orderNumber: String,
            shopName: String,
            timeFormatted: String,
            progressPercent: Int,
            isCooked: Boolean
        ) {
            val views = RemoteViews(context.packageName, R.layout.order_widget_layout)

            views.setTextViewText(R.id.tv_shop_name, shopName.ifEmpty { "Demo Coffee" })
            views.setTextViewText(R.id.tv_order_number, "Order $orderNumber")

            if (isCooked) {
                views.setViewVisibility(R.id.layout_cooking, View.GONE)
                views.setViewVisibility(R.id.layout_ready, View.VISIBLE)
                views.setProgressBar(R.id.progress_order, 100, 100, false)
            } else {
                views.setViewVisibility(R.id.layout_cooking, View.VISIBLE)
                views.setViewVisibility(R.id.layout_ready, View.GONE)
                views.setTextViewText(R.id.tv_timer, timeFormatted)
                views.setProgressBar(R.id.progress_order, 100, progressPercent, false)
            }

            // Clicking widget opens the order tracker
            val intent = Intent(context, OrderTrackingActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            val pendingIntent = PendingIntent.getActivity(
                context, 0, intent,
                PendingIntent.FLAG_UPDATE_CURRENT or (if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) PendingIntent.FLAG_IMMUTABLE else 0)
            )
            views.setOnClickPendingIntent(R.id.widget_container, pendingIntent)

            appWidgetManager.updateAppWidget(appWidgetId, views)
        }
    }
}
