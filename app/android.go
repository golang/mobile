// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

// Go runtime entry point for apps running on android.
// Sets up everything the runtime needs and exposes
// the entry point to JNI.

package app

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
#include <dlfcn.h>
#include <errno.h>
#include <fcntl.h>
#include <jni.h>
#include <pthread.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

pthread_cond_t go_started_cond;
pthread_mutex_t go_started_mu;
int go_started;

static int (*_rt0_arm_linux1)(int argc, char** argv);

jint JNI_OnLoad(JavaVM* vm, void* reserved) {
	JNIEnv* env;
	if ((*vm)->GetEnv(vm, (void**)&env, JNI_VERSION_1_6) != JNI_OK) {
		return -1;
	}

	_rt0_arm_linux1 = (int (*)(int, char**))dlsym(RTLD_DEFAULT, "_rt0_arm_linux1");
	if (_rt0_arm_linux1 == NULL) {
		__android_log_print(ANDROID_LOG_FATAL, "Go", "missing _rt0_arm_linux1");
	}

        pthread_mutex_lock(&go_started_mu);
	go_started = 0;
	pthread_mutex_unlock(&go_started_mu);
	pthread_cond_init(&go_started_cond, NULL);

	return JNI_VERSION_1_6;
}

// Runtime entry point.
JNIEXPORT void JNICALL
Java_go_Go_run(JNIEnv* env, jclass clazz) {
	// Defensively heap-allocate argv0, for setenv.
	char* argv0 = strdup("gojni");

	// Build argv, including the ELF auxiliary vector, which is loaded
	// from /proc/self/auxv. While there does not appear to be any
	// spec for this format, there are some notes in
	//
	// Phrack, V. 0x0b, Issue 0x3a, P. 0x05.
	// http://phrack.org/issues/58/5.html
	//
	// For our needs, we don't need to know the format beyond the fact
	// that argv is followed by a meaningless envp, then a series of
	// null terminated bytes that make up auxv.

	struct {
		char* argv[2];
		char* envp[2];
		char* auxv[1024];
	} x;
	x.argv[0] = argv0;
	x.argv[1] = NULL;
	x.envp[0] = argv0;
	x.envp[1] = NULL;

	int fd = open("/proc/self/auxv", O_RDONLY, 0);
	if (fd == -1) {
		__android_log_print(ANDROID_LOG_FATAL, "Go", "cannot open /proc/self/auxv: %s", strerror(errno));
	}
	int n = read(fd, &x.auxv, sizeof x.auxv - 1);
	if (n < 0) {
		__android_log_print(ANDROID_LOG_FATAL, "Go", "error reading /proc/self/auxv: %s", strerror(errno));
	}
	if (n == sizeof x.auxv - 1) { // x.auxv should be more than plenty.
		__android_log_print(ANDROID_LOG_FATAL, "Go", "/proc/self/auxv too big");
	}
	close(fd);
	x.auxv[n] = NULL;

        int32_t argc = 1;
        _rt0_arm_linux1(argc, x.argv);
}

// Used by Java initialization code to know when it can use cgocall.
JNIEXPORT void JNICALL
Java_go_Go_waitForRun(JNIEnv* env, jclass clazz) {
	pthread_mutex_lock(&go_started_mu);
	while (go_started == 0) {
		pthread_cond_wait(&go_started_cond, &go_started_mu);
	}
	pthread_mutex_unlock(&go_started_mu);
	__android_log_print(ANDROID_LOG_INFO, "Go", "gojni has started");
}
*/
import "C"
import "unsafe"

func run() {
	// TODO(crawshaw): replace os.Stderr / os.Stdio.

	ctag := C.CString("Go")
	cstr := C.CString("android.Run started")
	C.__android_log_write(C.ANDROID_LOG_INFO, ctag, cstr)
	C.free(unsafe.Pointer(ctag))
	C.free(unsafe.Pointer(cstr))

	// Inform Java that the program is initialized.
	C.pthread_mutex_lock(&C.go_started_mu)
	C.go_started = 1
	C.pthread_cond_signal(&C.go_started_cond)
	C.pthread_mutex_unlock(&C.go_started_mu)

	select {}
}
