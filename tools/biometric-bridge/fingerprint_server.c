/*
 * fingerprint_server.c - RESTful HTTP API Server  (v2.0)
 *
 * Implements a standards-compliant RESTful HTTP/1.1 server using GNU
 * libmicrohttpd. A single port exposes all endpoints under the /api/v1
 * prefix, following REST conventions for URL design, HTTP verbs, and
 * status codes.
 *
 * ╔══════════════════════════════════════════════════════════════════════╗
 * ║ API SURFACE  (base URL: http://127.0.0.1:5050)                      ║
 * ╠══════════════════════════════════════════════════════════════════════╣
 * ║ GET    /api/v1/device         Reader connection status               ║
 * ║ GET    /api/v1/device/info    Device metadata (name, DPI, serial)   ║
 * ║ POST   /api/v1/capture        Capture image → Base64 BMP + FMD      ║
 * ║ DELETE /api/v1/capture        Cancel an in-progress session          ║
 * ╚══════════════════════════════════════════════════════════════════════╝
 *
 * JSON RESPONSE ENVELOPE
 *   Success : { "success": true,  "data":  { ... } }
 *   Error   : { "success": false, "error": { "code": <int>, "message": "<str>" } }
 *
 * HTTP STATUS CODES
 *   200 OK                   - Request succeeded
 *   204 No Content           - OPTIONS preflight
 *   404 Not Found            - Unknown endpoint
 *   405 Method Not Allowed   - Wrong HTTP verb for this endpoint
 *   500 Internal Server Error - SDK or system failure
 *   503 Service Unavailable  - No fingerprint reader connected
 *
 * Copyright (C) 2011 Digital Persona, Inc. (SDK); project contributors (microservice)
 */
#include "fingerprint_capture.h"
#include "fingerprint_selection.h"
#include <locale.h>
#include <pthread.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <microhttpd.h>
#include <time.h>

/* ── Configuration ────────────────────────────────────────────────────────── */
#define API_PORT         5050
#define CONNECTION_LIMIT 10
#define API_PREFIX       "/api/v1"
#define CONTENT_JSON     "application/json; charset=utf-8"

#ifndef DPFPDD_MAX_STR_LENGTH
#define DPFPDD_MAX_STR_LENGTH MAX_DEVICE_NAME_LENGTH
#endif

#if defined(MHD_VERSION) && (MHD_VERSION >= 0x00097002)
typedef enum MHD_Result mhd_result;
#else
typedef int mhd_result;
#endif

/* ── JSON / HTTP Response Helpers ─────────────────────────────────────────── */

/**
 * @brief Queue a JSON HTTP response with CORS and security headers.
 */
static mhd_result send_json(struct MHD_Connection *conn, unsigned int status,
                     const char *body) {
    struct MHD_Response *r =
        MHD_create_response_from_buffer(strlen(body), (void *)body,
                                        MHD_RESPMEM_MUST_COPY);
    if (!r) return MHD_NO;
    MHD_add_response_header(r, "Content-Type",                CONTENT_JSON);
    MHD_add_response_header(r, "Access-Control-Allow-Origin", "*");
    MHD_add_response_header(r, "Access-Control-Allow-Methods","GET, POST, DELETE, OPTIONS");
    MHD_add_response_header(r, "Access-Control-Allow-Headers","Content-Type");
    MHD_add_response_header(r, "X-Content-Type-Options",      "nosniff");
    mhd_result ret = (mhd_result)MHD_queue_response(conn, status, r);
    MHD_destroy_response(r);
    return ret;
}

/** Queue a JSON error envelope. */
static mhd_result send_error(struct MHD_Connection *conn, unsigned int status,
                      int code, const char *msg) {
    char buf[512];
    snprintf(buf, sizeof(buf),
             "{\"success\":false,\"error\":{\"code\":%d,\"message\":\"%s\"}}",
             code, msg);
    return send_json(conn, status, buf);
}

