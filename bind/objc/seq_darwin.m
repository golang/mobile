// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <Foundation/Foundation.h>
#include "seq.h"

#ifdef DEBUG
#define LOG_DEBUG(...) NSLog(__VA_ARGS__);
#else
#define LOG_DEBUG(...) ;
#endif

#define LOG_INFO(...) NSLog(__VA_ARGS__);
#define LOG_FATAL(...)                                                         \
  @throw                                                                       \
      [NSException exceptionWithName:NSInternalInconsistencyException          \
                              reason:[NSString stringWithFormat:__VA_ARGS__]   \
                            userInfo:NULL];

// mem is a simple C equivalent of seq.Buffer.
//
// Many of the allocations around mem could be avoided to improve
// function call performance, but the goal is to start simple.
typedef struct mem {
  uint8_t *buf;
  size_t off;
  size_t len;
  size_t cap;

  // TODO(hyangah): array support. See code around struct pinned in
  // bind/javaj/seq_android.c
} mem;

// mem_ensure ensures that m has at least size bytes free.
// If m is NULL, it is created.
static mem *mem_ensure(mem *m, uint32_t size) {
  if (m == NULL) {
    m = (mem *)malloc(sizeof(mem));
    if (m == NULL) {
      LOG_FATAL(@"mem_ensure malloc failed");
    }
    m->cap = 0;
    m->off = 0;
    m->len = 0;
    m->buf = NULL;
  }
  if (m->cap > m->off + size) {
    return m;
  }
  m->buf = (uint8_t *)realloc((void *)m->buf, m->off + size);
  if (m->buf == NULL) {
    LOG_FATAL(@"mem_ensure realloc failed, off=%zu, size=%u", m->off, size);
  }
  m->cap = m->off + size;
  return m;
}

static uint32_t align(uint32_t offset, uint32_t alignment) {
  uint32_t pad = offset % alignment;
  if (pad > 0) {
    pad = alignment - pad;
  }
  return pad + offset;
}

static uint8_t *mem_read(GoSeq *seq, uint32_t size, uint32_t alignment) {
  mem *m = (mem *)(seq->mem_ptr);
  if (size == 0) {
    return NULL;
  }
  if (m == NULL) {
    LOG_FATAL(@"mem_read on NULL mem");
  }
  uint32_t offset = align(m->off, alignment);

  if (m->len - offset < size) {
    LOG_FATAL(@"short read");
  }
  uint8_t *res = m->buf + offset;
  m->off = offset + size;
  return res;
}

uint8_t *mem_write(GoSeq *seq, uint32_t size, uint32_t alignment) {
  mem *m = (mem *)(seq->mem_ptr);
  if (m == NULL) {
    LOG_FATAL(@"mem_write on NULL mem");
  }
  if (m->off != m->len) {
    LOG_FATAL(@"write can only append to seq, size: (off=%zu len=%zu, size=%u)",
              m->off, m->len, size);
  }
  uint32_t offset = align(m->off, alignment);
  uint32_t cap = m->cap;
  while (offset + size > cap) {
    cap *= 2;
  }
  m = mem_ensure(m, cap);
  uint8_t *res = m->buf + offset;
  m->off = offset + size;
  m->len = offset + size;
  return res;
}

// extern
void go_seq_init(GoSeq *seq) { seq->mem_ptr = mem_ensure(seq->mem_ptr, 64); }

// extern
void go_seq_free(GoSeq *seq) {
  mem *m = (mem *)(seq->mem_ptr);
  if (m != NULL) {
    free(m->buf);
    free(m);
  }
}
