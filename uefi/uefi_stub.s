// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !amd64

#include "textflag.h"

// func callFn(fn uint64, n int, args []uint64) (status uint64)
TEXT ·callFn(SB),NOSPLIT,$0-48
	RET
