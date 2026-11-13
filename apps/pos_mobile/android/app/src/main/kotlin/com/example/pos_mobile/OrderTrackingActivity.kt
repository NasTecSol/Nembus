package com.example.pos_mobile

import android.app.Activity
import android.media.AudioAttributes
import android.media.MediaPlayer
import android.net.Uri
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
        handleIncomingIntent()
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
    }

    private fun handleIncomingIntent() {
        val data: Uri? = intent?.data
        if (data != null) {
            val urlString = data.toString()
            // Check direct JSON in URL
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

            // Path: /o/:orderId
            val pathSegments = data.pathSegments
            val oIndex = pathSegments.indexOf("o")
            if (oIndex != -1 && oIndex + 1 < pathSegments.size) {
                val orderId = pathSegments[oIndex + 1]
                fetchOrderDetails(orderId)
                return
            }
        }

        // Default demo order
        startTimer(etaMinutes * 60)
    }

    private fun fetchOrderDetails(orderId: String) {
        thread {
            try {
                val url = URL("$apiBaseUrl/v1/orders/$orderId")
                val conn = url.openConnection() as HttpURLConnection
                conn.requestMethod = "GET"
                conn.connectTimeout = 8000
                conn.readTimeout = 8000

                if (conn.responseCode in 200..299) {
                    val reader = BufferedReader(InputStreamReader(conn.inputStream))
                    val response = reader.readText()
                    reader.close()

                    val rawJson = JSONObject(response)
                    val orderJson = if (rawJson.has("order")) rawJson.getJSONObject("order") else rawJson

                    runOnUiThread {
                        parseOrderJson(orderJson, orderId)
                    }
                }
            } catch (e: Exception) {
                e.printStackTrace()
                runOnUiThread {
                    // Fallback to initial demo
                    startTimer(25 * 60)
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
