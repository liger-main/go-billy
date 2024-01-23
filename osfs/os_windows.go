//go:build windows
// +build windows

package osfs

import (
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32DLL    = windows.NewLazySystemDLL("kernel32.dll")
	lockFileExProc = kernel32DLL.NewProc("LockFileEx")
	unlockFileProc = kernel32DLL.NewProc("UnlockFile")
)

const (
	lockfileExclusiveLock = 0x2
)

func (f *file) Lock() error {
	f.m.Lock()
	defer f.m.Unlock()

	var overlapped windows.Overlapped
	// err is always non-nil as per sys/windows semantics.
	ret, _, err := lockFileExProc.Call(f.File.Fd(), lockfileExclusiveLock, 0, 0xFFFFFFFF, 0,
		uintptr(unsafe.Pointer(&overlapped)))
	runtime.KeepAlive(&overlapped)
	if ret == 0 {
		return err
	}
	return nil
}

func (f *file) Unlock() error {
	f.m.Lock()
	defer f.m.Unlock()

	// err is always non-nil as per sys/windows semantics.
	ret, _, err := unlockFileProc.Call(f.File.Fd(), 0, 0, 0xFFFFFFFF, 0)
	if ret == 0 {
		return err
	}
	return nil
}

func rename(from, to string) error {
	return os.Rename(from, to)
}

func umask(new int) func() {
	return func() {
	}
}

func objectID(path string, fi os.FileInfo) uint64 {
	pathp, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	handle, err := windows.CreateFile(
		pathp,
		windows.GENERIC_READ,
		0,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)

	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(handle)

	var data windows.ByHandleFileInformation

	if err = windows.GetFileInformationByHandle(handle, &data); err != nil {
		return 0, err
	}

	return (uint64(data.FileIndexHigh) << 32) | uint64(data.FileIndexLow), nil
}
