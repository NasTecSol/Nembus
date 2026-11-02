/*
 * fingerprint_capture.c - Biometric Capture, FMD Extraction, and BMP Serialization
 *
 * Implements the full fingerprint capture pipeline:
 *   1. CaptureFingerprintImage() - raw image via dpfpdd_capture()
 *   2. ExtractFMD()              - minutiae template via dpfj_create_fmd_from_fid()
 *   3. AssessImageQuality()      - quality score via dpfj_get_fid_quality()
 *   4. SaveBMP()                 - cropped 24-bit BMP serialization
 *
 * Copyright (C) 2011 Digital Persona, Inc. (SDK); project contributors (microservice)
 */
#include "fingerprint_capture.h"
#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <stdint.h>

/* ── BMP File Format ──────────────────────────────────────────────────────── */
#pragma pack(push, 1)
typedef struct {
    unsigned short  type;
    unsigned int    file_size;
    unsigned short  reserved1;
    unsigned short  reserved2;
    unsigned int    pixel_offset;
    unsigned int    header_size;
    int             width;
    int             height;
    unsigned short  planes;
    unsigned short  bits_per_pixel;
    unsigned int    compression;
    unsigned int    image_size;
    int             x_pixels_per_meter;
    int             y_pixels_per_meter;
    unsigned int    colors_used;
    unsigned int    colors_important;
} BMPHeader;
#pragma pack(pop)

static DPFPDD_DEV g_hActiveReader = NULL;

int CancelActiveReaderCapture(void) {
    DPFPDD_DEV h = g_hActiveReader;
    if (NULL != h) {
        fprintf(stderr, "[INFO] In-process dpfpdd_cancel() invoked.\n");
        return dpfpdd_cancel(h);
    }
    return 0;
}

static void capture_signal_handler(int sig) {
    if (SIGINT == sig && NULL != g_hActiveReader) {
        fprintf(stderr, "[INFO] Capture interrupted; cancelling reader.\n");
        dpfpdd_cancel(g_hActiveReader);
    }
}

/* ── CaptureFingerprintImage ──────────────────────────────────────────────── */
int CaptureFingerprintImage(DPFPDD_DEV hReader, int dpi,
                            FingerprintCaptureResult *out) {
    if (!hReader || !out) return -1;
    memset(out, 0, sizeof(*out));
    out->quality = -1;

    DPFPDD_CAPTURE_PARAM cparam = {0};
    cparam.size      = sizeof(cparam);
    cparam.image_fmt = DPFPDD_IMG_FMT_ISOIEC19794;
    cparam.image_proc = DPFPDD_IMG_PROC_NONE;
    cparam.image_res = (unsigned int)dpi;

    DPFPDD_CAPTURE_RESULT cresult = {0};
    cresult.size      = sizeof(cresult);
    cresult.info.size = sizeof(cresult.info);

    /* Probe required buffer size (non-blocking, timeout=0) */
    unsigned int image_size = 0;
    int rc = dpfpdd_capture(hReader, &cparam, 0, &cresult, &image_size, NULL);
    if (DPFPDD_E_MORE_DATA != rc) {
        snprintf(out->error_msg, sizeof(out->error_msg),
                 "dpfpdd_capture() probe failed (rc=%d)", rc);
        return rc;
    }

    unsigned char *buf = (unsigned char *)malloc(image_size);
    if (!buf) {
        snprintf(out->error_msg, sizeof(out->error_msg),
                 "malloc failed for image buffer (%u bytes)", image_size);
        return -1;
    }

    /* Register temporary SIGINT handler to support cancellation */
    g_hActiveReader = hReader;
    struct sigaction new_sa, old_sa;
    new_sa.sa_handler = capture_signal_handler;
    sigemptyset(&new_sa.sa_mask);
    new_sa.sa_flags = 0;
    sigaction(SIGINT, &new_sa, &old_sa);
    sigset_t new_mask, old_mask;
    sigemptyset(&new_mask);
    sigaddset(&new_mask, SIGINT);
    pthread_sigmask(SIG_UNBLOCK, &new_mask, &old_mask);

    /* Blocking capture (timeout=-1: wait indefinitely for finger) */
    unsigned int actual = image_size;
    rc = dpfpdd_capture(hReader, &cparam, -1, &cresult, &actual, buf);

    sigaction(SIGINT, &old_sa, NULL);
    pthread_sigmask(SIG_SETMASK, &old_mask, NULL);
    g_hActiveReader = NULL;

    if (DPFPDD_SUCCESS != rc) {
        free(buf);
        snprintf(out->error_msg, sizeof(out->error_msg),
                 "dpfpdd_capture() failed (rc=%d)", rc);
        return rc;
    }

    if (!cresult.success) {
        free(buf);
        if (DPFPDD_QUALITY_CANCELED == cresult.quality)
            snprintf(out->error_msg, sizeof(out->error_msg),
                     "Capture was cancelled.");
        else
            snprintf(out->error_msg, sizeof(out->error_msg),
                     "Capture unsuccessful; quality code=%d", cresult.quality);
        return -1;
    }

    out->image_data   = buf;
    out->image_size   = actual;
    out->image_width  = cresult.info.width;
    out->image_height = cresult.info.height;
    out->success      = 1;
    return 0;
}

