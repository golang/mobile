// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#import "IvyController.h"
#import "mobile/Mobile.h"

@interface IvyController ()

@end

@implementation IvyController {
  NSArray *demo_lines;
  int demo_index;
}

- (void)viewDidLoad {
  [super viewDidLoad];

  self.input.delegate = self;
  self.input.autocorrectionType = UITextAutocorrectionTypeNo;
  self.input.keyboardType = UIKeyboardTypeNumbersAndPunctuation;

  self.suggestionView = [[Suggestion alloc] init];
  self.suggestionView.delegate = self;

  self.tape.UIDelegate = self;
  self->demo_lines = NULL;

  [self.okButton setTitle:@"" forState:UIControlStateNormal];
  [self.okButton setHidden:TRUE];

  [[NSNotificationCenter defaultCenter] addObserver:self
                                           selector:@selector(textDidChange:)
                                               name:UITextFieldTextDidChangeNotification
                                             object:self.input];
  [[NSNotificationCenter defaultCenter] addObserver:self
                                           selector:@selector(keyboardWillShow:)
                                               name:UIKeyboardWillShowNotification
                                             object:nil];
  [[NSNotificationCenter defaultCenter] addObserver:self
                                           selector:@selector(keyboardWillHide:)
                                               name:UIKeyboardWillHideNotification
                                             object:nil];

  [self.input becomeFirstResponder];
  [self clear:NULL];
}

- (BOOL)textFieldShouldBeginEditing:(UITextField *)textField {
  if ([textField isEqual:self.input]) {
    textField.inputAccessoryView = self.suggestionView;
    textField.autocorrectionType = UITextAutocorrectionTypeNo;
    [textField reloadInputViews];
  }
  return YES;
}

- (BOOL)textFieldShouldEndEditing:(UITextField *)textField {
  if ([textField isEqual:self.input]) {
    textField.inputAccessoryView = nil;
    [textField reloadInputViews];
  }
  return YES;
}

- (void)textDidChange:(NSNotification *)notif {
  [self.suggestionView suggestFor:self.input.text];
}

- (void)suggestionReplace:(NSString *)text {
  self.input.text = text;
  [self.suggestionView suggestFor:text];
}

- (void)keyboardWillShow:(NSNotification *)aNotification {
  // Move the input text field up, as the keyboard has taken some of the screen.
  NSDictionary *info = [aNotification userInfo];
  CGRect kbFrame = [[info objectForKey:UIKeyboardFrameEndUserInfoKey] CGRectValue];
  NSNumber *duration = [info objectForKey:UIKeyboardAnimationDurationUserInfoKey];

  UIViewAnimationCurve keyboardTransitionAnimationCurve;
  [[info valueForKey:UIKeyboardAnimationCurveUserInfoKey]
      getValue:&keyboardTransitionAnimationCurve];
  UIViewAnimationOptions options =
      keyboardTransitionAnimationCurve | keyboardTransitionAnimationCurve << 16;

  [UIView animateWithDuration:duration.floatValue
      delay:0
      options:options
      animations:^{
        self.bottomConstraint.constant = 0 - kbFrame.size.height;
        [self.view layoutIfNeeded];
      }
      completion:^(BOOL finished) {
        [self scrollTapeToBottom];
      }];
}

- (void)keyboardWillHide:(NSNotification *)aNotification {
  // Move the input text field back down.
  NSDictionary *info = [aNotification userInfo];

  NSNumber *duration = [info objectForKey:UIKeyboardAnimationDurationUserInfoKey];

  UIViewAnimationCurve keyboardTransitionAnimationCurve;
  [[info valueForKey:UIKeyboardAnimationCurveUserInfoKey]
      getValue:&keyboardTransitionAnimationCurve];
  UIViewAnimationOptions options =
      keyboardTransitionAnimationCurve | keyboardTransitionAnimationCurve << 16;

  int offset = self.input.inputAccessoryView != NULL ? self.suggestionView.frame.size.height : 0;

  [UIView animateWithDuration:duration.floatValue
      delay:0
      options:options
      animations:^{
        self.bottomConstraint.constant = 0 - offset;
        [self.view layoutIfNeeded];
      }
      completion:^(BOOL finished) {
        [self scrollTapeToBottom];
      }];
}

- (void)enterPressed {
  NSString *text = self.input.text;
  if ([text isEqual:@""]) {
    if (self->demo_lines == NULL) {
      return;
    }
    while (demo_index < self->demo_lines.count) {
      NSString *line = self->demo_lines[self->demo_index++];
      if ([line hasPrefix:@"#"]) {
        [self appendTape:line tag:@"comment"];
      } else {
        self.input.text = line;
        break;
      }
    }
  } else if (self->demo_lines != NULL && [text isEqual:@"quit"]) {
    [self unloadDemo];
  } else {
    [self appendTape:text tag:@"expr"];
    NSString *expr = [text stringByAppendingString:@"\n"];
    NSError *err;
    NSString *result = MobileEval(expr, &err);
    if (err != nil) {
      result = err.description;
    }
    result = [result stringByTrimmingCharactersInSet:[NSCharacterSet newlineCharacterSet]];
    result = [result stringByReplacingOccurrencesOfString:@"<" withString:@"&lt;"];
    result = [result stringByReplacingOccurrencesOfString:@">" withString:@"&gt;"];
    NSMutableArray *lines = (NSMutableArray *)[result componentsSeparatedByString:@"\n"];
    for (NSMutableString *line in lines) {
      if ([line hasPrefix:@"#"])
        [self appendTape:line tag:@"comment"];
      else
        [self appendTape:line tag:@"result"];
    }
    self.input.text = @"";
  }
  [self.input becomeFirstResponder];
}

- (void)scrollTapeToBottom {
  NSString *scroll = @"window.scrollBy(0, document.body.offsetHeight);";
  [self.tape evaluateJavaScript:scroll completionHandler:nil];
}

- (void)appendTape:(NSString *)text tag:(NSString *)tag {
  NSString *injectSrc = @"appendDiv('%@','%@');";
  NSString *runToInject = [NSString stringWithFormat:injectSrc, text, tag];
  [self.tape evaluateJavaScript:runToInject completionHandler:nil];
  [self scrollTapeToBottom];
}

- (void)loadDemo {
  [self.okButton setHidden:FALSE];
  NSString *text = DemoText();

  self->demo_lines =
      [text componentsSeparatedByCharactersInSet:[NSCharacterSet newlineCharacterSet]];
  self->demo_index = 0;
  self.input.text = @"";
  self.input.enablesReturnKeyAutomatically = TRUE;
  [self enterPressed];
}
- (void)unloadDemo {
  [self.okButton setHidden:TRUE];
  self.input.enablesReturnKeyAutomatically = FALSE;
  self->demo_lines = NULL;
  self.input.text = @"";
}
- (IBAction)okPressed:(id)sender {
  [self enterPressed];
}

- (IBAction)demo:(id)sender {
  if (self->demo_lines) {  // demo already running
    [self enterPressed];
  } else {
    [self loadDemo];
  }
}

- (IBAction)clear:(id)sender {
  [self unloadDemo];
  NSString *string = [NSString
      stringWithContentsOfFile:[[NSBundle mainBundle] pathForResource:@"tape" ofType:@"html"]
                      encoding:NSUTF8StringEncoding
                         error:NULL];
  [self.tape loadHTMLString:string baseURL:NULL];
}
@end

