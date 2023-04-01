#include <Windows.h>

BOOL APIENTRY DllMain(
    HINSTANCE hinstDLL,  // handle to DLL module
    DWORD fdwReason,     // reason for calling function
    LPVOID lpvReserved )  // reserved
{
    // Perform actions based on the reason for calling.
    switch( fdwReason ) 
    { 
        case DLL_PROCESS_ATTACH:

            //STARTUPINFOA info = { sizeof(info) };
            //PROCESS_INFORMATION processInfo;
            //if (CreateProcessA((LPCSTR)"main.exe", NULL, NULL, NULL, TRUE, 0, NULL, NULL, &info, &processInfo))
            //{
            //    WaitForSingleObject(processInfo.hProcess, INFINITE);
            //    CloseHandle(processInfo.hProcess);
            //    CloseHandle(processInfo.hThread);
            //}

            // Begin injection of remote PE into memory
            break;
    }
    return TRUE;  // Successful DLL_PROCESS_ATTACH.
}

// These functions enable the sideloading
extern "C" __declspec (dllexport) void DevObjCreateDeviceInfoList(){}

extern "C" __declspec (dllexport) void DevObjUninstallDevice(){}

extern "C" __declspec (dllexport) void DevObjOpenDevRegKey(){}

extern "C" __declspec (dllexport) void DevObjEnumDeviceInfo(){}

extern "C" __declspec (dllexport) void DevObjGetClassDevs() {}

extern "C" __declspec (dllexport) void DevObjGetDeviceInstanceId() {}

extern "C" __declspec (dllexport) void DevObjDestroyDeviceInfoList() {}