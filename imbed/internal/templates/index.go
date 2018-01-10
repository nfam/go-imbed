// Code generated by go-imbed. DO NOT EDIT.

// Package templates holds binary resources embedded into Go executable
package templates

import (
	"os"
	"io"
	"path/filepath"
	"io/ioutil"
	"strings"
	"path"
	"bytes"
	"compress/gzip"
	"time"
)

func blob_bytes(uint32) []byte
func blob_string(uint32) string

// Asset represents binary resource stored within Go executable. Asset implements
// fmt.Stringer and io.WriterTo interfaces, decompressing binary data if necessary.
type Asset struct {
	name         string // File name
	size         int32  // File size (uncompressed)
	blob         []byte // Resource blob []byte
	str_blob     string // Resource blob as string
	isCompressed bool   // true if resources was compressed with gzip
	mime         string // MIME Type
	tag          string // Tag is essentially a Tag of resource content and can be used as a value for "Etag" HTTP header
}

// Name returns the base name of the asset
func (a *Asset) Name() string       { return a.name }
// MimeType returns MIME Type of the asset
func (a *Asset) MimeType() string   { return a.mime }
// IsCompressed returns true of asset has been compressed
func (a *Asset) IsCompressed() bool { return a.isCompressed }

// Size implements os.FileInfo and returns the size of the asset (uncompressed, if asset has been compressed)
func (a *Asset) Size() int64        { return int64(a.size) }
// Mode implements os.FileInfo and always returns 0444
func (a *Asset) Mode() os.FileMode  { return 0444 }
// ModTime implements os.FileInfo and returns the time stamp when this package has been produced (the same value for all the assets)
func (a *Asset) ModTime() time.Time { return stamp }
// IsDir implements os.FileInfo and returns false
func (a *Asset) IsDir() bool        { return false }
// Sys implements os.FileInfo and returns nil
func (a *Asset) Sys() interface{}   { return nil }

// WriteTo implements io.WriterTo interface and writes content of the asset to w
func (a *Asset) WriteTo(w io.Writer) (int64, error) {
	if a.isCompressed {
		ungzip, _ := gzip.NewReader(bytes.NewReader(a.blob))
		defer ungzip.Close()
		return io.Copy(w, ungzip)
	}
	n, err := w.Write(a.blob)
	return int64(n), err
}

// The CopyTo method copies asset content to the target directory.
// If file with the same name, size and modification time exists,
// it will not be overwritten, unless overwrite = true is specified.
func (a *Asset) CopyTo(target string, mode os.FileMode, overwrite bool) error {
	fname := filepath.Join(target, a.name)
	fs, err := os.Stat(fname)
	if err == nil {
		if fs.IsDir() {
			return os.ErrExist
		} else if !overwrite && fs.Size() == a.Size() && fs.ModTime().Equal(a.ModTime()) {
			return nil
		}
	}
	file, err := ioutil.TempFile(target, ".imbed")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = a.WriteTo(file)
	if err != nil {
		return err
	}
	file.Close()
	os.Chtimes(file.Name(), a.ModTime(), a.ModTime())
	os.Chmod(file.Name(), mode)
	return os.Rename(file.Name(), fname)
}

// String returns (uncompressed) content of asset as a string
func (a *Asset) String() string {
	if a.isCompressed {
		ungzip, _ := gzip.NewReader(bytes.NewReader(a.blob))
		ret, _ := ioutil.ReadAll(ungzip)
		ungzip.Close()
		return string(ret)
	}
	return a.str_blob
}

type assetReader struct {
	bytes.Reader
}

func (r *assetReader) Close() error {
	r.Reset(nil)
	return nil
}

// Opens asset as an io.ReadCloser. Returns os.ErrNotExist if no asset is found.
func Open(name string) (File, error) {
	return root.Open(name)
}

// Gets asset by name. Returns nil if no asset found.
func Get(name string) *Asset {
	if entry, ok := idx[name]; ok {
		return entry
	} else {
		return nil
	}
}

// Get asset by name. Panics if no asset found.
func Must(name string) *Asset {
	if entry, ok := idx[name]; ok {
		return entry
	} else {
		panic("asset " + name + " not found")
	}
}

