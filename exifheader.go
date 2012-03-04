package exif4go

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// newRatio creates a ratio structure that eventually will be able to reduce itself to the lowest 
// common denominator for printing.
func newRatio(num int, den int) *ratio {
	return &ratio{num, den, false}
}

type ratio struct {
	num     int
	den     int
	reduced bool
}

func gcd(a int, b int) int {
	if b == 0 {
		return a
	}
	return gcd(b, a%b)
}

func (r *ratio) reduce() {
	div := gcd(r.num, r.den)
	if div > 1 {
		r.num = r.num / div
		r.den = r.den / div
	}

}

func (r *ratio) String() string {
	if !r.reduced {
		r.reduce()
		r.reduced = true
	}
	if r.den == 1 {
		return strconv.Itoa(r.num)
	}
	return fmt.Sprintf("%d/%d", r.num, r.den)
}

// IfdTag, used to deal with tags.
type IfdTag struct {
	// printable version of data
	Printable string
	// tag ID number
	tag int
	//field type as index into FIELD_TYPES
	Fieldtype int
	// offset of start of field in bytes from beginning of IFD
	fieldoffset int
	// length of data field in bytes
	fieldlength int
	// either a string or array of data items
	Values []string
}

func (t *IfdTag) String() string {
	return fmt.Sprintf("(0x%04X) %s=%s @ %d", t.tag,
		FIELD_TYPES[t.Fieldtype].Name,
		t.Printable,
		t.fieldoffset)
}

type exifHeader struct {
	file     *os.File
	endian   []byte
	offset   int64
	fakeExif bool
	strict   bool
	debug    bool
	tags     map[string]*IfdTag
}

func newExifHeader(file *os.File,
	endian []byte,
	offset int64,
	fakeExif bool,
	strict bool,
	debug bool) *exifHeader {
	tags := make(map[string]*IfdTag)
	hdr := &exifHeader{file, endian, offset, fakeExif, strict, debug, tags}
	return hdr
}

// Extract multibyte integer in Motorola format (little endian).
func (eh *exifHeader) s2n_motorola(str []byte) (val int) {
	var x int = 0
	for _, c := range str {
		x = (x << 8) | int(c)
	}
	return x

}

// Extract multibyte integer in Intel format (big endian)
func (eh *exifHeader) s2n_intel(str []byte) (val int) {
	var x int = 0
	var y byte = 0
	for _, c := range str {
		// NOTE: (int conversion is fundamental to avoid byte(1) << 8 being 0)
		x = (x | (int(c) << y))
		y = y + 8
	}

	return x
}

/*
   s2n convert a slice to an integer, based on the sign and endian flags.
   Usually this offset is assumed to be relative to the beginning of the
   start of the EXIF information. 
   For some cameras that use relative tags, this offset may be relative to some other starting point.
*/
func (eh *exifHeader) s2n(offset int, length uint, signed bool) (val int, err error) {

	eh.file.Seek(eh.offset+int64(offset), 0)

	s := make([]byte, length)

	if _, err1 := eh.file.Read(s); err1 != nil {
		return -1, err1
	}
	if eh.endian[0] == 'I' {
		val = int(eh.s2n_intel(s))

	} else {

		val = int(eh.s2n_motorola(s))
	}

	// Sign extension ?
	if signed {
		msb := 1 << (8*length - 1)
		if val&msb > 0 {
			val = val - (msb << 1)
		}
	}
	return val, err
}

// Convert offset to string.
func (eh *exifHeader) n2s(offset int, length int) string {
	s := ""
	for dummy := 0; dummy < length; dummy++ {
		if eh.endian[0] == 'I' {
			s = s + strconv.Itoa(offset&0xFF)
		} else {
			s = strconv.Itoa(offset&0xFF) + s
			offset = offset >> 8
		}
	}
	return s
}

// Return first IFD.
func (eh *exifHeader) firstIfd() (int, error) {
	return eh.s2n(4, 4, false)
}

// Return a pointer to next IFD.
func (eh *exifHeader) nextIfd(ifd int) (val int, err error) {
	entries, err := eh.s2n(ifd, 2, false)
	if err != nil {
		return val, err
	}
	val, err = eh.s2n(ifd+2+12*entries, 4, false)
	return
}

// Return a list of IFDs in the header.
func (eh *exifHeader) listIfds() (a []int, err error) {
	var i int
	if i, err = eh.firstIfd(); err != nil {
		return
	}
	//a = make([]int, 0)
	a = []int{}
	for i > 0 {
		a = append(a, i)
		if i, err = eh.nextIfd(i); err != nil {
			return nil, err
		}
	}
	return a, err
}