static mhd_result send_not_found(struct MHD_Connection *conn) {
    return send_error(conn, MHD_HTTP_NOT_FOUND, 404,
                      "The requested endpoint does not exist.");
}

static mhd_result send_method_not_allowed(struct MHD_Connection *conn) {
    return send_error(conn, MHD_HTTP_METHOD_NOT_ACCEPTABLE, 405,
                      "HTTP method not allowed for this endpoint.");
}

/* ── CORS Preflight ───────────────────────────────────────────────────────── */
static mhd_result handle_options(struct MHD_Connection *conn) {
    struct MHD_Response *r =
        MHD_create_response_from_buffer(0, (void *)"", MHD_RESPMEM_PERSISTENT);
    if (!r) return MHD_NO;
    MHD_add_response_header(r, "Access-Control-Allow-Origin",  "*");
    MHD_add_response_header(r, "Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS");
    MHD_add_response_header(r, "Access-Control-Allow-Headers", "Content-Type");
    MHD_add_response_header(r, "Access-Control-Max-Age",       "86400");
    mhd_result ret = (mhd_result)MHD_queue_response(conn, MHD_HTTP_NO_CONTENT, r);
    MHD_destroy_response(r);
    return ret;
}

/* ── SDK Lifecycle Helpers ────────────────────────────────────────────────── */

/**
 * @brief Initialise SDK and open the first connected reader.
 * Sends the appropriate HTTP error and returns 0 if the reader is unavailable.
 */
static int open_reader(struct MHD_Connection *conn,
                       DPFPDD_DEV *hReader, int *dpi,
                       char *szReader, size_t len) {
    sigset_t mask; sigfillset(&mask);
    pthread_sigmask(SIG_BLOCK, &mask, NULL);
    setlocale(LC_ALL, "");

    int rc = dpfpdd_init();
    if (DPFPDD_SUCCESS != rc) {
        fprintf(stderr, "[ERROR] dpfpdd_init() rc=%d\n", rc);
        send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                   "Failed to initialise the fingerprint SDK.");
        return 0;
    }

    *hReader = SelectAndOpenReader(szReader, len, dpi);
    if (!(*hReader)) {
        dpfpdd_exit();
        send_error(conn, MHD_HTTP_SERVICE_UNAVAILABLE, 503,
                   "No fingerprint reader detected. "
                   "Verify USB connection and udev rules.");
        return 0;
    }
    return 1;
}

/** Close reader handle and shut down the SDK. */
static void close_reader(DPFPDD_DEV *hReader) {
    if (hReader && *hReader) { dpfpdd_close(*hReader); *hReader = NULL; }
    dpfpdd_exit();
}

/* ── Endpoint: GET /api/v1/device ─────────────────────────────────────────── */
/**
 * Reports whether a fingerprint reader is connected and ready.
 *
 * 200 OK:  { "success": true,  "data": { "connected": true } }
 * 503:     { "success": false, "error": { "code": 503, "message": "..." } }
 */
static mhd_result handle_device_status(struct MHD_Connection *conn) {
    DPFPDD_DEV h = NULL; int dpi = 0;
    char szR[DPFPDD_MAX_STR_LENGTH] = {0};
    if (!open_reader(conn, &h, &dpi, szR, sizeof(szR))) return MHD_YES;
    close_reader(&h);
    return send_json(conn, MHD_HTTP_OK,
        "{\"success\":true,\"data\":{\"connected\":true}}");
}

/* ── Endpoint: GET /api/v1/device/info ────────────────────────────────────── */
/**
 * Returns metadata about the connected fingerprint reader.
 *
 * 200 OK:
 *   { "success": true, "data": {
 *       "connected": true, "device_name": "...",
 *       "serial_number": "...", "resolution_dpi": 500 } }
 */
