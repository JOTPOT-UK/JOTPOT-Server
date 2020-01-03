package jps

import (
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/mediatype"
)

type MetadataWanted byte

const (
	//MetadataHasEffect bit will be set on a MetadataWanted value if that metadata value can be transfered by the protocol. If this bit is absent, then calling the set method will have no effect, and the get method will never return a value.
	MetadataHasEffect MetadataWanted = 1 << 0
	//MetadataRecomended bit will be set on a MetadataWanted value if the implementer of the metadata setter stringly recomends that you set that metadata proerty if possible.
	MetadataRecomended MetadataWanted = 1 << 1
	//MetadataRequired bits will be set on a MetadataWanted value if the protocol requires that metadata to be sent.
	MetadataRequired MetadataWanted = 1<<2 | MetadataRecomended
	//MetadataNoDefault bit will be set on a MetadataWanted value if the metadata is required AND the implementation of the MetadataSetter does not have a default value. If this is present with MetadataRequired, then the value must be set explicitly.
	MetadataNoDefault MetadataWanted = 1 << 3
)

type FileType uint8

const (
	FileTypeFile = FileType(iota)
	FileTypeDir
	FileTypeLink
	FileTypeOther
)

//MetadataGetter provides an interface to get standard metadata types from bodies.
type MetadataGetter interface {
	//Location gets the location value - see MetadataSetter.SetLocation
	Location() (string, error)
	//Size gets the size of the file requested - see MetadataSetter.Size
	Size() (int64, error)
	Range() (Range, error)
	Type() (*mediatype.Type, error)
	TypeString() (string, error)
	Encodings() ([]string, error)
	Languages() ([]string, error)
	MTime() (time.Time, error)
	ETag() (string, bool, error)
	//TODO: FileType() FileType
	//TODO: Permissions?
}

type MetadataSetter interface {
	WantLocation() MetadataWanted
	SetLocation(string) error
	WantSize() MetadataWanted
	//SetSize sets the size of the body (this is the complete size, if the whole resource was sent)
	// A value of -1 indicates an unknown size. This should be the default size.
	SetSize(int64) error
	WantRange() MetadataWanted
	//SetRange sets the range of the whole data which this body contains.
	//start is the byte offset in the file which this body starts at; 0 is the first byte of the file.
	//end is the byte offset in the whole file which corresponds to the last byte sent.
	//For example, if the file being sent has a size of 10 bytes, then SetRange(0, 9) would be thr range of the entire file.
	//setting start to -1 indicates the end of the file.
	SetRange(Range) error
	WantType() MetadataWanted
	SetType(*mediatype.Type) error
	SetTypeString(string) error
	WantEncodings() MetadataWanted
	SetEncodings([]string) error
	WantLanguages() MetadataWanted
	SetLanguages([]string) error
	WantMTime() MetadataWanted
	SetMTime(time.Time) error
	WantETag() MetadataWanted
	SetETag(etag string, strong bool) error
}

type Metadata interface {
	MetadataGetter
	MetadataSetter
}

//IncomingBody contains a Reader to read the body from and a MetadataGetter of properties about this body.
type IncomingBody struct {
	Reader
	MetadataGetter
}

//OutgoingBody contains a Writer to write the body to and a Metadata interface to set and get properties about the body.
type OutgoingBody struct {
	Writer
	Metadata
}
