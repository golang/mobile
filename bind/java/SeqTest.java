package go;

import go.testpkg.Testpkg;

import junit.framework.TestCase;

public class SeqTest extends TestCase {
	static {
		Go.init(null);
	}

	public void testAdd() {
		long res = Testpkg.instance.Add(3, 4);
		assertEquals("Unexpected arithmetic failure", 7, res);
	}

	public void testGoRefGC() {
		Testpkg.S s = Testpkg.instance.New();
		System.gc();
		Testpkg.instance.GC();
		long collected = Testpkg.instance.NumSCollected();
		assertEquals("Only S should be pinned", 0, collected);

		s = null;
		System.gc();
		Testpkg.instance.GC();
		collected = Testpkg.instance.NumSCollected();
		assertEquals("S should be collected", 1, collected);
	}

	boolean finalizedAnI;

	private class AnI extends Testpkg.I.Stub {
		boolean called;
		public void F() {
			called = true;
		}
		@Override
		public void finalize() throws Throwable {
			finalizedAnI = true;
			super.finalize();
		}
	}

	public void testJavaRefGC() {
		finalizedAnI = false;
		AnI obj = new AnI();
		Testpkg.instance.Call(obj);
		assertTrue("want F to be called", obj.called);
		obj = null;
		Testpkg.instance.GC();
		System.gc();
		assertTrue("want obj to be collected", finalizedAnI);
	}

	public void testJavaRefKeep() {
		finalizedAnI = false;
		AnI obj = new AnI();
		Testpkg.instance.Keep(obj);
		obj = null;
		System.gc();
		assertFalse("want obj to be kept live by Go", finalizedAnI);
	}
}