/* ── ExtractFMD ───────────────────────────────────────────────────────────── */
int ExtractFMD(FingerprintCaptureResult *result, DPFJ_FMD_FORMAT fmt) {
    if (!result || !result->image_data) return -1;

    unsigned int fmd_size = MAX_FMD_SIZE;
    unsigned char *fmd_buf = (unsigned char *)malloc(fmd_size);
    if (!fmd_buf) {
        snprintf(result->error_msg, sizeof(result->error_msg),
                 "malloc failed for FMD buffer");
        return -1;
    }

    struct timeval t1, t2;
    gettimeofday(&t1, NULL);

    int rc = dpfj_create_fmd_from_fid(
        DPFJ_FID_ISO_19794_4_2005,
        result->image_data, result->image_size,
        fmt,
        fmd_buf, &fmd_size);

    gettimeofday(&t2, NULL);
    long ms = (t2.tv_sec - t1.tv_sec) * 1000 + (t2.tv_usec - t1.tv_usec) / 1000;
    fprintf(stderr, "[INFO] dpfj_create_fmd_from_fid() completed in %ld ms\n", ms);

    if (DPFJ_SUCCESS != rc) {
        free(fmd_buf);
        snprintf(result->error_msg, sizeof(result->error_msg),
                 "dpfj_create_fmd_from_fid() failed (rc=%d)", rc);
        return rc;
    }

    result->fmd_data   = fmd_buf;
    result->fmd_size   = fmd_size;
    result->fmd_format = fmt;
    return 0;
}

/* ── AssessImageQuality ───────────────────────────────────────────────────── */
int AssessImageQuality(FingerprintCaptureResult *result) {
    if (!result || !result->image_data) return -1;

    unsigned int quality = 0;
    int rc = dpfj_quality_nfiq_from_fid(
        DPFJ_FID_ISO_19794_4_2005,
        result->image_data, result->image_size,
        0,
        DPFJ_QUALITY_NFIQ_NIST,
        &quality);

    if (DPFJ_SUCCESS == rc)
        result->quality = (int)quality;
    else {
        fprintf(stderr, "[WARN] dpfj_quality_nfiq_from_fid() failed (rc=%d)\n", rc);
        result->quality = -1;
    }
    return rc;
}

/* ── SaveBMP ──────────────────────────────────────────────────────────────── */
int SaveBMP(const char *filename, unsigned char *image_data, int width, int height) {
    /* ISO/IEC 19794-4:2005 FID has a 46-byte record+view header before raw grayscale pixels */
    const int ISO_HEADER_OFFSET = 46;
    const unsigned char *pixels = image_data + ISO_HEADER_OFFSET;
    int stride   = width * 3;
    int padding  = (4 - (stride % 4)) % 4;
    int pix_size = (stride + padding) * height;

    BMPHeader hdr;
    memset(&hdr, 0, sizeof(hdr));
    hdr.type            = 0x4D42;
    hdr.file_size       = (unsigned int)(sizeof(BMPHeader) + pix_size);
    hdr.pixel_offset    = sizeof(BMPHeader);
    hdr.header_size     = 40;
    hdr.width           = width;
    hdr.height          = height;
    hdr.planes          = 1;
    hdr.bits_per_pixel  = 24;
    hdr.image_size      = (unsigned int)pix_size;

    FILE *fp = fopen(filename, "wb");
    if (!fp) {
        fprintf(stderr, "[ERROR] SaveBMP: cannot open '%s': %s\n",
                filename, strerror(errno));
        return -1;
    }

    fwrite(&hdr, sizeof(BMPHeader), 1, fp);
    for (int y = height - 1; y >= 0; --y) {
        for (int x = 0; x < width; ++x) {
            unsigned char px = pixels[y * width + x];
            fwrite(&px, 1, 1, fp);
            fwrite(&px, 1, 1, fp);
            fwrite(&px, 1, 1, fp);
        }
        for (int p = 0; p < padding; ++p) fputc(0, fp);
    }

    fclose(fp);
    return 0;
}

