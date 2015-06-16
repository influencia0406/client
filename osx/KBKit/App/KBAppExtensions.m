//
//  KBAppExtensions.m
//  Keybase
//
//  Created by Gabriel on 6/10/15.
//  Copyright (c) 2015 Keybase. All rights reserved.
//

#import "KBAppExtensions.h"

#import "KBService.h"
#import "KBPGPEncryptActionView.h"
#import "KBWorkspace.h"
#import "KBPGPEncrypt.h"
#import "KBWork.h"

@interface KBAppExtensions ()
@property KBService *service;
@end

@implementation KBAppExtensions

- (KBService *)loadService {
  if (!_service) {
    KBEnvConfig *config = [KBEnvConfig loadFromUserDefaults:[KBWorkspace userDefaults]];
    _service = [[KBService alloc] initWithConfig:config];
  }
  return _service;
}

- (NSView *)encryptViewWithExtensionItem:(NSExtensionItem *)extensionItem completion:(KBOnExtension)completion {
  KBPGPEncryptActionView *encryptView = [[KBPGPEncryptActionView alloc] initWithFrame:CGRectMake(0, 0, 400, 300)];
  encryptView.extensionItem = extensionItem;
  encryptView.client = [self loadService].client;
  encryptView.completion = completion;
  return encryptView;
}

- (void)encryptExtensionItem:(NSExtensionItem *)extensionItem usernames:(NSArray *)usernames sender:(id)sender completion:(KBOnExtension)completion {
  KBPGPEncrypt *encrypt = [[KBPGPEncrypt alloc] init];
  KBRPClient *client = [self loadService].client;
  NSString *text = extensionItem.attributedContentText.string;
  [encrypt encryptText:text usernames:usernames client:client sender:sender completion:^(KBWork *work) {
    //NSError *error = [work error];
    // TODO: Handle error
    KBStream *stream = [work output];
    NSExtensionItem *item = [[NSExtensionItem alloc] init];
    NSAttributedString *text = [[NSAttributedString alloc] initWithString:[[NSString alloc] initWithData:stream.writer.data encoding:NSUTF8StringEncoding]];
    item.attributedContentText = text;
    completion(sender, item);
  }];
}

@end
