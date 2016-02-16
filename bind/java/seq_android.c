// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <android/log.h>
#include <errno.h>
#include <jni.h>
#include <stdint.h>
#include <stdio.h>
#include <unistd.h>
#include <pthread.h>
#include "seq_android.h"
#include "_cgo_export.h"

#define LOG_INFO(...) __android_log_print(ANDROID_LOG_INFO, "go/Seq", __VA_ARGS__)
#define LOG_FATAL(...) __android_log_print(ANDROID_LOG_FATAL, "go/Seq", __VA_ARGS__)

static jfieldID memptr_id;

static jclass jbytearray_clazz;

static jclass seq_clazz;
static jmethodID seq_cons;
static jmethodID seq_recv;

static JavaVM *jvm;
// jnienvs holds the per-thread JNIEnv* for Go threads where we called AttachCurrentThread.
// A pthread key destructor is supplied to call DetachCurrentThread on exit. This trick is
// documented in http://developer.android.com/training/articles/perf-jni.html under "Threads".
static pthread_key_t jnienvs;

// pinned represents a pinned array to be released at the end of Send call.
typedef struct pinned {
	jobject ref;
	void* ptr;
	struct pinned* next;
} pinned;

// mem is a simple C equivalent of seq.Buffer.
//
// Many of the allocations around mem could be avoided to improve
// function call performance, but the goal is to start simple.
typedef struct mem {
	uint8_t *buf;
	size_t off;
	size_t len;
	size_t cap;

	// TODO(hyangah): have it as a separate field outside mem?
	pinned* pinned;
} mem;

// mem_ensure ensures that m has at least size bytes free.
// If m is NULL, it is created.
static mem *mem_ensure(mem *m, uint32_t size) {
	if (m == NULL) {
		m = (mem*)malloc(sizeof(mem));
		if (m == NULL) {
			LOG_FATAL("mem_ensure malloc failed");
		}
		m->cap = 0;
		m->off = 0;
		m->len = 0;
		m->buf = NULL;
		m->pinned = NULL;
	}
	uint32_t cap = m->cap;
	if (m->cap > m->off+size) {
		return m;
	}
	if (cap == 0) {
		cap = 64;
	}
	// TODO(hyangah): consider less aggressive allocation such as
	//   cap += max(pow2round(size), 64)
	while (cap < m->off+size) {
		cap *= 2;
	}
	m->buf = (uint8_t*)realloc((void*)m->buf, cap);
	if (m->buf == NULL) {
		LOG_FATAL("mem_ensure realloc failed, off=%d, size=%d", m->off, size);
	}
	m->cap = cap;
	return m;
}

static mem *mem_get(JNIEnv *env, jobject obj) {
	if (obj == NULL) {
		return NULL;
	}
	// Storage space for pointer is always 64-bits, even on 32-bit
	// machines. Cast to uintptr_t to avoid -Wint-to-pointer-cast.
	return (mem*)(uintptr_t)(*env)->GetLongField(env, obj, memptr_id);
}

static uint32_t align(uint32_t offset, uint32_t alignment) {
	uint32_t pad = offset % alignment;
	if (pad > 0) {
		pad = alignment-pad;
	}
	return pad+offset;
}

static uint8_t *mem_read(JNIEnv *env, jobject obj, uint32_t size, uint32_t alignment) {
	if (size == 0) {
		return NULL;
	}
	mem *m = mem_get(env, obj);
	if (m == NULL) {
		LOG_FATAL("mem_read on NULL mem");
	}
	uint32_t offset = align(m->off, alignment);

	if (m->len-offset < size) {
		LOG_FATAL("short read");
	}
	uint8_t *res = m->buf+offset;
	m->off = offset+size;
	return res;
}

uint8_t *mem_write(JNIEnv *env, jobject obj, uint32_t size, uint32_t alignment) {
	mem *m = mem_get(env, obj);
	if (m == NULL) {
		LOG_FATAL("mem_write on NULL mem");
	}
	if (m->off != m->len) {
		LOG_FATAL("write can only append to seq, size: (off=%d, len=%d, size=%d", m->off, m->len, size);
	}
	uint32_t offset = align(m->off, alignment);
	m = mem_ensure(m, offset - m->off + size);
	uint8_t *res = m->buf+offset;
	m->off = offset+size;
	m->len = offset+size;
	return res;
}

static void *pin_array(JNIEnv *env, jobject obj, jobject arr) {
	mem *m = mem_get(env, obj);
	if (m == NULL) {
		m = mem_ensure(m, 64);
	}
	pinned *p = (pinned*) malloc(sizeof(pinned));
	if (p == NULL) {
		LOG_FATAL("pin_array malloc failed");
	}
	p->ref = (*env)->NewGlobalRef(env, arr);

	if ((*env)->IsInstanceOf(env, p->ref, jbytearray_clazz)) {
		p->ptr = (*env)->GetByteArrayElements(env, p->ref, NULL);
	} else {
		LOG_FATAL("unsupported array type");
	}

	p->next = m->pinned;
	m->pinned = p;
	return p->ptr;
}