static mhd_result handle_device_info(struct MHD_Connection *conn) {
    DPFPDD_DEV h = NULL; int dpi = 0;
    char szR[DPFPDD_MAX_STR_LENGTH] = {0};
    if (!open_reader(conn, &h, &dpi, szR, sizeof(szR))) return MHD_YES;

    char serial[DPFPDD_MAX_STR_LENGTH] = {0};
    QueryDeviceSerial(serial, sizeof(serial));
    close_reader(&h);

    char buf[2048];
    snprintf(buf, sizeof(buf),
             "{\"success\":true,\"data\":{"
             "\"connected\":true,"
             "\"device_name\":\"%s\","
             "\"serial_number\":\"%s\","
             "\"resolution_dpi\":%d}}",
             szR, serial, dpi);
    return send_json(conn, MHD_HTTP_OK, buf);
}

/* ── Endpoint: POST /api/v1/capture ───────────────────────────────────────── */
/**
 * Full fingerprint acquisition cycle:
 *   1. Open reader
 *   2. Block until finger placed; capture FID image
 *   3. Assess image quality via dpfj_get_fid_quality()
 *   4. Extract FMD template via dpfj_create_fmd_from_fid() (ISO 19794-2:2005)
 *   5. Serialize cropped BMP to disk; encode as Base64
 *   6. Return Base64 BMP image + Base64 FMD + quality + serial
 *
 * 200 OK:
 *   { "success": true, "data": {
 *       "image_base64": "...", "fmd_base64": "...",
 *       "fmd_format": "ISO_19794_2_2005",
 *       "quality": 87, "serial_number": "..." } }
 */
static mhd_result handle_capture(struct MHD_Connection *conn) {
    DPFPDD_DEV h = NULL; int dpi = 0;
    char szR[DPFPDD_MAX_STR_LENGTH] = {0};
    if (!open_reader(conn, &h, &dpi, szR, sizeof(szR))) return MHD_YES;

    char serial[DPFPDD_MAX_STR_LENGTH] = {0};
    QueryDeviceSerial(serial, sizeof(serial));

    FingerprintCaptureResult res; memset(&res, 0, sizeof(res)); res.quality = -1;

    int rc = CaptureFingerprintImage(h, dpi, &res);
    if (0 != rc) {
        close_reader(&h);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          res.error_msg[0] ? res.error_msg : "Fingerprint capture failed.");
    }

    AssessImageQuality(&res);

    rc = ExtractFMD(&res, DEFAULT_FMD_FORMAT);
    if (0 != rc) {
        FreeCaptureResult(&res); close_reader(&h);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          res.error_msg[0] ? res.error_msg : "FMD extraction failed.");
    }

    size_t img_b64_len = 0;
    char *img_b64 = SerializeBMPBase64(res.image_data,
                                      (int)res.image_width, (int)res.image_height,
                                      &img_b64_len);
    size_t fmd_b64_len = 0;
    char *fmd_b64 = Base64Encode(res.fmd_data, res.fmd_size, &fmd_b64_len);
    int quality   = res.quality;
    close_reader(&h);

    if (!img_b64 || !fmd_b64) {
        free(img_b64); free(fmd_b64); FreeCaptureResult(&res);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          "Base64 encoding failed.");
    }

    size_t jlen = img_b64_len + fmd_b64_len + 256;
    char *jbuf  = (char *)malloc(jlen);
    if (!jbuf) {
        free(img_b64); free(fmd_b64); FreeCaptureResult(&res);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          "Memory allocation failed for response buffer.");
    }

    snprintf(jbuf, jlen,
             "{\"success\":true,\"data\":{"
             "\"image_base64\":\"%s\","
             "\"fmd_base64\":\"%s\","
             "\"fmd_format\":\"ISO_19794_2_2005\","
             "\"quality\":%d,"
             "\"serial_number\":\"%s\"}}",
             img_b64, fmd_b64, quality, serial);

    free(img_b64); free(fmd_b64); FreeCaptureResult(&res);
    int ret = send_json(conn, MHD_HTTP_OK, jbuf);
    free(jbuf);
    return ret;
}


