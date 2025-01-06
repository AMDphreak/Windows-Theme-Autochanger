package main

import (
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func isAdmin() bool {
	var sid *windows.SID

	// Try to get the administrator SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// Check if the current process token is a member of the administrator SID
	token := windows.Token(0)
	member := false

	err = windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY|windows.TOKEN_DUPLICATE, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	member, err = token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}

func runMeElevated() error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	// Add a flag to indicate this is the elevated process
	args = args + " --elevated"

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}

	// Exit the non-elevated process
	os.Exit(0)
	return nil
}