// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

const (
	testdatapath = "testdata"
	RmFile       = "filetobeRemove.yml"
)

var (
	testErr = fmt.Errorf("test error")
)

// nolint
func patchReadlink(t *testing.T, name string, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Readlink, func(_ string) (string, error) {
		unpatch(t, patch)
		return name, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchOsStat(t *testing.T, fileInfo fs.FileInfo, err error) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Stat, func(_ string) (fs.FileInfo, error) {
		return fileInfo, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

// nolint
func patchOsChmod(t *testing.T, err error, nextPatch func()) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Chmod, func(_ string, _ fs.FileMode) error {
		unpatch(t, patch)
		if nextPatch != nil {
			nextPatch()
		}
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchOsOpenFile(t *testing.T, file *os.File, err error) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.OpenFile, func(fileName string, _ int, _ fs.FileMode) (*os.File, error) {
		return file, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

// nolint
func patchOsOpen(t *testing.T, file *os.File, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Open, func(_ string) (*os.File, error) {
		unpatch(t, patch)
		return file, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

// nolint
func patchOsCreate(t *testing.T, file *os.File, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Create, func(_ string) (*os.File, error) {
		unpatch(t, patch)
		return file, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchHttpGet(t *testing.T, resp *http.Response, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(http.Get, func(_ string) (*http.Response, error) {
		unpatch(t, patch)
		return resp, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

//nolint:unparam
func patchIoCopy(t *testing.T, written int64, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(io.Copy, func(_ io.Writer, _ io.Reader) (int64, error) {
		unpatch(t, patch)
		return written, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchIoCopyReturnMpatch(t *testing.T, written int64, err error) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(io.Copy, func(_ io.Writer, _ io.Reader) (int64, error) {
		return written, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

// nolint
func patchFileWriteString(t *testing.T, n int, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&os.File{}), "WriteString", func(_ *os.File, _ string) (int, error) {
		unpatch(t, patch)
		return n, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

type JsonTest struct {
	Cn    string   `json:"CN"`
	Hosts []string `json:"hosts"`
}

func TestRemoveFile(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		retError       error
		expectError    bool
		funcBeforeTest func()
	}{
		{
			name:        "Remove file fail",
			input:       []string{filepath.Join(testdatapath, "not_exist.yml")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				var patch *mpatch.Patch
				var err error
				patch, err = mpatch.PatchMethod(os.Remove, func(_ string) error {
					unpatch(t, patch)
					return testErr
				})
				if err != nil {
					t.Errorf("patch error: %v", err)
				}
			},
		},
		{
			name:        "Remove file OK",
			input:       []string{filepath.Join(testdatapath, RmFile)},
			retError:    nil,
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			result := RemoveFile(tc.input[0])
			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	func_download_ok := func() []*mpatch.Patch {
		resp := &http.Response{Body: io.NopCloser(strings.NewReader("test reader"))}
		p1, err := mpatch.PatchMethod(http.Get, func(string) (*http.Response, error) { return resp, nil })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	cases := []struct {
		name        string
		input       []string
		retError    error
		expectError bool
		funcPatch   func() []*mpatch.Patch
	}{
		{
			"Download file fail",
			[]string{"", filepath.Join(testdatapath, "")},
			nil,
			true,
			nil,
		},
		{
			"Download http get error",
			[]string{filepath.Join(testdatapath, "kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			true,
			func() []*mpatch.Patch {
				patchHttpGet(t, nil, testErr)
				return nil
			},
		},
		{
			"Download http create file error",
			[]string{filepath.Join(testdatapath, "/kind/kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			true,
			func() []*mpatch.Patch {
				resp := &http.Response{}
				resp.Body = io.NopCloser(strings.NewReader(""))
				patchHttpGet(t, resp, nil)
				return nil
			},
		},
		{
			"Download http write file error",
			[]string{filepath.Join(testdatapath, "/kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			true,
			func() []*mpatch.Patch {
				resp := &http.Response{}
				resp.Body = io.NopCloser(strings.NewReader(""))
				resp.Body.Close()
				patchHttpGet(t, resp, nil)
				patchIoCopy(t, 0, testErr)
				return nil
			},
		},
		{
			"Download file from file",
			[]string{filepath.Join(testdatapath, "destfile.yml"), filepath.Join(testdatapath, "fileutil1.yml")},
			nil,
			true,
			nil,
		},
		{
			"Download file from file error",
			[]string{filepath.Join(testdatapath, "destfile.yml"), fmt.Sprintf("file://%s", filepath.Join(testdatapath, "aaa.yml"))},
			nil,
			true,
			nil,
		},
		{
			"Download file from https OK",
			[]string{filepath.Join(testdatapath, "kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			false,
			func_download_ok,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcPatch != nil {
				plist := tc.funcPatch()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			result := DownloadFile(tc.input[0], tc.input[1])
			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}
	//Remove downloaded files after download test
	err1 := RemoveFile(filepath.Join(testdatapath, "destfile.yml"))
	err2 := RemoveFile(filepath.Join(testdatapath, "kind"))
	if err1 != nil || err2 != nil {
		t.Error("Download file test not finished!")
	}
}

func TestCheckFileLink(t *testing.T) {
	type args struct {
		filename string
	}

	tests := []struct {
		name string
		args args
		want Filelink
	}{
		// TODO: Add test cases.
		{
			"check file, normal link",
			args{filename: filepath.Join(testdatapath, "fileutil1.yml")},
			Normalfile,
		},
		{
			"check file, symbol link",
			args{filename: filepath.Join(testdatapath, "symbol.yaml")},
			Symbollink,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckFileLink(tt.args.filename); got != tt.want {
				t.Errorf("CheckFileLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			"normal file is valid",
			args{filename: filepath.Join(testdatapath, "fileutil1.yml")},
			true,
		},
		{
			"symbol file is valid",
			args{filename: filepath.Join(testdatapath, "symbol.yaml")},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidFile(tt.args.filename); got != tt.want {
				t.Errorf("IsValidFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	type args struct {
		dstName string
		srcName string
	}
	tests := []struct {
		name      string
		args      args
		want      bool
		funcPatch func() []*mpatch.Patch
	}{
		{
			name: "Test Successful copy",
			args: args{
				dstName: "testdata/dstFile.yml",
				srcName: "testdata/fileutil1.yml",
			},
			want: false,
		},
		{
			name: "Test failure copy - os.Stat error",
			args: args{
				dstName: "testdata/dstFile.yml",
				srcName: "testdata/fileutil1.yml",
			},
			want: true,
			funcPatch: func() []*mpatch.Patch {
				patch1 := patchOsStat(t, nil, fmt.Errorf("error in os.Stat"))
				return []*mpatch.Patch{patch1}
			},
		},
		{
			name: "Test failure copy - OpenFile error",
			args: args{
				dstName: "testdata/dstFile.yml",
				srcName: "testdata/fileutil1.yml",
			},
			want: true,
			funcPatch: func() []*mpatch.Patch {
				patch1 := patchOsOpenFile(t, nil, fmt.Errorf("error in os.OpenFile"))
				return []*mpatch.Patch{patch1}
			},
		},
		{
			name: "Test failure copy - io.Copy error",
			args: args{
				dstName: "testdata/dstFile.yml",
				srcName: "testdata/fileutil1.yml",
			},
			want: true,
			funcPatch: func() []*mpatch.Patch {
				patch1 := patchIoCopyReturnMpatch(t, 10, fmt.Errorf("error in io.Copy"))
				return []*mpatch.Patch{patch1}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.funcPatch != nil {
				plist := tt.funcPatch()
				defer unpatchAll(t, plist)
			}
			written, err := CopyFile(tt.args.dstName, tt.args.srcName)
			if (err != nil) != tt.want {
				t.Errorf("got: %v, want: %v", err, tt.want)
				return
			}
			if err == nil && written == 0 {
				t.Errorf("no data written")
				return
			}
		})
	}
}
