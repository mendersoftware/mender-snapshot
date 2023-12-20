// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package utils

import (
	"github.com/mendersoftware/progressbar"
)

type ProgressWriter struct {
	bar *progressbar.Bar
}

func NewProgressWriter(size int64) *ProgressWriter {
	return &ProgressWriter{
		bar: progressbar.New(size),
	}
}

func (p *ProgressWriter) Tick(n uint64) {
	p.bar.Tick(int64(n))
}
