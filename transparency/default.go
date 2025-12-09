// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !transparency

package transparency

import (
	"io/fs"
)

func Check(fsys fs.FS, online bool) (err error) {
	return
}