type directoryAsset struct {
	name  string
	dirs  []directoryAsset
	files []Asset
}

var root *directoryAsset

// A File is returned by virtual FileSystem's Open method.
// The methods should behave the same as those on an *os.File.
type File interface {
	io.Closer
	io.Reader
	io.Seeker
	Readdir(count int) ([]os.FileInfo, error)
	Stat() (os.FileInfo, error)
	// The CopyTo method copies file content to the target path.
	// If file with the same name, size and modification time exists,
	// it will not be overwritten, unless overwrite = true is specified.
	// templates.Root().CopyTo(".", mode, false) will effectively
	// extract content of the filesystem to the current directory (which
	// makes it the most space-wise inefficient self-extracting archive
	// ever).
	CopyTo(target string, mode os.FileMode, overwrite bool) error
}

func (d *directoryAsset) Open(name string) (File, error) {
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	p := path.Clean(name)
	if p == "." {
		return &directoryAssetFile{dir: d}, nil
	} else {
		var first, rest string
		i := strings.IndexByte(p, '/')
		if i == -1 {
			first = p
		} else {
			first = p[:i]
			rest = p[i+1:]
		}
		for j := range d.dirs {
			if d.dirs[j].name == first {
				if rest == "" {
					return &directoryAssetFile{dir: &d.dirs[j]}, nil
				} else {
					return d.dirs[j].Open(rest)
				}
			}
		}
		if rest != "" {
			return nil, os.ErrNotExist
		}
		for j := range d.files {
			if d.files[j].name == first {
				if d.files[j].isCompressed {
					ret := &assetCompressedFile{asset: &d.files[j]}
					ret.Reset(bytes.NewReader(d.files[j].blob))
					return ret, nil
				} else {
					ret := &assetFile{asset: &d.files[j]}
					ret.Reset(d.files[j].blob)
					return ret, nil
				}
			}
		}
		return nil, os.ErrNotExist
	}
}

type directoryAssetFile struct {
	dir *directoryAsset
	pos int
}

func (d *directoryAssetFile) Close() error {
	if d.pos < 0 {
		return os.ErrClosed
	}
	d.pos = -1
	return nil
}

func (d *directoryAssetFile) Read([]byte) (int, error) {
	if d.pos < 0 {
		return 0, os.ErrClosed
	}
	return 0, io.EOF
}

func (d *directoryAssetFile) Stat() (os.FileInfo, error) {
	if d.pos < 0 {
		return nil, os.ErrClosed
	}
	return d.dir, nil
}

func (d *directoryAssetFile) Seek(pos int64, whence int) (int64, error) {
	if d.pos < 0 {
		return 0, os.ErrClosed
	}
	if whence == io.SeekStart && pos == 0 {
		d.pos = 0
		return 0, nil
	} else {
		return 0, os.ErrInvalid
	}
}

func (d *directoryAssetFile) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos < 0 {
		return nil, os.ErrClosed
	}
	ret := make([]os.FileInfo, len(d.dir.dirs) + len(d.dir.files))
	i := 0
	for j := range d.dir.dirs {
		ret[j + i] = &d.dir.dirs[j]
	}
	i = len(d.dir.dirs)
	for j := range d.dir.files {
		ret[j + i] = &d.dir.files[j]
	}
	if count <= 0 {
		return ret, nil
	} else if d.pos > len(ret) {
		return nil, io.EOF
	} else {
		return ret[d.pos:d.pos+count], nil
	}
}

func (d *directoryAsset) copyTo(target string, dirmode os.FileMode, mode os.FileMode, overwrite bool) error {
	dname := filepath.Join(target, d.name)
	err := os.MkdirAll(dname, dirmode)
	if err != nil {
		return err
	}
	for i := range d.dirs {
		if err = d.dirs[i].copyTo(dname, dirmode, mode, overwrite); err != nil {
			return err
		}
	}
	for i := range d.files {
		if err = d.files[i].CopyTo(dname, mode, overwrite); err != nil {
			return err
		}
	}
	return nil
}

