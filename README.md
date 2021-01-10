# godump
Process dumping utility in golang

## Background Info
Inspired by SharpDump(https://github.com/GhostPack/SharpDump) and several other process memory dumping utilities, acts as a replacement for procdump.exe
- Uses golang's /x/sys/windows library to import and call MiniDumpWriteDump on a process
- Currently does not enable seDebugPriv on current process, so you should run from elevated context
- No *nix support (yet)

## Usage
### Build
```
> go build godump.go
```
### Arguments
- pid: Process Id to dump memory of
- verbose: boolean option to print status updates on program
### Example
```
> godump.exe -pid 1234 -verbose
```
