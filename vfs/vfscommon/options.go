package vfscommon

import (
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/filter"
)

// Options is options for creating the vfs
type Options struct {
	NoSeek             bool          // don't allow seeking if set
	NoChecksum         bool          // don't check checksums if set
	ReadOnly           bool          // if set VFS is read only
	NoModTime          bool          // don't read mod times for files
	DirCacheTime       time.Duration // how long to consider directory listing cache valid
	PollInterval       time.Duration
	Umask              int
	UID                uint32
	GID                uint32
	DirPerms           os.FileMode
	FilePerms          os.FileMode
	ChunkSize          fs.SizeSuffix // if > 0 read files in chunks
	ChunkSizeLimit     fs.SizeSuffix // if > ChunkSize double the chunk size after each chunk until reached
	CacheMode          CacheMode
	CacheMaxAge        time.Duration
	CacheMaxSize       fs.SizeSuffix
	CacheMinFreeSpace  fs.SizeSuffix
	CachePollInterval  time.Duration
	CaseInsensitive    bool
	WriteWait          time.Duration // time to wait for in-sequence write
	ReadWait           time.Duration // time to wait for in-sequence read
	WriteBack          time.Duration // time to wait before writing back dirty files
	ReadAhead          fs.SizeSuffix // bytes to read ahead in cache mode "full"
	UsedIsSize         bool          // if true, use the `rclone size` algorithm for Used size
	FastFingerprint    bool          // if set use fast fingerprints
	DiskSpaceTotalSize fs.SizeSuffix
	VfsUploadExclude   []string
	VfsExcludeRegex    []*regexp.Regexp
}

// DefaultOpt is the default values uses for Opt
var DefaultOpt = Options{
	NoModTime:          false,
	NoChecksum:         false,
	NoSeek:             false,
	DirCacheTime:       5 * 60 * time.Second,
	PollInterval:       time.Minute,
	ReadOnly:           false,
	Umask:              0,
	UID:                ^uint32(0), // these values instruct WinFSP-FUSE to use the current user
	GID:                ^uint32(0), // overridden for non windows in mount_unix.go
	DirPerms:           os.FileMode(0777),
	FilePerms:          os.FileMode(0666),
	CacheMode:          CacheModeOff,
	CacheMaxAge:        3600 * time.Second,
	CachePollInterval:  60 * time.Second,
	ChunkSize:          128 * fs.Mebi,
	ChunkSizeLimit:     -1,
	CacheMaxSize:       -1,
	CacheMinFreeSpace:  -1,
	CaseInsensitive:    runtime.GOOS == "windows" || runtime.GOOS == "darwin", // default to true on Windows and Mac, false otherwise
	WriteWait:          1000 * time.Millisecond,
	ReadWait:           20 * time.Millisecond,
	WriteBack:          5 * time.Second,
	ReadAhead:          0 * fs.Mebi,
	UsedIsSize:         false,
	DiskSpaceTotalSize: -1,
}

func initializeExclusionPatterns(opt *Options) {
	for _, pattern := range opt.VfsUploadExclude {
		regexPattern, err := filter.GlobToRegexp(pattern, true)
		if err != nil {
			fs.Errorf(pattern, "Could not generate regex from glob for VFS cache exclusion: %v", err)
			continue
		}
		opt.VfsExcludeRegex = append(opt.VfsExcludeRegex, regexPattern)
	}
}

// Init the options, making sure everything is within range
func (opt *Options) Init() {
	// Mask the permissions with the umask
	opt.DirPerms &= ^os.FileMode(opt.Umask)
	opt.FilePerms &= ^os.FileMode(opt.Umask)

	// Make sure directories are returned as directories
	opt.DirPerms |= os.ModeDir
	initializeExclusionPatterns(opt)
}
