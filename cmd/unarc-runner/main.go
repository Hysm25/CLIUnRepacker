//go:build windows && 386

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	virtualAlloc                = kernel32.NewProc("VirtualAlloc")
	addVectoredExceptionHandler = kernel32.NewProc("AddVectoredExceptionHandler")
	terminateThread             = kernel32.NewProc("TerminateThread")
	getCurrentThread            = kernel32.NewProc("GetCurrentThread")
)

// installWoW64ExceptionHandler intercepts the access violation caused by Wine's WoW64
// thread exit thunk trying to return to 0xff...... addresses.
func installWoW64ExceptionHandler() {
	const MEM_COMMIT = 0x1000
	const MEM_RESERVE = 0x2000
	const PAGE_EXECUTE_READWRITE = 0x40

	mem, _, _ := virtualAlloc.Call(0, 128, MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWRITE)
	if mem == 0 {
		panic("VirtualAlloc failed for exception handler")
	}

	gctAddr := getCurrentThread.Addr()
	ttAddr := terminateThread.Addr()

	code := []byte{
		0x55,                               // 00: push ebp
		0x89, 0xe5,                         // 01: mov ebp, esp
		0x8b, 0x45, 0x08,                   // 03: mov eax, [ebp+8]
		0x8b, 0x00,                         // 06: mov eax, [eax]
		0x81, 0x38, 0x05, 0x00, 0x00, 0xc0, // 08: cmp dword ptr [eax], 0xc0000005
		0x75, 0x27,                         // 0e: jne pass
		0x8b, 0x40, 0x0c,                   // 10: mov eax, [eax+12]
		0xc1, 0xe8, 0x18,                   // 13: shr eax, 24
		0x3d, 0xff, 0x00, 0x00, 0x00,       // 16: cmp eax, 0xff
		0x75, 0x1a,                         // 1b: jne pass
		0x6a, 0x00,                         // 1d: push 0
		0xb8, 0x00, 0x00, 0x00, 0x00,       // 1f: mov eax, gctAddr
		0xff, 0xd0,                         // 24: call eax
		0x50,                               // 26: push eax
		0xb8, 0x00, 0x00, 0x00, 0x00,       // 27: mov eax, ttAddr
		0xff, 0xd0,                         // 2c: call eax
		0xb8, 0xff, 0xff, 0xff, 0xff,       // 2e: mov eax, -1 (EXCEPTION_CONTINUE_EXECUTION)
		0x5d,                               // 33: pop ebp
		0xc2, 0x04, 0x00,                   // 34: ret 4
		// pass:
		0x31, 0xc0,                         // 37: xor eax, eax (EXCEPTION_CONTINUE_SEARCH)
		0x5d,                               // 39: pop ebp
		0xc2, 0x04, 0x00,                   // 3a: ret 4
	}

	*(*uint32)(unsafe.Pointer(&code[0x20])) = uint32(gctAddr)
	*(*uint32)(unsafe.Pointer(&code[0x28])) = uint32(ttAddr)

	for i, b := range code {
		*(*byte)(unsafe.Pointer(mem + uintptr(i))) = b
	}

	addVectoredExceptionHandler.Call(1, mem)
}

// makeCdeclThunk creates a machine-code bridge that dynamically pushes N arguments
// and calls a __cdecl function where the caller is responsible for cleaning up the stack.
func makeCdeclThunk() func(proc uintptr, n int, args *uintptr) uintptr {
	const MEM_COMMIT = 0x1000
	const MEM_RESERVE = 0x2000
	const PAGE_EXECUTE_READWRITE = 0x40

	mem, _, _ := virtualAlloc.Call(0, 128, MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWRITE)
	if mem == 0 {
		panic("VirtualAlloc failed for cdecl thunk")
	}

	code := []byte{
		0x55,                   // push ebp
		0x89, 0xe5,             // mov ebp, esp
		0x53,                   // push ebx
		0x56,                   // push esi
		0x57,                   // push edi
		0x8b, 0x4d, 0x0c,       // mov ecx, [ebp+12] (n)
		0x8b, 0x75, 0x10,       // mov esi, [ebp+16] (args array)
		0x85, 0xc9,             // test ecx, ecx
		0x74, 0x0f,             // jz call_it (+15)
		0x8d, 0x74, 0x8e, 0xfc, // lea esi, [esi+ecx*4-4]
		// loop_push:
		0xff, 0x36,       // push dword ptr [esi]
		0x83, 0xee, 0x04, // sub esi, 4
		0x49,             // dec ecx
		0x75, 0xf7,       // jnz loop_push
		// call_it:
		0xff, 0x55, 0x08, // call dword ptr [ebp+8] (proc)
		// Epilogue:
		0x8d, 0x65, 0xf4, // lea esp, [ebp-12]
		0x5f,             // pop edi
		0x5e,             // pop esi
		0x5b,             // pop ebx
		0x89, 0xec,       // mov esp, ebp
		0x5d,             // pop ebp
		0xc2, 0x0c, 0x00, // ret 12
	}

	for i, b := range code {
		*(*byte)(unsafe.Pointer(mem + uintptr(i))) = b
	}

	return func(proc uintptr, n int, args *uintptr) uintptr {
		r, _, _ := syscall.Syscall(mem, 3, proc, uintptr(n), uintptr(unsafe.Pointer(args)))
		return r
	}
}

