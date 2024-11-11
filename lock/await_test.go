// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"testing"
	"time"
)

func BenchmarkAwait(b *testing.B) {
	aw := NewAwait[int](10) // 使用较

	// 基准测试 100 个并发任务
	b.Run("Benchmark 100 tasks", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			idx := int64(i)
			go func(idx int64) {
				// 模拟工作，延迟 100 毫秒
				time.Sleep(1 * time.Millisecond)
				aw.Close(idx)
			}(idx)
			// 等待任务完成
			aw.Wait(idx, time.Millisecond)
		}
	})
}