char *SerializeBMPBase64(unsigned char *image_data, int width, int height, size_t *out_len) {
    if (!image_data || width <= 0 || height <= 0 || !out_len) return NULL;
    const int ISO_HEADER_OFFSET = 46;
    const unsigned char *pixels = image_data + ISO_HEADER_OFFSET;
    int stride   = width * 3;
    int padding  = (4 - (stride % 4)) % 4;
    int pix_size = (stride + padding) * height;

    BMPHeader hdr;
    memset(&hdr, 0, sizeof(hdr));
    hdr.type            = 0x4D42;
    hdr.file_size       = (unsigned int)(sizeof(BMPHeader) + pix_size);
    hdr.pixel_offset    = sizeof(BMPHeader);
    hdr.header_size     = 40;
    hdr.width           = width;
    hdr.height          = height;
    hdr.planes          = 1;
    hdr.bits_per_pixel  = 24;
    hdr.image_size      = (unsigned int)pix_size;

    size_t total = sizeof(BMPHeader) + (size_t)pix_size;
    unsigned char *buf = (unsigned char *)malloc(total);
    if (!buf) return NULL;
    memcpy(buf, &hdr, sizeof(BMPHeader));
    unsigned char *dst = buf + sizeof(BMPHeader);
    for (int y = height - 1; y >= 0; --y) {
        for (int x = 0; x < width; ++x) {
            unsigned char px = pixels[y * width + x];
            *dst++ = px;
            *dst++ = px;
            *dst++ = px;
        }
        for (int p = 0; p < padding; ++p) {
            *dst++ = 0;
        }
    }
    char *encoded = Base64Encode(buf, total, out_len);
    free(buf);
    return encoded;
}

/* ── FreeCaptureResult ────────────────────────────────────────────────────── */
void FreeCaptureResult(FingerprintCaptureResult *result) {
    if (!result) return;
    free(result->image_data);
    free(result->fmd_data);
    memset(result, 0, sizeof(*result));
    result->quality = -1;
}

/* ── Base64Encode ─────────────────────────────────────────────────────────── */
char *Base64Encode(const unsigned char *input, size_t length, size_t *out_len) {
    static const char CHARS[] =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    *out_len = 4 * ((length + 2) / 3);
    char *out = (char *)malloc(*out_len + 1);
    if (!out) return NULL;

    size_t i, j;
    for (i = 0, j = 0; i < length; i += 3, j += 4) {
        uint32_t a = (i     < length) ? input[i]     : 0;
        uint32_t b = (i + 1 < length) ? input[i + 1] : 0;
        uint32_t c = (i + 2 < length) ? input[i + 2] : 0;
        uint32_t t = (a << 16) | (b << 8) | c;
        out[j]     = CHARS[(t >> 18) & 0x3F];
        out[j + 1] = CHARS[(t >> 12) & 0x3F];
        out[j + 2] = (i + 1 < length) ? CHARS[(t >> 6) & 0x3F] : '=';
        out[j + 3] = (i + 2 < length) ? CHARS[t & 0x3F]        : '=';
    }
    out[*out_len] = '\0';
    return out;
}

/* ── ReadFileBinary ───────────────────────────────────────────────────────── */
unsigned char *ReadFileBinary(const char *filename, size_t *length) {
    FILE *fp = fopen(filename, "rb");
    if (!fp) { fprintf(stderr, "[ERROR] ReadFileBinary: %s\n", strerror(errno)); return NULL; }
    fseek(fp, 0, SEEK_END); *length = (size_t)ftell(fp); fseek(fp, 0, SEEK_SET);
    unsigned char *buf = (unsigned char *)malloc(*length);
    if (!buf) { fclose(fp); return NULL; }
    fread(buf, 1, *length, fp);
    fclose(fp);
    return buf;
}