func cString(p uintptr) string {
	if p == 0 {
		return ""
	}
	var b []byte
	for {
		c := *(*byte)(unsafe.Pointer(p))
		if c == 0 {
			break
		}
		b = append(b, c)
		p++
	}
	return string(b)
}

func goProgress(what uintptr, int1 uintptr, int2 uintptr, str uintptr) uintptr {
	wStr := cString(what)
	sStr := cString(str)
	if wStr == "overwrite?" {
		return 'y'
	}
	if wStr == "write" {
		// Suppress high-frequency write spam unless needed
		return 1
	}
	fmt.Printf("[%s] %s (%d, %d)\n", wStr, sStr, int1, int2)
	return 1
}

func main() {
	var workDir string
	var arcPath string
	var outDir string
	var mode string

	flag.StringVar(&workDir, "work", "", "Working directory containing unarc.dll and arc.ini")
	flag.StringVar(&arcPath, "arc", "", "Path to the archive (.bin)")
	flag.StringVar(&outDir, "out", "", "Output directory for extracted files")
	flag.StringVar(&mode, "mode", "x", "FreeArc mode (x=extract, l=list)")
	flag.Parse()

	if workDir == "" || arcPath == "" || outDir == "" {
		if flag.NArg() >= 3 {
			workDir = flag.Arg(0)
			arcPath = flag.Arg(1)
			outDir = flag.Arg(2)
		} else {
			fmt.Println("Usage: unarc-runner -work <workdir> -arc <archive.bin> -out <outdir> [-mode x|l]")
			os.Exit(1)
		}
	}

	absWorkDir, err := filepath.Abs(workDir)
	if err == nil {
		workDir = absWorkDir
	}
	absArcPath, err := filepath.Abs(arcPath)
	if err == nil {
		arcPath = absArcPath
	}
	absOutDir, err := filepath.Abs(outDir)
	if err == nil {
		outDir = absOutDir
	}

	_ = os.Chdir(workDir)
	setCurrentDirectory := kernel32.NewProc("SetCurrentDirectoryW")
	if setCurrentDirectory.Find() == nil {
		wDir, _ := syscall.UTF16PtrFromString(workDir)
		setCurrentDirectory.Call(uintptr(unsafe.Pointer(wDir)))
	}

	installWoW64ExceptionHandler()

	dllPath := filepath.Join(workDir, "unarc.dll")
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		fmt.Printf("[!] Failed to load unarc.dll from %s: %v\n", dllPath, err)
		os.Exit(1)
	}
	defer dll.Release()

	proc, err := dll.FindProc("FreeArcExtract")
	if err != nil {
		fmt.Printf("[!] Failed to find FreeArcExtract in unarc.dll: %v\n", err)
		os.Exit(1)
	}

	callCdecl := makeCdeclThunk()
	cb := syscall.NewCallback(goProgress)

	cfgPath := filepath.Join(workDir, "arc.ini")
	cmd1, _ := syscall.BytePtrFromString(mode)
	cmd2, _ := syscall.BytePtrFromString("-o+")
	cmd3, _ := syscall.BytePtrFromString("-cfg" + cfgPath)
	cmd4, _ := syscall.BytePtrFromString("-w" + workDir)
	cmd5, _ := syscall.BytePtrFromString("-dp" + outDir)
	cmd6, _ := syscall.BytePtrFromString("--")
	cmd7, _ := syscall.BytePtrFromString(arcPath)

	args := []uintptr{
		cb,
		uintptr(unsafe.Pointer(cmd1)),
		uintptr(unsafe.Pointer(cmd2)),
		uintptr(unsafe.Pointer(cmd3)),
		uintptr(unsafe.Pointer(cmd4)),
		uintptr(unsafe.Pointer(cmd5)),
		uintptr(unsafe.Pointer(cmd6)),
		uintptr(unsafe.Pointer(cmd7)),
		0,
	}

	r := callCdecl(proc.Addr(), len(args), &args[0])
	exitCode := int(r)
	if exitCode != 0 {
		fmt.Printf("[!] FreeArcExtract returned error code: %d\n", exitCode)
		os.Exit(exitCode)
	}
	fmt.Println("[✓] FreeArcExtract completed successfully.")
}