static void unpin_arrays(JNIEnv *env, mem *m) {
	pinned* p = m->pinned;
	while (p != NULL) {
		if ((*env)->IsInstanceOf(env, p->ref, jbytearray_clazz)) {
			(*env)->ReleaseByteArrayElements(env, p->ref, (jbyte*)p->ptr, JNI_ABORT);
		} else {
			LOG_FATAL("invalid array type");
		}

		(*env)->DeleteGlobalRef(env, p->ref);

		pinned* o = p;
		p = p->next;
		free(o);
	}
	m->pinned = NULL;
}

static void describe_exception(JNIEnv* env) {
	jthrowable exc = (*env)->ExceptionOccurred(env);
	if (exc) {
		(*env)->ExceptionDescribe(env);
		(*env)->ExceptionClear(env);
	}
}

static jfieldID find_field(JNIEnv *env, const char *class_name, const char *field_name, const char *field_type) {
	jclass clazz = (*env)->FindClass(env, class_name);
	if (clazz == NULL) {
		describe_exception(env);
		LOG_FATAL("cannot find %s", class_name);
		return NULL;
	}
	jfieldID id = (*env)->GetFieldID(env, clazz, field_name , field_type);
	if(id == NULL) {
		describe_exception(env);
		LOG_FATAL("no %s/%s field", field_name, field_type);
		return NULL;
	}
	return id;
}

static jclass find_class(JNIEnv *env, const char *class_name) {
	jclass clazz = (*env)->FindClass(env, class_name);
	if (clazz == NULL) {
		describe_exception(env);
		LOG_FATAL("cannot find %s", class_name);
		return NULL;
	}
	return (*env)->NewGlobalRef(env, clazz);
}

static jmethodID get_method_id(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	jmethodID m = (*env)->GetMethodID(env, clazz, name, sig);
	if (m == NULL) {
		describe_exception(env);
		LOG_FATAL("cannot find method %s", name);
	}
	return m;
}

static jmethodID get_static_method_id(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	jmethodID m = (*env)->GetStaticMethodID(env, clazz, name, sig);
	if (m == NULL) {
		describe_exception(env);
		LOG_FATAL("cannot find static method %s", name);
	}
	return m;
}

void recv(int32_t ref, int code, uint8_t *in_ptr, size_t in_len, uint8_t **out_ptr, size_t *out_len) {
	jobject out;
	mem *out_mem;
	mem *in_mem;
	JNIEnv *env;
	jobject in;
	jint ret;

	ret = (*jvm)->GetEnv(jvm, (void **)&env, JNI_VERSION_1_6);
	if (ret != JNI_OK) {
		if (ret != JNI_EDETACHED) {
			LOG_FATAL("failed to get thread env");
			return;
		}
		if ((*jvm)->AttachCurrentThread(jvm, &env, NULL) != JNI_OK) {
			LOG_FATAL("failed to attach current thread");
			return;
		}
		pthread_setspecific(jnienvs, env);
	}

	in = (*env)->NewObject(env, seq_clazz, seq_cons);
	if (in == NULL) {
		describe_exception(env);
		LOG_FATAL("cannot instantiate Seq");
		return;
	}
	in_mem = mem_get(env, in);
	if (in_mem == NULL) {
		LOG_FATAL("recv on NULL in_mem");
		return;
	}
	memcpy(mem_write(env, in, in_len, 1), in_ptr, in_len);
	in_mem->off = 0;
	out = (*env)->CallStaticObjectMethod(env, seq_clazz, seq_recv, in, code, ref);
	(*env)->DeleteLocalRef(env, in);
	if (out == NULL) {
		describe_exception(env);
		LOG_FATAL("failed to invoke Seq.recv");
		return;
	}
	out_mem = mem_get(env, out);
	(*env)->DeleteLocalRef(env, out);
	if (out_mem == NULL) {
		LOG_FATAL("recv on NULL out_mem");
		return;
	}
	*out_ptr = out_mem->buf;
	*out_len = out_mem->len;
}

// env_destructor is registered as a thread data key destructor to
// clean up a Go thread that is attached to the JVM.
static void env_destructor(void *env) {
	if ((*jvm)->DetachCurrentThread(jvm) != JNI_OK) {
		LOG_INFO("failed to detach current thread");
	}
}

