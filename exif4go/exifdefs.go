package exif4go

// makestring removes non-printing characters from a string. 
// It does not throw an exception when given an out of range character.
func makestring(values []string) string {
	if len(values) == 0 {
		return ""
	}
	seq := []byte(values[0])
	out := []byte{}
	for _, c := range seq {
		// Screen out non-printing characters
		if 32 <= c && int(c) < 256 {
			out = append(out, c)
		}
	}
	// If no printing chars
	if len(out) == 0 {
		return string(seq)
	}
	return string(out)
}

// makestringuc is a special version of makestring function to deal with the code in the first 8 bytes of a user comment.
// First 8 bytes gives coding system e.g. ASCII vs. JIS vs Unicode.
func makestringuc(values []string) string {
	if len(values) == 0 {
		return ""
	}
	seq0 := []byte(values[0])
	//code := seq[0:8]
	str := ""
	if len(seq0) >= 8 {
		str = string([]byte(values[0])[8:])
		//writeInfo("The value of str in makestringuc is", str)
	}

	// Of course, this is only correct if ASCII, and the standard explicitly
	// allows JIS and Unicode.
	return makestring([]string{str})
}

type FieldType struct {
	Size uint
	Code string
	Name string
}

// field type descriptions as (length, abbreviation, full name) tuples.
var FIELD_TYPES = []*FieldType{
	&FieldType{0, "X", "Proprietary"}, // no such type
	&FieldType{1, "B", "Byte"},
	&FieldType{1, "A", "ASCII"},
	&FieldType{2, "S", "Short"},
	&FieldType{4, "L", "Long"},
	&FieldType{8, "R", "Ratio"},
	&FieldType{1, "SB", "Signed Byte"},
	&FieldType{1, "U", "Undefined"},
	&FieldType{2, "SS", "Signed Short"},
	&FieldType{4, "SL", "Signed Long"},
	&FieldType{8, "SR", "Signed Ratio"},
}

type exifTag struct {
	name     string
	fields   map[int]string
	function func([]string) string
}

