package binres

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

func apiJarPath() (string, error) {
	sdkdir := os.Getenv("ANDROID_HOME")
	if sdkdir == "" {
		return "", fmt.Errorf("ANDROID_HOME env var not set")
	}
	return path.Join(sdkdir, "platforms/android-15/android.jar"), nil
}

func apiResources() ([]byte, error) {
	p, err := apiJarPath()
	if err != nil {
		return nil, err
	}
	zr, err := zip.OpenReader(p)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	buf := new(bytes.Buffer)
	for _, f := range zr.File {
		if f.Name == "resources.arsc" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(buf, rc)
			if err != nil {
				return nil, err
			}
			rc.Close()
			break
		}
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("failed to read resources.arsc")
	}
	return buf.Bytes(), nil
}

// packResources produces a stripped down version of the resources.arsc from api jar.
func packResources() error {
	tbl, err := OpenSDKTable()
	if err != nil {
		return err
	}

	tbl.pool.strings = []string{} // should not be needed
	pkg := tbl.pkgs[0]

	// drop language string entries
	for _, typ := range pkg.specs[3].types {
		if typ.config.locale.language != 0 {
			for j, nt := range typ.entries {
				if nt == nil { // NoEntry
					continue
				}
				pkg.keyPool.strings[nt.key] = ""
				typ.indices[j] = NoEntry
				typ.entries[j] = nil
			}
		}
	}

	// drop strings from pool for specs to be dropped
	for _, spec := range pkg.specs[4:] {
		for _, typ := range spec.types {
			for _, nt := range typ.entries {
				if nt == nil { // NoEntry
					continue
				}
				// don't drop if there's a collision
				var collision bool
				for _, xspec := range pkg.specs[:4] {
					for _, xtyp := range xspec.types {
						for _, xnt := range xtyp.entries {
							if xnt == nil {
								continue
							}
							if collision = nt.key == xnt.key; collision {
								break
							}
						}
					}
				}
				if !collision {
					pkg.keyPool.strings[nt.key] = ""
				}
			}
		}
	}

	// entries are densely packed but probably safe to drop nil entries off the end
	for _, spec := range pkg.specs[:4] {
		for _, typ := range spec.types {
			var last int
			for i, nt := range typ.entries {
				if nt != nil {
					last = i
				}
			}
			typ.entries = typ.entries[:last+1]
			typ.indices = typ.indices[:last+1]
		}
	}

	// keeping 0:attr, 1:id, 2:style, 3:string
	pkg.typePool.strings = pkg.typePool.strings[:4]
	pkg.specs = pkg.specs[:4]

	bin, err := tbl.MarshalBinary()
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join("data", "packed.arsc.gz"))
	if err != nil {
		return err
	}
	defer f.Close()

	zw := gzip.NewWriter(f)
	defer zw.Close()

	if _, err := zw.Write(bin); err != nil {
		return err
	}
	return nil
}
