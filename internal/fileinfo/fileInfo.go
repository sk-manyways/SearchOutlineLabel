package fileinfo

import "io/fs"

type Full struct {
	fs.FileInfo
	fullPath string
}

func NewFull(fi fs.FileInfo, fullPath string) Full {
	return Full{
		FileInfo: fi,
		fullPath: fullPath,
	}
}

func (f Full) FullPath() string {
	return f.fullPath
}
