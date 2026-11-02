/*
 * fingerprint_selection.c - Device Enumeration and Selection Implementation
 *
 * Implements device discovery using dpfpdd_query_devices(), dpfpdd_open(),
 * and dpfpdd_get_device_capabilities() from the DigitalPersona U.are.U SDK.
 * A reusable enumerate_devices() helper handles the two-pass reallocation
 * pattern required by the SDK API.
 *
 * Copyright (C) 2011 Digital Persona, Inc. (SDK); project contributors (microservice)
 */
#include "fingerprint_selection.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* ── Internal Helper: enumerate_devices ───────────────────────────────────── */

/**
 * @brief Allocate and populate a DPFPDD_DEV_INFO array for all connected devices.
 *
 * Handles the SDK two-pass reallocation pattern: dpfpdd_query_devices()
 * returns DPFPDD_E_MORE_DATA when the buffer is too small, then updates
 * the count; the buffer is reallocated and the call is retried.
 *
 * @param[out] pp_info  Set to a heap-allocated DPFPDD_DEV_INFO array; caller must free().
 * @param[out] p_count  Receives the number of entries written.
 * @return              0 on success; non-zero error code on failure.
 */
static int enumerate_devices(DPFPDD_DEV_INFO **pp_info, unsigned int *p_count) {
    unsigned int cnt = 1;
    DPFPDD_DEV_INFO *info =
        (DPFPDD_DEV_INFO *)malloc(sizeof(DPFPDD_DEV_INFO) * cnt);
    if (!info) { fprintf(stderr, "[ERROR] enumerate_devices: malloc failed\n"); return -1; }

    for (;;) {
        unsigned int i;
        for (i = 0; i < cnt; ++i) info[i].size = sizeof(DPFPDD_DEV_INFO);

        unsigned int new_cnt = cnt;
        int rc = dpfpdd_query_devices(&new_cnt, info);

        if (DPFPDD_SUCCESS == rc) { cnt = new_cnt; break; }

        if (DPFPDD_E_MORE_DATA == rc) {
            DPFPDD_DEV_INFO *tmp =
                (DPFPDD_DEV_INFO *)realloc(info, sizeof(DPFPDD_DEV_INFO) * new_cnt);
            if (!tmp) {
                fprintf(stderr, "[ERROR] enumerate_devices: realloc failed\n");
                free(info); return -1;
            }
            info = tmp; cnt = new_cnt; continue;
        }

        fprintf(stderr, "[ERROR] dpfpdd_query_devices() rc=%d\n", rc);
        free(info); return rc;
    }

    *pp_info = info; *p_count = cnt;
    return 0;
}

/* ── SelectAndOpenReader ─────────────────────────────────────────────────── */
DPFPDD_DEV SelectAndOpenReader(char *szReader, size_t nReaderLen, int *pDPI) {
    if (!szReader || nReaderLen == 0 || !pDPI) return NULL;
    strncpy(szReader, "", nReaderLen);
    *pDPI = 0;

    DPFPDD_DEV_INFO *info = NULL; unsigned int cnt = 0;
    if (0 != enumerate_devices(&info, &cnt) || cnt == 0) { free(info); return NULL; }

    DPFPDD_DEV hReader = NULL;
    int rc = dpfpdd_open(info[0].name, &hReader);
    if (DPFPDD_SUCCESS != rc) {
        fprintf(stderr, "[ERROR] dpfpdd_open() rc=%d\n", rc);
        free(info); return NULL;
    }

    strncpy(szReader, info[0].name, nReaderLen - 1);
    szReader[nReaderLen - 1] = '\0';

    /* Query device capabilities to retrieve native resolution (DPI) */
    unsigned int caps_size = sizeof(DPFPDD_DEV_CAPS);
    DPFPDD_DEV_CAPS *caps = (DPFPDD_DEV_CAPS *)malloc(caps_size);
    if (caps) {
        caps->size = caps_size;
        rc = dpfpdd_get_device_capabilities(hReader, caps);

        if (DPFPDD_E_MORE_DATA == rc) {
            caps_size = caps->size;
            DPFPDD_DEV_CAPS *tmp = (DPFPDD_DEV_CAPS *)realloc(caps, caps_size);
            if (tmp) { caps = tmp; caps->size = caps_size;
                       rc = dpfpdd_get_device_capabilities(hReader, caps); }
            else rc = -1;
        }

        if (DPFPDD_SUCCESS == rc && caps->resolution_cnt > 0)
            *pDPI = (int)caps->resolutions[0];
        else
            fprintf(stderr, "[WARN] dpfpdd_get_device_capabilities() rc=%d; DPI unknown\n", rc);

        free(caps);
    }

    free(info);
    return hReader;
}

/* ── QueryDeviceSerial ───────────────────────────────────────────────────── */
int QueryDeviceSerial(char *serial_out, size_t buf_len) {
    if (!serial_out || buf_len == 0) return -1;
    strncpy(serial_out, "", buf_len);

    DPFPDD_DEV_INFO *info = NULL; unsigned int cnt = 0;
    if (0 != enumerate_devices(&info, &cnt) || cnt == 0) { free(info); return -1; }

    strncpy(serial_out, info[0].descr.serial_num, buf_len - 1);
    serial_out[buf_len - 1] = '\0';
    free(info);
    return 0;
}
