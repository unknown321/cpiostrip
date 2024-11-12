package strip

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

const MtimeOffset = 6 + 5*8
const HeaderSize = 110

var Trailer = []byte("TRAILER!!!")
var Magic = []byte("070701")
var EpochBytes = [8]byte{}

var EpochStart = time.Unix(0, 0)

type CpioHeader struct {
	Magic     [6]byte
	Ino       [8]byte
	Mode      [8]byte
	Uid       [8]byte
	Gid       [8]byte
	Nlink     [8]byte
	Mtime     [8]byte
	Filesize  [8]byte
	Devmajor  [8]byte
	Devminor  [8]byte
	Rdevmajor [8]byte
	Rdevminor [8]byte
	Namesize  [8]byte
	Check     [8]byte
}

func (m *CpioHeader) String() string {
	return fmt.Sprintf("m.Magic:     %v\nm.Ino:       %v\nm.Mode:      %v\nm.Uid:       %v\nm.Gid:       %v\nm.Nlink:     %v\nm.Mtime:     %v\nm.Filesize:  %v\nm.Devmajor:  %v\nm.Devminor:  %v\nm.Rdevmajor: %v\nm.Rdevminor: %v\nm.Namesize:  %v\nm.Check:     %v\n",
		m.Magic,
		m.Ino,
		m.Mode,
		m.Uid,
		m.Gid,
		m.Nlink,
		m.Mtime,
		m.Filesize,
		m.Devmajor,
		m.Devminor,
		m.Rdevmajor,
		m.Rdevminor,
		m.Namesize,
		m.Check)
}

func FromHex(in []byte) (int, error) {
	dst := make([]byte, hex.DecodedLen(len(in)))
	if _, err := hex.Decode(dst, in); err != nil {
		return 0, err
	}

	return int(binary.BigEndian.Uint32(dst)), nil
}

func readHeader(f *os.File) (*CpioHeader, error) {
	var err error
	h := &CpioHeader{}
	b := []byte{}
	buf := bytes.NewBuffer(b)
	if _, err = io.CopyN(buf, f, HeaderSize); err != nil {
		return nil, err
	}

	if err = binary.Read(buf, binary.LittleEndian, h); err != nil {
		return nil, fmt.Errorf("cannot read %w", err)
	}

	if bytes.Compare(h.Magic[:], Magic) != 0 {
		return nil, fmt.Errorf("wrong magic")
	}

	return h, nil
}

func getName(h *CpioHeader, data *bytes.Reader) ([]byte, error) {
	size, err := FromHex(h.Namesize[:])
	if err != nil {
		return nil, err
	}

	namePadding := (2 - size) & 3
	nameBuf := make([]byte, size+namePadding)
	_, err = data.Read(nameBuf)
	if err != nil {
		return nil, err
	}

	nameBuf = bytes.TrimRight(nameBuf, "\x00")
	return nameBuf, nil
}

func getData(h *CpioHeader, f *os.File) (data []byte, err error) {
	nameSize, err := FromHex(h.Namesize[:])
	if err != nil {
		return nil, err
	}

	namePadding := (2 - nameSize) & 3

	dataSize, err := FromHex(h.Filesize[:])
	if err != nil {
		return nil, err
	}

	dataPadding := 3 & -int64(dataSize)

	slog.Debug("", "nameSize", nameSize, "namePadding", namePadding, "dataSize", dataSize, "dataPadding", dataPadding)
	buf := make([]byte, nameSize+namePadding+dataSize+int(dataPadding))
	if _, err = f.Read(buf); err != nil {
		return nil, fmt.Errorf("cannot read data after header: %w", err)
	}

	return buf, nil
}

func inplace(f *os.File) (err error) {
	var h *CpioHeader
	for {
		if h, err = readHeader(f); err != nil {
			return err
		}

		if _, err = f.Seek(-HeaderSize+MtimeOffset, io.SeekCurrent); err != nil {
			return fmt.Errorf("cannot seek to mtime: %w", err)
		}

		if _, err = f.Write(EpochBytes[:]); err != nil {
			return err
		}

		if _, err = f.Seek(HeaderSize-MtimeOffset-8, io.SeekCurrent); err != nil {
			return err
		}

		fSize, err := FromHex(h.Filesize[:])
		if err != nil {
			return err
		}

		fNameSize, err := FromHex(h.Namesize[:])
		if err != nil {
			return err
		}

		namepad := (2 - fNameSize) & 3
		nameBuf := make([]byte, fNameSize+namepad)
		_, err = f.Read(nameBuf)
		if err != nil {
			return err
		}

		nameBuf = bytes.TrimRight(nameBuf, "\x00")

		filePadding := 3 & -int64(fSize)
		if _, err = f.Seek(int64(fSize)+filePadding, io.SeekCurrent); err != nil {
			return err
		}

		if bytes.Compare(nameBuf, Trailer) == 0 {
			break
		}

		slog.Info("stripped", "filename", string(nameBuf))
	}

	return nil
}

func toFile(f *os.File, out *os.File) (err error) {
	var h *CpioHeader
	for {
		if h, err = readHeader(f); err != nil {
			return err
		}

		h.Mtime = EpochBytes
		if err = binary.Write(out, binary.LittleEndian, h); err != nil {
			return fmt.Errorf("cannot write header: %w", err)
		}

		var data []byte
		data, err = getData(h, f)
		if err != nil {
			return fmt.Errorf("cannot get data: %w", err)
		}

		var name []byte
		if name, err = getName(h, bytes.NewReader(data)); err != nil {
			return fmt.Errorf("cannot get name: %w", err)
		}

		if _, err = out.Write(data); err != nil {
			return fmt.Errorf("cannot write to outfile: %w", err)
		}

		if bytes.Compare(name, Trailer) == 0 {
			var curpos int64
			if curpos, err = f.Seek(0, 1); err != nil {
				return fmt.Errorf("cannot get curpos: %w", err)
			}

			var info os.FileInfo
			if info, err = f.Stat(); err != nil {
				return fmt.Errorf("cannot stat file: %w", err)
			}

			buf := make([]byte, info.Size()-curpos)
			if _, err = f.Read(buf); err != nil {
				return fmt.Errorf("cannot read rest of the file: %w", err)
			}

			if _, err = out.Write(buf); err != nil {
				return fmt.Errorf("cannot write rest of the file: %w", err)
			}

			break
		}

		slog.Info("stripped", "filename", string(name))
	}

	return nil
}

// Strip sets modification time for all files and directories in cpio archive to `Thu Jan  1 12:00:00 AM GMT 1970`
func Strip(filename string, outName string) error {
	binary.BigEndian.PutUint32(EpochBytes[:], uint32(EpochStart.Unix()))
	var err error

	f, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("cannot open file for stripping: %w", err)
	}
	defer f.Close()

	if filename == outName {
		slog.Info("stripping inplace", "filename", outName)
		if err = inplace(f); err != nil {
			return err
		}
	} else {
		slog.Info("stripping into", "filename", outName)
		var out *os.File
		out, err = os.OpenFile(outName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("cannot open file for stripping: %w", err)
		}
		defer out.Close()

		if err = toFile(f, out); err != nil {
			return err
		}

	}

	return nil
}
