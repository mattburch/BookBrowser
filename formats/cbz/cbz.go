package cbz

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/geek1011/BookBrowser/booklist"
	"github.com/geek1011/BookBrowser/formats"

	"github.com/pkg/errors"
)

type cbz struct {
	hascover  bool
	book      *booklist.Book
	coverpath *string
}

func (e *cbz) Book() *booklist.Book {
	return e.book
}

func (e *cbz) HasCover() bool {
	return e.coverpath != nil
}

func (e *cbz) GetCover() (i image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("panic while decoding cover image")
		}
	}()

	zr, err := zip.OpenReader(e.book.FilePath)
	if err != nil {
		return nil, errors.Wrap(err, "error opening cbz as zip")
	}
	defer zr.Close()

	cr, err := zr.File[0].Open()
	if err != nil {
		return nil, errors.Wrapf(err, "could not open cover '%s'", *e.coverpath)
	}

	i, _, err = image.Decode(cr)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding image")
	}

	return i, nil
}

func load(filename string) (formats.BookInfo, error) {
	e := &cbz{book: &booklist.Book{}, hascover: false}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, errors.Wrapf(err, "could not stat book")
	}
	e.book.FilePath = filename
	e.book.FileSize = fi.Size()
	e.book.ModTime = fi.ModTime()

	s := sha1.New()
	i, err := io.Copy(s, f)
	if err == nil && i != fi.Size() {
		err = errors.New("could not read whole file")
	}
	if err != nil {
		f.Close()
		return nil, errors.Wrap(err, "could not hash book")
	}
	e.book.Hash = fmt.Sprintf("%x", s.Sum(nil))

	f.Close()

	zr, err := zip.OpenReader(filename)
	if err != nil {
		return nil, errors.Wrap(err, "error opening cbz as zip")
	}
	defer zr.Close()

	// CBZ format does not carry metadata
	e.book.Title = filepath.Base(filename)
	e.coverpath = &zr.File[0].Name

	return e, nil
}

func init() {
	formats.Register("cbz", load)
}