/* ── Endpoint: DELETE /api/v1/capture ─────────────────────────────────────── */
/**
 * Cancels an in-progress fingerprint capture session by sending SIGINT
 * to the tmux session hosting the blocking dpfpdd_capture() call.
 * The registered signal handler (capture_signal_handler) will invoke
 * dpfpdd_cancel() to unblock the call cleanly.
 *
 * 200 OK: { "success": true, "data": { "message": "Capture session cancelled." } }
 */
static mhd_result handle_capture_cancel(struct MHD_Connection *conn) {
    CancelActiveReaderCapture();
    return send_json(conn, MHD_HTTP_OK,
        "{\"success\":true,\"data\":{"
        "\"message\":\"Capture session cancelled.\"}}");
}


/* ── Endpoint: POST /api/v1/enrollment ───────────────────────────────────── */
/**
 * Performs a multi-sample fingerprint enrollment session.
 *
 * The caller specifies how many capture samples are required via the JSON
 * request body. The server captures each sample sequentially, blocking
 * between captures until the user places (and removes) their finger.
 *
 * Request body (JSON):
 *   {
 *     "count":        4,              // Number of captures required (1-10)
 *     "finger_index": 1,              // Optional: NIST finger position code (1=right thumb)
 *     "session_id":   "enroll-abc123" // Optional: caller-supplied session identifier
 *   }
 *
 * Response 200:
 *   {
 *     "success": true,
 *     "data": {
 *       "session_id":    "enroll-abc123",
 *       "finger_index":  1,
 *       "fmd_format":    "ISO_19794_2_2005",
 *       "serial_number": "5000216000000000",
 *       "samples_requested": 4,
 *       "samples_captured":  4,
 *       "samples": [
 *         { "index": 0, "fmd_base64": "...", "quality": 87 },
 *         { "index": 1, "fmd_base64": "...", "quality": 91 },
 *         { "index": 2, "fmd_base64": "...", "quality": 82 },
 *         { "index": 3, "fmd_base64": "...", "quality": 89 }
 *       ]
 *     }
 *   }
 *
 * If a single sample capture fails, the session is aborted and the endpoint
 * returns a 500 error with the sample index that failed.
 */
