// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <android/log.h>
#include <errno.h>
#include <jni.h>
#include <stdint.h>
#include <stdio.h>
#include <unistd.h>
#include "seq_android.h"
#include "_cgo_export.h"

#define LOG_INFO(...) __android_log_print(ANDROID_LOG_INFO, "go/Seq", __VA_ARGS__)
#define LOG_FATAL(...) __android_log_print(ANDROID_LOG_FATAL, "go/Seq", __VA_ARGS__)

static jfieldID memptr_id;
static jfieldID receive_refnum_id;
static jfieldID receive_code_id;
static jfieldID receive_handle_id;

// mem is a simple C equivalent of seq.Buffer.
//
// Many of the allocations around mem could be avoided to improve
// function call performance, but the goal is to start simple.
typedef struct mem {
	uint8_t *buf;
	uint32_t off;
	uint32_t len;
	uint32_t cap;
} mem;

static mem *mem_resize(mem *m, uint32_t size) {
	if (m == NULL) {
		m = (mem*)malloc(sizeof(mem));
		if (m == NULL) {
			LOG_FATAL("mem_resize malloc failed");
		}
		m->off = 0;
		m->len = 0;
		m->buf = NULL;
	}
	m->buf = (uint8_t*)realloc((void*)m->buf, size);
	if (m->buf == NULL) {
		LOG_FATAL("mem_resize realloc failed, size=%d", size);
	}
	m->cap = size;
	return m;
}

static mem *mem_get(JNIEnv *env, jobject obj) {
	// Storage space for pointer is always 64-bits, even on 32-bit
	// machines. Cast to uintptr_t to avoid -Wint-to-pointer-cast.
	return (mem*)(uintptr_t)(*env)->GetLongField(env, obj, memptr_id);
}

static uint8_t *mem_read(JNIEnv *env, jobject obj, uint32_t size) {
	mem *m = mem_get(env, obj);
	if (m == NULL) {
		LOG_FATAL("mem_read on NULL mem");
	}
	if (m->len-m->off < size) {
		LOG_FATAL("short read, size: %d", size);
	}
	uint8_t *res = m->buf+m->off;
	m->off += size;
	return res;
}

uint8_t *mem_write(JNIEnv *env, jobject obj, uint32_t size) {
	mem *m = mem_get(env, obj);
	if (m == NULL) {
		LOG_FATAL("mem_write on NULL mem");
	}
	if (m->off != m->len) {
		LOG_FATAL("write can only append to seq, size: (off=%d, len=%d, size=%d", m->off, m->len, size);
	}
	if (m->off+size > m->cap) {
		m = mem_resize(m, 2*m->cap);
	}
	uint8_t *res = m->buf+m->off;
	m->off += size;
	m->len += size;
	return res;
}

static jfieldID find_field(JNIEnv *env, const char *class_name, const char *field_name, const char *field_type) {
	jclass clazz = (*env)->FindClass(env, class_name);
	if (clazz == NULL) {
		LOG_FATAL("cannot find %s", class_name);
		return NULL;
	}
	jfieldID id = (*env)->GetFieldID(env, clazz, field_name , field_type);
	if(id == NULL) {
		LOG_FATAL("no %s/%s field", field_name, field_type);
		return NULL;
	}
	return id;
}

void init_seq(void *javavm) {
	JavaVM *vm = (JavaVM*)javavm;
	JNIEnv *env;
	if ((*vm)->GetEnv(vm, (void**)&env, JNI_VERSION_1_6) != JNI_OK) {
		LOG_FATAL("bad vm env");
	}

	memptr_id = find_field(env, "go/Seq", "memptr", "J");
	receive_refnum_id = find_field(env, "go/Seq$Receive", "refnum", "I");
	receive_handle_id = find_field(env, "go/Seq$Receive", "handle", "I");
	receive_code_id = find_field(env, "go/Seq$Receive", "code", "I");

	LOG_INFO("loaded go/Seq");
}

JNIEXPORT void JNICALL
Java_go_Seq_ensure(JNIEnv *env, jobject obj, jint size) {
	mem *m = mem_get(env, obj);
	if (m == NULL || size > m->cap - m->off) {
		m = mem_resize(m, size);
		(*env)->SetLongField(env, obj, memptr_id, (jlong)(uintptr_t)m);
	}
}

JNIEXPORT void JNICALL
Java_go_Seq_free(JNIEnv *env, jobject obj) {
	mem *m = mem_get(env, obj);
	if (m != NULL) {
		free((void*)m->buf);
		free((void*)m);
	}
}

#define MEM_READ(obj, ty) ((ty*)mem_read(env, obj, sizeof(ty)))

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
	return (*env)->NewString(env, (jchar*)mem_read(env, obj, 2*size), size);
}

#define MEM_WRITE(ty) (*(ty*)mem_write(env, obj, sizeof(ty)))

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
	(*env)->GetStringRegion(env, v, 0, size, (jchar*)mem_write(env, obj, 2*size));
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
	mem *src = mem_get(env, src_obj);
	if (src == NULL) {
		LOG_FATAL("send src is NULL");
	}
	mem *dst = mem_get(env, dst_obj);
	if (dst == NULL) {
		LOG_FATAL("send dst is NULL");
	}

	GoString desc;
	desc.p = (char*)(*env)->GetStringUTFChars(env, descriptor, NULL);
	if (desc.p == NULL) {
		LOG_FATAL("send GetStringUTFChars failed");
	}
	desc.n = (*env)->GetStringUTFLength(env, descriptor);
	Send(desc, (GoInt)code, src->buf, src->len, &dst->buf, &dst->len);
	(*env)->ReleaseStringUTFChars(env, descriptor, desc.p);
}

JNIEXPORT void JNICALL
Java_go_Seq_recv(JNIEnv *env, jclass clazz, jobject in_obj, jobject receive) {
	mem *in = mem_get(env, in_obj);
	if (in == NULL) {
		LOG_FATAL("recv in is NULL");
	}
	struct Recv_return ret = Recv(&in->buf, &in->len);
	(*env)->SetIntField(env, receive, receive_refnum_id, ret.r0);
	(*env)->SetIntField(env, receive, receive_code_id, ret.r1);
	(*env)->SetIntField(env, receive, receive_handle_id, ret.r2);
}

JNIEXPORT void JNICALL
Java_go_Seq_recvRes(JNIEnv *env, jclass clazz, jint handle, jobject out_obj) {
	mem *out = mem_get(env, out_obj);
	if (out == NULL) {
		LOG_FATAL("recvRes out is NULL");
	}
	RecvRes((int32_t)handle, out->buf, out->len);
}
