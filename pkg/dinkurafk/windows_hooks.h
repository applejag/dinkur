#include <windows.h>
#include <stdbool.h>

#pragma comment(lib, "user32.lib")

DWORD RegisterHooks();
DWORD UnregisterHooks();
DWORD GetLastEventTick();
DWORD GetThreadStatus();
