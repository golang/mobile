// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#import "DocsController.h"
#import "mobile/Mobile.h"

@interface DocsController () {
  WKWebView *webView;
}

@end

@implementation DocsController

- (void)viewDidLoad {
  [super viewDidLoad];
  webView = (WKWebView *)[self.view viewWithTag:11];
  NSString *helpHTML = MobileHelp();
  NSRange r = [helpHTML rangeOfString:@"<head>"];
  NSString *html = [helpHTML substringToIndex:r.location];

  // With the following meta tag, WKWebView displays the fonts more nicely.
  NSString *meta = @"<meta name='viewport' \
        content='width=device-width, "
                   @"initial-scale=1.0, maximum-scale=1.0, \
        minimum-scale=1.0, "
                   @"user-scalable=no'>";

  html = [html stringByAppendingString:@"<head>"];
  html = [html stringByAppendingString:meta];
  html = [html stringByAppendingString:[helpHTML substringFromIndex:r.location]];

  [webView loadHTMLString:html baseURL:NULL];
}

- (void)didReceiveMemoryWarning {
  [super didReceiveMemoryWarning];
}

@end