/* 
Map of main EXIF tag names. 
The first element of tuple is the tag name, 
the second optional element is another dictionary giving names to values

TODO:
	Replace make_string_uc and make_string with something meaningful.

*/
var exifTags = map[int]*exifTag{
	0x0100: &exifTag{"ImageWidth", nil, nil},

	0x0101: &exifTag{"ImageLength", nil, nil},
	0x0102: &exifTag{"BitsPerSample", nil, nil},
	0x0103: &exifTag{"Compression",
		map[int]string{
			1:     "Uncompressed",
			2:     "CCITT 1D",
			3:     "T4/Group 3 Fax",
			4:     "T6/Group 4 Fax",
			5:     "LZW",
			6:     "JPEG (old-style)",
			7:     "JPEG",
			8:     "Adobe Deflate",
			9:     "JBIG B&W",
			10:    "JBIG Color",
			32766: "Next",
			32769: "Epson ERF Compressed",
			32771: "CCIRLEW",
			32773: "PackBits",
			32809: "Thunderscan",
			32895: "IT8CTPAD",
			32896: "IT8LW",
			32897: "IT8MP",
			32898: "IT8BL",
			32908: "PixarFilm",
			32909: "PixarLog",
			32946: "Deflate",
			32947: "DCS",
			34661: "JBIG",
			34676: "SGILog",
			34677: "SGILog24",
			34712: "JPEG 2000",
			34713: "Nikon NEF Compressed",
			65000: "Kodak DCR Compressed",
			65535: "Pentax PEF Compressed"}, nil},
	0x0106: &exifTag{"PhotometricInterpretation", nil, nil},
	0x0107: &exifTag{"Thresholding", nil, nil},
	0x010A: &exifTag{"FillOrder", nil, nil},
	0x010D: &exifTag{"DocumentName", nil, nil},
	0x010E: &exifTag{"ImageDescription", nil, nil},
	0x010F: &exifTag{"Make", nil, nil},
	0x0110: &exifTag{"Model", nil, nil},
	0x0111: &exifTag{"StripOffsets", nil, nil},
	0x0112: &exifTag{"Orientation",
		map[int]string{
			1: "Horizontal (normal)",
			2: "Mirrored horizontal",
			3: "Rotated 180",
			4: "Mirrored vertical",
			5: "Mirrored horizontal then rotated 90 CCW",
			6: "Rotated 90 CW",
			7: "Mirrored horizontal then rotated 90 CW",
			8: "Rotated 90 CCW"}, nil},
	0x0115: &exifTag{"SamplesPerPixel", nil, nil},
	0x0116: &exifTag{"RowsPerStrip", nil, nil},
	0x0117: &exifTag{"StripByteCounts", nil, nil},
	0x011A: &exifTag{"XResolution", nil, nil},
	0x011B: &exifTag{"YResolution", nil, nil},
	0x011C: &exifTag{"PlanarConfiguration", nil, nil},
	0x011D: &exifTag{"PageName", nil, makestring},
	0x0128: &exifTag{"ResolutionUnit",
		map[int]string{
			1: "Not Absolute",
			2: "Pixels/Inch",
			3: "Pixels/Centimeter"}, nil},
	0x012D: &exifTag{"TransferFunction", nil, nil},
	0x0131: &exifTag{"Software", nil, nil},
	0x0132: &exifTag{"DateTime", nil, nil},
	0x013B: &exifTag{"Artist", nil, nil},
	0x013E: &exifTag{"WhitePoint", nil, nil},
	0x013F: &exifTag{"PrimaryChromaticities", nil, nil},
	0x0156: &exifTag{"TransferRange", nil, nil},
	0x0200: &exifTag{"JPEGProc", nil, nil},
	0x0201: &exifTag{"JPEGInterchangeFormat", nil, nil},
	0x0202: &exifTag{"JPEGInterchangeFormatLength", nil, nil},
	0x0211: &exifTag{"YCbCrCoefficients", nil, nil},
	0x0212: &exifTag{"YCbCrSubSampling", nil, nil},
	0x0213: &exifTag{"YCbCrPositioning",
		map[int]string{
			1: "Centered",
			2: "Co-sited"}, nil},
	0x0214: &exifTag{"ReferenceBlackWhite", nil, nil},

	0x4746: &exifTag{"Rating", nil, nil},

	0x828D: &exifTag{"CFARepeatPatternDim", nil, nil},
	0x828E: &exifTag{"CFAPattern", nil, nil},
	0x828F: &exifTag{"BatteryLevel", nil, nil},
	0x8298: &exifTag{"Copyright", nil, nil},
	0x829A: &exifTag{"ExposureTime", nil, nil},
	0x829D: &exifTag{"FNumber", nil, nil},
	0x83BB: &exifTag{"IPTC/NAA", nil, nil},
	0x8769: &exifTag{"ExifOffset", nil, nil},
	0x8773: &exifTag{"InterColorProfile", nil, nil},
	0x8822: &exifTag{"ExposureProgram",
		map[int]string{
			0: "Unidentified",
			1: "Manual",
			2: "Program Normal",
			3: "Aperture Priority",
			4: "Shutter Priority",
			5: "Program Creative",
			6: "Program Action",
			7: "Portrait Mode",
			8: "Landscape Mode"}, nil},
	0x8824: &exifTag{"SpectralSensitivity", nil, nil},
	0x8825: &exifTag{"GPSInfo", nil, nil},
	0x8827: &exifTag{"ISOSpeedRatings", nil, nil},
	0x8828: &exifTag{"OECF", nil, nil},
	0x9000: &exifTag{"ExifVersion", nil, makestring},
	0x9003: &exifTag{"DateTimeOriginal", nil, nil},
	0x9004: &exifTag{"DateTimeDigitized", nil, nil},
	0x9101: &exifTag{"ComponentsConfiguration",
		map[int]string{
			0: "",
			1: "Y",
			2: "Cb",
			3: "Cr",
			4: "Red",
			5: "Green",
			6: "Blue"}, nil},
	0x9102: &exifTag{"CompressedBitsPerPixel", nil, nil},
	0x9201: &exifTag{"ShutterSpeedValue", nil, nil},
	0x9202: &exifTag{"ApertureValue", nil, nil},
	0x9203: &exifTag{"BrightnessValue", nil, nil},
	0x9204: &exifTag{"ExposureBiasValue", nil, nil},
	0x9205: &exifTag{"MaxApertureValue", nil, nil},
	0x9206: &exifTag{"SubjectDistance", nil, nil},
	0x9207: &exifTag{"MeteringMode",
		map[int]string{
			0: "Unidentified",
			1: "Average",
			2: "CenterWeightedAverage",
			3: "Spot",
			4: "MultiSpot",
			5: "Pattern"}, nil},
	0x9208: &exifTag{"LightSource",
		map[int]string{
			0:   "Unknown",
			1:   "Daylight",
			2:   "Fluorescent",
			3:   "Tungsten",
			9:   "Fine Weather",
			10:  "Flash",
			11:  "Shade",
			12:  "Daylight Fluorescent",
			13:  "Day White Fluorescent",
			14:  "Cool White Fluorescent",
			15:  "White Fluorescent",
			17:  "Standard Light A",
			18:  "Standard Light B",
			19:  "Standard Light C",
			20:  "D55",
			21:  "D65",
			22:  "D75",
			255: "Other"}, nil},
	0x9209: &exifTag{"Flash",
		map[int]string{
			0:  "No",
			1:  "Fired",
			5:  "Fired (?)", // no return sensed
			7:  "Fired (!)", // return sensed
			9:  "Fill Fired",
			13: "Fill Fired (?)",
			15: "Fill Fired (!)",
			16: "Off",
			24: "Auto Off",
			25: "Auto Fired",
			29: "Auto Fired (?)",
			31: "Auto Fired (!)",
			32: "Not Available"}, nil},
	0x920A: &exifTag{"FocalLength", nil, nil},
	0x9214: &exifTag{"SubjectArea", nil, nil},
	0x927C: &exifTag{"MakerNote", nil, nil},
	0x9286: &exifTag{"UserComment", nil, makestringuc},
	0x9290: &exifTag{"SubSecTime", nil, nil},
	0x9291: &exifTag{"SubSecTimeOriginal", nil, nil},
	0x9292: &exifTag{"SubSecTimeDigitized", nil, nil},

	// used by Windows Explorer
	0x9C9B: &exifTag{"XPTitle", nil, nil},
	0x9C9C: &exifTag{"XPComment", nil, nil},
	0x9C9D: &exifTag{"XPAuthor", nil, nil}, //(ignored by Windows Explorer if Artist exists)
	0x9C9E: &exifTag{"XPKeywords", nil, nil},
	0x9C9F: &exifTag{"XPSubject", nil, nil},

	0xA000: &exifTag{"FlashPixVersion", nil, makestring},
	0xA001: &exifTag{"ColorSpace",
		map[int]string{
			1:     "sRGB",
			2:     "Adobe RGB",
			65535: "Uncalibrated"}, nil},
	0xA002: &exifTag{"ExifImageWidth", nil, nil},
	0xA003: &exifTag{"ExifImageLength", nil, nil},
	0xA005: &exifTag{"InteroperabilityOffset", nil, nil},
	0xA20B: &exifTag{"FlashEnergy", nil, nil},              // 0x920B in TIFF/EP
	0xA20C: &exifTag{"SpatialFrequencyResponse", nil, nil}, // 0x920C
	0xA20E: &exifTag{"FocalPlaneXResolution", nil, nil},    // 0x920E
	0xA20F: &exifTag{"FocalPlaneYResolution", nil, nil},    // 0x920F
	0xA210: &exifTag{"FocalPlaneResolutionUnit", nil, nil}, // 0x9210
	0xA214: &exifTag{"SubjectLocation", nil, nil},          // 0x9214
	0xA215: &exifTag{"ExposureIndex", nil, nil},            // 0x9215
	0xA217: &exifTag{"SensingMethod", // 0x9217
		map[int]string{
			1: "Not defined",
			2: "One-chip color area",
			3: "Two-chip color area",
			4: "Three-chip color area",
			5: "Color sequential area",
			7: "Trilinear",
			8: "Color sequential linear"}, nil},
	0xA300: &exifTag{"FileSource",
		map[int]string{
			1: "Film Scanner",
			2: "Reflection Print Scanner",
			3: "Digital Camera"}, nil},
	0xA301: &exifTag{"SceneType",
		map[int]string{
			1: "Directly Photographed"}, nil},
	0xA302: &exifTag{"CVAPattern", nil, nil},
	0xA401: &exifTag{"CustomRendered",
		map[int]string{
			0: "Normal",
			1: "Custom"}, nil},
	0xA402: &exifTag{"ExposureMode",
		map[int]string{
			0: "Auto Exposure",
			1: "Manual Exposure",
			2: "Auto Bracket"}, nil},
	0xA403: &exifTag{"WhiteBalance",
		map[int]string{
			0: "Auto",
			1: "Manual"}, nil},
	0xA404: &exifTag{"DigitalZoomRatio", nil, nil},
	0xA405: &exifTag{"FocalLengthIn35mmFilm", nil, nil},
	0xA406: &exifTag{"SceneCaptureType",
		map[int]string{
			0: "Standard",
			1: "Landscape",
			2: "Portrait",
			3: "Night)"}, nil},
	0xA407: &exifTag{"GainControl",
		map[int]string{
			0: "None",
			1: "Low gain up",
			2: "High gain up",
			3: "Low gain down",
			4: "High gain down"}, nil},
	0xA408: &exifTag{"Contrast",
		map[int]string{
			0: "Normal",
			1: "Soft",
			2: "Hard"}, nil},
	0xA409: &exifTag{"Saturation",
		map[int]string{
			0: "Normal",
			1: "Soft",
			2: "Hard"}, nil},
	0xA40A: &exifTag{"Sharpness",
		map[int]string{
			0: "Normal",
			1: "Soft",
			2: "Hard"}, nil},
	0xA40B: &exifTag{"DeviceSettingDescription", nil, nil},
	0xA40C: &exifTag{"SubjectDistanceRange", nil, nil},
	0xA500: &exifTag{"Gamma", nil, nil},
	0xC4A5: &exifTag{"PrintIM", nil, nil},
	0xEA1C: &exifTag{"Padding", nil, nil},
}

