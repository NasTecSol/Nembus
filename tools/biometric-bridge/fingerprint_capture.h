/*
 * fingerprint_capture.h - Biometric Capture and Feature Extraction Interface
 *
 * Declares the public API for fingerprint image acquisition,
 * FMD template generation, image quality assessment, and BMP serialization
 * using the DigitalPersona U.are.U SDK (libdpfpdd, libdpfj).
 *
 * SDK components used:
 *   dpfpdd.h           - Device enumeration, open/close, image capture
 *   dpfj.h             - FMD creation and biometric matching
 *   dpfj_compression.h - FMD compression utilities
 *   dpfj_quality.h     - Image quality assessment
 *
 * Copyright (C) 2011 Digital Persona, Inc. (SDK); project contributors (microservice)
 */
#pragma once
#include <dpfpdd.h>
#include <dpfj.h>
#include <dpfj_compression.h>
#include <dpfj_quality.h>
#include <stddef.h>

#ifndef MAX_FMD_SIZE
#define MAX_FMD_SIZE       (100 * 1024)
#endif
#define DEFAULT_FMD_FORMAT DPFJ_FMD_ISO_19794_2_2005

typedef struct {
    unsigned char  *image_data;
    unsigned int    image_size;
    unsigned int    image_width;
    unsigned int    image_height;
    unsigned char  *fmd_data;
    unsigned int    fmd_size;
    DPFJ_FMD_FORMAT fmd_format;
    int             quality;
    int             success;
    char            error_msg[256];
} FingerprintCaptureResult;

int  CaptureFingerprintImage(DPFPDD_DEV hReader, int dpi, FingerprintCaptureResult *out);
int  CancelActiveReaderCapture(void);
int  ExtractFMD(FingerprintCaptureResult *result, DPFJ_FMD_FORMAT fmt);
int  AssessImageQuality(FingerprintCaptureResult *result);
int  SaveBMP(const char *filename, unsigned char *image_data, int width, int height);
char *SerializeBMPBase64(unsigned char *image_data, int width, int height, size_t *out_len);
void FreeCaptureResult(FingerprintCaptureResult *result);
char          *Base64Encode(const unsigned char *input, size_t length, size_t *out_len);
unsigned char *ReadFileBinary(const char *filename, size_t *length);
