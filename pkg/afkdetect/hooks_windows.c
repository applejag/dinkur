// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

#include "hooks_windows.h"
#include "_cgo_export.h"

// Global vars
HANDLE threadHandle = NULL;
bool hookedIn = false;
HHOOK hhookKeyboard;
HHOOK hhookMouse;
DWORD lastEventTick;
DWORD lastManualTick;
const DWORD throttleMs = 5000;

LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam);
LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam);
DWORD WINAPI ThreadProc(LPVOID lpParameter);

DWORD GetLastEventTickMs()
{
	return lastEventTick;
}

DWORD GetTickMs()
{
	return GetTickCount();
}

bool GetWorkstationLocked()
{
	HDESK desktop = OpenInputDesktop(0, false, DESKTOP_READOBJECTS);
	DWORD err = GetLastError();
	if (desktop != NULL)
	{
		CloseDesktop(desktop);
	}
	return err == ERROR_ACCESS_DENIED;
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

	hookedIn = true;
	lastEventTick = GetTickCount();
	lastManualTick = lastEventTick;
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

	hhookMouse = SetWindowsHookEx(WH_MOUSE_LL, MouseProc, (HINSTANCE) NULL, dwThreadId);
	if (hhookMouse == NULL)
	{
		return GetLastError();
	}

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
	if (lastEventTick - lastManualTick >= throttleMs) {
		lastManualTick = lastEventTick;
		goTriggerTick();
	}
    return CallNextHookEx(hhookKeyboard, nCode, wParam, lParam);
}

LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam)
{
	// TODO: this value wraps after 49.7 days of computer uptime
	lastEventTick = GetTickCount();
	if (lastEventTick - lastManualTick >= throttleMs) {
		lastManualTick = lastEventTick;
		goTriggerTick();
	}
    return CallNextHookEx(hhookMouse, nCode, wParam, lParam);
}
