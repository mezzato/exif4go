package exif4go

import (
	"testing"
	"os"
)

func TestProcessFile(t *testing.T) {
	//setDebug(true)
	fpath := "./test/test.jpg"
	f, err := os.Open(fpath)
	if err != nil {
		t.Fatal("Error opening file:", fpath)
	}
	var tags map[string]*IfdTag
	tags, err = Process(f, true)
	if err != nil {
		t.Fatalf("There was an error: %s", err)
	}
	t.Logf("The file name is %s", f.Name())

	if err != nil {
		t.Fatal("Error parsing via exif the file at:", fpath)
	}
	if len(tags) == 0 {
		t.Error("No tags found")
	}

	var checker = func(title string, key string, val string) bool {
		var ok bool = true
		var tag *IfdTag
		t.Log("Checking:", title)

		if tag, ok = tags[key]; !ok {
			t.Errorf("The key %s is missing", key)
		} else if len(tag.Values) > 0 && val != tag.Values[0] {
			t.Errorf("The data value %s is not the same as the expected value %s", tag.Values[0], val)
			ok = false
		}
		return ok
	}

	checker("Make", "Image Make", "Canon")
	checker("Model", "Image Model", "Canon EOS 1000D")
	checker("DateTime", "Image DateTime", "2010:11:28 16:42:18")
	checker("EXIF DateTimeOriginal", "EXIF DateTimeOriginal", "2010:11:28 16:42:18")

}
