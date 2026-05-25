// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#import <UIKit/UIKit.h>
#import <WebKit/WebKit.h>
#import "Suggestion.h"

// IvyController displays the main app view.
@interface IvyController : UIViewController <UITextFieldDelegate, WKUIDelegate, SuggestionDelegate>

@property(weak, nonatomic) IBOutlet NSLayoutConstraint *bottomConstraint;

// A text input field coupled to an output "tape", rendered with a WKWebView.
@property(weak, nonatomic) IBOutlet UITextField *input;
@property(strong, nonatomic) Suggestion *suggestionView;
@property(weak, nonatomic) IBOutlet WKWebView *tape;
@property(weak, nonatomic) IBOutlet UIButton *okButton;

- (IBAction)clear:(id)sender;
- (IBAction)demo:(id)sender;
- (IBAction)okPressed:(id)sender;

@end

