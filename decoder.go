package drum

import (
    "os"
    "fmt"
    "encoding/binary"
    "bytes"
    "errors"
    "strings"
    "math"
)

func readString(r *bytes.Reader) string {
    var b byte
    var err error
    result := make([]byte,0)
    for {
        b, err = r.ReadByte()
        if err != nil || b==0 {
            break
        }
        result = append(result,b)
    }
    return string(result)
}

func readStringLength(r *bytes.Reader, length uint8) string {
    result := make([]byte, length)
    r.Read(result)
    return string(result)
}

func skipNull(r *bytes.Reader) error{
    var b byte
    var err error
    for {
        b,err=r.ReadByte()
        if err != nil {
            return err
        }
        if b != 0 {
            r.UnreadByte()
            return nil
        }
    }
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	p := NewPattern()

        f, err := os.Open(path)
        if err != nil {
            return nil, err
        }
        defer f.Close()

        buf := make([]byte, 1024)

        bytes_read, err := f.Read(buf)
        if err != nil {
            return nil, err
        }

        data := bytes.NewReader(buf[:bytes_read])

        //We read in the header
        if readString(data) != "SPLICE" {
            return nil, errors.New(fmt.Sprintf("File %s does not have a correct header!",path))
        }
        skipNull(data)

        //Read the version
        data.ReadByte()
        p.Version = readString(data)

        //We read in the tempo from offset 46 as a float32
        data.Seek(46,0)
        err = binary.Read(data,binary.LittleEndian,&p.Tempo)
        if err !=nil {
            fmt.Printf(err.Error())
        }

        var namelen uint8
        for err==nil {
            track := NewTrack()
            err = binary.Read(data,binary.LittleEndian,&track.Id)

            if (track.Id > 255) {
                break
            }
            if err != nil {
                break
            }

            err = binary.Read(data,binary.LittleEndian,&namelen)
            track.Name=readStringLength(data,namelen)
            bytes_read, err = data.Read(track.Steps)

            p.Tracks = append(p.Tracks,track)
        }

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct{
    Version string
    Tempo float32
    Tracks []*Track
}

func NewPattern() *Pattern {
    p := new(Pattern)
    p.Tracks = make([]*Track,0)
    return p
}

func format_float(f float32) string {
    ff := float64(f)
    main := math.Floor(ff)
    if ff - main > 0.0 {
        return fmt.Sprintf("%.1f",f);
    } else {
        return fmt.Sprintf("%.f",f);
    }
}

func print_track(track []byte) string {
    out := make([]string,0)
    out = append(out, "|")

    for i,b := range(track) {
        if b != 0 {
            out = append(out,"x")
        } else {
            out = append(out,"-")
        }

        if (i+1)%4==0 {
            out = append(out,"|")
        }
    }

    return strings.Join(out,"")
}

func (p *Pattern) String() string {
    out := make([]string,0)

    out = append(out,fmt.Sprintf("Saved with HW Version: %s\nTempo: %s", p.Version,format_float(p.Tempo)))

    for _,track := range p.Tracks {
        out = append(out,fmt.Sprintf("(%d) %s\t%21s",track.Id,track.Name,print_track(track.Steps)))
    }

    return fmt.Sprintln(strings.Join(out,"\n"));
}

type Track struct{
    Name string
    Id uint32
    Steps []byte
}

func NewTrack() *Track {
    t:=new(Track)
    t.Steps = make([]byte,16)
    return t
}
