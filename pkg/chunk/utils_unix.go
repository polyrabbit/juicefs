//go:build !windows
// +build !windows

/*
 * JuiceFS, Copyright 2020 Juicedata, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package chunk

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func getNlink(fi os.FileInfo) int {
	if sst, ok := fi.Sys().(*syscall.Stat_t); ok {
		return int(sst.Nlink)
	}
	return 1
}

func getDiskUsage(path string) (uint64, uint64, uint64, uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err == nil {
		return stat.Blocks * uint64(stat.Bsize), stat.Bavail * uint64(stat.Bsize), stat.Files, stat.Ffree
	} else {
		logger.Warnf("statfs %s: %s", path, err)
		return 1, 1, 1, 1
	}
}

func changeMode(dir string, st os.FileInfo, mode os.FileMode) {
	sst := st.Sys().(*syscall.Stat_t)
	if os.Getuid() == int(sst.Uid) {
		_ = os.Chmod(dir, mode)
	}
}

func inRootVolume(dir string) bool {
	dstat, err := os.Stat(dir)
	if err != nil {
		logger.Warnf("stat `%s`: %s", dir, err.Error())
		return false
	}
	rstat, err := os.Stat("/")
	if err != nil {
		logger.Warnf("stat `/`: %s", err.Error())
		return false
	}
	return dstat.Sys().(*syscall.Stat_t).Dev == rstat.Sys().(*syscall.Stat_t).Dev
}

var dropPageCache = os.Getenv("JFS_DROP_PAGE_CACHE") == "1"

func (cf *cacheFile) Close() error {
	if dropPageCache {
		err := unix.Fadvise(int(cf.File.Fd()), 0, 0, unix.FADV_DONTNEED)
		if err != nil {
			logger.Warnf("fadvise %s: %s", cf.File.Name(), err)
		}
	}
	return cf.File.Close()
}

func (cw noPageCacheWriter) Close() error {
	if dropPageCache {
		err := cw.File.Sync()
		if err != nil {
			return err
		}
		err = unix.Fadvise(int(cw.File.Fd()), 0, 0, unix.FADV_DONTNEED)
		if err != nil {
			logger.Warnf("fadvise %s: %s", cw.File.Name(), err)
		}
	}
	return cw.File.Close()
}
