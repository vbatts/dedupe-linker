package file

import (
	"fmt"
	"os"
	"syscall"
)

// SameInodePaths checks whether path1 and path2 are the same inode
func SameInodePaths(path1, path2 string) (match bool, err error) {
	var inode1, inode2 uint64
	if inode1, err = GetInode(path1); err != nil {
		return false, err
	}
	if inode2, err = GetInode(path2); err != nil {
		return false, err
	}
	return inode1 == inode2, nil
}

// SameDevPaths checks whether path1 and path2 are on the same device
func SameDevPaths(path1, path2 string) (match bool, err error) {
	var dev1, dev2 uint64
	if dev1, err = GetDev(path1); err != nil {
		return false, err
	}
	if dev2, err = GetDev(path2); err != nil {
		return false, err
	}
	return dev1 == dev2, nil
}

// FormatDev has a scary name, but just pretty prints the stat_t.dev as "major:minor"
func FormatDev(stat *syscall.Stat_t) string {
	return fmt.Sprintf("%d:%d", MajorDev(stat.Dev), MinorDev(stat.Dev))
}

// MajorDev provides the major device number from a stat_t.dev
func MajorDev(dev uint64) uint64 {
	return (((dev >> 8) & 0xfff) | ((dev >> 32) & ^uint64(0xfff)))
}

// MinorDev provides the minor device number from a stat_t.dev
func MinorDev(dev uint64) uint64 {
	return ((dev & 0xff) | ((dev >> 12) & ^uint64(0xff)))
}

// GetLstat returns the system stat_t for the file at path.
// (symlinks are not deferenced)
func GetLstat(path string) (*syscall.Stat_t, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	return fi.Sys().(*syscall.Stat_t), nil
}

// GetStat returns the system stat_t for the file at path.
// (symlinks are deferenced)
func GetStat(path string) (*syscall.Stat_t, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return fi.Sys().(*syscall.Stat_t), nil
}

// GetInode returns the inode for path
func GetInode(path string) (uint64, error) {
	stat, err := GetStat(path)
	if err != nil {
		return 0, err
	}
	return stat.Ino, nil
}

// GetDev returns the device for path
func GetDev(path string) (uint64, error) {
	stat, err := GetStat(path)
	if err != nil {
		return 0, err
	}
	return stat.Dev, nil
}

// SameFile checks whether the two paths are same device and inode
func SameFile(fpath1, fpath2 string) bool {
	bStat, err := GetStat(fpath1)
	if err != nil {
		return false
	}
	dStat, err := GetStat(fpath2)
	if err != nil {
		return false
	}
	if bStat.Dev != dStat.Dev {
		return false
	}
	if bStat.Ino != dStat.Ino {
		return false
	}
	// if we made it here, we must be ok
	return true

}

// GetNlink returns the number of links for path. For directories, that is
// number of entries. For regular files, that is number of hardlinks.
func GetNlink(path string) (uint64, error) {
	stat, err := GetStat(path)
	if err != nil {
		return 0, err
	}
	return stat.Nlink, nil
}
