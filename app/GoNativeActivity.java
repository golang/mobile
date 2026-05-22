package org.golang.app;

import android.app.Activity;
import android.app.NativeActivity;
import android.content.Context;
import android.content.pm.ActivityInfo;
import android.content.pm.PackageManager;
import android.os.Build;
import android.os.Bundle;
import android.util.Log;
import android.view.KeyCharacterMap;
import android.view.View;
import android.view.WindowManager;
import android.view.inputmethod.InputMethodManager;

public class GoNativeActivity extends NativeActivity {
	private static GoNativeActivity goNativeActivity;

	public GoNativeActivity() {
		super();
		goNativeActivity = this;
	}

	String getTmpdir() {
		return getCacheDir().getAbsolutePath();
	}

	static int getRune(int deviceId, int keyCode, int metaState) {
		try {
			int rune = KeyCharacterMap.load(deviceId).get(keyCode, metaState);
			if (rune == 0) {
				return -1;
			}
			return rune;
		} catch (KeyCharacterMap.UnavailableException e) {
			return -1;
		} catch (Exception e) {
			Log.e("Go", "exception reading KeyCharacterMap", e);
			return -1;
		}
	}

	private void load() {
		// Interestingly, NativeActivity uses a different method
		// to find native code to execute, avoiding
		// System.loadLibrary. The result is Java methods
		// implemented in C with JNIEXPORT (and JNI_OnLoad) are not
		// available unless an explicit call to System.loadLibrary
		// is done. So we do it here, borrowing the name of the
		// library from the same AndroidManifest.xml metadata used
		// by NativeActivity.
		try {
			ActivityInfo ai = getPackageManager().getActivityInfo(
					getIntent().getComponent(), PackageManager.GET_META_DATA);
			if (ai.metaData == null) {
				Log.e("Go", "loadLibrary: no manifest metadata found");
				return;
			}
			String libName = ai.metaData.getString("android.app.lib_name");
			System.loadLibrary(libName);
		} catch (Exception e) {
			Log.e("Go", "loadLibrary failed", e);
		}
	}

	@Override
	public void onCreate(Bundle savedInstanceState) {
		load();
		super.onCreate(savedInstanceState);

		// Force the soft keyboard hidden on every window focus. Without this
		// the IME can show on its own (some devices auto-show on launch even
		// though the GLSurfaceView has no editable text field).
		// showSoftKeyboard() forces the IME up via toggleSoftInput(SHOW_FORCED),
		// which overrides this flag when text input is genuinely needed.
		getWindow().setSoftInputMode(
				WindowManager.LayoutParams.SOFT_INPUT_STATE_ALWAYS_HIDDEN
						| WindowManager.LayoutParams.SOFT_INPUT_ADJUST_NOTHING);

		// Enable rendering into display cutout area (notch) for API 28+
		if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
			WindowManager.LayoutParams params = getWindow().getAttributes();
			params.layoutInDisplayCutoutMode = WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES;
			getWindow().setAttributes(params);
			Log.i("Go", "Display cutout mode set to SHORT_EDGES");
		}
	}

	@Override
	public void onWindowFocusChanged(boolean hasFocus) {
		super.onWindowFocusChanged(hasFocus);
		if (hasFocus) {
			// Belt-and-suspenders: explicitly dismiss the IME on every focus
			// gain. STATE_ALWAYS_HIDDEN alone isn't enough on some devices
			// where the IME has been brought up by a prior activity instance.
			try {
				InputMethodManager imm = (InputMethodManager) getSystemService(Context.INPUT_METHOD_SERVICE);
				if (imm != null) {
					View view = getWindow().getDecorView();
					if (view != null) {
						imm.hideSoftInputFromWindow(view.getWindowToken(), 0);
					}
				}
			} catch (Exception e) {
				Log.e("Go", "onWindowFocusChanged hide IME failed", e);
			}
		}
	}

	@Override
	protected void onResume() {
		super.onResume();
		// Also hide on every resume — covers the launch path and returning from
		// background.
		try {
			InputMethodManager imm = (InputMethodManager) getSystemService(Context.INPUT_METHOD_SERVICE);
			if (imm != null) {
				View view = getWindow().getDecorView();
				if (view != null) {
					imm.hideSoftInputFromWindow(view.getWindowToken(), 0);
				}
			}
		} catch (Exception e) {
			Log.e("Go", "onResume hide IME failed", e);
		}
	}

	// Show the soft keyboard
	void showSoftKeyboard() {
		runOnUiThread(new Runnable() {
			@Override
			public void run() {
				try {
					InputMethodManager imm = (InputMethodManager) getSystemService(Context.INPUT_METHOD_SERVICE);
					if (imm != null) {
						// For NativeActivity, we need to use toggleSoftInput with SHOW_FORCED
						// First, try to ensure the window can receive input
						View decorView = getWindow().getDecorView();
						decorView.setFocusable(true);
						decorView.setFocusableInTouchMode(true);
						decorView.requestFocus();

						// Use toggleSoftInput which works more reliably with NativeActivity
						imm.toggleSoftInput(InputMethodManager.SHOW_FORCED, InputMethodManager.HIDE_IMPLICIT_ONLY);
						Log.i("Go", "showSoftKeyboard: toggleSoftInput called");
					}
				} catch (Exception e) {
					Log.e("Go", "showSoftKeyboard failed", e);
				}
			}
		});
	}

	// Hide the soft keyboard
	void hideSoftKeyboard() {
		runOnUiThread(new Runnable() {
			@Override
			public void run() {
				try {
					InputMethodManager imm = (InputMethodManager) getSystemService(Context.INPUT_METHOD_SERVICE);
					if (imm != null) {
						View view = getWindow().getDecorView();
						if (view != null) {
							// Use hideSoftInputFromWindow only. Do NOT call
							// toggleSoftInput here — toggle SHOWS the IME when
							// it is already hidden, which is exactly the
							// opposite of what we want.
							imm.hideSoftInputFromWindow(view.getWindowToken(), 0);
						}
						Log.i("Go", "hideSoftKeyboard: called");
					}
				} catch (Exception e) {
					Log.e("Go", "hideSoftKeyboard failed", e);
				}
			}
		});
	}
}
