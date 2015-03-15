This is an exif library to extract information from digital images.
The project is based on a Python version of this exif functionality.

The original Python library can be found at: [exif-py](http://sourceforge.net/projects/exif-py/)

# Installation #

The recommended current way is to use to install this library in $GOROOT with goinstall.
```
goinstall exif4go.googlecode.com/hg/exif4go
```
# How to use it #

Import the library in your code and call the method:

```go

func Process(f *os.File, debug bool) (map[string]*IfdTag, os.Error)
```

# TODOs #

  * The library has been testing only with jpeg images, the other formats might not be working properly yet, event if translated.