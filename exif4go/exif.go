package exif4go

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

var detailed bool
var debug bool

func setDebug(deb bool) {
	debug = deb
}

func writeInfo(a ...interface{}) (n int, err error) {
	if debug {
		if n, err = fmt.Println(a...); err != nil {
			print("Error printing")
		}
	}
	return
}

type StringSlice []string

func (p StringSlice) contains(s string) bool {
	for _, t := range p {
		if t == s {
			return true
		}
	}
	return false
}

type IntSlice []int

func (p IntSlice) contains(i int) bool {
	for _, t := range p {
		if t == i {
			return true
		}
	}
	return false
}

// Process process an images file calling the ProcessFile function with default parameters.
func Process(f *os.File, debug bool) (map[string]*IfdTag, error) {
	writeInfo("Processing using default parameters")
	return ProcessFile(f, "UNDEF", true, false, debug)
}

/* 
	ProcessFile processes an image file (expects an open file object). 
	This is the function that has to deal with all the arbitrary nasty bits of the EXIF standard.
*/
func ProcessFile(f *os.File, stop_tag string, details bool, strict bool, debug bool) (map[string]*IfdTag, error) {
	// yah it"s cheesy...
	if len(stop_tag) == 0 {
		stop_tag = "UNDEF"
	}
	//global detailed
	detailed = details

	// by default do not fake an EXIF beginning
	//fake_exif := 0

	// determine whether it"s a JPEG or TIFF
	data := make([]byte, 12)
	_, err := io.ReadAtLeast(f, data, 12)
	if err != nil {
		return nil, err
	}

	writeInfo("data has value:", data)

	s := StringSlice{"II*\x00", "MM\x00*"}
	var offset int64
	var endian []byte
	var fakeexif bool

	switch {
	case s.contains(string(data[0:4])):
		// it"s a TIFF file
		writeInfo("TIFF file")
		f.Seek(0, 0) // 0 relative to the beginning of the file
		endian = make([]byte, 1)
		if _, err := io.ReadAtLeast(f, endian, 1); err != nil {
			return nil, err
		}
		io.ReadAtLeast(f, endian, 1) // read again
		offset = 0
	case string(data[0:2]) == "\xFF\xD8":
		// it's a JPEG file
		writeInfo("JPEG file")
		s := StringSlice{"JFIF", "JFXX", "OLYM", "Phot"}

		token := string(data[6:10])
		for ; data[2] == 0xFF && s.contains(token); token = string(data[6:10]) {
			writeInfo("String token data[6:10]:", token)
			length := int(data[4])*256 + int(data[5])
			jump := make([]byte, length-8)
			//f.Read(jump) // advance
			//writeInfo("The string value of jump is:", string(jump))
			f.Seek(int64(length-8), 1) // advance relative to the current offset

			jump = make([]byte, 10)
			if _, err := f.Read(jump); err != nil {
				return nil, err
			}

			sj := string(jump)
			writeInfo("The string value of jump is:", sj)
			data = []byte("\xFF\x00" + sj)
			fakeexif = true
		}
		writeInfo("fakeexif:", fakeexif)

		if data[2] == 0xFF && string(data[6:10]) == "Exif" {
			//detected EXIF header
			writeInfo("detected EXIF header")
			offset, _ = f.Seek(0, 1)
			endian = make([]byte, 1)
			if _, err := f.Read(endian); err != nil {
				return nil, err
			}
		} else {
			// no EXIF information
			return nil, nil
		}
	default:
		// file format not recognized
		return nil, nil
	}
	// deal with the EXIF info we found
	writeInfo("The offset is:", offset, "\nThe endian value is:", string(endian), ", where 'I' => 'Intel', 'M' => 'Motorola'")

	hdr := newExifHeader(f, endian, offset, fakeexif, strict, debug)
	ifdlist, err := hdr.listIfds()
	if err != nil {
		return nil, err
	}
	writeInfo("The length of ifdlist is:", len(ifdlist))
	var ctr int
	for _, i := range ifdlist {
		var ifdname string
		var thumbifd int
		switch {
		case ctr == 0:
			ifdname = "Image"
		case ctr == 1:
			ifdname = "Thumbnail"
			thumbifd = i
		default:
			ifdname = fmt.Sprintf("IFD %d", ctr)
		}
		writeInfo(fmt.Sprintf(" IFD %d (%s) at offset %d:", ctr, ifdname, i))

		hdr.dumpIfd(i, ifdname, exifTags, 0, stop_tag)

		if exifoff, ok := hdr.tags[ifdname+" ExifOffset"]; ok {
			writeInfo(fmt.Sprintf(" EXIF SubIFD at offset %d:", exifoff.Values[0]))
			v, _ := strconv.Atoi(exifoff.Values[0])
			hdr.dumpIfd(v, "EXIF", exifTags, 0, stop_tag)

			// Interoperability IFD contained in EXIF IFD
			if introff, ok := hdr.tags["EXIF SubIFD InteroperabilityOffset"]; ok {
				writeInfo(fmt.Sprintf(" EXIF Interoperability SubSubIFD at offset %d:", introff.Values[0]))
				v, _ := strconv.Atoi(introff.Values[0])
				hdr.dumpIfd(v, "EXIF Interoperability", interTags, 0, stop_tag)
			}

		}

		// GPS IFD
		if gpsoff, ok := hdr.tags[ifdname+" GPSInfo"]; ok {
			writeInfo(fmt.Sprintf(" GPS SubIFD at offset %d:", gpsoff.Values[0]))
			v, _ := strconv.Atoi(gpsoff.Values[0])
			hdr.dumpIfd(v, "GPS", gpsTags, 0, stop_tag)
			//hdr.dump_IFD(gps_off.values[0], 'GPS', dict = GPS_TAGS, stop_tag = stop_tag)
		}
		writeInfo("thumbifd:", thumbifd)
		ctr += 1

	}

	// extract uncompressed TIFF thumbnail
	thumb, ok := hdr.tags["Thumbnail Compression"]
	if ok && thumb.Printable == "Uncompressed TIFF" {
		//hdr.extract_TIFF_thumbnail(thumb_ifd)
	}

	//JPEG thumbnail (thankfully the JPEG data is stored as a unit)
	if thumboff, ok := hdr.tags["Thumbnail JPEGInterchangeFormat"]; ok {
		j, _ := strconv.Atoi(thumboff.Values[0])
		f.Seek(offset+int64(j), 0)
		size, _ := strconv.Atoi(hdr.tags["Thumbnail JPEGInterchangeFormatLength"].Values[0])
		t := make([]byte, size)
		n, err := f.Read(t)
		if err != nil {
			return nil, err
		}
		writeInfo("Return number of bytes:", n)
		// hdr.tags["JPEGThumbnail"] = string(t) why this
	}

	// deal with MakerNote contained in EXIF IFD
	// (Some apps use MakerNote tags but do not use a format for which we
	// have a description, do not process these).

	_, ok1 := hdr.tags["EXIF MakerNote"]
	_, ok2 := hdr.tags["Image Make"]
	if ok1 && ok2 && detailed {
		//hdr.decode_maker_note()
	}

	// Sometimes in a TIFF file, a JPEG thumbnail is hidden in the MakerNote
	// since it's not allowed in a uncompressed TIFF IFD
	if _, ok := hdr.tags["JPEGThumbnail"]; !ok {

		if thumboff, ok := hdr.tags["MakerNote JPEGThumbnail"]; ok {
			j, _ := strconv.Atoi(thumboff.Values[0])
			f.Seek(offset+int64(j), 0)
			t := make([]byte, thumboff.fieldlength)
			n, err := f.Read(t)
			if err != nil {
				return nil, err
			}
			writeInfo("Return number of bytes:", n)
			// hdr.tags["JPEGThumbnail"] = t why this
		}
	}
	return hdr.tags, nil

}
