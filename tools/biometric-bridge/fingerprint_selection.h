/*
 * fingerprint_selection.h - Device Enumeration and Selection Interface
 *
 * Declares functions for enumerating connected biometric readers, opening
 * the first available device, and querying device metadata using the
 * DigitalPersona U.are.U SDK (dpfpdd.h).
 *
 * Copyright (C) 2011 Digital Persona, Inc. (SDK); project contributors (microservice)
 */
#pragma once
#include <dpfpdd.h>
#include <stddef.h>

/**
 * @brief Enumerate connected readers and open the first available device.
 *
 * Calls dpfpdd_query_devices() then dpfpdd_open() on the first entry.
 * Device capabilities are retrieved via dpfpdd_get_device_capabilities()
 * to extract the native sensor DPI. The caller must call dpfpdd_close()
 * on the returned handle when finished.
 *
 * @param szReader    Buffer to receive the device name string.
 * @param nReaderLen  Byte capacity of szReader.
 * @param pDPI        Receives the device's first supported resolution (DPI).
 * @return            Open DPFPDD_DEV handle on success, or NULL on failure.
 */
DPFPDD_DEV SelectAndOpenReader(char *szReader, size_t nReaderLen, int *pDPI);

/**
 * @brief Query the serial number of the first connected reader.
 *
 * Enumerates devices and copies the serial number into serial_out.
 * The buffer should be at least DPFPDD_MAX_STR_LENGTH bytes.
 *
 * @param serial_out  Buffer to receive the serial number string.
 * @param buf_len     Byte capacity of serial_out.
 * @return            0 on success; -1 if no device found or on error.
 */
int QueryDeviceSerial(char *serial_out, size_t buf_len);
