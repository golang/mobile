// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

void init_seq(void* vm, void* classfinder);
JNIEnv *get_thread_env(void);
void recv(int32_t ref, int code, uint8_t *in_ptr, size_t in_len, uint8_t **out_ptr, size_t *out_len);