static mhd_result handle_enrollment(struct MHD_Connection *conn,
                             const char *upload_data,
                             size_t *upload_data_size,
                             void **con_cls) {
    /*
     * libmicrohttpd delivers POST body in chunks.  We accumulate the body
     * across calls using a heap buffer stored in con_cls.
     *
     * State machine:
     *   con_cls == NULL  → first call, allocate accumulator
     *   upload_data_size > 0 → body chunk arriving, append
     *   upload_data_size == 0 → body complete, process request
     */

    /* ── Accumulator for chunked POST body ── */
    typedef struct { char *buf; size_t len; } BodyBuf;

    if (*con_cls == NULL) {
        BodyBuf *bb = (BodyBuf *)calloc(1, sizeof(BodyBuf));
        if (!bb) return MHD_NO;
        *con_cls = bb;
        return MHD_YES;
    }

    BodyBuf *bb = (BodyBuf *)(*con_cls);

    if (*upload_data_size > 0) {
        char *tmp = (char *)realloc(bb->buf, bb->len + *upload_data_size + 1);
        if (!tmp) return MHD_NO;
        bb->buf = tmp;
        memcpy(bb->buf + bb->len, upload_data, *upload_data_size);
        bb->len += *upload_data_size;
        bb->buf[bb->len] = '\0';
        *upload_data_size = 0;
        return MHD_YES;
    }

    /* ── Body fully received — parse JSON fields ── */
    int    count        = 1;
    int    finger_index = 0;
    char   session_id[64] = {0};

    const char *body = (bb->buf && bb->len > 0) ? bb->buf : "{}";

    /* Minimal JSON field extraction (no external JSON library dependency) */
    {
        const char *p;

        /* "count": N */
        p = strstr(body, "\"count\"");
        if (p) {
            p = strchr(p, ':');
            if (p) count = atoi(p + 1);
        }

        /* "finger_index": N */
        p = strstr(body, "\"finger_index\"");
        if (p) {
            p = strchr(p, ':');
            if (p) finger_index = atoi(p + 1);
        }

        /* "session_id": "..." */
        p = strstr(body, "\"session_id\"");
        if (p) {
            p = strchr(p, '"');  /* skip key */
            if (p) p = strchr(p + 1, '"');  /* skip closing quote of key */
            if (p) p = strchr(p + 1, '"');  /* open quote of value */
            if (p) {
                const char *end = strchr(p + 1, '"');
                if (end) {
                    size_t slen = (size_t)(end - p - 1);
                    if (slen >= sizeof(session_id)) slen = sizeof(session_id) - 1;
                    strncpy(session_id, p + 1, slen);
                    session_id[slen] = '\0';
                }
            }
        }
    }

    /* Clamp count to sane bounds */
    if (count < 1)  count = 1;
    if (count > 10) count = 10;

    /* Generate a default session_id if not supplied */
    if (session_id[0] == '\0') {
        snprintf(session_id, sizeof(session_id), "session-%ld", (long)time(NULL));
    }

    /* Free body accumulator (no longer needed) */
    if (bb->buf) free(bb->buf);
    free(bb);
    *con_cls = NULL;

    /* ── Open the reader ── */
    DPFPDD_DEV hReader = NULL;
    int dpi = 0;
    char szReader[DPFPDD_MAX_STR_LENGTH] = {0};
    if (!open_reader(conn, &hReader, &dpi, szReader, sizeof(szReader)))
        return MHD_YES;

    char serial[DPFPDD_MAX_STR_LENGTH] = {0};
    QueryDeviceSerial(serial, sizeof(serial));

    /* ── Capture N samples sequentially ── */

    /* Allocate arrays for FMD base64 strings and quality scores */
    char   **fmd_b64_arr = (char **)calloc((size_t)count, sizeof(char *));
    int     *quality_arr = (int  *)calloc((size_t)count, sizeof(int));
    int      captured    = 0;

    if (!fmd_b64_arr || !quality_arr) {
        free(fmd_b64_arr); free(quality_arr);
        close_reader(&hReader);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          "Memory allocation failed for enrollment arrays.");
    }

    int capture_error = 0;
    char capture_error_msg[512] = {0};

    /* Start FingerJet consolidated enrollment state machine */
    dpfj_finish_enrollment();
    int enroll_rc = dpfj_start_enrollment(DEFAULT_FMD_FORMAT);
    if (DPFJ_SUCCESS != enroll_rc) {
        free(fmd_b64_arr); free(quality_arr);
        close_reader(&hReader);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          "dpfj_start_enrollment failed.");
    }

    for (int i = 0; i < count; ++i) {
        FingerprintCaptureResult res;
        memset(&res, 0, sizeof(res));
        res.quality = -1;

        fprintf(stderr, "[INFO] Enrollment %s: capturing sample %d/%d\n",
                session_id, i + 1, count);

        int rc = CaptureFingerprintImage(hReader, dpi, &res);
        if (0 != rc) {
            snprintf(capture_error_msg, sizeof(capture_error_msg),
                     "Sample %d capture failed: %s", i + 1,
                     res.error_msg[0] ? res.error_msg : "unknown error");
            FreeCaptureResult(&res);
            capture_error = 1;
            break;
        }

        AssessImageQuality(&res);
        /* Reject unusable NFIQ scans (NFIQ 4 or 5) and retry same step */
        if (res.quality >= 4) {
            fprintf(stderr, "[WARN] Poor NFIQ score (%d), retrying scan %d\n", res.quality, i + 1);
            FreeCaptureResult(&res);
            --i;
            continue;
        }
        quality_arr[i] = res.quality;

        rc = ExtractFMD(&res, DEFAULT_FMD_FORMAT);
        if (0 != rc) {
            snprintf(capture_error_msg, sizeof(capture_error_msg),
                     "Sample %d FMD extraction failed: %s", i + 1,
                     res.error_msg[0] ? res.error_msg : "unknown error");
            FreeCaptureResult(&res);
            capture_error = 1;
            break;
        }

        /* Add view FMD to FingerJet consolidated enrollment */
        dpfj_add_to_enrollment(DEFAULT_FMD_FORMAT, res.fmd_data, res.fmd_size, 0);

        size_t b64_len = 0;
        fmd_b64_arr[i] = Base64Encode(res.fmd_data, res.fmd_size, &b64_len);
        FreeCaptureResult(&res);

        if (!fmd_b64_arr[i]) {
            snprintf(capture_error_msg, sizeof(capture_error_msg),
                     "Sample %d Base64 encoding failed.", i + 1);
            capture_error = 1;
            break;
        }

        captured++;
    }

    /* Extract consolidated ISO 19794-2:2005 enrollment FMD */
    char *consolidated_b64 = NULL;
    if (!capture_error && captured > 0) {
        unsigned int cons_size = MAX_FMD_SIZE;
        unsigned char *cons_buf = (unsigned char *)malloc(cons_size);
        if (cons_buf) {
            if (DPFJ_SUCCESS == dpfj_create_enrollment_fmd(cons_buf, &cons_size)) {
                size_t cb_len = 0;
                consolidated_b64 = Base64Encode(cons_buf, cons_size, &cb_len);
            }
            free(cons_buf);
        }
    }
    dpfj_finish_enrollment();

    close_reader(&hReader);

    if (capture_error) {
        free(consolidated_b64);
        for (int i = 0; i < captured; ++i) free(fmd_b64_arr[i]);
        free(fmd_b64_arr); free(quality_arr);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          capture_error_msg);
    }

    /* ── Build JSON response ── */
    size_t total_b64_len = consolidated_b64 ? strlen(consolidated_b64) : 0;
    for (int i = 0; i < captured; ++i)
        total_b64_len += strlen(fmd_b64_arr[i]);

    size_t jlen = total_b64_len + (size_t)(captured * 64) + 1024;
    char *jbuf  = (char *)malloc(jlen);
    if (!jbuf) {
        free(consolidated_b64);
        for (int i = 0; i < captured; ++i) free(fmd_b64_arr[i]);
        free(fmd_b64_arr); free(quality_arr);
        return send_error(conn, MHD_HTTP_INTERNAL_SERVER_ERROR, 500,
                          "Memory allocation failed for response buffer.");
    }

    const char *final_fmd = consolidated_b64 ? consolidated_b64 : (captured > 0 ? fmd_b64_arr[0] : "");

    /* Build the response including consolidated_fmd */
    size_t offset = 0;
    offset += (size_t)snprintf(jbuf + offset, jlen - offset,
        "{\"success\":true,\"consolidated_fmd\":\"%s\",\"fmd\":\"%s\",\"data\":{"
        "\"session_id\":\"%s\","
        "\"finger_index\":%d,"
        "\"fmd_format\":\"ISO_19794_2_2005\","
        "\"consolidated_fmd\":\"%s\","
        "\"serial_number\":\"%s\","
        "\"samples_requested\":%d,"
        "\"samples_captured\":%d,"
        "\"samples\":[",
        final_fmd, final_fmd, session_id, finger_index, final_fmd, serial, count, captured);

    for (int i = 0; i < captured; ++i) {
        offset += (size_t)snprintf(jbuf + offset, jlen - offset,
            "%s{\"index\":%d,\"fmd_base64\":\"%s\",\"quality\":%d}",
            (i > 0) ? "," : "",
            i, fmd_b64_arr[i], quality_arr[i]);
        free(fmd_b64_arr[i]);
    }

    snprintf(jbuf + offset, jlen - offset, "]}}");

    free(consolidated_b64);
    free(fmd_b64_arr);
    free(quality_arr);

    int ret = send_json(conn, MHD_HTTP_OK, jbuf);
    free(jbuf);
    return ret;
}

