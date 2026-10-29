# ==============================================================================
# Makefile: DigitalPersona U.are.U HTTP Microservice
# ==============================================================================

CC ?= gcc
EXE_NAME = finger

SDK_INCLUDES ?= ../Crossmatch/urusdk-linux/Include
OUT_DIR ?= .
LIB_OUT_DIR ?= ../Crossmatch/urusdk-linux/Linux/lib/x64

CCFLAGS = -g -Wall -I$(SDK_INCLUDES) $(CFLAGS)

# Function to obtain real target link path under Cygwin environments
getlink = $(shell cygpath -wm `readlink -f $(1)`)

ifeq ($(findstring arm,$(CFLAGS))$(findstring CYGWIN,$(shell uname)),armCYGWIN)
	# Code Sourcery toolchain under Cygwin cannot dereference symlinks; link to real library path
	LDFLAGS = -lpthread -lm -lc $(CFLAGS) $(call getlink,$(LIB_OUT_DIR)/libdpfpdd.so) $(call getlink,$(LIB_OUT_DIR)/libdpfj.so)
else
	LDFLAGS = -lpthread -lm -lc -ldl $(CFLAGS) -L$(LIB_OUT_DIR) -L/usr/lib -ldpfpdd -ldpfj
endif

# Link GNU libmicrohttpd for HTTP server functionality
LDFLAGS += -lmicrohttpd

OBJS = fingerprint_server.o fingerprint_capture.o fingerprint_selection.o

.PHONY: all all_debug clean install help

all: $(OBJS)
	mkdir -p $(OUT_DIR)
	$(CC) $(OBJS) -Wl,--no-as-needed $(LDFLAGS) -o $(OUT_DIR)/$(EXE_NAME)

all_debug: CCFLAGS += -DDEBUG
all_debug: all

clean:
	rm -f $(OUT_DIR)/$(EXE_NAME) *.o *~

%.o: %.c
	$(CC) $(CCFLAGS) -c $< -o $@

install: all
	install -d $(DESTDIR)/usr/local/bin
	install -m 755 $(OUT_DIR)/$(EXE_NAME) $(DESTDIR)/usr/local/bin/

help:
	@echo "Available targets:"
	@echo "  all        Build the microservice executable (default)"
	@echo "  all_debug  Build with debug symbols enabled"
	@echo "  clean      Remove compiled objects and executable"
	@echo "  install    Install binary to system path"
	@echo "  help       Show this help message"
