package go;

import android.test.suitebuilder.annotation.Suppress;
import android.test.MoreAsserts;
import java.util.Arrays;
import java.util.Random;

import go.testpkg.Testpkg;

import junit.framework.TestCase;

public class SeqTest extends TestCase {
  static {
    Go.init(null);
  }

  public void testAdd() {
    long res = Testpkg.Add(3, 4);
    assertEquals("Unexpected arithmetic failure", 7, res);
  }

  public void testShortString() {
    String want = "a short string";
    String got = Testpkg.StrDup(want);
    assertEquals("Strings should match", want, got);
  }

  public void testLongString() {
    StringBuilder b = new StringBuilder();
    for (int i = 0; i < 128*1024; i++) {
      b.append("0123456789");
    }
    String want = b.toString();
    String got = Testpkg.StrDup(want);
    assertEquals("Strings should match", want, got);
  }

  public void testUnicode() {
    String want = "Hello, 世界";
    String got = Testpkg.StrDup(want);
    assertEquals("Strings should match", want, got);
  }

  public void testNilErr() throws Exception {
    Testpkg.Err(null); // returns nil, no exception
  }

  public void testErr() {
    String msg = "Go errors are dropped into the confusing space of exceptions";
    try {
      Testpkg.Err(msg);
      fail("expected non-nil error to be turned into an exception");
    } catch (Exception e) {
      assertEquals("messages should match", msg, e.getMessage());
    }
  }

  public void testByteArray() {
    for (int i = 0; i < 2048; i++) {
      if (i == 0) {
        byte[] got = Testpkg.BytesAppend(null, null);
        assertEquals("Bytes(null+null) should match", (byte[])null, got);
        got = Testpkg.BytesAppend(new byte[0], new byte[0]);
        assertEquals("Bytes(empty+empty) should match", (byte[])null, got);
        continue;
      }

      byte[] want = new byte[i];
      new Random().nextBytes(want);

      byte[] s1 = null;
      byte[] s2 = null;
      if (i > 0) {
        s1 = Arrays.copyOfRange(want, 0, 1);
      }
      if (i > 1) {
        s2 = Arrays.copyOfRange(want, 1, i);
      }
      byte[] got = Testpkg.BytesAppend(s1, s2);
      MoreAsserts.assertEquals("Bytes(len="+i+") should match", want, got);
    }
  }

  public void testGoRefGC() {
    Testpkg.S s = Testpkg.New();
    runGC();
    long collected = Testpkg.NumSCollected();
    assertEquals("Only S should be pinned", 0, collected);

    s = null;
    runGC();
    collected = Testpkg.NumSCollected();
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

  /* Suppress this test for now; it's flaky or broken. */
  @Suppress
  public void testJavaRefGC() {
    finalizedAnI = false;
    AnI obj = new AnI();
    runGC();
    Testpkg.Call(obj);
    assertTrue("want F to be called", obj.called);
    obj = null;
    runGC();
    assertTrue("want obj to be collected", finalizedAnI);
  }

  public void testJavaRefKeep() {
    finalizedAnI = false;
    AnI obj = new AnI();
    Testpkg.Keep(obj);
    obj = null;
    runGC();
    assertFalse("want obj to be kept live by Go", finalizedAnI);
  }

  private void runGC() {
    System.gc();
    System.runFinalization();
    Testpkg.GC();
    System.gc();
    System.runFinalization();
  }
}

