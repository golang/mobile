// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go;

import java.util.Arrays;
import java.util.logging.Logger;

// Seq is a sequence of machine-dependent encoded values.
// Used by automatically generated language bindings to talk to Go.
public class Seq {
	private static Logger log = Logger.getLogger("GoSeq");

	static {
		// Look for the shim class auto-generated by gomobile bind.
		// Its only purpose is to call System.loadLibrary.
		try {
			Class loadJNI = Class.forName("go.LoadJNI");
			setContext(loadJNI.getDeclaredField("ctx").get(null));
		} catch (ClassNotFoundException e) {
			// Ignore, assume the user will load JNI for it.
			log.warning("LoadJNI class not found");
		} catch (NoSuchFieldException e) {
			log.severe("LoadJNI class missing field: " + e);
		} catch (IllegalAccessException e) {
			log.severe("LoadJNI class bad field: " + e);
		}

		initSeq();
	}

	@SuppressWarnings("UnusedDeclaration")
	private long memptr; // holds C-allocated pointer

	public Seq() {
		ensure(64);
	}

	// ctx is an android.context.Context.
	static native void setContext(java.lang.Object ctx);

	// Ensure that at least size bytes can be written to the Seq.
	// Any existing data in the buffer is preserved.
	public native void ensure(int size);

	// Moves the internal buffer offset back to zero.
	// Length and contents are maintained. Data can be read after a reset.
	public native void resetOffset();

	public native void log(String label);

	public native boolean readBool();
	public native byte readInt8();
	public native short readInt16();
	public native int readInt32();
	public native long readInt64();
	public long readInt() { return readInt64(); }

	public native float readFloat32();
	public native double readFloat64();
	public native String readUTF16();
	public String readString() { return readUTF16(); }
	public native byte[] readByteArray();

	public native void writeBool(boolean v);
	public native void writeInt8(byte v);
	public native void writeInt16(short v);
	public native void writeInt32(int v);
	public native void writeInt64(long v);
	public void writeInt(long v) { writeInt64(v); }

	public native void writeFloat32(float v);
	public native void writeFloat64(double v);
	public native void writeUTF16(String v);
	public void writeString(String v) { writeUTF16(v); }
	public native void writeByteArray(byte[] v);

	public void writeRef(Ref ref) {
		if (ref == null)
			ref = RefTracker.nullRef;
		tracker.inc(ref);
		writeInt32(ref.refnum);
	}

	public Ref readRef() {
		int refnum = readInt32();
		return tracker.get(refnum);
	}

	static native void initSeq();

	// Informs the Go ref tracker that Java is done with this ref.
	static native void destroyRef(int refnum);

	// createRef creates a Ref to a Java object.
	public static Ref createRef(Seq.Object o) {
		return tracker.createRef(o);
	}

	// sends a function invocation request to Go.
	//
	// Blocks until the function completes.
	// If the request is for a method, the first element in src is
	// a Ref to the receiver.
	public static native void send(String descriptor, int code, Seq src, Seq dst);

	protected void finalize() throws Throwable {
		super.finalize();
		free();
	}
	private native void free();

	public static Seq recv(Seq in, int code, int refnum) {
		Seq out = new Seq();
		if (code == -1) {
			// Special signal from seq.FinalizeRef.
			tracker.dec(refnum);
			return out;
		}

		Ref r = tracker.get(refnum);
		r.obj.call(code, in, out);
		return out;
	}

	// An Object is a Java object that matches a Go object.
	// The implementation of the object may be in either Java or Go,
	// with a proxy instance in the other language passing calls
	// through to the other language.
	//
	// Don't implement an Object directly. Instead, look for the
	// generated abstract Stub.
	public interface Object {
		public Ref ref();
		public void call(int code, Seq in, Seq out);
	}

	// A Ref is an object tagged with an integer for passing back and
	// forth across the language boundary.
	//
	// A Ref may represent either an instance of a Java Object subclass,
	// or an instance of a Go object. The explicit allocation of a Ref
	// is used to pin Go object instances when they are passed to Java.
	// The Go Seq library maintains a reference to the instance in a map
	// keyed by the Ref number. When the JVM calls finalize, we ask Go
	// to clear the entry in the map.
	public static final class Ref {
		// refnum < 0: Go object tracked by Java
		// refnum > 0: Java object tracked by Go
		int refnum;

		int refcnt;  // for Java obj: track how many times sent to Go.

		public Seq.Object obj;  // for Java obj: pointers to the Java obj.

		Ref(int refnum, Seq.Object o) {
			this.refnum = refnum;
			this.refcnt = 0;
			this.obj = o;
		}

		@Override
		protected void finalize() throws Throwable {
			if (refnum < 0) {
				// Go object: signal Go to decrement the reference count.
				Seq.destroyRef(refnum);
			}
			super.finalize();
		}
	}

	static final RefTracker tracker = new RefTracker();

	static final class RefTracker {
		private static final int REF_OFFSET = 42;
		private static final int NULL_REFNUM = 41; // also known to bind/seq/ref.go

		// use single Ref for null Seq.Object
		private static final Ref nullRef = new Ref(NULL_REFNUM, null);

		// Next Java object reference number.
		//
		// Reference numbers are positive for Java objects,
		// and start, arbitrarily at a different offset to Go
		// to make debugging by reading Seq hex a little easier.
		private int next = REF_OFFSET; // next Java object ref

