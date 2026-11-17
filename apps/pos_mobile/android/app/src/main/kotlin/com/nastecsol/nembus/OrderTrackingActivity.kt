package com.nastecsol.nembus

import android.app.Activity
import android.appwidget.AppWidgetManager
import android.content.ComponentName
import android.content.pm.PackageManager
import android.media.AudioAttributes
import android.media.MediaPlayer
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.os.CountDownTimer
import android.view.View
import android.view.animation.Animation
import android.view.animation.RotateAnimation
import android.widget.Button
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ProgressBar
import android.widget.TextView
import android.widget.Toast
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL
import kotlin.concurrent.thread

class OrderTrackingActivity : Activity() {

    private val apiBaseUrl = "https://nac-production.up.railway.app"

    private var orderNumber = "A-1"
    private var shopName = "Demo Coffee"
    private var itemsSummary = ""
    private var status = "preparing"
    private var etaMinutes = 25

    private var remainingSeconds = 0
    private var totalSeconds = 0
    private var timer: CountDownTimer? = null
    private var mediaPlayer: MediaPlayer? = null

    // UI Elements
    private lateinit var tvShopName: TextView
    private lateinit var tvOrderNumber: TextView
    private lateinit var tvItemsSummary: TextView
    private lateinit var tvCountdown: TextView
    private lateinit var progressBar: ProgressBar
    private lateinit var cardCooking: LinearLayout
    private lateinit var cardReady: LinearLayout
    private lateinit var ivPanAnimation: ImageView
    private lateinit var tvReadyOrderNum: TextView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_order_tracking)

        initViews()
        setupListeners()
        requestNotificationPermissionIfNeeded()
        handleIncomingIntent()
    }

    private fun requestNotificationPermissionIfNeeded() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (checkSelfPermission(android.Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
                requestPermissions(arrayOf(android.Manifest.permission.POST_NOTIFICATIONS), 101)
            }
        }
    }

    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<out String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == 101 && grantResults.isNotEmpty() && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
            updateTimerUI()
        }
    }

    private fun initViews() {
        tvShopName = findViewById(R.id.tv_main_shop_name)
        tvOrderNumber = findViewById(R.id.tv_main_order_number)
        tvItemsSummary = findViewById(R.id.tv_main_items_summary)
        tvCountdown = findViewById(R.id.tv_countdown)
        progressBar = findViewById(R.id.progress_bar_cooking)
        cardCooking = findViewById(R.id.card_cooking)
        cardReady = findViewById(R.id.card_ready)
        ivPanAnimation = findViewById(R.id.iv_pan_animation)
        tvReadyOrderNum = findViewById(R.id.tv_ready_order_num)

        startPanAnimation()
    }

    private fun startPanAnimation() {
        val rotate = RotateAnimation(
            -5f, 5f,
            Animation.RELATIVE_TO_SELF, 0.5f,
            Animation.RELATIVE_TO_SELF, 0.5f
        ).apply {
            duration = 800
            repeatMode = Animation.REVERSE
            repeatCount = Animation.INFINITE
        }
        ivPanAnimation.startAnimation(rotate)
    }

    private fun setupListeners() {
        findViewById<Button>(R.id.btn_simulate_ready).setOnClickListener {
            timer?.cancel()
            remainingSeconds = 0
            onTimerComplete()
        }

        findViewById<Button>(R.id.btn_test_1m).setOnClickListener {
            startTimer(60)
        }

        findViewById<Button>(R.id.btn_replay_alarm).setOnClickListener {
            playTimerAlarmSound()
        }

        findViewById<Button>(R.id.btn_pin_widget)?.setOnClickListener {
            pinWidgetToHomeScreen()
        }
    }

    private fun pinWidgetToHomeScreen() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val appWidgetManager = AppWidgetManager.getInstance(this)
            val myProvider = ComponentName(this, OrderWidgetProvider::class.java)
            if (appWidgetManager.isRequestPinAppWidgetSupported) {
                appWidgetManager.requestPinAppWidget(myProvider, null, null)
                Toast.makeText(this, "Order Widget added to Home Screen!", Toast.LENGTH_SHORT).show()
            } else {
                Toast.makeText(this, "Please add widget from your Home Screen launcher menu", Toast.LENGTH_LONG).show()
            }
        } else {
            Toast.makeText(this, "Please add widget from your Home Screen launcher menu", Toast.LENGTH_LONG).show()
        }
    }

    private fun handleIncomingIntent() {
        // 1. Direct Intent Extras (from Flutter or local launcher)
        val extraJson = intent?.getStringExtra("order_json")
        if (!extraJson.isNullOrEmpty()) {
            try {
                parseOrderJson(JSONObject(extraJson), "A-1")
                return
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }

        val data: Uri? = intent?.data
        if (data != null) {
            val urlString = data.toString()

            // 2. Query param ?data= or ?json=
            val queryData = data.getQueryParameter("data") ?: data.getQueryParameter("json")
            if (!queryData.isNullOrEmpty()) {
                try {
                    parseOrderJson(JSONObject(queryData), "A-1")
                    return
                } catch (e: Exception) {
                    e.printStackTrace()
                }
            }

            // 3. Query params: ?orderNumber=...&shopName=...&etaMinutes=...
            if (data.getQueryParameter("orderNumber") != null || data.getQueryParameter("order_number") != null) {
                val qOrderNum = data.getQueryParameter("orderNumber") ?: data.getQueryParameter("order_number") ?: "A-1"
                val qShopName = data.getQueryParameter("shopName") ?: data.getQueryParameter("shop_name") ?: "Demo Coffee"
                val qEta = data.getQueryParameter("etaMinutes")?.toIntOrNull() ?: data.getQueryParameter("eta_minutes")?.toIntOrNull() ?: 25
                val qStatus = data.getQueryParameter("status") ?: "preparing"

                val json = JSONObject().apply {
                    put("orderNumber", qOrderNum)
                    put("shopName", qShopName)
                    put("etaMinutes", qEta)
                    put("status", qStatus)
                }
                parseOrderJson(json, qOrderNum)
                return
            }

            // 4. Extract order ID from URL path (e.g. /v1/orders/:orderId, /orders/:orderId, /o/:orderId)
            val extractedId = extractOrderId(data)
            if (!extractedId.isNullOrEmpty()) {
                fetchOrderDetails(extractedId)
                return
            }

            // 5. Raw JSON substring in URL fallback
            if (urlString.contains("{") && urlString.contains("}")) {
                try {
                    val jsonStart = urlString.indexOf('{')
                    val jsonEnd = urlString.lastIndexOf('}') + 1
                    val json = JSONObject(urlString.substring(jsonStart, jsonEnd))
                    parseOrderJson(json, "A-1")
                    return
                } catch (e: Exception) {
                    e.printStackTrace()
                }
            }
        }

        // Default demo order
        startTimer(etaMinutes * 60)
    }

    private fun extractOrderId(uri: Uri): String? {
        // Query param check
        uri.getQueryParameter("orderId")?.let { if (it.isNotBlank()) return it }
        uri.getQueryParameter("order_id")?.let { if (it.isNotBlank()) return it }
        uri.getQueryParameter("orderNumber")?.let { if (it.isNotBlank()) return it }
        uri.getQueryParameter("id")?.let { if (it.isNotBlank()) return it }

        // Path segments check
        val segments = uri.pathSegments
        if (segments.isNullOrEmpty()) return null

        // Pattern matching: /v1/orders/:id, /orders/:id, /order/:id, /o/:id
        for (i in segments.indices) {
            val seg = segments[i].lowercase()
            if ((seg == "orders" || seg == "order" || seg == "o") && i + 1 < segments.size) {
                return segments[i + 1]
            }
        }

        // Return last segment if not a known prefix
        val last = segments.last()
        if (last.lowercase() !in listOf("o", "order", "orders", "v1", "api")) {
            return last
        }

        return null
    }

    private fun fetchOrderDetails(orderId: String) {
        orderNumber = orderId
        tvOrderNumber.text = "Order $orderNumber"
        tvReadyOrderNum.text = "Order #$orderNumber"
        tvItemsSummary.text = "Fetching order details..."

        // Start immediate lock screen & widget status while fetching
        startTimer(etaMinutes * 60)

        thread {
            try {
                val url = URL("$apiBaseUrl/v1/orders/$orderId")
                val conn = url.openConnection() as HttpURLConnection
                conn.requestMethod = "GET"
                conn.setRequestProperty("Accept", "application/json")
                conn.connectTimeout = 10000
                conn.readTimeout = 10000

                if (conn.responseCode in 200..299) {
                    val reader = BufferedReader(InputStreamReader(conn.inputStream))
                    val response = reader.readText()
                    reader.close()

                    val rawJson = JSONObject(response)
                    val orderJson = when {
                        rawJson.has("data") -> rawJson.getJSONObject("data")
                        rawJson.has("order") -> rawJson.getJSONObject("order")
                        else -> rawJson
                    }

                    runOnUiThread {
                        parseOrderJson(orderJson, orderId)
                    }
                } else {
                    runOnUiThread {
                        tvItemsSummary.text = "Order #$orderId in preparation"
                    }
                }
            } catch (e: Exception) {
                e.printStackTrace()
                runOnUiThread {
                    tvItemsSummary.text = "Tracking active for #$orderId"
                }
            }
        }
    }

    private fun parseOrderJson(json: JSONObject, fallbackId: String) {
        orderNumber = json.optString("orderNumber", json.optString("order_number", fallbackId))
        shopName = json.optString("shopName", json.optString("shop_name", "Demo Coffee"))
        status = json.optString("status", "preparing")
        etaMinutes = json.optInt("etaMinutes", json.optInt("eta_minutes", 25))

        tvShopName.text = shopName
        tvOrderNumber.text = "Order $orderNumber"
        tvReadyOrderNum.text = "Order #$orderNumber"

        if (status.equals("ready", ignoreCase = true) || status.equals("picked_up", ignoreCase = true)) {
            onTimerComplete()
        } else {
            startTimer(maxOf(etaMinutes, 1) * 60)
        }
    }

    private fun startTimer(seconds: Int) {
        timer?.cancel()
        remainingSeconds = seconds
        totalSeconds = seconds

        cardCooking.visibility = View.VISIBLE
        cardReady.visibility = View.GONE

        timer = object : CountDownTimer((seconds * 1000).toLong(), 1000) {
            override fun onTick(millisUntilFinished: Long) {
                remainingSeconds = (millisUntilFinished / 1000).toInt()
                updateTimerUI()
            }

            override fun onFinish() {
                remainingSeconds = 0
                onTimerComplete()
            }
        }.start()

        updateTimerUI()
    }

    private fun updateTimerUI() {
        val m = remainingSeconds / 60
        val s = remainingSeconds % 60
        val formatted = String.format("%02d:%02d", m, s)

        tvCountdown.text = formatted

        val progress = if (totalSeconds > 0) {
            (((totalSeconds - remainingSeconds).toDouble() / totalSeconds) * 100).toInt()
        } else 0
        progressBar.progress = progress

        // Update Lock Screen Notification
        OrderNotificationManager.showCookingNotification(
            this, orderNumber, shopName, remainingSeconds, totalSeconds
        )

        // Update Home Screen Widget
        OrderWidgetProvider.updateAllWidgets(
            this, orderNumber, shopName, formatted, progress, false
        )
    }

    private fun onTimerComplete() {
        cardCooking.visibility = View.GONE
        cardReady.visibility = View.VISIBLE

        // 1. Auto-play timer.mp3 ringtone
        playTimerAlarmSound()

        // 2. Fire Lock Screen notification with timer.mp3 sound
        OrderNotificationManager.showOrderCookedNotification(this, orderNumber, shopName)

        // 3. Update Home Screen Widget to ready
        OrderWidgetProvider.updateAllWidgets(
            this, orderNumber, shopName, "READY", 100, true
        )
    }

    private fun playTimerAlarmSound() {
        try {
            mediaPlayer?.release()
            mediaPlayer = MediaPlayer.create(this, R.raw.timer)?.apply {
                setAudioAttributes(
                    AudioAttributes.Builder()
                        .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION)
                        .setUsage(AudioAttributes.USAGE_ALARM)
                        .build()
                )
                start()
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        timer?.cancel()
        mediaPlayer?.release()
    }
}
