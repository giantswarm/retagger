//go:build !linux
// +build !linux

package main

func reexecIfNecessaryForImages(inputImageNames ...string) error {
	return nil
}