JNIEXPORT void JNICALL
Java_go_Seq_initSeq(JNIEnv *env, jclass clazz) {
	seq_clazz = (*env)->NewGlobalRef(env, clazz);
	seq_recv = get_static_method_id(env, seq_clazz, "recv", "(Lgo/Seq;II)Lgo/Seq;");
	seq_cons = get_method_id(env, seq_clazz, "<init>", "()V");

	memptr_id = find_field(env, "go/Seq", "memptr", "J");

	jclass bclazz = find_class(env, "[B");
	jbytearray_clazz = (*env)->NewGlobalRef(env, bclazz);

	if ((*env)->GetJavaVM(env, &jvm) != 0) {
		LOG_FATAL("failed to get JVM");
		return;
	}
	if (pthread_key_create(&jnienvs, env_destructor) != 0) {
		LOG_FATAL("failed to initialize jnienvs thread local storage");
		return;
	}
}

JNIEXPORT void JNICALL
Java_go_Seq_ensure(JNIEnv *env, jobject obj, jint size) {
	mem *m = mem_get(env, obj);
	if (m == NULL || m->off+size > m->cap) {
		m = mem_ensure(m, size);
		(*env)->SetLongField(env, obj, memptr_id, (jlong)(uintptr_t)m);
	}
}

JNIEXPORT void JNICALL
Java_go_Seq_free(JNIEnv *env, jobject obj) {
	mem *m = mem_get(env, obj);
	if (m != NULL) {
		unpin_arrays(env, m);
		free((void*)m->buf);
		free((void*)m);
	}
}

#define MEM_READ(obj, ty) ((ty*)mem_read(env, obj, sizeof(ty), sizeof(ty)))

JNIEXPORT jboolean JNICALL
Java_go_Seq_readBool(JNIEnv *env, jobject obj) {
	int8_t *v = MEM_READ(obj, int8_t);
	if (v == NULL) {
		return 0;
	}
	return *v != 0 ? 1 : 0;
}

JNIEXPORT jbyte JNICALL
Java_go_Seq_readInt8(JNIEnv *env, jobject obj) {
	uint8_t *v = MEM_READ(obj, uint8_t);
	if (v == NULL) {
		return 0;
	}
	return *v;
}

JNIEXPORT jshort JNICALL
Java_go_Seq_readInt16(JNIEnv *env, jobject obj) {
	int16_t *v = MEM_READ(obj, int16_t);
	return v == NULL ? 0 : *v;
}

JNIEXPORT jint JNICALL
Java_go_Seq_readInt32(JNIEnv *env, jobject obj) {
	int32_t *v = MEM_READ(obj, int32_t);
	return v == NULL ? 0 : *v;
}

JNIEXPORT jlong JNICALL
Java_go_Seq_readInt64(JNIEnv *env, jobject obj) {
	int64_t *v = MEM_READ(obj, int64_t);
	return v == NULL ? 0 : *v;
}

JNIEXPORT jfloat JNICALL
Java_go_Seq_readFloat32(JNIEnv *env, jobject obj) {
	float *v = MEM_READ(obj, float);
	return v == NULL ? 0 : *v;
}

JNIEXPORT jdouble JNICALL
Java_go_Seq_readFloat64(JNIEnv *env, jobject obj) {
	double *v = MEM_READ(obj, double);
	return v == NULL ? 0 : *v;
}

JNIEXPORT jstring JNICALL
Java_go_Seq_readUTF16(JNIEnv *env, jobject obj) {
	int32_t size = *MEM_READ(obj, int32_t);
	if (size == 0) {
		return (*env)->NewString(env, NULL, 0);
	}
	return (*env)->NewString(env, (jchar*)mem_read(env, obj, 2*size, 1), size);
}

JNIEXPORT jbyteArray JNICALL
Java_go_Seq_readByteArray(JNIEnv *env, jobject obj) {
	// Send the (array length, pointer) pair encoded as two int64.
	// The pointer value is omitted if array length is 0.
	jlong size = Java_go_Seq_readInt64(env, obj);
	if (size == 0) {
		return NULL;
	}
	jbyteArray res = (*env)->NewByteArray(env, size);
	jlong ptr = Java_go_Seq_readInt64(env, obj);
	(*env)->SetByteArrayRegion(env, res, 0, size, (jbyte*)(intptr_t)(ptr));
	return res;
}

#define MEM_WRITE(ty) (*(ty*)mem_write(env, obj, sizeof(ty), sizeof(ty)))

JNIEXPORT void JNICALL
Java_go_Seq_writeBool(JNIEnv *env, jobject obj, jboolean v) {
	MEM_WRITE(int8_t) = v ? 1 : 0;
}

JNIEXPORT void JNICALL
Java_go_Seq_writeInt8(JNIEnv *env, jobject obj, jbyte v) {
	MEM_WRITE(int8_t) = v;
}