/* ── Request Dispatcher ───────────────────────────────────────────────────── */
/**
 * @brief libmicrohttpd request callback — routes all incoming HTTP requests.
 *
 * Routing table:
 *   OPTIONS *                   → 204 CORS preflight
 *   GET     /api/v1/device      → handle_device_status()
 *   GET     /api/v1/device/info → handle_device_info()
 *   POST    /api/v1/capture     → handle_capture()
 *   DELETE  /api/v1/capture     → handle_capture_cancel()
 *   *                           → 404 Not Found
 */
static mhd_result request_handler(void *cls,
                            struct MHD_Connection *conn,
                            const char *url, const char *method,
                            const char *version,
                            const char *upload_data, size_t *upload_data_size,
                            void **con_cls) {
    (void)cls; (void)version; (void)upload_data;
    (void)upload_data_size; (void)con_cls;

    fprintf(stderr, "[INFO] %s %s\n", method, url);

    if (0 == strcmp(method, "OPTIONS")) return handle_options(conn);

    if (0 == strcmp(url, API_PREFIX "/device")) {
        if (0 != strcmp(method, "GET")) return send_method_not_allowed(conn);
        return handle_device_status(conn);
    }
    if (0 == strcmp(url, API_PREFIX "/device/info")) {
        if (0 != strcmp(method, "GET")) return send_method_not_allowed(conn);
        return handle_device_info(conn);
    }
    if (0 == strcmp(url, API_PREFIX "/enrollment")) {
        if (0 != strcmp(method, "POST")) return send_method_not_allowed(conn);
        return handle_enrollment(conn, upload_data, upload_data_size, con_cls);
    }

    if (0 == strcmp(url, API_PREFIX "/capture")) {
        if (0 == strcmp(method, "POST"))   return handle_capture(conn);
        if (0 == strcmp(method, "DELETE")) return handle_capture_cancel(conn);
        return send_method_not_allowed(conn);
    }

    return send_not_found(conn);
}

