// ======================================================================================
// cgo compilation (for desktop platforms and local tests)
// ======================================================================================

#include <stdio.h>
#include <stddef.h>
#include <stdbool.h>
#include "_cgo_export.h"

typedef void (*keycardCallback)(const char *jsonEvent);
keycardCallback gCallback = 0;

bool KeycardServiceSignalEvent(const char *jsonEvent) {
	if (gCallback) {
		gCallback(jsonEvent);
	}

	return true;
}

void KeycardSetEventCallback(void *cb) {
	gCallback = (keycardCallback)cb;
}
