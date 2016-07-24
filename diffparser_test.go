// Copyright (c) 2015 Jesse Meek <https://github.com/waigani>
// This program is Free Software see LICENSE file for details.

package diffparser_test

import (
	"io/ioutil"
	"testing"

	"github.com/anmic/diffparser"
	jt "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type suite struct {
	jt.CleanupSuite
	rawdiff  string
	rawdiff2 string
	diff     *diffparser.Diff
}

var _ = gc.Suite(&suite{})

func (s *suite) SetUpSuite(c *gc.C) {
	byt, err := ioutil.ReadFile("example.diff")
	c.Assert(err, jc.ErrorIsNil)
	s.rawdiff = string(byt)
	byt, err = ioutil.ReadFile("example2.diff")
	c.Assert(err, jc.ErrorIsNil)
	s.rawdiff2 = string(byt)
}

// TODO(waigani) tests are missing more creative names (spaces, special
// chars), and diffed files that are not in the current directory.

func (s *suite) TestFileModeAndNaming(c *gc.C) {
	diff, err := diffparser.Parse(s.rawdiff)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(diff.Files, gc.HasLen, 6)

	for i, expected := range []struct {
		mode      diffparser.FileMode
		origName  string
		newName   string
		Additions int
		Deletions int
		Hash      string
	}{
		{
			mode:      diffparser.MODIFIED,
			origName:  "file1",
			newName:   "file1",
			Additions: 1,
			Deletions: 1,
			Hash:      "30f1681a9246bae4a64428a28e5e91136c5af6a6",
		},
		{
			mode:      diffparser.DELETED,
			origName:  "file2",
			newName:   "file2",
			Additions: 0,
			Deletions: 4,
			Hash:      "3dec22e1cea6483677dfa6a7f0e89f9f5f4ecb5d",
		},
		{
			mode:      diffparser.DELETED,
			origName:  "file3",
			newName:   "file3",
			Additions: 0,
			Deletions: 4,
			Hash:      "e2ffa21c1c3f03c9b001574e5176bbb9daa37b10",
		},
		{
			mode:      diffparser.NEW,
			origName:  "file4",
			newName:   "file4",
			Additions: 1,
			Deletions: 0,
			Hash:      "89bf72b8516fcf5bf50d82963128329fc4ad32da",
		},
		{
			mode:      diffparser.NEW,
			origName:  "newname",
			newName:   "newname",
			Additions: 4,
			Deletions: 0,
			Hash:      "b238bbc90ba9d102974b4470822b8e5f2da006b5",
		},
		{
			mode:      diffparser.MODIFIED,
			origName:  "static/img/background/Image/image 9.jpg",
			newName:   "static/img/background/Image/image 9.jpg",
			Additions: 0,
			Deletions: 0,
			Hash:      "2cc10b8e8c524a719e809ff2a5e565f8181cd1e0",
		},
	} {
		file := diff.Files[i]
		c.Logf("testing file: %v", file)
		c.Assert(file.Mode, gc.Equals, expected.mode)
		c.Assert(file.OrigName, gc.Equals, expected.origName)
		c.Assert(file.NewName, gc.Equals, expected.newName)
		c.Assert(file.Additions, gc.Equals, expected.Additions)
		c.Assert(file.Deletions, gc.Equals, expected.Deletions)
		c.Assert(file.Hash, gc.Equals, expected.Hash)
	}
}

func (s *suite) TestHunk(c *gc.C) {
	diff, err := diffparser.Parse(s.rawdiff)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(diff.Files, gc.HasLen, 6)

	expectedOrigLines := []diffparser.DiffLine{
		{
			Mode:     diffparser.UNCHANGED,
			Number:   1,
			Content:  "some",
			Position: 2,
		}, {
			Mode:     diffparser.UNCHANGED,
			Number:   2,
			Content:  "lines",
			Position: 3,
		}, {
			Mode:     diffparser.REMOVED,
			Number:   3,
			Content:  "in",
			Position: 4,
		}, {
			Mode:     diffparser.UNCHANGED,
			Number:   4,
			Content:  "file1",
			Position: 5,
		},
	}

	expectedNewLines := []diffparser.DiffLine{
		{
			Mode:     diffparser.ADDED,
			Number:   1,
			Content:  "add a line",
			Position: 1,
		}, {
			Mode:     diffparser.UNCHANGED,
			Number:   2,
			Content:  "some",
			Position: 2,
		}, {
			Mode:     diffparser.UNCHANGED,
			Number:   3,
			Content:  "lines",
			Position: 3,
		}, {
			Mode:     diffparser.UNCHANGED,
			Number:   4,
			Content:  "file1",
			Position: 5,
		},
	}

	file := diff.Files[0]
	origRange := file.Hunks[0].OrigRange
	newRange := file.Hunks[0].NewRange

	c.Assert(origRange.Start, gc.Equals, 1)
	c.Assert(origRange.Length, gc.Equals, 4)
	c.Assert(newRange.Start, gc.Equals, 1)
	c.Assert(newRange.Length, gc.Equals, 4)

	for i, line := range expectedOrigLines {
		c.Assert(*origRange.Lines[i], gc.DeepEquals, line)
	}
	for i, line := range expectedNewLines {
		c.Assert(*newRange.Lines[i], gc.DeepEquals, line)
	}
}

func (s *suite) TestParseHeader(c *gc.C) {
	diff, err := diffparser.Parse(s.rawdiff2)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(diff.Files, gc.HasLen, 1)

	c.Assert(diff.Files[0].OrigName, gc.Equals, "includes/s_header.php")
	c.Assert(diff.Files[0].NewName, gc.Equals, "includes/s_header.php")
	c.Assert(diff.Files[0].Mode, gc.Equals, diffparser.DELETED)
}

func (s *suite) TestEmptyDiff(c *gc.C) {
	diff, err := diffparser.Parse("")

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(diff.Files, gc.HasLen, 0)
}

func (s *suite) TestCorruptDiff(c *gc.C) {
	diff, err := diffparser.Parse("diff --git a/file1")

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(diff.Files, gc.HasLen, 0)
}