// interoperability tags.
var interTags = map[int]*exifTag{
	0x0001: &exifTag{"InteroperabilityIndex", nil, nil},
	0x0002: &exifTag{"InteroperabilityVersion", nil, nil},
	0x1000: &exifTag{"RelatedImageFileFormat", nil, nil},
	0x1001: &exifTag{"RelatedImageWidth", nil, nil},
	0x1002: &exifTag{"RelatedImageLength", nil, nil},
}

// GPS tags.
var gpsTags = map[int]*exifTag{
	0x0000: &exifTag{"GPSVersionID", nil, nil},
	0x0001: &exifTag{"GPSLatitudeRef", nil, nil},
	0x0002: &exifTag{"GPSLatitude", nil, nil},
	0x0003: &exifTag{"GPSLongitudeRef", nil, nil},
	0x0004: &exifTag{"GPSLongitude", nil, nil},
	0x0005: &exifTag{"GPSAltitudeRef", nil, nil},
	0x0006: &exifTag{"GPSAltitude", nil, nil},
	0x0007: &exifTag{"GPSTimeStamp", nil, nil},
	0x0008: &exifTag{"GPSSatellites", nil, nil},
	0x0009: &exifTag{"GPSStatus", nil, nil},
	0x000A: &exifTag{"GPSMeasureMode", nil, nil},
	0x000B: &exifTag{"GPSDOP", nil, nil},
	0x000C: &exifTag{"GPSSpeedRef", nil, nil},
	0x000D: &exifTag{"GPSSpeed", nil, nil},
	0x000E: &exifTag{"GPSTrackRef", nil, nil},
	0x000F: &exifTag{"GPSTrack", nil, nil},
	0x0010: &exifTag{"GPSImgDirectionRef", nil, nil},
	0x0011: &exifTag{"GPSImgDirection", nil, nil},
	0x0012: &exifTag{"GPSMapDatum", nil, nil},
	0x0013: &exifTag{"GPSDestLatitudeRef", nil, nil},
	0x0014: &exifTag{"GPSDestLatitude", nil, nil},
	0x0015: &exifTag{"GPSDestLongitudeRef", nil, nil},
	0x0016: &exifTag{"GPSDestLongitude", nil, nil},
	0x0017: &exifTag{"GPSDestBearingRef", nil, nil},
	0x0018: &exifTag{"GPSDestBearing", nil, nil},
	0x0019: &exifTag{"GPSDestDistanceRef", nil, nil},
	0x001A: &exifTag{"GPSDestDistance", nil, nil},
	0x001D: &exifTag{"GPSDate", nil, nil},
}

