package compare

import (
	"fmt"
	"github.com/cavaliergopher/cpio"
	"github.com/r3labs/diff/v3"
	"io"
	"log"
	"os"
	"strings"
)

func CompareHeaders(h1 *cpio.Header, h2 *cpio.Header) error {
	if h1.Name != h2.Name {
		return fmt.Errorf("name mismatch, %s != %s", h1.Name, h2.Name)
	}

	c, err := diff.Diff(h1, h2)
	if err != nil {
		return err
	}

	for _, cc := range c {
		p := strings.Join(cc.Path, ",")
		fmt.Printf("file: %s\t%s %s:%d -> %d\t\n", h1.Name, cc.Type, p, cc.From, cc.To)
	}

	return nil
}

func Compare(filename1 string, filename2 string) error {
	names := []string{
		filename1,
		filename2,
	}

	headers := map[string][]*cpio.Header{}
	for _, name := range names {
		f, err := os.Open(name)
		if err != nil {
			return fmt.Errorf("cannot open file: %w", err)
		}

		r := cpio.NewReader(f)

		// Iterate through the files in the archive.
		for {
			hdr, err := r.Next()
			if err == io.EOF {
				// end of cpio archive
				break
			}
			if err != nil {
				log.Fatalln(err)
			}

			headers[name] = append(headers[name], hdr)
		}

		f.Close()
	}

	file2 := headers[names[1]]
	for n, hdr := range headers[names[0]] {
		if n > len(file2) {
			fmt.Println("WRONG FILE")
			continue
		}

		h2 := file2[n]

		if err := CompareHeaders(hdr, h2); err != nil {
			return err
		}
	}

	return nil
}