		// Java objects that have been passed to Go. refnum -> Ref
		// The Ref obj field is non-null.
		// This map pins Java objects so they don't get GCed while the
		// only reference to them is held by Go code.
		private RefMap javaObjs = new RefMap();

		// inc increments the reference count of a Java object when it
		// is sent to Go.
		synchronized void inc(Ref ref) {
			int refnum = ref.refnum;
			if (refnum <= 0) {
				// We don't keep track of the Go object.
				return;
			}
			if (refnum == nullRef.refnum) {
				return;
			}
			// Count how many times this ref's Java object is passed to Go.
			if (ref.refcnt == Integer.MAX_VALUE) {
				throw new RuntimeException("refnum " + refnum + " overflow");
			}
			ref.refcnt++;
			Ref obj = javaObjs.get(refnum);
			if (obj == null) {
				javaObjs.put(refnum, ref);
			}
		}

		// dec decrements the reference count of a Java object when
		// Go signals a corresponding proxy object is finalized.
		// If the count reaches zero, the Java object is removed
		// from the javaObjs map.
		synchronized void dec(int refnum) {
			if (refnum <= 0) {
				// We don't keep track of the Go object.
				// This must not happen.
				log.severe("dec request for Go object "+ refnum);
				return;
			}
			if (refnum == nullRef.refnum) {
				return;
			}
			// Java objects are removed on request of Go.
			Ref obj = javaObjs.get(refnum);
			if (obj == null) {
				throw new RuntimeException("referenced Java object is not found: refnum="+refnum);
			}
			obj.refcnt--;
			if (obj.refcnt <= 0) {
				javaObjs.remove(refnum);
			}
		}

		synchronized Ref createRef(Seq.Object o) {
			if (o == null) {
				return nullRef;
			}
			if (next == Integer.MAX_VALUE) {
				throw new RuntimeException("createRef overflow for " + o);
			}
			int refnum = next++;
			Ref ref = new Ref(refnum, o);
			return ref;
		}

		// get returns an existing Ref to either a Java or Go object.
		// It may be the first time we have seen the Go object.
		//
		// TODO(crawshaw): We could cut down allocations for frequently
		// sent Go objects by maintaining a map to weak references. This
		// however, would require allocating two objects per reference
		// instead of one. It also introduces weak references, the bane
		// of any Java debugging session.
		//
		// When we have real code, examine the tradeoffs.
		synchronized Ref get(int refnum) {
			if (refnum > 0) {
				if (refnum == nullRef.refnum) {
					return nullRef;
				}
				Ref ref = javaObjs.get(refnum);
				if (ref == null) {
					throw new RuntimeException("unknown java Ref: "+refnum);
				}
				return ref;
			} else {
				// Go object.
				return new Ref(refnum, null);
			}
		}
	}

	// RefMap is a mapping of integers to Ref objects.
	//
	// The integers can be sparse. In Go this would be a map[int]*Ref.
	static final class RefMap {
		private int next = 0;
		private int live = 0;
		private int[] keys = new int[16];
		private Ref[] objs = new Ref[16];

		RefMap() {}

		Ref get(int key) {
			int i = Arrays.binarySearch(keys, 0, next, key);
			if (i >= 0) {
				return objs[i];
			}
			return null;
		}

		void remove(int key) {
			int i = Arrays.binarySearch(keys, 0, next, key);
			if (i >= 0) {
				if (objs[i] != null) {
					objs[i] = null;
					live--;
				}
			}
		}

		void put(int key, Ref obj) {
			if (obj == null) {
				throw new RuntimeException("put a null ref (with key "+key+")");
			}
			int i = Arrays.binarySearch(keys, 0, next, key);
			if (i >= 0) {
				if (objs[i] == null) {
					objs[i] = obj;
					live++;
				}
				if (objs[i] != obj) {
					throw new RuntimeException("replacing an existing ref (with key "+key+")");
				}
				return;
			}
			if (next >= keys.length) {
				grow();
				i = Arrays.binarySearch(keys, 0, next, key);
			}
			i = ~i;
			if (i < next) {
				// Insert, shift everything afterwards down.
				System.arraycopy(keys, i, keys, i+1, next-i);
				System.arraycopy(objs, i, objs, i+1, next-i);
			}
			keys[i] = key;
			objs[i] = obj;
			live++;
			next++;
		}

		private void grow() {
			// Compact and (if necessary) grow backing store.
			int[] newKeys;
			Ref[] newObjs;
			int len = 2*roundPow2(live);
			if (len > keys.length) {
				newKeys = new int[keys.length*2];
				newObjs = new Ref[objs.length*2];
			} else {
				newKeys = keys;
				newObjs = objs;
			}

			int j = 0;
			for (int i = 0; i < keys.length; i++) {
				if (objs[i] != null) {
					newKeys[j] = keys[i];
					newObjs[j] = objs[i];
					j++;
				}
			}
			for (int i = j; i < newKeys.length; i++) {
				newKeys[i] = 0;
				newObjs[i] = null;
			}

			keys = newKeys;
			objs = newObjs;
			next = j;

			if (live != next) {
				throw new RuntimeException("bad state: live="+live+", next="+next);
			}
		}

		private static int roundPow2(int x) {
			int p = 1;
			while (p < x) {
				p *= 2;
			}
			return p;
		}
	}
}