// Ignore these tags when quick processing.
// - 0x927C is MakerNote Tags
// - 0x9286 is a user comment
var ignoreTags = []int{0x9286, 0x927C}

// Canon tags
var makerNoteCanonTags = map[int]*exifTag{
	0x0006: &exifTag{"ImageType", nil, nil},
	0x0007: &exifTag{"FirmwareVersion", nil, nil},
	0x0008: &exifTag{"ImageNumber", nil, nil},
	0x0009: &exifTag{"OwnerName", nil, nil},
}

// This is in element offset, name, optional value dictionary format.
var makerNoteCanonTags_0x001 = map[int]*exifTag{
	1: &exifTag{"Macromode",
		map[int]string{
			1: "Macro",
			2: "Normal"}, nil},
	2: &exifTag{"SelfTimer", nil, nil},
	3: &exifTag{"Quality",
		map[int]string{
			2: "Normal",
			3: "Fine",
			5: "Superfine"}, nil},
	4: &exifTag{"FlashMode",
		map[int]string{
			0:  "Flash Not Fired",
			1:  "Auto",
			2:  "On",
			3:  "Red-Eye Reduction",
			4:  "Slow Synchro",
			5:  "Auto + Red-Eye Reduction",
			6:  "On + Red-Eye Reduction",
			16: "external flash"}, nil},
	5: &exifTag{"ContinuousDriveMode",
		map[int]string{
			0: "Single Or Timer",
			1: "Continuous"}, nil},
	7: &exifTag{"FocusMode",
		map[int]string{
			0: "One-Shot",
			1: "AI Servo",
			2: "AI Focus",
			3: "MF",
			4: "Single",
			5: "Continuous",
			6: "MF"}, nil},
	10: &exifTag{"ImageSize",
		map[int]string{
			0: "Large",
			1: "Medium",
			2: "Small"}, nil},
	11: &exifTag{"EasyShootingMode",
		map[int]string{
			0:  "Full Auto",
			1:  "Manual",
			2:  "Landscape",
			3:  "Fast Shutter",
			4:  "Slow Shutter",
			5:  "Night",
			6:  "B&W",
			7:  "Sepia",
			8:  "Portrait",
			9:  "Sports",
			10: "Macro/Close-Up",
			11: "Pan Focus"}, nil},
	12: &exifTag{"DigitalZoom",
		map[int]string{
			0: "None",
			1: "2x",
			2: "4x"}, nil},
	13: &exifTag{"Contrast",
		map[int]string{
			0xFFFF: "Low",
			0:      "Normal",
			1:      "High"}, nil},
	14: &exifTag{"Saturation",
		map[int]string{
			0xFFFF: "Low",
			0:      "Normal",
			1:      "High"}, nil},
	15: &exifTag{"Sharpness",
		map[int]string{
			0xFFFF: "Low",
			0:      "Normal",
			1:      "High"}, nil},
	16: &exifTag{"ISO",
		map[int]string{
			0:  "See ISOSpeedRatings Tag",
			15: "Auto",
			16: "50",
			17: "100",
			18: "200",
			19: "400"}, nil},
	17: &exifTag{"MeteringMode",
		map[int]string{
			3: "Evaluative",
			4: "Partial",
			5: "Center-weighted"}, nil},
	18: &exifTag{"FocusType",
		map[int]string{
			0: "Manual",
			1: "Auto",
			3: "Close-Up (Macro)",
			8: "Locked (Pan Mode)"}, nil},
	19: &exifTag{"AFPointSelected",
		map[int]string{
			0x3000: "None (MF)",
			0x3001: "Auto-Selected",
			0x3002: "Right",
			0x3003: "Center",
			0x3004: "Left"}, nil},
	20: &exifTag{"ExposureMode",
		map[int]string{
			0: "Easy Shooting",
			1: "Program",
			2: "Tv-priority",
			3: "Av-priority",
			4: "Manual",
			5: "A-DEP"}, nil},
	23: &exifTag{"LongFocalLengthOfLensInFocalUnits", nil, nil},
	24: &exifTag{"ShortFocalLengthOfLensInFocalUnits", nil, nil},
	25: &exifTag{"FocalUnitsPerMM", nil, nil},
	28: &exifTag{"FlashActivity",
		map[int]string{
			0: "Did Not Fire",
			1: "Fired"}, nil},
	29: &exifTag{"FlashDetails",
		map[int]string{
			14: "External E-TTL",
			13: "Internal Flash",
			11: "FP Sync Used",
			7:  "2nd&exifTag{\"Rear\")-Curtain Sync Used",
			4:  "FP Sync Enabled"}, nil},
	32: &exifTag{"FocusMode",
		map[int]string{
			0: "Single",
			1: "Continuous"}, nil},
}

var makerNoteCanonTags_0x004 = map[int]*exifTag{
	7: &exifTag{"WhiteBalance",
		map[int]string{
			0: "Auto",
			1: "Sunny",
			2: "Cloudy",
			3: "Tungsten",
			4: "Fluorescent",
			5: "Flash",
			6: "Custom"}, nil},
	9:  &exifTag{"SequenceNumber", nil, nil},
	14: &exifTag{"AFPointUsed", nil, nil},
	15: &exifTag{"FlashBias",
		map[int]string{
			0xFFC0: "-2 EV",
			0xFFCC: "-1.67 EV",
			0xFFD0: "-1.50 EV",
			0xFFD4: "-1.33 EV",
			0xFFE0: "-1 EV",
			0xFFEC: "-0.67 EV",
			0xFFF0: "-0.50 EV",
			0xFFF4: "-0.33 EV",
			0x0000: "0 EV",
			0x000C: "0.33 EV",
			0x0010: "0.50 EV",
			0x0014: "0.67 EV",
			0x0020: "1 EV",
			0x002C: "1.33 EV",
			0x0030: "1.50 EV",
			0x0034: "1.67 EV",
			0x0040: "2 EV"}, nil},
	19: &exifTag{"SubjectDistance", nil, nil},
}
