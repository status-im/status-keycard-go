// ======================================================================================
// cgo compilation (for desktop platforms and local tests)
// ======================================================================================

#include <stdio.h>
#include <stddef.h>
#include <stdbool.h>
#include "_cgo_export.h"

typedef void (*callback)(const char *jsonEvent);
callback gCallback = 0;

bool KeycardServiceSignalEvent(const char *jsonEvent) {
	if (gCallback) {
		gCallback(jsonEvent);
	} else {
		NotifyNode((char *)jsonEvent); // re-send notification back to status node
	}

	return true;
}

void KeycardSetEventCallback(void *cb) {
	gCallback = (callback)cb;
}
