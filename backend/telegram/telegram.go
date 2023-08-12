package telegram

import (
	"context"
	"errors"
	"io"
	"strings"
	"github.com/dceldran/rclone/fs"
	"github.com/dceldran/rclone/fs/config"
	"github.com/dceldran/rclone/fs/config/configmap"
	"github.com/dceldran/rclone/fs/config/configstruct"
	"github.com/dceldran/rclone/fs/fshttp"
	"github.com/dceldran/rclone/fs/hash"
	"github.com/dceldran/rclone/lib/readers"
	"github.com/dceldran/rclone/vfs"
	"gopkg.in/telebot.v3"
	"time"
)

// Register with Fs
func init() {
	fs.Register(&fs.RegInfo{
		Name:        "telegram",
		Description: "Telegram",
		NewFs:       NewFs,
		Options: []fs.Option{{
			Name:     "token",
			Help:     "Telegram bot token.",
			Required: true,
		}, {
			Name:     "chat_id",
			Help:     "Telegram chat ID.",
			Required: true,
		}},
	})
}

// Options defines the configuration for this backend
type Options struct {
	Token  string `config:"token"`
	ChatID int64 `config:"chat_id"`
}

// Fs represents a remote Telegram chat
type Fs struct {
	name     string
	root     string
	features *fs.Features
	bot      *telebot.Bot
	chatID   int64
}

// NewFs constructs a new Fs
func NewFs(ctx context.Context, name, root string, m configmap.Mapper) (fs.Fs, error) {
	options := new(Options)
	if err := configstruct.Set(m, options); err != nil {
		return nil, err
	}

	bot, err := telebot.NewBot(telebot.Settings{Token: options.Token})
	if err != nil {
		return nil, err
	}

	f := &Fs{
		name:     name,
		root:     "/",
		bot:      bot,
		chatID:   options.ChatID,
	}

	return f, nil
}

// Name of the remote (as passed into NewFs)
func (f *Fs) Name() string {
	return f.name
}

// Root of the remote (as passed into NewFs)
func (f *Fs) Root() string {
	return f.root
}

// String returns a description of the FS
func (f *Fs) String() string {
	return "Telegram chat " + f.name
}

// Features returns the optional features of this Fs
func (f *Fs) Features() *fs.Features {
	return f.features
}

// Hashes returns the supported hash sets
func (f *Fs) Hashes() hash.Set {
	return hash.Set(hash.None)
}

// Put uploads contents to the remote path
func (f *Fs) Put(ctx context.Context, in io.ReadCloser, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	if src.Size() > int64(2<<30) {
		return nil, errors.New("telegram backend only supports files up to 2GB in size")
	}

	fileName := src.Remote()
	if strings.HasPrefix(fileName, "/") {
		fileName = fileName[1:]
	}

	file := telebot.Document{
		File:   telebot.FromReader(readers.NewLimitedReadCloser(in, src.Size())),
		Caption: fileName,
	}

	message, err := f.bot.Send(f.bot.Me, &file)
	if err != nil {
		return nil, err
	}

	return &Object{
		fs:        f,
		remote:	   fileName,
		path:      "/" + message.Document.FileID,
		name:      fileName,
		size:      src.Size(),
		modTime:   message.Date,
		isDir:     false,
	}, nil
}

// List returns a channel to the objects and subdirectories
// in dir with directory entries popped from the channel
func (f *Fs) List(ctx context.Context, dir string) (fs.DirChan, fs.EntryChan, error) {
	return nil, nil, errors.New("telegram backend does not support directory listing")
}

// NewObject finds the Object at remote
func (f *Fs) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	message, err := f.bot.GetFile(remote[1:])
	if err != nil {
		return nil, err
	}

	return &Object{
		fs:        f,
		path:      remote,
		name:      message.FilePath,
		size:      int64(message.FileSize),
		modTime:   time.Unix(message.Date, 0),
		isDir:     false,
	}, nil
}

// PutStream uploads contents to the remote path using a stream
func (f *Fs) PutStream(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	return f.Put(ctx, in, src, options...)
}

// Mkdir creates the directory if it doesn't exist
func (f *Fs) Mkdir(ctx context.Context, dir string) error {
	return errors.New("telegram backend does not support directory creation")
}

// Rmdir removes the directory
func (f *Fs) Rmdir(ctx context.Context, dir string) error {
	return errors.New("telegram backend does not support directory removal")
}

// Precision of the ModTimes in this Fs
func (f *Fs) Precision() time.Duration {
	return fs.ModTimeNotSupported
}

// Object represents a remote Telegram file
type Object struct {
	fs        *Fs
	remote    string
	path      string
	name      string
	size      int64
	modTime   time.Time
	isDir     bool
}

// Fs returns the parent Fs
func (o *Object) Fs() fs.Info {
	return o.fs
}

// String returns a description of the Object
func (o *Object) String() string {
	return o.path
}

// Remote returns the remote path
func (o *Object) Remote() string {
	return o.path
}

// Hash returns the selected checksum of the file
// If no checksum is available it returns ""
func (o *Object) Hash(r hash.Type) (string, error) {
	return "", hash.ErrUnsupported
}

// Size returns the size of the file
func (o *Object) Size() int64 {
	return o.size
}

// ModTime returns the modification time of the file
func (o *Object) ModTime(ctx context.Context) time.Time {
	return o.modTime
}

// SetModTime sets the modification time of the file
func (o *Object) SetModTime(ctx context.Context, modTime time.Time) error {
	return errors.New("telegram backend does not support modification times")
}

// Storable returns whether this object can be stored
func (o *Object) Storable() bool {
	return true
}

// Open opens the file for read
func (o *Object) Open(ctx context.Context, options ...fs.OpenOption) (io.ReadCloser, error) {
	fileBytes, err := o.fs.bot.Download(&telebot.File{FileID: o.path[1:]})
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

// Update updates the object from in with modTime
func (o *Object) Update(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) error {
	return errors.New("telegram backend does not support object updates")
}

// Remove deletes the remote object
func (o *Object) Remove(ctx context.Context) error {
	_, err := o.fs.bot.Delete(&telebot.Message{Document: &telebot.Document{FileID: o.path[1:]}})
	return err
}
