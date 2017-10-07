// The Isomorphic Go Project
// Copyright (c) Wirecog, LLC. All rights reserved.
// Use of this source code is governed by a BSD-style
// license, which can be found in the LICENSE file.

package html

func (t *Tokenizer) CurrentOffset() [2]int {
	return [2]int{t.data.start, t.data.end}
}
