// Copyright 2026 The Enjin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// FilesDir support for the engine's asset cache. Exposes the absolute path
// of the running app's internal storage directory (Context.getFilesDir()),
// which is the only writable location guaranteed to persist across boots on
// Android. Not in upstream golang.org/x/mobile.

package asset

/*
#cgo LDFLAGS: -landroid
#include <android/asset_manager.h>
#include <android/asset_manager_jni.h>
#include <jni.h>
#include <stdlib.h>
#include <string.h>

// Calls ctx.getFilesDir().getAbsolutePath() and returns a malloc'd C string
// the caller must free. Returns NULL on any failure.
static char* files_dir_path(uintptr_t java_vm, uintptr_t jni_env, jobject ctx) {
	JavaVM* vm = (JavaVM*)java_vm;
	JNIEnv* env = (JNIEnv*)jni_env;
	(void)vm;

	jclass ctx_clazz = (*env)->FindClass(env, "android/content/Context");
	if (ctx_clazz == NULL) return NULL;
	jmethodID get_files_dir = (*env)->GetMethodID(env, ctx_clazz, "getFilesDir", "()Ljava/io/File;");
	if (get_files_dir == NULL) return NULL;
	jobject file = (*env)->CallObjectMethod(env, ctx, get_files_dir);
	if (file == NULL) return NULL;

	jclass file_clazz = (*env)->FindClass(env, "java/io/File");
	if (file_clazz == NULL) return NULL;
	jmethodID get_abs_path = (*env)->GetMethodID(env, file_clazz, "getAbsolutePath", "()Ljava/lang/String;");
	if (get_abs_path == NULL) return NULL;
	jstring jpath = (jstring)(*env)->CallObjectMethod(env, file, get_abs_path);
	if (jpath == NULL) return NULL;

	const char* utf = (*env)->GetStringUTFChars(env, jpath, NULL);
	if (utf == NULL) return NULL;
	char* dup = strdup(utf);
	(*env)->ReleaseStringUTFChars(env, jpath, utf);
	return dup;
}
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"vortex.studio/mobile/internal/mobileinit"
)

var (
	filesDirOnce sync.Once
	filesDirVal  string
	filesDirErr  error
)

// FilesDir returns the absolute filesystem path of the running app's
// internal storage directory (the Java-level Context.getFilesDir()). Used
// by the engine's asset cache to pick a write location that survives across
// app launches.
//
// FilesDir is cached after the first successful call; the JVM context is
// process-stable, so re-resolving on every call would just waste JNI hops.
func FilesDir() (string, error) {
	filesDirOnce.Do(func() {
		err := mobileinit.RunOnJVM(func(vm, env, ctx uintptr) error {
			cstr := C.files_dir_path(C.uintptr_t(vm), C.uintptr_t(env), C.jobject(ctx))
			if cstr == nil {
				return fmt.Errorf("asset: files_dir_path returned NULL (JNI lookup failed)")
			}
			defer C.free(unsafe.Pointer(cstr))
			filesDirVal = C.GoString(cstr)
			return nil
		})
		if err != nil {
			filesDirErr = fmt.Errorf("asset: resolve filesDir: %w", err)
		}
	})
	return filesDirVal, filesDirErr
}
