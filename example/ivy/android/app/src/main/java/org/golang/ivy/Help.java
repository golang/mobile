/**
 * Copyright 2015 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package org.golang.ivy;


import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.view.MenuItem;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.appcompat.app.AppCompatActivity;

import mobile.Mobile;

/*
 * Displays the help message for Ivy.
 */
public class Help extends AppCompatActivity {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_help);
        WebView webView = (WebView) findViewById(R.id.help_webview);
        webView.setWebViewClient(new WebViewClient() {
            public boolean shouldOverrideUrlLoading(WebView view, String url) {
                // we are not a browser; redirect the request to proper apps.
                if (url != null) {
                    view.getContext().startActivity(new Intent(Intent.ACTION_VIEW, Uri.parse(url)));
                }
                return true;
            }
        });
        webView.getSettings().setDefaultTextEncodingName("utf-8");
        // mobile.Mobile was generated using gomobile bind robpike.io/ivy/mobile.
        String helpMsg = Mobile.help();

        // loadData has a rendering bug: https://code.google.com/p/android/issues/detail?id=6965
        webView.loadDataWithBaseURL("http://pkg.go.dev/robpike.io/ivy", helpMsg, "text/html", "UTF-8", null);
        webView.setBackgroundColor(getResources().getColor(R.color.body));
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        onBackPressed();  // back to parent.
        return true;
    }
}
