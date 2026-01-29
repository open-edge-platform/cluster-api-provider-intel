// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"syscall"

	"github.com/naughtygopher/errors"
)

type Filelink string

const (
	Symbollink Filelink = "symbol"
	Hardlink   Filelink = "hard"
	Normalfile Filelink = "normal"
	Wrongfile  Filelink = ""
)

func RemoveFile(name string) error {
	err := os.Remove(name)
	if err != nil && !os.IsNotExist(err) {
		return errors.InternalErrf(err, "Failed to remove %s", name)
	}
	return nil
}

func CheckFileLink(filename string) Filelink {
	fi, err := os.Lstat(filename)
	if err != nil {
		log.Error().Msgf("Check link failure: %v", err)
		return Wrongfile
	}
	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		log.Error().Msgf("Check stat value failure: %v", err)
		return Wrongfile
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		_, err = os.Readlink(filename)
		if err != nil {
			log.Error().Msgf("Read symbol link error : %v", err)
			return Wrongfile
		}
		return Symbollink
	}

	nlink := uint32(s.Nlink) //nolint:gosec // unknown why this cast exists, leaving as is
	if nlink > 1 {
		return Hardlink
	} else {
		return Normalfile
	}
}

func IsValidFile(filename string) bool {
	flinktype := CheckFileLink(filename)
	if flinktype == Normalfile || flinktype == Symbollink {
		return true
	}
	return false
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		log.Error().Msgf("error opening source file: %v", err)
		return 0, errors.InternalErrf(err, "error opening source file: %v", srcName)
	}
	defer src.Close()

	validfile := IsValidFile(srcName)

	if !validfile {
		return 0, errors.Validationf("not a valid src file: %v", srcName)
	}
	if FileExists(dstName) {
		validfile = IsValidFile(dstName)
		if !validfile {
			return 0, errors.Validationf("not a valid dst file: %v", dstName)
		}
	}

	info, err := os.Stat(srcName)
	if err != nil {
		return 0, errors.InternalErrf(err, "error fetching file information for file: %v", srcName)
	}

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return 0, errors.InternalErrf(err, "error opening dst file: %v", dstName)
	}
	defer dst.Close()
	written, err = io.Copy(dst, src)
	if err != nil {
		return 0, errors.InternalErrf(err, "error copying file from source %v to destination %v",
			srcName, dstName)
	}
	return written, nil
}

// DownloadFile Downloads file through url and copies it to the specified path
func DownloadFile(filepath string, fileurl string) error {
	log.Info().Msgf("Downloading: %s to %s", fileurl, filepath)
	ufile, _ := url.Parse(fileurl)
	switch ufile.Scheme {
	case "http", "https":
		resp, err := http.Get(fileurl) //nolint: noctx, gosec // FIXME: This is not the right way per linter issue. Check more
		// here https://github.com/sonatard/noctx to fix it. More effort and change in this PR and hence skipping this.
		if err != nil {
			return errors.InternalErrf(err, "Failed to get %s", fileurl)
		}
		// close http client before exiting the function
		defer func() {
			if err = resp.Body.Close(); err != nil {
				log.Warn().Msgf("error closing http client: %v", err)
			}
		}()

		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			return errors.InputBodyErrf(err, "Failed to create %s", filepath)
		}
		defer func() {
			_ = out.Close()
		}()
		// Write the body to file
		if _, err = io.Copy(out, resp.Body); err != nil {
			return errors.InternalErrf(err, "error copying file")
		}
		return nil
	case "file":
		_, err := CopyFile(filepath, ufile.Path)
		return err

	default:
		return errors.Validationf("unknown url schema: %v", ufile.Scheme)

	}
}