/* ── Signal Handler ───────────────────────────────────────────────────────── */
static struct MHD_Daemon *g_daemon = NULL;

static void sigint_handler(int sig) {
    (void)sig;
    fprintf(stderr, "\n[INFO] Shutdown signal received; stopping server...\n");
    if (g_daemon) { MHD_stop_daemon(g_daemon); g_daemon = NULL; }
}

/* ── main ─────────────────────────────────────────────────────────────────── */
int main(void) {
    signal(SIGINT, sigint_handler);

    g_daemon = MHD_start_daemon(
        MHD_USE_INTERNAL_POLLING_THREAD | MHD_USE_THREAD_PER_CONNECTION | MHD_USE_DEBUG,
        API_PORT, NULL, NULL,
        &request_handler, NULL,
        MHD_OPTION_CONNECTION_LIMIT, (unsigned int)CONNECTION_LIMIT,
        MHD_OPTION_END);

    if (!g_daemon) {
        fprintf(stderr, "[FATAL] Failed to start HTTP daemon on port %d.\n", API_PORT);
        return EXIT_FAILURE;
    }

    printf("\n");
    printf("=================================================================\n");
    printf("  DigitalPersona U.are.U Fingerprint Microservice  v2.0\n");
    printf("=================================================================\n");
    printf("  Base URL : http://127.0.0.1:%d%s\n", API_PORT, API_PREFIX);
    printf("\n");
    printf("  GET    /api/v1/device          Reader connection status\n");
    printf("  GET    /api/v1/device/info     Reader metadata\n");
    printf("  POST   /api/v1/capture         Capture image + FMD template\n");
    printf("  DELETE /api/v1/capture         Cancel active capture session\n");
    printf("  POST   /api/v1/enrollment       Multi-sample enrollment session\n");
    printf("\n");
    printf("  Press Enter to stop...\n");
    printf("=================================================================\n\n");

    getchar();

    MHD_stop_daemon(g_daemon);
    g_daemon = NULL;
    printf("[INFO] Server stopped.\n");
    return EXIT_SUCCESS;
}
