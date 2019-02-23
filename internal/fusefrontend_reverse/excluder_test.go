package fusefrontend_reverse

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/rfjakob/gocryptfs/internal/configfile"
	"github.com/rfjakob/gocryptfs/internal/fusefrontend"
	"github.com/rfjakob/gocryptfs/internal/nametransform"
)

func TestShouldNoCreateExcluderIfNoPattersWereSpecified(t *testing.T) {
	var rfs ReverseFS
	var args fusefrontend.Args
	rfs.prepareExcluder(args)
	if rfs.excluder != nil {
		t.Error("Should not have created excluder")
	}
}

func TestShouldPrefixExcludeValuesWithSlash(t *testing.T) {
	var args fusefrontend.Args
	args.Exclude = []string{"file1", "dir1/file2.txt"}
	args.ExcludeWildcard = []string{"*~", "build/*.o"}

	expected := []string{"/file1", "/dir1/file2.txt", "*~", "build/*.o"}

	patterns := getExclusionPatterns(args)
	if !reflect.DeepEqual(patterns, expected) {
		t.Errorf("expected %q, got %q", expected, patterns)
	}
}

func TestShouldReadExcludePatternsFromFiles(t *testing.T) {
	tmpfile1, err := ioutil.TempFile("", "excludetest")
	if err != nil {
		t.Fatal(err)
	}
	exclude1 := tmpfile1.Name()
	defer os.Remove(exclude1)
	defer tmpfile1.Close()

	tmpfile2, err := ioutil.TempFile("", "excludetest")
	if err != nil {
		t.Fatal(err)
	}
	exclude2 := tmpfile2.Name()
	defer os.Remove(exclude2)
	defer tmpfile2.Close()

	tmpfile1.WriteString("file1.1\n")
	tmpfile1.WriteString("file1.2\n")
	tmpfile2.WriteString("file2.1\n")
	tmpfile2.WriteString("file2.2\n")

	var args fusefrontend.Args
	args.ExcludeWildcard = []string{"cmdline1"}
	args.ExcludeFrom = []string{exclude1, exclude2}

	// An empty string is returned for the last empty line
	// It's ignored when the patterns are actually compiled
	expected := []string{"cmdline1", "file1.1", "file1.2", "", "file2.1", "file2.2", ""}

	patterns := getExclusionPatterns(args)
	if !reflect.DeepEqual(patterns, expected) {
		t.Errorf("expected %q, got %q", expected, patterns)
	}
}

type IgnoreParserMock struct {
	calledWith string
}

func (parser *IgnoreParserMock) MatchesPath(f string) bool {
	parser.calledWith = f
	return false
}

// Note: See also the integration tests in
// tests/reverse/exclude_test.go
func TestShouldNotCallIgnoreParserForTranslatedConfig(t *testing.T) {
	ignorerMock := &IgnoreParserMock{}
	var rfs ReverseFS
	rfs.excluder = ignorerMock

	if excluded, _, _ := rfs.isExcludedCipher(configfile.ConfDefaultName); excluded {
		t.Error("Should not exclude translated config")
	}
	if ignorerMock.calledWith != "" {
		t.Error("Should not call IgnoreParser for translated config")
	}
}

func TestShouldNotCallIgnoreParserForDirIV(t *testing.T) {
	ignorerMock := &IgnoreParserMock{}
	var rfs ReverseFS
	rfs.excluder = ignorerMock

	if excluded, _, _ := rfs.isExcludedCipher(nametransform.DirIVFilename); excluded {
		t.Error("Should not exclude DirIV")
	}
	if ignorerMock.calledWith != "" {
		t.Error("Should not call IgnoreParser for DirIV")
	}
}

func TestShouldReturnFalseIfThereAreNoExclusions(t *testing.T) {
	var rfs ReverseFS
	if rfs.isExcludedPlain("any/path") {
		t.Error("Should not exclude any path if no exclusions were specified")
	}
}

func TestShouldCallIgnoreParserToCheckExclusion(t *testing.T) {
	ignorerMock := &IgnoreParserMock{}
	var rfs ReverseFS
	rfs.excluder = ignorerMock

	rfs.isExcludedPlain("some/path")
	if ignorerMock.calledWith != "some/path" {
		t.Error("Failed to call IgnoreParser")
	}

}