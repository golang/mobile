// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mobileinit

/*
To view the log output run:
adb logcat GoLog:I *:S
*/

// Android redirects stdout and stderr to /dev/null.
// As these are common debugging utilities in Go,
// we redirect them to logcat.
//
// Unfortunately, logcat is line oriented, so we must buffer.
//
// The redirection code is written in C, as stderr may be
// written while the the world is frozen (e.g. the runtime is
// printing the traceback of a panic). See issue #35590.

/*
#cgo LDFLAGS: -landroid -llog -pthread

#include <android/log.h>
#include <assert.h>
#include <errno.h>
#include <pthread.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define LOG_FATAL(...) __android_log_print(ANDROID_LOG_FATAL, CTAG, __VA_ARGS__)
#define LOG_DEBUG(...) __android_log_print(ANDROID_LOG_DEBUG, CTAG, __VA_ARGS__)

#define CTAG			"GoLog"	// logcat TAG
#define BUFFER_SIZE		1024	// matches android/log.h
#define REDIR_COUNT		2		// two fds: stdout & stderr

// Handler to redirect a file descriptor to Android's logcat
struct fd_redir {
	int fd;
	size_t buffer_cursor;
	int priority;

	// pre-allocated buffer. In case the buffer is full (buffer_cursor ==
	// BUFFER_SIZE) without any detected newline, the extra character at the end
	// of the buffer will be set to '\0' to print the current buffer as if the
	// last character was a newline.
	char buffer[BUFFER_SIZE + 1];
};

static struct fd_redir redirs[REDIR_COUNT];

// Read a line
void _redir_handle(struct fd_redir* redir) {
	assert(redir);
	assert(redir->buffer_cursor < BUFFER_SIZE);

	// Read the file descriptor into the pre-allocated buffer
	ssize_t read_ret = read(
		redir->fd,
		redir->buffer + redir->buffer_cursor,
		BUFFER_SIZE - redir->buffer_cursor);

	// Could not read the file
	if (read_ret < 0) {
		LOG_FATAL("read(%d) failed: %d. stop redirection.", redir->fd, errno);
		redir->fd = -1;
		return;
	}

	// File had been closed
	else if (read_ret == 0) {
		redir->fd = -1;
		return;
	}

	// redir->buffer had been populated with read buffer
	assert(read_ret > 0);
	redir->buffer_cursor += (size_t)read_ret;
	assert(redir->buffer_cursor <= BUFFER_SIZE);

	while (redir->buffer_cursor > 0) {
		// look for an endline character
		char* endline = memchr(redir->buffer, '\n', redir->buffer_cursor);

		// if no newline were detected ...
		if (!endline) {
			// ... and if buffer is not full, pass
			if (redir->buffer_cursor < BUFFER_SIZE) {
				break;
			}

			// if buffer is full, then prints the current buffer as a line.
			// buffer has an extra character so that it is safe to set the
			// redir->buffer[BUFFER_SIZE] to '\0'
			endline = redir->buffer + redir->buffer_cursor;
		}

		*endline = '\0';
		__android_log_write(redir->priority, CTAG, redir->buffer);

		// calculate the length of the remaining buffer. it can be negative
		// if the buffer was full and no newline were detected.
		ssize_t remaining_buffer_size = (ssize_t)redir->buffer_cursor - (endline + 1 - redir->buffer);

		if (remaining_buffer_size > 0) {
			memmove(redir->buffer, endline + 1, remaining_buffer_size);
		}

		redir->buffer_cursor = remaining_buffer_size > 0 ? remaining_buffer_size : 0;
	}
}

// Redirection thread's entrypoint
void* _redir_entrypoint(void* _unused) {
	LOG_DEBUG("stdout & stderr redirection thread has started");

	// Block all signals
	sigset_t set;
	sigfillset(&set);
	pthread_sigmask(SIG_SETMASK, &set, NULL);

	for (;;) {
		// Perform a select() call to wait for any readable fd in the
		// redirections
		fd_set read_fds;
		int max_fd = -1;

		FD_ZERO(&read_fds);
		for (int i=0; i<REDIR_COUNT; ++i) {
			int fd = redirs[i].fd;

			// File descriptor set to a negative value had been closed and
			// should be ignored.
			if (fd < 0) {
				continue;
			}

			FD_SET(fd, &read_fds);
			if (fd > max_fd) {
				max_fd = fd;
			}
		}

		// If there are no more file descriptor to read, break the loop and exit
		// nicely
		if (max_fd < 0) {
			LOG_DEBUG("no more file descriptors to redirect");
			break;
		}

		// Wait for any readable fd
		int select_ret = select(max_fd + 1, &read_fds, NULL, NULL, NULL);

		// select() returned an error
		if (select_ret < 0) {
			LOG_FATAL("select() failed: %d. stop all redirections.", errno);
			break;
		}

		// select() returned no fd are readable. This should not happen!
		else if (select_ret == 0) {
			LOG_FATAL("select() returned 0. stop all redirections.");
			break;
		}

		// call _redir_handle for every readable fds
		for (int i=0; i<REDIR_COUNT; ++i) {
			if (!FD_ISSET(redirs[i].fd, &read_fds)) {
				continue;
			}

			_redir_handle(redirs+i);
		}
	}

	LOG_DEBUG("exiting stdout & stderr redirection thread");

	return NULL;
}

// Handle of thread which will redirect the files to Android's logger
static pthread_t thread_redir;

// Initialize the redirection of stdout and stderr. Starts the redirection
// thread. The latter will automatically quit when stdout and stderr will not be
// readable anymore.
void initRedirection(int stdoutFd, int stderrFd) {
	redirs[0].buffer_cursor = 0;
	redirs[0].fd = stdoutFd;
	redirs[0].priority = ANDROID_LOG_INFO;

	redirs[1].buffer_cursor = 0;
	redirs[1].fd = stderrFd;
	redirs[1].priority = ANDROID_LOG_ERROR;

	LOG_DEBUG("starting stdout & stderr redirection thread");
	errno = pthread_create(&thread_redir, NULL, _redir_entrypoint, NULL);
	if (errno != 0) {
		// error will be caught by Go
		return;
	}

	errno = pthread_detach(thread_redir);
}
*/
import "C"

import (
	"log"
	"os"
	"syscall"
	"unsafe"
)

var (
	ctag = C.CString("GoLog")
	// Store the reader and writer end of the redirected stderr
	// and stdout so that they are not garbage collected and closed.
	stderrW, stdoutW *os.File
	stderrR, stdoutR *os.File
)

type infoWriter struct{}

func (infoWriter) Write(p []byte) (n int, err error) {
	cstr := C.CString(string(p))
	C.__android_log_write(C.ANDROID_LOG_INFO, ctag, cstr)
	C.free(unsafe.Pointer(cstr))
	return len(p), nil
}

func init() {
	log.SetOutput(infoWriter{})
	// android logcat includes all of log.LstdFlags
	log.SetFlags(log.Flags() &^ log.LstdFlags)

	var err error
	stderrR, stderrW, err = os.Pipe()
	if err != nil {
		panic(err)
	}
	if err := syscall.Dup3(int(stderrW.Fd()), int(os.Stderr.Fd()), 0); err != nil {
		panic(err)
	}

	stdoutR, stdoutW, err = os.Pipe()
	if err != nil {
		panic(err)
	}
	if err := syscall.Dup3(int(stdoutW.Fd()), int(os.Stdout.Fd()), 0); err != nil {
		panic(err)
	}

	if _, err := C.initRedirection(C.int(stdoutR.Fd()), C.int(stderrR.Fd())); err != nil {
		panic(err)
	}
}
