//go:build !linux

package main

import "syscall"

var sigInfo = syscall.SIGINFO
