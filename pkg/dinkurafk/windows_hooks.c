#include "windows_hooks.h"
#include "_cgo_export.h"

// Global vars
HANDLE threadHandle = NULL;
bool hookedIn = false;
HHOOK hhookKeyboard;
HHOOK hhookMouse;
DWORD lastEventTick;

LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam);
LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam);
DWORD WINAPI ThreadProc(LPVOID lpParameter);

DWORD GetLastEventTick()
{
	return lastEventTick;
}

DWORD GetThreadStatus()
{
	if (threadHandle == NULL)
	{
		return 0;
	}

	DWORD exitCode;
	if (GetExitCodeThread(threadHandle, &exitCode) == 0)
	{
		return GetLastError();
	}

	return exitCode;
}

DWORD RegisterHooks()
{
	if (hookedIn)
	{
		return 1;
	}

	if (threadHandle != NULL)
	{
		return 1;
	}

	threadHandle = CreateThread(NULL, 0, ThreadProc, NULL, 0, NULL);
	DWORD err = GetLastError();
	if (err != 0) {
		return err;
	}

	//HINSTANCE hmod = GetModuleHandle(NULL);
	//DWORD err = GetLastError();
	//if (err != 0) {
	//	return err;
	//}
	//const DWORD dwThreadId = 0; // all threads on the computer

	//hhookKeyboard = SetWindowsHookEx(WH_KEYBOARD_LL, KeyboardProc, hmod, dwThreadId);
	//if (hhookKeyboard == NULL)
	//{
	//	return GetLastError();
	//}

	//hhookMouse = SetWindowsHookEx(WH_MOUSE, MouseProc, hmod, dwThreadId);
	//if (hhookMouse == NULL)
	//{
	//	return GetLastError();
	//}

	hookedIn = true;
	lastEventTick = GetTickCount();
	return 0;
}

DWORD UnregisterHooks()
{
	if (!hookedIn)
	{
		return 1;
	}

	UnhookWindowsHookEx(hhookKeyboard);
	UnhookWindowsHookEx(hhookMouse);
	hookedIn = false;
	return 0;
}

DWORD WINAPI ThreadProc(LPVOID lpParameter)
{
	MSG msg;

	const DWORD dwThreadId = 0; // all threads on the computer

	hhookKeyboard = SetWindowsHookEx(WH_KEYBOARD_LL, KeyboardProc, (HINSTANCE) NULL, dwThreadId);
	if (hhookKeyboard == NULL)
	{
		return GetLastError();
	}

	//hhookMouse = SetWindowsHookEx(WH_MOUSE, MouseProc, (HINSTANCE) NULL, dwThreadId);
	//if (hhookMouse == NULL)
	//{
	//	return GetLastError();
	//}

	while (GetMessage(&msg, NULL, 0, 0))
	{
		TranslateMessage(&msg);
		DispatchMessage(&msg);
	}

	return 0;
}

LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam)
{
	lastEventTick = GetTickCount();
	goWindowsKeyboardEvent();
    return CallNextHookEx(hhookKeyboard, nCode, wParam, lParam);
}

LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam)
{
	lastEventTick = GetTickCount();
    return CallNextHookEx(hhookMouse, nCode, wParam, lParam);
}