func (d *directoryAssetFile) CopyTo(target string, mode os.FileMode, overwrite bool) error {
	dirmode := ((mode&0444)>>2)|mode
	return d.dir.copyTo(target, dirmode, mode, overwrite)
}

func (d *directoryAsset) Name() string       { return d.name }
func (d *directoryAsset) Size() int64        { return 0 }
func (d *directoryAsset) Mode() os.FileMode  { return os.ModeDir | 0555 }
func (d *directoryAsset) ModTime() time.Time { return stamp }
func (d *directoryAsset) IsDir() bool        { return true }
func (d *directoryAsset) Sys() interface{}   { return nil }

type assetFile struct {
	assetReader
	asset *Asset
}

func (a *assetFile) Stat() (os.FileInfo, error) {
	return a.asset, nil
}

func (a *assetFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (a *assetFile) CopyTo(target string, mode os.FileMode, overwrite bool) error {
	return a.asset.CopyTo(target, mode, overwrite)
}
type assetCompressedFile struct {
	gzip.Reader
	asset *Asset
}

func (a *assetCompressedFile) Stat() (os.FileInfo, error) {
	return a.asset, nil
}

func (a *assetCompressedFile) Seek(int64, int) (int64, error) {
	return 0, os.ErrInvalid
}

func (a *assetCompressedFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (a *assetCompressedFile) CopyTo(target string, mode os.FileMode, overwrite bool) error {
	return a.asset.CopyTo(target, mode, overwrite)
}

var idx = make(map[string]*Asset)
var stamp time.Time

func init() {
	stamp = time.Unix(1515560467,162862000)
	bb := blob_bytes(5472)
	bs := blob_string(5472)
	root = &directoryAsset{
		files: []Asset{
			{
				name:         "index.go",
				blob:         bb[0:3784],
				str_blob:     bs[0:3784],
				mime:         "text/x-golang; charset=utf-8",
				tag:          "t2uskxucxck66",
				size:         13115,
				isCompressed: true,
			},
			{
				name:         "index_386.s",
				blob:         bb[3784:3981],
				str_blob:     bs[3784:3981],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "ihibzfvzsneuc",
				size:         327,
				isCompressed: true,
			},
			{
				name:         "index_amd64.s",
				blob:         bb[3984:4192],
				str_blob:     bs[3984:4192],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "dcfwghvd5ccho",
				size:         361,
				isCompressed: true,
			},
			{
				name:         "index_arm.s",
				blob:         bb[4192:4384],
				str_blob:     bs[4192:4384],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "37iklflan4xc4",
				size:         327,
				isCompressed: true,
			},
			{
				name:         "index_arm64.s",
				blob:         bb[4384:4583],
				str_blob:     bs[4384:4583],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "slzknys4x76m2",
				size:         329,
				isCompressed: true,
			},
			{
				name:         "index_mips64x.s",
				blob:         bb[4584:4805],
				str_blob:     bs[4584:4805],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "fu6srp6typnuy",
				size:         368,
				isCompressed: true,
			},
			{
				name:         "index_mipsx.s",
				blob:         bb[4808:5026],
				str_blob:     bs[4808:5026],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "ev3x2mivyfazw",
				size:         362,
				isCompressed: true,
			},
			{
				name:         "index_ppc64x.s",
				blob:         bb[5032:5246],
				str_blob:     bs[5032:5246],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "uvrbimoxyzy3a",
				size:         354,
				isCompressed: true,
			},
			{
				name:         "index_s390x.s",
				blob:         bb[5248:5465],
				str_blob:     bs[5248:5465],
				mime:         "text/x-asm; charset=utf-8",
				tag:          "kyrcf7qm7xney",
				size:         311,
				isCompressed: true,
			},
		},
	}
	idx["index.go"] = &root.files[0]
	idx["index_386.s"] = &root.files[1]
	idx["index_amd64.s"] = &root.files[2]
	idx["index_arm.s"] = &root.files[3]
	idx["index_arm64.s"] = &root.files[4]
	idx["index_mips64x.s"] = &root.files[5]
	idx["index_mipsx.s"] = &root.files[6]
	idx["index_ppc64x.s"] = &root.files[7]
	idx["index_s390x.s"] = &root.files[8]
}