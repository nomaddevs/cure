/*
	Compile with:
		go build -o cure.so -buildmode=c-shared cure.go
*/
package main

import "C"

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// TH32CS_SNAPPROCESS is described in https://msdn.microsoft.com/de-de/library/windows/desktop/ms682489(v=vs.85).aspx
const TH32CS_SNAPPROCESS		= 0x00000002 // Includes all processes in the system in the snapshot.
const PROCESS_ALL_ACCESS 		= 0x1F0FFF
const START_LOOP_BASE_ADDRESS 	= 0x400000
const END_LOOP_BASE_ADDRESS 	= 0x41DFFF//0xFFFFFF
const FFXIV_EXE_NAME_DX11 		= "ffxiv_dx11.exe"
const FFXIV_EXE_NAME_DX9 		= "ffxiv_dx9.exe"

func OpenProcessHandle(processId int) uintptr {
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	proc := kernel32.MustFindProc("OpenProcess")
	handle, _, _ := proc.Call(ptr(PROCESS_ALL_ACCESS), ptr(true), ptr(processId))
	return uintptr(handle)
}

func ptr(val interface{}) uintptr {
	switch val.(type) {
	case string:
		return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(val.(string))))
	case int:
		return uintptr(val.(int))
	default:
		return uintptr(0)
	}
}

//export GetProc
func GetProc() int {
	procs, err := processes()
	if err != nil {
		log.Fatal(err)
	}
	ffxiv := findProcessByName(procs, FFXIV_EXE_NAME_DX11)
	if ffxiv != nil {
		return ffxiv.ProcessID
	}
	ffxiv = findProcessByName(procs, FFXIV_EXE_NAME_DX9)
	if ffxiv != nil {
		return ffxiv.ProcessID
	}
	return -1
}

// WindowsProcess is an implementation of Process for Windows.
type WindowsProcess struct {
	ProcessID       int
	ParentProcessID int
	Exe             string
}

func processes() ([]WindowsProcess, error) {
	handle, err := windows.CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(handle)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	// get the first process
	err = windows.Process32First(handle, &entry)
	if err != nil {
		return nil, err
	}

	results := make([]WindowsProcess, 0, 50)
	for {
		results = append(results, newWindowsProcess(&entry))

		err = windows.Process32Next(handle, &entry)
		if err != nil {
			// windows sends ERROR_NO_MORE_FILES on last process
			if err == syscall.ERROR_NO_MORE_FILES {
				return results, nil
			}
			return nil, err
		}
	}
}

func findProcessByName(processes []WindowsProcess, name string) *WindowsProcess {
	for _, p := range processes {
		if strings.ToLower(p.Exe) == strings.ToLower(name) {
			return &p
		}
	}
	return nil
}

func newWindowsProcess(e *windows.ProcessEntry32) WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return WindowsProcess{
		ProcessID:       int(e.ProcessID),
		ParentProcessID: int(e.ParentProcessID),
		Exe:             syscall.UTF16ToString(e.ExeFile[:end]),
	}
}

func main(){
	fmt.Println("Getting FFXIV PID...")
	pid := GetProc()
	if pid < 0 {
		fmt.Printf("Couldn't get FFXIV PID.")
		return
	}

	fmt.Printf("Found PID %d\n", pid)
	handle := OpenProcessHandle(pid)
	fmt.Printf("Handle: %d\n", handle)
	fmt.Printf("Reading process memory...\n")

	//buffer, numBytesRead, ok := w32.ReadProcessMemory(handle, lp, unsafe.Sizeof(poa))
	procReadProcessMemory := windows.MustLoadDLL("kernel32.dll").MustFindProc("ReadProcessMemory")
	if nil == procReadProcessMemory {
		fmt.Printf("Nil procReadProcessMemory")
		return
	}

	var data [2]byte
    var length uint32
    var first = -1
    var last = -1

	for i := START_LOOP_BASE_ADDRESS; i < END_LOOP_BASE_ADDRESS; i += 2 {
		ret1, _, err := procReadProcessMemory.Call(uintptr(handle), uintptr(i), uintptr(unsafe.Pointer(&data[0])), 2, uintptr(unsafe.Pointer(&length)))
		if err != syscall.Errno(0) {
			fmt.Printf("First hit: 0x%X Last hit: 0x%X Current: 0x%X Last Error: %s\r", first, last, i, err.Error())
		}
		if ret1 != 0 {
			// dubug stuff
			if first == -1 {
				first = i
			}
			if last == -1 || last < i {
				last = i
			}
			// /debug 
			fmt.Printf("0x%X Length: %v\n%v\n", i, length, data)
		}
	}

	fmt.Printf("First address: 0x%X\nLast address: 0x%X\n", first, last)
}