//go:build !cgo && windows

package main

func refuseDirectSetuidLaunch() error {
	return nil
}
