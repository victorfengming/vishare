//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// vishare_platform_setup must be called on the OS main thread before
// systray.Run().  It ensures NSApp is initialised and the activation
// policy is set to Accessory (menu-bar-only, no Dock icon) so that
// NSStatusItem menus receive mouse events in non-bundled binaries.
void vishare_platform_setup(void) {
    [NSApplication sharedApplication];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    [NSApp activateIgnoringOtherApps:YES];
}
*/
import "C"

import "runtime"

func platformSetup() {
    // Cocoa UI calls must happen on the OS main thread.
    runtime.LockOSThread()
    C.vishare_platform_setup()
}