// Return list of entries in this IFD.
func (eh *exifHeader) dumpIfd(ifd int,
	ifdname string,
	dict map[int]*exifTag,
	relative int,
	stoptag string) (err error) {

	if dict == nil {
		dict = exifTags
	}
	if len(stoptag) == 0 {
		stoptag = "UNDEF"
	}

	entries, err := eh.s2n(ifd, 2, false)
	if err != nil {
		return err
	}

	for i := 0; i < entries; i++ {
		// entry is index of start of this IFD in the file
		entry := ifd + 2 + 12*i
		tag, err := eh.s2n(entry, 2, false)

		if err != nil {
			return err
		}

		var tagname string
		tagentry, ok := dict[tag]
		// get tag name early to avoid errors, help debug
		if ok {
			tagname = tagentry.name
			//writeInfo(fmt.Sprintf("Tag found, value %d and name %s",tag, tagname))
			//writeInfo("dumpIfd, tagname:", tagname)
		} else {
			tagname = fmt.Sprintf("Tag 0x%04X", tag)
			//writeInfo("Tag not in dictionary")
		}

		//writeInfo(fmt.Sprintf("entry no %d, tag %d, tagname %s", i, tag, tagname))
		// ignore certain tags for faster processing
		if detailed || !IntSlice(ignoreTags).contains(tag) {
			fieldtype, err := eh.s2n(entry+2, 2, false)
			if err != nil {
				return err
			}

			// unknown field type
			if !(0 < fieldtype && fieldtype < len(FIELD_TYPES)) {
				if !eh.strict {
					continue
				} else {
					return errors.New(fmt.Sprintf("unknown type %d in tag 0x%04X", fieldtype, tag))
				}
			}

			//writeInfo("field type:", fieldtype)
			typelen := FIELD_TYPES[fieldtype].Size
			count, err := eh.s2n(entry+4, 4, false)

			if err != nil {
				return err
			}
			// Adjust for tag id/type/count (2+2+4 bytes)
			// Now we point at either the data or the 2nd level offset
			offset := entry + 8

			// If the value fits in 4 bytes, it is inlined, else we
			// need to jump ahead again.
			if count*int(typelen) > 4 {
				// offset is not the value; it's a pointer to the value
				// if relative we set things up so s2n will seek to the right
				// place when it adds self.offset.  Note that this 'relative'
				// is for the Nikon type 3 makernote.  Other cameras may use
				// other relative offsets, which would have to be computed here
				// slightly differently.
				if relative > 0 {
					tmp_offset, err := eh.s2n(offset, 4, false)
					if err != nil {
						return err
					}
					offset = tmp_offset + ifd - 8
					if eh.fakeExif {
						offset = offset + 18
					}
				} else {
					offset, err = eh.s2n(offset, 4, false)
					if err != nil {
						return err
					}
				}
			}

			fieldoffset := offset
			values := []string{}
			if fieldtype == 2 {
				// special case: null-terminated ASCII string
				// XXX investigate
				// sometimes gets too big to fit in int value
				if count != 0 && count < (2^31) {
					eh.file.Seek(eh.offset+int64(offset), 0)
					vals := make([]byte, count)
					if _, err := eh.file.Read(vals); err != nil {
						return err
					}
					// Drop any garbage after a null.
					// IMPORTANT: return at most 2 substrings from Split, but NOT JUST 1
					cleanedVal := strings.SplitN(string(vals), "\x00", 2)[0]
					values = append(values, cleanedVal)

				} else {
					values = append(values, "")
				}
			} else {
				//values = ""
				is := IntSlice{6, 8, 9, 10}
				signed := is.contains(fieldtype)

				// XXX investigate
				// some entries get too big to handle could be malformed
				// file or problem with self.s2n
				if count < 1000 {
					for dummy := 0; dummy < count; dummy++ {
						var value string
						switch fieldtype {
						case 5, 10:
							// a ratio

							num, err := eh.s2n(offset, 4, signed)
							if err != nil {
								return err
							}
							den, err := eh.s2n(offset+4, 4, signed)
							if err != nil {
								return err
							}
							r := newRatio(num, den)
							value = r.String()
						default:
							v, err := eh.s2n(offset, typelen, signed)
							if err != nil {
								return err
							}
							value = strconv.Itoa(v)
						}

						values = append(values, value)
						offset = offset + int(typelen)
					}
					// The test above causes problems with tags that are 
					// supposed to have long values!  Fix up one important case.
				} else if tagname == "MakerNote" {
					for dummy := 0; dummy < count; dummy++ {
						v, err := eh.s2n(offset, typelen, signed)
						if err != nil {
							return err
						}
						value := strconv.Itoa(v)
						values = append(values, value)
						offset = offset + int(typelen)
					}
				}
				//else :
				//    print "Warning: dropping large tag:", tag, tag_name
			}

			var printable string
			// now 'values' is either a string or an array
			if count == 1 && fieldtype != 2 {
				printable = values[0]
			} else if count > 50 && len(values) > 20 {
				printable = "[" + strings.Join(values[0:20], ", ") + ", ... ]"
			} else {
				printable = strings.Join(values, ", ")
				if fieldtype == 2 { // ASCII
					printable = strconv.Quote(printable)
				}
			}

			// compute printable version of values
			if tagentry != nil {
				//writeInfo("Processing tag entry")

				// optional 2nd tag element is present
				if tagentry.function != nil {
					//if hasattr(tag_entry[1], '__call__'):
					// call mapping function
					printable = tagentry.function(values)
				} else if tagentry.fields != nil {
					printable = ""

					for _, i := range values {
						// use lookup table for this tag
						var ok bool
						var s string
						j, err := strconv.Atoi(i)
						if err != nil {
							return err
						}
						s, ok = tagentry.fields[j]

						if ok {
							printable += s
						} else {
							printable += i
						}

					}
				}

				k := ifdname + " " + tagname
				//writeInfo("Setting tag key:", k,"and offset",fieldoffset)
				eh.tags[k] = &IfdTag{printable, tag,
					fieldtype,
					fieldoffset,
					count * int(typelen),
					values}
				testval, _ := eh.tags[k]
				writeInfo(fmt.Sprintf(" DEBUG:   %s: %s.", tagname, testval))
			}
		}

		if tagname == stoptag {
			writeInfo("dumpIfd: breaking out of loop, reached stop tag:", stoptag)
			break
		}
	}
	return nil
}