JNIEXPORT void JNICALL
Java_go_Seq_writeInt16(JNIEnv *env, jobject obj, jshort v) {
	MEM_WRITE(int16_t) = v;
}

JNIEXPORT void JNICALL
Java_go_Seq_writeInt32(JNIEnv *env, jobject obj, jint v) {
	MEM_WRITE(int32_t) = v;
}

JNIEXPORT void JNICALL
Java_go_Seq_writeInt64(JNIEnv *env, jobject obj, jlong v) {
	MEM_WRITE(int64_t) = v;
}

JNIEXPORT void JNICALL
Java_go_Seq_writeFloat32(JNIEnv *env, jobject obj, jfloat v) {
	MEM_WRITE(float) = v;
}

JNIEXPORT void JNICALL
Java_go_Seq_writeFloat64(JNIEnv *env, jobject obj, jdouble v) {
	MEM_WRITE(double) = v;
}

JNIEXPORT void JNICALL
Java_go_Seq_writeUTF16(JNIEnv *env, jobject obj, jstring v) {
	if (v == NULL) {
		MEM_WRITE(int32_t) = 0;
		return;
	}
	int32_t size = (*env)->GetStringLength(env, v);
	MEM_WRITE(int32_t) = size;
	(*env)->GetStringRegion(env, v, 0, size, (jchar*)mem_write(env, obj, 2*size, 1));
}

JNIEXPORT void JNICALL
Java_go_Seq_writeByteArray(JNIEnv *env, jobject obj, jbyteArray v) {
	// For Byte array, we pass only the (array length, pointer) pair
	// encoded as two int64 values. If the array length is 0,
	// the pointer value is omitted.
	if (v == NULL) {
		MEM_WRITE(int64_t) = 0;
		return;
	}

	jsize len = (*env)->GetArrayLength(env, v);
	MEM_WRITE(int64_t) = len;
	if (len == 0) {
		return;
	}

	jbyte* b = pin_array(env, obj, v);
	MEM_WRITE(int64_t) = (jlong)(uintptr_t)b;
}

JNIEXPORT void JNICALL
Java_go_Seq_resetOffset(JNIEnv *env, jobject obj) {
	mem *m = mem_get(env, obj);
	if (m == NULL) {
		LOG_FATAL("resetOffset on NULL mem");
	}
	m->off = 0;
}

JNIEXPORT void JNICALL
Java_go_Seq_log(JNIEnv *env, jobject obj, jstring v) {
	mem *m = mem_get(env, obj);
	const char *label = (*env)->GetStringUTFChars(env, v, NULL);
	if (label == NULL) {
		LOG_FATAL("log GetStringUTFChars failed");
	}
	if (m == NULL) {
		LOG_INFO("%s: mem=NULL", label);
	} else {
		LOG_INFO("%s: mem{off=%d, len=%d, cap=%d}", label, m->off, m->len, m->cap);
	}
	(*env)->ReleaseStringUTFChars(env, v, label);
}

JNIEXPORT void JNICALL
Java_go_Seq_destroyRef(JNIEnv *env, jclass clazz, jint refnum) {
	DestroyRef(refnum);
}

JNIEXPORT void JNICALL
Java_go_Seq_send(JNIEnv *env, jclass clazz, jstring descriptor, jint code, jobject src_obj, jobject dst_obj) {
	uint8_t* req = NULL;
	size_t reqlen = 0;
	mem *src = mem_get(env, src_obj);
	if (src != NULL) {
		req = src->buf;
		reqlen = src->len;
	}

	uint8_t** res = NULL;
	size_t* reslen = NULL;
	mem *dst = mem_get(env, dst_obj);
	if (dst != NULL) {
		res = &dst->buf;
		reslen = &dst->len;
	}

	GoString desc;
	desc.p = (char*)(*env)->GetStringUTFChars(env, descriptor, NULL);
	if (desc.p == NULL) {
		LOG_FATAL("send GetStringUTFChars failed");
	}
	desc.n = (*env)->GetStringUTFLength(env, descriptor);

	Send(desc, (GoInt)code, req, reqlen, res, reslen);
	(*env)->ReleaseStringUTFChars(env, descriptor, desc.p);

	if (src != NULL) {
		unpin_arrays(env, src);  // assume 'src' is no longer needed.
	}
}

JNIEXPORT void JNICALL
Java_go_Seq_setContext(JNIEnv* env, jclass clazz, jobject ctx) {
	JavaVM* vm;
        if ((*env)->GetJavaVM(env, &vm) != 0) {
		LOG_FATAL("failed to get JavaVM");
	}
	setContext(vm, (*env)->NewGlobalRef(env, ctx));
}
